package routing

import (
	"context"
	"fmt"
	"strings"
)

type NFTManager struct {
	Table  string
	Family string
}

func (m *NFTManager) Ensure(ctx context.Context, policies []PolicyStatus) error {
	if m.Table == "" {
		m.Table = "octaroute"
	}
	if m.Family == "" {
		m.Family = "inet"
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if err := m.ensureChain(ctx); err != nil {
		return err
	}
	if err := run(ctx, "nft", "flush", "chain", m.Family, m.Table, "prerouting"); err != nil {
		return fmt.Errorf("flush nft chain: %w", err)
	}
	for _, policy := range policies {
		if err := m.ensureDNSSet(ctx, policy.Name); err != nil {
			return err
		}
	}
	for _, policy := range policies {
		if err := m.addPolicyRules(ctx, policy); err != nil {
			return err
		}
	}
	return nil
}

func (m *NFTManager) ensureTable(ctx context.Context) error {
	if err := run(ctx, "nft", "list", "table", m.Family, m.Table); err == nil {
		return nil
	}
	if err := run(ctx, "nft", "add", "table", m.Family, m.Table); err != nil {
		return fmt.Errorf("add nft table: %w", err)
	}
	return nil
}

func (m *NFTManager) ensureChain(ctx context.Context) error {
	if err := run(ctx, "nft", "list", "chain", m.Family, m.Table, "prerouting"); err == nil {
		return nil
	}
	rule := fmt.Sprintf("add chain %s %s prerouting { type filter hook prerouting priority mangle ; policy accept ; }", m.Family, m.Table)
	if err := run(ctx, "nft", ruleParts(rule)...); err != nil {
		return fmt.Errorf("add nft chain: %w", err)
	}
	return nil
}

func (m *NFTManager) ensureDNSSet(ctx context.Context, name string) error {
	setName := dnsSetName(name)
	if err := run(ctx, "nft", "list", "set", m.Family, m.Table, setName); err == nil {
		return nil
	}
	rule := fmt.Sprintf("add set %s %s %s { type ipv4_addr ; flags interval ; }", m.Family, m.Table, setName)
	if err := run(ctx, "nft", ruleParts(rule)...); err != nil {
		return fmt.Errorf("add nft set: %w", err)
	}
	return nil
}

func (m *NFTManager) addPolicyRules(ctx context.Context, policy PolicyStatus) error {
	if policy.Action != "" && policy.Action != "allow" {
		return nil
	}
	base := []string{"add", "rule", m.Family, m.Table, "prerouting"}
	match := []string{}
	if len(policy.SourceCIDRs) > 0 {
		match = append(match, "ip", "saddr", "{", strings.Join(policy.SourceCIDRs, ","), "}")
	}
	if len(policy.DestinationCIDRs) > 0 {
		match = append(match, "ip", "daddr", "{", strings.Join(policy.DestinationCIDRs, ","), "}")
	}
	mark := []string{"meta", "mark", "set", fmt.Sprint(policy.Mark)}
	if len(match) > 0 || (len(policy.SourceCIDRs) == 0 && len(policy.DestinationCIDRs) == 0 && len(policy.Domains) == 0) {
		args := append(base, match...)
		args = append(args, mark...)
		if err := run(ctx, "nft", args...); err != nil {
			return fmt.Errorf("add nft rule: %w", err)
		}
	}
	if len(policy.Domains) > 0 {
		args := append(base, "ip", "daddr", "@"+dnsSetName(policy.Name))
		if len(policy.SourceCIDRs) > 0 {
			args = append(base, "ip", "saddr", "{", strings.Join(policy.SourceCIDRs, ","), "}", "ip", "daddr", "@"+dnsSetName(policy.Name))
		}
		args = append(args, mark...)
		if err := run(ctx, "nft", args...); err != nil {
			return fmt.Errorf("add nft dns rule: %w", err)
		}
	}
	return nil
}

func (m *NFTManager) AddDomainIPs(ctx context.Context, policyName string, ips []string) error {
	if len(ips) == 0 {
		return nil
	}
	args := []string{"add", "element", m.Family, m.Table, dnsSetName(policyName), "{" + strings.Join(ips, ",") + "}"}
	if err := run(ctx, "nft", args...); err != nil {
		return fmt.Errorf("add nft element: %w", err)
	}
	return nil
}

func dnsSetName(policyName string) string {
	return "dns_" + sanitizeName(policyName)
}

func sanitizeName(value string) string {
	value = strings.ToLower(value)
	value = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '_':
			return r
		case r == '-':
			return '_'
		default:
			return '_'
		}
	}, value)
	return value
}

func ruleParts(rule string) []string {
	return strings.Fields(rule)
}
