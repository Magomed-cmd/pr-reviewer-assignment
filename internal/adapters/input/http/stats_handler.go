package http

import (
	"net/http"

	serviceports "pr-reviewer-assignment/internal/core/ports/services"
	"pr-reviewer-assignment/internal/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StatsHandler struct {
	service serviceports.StatsService
	logger  *zap.Logger
}

func NewStatsHandler(service serviceports.StatsService, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{
		service: service,
		logger:  logger,
	}
}

func (h *StatsHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/stats", h.GetStats)
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Warn("Get stats failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.StatsResponse{
		Stats: dto.StatsDTO{
			Teams:        stats.Teams,
			Users:        stats.Users,
			PullRequests: stats.PullRequests,
			Assignments:  stats.Assignments,
		},
	})
}
