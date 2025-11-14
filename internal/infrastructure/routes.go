package infrastructure

import (
	adapterhttp "pr-reviewer-assignment/internal/adapters/input/http"
	"pr-reviewer-assignment/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(logger *zap.Logger, auth *middleware.AuthMiddleware, health *adapterhttp.HealthHandler, team *adapterhttp.TeamHandler, user *adapterhttp.UserHandler, pr *adapterhttp.PullRequestHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	if logger != nil {
		r.Use(middleware.RequestLogger(logger))
	}

	if health != nil {
		health.Register(r)
	}

	registerTeamRoutes(r, auth, team)
	registerUserRoutes(r, auth, user)
	registerPullRequestRoutes(r, auth, pr)

	return r
}

func registerTeamRoutes(r *gin.Engine, auth *middleware.AuthMiddleware, handler *adapterhttp.TeamHandler) {
	if handler == nil {
		return
	}

	adminOnly := r.Group("/team")
	reader := r.Group("/team")
	if auth != nil {
		adminOnly.Use(auth.Require(middleware.RoleAdmin))
		reader.Use(auth.Require(middleware.RoleAdmin, middleware.RoleUser))
	}

	adminOnly.POST("/add", handler.CreateTeam)
	reader.GET("/get", handler.GetTeam)
}

func registerUserRoutes(r *gin.Engine, auth *middleware.AuthMiddleware, handler *adapterhttp.UserHandler) {
	if handler == nil {
		return
	}

	setGroup := r.Group("/users")
	getGroup := r.Group("/users")
	if auth != nil {
		setGroup.Use(auth.Require(middleware.RoleAdmin))
		getGroup.Use(auth.Require(middleware.RoleAdmin, middleware.RoleUser))
	}

	setGroup.POST("/setIsActive", handler.SetActivity)
	getGroup.GET("/getReview", handler.GetReviewerAssignments)
}

func registerPullRequestRoutes(r *gin.Engine, auth *middleware.AuthMiddleware, handler *adapterhttp.PullRequestHandler) {
	if handler == nil {
		return
	}

	group := r.Group("/pullRequest")
	if auth != nil {
		group.Use(auth.Require(middleware.RoleAdmin))
	}

	group.POST("/create", handler.Create)
	group.POST("/merge", handler.Merge)
	group.POST("/reassign", handler.Reassign)
}
