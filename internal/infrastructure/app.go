package infrastructure

import (
	"fmt"
	"net/http"

	adapterhttp "pr-reviewer-assignment/internal/adapters/input/http"
	adapterdb "pr-reviewer-assignment/internal/adapters/output/database"
	"pr-reviewer-assignment/internal/config"
	"pr-reviewer-assignment/internal/core/services"
	"pr-reviewer-assignment/internal/infrastructure/database/postgres"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	httpServer *http.Server
	dbPool     *pgxpool.Pool
}

func NewApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	gin.SetMode(cfg.Server.Mode)

	dbPool, err := postgres.NewConnection(&cfg.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	txManager := postgres.NewTransactionManager(dbPool, logger)

	teamRepo := adapterdb.NewTeamRepository(dbPool, logger)
	userRepo := adapterdb.NewUserRepository(dbPool, logger)
	prRepo := adapterdb.NewPullRequestRepository(dbPool, logger)

	teamService := services.NewTeamService(teamRepo, userRepo, logger, txManager)
	userService := services.NewUserService(userRepo, prRepo, logger)
	prService := services.NewPullRequestService(prRepo, userRepo, teamRepo, logger, txManager)

	healthHandler := adapterhttp.NewHealthHandler()
	teamHandler := adapterhttp.NewTeamHandler(teamService, logger)
	userHandler := adapterhttp.NewUserHandler(userService, logger)
	prHandler := adapterhttp.NewPullRequestHandler(prService, logger)

	router := NewRouter(logger, healthHandler, teamHandler, userHandler, prHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		httpServer: server,
		dbPool:     dbPool,
	}, nil
}

func (a *App) HTTPServer() *http.Server {
	return a.httpServer
}

func (a *App) Close() {
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}
