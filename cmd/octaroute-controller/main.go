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
	"octaroute/internal/controllerdb"
	"octaroute/internal/netutil"
)

type apiServer struct {
	store *controllerdb.Store
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to JSON config")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store, err := controllerdb.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer store.Close()

	api := &apiServer{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/nodes", api.handleNodes)
	mux.HandleFunc("/api/policies", api.handlePolicies)
	mux.HandleFunc("/api/routes", api.handleRoutes)

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ln, err := netutil.ListenWithOptionalDevice(ctx, "tcp", cfg.Server.Address, bindDevice(cfg.Server.BindTailscale))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("controller listening on %s", cfg.Server.Address)
	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func bindDevice(bindTailscale bool) string {
	if bindTailscale {
		return "tailscale0"
	}
	return ""
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (a *apiServer) handleNodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		nodes, err := a.store.ListNodes(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, nodes)
	case http.MethodPost:
		var n controllerdb.Node
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if n.Name == "" || n.Address == "" || n.Zone == "" {
			writeError(w, http.StatusBadRequest, errMissingFields)
			return
		}
		created, err := a.store.CreateNode(r.Context(), n)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *apiServer) handlePolicies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		policies, err := a.store.ListPolicies(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, policies)
	case http.MethodPost:
		var p controllerdb.Policy
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if p.Name == "" || p.Source == "" || p.Destination == "" || p.Action == "" {
			writeError(w, http.StatusBadRequest, errMissingFields)
			return
		}
		created, err := a.store.CreatePolicy(r.Context(), p)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *apiServer) handleRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		routes, err := a.store.ListRoutes(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, routes)
	case http.MethodPost:
		var rt controllerdb.Route
		if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if rt.CIDR == "" || rt.NextHop == "" || rt.NodeID == 0 {
			writeError(w, http.StatusBadRequest, errMissingFields)
			return
		}
		created, err := a.store.CreateRoute(r.Context(), rt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

var errMissingFields = &apiError{Message: "missing required fields"}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type apiError struct {
	Message string `json:"message"`
}

func (e *apiError) Error() string {
	return e.Message
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, apiError{Message: err.Error()})
}

func init() {
	if _, ok := os.LookupEnv("TZ"); !ok {
		_ = os.Setenv("TZ", "UTC")
	}
}
