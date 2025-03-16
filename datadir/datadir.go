package datadir

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const appName = "download-manager"

func GetAppDataDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			return "", errors.New("LOCALAPPDATA environment variable not set")
		}
		baseDir = filepath.Join(appData, appName)
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(homeDir, "Library", "Application Support", appName)
	default:
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			dataHome = filepath.Join(homeDir, ".local", "share")
		}
		baseDir = filepath.Join(dataHome, appName)
	}

	return baseDir, os.MkdirAll(baseDir, 0755)
}
