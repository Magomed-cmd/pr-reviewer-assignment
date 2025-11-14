package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type AuthMiddleware struct {
	logger     *zap.Logger
	tokenRoles map[string]string
}

func NewAuthMiddleware(adminToken, userToken string, logger *zap.Logger) *AuthMiddleware {
	if logger == nil {
		logger = zap.NewNop()
	}

	tokens := make(map[string]string)
	if adminToken != "" {
		tokens[adminToken] = RoleAdmin
	}
	if userToken != "" {
		tokens[userToken] = RoleUser
	}

	if len(tokens) == 0 {
		return nil
	}

	return &AuthMiddleware{
		logger:     logger,
		tokenRoles: tokens,
	}
}

func (m *AuthMiddleware) Require(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		if m == nil {
			c.Next()
			return
		}

		token := extractToken(c.GetHeader("Authorization"))
		if token == "" {
			abortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")
			return
		}

		role, ok := m.tokenRoles[token]
		if !ok {
			abortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			return
		}

		if len(allowed) > 0 {
			if _, ok := allowed[role]; !ok {
				abortWithError(c, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}
		}

		c.Set("role", role)
		c.Next()
	}
}

func extractToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.Fields(header)
	if len(parts) == 0 {
		return ""
	}

	if strings.EqualFold(parts[0], "Bearer") {
		if len(parts) > 1 {
			return parts[1]
		}
		return ""
	}

	return parts[0]
}

func abortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
