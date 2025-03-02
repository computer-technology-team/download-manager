package logging

import (
	"log/slog"
	"os"
)

func InitializeLogger() (func() error, error) {
	logFile, err := os.OpenFile("download-manager.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return logFile.Close, err
	}

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	return logFile.Close, nil
}
