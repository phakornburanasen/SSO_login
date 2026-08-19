package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// api_gatewayGo — reverse proxy
//   Listens on :18000 (configurable via GATEWAY_ADDR)
//   Strips the public prefix /api/SSO_login and forwards the rest
//   to the SSO Login backend on :12080 (configurable via BACKEND_URL)
//
//   Examples:
//     http://localhost:18000/api/SSO_login/api/auth/login
//       -> http://localhost:12080/api/auth/login
//     http://localhost:18000/api/SSO_login/health
//       -> http://localhost:12080/health

const (
	defaultGatewayAddr = ":18000"
	defaultBackendURL  = "http://127.0.0.1:12080"
	publicPrefix       = "/api/SSO_login"
)

func main() {
	gatewayAddr := envOr("GATEWAY_ADDR", defaultGatewayAddr)
	backendRaw := envOr("BACKEND_URL", defaultBackendURL)
	backendRaw = strings.TrimRight(backendRaw, "/")

	backendURL, err := url.Parse(backendRaw)
	if err != nil {
		log.Fatalf("invalid BACKEND_URL %q: %v", backendRaw, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[gateway] upstream error: %v (path=%s)", err, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream backend unavailable"}`))
	}

	mux := http.NewServeMux()

	// /health — own endpoint, needs CORS for browser access
	mux.Handle("/health", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"api_gatewayGo"}`))
	})))

	// All other paths: reverse proxy.
	// NOTE: do NOT add CORS headers here — the backend already sets them
	// via its own CORS middleware. Adding them again would create duplicate
	// Access-Control-Allow-Origin headers, which Chrome rejects with
	// "net::ERR_FAILED" (only one ACAO header is allowed per response).
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only proxy paths that begin with the public prefix
		if !strings.HasPrefix(r.URL.Path, publicPrefix) {
			http.NotFound(w, r)
			return
		}

		// Rewrite: /api/SSO_login/api/auth/login -> /api/auth/login
		original := r.URL.Path
		r.URL.Path = strings.TrimPrefix(original, publicPrefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		log.Printf("[gateway] %s %s -> %s%s",
			r.Method, original, backendURL.String(), r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:              gatewayAddr,
		Handler:           logMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("api_gatewayGo listening on %s -> %s (prefix=%s)",
			gatewayAddr, backendURL.String(), publicPrefix)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("gateway: %v", err)
		}
	}()

	<-stop
	log.Printf("shutting down gateway...")
	shCtx, shCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shCancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func envOr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}

// ---- middleware ----

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		log.Printf("[gateway] %s %s -> %d (%s)",
			r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
