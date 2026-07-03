package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/frank-chris/burrow/internal/constants"
)

type Auth struct {
	Provider  string `json:"provider"`
	APIToken  string `json:"api_token"`
	AccountID string `json:"account_id"`
}

func authPath() (string, error) {
	if dir := os.Getenv(constants.ConfigDirEnvVar); dir != "" {
		return filepath.Join(dir, constants.AuthFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, constants.ConfigDirName, constants.AuthFileName), nil
}

func AuthExists() bool {
	path, err := authPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func LoadAuth() (*Auth, error) {
	path, err := authPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not initialized - run `burrow init` first")
		}
		return nil, fmt.Errorf("could not read auth file: %w", err)
	}
	var auth Auth
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("auth file is corrupted - run `burrow init` to reconfigure")
	}
	if auth.APIToken == "" || auth.AccountID == "" {
		return nil, fmt.Errorf("auth file is incomplete - run `burrow init` to reconfigure")
	}
	return &auth, nil
}

func SaveAuth(auth *Auth) error {
	path, err := authPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("could not save credentials: %w", err)
	}
	return nil
}
