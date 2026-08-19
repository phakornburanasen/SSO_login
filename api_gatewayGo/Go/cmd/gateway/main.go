package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"api-gateway-go/internal/config"
	"api-gateway-go/internal/gateway"
)

func main() {
	var (
		addr                  string
		configPath            string
		reloadInterval        time.Duration
		upstreamHeaderTimeout time.Duration
		shutdownTimeout       time.Duration
		allowCredentialedCORS bool
	)

	flag.StringVar(&addr, "addr", defaultAddr(), "gateway listen address")
	flag.StringVar(&configPath, "config", "", "path to service config json")
	flag.DurationVar(&reloadInterval, "reload-interval", config.DefaultReloadInterval, "service config reload check interval")
	flag.DurationVar(&upstreamHeaderTimeout, "upstream-header-timeout", 300*time.Second, "upstream response header timeout")
	flag.DurationVar(&shutdownTimeout, "shutdown-timeout", 15*time.Second, "graceful shutdown timeout")
	flag.BoolVar(&allowCredentialedCORS, "cors-credentials", true, "set Access-Control-Allow-Credentials for browser requests")
	flag.Parse()

	resolvedConfigPath, err := config.ResolvePath(configPath)
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}

	serviceStore := config.NewStore(resolvedConfigPath, reloadInterval)
	if err := serviceStore.Load(); err != nil {
		log.Fatalf("load services: %v", err)
	}

	handler := gateway.New(serviceStore, gateway.Options{
		Transport: gateway.NewTransport(upstreamHeaderTimeout),
		CORS: gateway.CORSOptions{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: allowCredentialedCORS,
		},
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("gateway listening on %s", addr)
		log.Printf("service config: %s", serviceStore.Path())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Printf("gateway shutting down")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
}

func defaultAddr() string {
	if addr := strings.TrimSpace(os.Getenv("ADDR")); addr != "" {
		return addr
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return fmt.Sprintf(":%s", port)
	}
	return ":60000"
}
