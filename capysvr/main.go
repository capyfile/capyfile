package main

import (
	"capyfile/capyerr"
	"capyfile/capyfs"
	"capyfile/capysvc"
	"capyfile/capysvc/common"
	"capyfile/capysvc/httpio"
	"capyfile/capysvc/service"
	"errors"
	"fmt"
	"golang.org/x/exp/slog"
	"net/http"
)

func main() {
	capyfs.InitOsFilesystem()

	initLoggerErr := common.InitLogger()
	if initLoggerErr != nil {
		panic(initLoggerErr)
	}

	initEtcdClientErr := common.InitEtcdClient()
	if initEtcdClientErr != nil {
		panic(initEtcdClientErr)
	}

	common.Logger.Info("server is initialized")

	sdLoadErr := capysvc.LoadServiceDefinition()
	if sdLoadErr != nil {
		panic(sdLoadErr)
	}

	common.Logger.Info("service definition is loaded")

	http.HandleFunc(
		fmt.Sprintf("/%s/", capysvc.ServiceDefinition.Name),
		processFiles)

	common.Logger.Info("listening on port 8024")

	svrErr := http.ListenAndServe(":8024", nil)
	if svrErr != nil {
		panic(svrErr)
	}
}

func processFiles(respWriter http.ResponseWriter, req *http.Request) {
	common.Logger.Info(
		"serving request",
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path),
	)

	respWriter.Header().Set("Content-Type", "application/json")

	// We have something like "/:serviceName/:processorName"
	var serviceProcessor = req.URL.Path
	var processorName = req.URL.Path[len(capysvc.ServiceDefinition.Name)+2:]

	in, inputReaderErr := httpio.ReadInput(req)
	if inputReaderErr != nil {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				500,
				"REQUEST_INPUT_READING_FAILURE",
				"request input reading failure",
				inputReaderErr,
			),
			respWriter,
		)
		return
	}

	common.Logger.Info(
		"input has been extracted",
		slog.String("serviceProcessor", serviceProcessor),
	)

	if len(in) == 0 {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				400,
				"NO_INPUT_PROVIDED",
				"no input provided",
				nil,
			),
			respWriter,
		)
		return
	}

	out, procErr := capysvc.ServiceDefinition.RunProcessor(
		&service.ServerContext{
			Req:        req,
			EtcdClient: common.EtcdClient,
		},
		processorName,
		in)
	if procErr != nil {
		common.Logger.Error(
			"service processor error",
			slog.String("serviceProcessor", serviceProcessor),
			slog.Any("error", procErr),
		)

		var procNotFound *capyerr.ProcessorNotFoundType
		if errors.As(procErr, &procNotFound) {
			_ = httpio.WriteError(
				httpio.NewHTTPAwareError(
					404,
					procNotFound.Code(),
					procNotFound.Message(),
					procErr,
				),
				respWriter,
			)
			return
		}

		var opCfg *capyerr.OperationConfigurationType
		if errors.As(procErr, &opCfg) {
			common.Logger.Error(
				"operation configuration error",
				slog.String("serviceProcessor", serviceProcessor),
				slog.Any("error", opCfg),
				slog.Any("origError", opCfg.OriginalError()),
			)

			_ = httpio.WriteError(
				httpio.NewHTTPAwareError(
					500,
					opCfg.Code(),
					opCfg.Message(),
					procErr,
				),
				respWriter,
			)
			return
		}

		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				500,
				"UNKNOWN",
				"unknown",
				procErr,
			),
			respWriter,
		)
		return
	}

	common.Logger.Info(
		"input has been processed",
		slog.String("serviceProcessor", serviceProcessor),
	)

	outputWriterErr := httpio.WriteOutput(out, respWriter)
	if outputWriterErr != nil {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				500,
				"REQUEST_OUTPUT_WRITING_FAILURE",
				"request output writing failure",
				outputWriterErr,
			),
			respWriter,
		)
		return
	}
}
