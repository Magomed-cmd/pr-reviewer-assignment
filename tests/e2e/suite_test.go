package e2e

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	adapterhttp "pr-reviewer-assignment/internal/adapters/input/http"
	adapterdb "pr-reviewer-assignment/internal/adapters/output/database"
	"pr-reviewer-assignment/internal/core/services"
	"pr-reviewer-assignment/internal/infrastructure"
	"pr-reviewer-assignment/internal/infrastructure/database/postgres"
	helpers "pr-reviewer-assignment/tests/shared"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	testSuite  *helpers.E2ESuite
	testPool   *pgxpool.Pool
	testServer *httptest.Server
	testLogger = zap.NewNop()
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	cfg := helpers.LoadTestDBConfig()
	pool, err := postgres.NewConnection(&cfg, testLogger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lock, err := helpers.AcquireDBLock(ctx, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to acquire db lock: %v\n", err)
		os.Exit(1)
	}
	defer lock.Release()

	if err := helpers.ResetSchema(ctx, pool, helpers.MigrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to reset schema: %v\n", err)
		os.Exit(1)
	}

	router := buildRouter(pool)
	server := httptest.NewServer(router)
	defer server.Close()

	testSuite = helpers.NewE2ESuite(server.URL, server.Client(), pool)
	testPool = pool
	testServer = server

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

func resetTables(t testing.TB) {
	t.Helper()
	testSuite.ResetTables(t)
}
