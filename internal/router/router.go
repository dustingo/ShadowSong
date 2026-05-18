package router

import (
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/handlers"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// isAllowedOrigin checks if the origin matches any allowed pattern.
// Patterns ending with "*" are treated as prefix matches.
func isAllowedOrigin(origin string, allowed []string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}
	for _, pattern := range allowed {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "*") {
			if strings.HasPrefix(origin, strings.TrimSuffix(pattern, "*")) {
				return true
			}
			continue
		}
		if origin == pattern {
			return true
		}
	}
	return false
}

func Setup(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	allowedOrigins := cfg.Server.AllowedOrigins
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if isAllowedOrigin(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize handlers
	alertHandler := handlers.NewAlertHandler(db)
	configHandler := handlers.NewConfigHandler(db)
	deliveryHandler := handlers.NewDeliveryHandler(db, delivery.NewService(db))
	webhookHandler := handlers.NewWebhookHandler(db, redisClient)
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	metricsHandler := handlers.NewMetricsHandler(db)
	channelHealthHandler := handlers.NewChannelHealthHandler(db)

	// Initialize auth
	jwtAuth := auth.NewJWT(&cfg.Security)
	userHandler := handlers.NewUserHandler(db, jwtAuth)
	wsHandler := handlers.NewWSHandler(db, jwtAuth, cfg.Server.AllowedOrigins)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Readiness check (OPER-05)
	r.GET("/ready", healthHandler.Readiness)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/logout", userHandler.Logout)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.JWTAuth(jwtAuth, db))
		{
			users.GET("", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.ListUsers)
			users.GET("/me", userHandler.GetCurrentUser)
			users.POST("", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.CreateUser)
			users.PATCH("/:id", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.AdminUpdateUser)
			users.PATCH("/me/profile", userHandler.UpdateOwnProfile)
			users.PUT("/me/password", userHandler.UpdateOwnPassword)
			users.DELETE("/:id", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.DeleteUser)
			users.GET("/audit-logs", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.ListAuditLogs)
		}

		// Alert routes (protected)
		alerts := v1.Group("/alerts")
		alerts.Use(middleware.JWTAuth(jwtAuth, db))
		{
			alerts.GET("", alertHandler.List)
			alerts.GET("/stats", alertHandler.Stats)
			alerts.GET("/active", alertHandler.Active)
			alerts.GET("/:id/deliveries", alertHandler.AlertDeliveries)
			alerts.GET("/:id", alertHandler.Get)
			alerts.POST("/:id/ack", middleware.RequireCapability(authz.CapabilityProcessAlerts), alertHandler.Ack)
			alerts.POST("/:id/quick-silence", middleware.RequireCapability(authz.CapabilityProcessAlerts), alertHandler.QuickSilence)
		}

		// DataSource routes (protected)
		datasources := v1.Group("/datasources")
		datasources.Use(middleware.JWTAuth(jwtAuth, db))
		{
			datasources.GET("", configHandler.ListDataSources)
			datasources.GET("/:id", configHandler.GetDataSource)
			datasources.POST("", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.CreateDataSource)
			datasources.POST("/preview", configHandler.PreviewDataSource)
			datasources.PUT("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.UpdateDataSource)
			datasources.DELETE("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.DeleteDataSource)
			datasources.PATCH("/:id/toggle", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.ToggleDataSource)
		}

		// Channel routes (protected)
		channels := v1.Group("/channels")
		channels.Use(middleware.JWTAuth(jwtAuth, db))
		{
			channels.GET("", configHandler.ListChannels)
			channels.GET("/:id", configHandler.GetChannel)
			channels.POST("", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.CreateChannel)
			channels.PUT("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.UpdateChannel)
			channels.DELETE("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.DeleteChannel)
			channels.PATCH("/:id/toggle", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.ToggleChannel)
			channels.POST("/:id/test", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.TestChannel)
				channels.GET("/:id/health", middleware.RequireCapability(authz.CapabilityViewConfig), channelHealthHandler.GetChannelHealth)
		}

		// RouteRule routes (protected)
		routes := v1.Group("/routes")
		routes.Use(middleware.JWTAuth(jwtAuth, db))
		{
			routes.GET("", configHandler.ListRouteRules)
			routes.GET("/:id", configHandler.GetRouteRule)
			routes.POST("", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.CreateRouteRule)
			routes.PUT("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.UpdateRouteRule)
			routes.DELETE("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.DeleteRouteRule)
			routes.POST("/reorder", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.ReorderRouteRules)
		}

		// SilenceRule routes (protected)
		silences := v1.Group("/silences")
		silences.Use(middleware.JWTAuth(jwtAuth, db))
		{
			silences.GET("", configHandler.ListSilenceRules)
			silences.GET("/:id", configHandler.GetSilenceRule)
			silences.POST("", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.CreateSilenceRule)
			silences.PUT("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.UpdateSilenceRule)
			silences.DELETE("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.DeleteSilenceRule)
			silences.POST("/from-alert/:alertId", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.CreateSilenceFromAlert)
		}

		// Delivery routes (protected)
		deliveries := v1.Group("/deliveries")
		deliveries.Use(middleware.JWTAuth(jwtAuth, db))
		{
			deliveries.GET("", middleware.RequireCapability(authz.CapabilityViewConfig), deliveryHandler.List)
			deliveries.GET("/:id", middleware.RequireCapability(authz.CapabilityViewConfig), deliveryHandler.Get)
			deliveries.POST("/:id/retry", middleware.RequireCapability(authz.CapabilityProcessAlerts), deliveryHandler.Retry)
			deliveries.POST("/:id/replay", middleware.RequireCapability(authz.CapabilityProcessAlerts), deliveryHandler.Replay)
		}

		// Metrics routes (protected) - OPER-03
		v1.GET("/metrics", middleware.JWTAuth(jwtAuth, db), middleware.RequireCapability(authz.CapabilityViewConfig), metricsHandler.GetMetrics)

	}

	// Webhook routes (INGR-01: size limit, INGR-02: rate limit)
	webhook := r.Group("/webhook")
	webhook.Use(middleware.RequestSizeLimit(1 * 1024 * 1024)) // 1MB default
	webhook.Use(middleware.RateLimit(
		middleware.NewInMemoryRateLimiter(1000, time.Minute), // 1000 req/min per source
		func(c *gin.Context) string { return c.Param("source_name") },
	))
	{
		webhook.POST("/:source_name", webhookHandler.HandleWebhook)
		webhook.POST("/test-template", webhookHandler.TestInputTemplate)
	}

	// WebSocket routes
	r.GET("/ws/alerts", wsHandler.HandleAlerts)

	return r
}
