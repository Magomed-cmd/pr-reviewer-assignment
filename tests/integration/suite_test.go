package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	adapterhttp "pr-reviewer-assignment/internal/adapters/input/http"
	adapterdb "pr-reviewer-assignment/internal/adapters/output/database"
	"pr-reviewer-assignment/internal/config"
	"pr-reviewer-assignment/internal/core/services"
	"pr-reviewer-assignment/internal/infrastructure"
	"pr-reviewer-assignment/internal/infrastructure/database/postgres"
	"pr-reviewer-assignment/tests/integration/helpers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	testPool      *pgxpool.Pool
	testRouter    *gin.Engine
	testSuite     *helpers.Suite
	testLogger    = zap.NewNop()
	migrationsDir = filepath.Join("..", "..", "migrations")
)

const (
	testTeamCore     = "core-team"
	testTeamPlatform = "platform-team"
	testAuthorID     = "author-1"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	cfg := loadTestDBConfig()
	pool, err := postgres.NewConnection(&cfg, testLogger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset schema: %v\n", err)
		os.Exit(1)
	}

	testPool = pool
	testRouter = buildRouter(pool)
	testSuite = helpers.NewSuite(testRouter, testPool)

	code := m.Run()
	os.Exit(code)
}

func buildRouter(pool *pgxpool.Pool) *gin.Engine {
	txManager := postgres.NewTransactionManager(pool, testLogger)

	teamRepo := adapterdb.NewTeamRepository(pool, testLogger)
	userRepo := adapterdb.NewUserRepository(pool, testLogger)
	prRepo := adapterdb.NewPullRequestRepository(pool, testLogger)

	teamService := services.NewTeamService(teamRepo, userRepo, testLogger, txManager)
	userService := services.NewUserService(userRepo, prRepo, testLogger)
	prService := services.NewPullRequestService(prRepo, userRepo, teamRepo, testLogger, txManager)

	healthHandler := adapterhttp.NewHealthHandler()
	teamHandler := adapterhttp.NewTeamHandler(teamService, testLogger)
	userHandler := adapterhttp.NewUserHandler(userService, testLogger)
	prHandler := adapterhttp.NewPullRequestHandler(prService, testLogger)

	return infrastructure.NewRouter(testLogger, healthHandler, teamHandler, userHandler, prHandler)
}

func loadTestDBConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     envOr("TEST_DB_HOST", "127.0.0.1"),
		Port:     envOrInt("TEST_DB_PORT", 5434),
		Name:     envOr("TEST_DB_NAME", "pr_reviewer_test_db"),
		User:     envOr("TEST_DB_USER", "postgres"),
		Password: envOr("TEST_DB_PASSWORD", "password"),
		SSLMode:  envOr("TEST_DB_SSLMODE", "disable"),
	}
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

func resetSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if err := runSQLFile(ctx, pool, "001_init.down.sql"); err != nil {
		return err
	}
	if err := runSQLFile(ctx, pool, "001_init.up.sql"); err != nil {
		return err
	}
	return nil
}

func runSQLFile(ctx context.Context, pool *pgxpool.Pool, fileName string) error {
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

func resetTables(t testing.TB) {
	t.Helper()
	testSuite.ResetTables(t)
}
