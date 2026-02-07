package routing

import "time"

type ApplyRequest struct {
	Nodes    []EgressNode  `json:"nodes"`
	Policies []PolicyGroup `json:"policies"`
	Routes   []StaticRoute `json:"routes"`
}

type EgressNode struct {
	Name                string   `json:"name"`
	Endpoint            string   `json:"endpoint"`
	PublicKey           string   `json:"publicKey"`
	AllowedIPs          []string `json:"allowedIps"`
	LocalAddress        string   `json:"localAddress"`
	PersistentKeepalive int      `json:"persistentKeepalive"`
}

type PolicyGroup struct {
	Name             string   `json:"name"`
	Node             string   `json:"node"`
	SourceCIDRs      []string `json:"sourceCidrs"`
	DestinationCIDRs []string `json:"destinationCidrs"`
	Domains          []string `json:"domains"`
	Action           string   `json:"action"`
}

type StaticRoute struct {
	CIDR    string `json:"cidr"`
	NextHop string `json:"nextHop"`
	Node    string `json:"node"`
}

type RoutingState struct {
	AppliedAt time.Time      `json:"appliedAt"`
	Nodes     []NodeStatus   `json:"nodes"`
	Policies  []PolicyStatus `json:"policies"`
	Routes    []StaticRoute  `json:"routes"`
}

type NodeStatus struct {
	EgressNode
	Interface string `json:"interface"`
	TableID   int    `json:"tableId"`
}

type PolicyStatus struct {
	PolicyGroup
	Mark   int  `json:"mark"`
	Table  int  `json:"table"`
	Active bool `json:"active"`
}
