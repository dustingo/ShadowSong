package router

import (
	"strings"

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

func Setup(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "http://localhost") {
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
		}

		// Alert routes (protected)
		alerts := v1.Group("/alerts")
		alerts.Use(middleware.JWTAuth(jwtAuth, db))
		{
			alerts.GET("", alertHandler.List)
			alerts.GET("/stats", alertHandler.Stats)
			alerts.GET("/active", alertHandler.Active)
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

		// OnDuty routes (protected)
		onduty := v1.Group("/onduty")
		onduty.Use(middleware.JWTAuth(jwtAuth, db))
		{
			onduty.GET("", configHandler.ListOnDuty)
			onduty.GET("/current", configHandler.CurrentOnDuty)
			onduty.GET("/:id", configHandler.GetOnDuty)
			onduty.POST("", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.CreateOnDuty)
			onduty.PUT("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.UpdateOnDuty)
			onduty.DELETE("/:id", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.DeleteOnDuty)
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

	}

	// Webhook routes
	webhook := r.Group("/webhook")
	{
		webhook.POST("/:source_name", webhookHandler.HandleWebhook)
		webhook.POST("/test-template", webhookHandler.TestInputTemplate)
	}

	// WebSocket routes
	r.GET("/ws/alerts", wsHandler.HandleAlerts)

	return r
}
