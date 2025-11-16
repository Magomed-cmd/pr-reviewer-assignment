package infrastructure

import (
	adapterhttp "pr-reviewer-assignment/internal/adapters/input/http"
	"pr-reviewer-assignment/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, health *adapterhttp.HealthHandler, team *adapterhttp.TeamHandler, user *adapterhttp.UserHandler, pr *adapterhttp.PullRequestHandler, stats *adapterhttp.StatsHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if logger != nil {
		r.Use(middleware.RequestLogger(logger))
	}

	if health != nil {
		health.Register(r)
	}

	registerTeamRoutes(r, team)
	registerUserRoutes(r, user)
	registerPullRequestRoutes(r, pr)
	registerStatsRoutes(r, stats)

	return r
}

func registerTeamRoutes(r *gin.Engine, handler *adapterhttp.TeamHandler) {
	if handler == nil {
		return
	}

	group := r.Group("/team")

	group.POST("/add", handler.CreateTeam)
	group.GET("/get", handler.GetTeam)
}

func registerUserRoutes(r *gin.Engine, handler *adapterhttp.UserHandler) {
	if handler == nil {
		return
	}

	group := r.Group("/users")
	group.POST("/setIsActive", handler.SetActivity)
	group.GET("/getReview", handler.GetReviewerAssignments)
}

func registerPullRequestRoutes(r *gin.Engine, handler *adapterhttp.PullRequestHandler) {
	if handler == nil {
		return
	}

	group := r.Group("/pullRequest")

	group.POST("/create", handler.Create)
	group.POST("/merge", handler.Merge)
	group.POST("/reassign", handler.Reassign)
}

func registerStatsRoutes(r *gin.Engine, handler *adapterhttp.StatsHandler) {
	if handler == nil {
		return
	}
	handler.RegisterRoutes(r)
}
