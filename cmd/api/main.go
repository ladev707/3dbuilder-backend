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

	"github.com/joho/godotenv"

	"github.com/ladev707/3dbuilder-backend/internal/api"
	"github.com/ladev707/3dbuilder-backend/internal/config"
	"github.com/ladev707/3dbuilder-backend/pkg/logger"
)

func main() {
	_ = godotenv.Load()
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := api.NewApplication(ctx, cfg)
	if err != nil {
		log.Fatalf("initialize application: %v", err)
	}
	defer application.Close()

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           application.SetupRoutes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info("API started", "port", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("API server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("API shutdown failed", "error", err)
	}
}
