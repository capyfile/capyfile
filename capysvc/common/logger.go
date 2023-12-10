package common

import (
	"golang.org/x/exp/slog"
	"io"
	"os"
)

var Logger *slog.Logger

func InitDefaultServerLogger() error {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout))
	return nil
}

func InitDefaultCliLogger() error {
	//logFile, err := os.OpenFile("/var/log/capyfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	logFile, err := os.OpenFile("/tmp/capyfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	Logger = slog.New(slog.NewTextHandler(logFile))

	return nil
}

func InitDefaultWorkerLogger(filename string) error {
	if filename == "" {
		Logger = slog.New(slog.NewJSONHandler(os.Stdout))

		return nil
	}

	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	Logger = slog.New(slog.NewJSONHandler(logFile))

	return nil
}

func InitTestLogger() error {
	Logger = slog.New(slog.NewJSONHandler(io.Discard))
	return nil
}
