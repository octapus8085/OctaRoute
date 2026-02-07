package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	Address       string `json:"address"`
	BindTailscale bool   `json:"bindTailscale"`
	APIKey        string `json:"apiKey"`
}

type DNSConfig struct {
	ListenAddress string `json:"listenAddress"`
	Upstream      string `json:"upstream"`
}

type Config struct {
	Server   ServerConfig `json:"server"`
	Database string       `json:"database"`
	DNS      DNSConfig    `json:"dns"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address: ":8080",
		},
		Database: "octaroute.db",
		DNS: DNSConfig{
			ListenAddress: "127.0.0.1:5353",
			Upstream:      "1.1.1.1:53",
		},
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
