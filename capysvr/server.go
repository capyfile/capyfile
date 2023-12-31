package main

import (
	"capyfile/capyerr"
	"capyfile/capyfs"
	"capyfile/capysvc"
	"capyfile/capysvc/common"
	"capyfile/capysvr/httpio"
	"capyfile/files"
	"capyfile/operations"
	"context"
	"errors"
	"golang.org/x/exp/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	ServiceDefinitionFile string
	Concurrency           bool

	Addr     string
	CertFile string
	KeyFile  string

	ShutdownTimeout time.Duration
}

func (s *Server) Init() error {
	// Initialize the filesystem that Capyfile will work with.
	// By default, it's OS filesystem. In theory, it can be changed to any filesystem supported
	// by github.com/spf13/afero package since we use it globally as the filesystem abstraction
	// layer.
	capyfs.InitOsFilesystem()

	initLoggerErr := common.InitDefaultServerLogger()
	if initLoggerErr != nil {
		return initLoggerErr
	}

	// Initialize the etcd client.
	// It's optional. So far, it's only used for service definition parameters loading.
	initEtcdErr := common.InitEtcdClient()
	if initEtcdErr != nil {
		return initEtcdErr
	}

	// Load the service definition after which the server is ready to be running.
	sdLoadErr := capysvc.LoadServiceDefinition(s.ServiceDefinitionFile)
	if sdLoadErr != nil {
		return sdLoadErr
	}

	if addr, ok := os.LookupEnv("SERVER_ADDR"); ok {
		s.Addr = addr
	}

	s.CertFile = os.Getenv("SERVER_CERT_FILE")
	s.KeyFile = os.Getenv("SERVER_KEY_FILE")

	if st, ok := os.LookupEnv("SERVER_SHUTDOWN_TIMEOUT"); ok {
		t, convErr := strconv.Atoi(st)
		if convErr != nil {
			return convErr
		}
		s.ShutdownTimeout = time.Duration(t)
	}

	return nil
}

func (s *Server) Run() error {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:    s.Addr,
		Handler: http.HandlerFunc(s.Handler),
		BaseContext: func(listener net.Listener) context.Context {
			return mainCtx
		},
	}

	go func() {
		common.Logger.Info(
			"server is listening",
			slog.String("addr", s.Addr),
			slog.String("version", version),
			slog.String("commit", commit),
			slog.String("date", date),
		)

		var err error
		if s.CertFile != "" && s.KeyFile != "" {
			err = httpServer.ListenAndServeTLS(s.CertFile, s.KeyFile)
		} else {
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			common.Logger.Error(
				"server error",
				slog.Any("error", err),
			)
		}
	}()

	// Wait for the termination signal or the server to be closed.
	select {
	case <-mainCtx.Done():
		common.Logger.Info("shutting down gracefully")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		common.Logger.Error(
			"shutting down server error",
			slog.Any("error", err),
		)
	}

	common.Logger.Info("server is shut down")

	return nil
}

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// What we have so far is /:service-name/:processor-name as file uploading
	// endpoints (e.g. /messenger/avatar, /messenger/attachment, /images/upload).
	path := strings.Split(
		strings.TrimLeft(r.URL.Path, "/"),
		"/")
	if len(path) != 2 {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				404,
				"NOT_FOUND",
				"not found",
				nil,
			),
			w,
		)
		return
	}

	serviceName := path[0]
	processorName := path[1]

	svc := capysvc.FindService(serviceName)
	if svc == nil {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				404,
				"SERVICE_NOT_FOUND",
				"service not found",
				nil,
			),
			w,
		)
		return
	}

	proc := svc.FindProcessor(processorName)
	if proc == nil {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				404,
				"PROCESSOR_NOT_FOUND",
				"processor not found",
				nil,
			),
			w,
		)
		return
	}

	errorCh := make(chan operations.OperationError)
	notificationCh := make(chan operations.OperationNotification)

	// Log errors and notifications.
	go readErrorChAndLog(svc.Name, proc.Name, errorCh)
	go readNotificationChAndLog(svc.Name, proc.Name, notificationCh)

	// Here we do not check the http input is empty. It gives an option to trigger the
	// processor that retrieves an input from elsewhere (e.g. http, filesystem, some queue).

	var out []files.ProcessableFile
	var procErr error
	if s.Concurrency {
		out, procErr = svc.RunProcessorConcurrently(
			capysvc.NewServerContext(r, common.EtcdClient),
			proc.Name,
			[]files.ProcessableFile{},
			errorCh,
			notificationCh,
		)
	} else {
		out, procErr = svc.RunProcessor(
			capysvc.NewServerContext(r, common.EtcdClient),
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

		var procNotFound *capyerr.ProcessorNotFoundType
		if errors.As(procErr, &procNotFound) {
			_ = httpio.WriteError(
				httpio.NewHTTPAwareError(
					404,
					procNotFound.Code(),
					procNotFound.Message(),
					procErr,
				),
				w,
			)
			return
		}

		var opCfg *capyerr.OperationConfigurationType
		if errors.As(procErr, &opCfg) {
			common.Logger.Error(
				"operation configuration error",
				slog.String("service", svc.Name),
				slog.String("processor", proc.Name),
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
				w,
			)
			return
		}

		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				500,
				"INTERNAL",
				"something went wrong",
				procErr,
			),
			w,
		)
		return
	}

	defer func() {
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
	}()

	common.Logger.Info(
		"input has been processed",
		slog.String("service", svc.Name),
		slog.String("processor", proc.Name),
		slog.Int("outputLength", len(out)),
	)

	outputWriterErr := httpio.WriteOutput(out, w)
	if outputWriterErr != nil {
		_ = httpio.WriteError(
			httpio.NewHTTPAwareError(
				500,
				"REQUEST_OUTPUT_WRITING_FAILURE",
				"request output writing failure",
				outputWriterErr,
			),
			w,
		)
		return
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
