package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/frank-chris/burrow/internal/constants"
	"gopkg.in/yaml.v3"
)

type TunnelConfig struct {
	Name   string `yaml:"name"`
	Port   int    `yaml:"port"`
	Domain string `yaml:"domain"`
}

type Config struct {
	Provider string         `yaml:"provider"`
	Tunnels  []TunnelConfig `yaml:"tunnels"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = constants.BurrowConfigFile
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no %s found in current directory - create one first", constants.BurrowConfigFile)
		}
		return nil, fmt.Errorf("could not read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Tunnels) == 0 {
		return fmt.Errorf("no tunnels defined in %s", constants.BurrowConfigFile)
	}
	seen := make(map[string]bool)
	for _, t := range c.Tunnels {
		if t.Name == "" {
			return fmt.Errorf("a tunnel entry is missing a name")
		}
		if seen[t.Name] {
			return fmt.Errorf("duplicate tunnel name %q", t.Name)
		}
		seen[t.Name] = true
		if t.Port < 1 || t.Port > 65535 {
			return fmt.Errorf("tunnel %q has invalid port %d", t.Name, t.Port)
		}
		if t.Domain == "" {
			return fmt.Errorf("tunnel %q is missing a domain", t.Name)
		}
	}
	return nil
}
