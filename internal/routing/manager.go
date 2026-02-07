package routing

import (
	"context"
	"fmt"
	"time"
)

type Manager struct {
	WireGuard *WireGuardManager
	NFT       *NFTManager
	DNS       *DNSProxy
}

func (m *Manager) Apply(ctx context.Context, req ApplyRequest) (RoutingState, error) {
	if m.WireGuard == nil {
		m.WireGuard = &WireGuardManager{}
	}
	if m.NFT == nil {
		m.NFT = &NFTManager{}
	}
	if m.DNS == nil {
		m.DNS = &DNSProxy{NFT: m.NFT}
	} else if m.DNS.NFT == nil {
		m.DNS.NFT = m.NFT
	}
	nodeStatuses, policyStatuses, err := m.buildStatus(req)
	if err != nil {
		return RoutingState{}, err
	}
	if err := m.WireGuard.Ensure(ctx, nodeStatuses); err != nil {
		return RoutingState{}, err
	}
	if err := ensureIPRules(ctx, nodeStatuses); err != nil {
		return RoutingState{}, err
	}
	if err := m.NFT.Ensure(ctx, policyStatuses); err != nil {
		return RoutingState{}, err
	}
	if err := m.DNS.Start(); err != nil {
		return RoutingState{}, err
	}
	m.DNS.UpdatePolicies(policyStatuses)
	state := RoutingState{
		AppliedAt: time.Now().UTC(),
		Nodes:     nodeStatuses,
		Policies:  policyStatuses,
		Routes:    req.Routes,
	}
	return state, nil
}

func (m *Manager) buildStatus(req ApplyRequest) ([]NodeStatus, []PolicyStatus, error) {
	nodeStatuses := make([]NodeStatus, 0, len(req.Nodes))
	nodeTable := make(map[string]int, len(req.Nodes))
	for i, node := range req.Nodes {
		tableID := 101 + i
		nodeTable[node.Name] = tableID
		nodeStatuses = append(nodeStatuses, NodeStatus{
			EgressNode: node,
			Interface:  fmt.Sprintf("wg-egress-%s", sanitizeName(node.Name)),
			TableID:    tableID,
		})
	}
	policyStatuses := make([]PolicyStatus, 0, len(req.Policies))
	for _, policy := range req.Policies {
		tableID, ok := nodeTable[policy.Node]
		if !ok && policy.Node != "" {
			return nil, nil, fmt.Errorf("unknown node %s for policy %s", policy.Node, policy.Name)
		}
		if policy.Node == "" && len(nodeStatuses) > 0 {
			tableID = nodeStatuses[0].TableID
		}
		mark := tableID
		policyStatuses = append(policyStatuses, PolicyStatus{
			PolicyGroup: policy,
			Mark:        mark,
			Table:       tableID,
			Active:      true,
		})
	}
	return nodeStatuses, policyStatuses, nil
}

func ensureIPRules(ctx context.Context, nodes []NodeStatus) error {
	for _, node := range nodes {
		if err := run(ctx, "ip", "rule", "del", "fwmark", fmt.Sprint(node.TableID), "lookup", fmt.Sprint(node.TableID)); err != nil {
			// ignore delete errors
		}
		if err := run(ctx, "ip", "rule", "add", "fwmark", fmt.Sprint(node.TableID), "lookup", fmt.Sprint(node.TableID)); err != nil {
			return fmt.Errorf("ip rule for table %d: %w", node.TableID, err)
		}
	}
	return nil
}
