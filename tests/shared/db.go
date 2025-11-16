package shared

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pr-reviewer-assignment/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

const TestDBAdvisoryLockID = 987654321

var MigrationsDir = filepath.Join("..", "..", "migrations")

func LoadTestDBConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     envOr("TEST_DB_HOST", "127.0.0.1"),
		Port:     envOrInt("TEST_DB_PORT", 5434),
		Name:     envOr("TEST_DB_NAME", "pr_reviewer_test_db"),
		User:     envOr("TEST_DB_USER", "postgres"),
		Password: envOr("TEST_DB_PASSWORD", "password"),
		SSLMode:  envOr("TEST_DB_SSLMODE", "disable"),
	}
}

func ResetSchema(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	drop := `DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`
	if _, err := pool.Exec(ctx, drop); err != nil {
		return err
	}
	_, _ = pool.Exec(ctx, `DROP TYPE IF EXISTS pr_status_enum`)
	return runSQLFile(ctx, pool, migrationsDir, "001_init.up.sql")
}

type DBLock struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func AcquireDBLock(ctx context.Context, pool *pgxpool.Pool) (*DBLock, error) {
	if _, err := pool.Exec(ctx, `SELECT pg_advisory_lock($1)`, TestDBAdvisoryLockID); err != nil {
		return nil, err
	}
	return &DBLock{ctx: ctx, pool: pool}, nil
}

func (l *DBLock) Release() {
	if l == nil || l.pool == nil {
		return
	}
	_, _ = l.pool.Exec(l.ctx, `SELECT pg_advisory_unlock($1)`, TestDBAdvisoryLockID)
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func runSQLFile(ctx context.Context, pool *pgxpool.Pool, migrationsDir, fileName string) error {
	path := filepath.Join(migrationsDir, fileName)
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", fileName, err)
	}

	for _, stmt := range splitStatements(string(contents)) {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("exec %s: %w", fileName, err)
		}
	}

	return nil
}

func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			statements = append(statements, trimmed)
		}
	}
	return statements
}
