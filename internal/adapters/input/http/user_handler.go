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

type UserHandler struct {
	service serviceports.UserService
	logger  *zap.Logger
}

func NewUserHandler(service serviceports.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/users/setIsActive", h.SetActivity)
	router.GET("/users/getReview", h.GetReviewerAssignments)
}

type setActivityRequest struct {
	UserID   string `json:"user_id"`
	IsActive *bool  `json:"is_active"`
}

func (h *UserHandler) SetActivity(c *gin.Context) {
	var payload setActivityRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid payload")
		return
	}

	if strings.TrimSpace(payload.UserID) == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "user_id is required")
		return
	}

	if payload.IsActive == nil {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "is_active is required")
		return
	}

	user, err := h.service.SetActivity(c.Request.Context(), payload.UserID, *payload.IsActive)
	if err != nil {
		h.logger.Warn("SetActivity failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": mappers.UserToDTO(user)})
}

func (h *UserHandler) GetReviewerAssignments(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	if userID == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "user_id is required")
		return
	}

	prs, err := h.service.GetReviewerAssignments(c.Request.Context(), userID)
	if err != nil {
		h.logger.Warn("GetReviewerAssignments failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	response := dto.ListPullRequestsResponse{
		UserID:       userID,
		PullRequests: mappers.PullRequestsToShortDTO(prs),
	}

	c.JSON(http.StatusOK, response)
}
