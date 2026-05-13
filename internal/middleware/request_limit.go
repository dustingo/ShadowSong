package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimit returns a middleware that rejects requests with body size exceeding maxBytes.
func RequestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("request body too large: max %d bytes", maxBytes),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}