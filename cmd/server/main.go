package main

import (
	"log"

	"pr-reviewer-assignment/internal/config"
	"pr-reviewer-assignment/internal/infrastructure/database/postgres"
	"pr-reviewer-assignment/internal/logger"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	zapLogger := logger.NewFromGinMode(cfg.Server.Mode)
	defer zapLogger.Sync()

	db, err := postgres.NewConnection(&cfg.Database, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	zapLogger.Info("Database connection successful!")
	zapLogger.Info("PR Reviewer Service is ready")
}
