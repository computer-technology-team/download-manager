package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/computer-technology-team/download-manager.git/datadir"
)

func InitializeLogger() (func() error, error) {
	appDataDir, err := datadir.GetAppDataDir()
	if err != nil {
		return nil, fmt.Errorf("could not get app data directory: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(appDataDir, "download-manager.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return logFile.Close, fmt.Errorf("could not open log file: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	return logFile.Close, nil
}
