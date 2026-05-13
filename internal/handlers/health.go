package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
	}
}

// Readiness checks if PostgreSQL and Redis are ready (OPER-05).
// Returns 200 if both are healthy, 503 otherwise.
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check PostgreSQL
	var pgHealthy bool
	sqlDB, err := h.db.DB()
	if err == nil {
		pgHealthy = sqlDB.PingContext(ctx) == nil
	}

	// Check Redis
	var redisHealthy bool
	if h.redisClient != nil {
		redisHealthy = h.redisClient.Ping(ctx).Err() == nil
	}

	if pgHealthy && redisHealthy {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"checks": gin.H{
				"postgresql": "healthy",
				"redis":      "healthy",
			},
		})
		return
	}

	// Return 503 with details
	checks := gin.H{}
	if !pgHealthy {
		checks["postgresql"] = "unhealthy"
	} else {
		checks["postgresql"] = "healthy"
	}
	if !redisHealthy {
		checks["redis"] = "unhealthy"
	} else {
		checks["redis"] = "healthy"
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{
		"status": "not_ready",
		"checks": checks,
	})
}