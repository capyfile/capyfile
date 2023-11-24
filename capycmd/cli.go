package main

import (
	"capyfile/capyfs"
	"capyfile/capysvc"
	"capyfile/capysvc/common"
	"capyfile/files"
	"capyfile/operations"
	"fmt"
	"strings"
)

type Cli struct {
	ServiceDefinitionFile string
}

func (s *Cli) Init() error {
	// Initialize the filesystem that Capyfile will work with.
	// By default, it's OS filesystem. In theory, it can be changed to any filesystem supported
	// by github.com/spf13/afero package since we use it globally as the filesystem abstraction
	// layer.
	capyfs.InitOsFilesystem()

	initLoggerErr := common.InitDefaultCliLogger()
	if initLoggerErr != nil {
		return initLoggerErr
	}

	// Load the service definition after which the cli app is ready to be running.
	sdLoadErr := capysvc.LoadServiceDefinition(s.ServiceDefinitionFile)
	if sdLoadErr != nil {
		return sdLoadErr
	}

	return nil
}

func (s *Cli) Run(serviceProcessor string) error {
	if serviceProcessor == "" {
		printlnErrorMsg("No service processor provided")
		return nil
	}
	path := strings.Split(serviceProcessor, ":")
	if len(path) != 2 {
		printlnErrorMsg("Invalid service processor provided. Expected format: <service>:<processor>")
		return nil
	}
	serviceName := path[0]
	processorName := path[1]

	svc := capysvc.FindService(serviceName)
	if svc == nil {
		printlnErrorMsg(
			fmt.Sprintf("Service \"%s\" not found", serviceName))
		return nil
	}

	proc := svc.FindProcessor(processorName)
	if proc == nil {
		printlnErrorMsg(
			fmt.Sprintf("Processor \"%s\" not found", processorName))
		return nil
	}

	errorCh := make(chan operations.OperationError)
	notificationCh := make(chan operations.OperationNotification)

	go func() {
		for err := range errorCh {
			printlnErrorMsg(
				fmt.Sprintf("%s: %s", err.OperationName, err.Err.Error()))
		}
	}()
	go func() {
		for notification := range notificationCh {
			var operationStatus string
			switch notification.OperationStatus {
			case operations.StatusSkipped:
				operationStatus = "SKIPPED"
			case operations.StatusStarted:
				operationStatus = "STARTED"
			case operations.StatusFinished:
				operationStatus = "FINISHED"
			case operations.StatusFailed:
				operationStatus = "FAILED"
			}

			var fileName = "-"
			if notification.ProcessableFile != nil {
				fileName = notification.ProcessableFile.OriginalFilename()
			}

			fmt.Println(
				fmt.Sprintf(
					"[%s] \033[1m%s %s\033[0m %s",
					fileName,
					notification.OperationName,
					operationStatus,
					notification.OperationStatusMessage,
				),
			)
		}
	}()

	fmt.Println(fmt.Sprintf("Running %s:%s service processor...", svc.Name, proc.Name))
	fmt.Println()

	out, procErr := svc.RunProcessorConcurrently(
		capysvc.NewCliContext(),
		proc.Name,
		[]files.ProcessableFile{},
		errorCh,
		notificationCh,
	)
	if procErr != nil {
		printlnErrorMsg("Failed to run the processor: " + procErr.Error())
		return nil
	}

	defer func() {
		for _, pf := range out {
			freeResourcesErr := pf.FreeResources()
			if freeResourcesErr != nil {
				printlnErrorMsg(
					fmt.Sprintf("Failed to free resources for %s: %s", pf.Name(), freeResourcesErr.Error()))
			}
		}
	}()

	fmt.Println()
	printlnSuccessMsg("Completed")

	var failure []files.ProcessableFile
	for _, pf := range out {
		if pf.HasFileProcessingError() {
			failure = append(failure, pf)
		}
	}
	if len(failure) > 0 {
		fmt.Println()
		fmt.Println("Failures:")
		for _, pf := range failure {
			fmt.Println(
				fmt.Sprintf("    [%s] %s", pf.OriginalFilename(), pf.FileProcessingError.Error()))
		}
	}

	return nil
}

func printlnErrorMsg(msg string) {
	fmt.Println("\033[31m" + msg + "\033[0m")
}

func printlnSuccessMsg(msg string) {
	fmt.Println("\033[32m" + msg + "\033[0m")
}
