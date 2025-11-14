package http

import (
	"net/http"
	"strings"

	"pr-reviewer-assignment/internal/core/mappers"
	serviceports "pr-reviewer-assignment/internal/core/ports/services"
	"pr-reviewer-assignment/internal/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TeamHandler struct {
	service serviceports.TeamService
	logger  *zap.Logger
}

func NewTeamHandler(service serviceports.TeamService, logger *zap.Logger) *TeamHandler {
	return &TeamHandler{service: service, logger: logger}
}

func (h *TeamHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/team/add", h.CreateTeam)
	router.GET("/team/get", h.GetTeam)
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var payload dto.TeamDTO
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid team payload")
		return
	}

	teamName := strings.TrimSpace(payload.TeamName)
	if teamName == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}

	members := mappers.TeamMembersFromDTO(teamName, payload.Members)
	team, err := h.service.CreateTeam(c.Request.Context(), teamName, members)
	if err != nil {
		h.logger.Warn("CreateTeam failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": mappers.TeamToDTO(team)})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := strings.TrimSpace(c.Query("team_name"))
	if teamName == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "team_name is required")
		return
	}

	team, err := h.service.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		h.logger.Warn("GetTeam failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, mappers.TeamToDTO(team))
}
