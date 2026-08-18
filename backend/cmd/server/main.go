package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sso-login/backend/internal/config"
	"sso-login/backend/internal/httpapi"
	"sso-login/backend/internal/service"
	"sso-login/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := store.Open(ctx, cfg.DatabaseURL, cfg.DBMaxOpen, cfg.DBMaxLifetime)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer repo.Close()
	log.Printf("connected to PostgreSQL")

	svc := service.New(repo)
	handler := httpapi.New(svc, cfg)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("sso-login backend listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-stop
	log.Printf("shutting down...")
	shCtx, shCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shCancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
