package routing

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type WireGuardManager struct{}

func (m *WireGuardManager) Ensure(ctx context.Context, nodes []NodeStatus) error {
	for _, node := range nodes {
		iface := node.Interface
		if iface == "" {
			return fmt.Errorf("missing interface name for node %s", node.Name)
		}
		if err := ensureWGInterface(ctx, iface); err != nil {
			return err
		}
		if node.LocalAddress != "" {
			if err := run(ctx, "ip", "address", "replace", node.LocalAddress, "dev", iface); err != nil {
				return fmt.Errorf("assign address %s: %w", iface, err)
			}
		}
		if err := configurePeer(ctx, iface, node.EgressNode); err != nil {
			return err
		}
		if err := run(ctx, "ip", "link", "set", "up", "dev", iface); err != nil {
			return fmt.Errorf("link up %s: %w", iface, err)
		}
		if err := run(ctx, "ip", "route", "replace", "default", "dev", iface, "table", fmt.Sprint(node.TableID)); err != nil {
			return fmt.Errorf("route table %d: %w", node.TableID, err)
		}
	}
	return nil
}

func ensureWGInterface(ctx context.Context, iface string) error {
	if err := run(ctx, "ip", "link", "show", iface); err == nil {
		return nil
	}
	if err := run(ctx, "ip", "link", "add", "dev", iface, "type", "wireguard"); err != nil {
		return fmt.Errorf("create interface %s: %w", iface, err)
	}
	return nil
}

func configurePeer(ctx context.Context, iface string, node EgressNode) error {
	if node.PublicKey == "" || node.Endpoint == "" {
		return fmt.Errorf("missing WireGuard peer data for %s", node.Name)
	}
	args := []string{"set", iface, "peer", node.PublicKey, "endpoint", node.Endpoint}
	if len(node.AllowedIPs) > 0 {
		args = append(args, "allowed-ips", strings.Join(node.AllowedIPs, ","))
	}
	if node.PersistentKeepalive > 0 {
		args = append(args, "persistent-keepalive", fmt.Sprint(node.PersistentKeepalive))
	}
	if err := run(ctx, "wg", args...); err != nil {
		return fmt.Errorf("configure wg %s: %w", iface, err)
	}
	return nil
}

func run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}
