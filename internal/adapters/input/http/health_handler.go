package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Register(router *gin.Engine) {
	router.GET("/health", h.Ping)
	router.HEAD("/health", h.Ping)
}

func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
