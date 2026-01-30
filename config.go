package main

import (
	"errors"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server struct {
		Listen string `toml:"listen"`
	} `toml:"server"`

	Timeouts struct {
		BackendDial         time.Duration `toml:"backend_dial"`
		HealthcheckDial     time.Duration `toml:"healthcheck_dial"`
		HealthcheckInterval time.Duration `toml:"healthcheck_interval"`
	} `toml:"timeouts"`

	Backends struct {
		Servers []string `toml:"servers"`
	} `toml:"backends"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = "config.toml"
	}

	if _, err := os.Stat(path); err != nil {
		return nil, errors.New("config file not found: " + path)
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	if cfg.Server.Listen == "" {
		cfg.Server.Listen = ":25565"
	}
  
	if cfg.Timeouts.BackendDial == 0 {
		cfg.Timeouts.BackendDial = 5 * time.Second
	}
  
	if cfg.Timeouts.HealthcheckDial == 0 {
		cfg.Timeouts.HealthcheckDial = 3 * time.Second
	}
  
	if cfg.Timeouts.HealthcheckInterval == 0 {
		cfg.Timeouts.HealthcheckInterval = 5 * time.Second
	}

	return &cfg, nil
}
