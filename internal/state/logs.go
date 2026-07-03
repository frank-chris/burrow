package state

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frank-chris/burrow/internal/constants"
)

func LogDir() (string, error) {
	if dir := os.Getenv(constants.ConfigDirEnvVar); dir != "" {
		return filepath.Join(dir, constants.LogsDir), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, constants.ConfigDirName, constants.LogsDir), nil
}

func LogPath(name string) (string, error) {
	dir, err := LogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".log"), nil
}
