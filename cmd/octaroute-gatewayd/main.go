package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"octaroute/internal/config"
	"octaroute/internal/netutil"
	"octaroute/internal/routing"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to JSON config")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	stateStore, err := routing.OpenStateStore(cfg.Database)
	if err != nil {
		log.Fatalf("open routing state: %v", err)
	}
	defer func() {
		_ = stateStore.Close()
	}()

	manager := &routing.Manager{
		DNS: &routing.DNSProxy{
			ListenAddr: cfg.DNS.ListenAddress,
			Upstream:   cfg.DNS.Upstream,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/status", requireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		state, ok, err := stateStore.Load(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if !ok {
			writeJSON(w, http.StatusOK, map[string]any{
				"applied": false,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"applied": true,
			"state":   state,
		})
	}))
	mux.HandleFunc("/apply", requireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req routing.ApplyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		state, err := manager.Apply(r.Context(), req)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if err := stateStore.Save(r.Context(), state); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, state)
	}))

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	log.Printf("gatewayd listening on %s", cfg.Server.Address)
	if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func requireAPIKey(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Server.APIKey == "" {
			next(w, r)
			return
		}
		if r.Header.Get("X-API-Key") != cfg.Server.APIKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func init() {
	if _, ok := os.LookupEnv("TZ"); !ok {
		_ = os.Setenv("TZ", "UTC")
	}
}
