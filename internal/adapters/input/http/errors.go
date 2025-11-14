package http

import (
	"errors"
	"net/http"

	domainErrors "pr-reviewer-assignment/internal/core/domain/errors"
	"pr-reviewer-assignment/internal/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	errorCodeBadRequest   = "BAD_REQUEST"
	errorCodeInternal     = "INTERNAL_ERROR"
	errorCodeUnauthorized = "UNAUTHORIZED"
)

func respondError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, dto.ErrorResponse{
		Error: dto.ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func handleServiceError(c *gin.Context, logger *zap.Logger, err error) {
	var dErr domainErrors.DomainError
	if errors.As(err, &dErr) {
		switch dErr.Code() {
		case domainErrors.ErrorCodeTeamExists:
			respondError(c, http.StatusBadRequest, string(dErr.Code()), dErr.Message())
		case domainErrors.ErrorCodePRExists:
			respondError(c, http.StatusConflict, string(dErr.Code()), dErr.Message())
		case domainErrors.ErrorCodePRMerged,
			domainErrors.ErrorCodeNotAssigned,
			domainErrors.ErrorCodeNoCandidate:
			respondError(c, http.StatusConflict, string(dErr.Code()), dErr.Message())
		case domainErrors.ErrorCodeNotFound:
			respondError(c, http.StatusNotFound, string(dErr.Code()), dErr.Message())
		default:
			logger.Warn("Unhandled domain error", zap.String("code", string(dErr.Code())), zap.String("message", dErr.Message()))
			respondError(c, http.StatusInternalServerError, errorCodeInternal, "internal server error")
		}
		return
	}

	logger.Error("Service error", zap.Error(err))
	respondError(c, http.StatusInternalServerError, errorCodeInternal, "internal server error")
}
