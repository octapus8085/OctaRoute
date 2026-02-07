package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	Address       string `json:"address"`
	BindTailscale bool   `json:"bindTailscale"`
}

type Config struct {
	Server   ServerConfig `json:"server"`
	Database string       `json:"database"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address: ":8080",
		},
		Database: "octaroute.db",
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
