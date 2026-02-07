package routing

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type DNSProxy struct {
	ListenAddr string
	Upstream   string
	NFT        *NFTManager

	mu       sync.RWMutex
	policies map[string]string
	server   *dns.Server
	started  bool
}

func (p *DNSProxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return nil
	}
	if p.ListenAddr == "" {
		p.ListenAddr = "127.0.0.1:5353"
	}
	if p.Upstream == "" {
		p.Upstream = "1.1.1.1:53"
	}
	if p.policies == nil {
		p.policies = make(map[string]string)
	}
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		p.handleQuery(w, r)
	})
	p.server = &dns.Server{
		Addr:    p.ListenAddr,
		Net:     "udp",
		Handler: handler,
	}
	p.started = true
	go func() {
		_ = p.server.ListenAndServe()
	}()
	return nil
}

func (p *DNSProxy) UpdatePolicies(policies []PolicyStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.policies == nil {
		p.policies = make(map[string]string)
	}
	for k := range p.policies {
		delete(p.policies, k)
	}
	for _, policy := range policies {
		for _, domain := range policy.Domains {
			p.policies[normalizeDomain(domain)] = policy.Name
		}
	}
}

func (p *DNSProxy) handleQuery(w dns.ResponseWriter, r *dns.Msg) {
	client := &dns.Client{Timeout: 5 * time.Second}
	resp, _, err := client.Exchange(r, p.Upstream)
	if err != nil {
		_ = w.WriteMsg(&dns.Msg{
			MsgHdr: dns.MsgHdr{Rcode: dns.RcodeServerFailure},
		})
		return
	}
	p.trackAnswers(resp)
	_ = w.WriteMsg(resp)
}

func (p *DNSProxy) trackAnswers(resp *dns.Msg) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.policies) == 0 || p.NFT == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, answer := range resp.Answer {
		switch rr := answer.(type) {
		case *dns.A:
			p.addIP(rr.Hdr.Name, rr.A.String(), ctx)
		case *dns.AAAA:
			p.addIP(rr.Hdr.Name, rr.AAAA.String(), ctx)
		}
	}
}

func (p *DNSProxy) addIP(name string, ip string, ctx context.Context) {
	policyName, ok := p.policies[normalizeDomain(name)]
	if !ok {
		return
	}
	if net.ParseIP(ip) == nil {
		return
	}
	_ = p.NFT.AddDomainIPs(ctx, policyName, []string{ip})
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSuffix(domain, ".")
	return strings.ToLower(domain)
}

func (p *DNSProxy) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.started || p.server == nil {
		return nil
	}
	p.started = false
	shutdownCh := make(chan error, 1)
	go func() {
		shutdownCh <- p.server.Shutdown()
	}()
	select {
	case err := <-shutdownCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *DNSProxy) Status() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	status := map[string]any{
		"listenAddress": p.ListenAddr,
		"upstream":      p.Upstream,
		"policyCount":   len(p.policies),
	}
	if p.server == nil {
		status["running"] = false
	} else {
		status["running"] = p.started
	}
	return status
}
