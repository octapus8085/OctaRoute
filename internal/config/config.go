package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ServerConfig defines how the HTTP server listens for control plane traffic.
type ServerConfig struct {
	Address       string `json:"address"`
	BindTailscale bool   `json:"bindTailscale"`
	APIKey        string `json:"apiKey"`
}

type DNSConfig struct {
	ListenAddress string `json:"listenAddress"`
	Upstream      string `json:"upstream"`
}

// AuthConfig defines the API key header used by control endpoints.
type AuthConfig struct {
	APIKey string `json:"apiKey"`
	Header string `json:"header"`
}

// WireGuardConfig describes the exit node WireGuard interface settings.
type WireGuardConfig struct {
	Enabled        bool   `json:"enabled"`
	Interface      string `json:"interface"`
	ListenPort     int    `json:"listenPort"`
	PrivateKeyPath string `json:"privateKeyPath"`
	Address        string `json:"address"`
}

// NATConfig defines nftables masquerade settings.
type NATConfig struct {
	Enable            bool   `json:"enable"`
	ExternalInterface string `json:"externalInterface"`
	InternalInterface string `json:"internalInterface"`
}

// Config is the top-level configuration for OctaRoute services.
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  string          `json:"database"`
	Auth      AuthConfig      `json:"auth"`
	WireGuard WireGuardConfig `json:"wireguard"`
	NAT       NATConfig       `json:"nat"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address: ":8080",
		},
		Database: "octaroute.db",
		Auth: AuthConfig{
			Header: "X-API-Key",
		},
		WireGuard: WireGuardConfig{
			Interface: "wg0",
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

	if cfg.Auth.Header == "" {
		cfg.Auth.Header = "X-API-Key"
	}
	if cfg.WireGuard.Interface == "" {
		cfg.WireGuard.Interface = "wg0"
	}

	return cfg, nil
}
