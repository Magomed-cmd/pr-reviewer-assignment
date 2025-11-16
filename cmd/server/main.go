package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pr-reviewer-assignment/internal/config"
	"pr-reviewer-assignment/internal/infrastructure"
	"pr-reviewer-assignment/internal/logger"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.NewFromGinMode(cfg.Server.Mode)
	defer func() {
		if err := appLogger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	app, err := infrastructure.NewApp(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("failed to init app", zap.Error(err))
	}
	defer app.Close()

	server := app.HTTPServer()

	go func() {
		appLogger.Info("HTTP server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server stopped unexpectedly", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Graceful shutdown failed", zap.Error(err))
	}
}
