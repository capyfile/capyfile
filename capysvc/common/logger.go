package common

import (
	"golang.org/x/exp/slog"
	"io"
	"os"
)

var Logger *slog.Logger

func InitLogger() error {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout))
	return nil
}

func InitTestLogger() error {
	Logger = slog.New(slog.NewJSONHandler(io.Discard))
	return nil
}
