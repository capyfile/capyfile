package main

import (
	"capyfile/capyfs"
	"capyfile/capysvc"
	"capyfile/capysvc/common"
	"capyfile/files"
	"capyfile/operations"
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Worker struct {
	ServiceDefinitionFile string
	Concurrency           bool
	SleepTime             int
	MaxIterations         int
	LogFile               string
}

func (s *Worker) Init() error {
	capyfs.InitOsFilesystem()

	initLoggerErr := common.InitDefaultWorkerLogger(s.LogFile)
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

func (s *Worker) Run(serviceProcessor string) error {
	if serviceProcessor == "" {
		fmt.Println("error: no service processor provided")
		return nil
	}
	path := strings.Split(serviceProcessor, ":")
	if len(path) != 2 {
		fmt.Println("error: invalid service processor provided (expected format: <service>:<processor>)")
		return nil
	}
	serviceName := path[0]
	processorName := path[1]

	svc := capysvc.FindService(serviceName)
	if svc == nil {
		fmt.Printf("error: service \"%s\" not found\n", serviceName)
		return nil
	}

	proc := svc.FindProcessor(processorName)
	if proc == nil {
		fmt.Printf("error: processor \"%s\" not found\n", processorName)
		return nil
	}

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errorCh := make(chan operations.OperationError)
	notificationCh := make(chan operations.OperationNotification)

	go readErrorChAndLog(serviceName, processorName, errorCh)
	go readNotificationChAndLog(serviceName, processorName, notificationCh)

	var iterations = 1
	for {
		select {
		case <-mainCtx.Done():
			common.Logger.Info(
				"shutting down gracefully",
				slog.String("service", serviceName),
				slog.String("processor", processorName),
			)

			fmt.Println("shutting down gracefully...")

			return nil
		default:
			var out []files.ProcessableFile
			var procErr error
			if s.Concurrency {
				out, procErr = svc.RunProcessorConcurrently(
					capysvc.NewWorkerContext(common.EtcdClient),
					proc.Name,
					[]files.ProcessableFile{},
					errorCh,
					notificationCh,
				)
			} else {
				out, procErr = svc.RunProcessor(
					capysvc.NewWorkerContext(common.EtcdClient),
					proc.Name,
					[]files.ProcessableFile{},
					errorCh,
					notificationCh,
				)
			}
			if procErr != nil {
				common.Logger.Error(
					"service processor error",
					slog.String("service", svc.Name),
					slog.String("processor", proc.Name),
					slog.Any("error", procErr),
				)

				return procErr
			}

			for _, pf := range out {
				freeResourcesErr := pf.FreeResources()
				if freeResourcesErr != nil {
					common.Logger.Error(
						"failed to free resources for file",
						slog.String("service", svc.Name),
						slog.String("processor", proc.Name),
						slog.String("file", pf.Name()),
						slog.Any("error", freeResourcesErr),
					)
				}
			}

			if s.MaxIterations > 0 && iterations >= s.MaxIterations {
				common.Logger.Info(
					"reached maximum number of iterations",
					slog.String("service", serviceName),
					slog.String("processor", processorName),
					slog.Int("iterations", iterations),
				)

				fmt.Println("reached maximum number of iterations")

				return nil
			}

			iterations++

			if s.SleepTime > 0 {
				common.Logger.Info(
					"sleeping",
					slog.String("service", serviceName),
					slog.String("processor", processorName),
					slog.Int("sleepTime", s.SleepTime),
				)

				fmt.Printf("sleeping for %d seconds...\n", s.SleepTime)

				time.Sleep(time.Duration(s.SleepTime) * time.Second)
			}
		}
	}
}

func readErrorChAndLog(svcName, procName string, errorCh chan operations.OperationError) {
	for err := range errorCh {
		var filename = "-"
		var origFilename = "-"
		if err.ProcessableFile != nil {
			filename = err.ProcessableFile.Filename()
			origFilename = err.ProcessableFile.OriginalFilename()
		}

		common.Logger.Error(
			"operation error",
			slog.String("service", svcName),
			slog.String("processor", procName),
			slog.String("operation", err.OperationName),
			slog.Any("error", err.Err),
			slog.String("filename", filename),
			slog.String("origFilename", origFilename),
		)
	}
}

func readNotificationChAndLog(svcName, procName string, notificationCh chan operations.OperationNotification) {
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

		var filename = "-"
		var origFilename = "-"
		if notification.ProcessableFile != nil {
			filename = notification.ProcessableFile.Filename()
			origFilename = notification.ProcessableFile.OriginalFilename()
		}

		common.Logger.Info(
			"operation notification",
			slog.String("service", svcName),
			slog.String("processor", procName),
			slog.String("operation", notification.OperationName),
			slog.String("operationStatus", operationStatus),
			slog.String("operationStatusMessage", notification.OperationStatusMessage),
			slog.String("filename", filename),
			slog.String("origFilename", origFilename),
		)
	}
}
