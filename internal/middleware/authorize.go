package middleware

import (
	"net/http"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/gin-gonic/gin"
)

// RequireCapability is the single route-facing authz seam for protected endpoints.
func RequireCapability(capability authz.Capability) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := GetPrincipal(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !authz.Can(principal.Role, capability) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
