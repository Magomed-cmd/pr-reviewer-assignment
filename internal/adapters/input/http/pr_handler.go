package http

import (
	"net/http"
	"strings"
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	"pr-reviewer-assignment/internal/core/mappers"
	serviceports "pr-reviewer-assignment/internal/core/ports/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PullRequestHandler struct {
	service serviceports.PullRequestService
	logger  *zap.Logger
}

func NewPullRequestHandler(service serviceports.PullRequestService, logger *zap.Logger) *PullRequestHandler {
	return &PullRequestHandler{service: service, logger: logger}
}

func (h *PullRequestHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/pullRequest/create", h.Create)
	router.POST("/pullRequest/merge", h.Merge)
	router.POST("/pullRequest/reassign", h.Reassign)
}

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

func (h *PullRequestHandler) Create(c *gin.Context) {
	var payload createPRRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid payload")
		return
	}

	if payload.PullRequestID == "" || payload.PullRequestName == "" || payload.AuthorID == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	pr := entities.NewPullRequest(payload.PullRequestID, payload.PullRequestName, payload.AuthorID, time.Time{})
	created, err := h.service.CreatePullRequest(c.Request.Context(), pr)
	if err != nil {
		h.logger.Warn("Create PR failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": mappers.PullRequestToDTO(created)})
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func (h *PullRequestHandler) Merge(c *gin.Context) {
	var payload mergePRRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid payload")
		return
	}

	if strings.TrimSpace(payload.PullRequestID) == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id is required")
		return
	}

	pr, err := h.service.MergePullRequest(c.Request.Context(), payload.PullRequestID)
	if err != nil {
		h.logger.Warn("Merge PR failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": mappers.PullRequestToDTO(pr)})
}

type reassignRequest struct {
	PullRequestID    string `json:"pull_request_id"`
	OldUserID        string `json:"old_user_id"`
	LegacyReviewerID string `json:"old_reviewer_id"`
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var payload reassignRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "invalid payload")
		return
	}

	oldUserID := payload.OldUserID
	if oldUserID == "" {
		oldUserID = payload.LegacyReviewerID
	}

	if payload.PullRequestID == "" || oldUserID == "" {
		respondError(c, http.StatusBadRequest, errorCodeBadRequest, "pull_request_id and old_user_id are required")
		return
	}

	pr, replacedBy, err := h.service.ReassignReviewer(c.Request.Context(), payload.PullRequestID, oldUserID)
	if err != nil {
		h.logger.Warn("Reassign reviewer failed", zap.Error(err))
		handleServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          mappers.PullRequestToDTO(pr),
		"replaced_by": replacedBy,
	})
}
