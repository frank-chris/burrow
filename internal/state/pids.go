package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frank-chris/burrow/internal/constants"
)

type TunnelProcess struct {
	Name string `json:"name"`
	PID  int    `json:"pid"`
}

func pidsPath() (string, error) {
	if dir := os.Getenv(constants.ConfigDirEnvVar); dir != "" {
		return filepath.Join(dir, constants.PIDsFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, constants.ConfigDirName, constants.PIDsFileName), nil
}

func SavePIDs(tunnels []TunnelProcess) error {
	path, err := pidsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(tunnels, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadPIDs() ([]TunnelProcess, error) {
	path, err := pidsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not read PID file: %w", err)
	}
	var tunnels []TunnelProcess
	if err := json.Unmarshal(data, &tunnels); err != nil {
		return nil, fmt.Errorf("PID file is corrupted")
	}
	return tunnels, nil
}

func ClearPIDs() error {
	path, err := pidsPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
