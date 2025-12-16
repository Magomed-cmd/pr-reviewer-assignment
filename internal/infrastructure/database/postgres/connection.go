package postgres

import (
	"context"
	"strings"
	"time"

	"pr-reviewer-assignment/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewConnection(cfg *config.DatabaseConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	dsn := cfg.GetDSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error("Failed to parse DSN", zap.Error(err))
		return nil, err
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	logger.Debug("Connecting to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int32("max_conns", poolConfig.MaxConns),
	)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("Failed to connect to Postgres", zap.Error(err))
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error("Failed to ping Postgres", zap.Error(err))
		return nil, err
	}

	safeDSN := strings.Replace(dsn, cfg.Password, "**hidden**", 1)
	logger.Info("PostgreSQL connection established",
		zap.String("dsn", safeDSN),
		zap.Int32("pool_max_conns", poolConfig.MaxConns),
	)

	return pool, nil
}
