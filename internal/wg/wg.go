package wg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Interface      string
	ListenPort     int
	PrivateKeyPath string
	Address        string
}

func EnsureServer(ctx context.Context, cfg Config) error {
	if cfg.Interface == "" {
		return fmt.Errorf("wireguard interface is required")
	}
	if cfg.ListenPort <= 0 {
		return fmt.Errorf("wireguard listen port must be set")
	}
	if cfg.PrivateKeyPath == "" {
		return fmt.Errorf("wireguard private key path must be set")
	}
	if _, err := os.Stat(cfg.PrivateKeyPath); err != nil {
		return fmt.Errorf("wireguard private key: %w", err)
	}

	exists := true
	if err := run(ctx, "ip", "link", "show", "dev", cfg.Interface); err != nil {
		exists = false
	}
	if !exists {
		if err := run(ctx, "ip", "link", "add", "dev", cfg.Interface, "type", "wireguard"); err != nil {
			return fmt.Errorf("create wireguard interface: %w", err)
		}
	}

	if cfg.Address != "" {
		addrOut, err := output(ctx, "ip", "-brief", "addr", "show", "dev", cfg.Interface)
		if err != nil {
			return fmt.Errorf("read interface address: %w", err)
		}
		if !strings.Contains(addrOut, cfg.Address) {
			if err := run(ctx, "ip", "address", "add", cfg.Address, "dev", cfg.Interface); err != nil {
				return fmt.Errorf("assign address: %w", err)
			}
		}
	}

	if err := run(ctx, "wg", "set", cfg.Interface, "listen-port", fmt.Sprintf("%d", cfg.ListenPort), "private-key", cfg.PrivateKeyPath); err != nil {
		return fmt.Errorf("configure wireguard: %w", err)
	}

	if err := run(ctx, "ip", "link", "set", "up", "dev", cfg.Interface); err != nil {
		return fmt.Errorf("bring up interface: %w", err)
	}

	return nil
}

func run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func output(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(data), nil
}
