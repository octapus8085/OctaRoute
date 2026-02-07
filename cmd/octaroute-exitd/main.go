package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"octaroute/internal/config"
	"octaroute/internal/netutil"
	"octaroute/internal/nft"
	"octaroute/internal/wg"
)

type healthMetrics struct {
	LatencyMs float64  `json:"latency_ms"`
	LossPct   float64  `json:"loss_pct"`
	HTTPSMs   float64  `json:"https_ms"`
	JitterMs  float64  `json:"jitter_ms"`
	BWMbps    *float64 `json:"bw_mbps,omitempty"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to JSON config")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if !cfg.Server.BindTailscale {
		log.Fatal("server.bindTailscale must be true to bind tailscale0")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.WireGuard.Enabled {
		if err := wg.EnsureServer(ctx, wg.Config{
			Interface:      cfg.WireGuard.Interface,
			ListenPort:     cfg.WireGuard.ListenPort,
			PrivateKeyPath: cfg.WireGuard.PrivateKeyPath,
			Address:        cfg.WireGuard.Address,
		}); err != nil {
			log.Fatalf("wireguard setup: %v", err)
		}
	}

	if cfg.NAT.Enable {
		if err := nft.EnsureMasquerade(ctx, nft.MasqueradeConfig{
			ExternalInterface: cfg.NAT.ExternalInterface,
			InternalInterface: cfg.NAT.InternalInterface,
		}); err != nil {
			log.Fatalf("nat setup: %v", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		metrics := healthMetrics{}
		writeJSON(w, http.StatusOK, metrics)
	})

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := netutil.ListenWithOptionalDevice(ctx, "tcp", cfg.Server.Address, "tailscale0")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("exitd listening on %s", cfg.Server.Address)
	if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func init() {
	if _, ok := os.LookupEnv("TZ"); !ok {
		_ = os.Setenv("TZ", "UTC")
	}
}
