package nft

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type MasqueradeConfig struct {
	ExternalInterface string
	InternalInterface string
}

func EnsureMasquerade(ctx context.Context, cfg MasqueradeConfig) error {
	if cfg.ExternalInterface == "" {
		return fmt.Errorf("nat external interface is required")
	}

	if err := ensureTable(ctx); err != nil {
		return err
	}
	if err := ensurePostroutingChain(ctx); err != nil {
		return err
	}
	return ensureMasqueradeRule(ctx, cfg)
}

func ensureTable(ctx context.Context) error {
	if err := run(ctx, "nft", "list", "table", "ip", "nat"); err == nil {
		return nil
	}
	if err := run(ctx, "nft", "add", "table", "ip", "nat"); err != nil {
		return fmt.Errorf("create nat table: %w", err)
	}
	return nil
}

func ensurePostroutingChain(ctx context.Context) error {
	if err := run(ctx, "nft", "list", "chain", "ip", "nat", "postrouting"); err == nil {
		return nil
	}
	args := []string{"add", "chain", "ip", "nat", "postrouting", "{", "type", "nat", "hook", "postrouting", "priority", "100", ";", "}"}
	if err := run(ctx, "nft", args...); err != nil {
		return fmt.Errorf("create postrouting chain: %w", err)
	}
	return nil
}

func ensureMasqueradeRule(ctx context.Context, cfg MasqueradeConfig) error {
	out, err := output(ctx, "nft", "list", "chain", "ip", "nat", "postrouting")
	if err != nil {
		return fmt.Errorf("list postrouting chain: %w", err)
	}

	ruleSnippet := fmt.Sprintf("oifname \"%s\" masquerade", cfg.ExternalInterface)
	args := []string{"add", "rule", "ip", "nat", "postrouting"}
	if cfg.InternalInterface != "" {
		ruleSnippet = fmt.Sprintf("iifname \"%s\" %s", cfg.InternalInterface, ruleSnippet)
		args = append(args, "iifname", cfg.InternalInterface)
	}
	args = append(args, "oifname", cfg.ExternalInterface, "masquerade")

	if strings.Contains(out, ruleSnippet) {
		return nil
	}

	if err := run(ctx, "nft", args...); err != nil {
		return fmt.Errorf("add masquerade rule: %w", err)
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
