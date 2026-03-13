package router

import (
	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/config"
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
		// 测试阶段暂时只允许本机
		c.Writer.Header().Set("Access-Control-Allow-Origin", "127.0.0.1")
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
	aiHandler := handlers.NewAIHandler(db, cfg)
	wsHandler := handlers.NewWSHandler(db)
	webhookHandler := handlers.NewWebhookHandler(db, redisClient)

	// Initialize auth
	jwtAuth := auth.NewJWT(&cfg.Security)
	userHandler := handlers.NewUserHandler(db, jwtAuth)

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
		users.Use(middleware.JWTAuth(jwtAuth))
		{
			users.GET("", middleware.RequireRole("admin"), userHandler.ListUsers)
			users.GET("/me", userHandler.GetCurrentUser)
			users.POST("", middleware.RequireRole("admin"), userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireRole("admin"), userHandler.DeleteUser)
		}

		// Alert routes (protected)
		alerts := v1.Group("/alerts")
		alerts.Use(middleware.JWTAuth(jwtAuth))
		{
			alerts.GET("", alertHandler.List)
			alerts.GET("/stats", alertHandler.Stats)
			alerts.GET("/active", alertHandler.Active)
			alerts.GET("/:id", alertHandler.Get)
			alerts.POST("/:id/ack", alertHandler.Ack)
			alerts.POST("/:id/quick-silence", alertHandler.QuickSilence)
		}

		// DataSource routes (protected)
		datasources := v1.Group("/datasources")
		datasources.Use(middleware.JWTAuth(jwtAuth))
		{
			datasources.GET("", configHandler.ListDataSources)
			datasources.GET("/:id", configHandler.GetDataSource)
			datasources.POST("", configHandler.CreateDataSource)
			datasources.PUT("/:id", configHandler.UpdateDataSource)
			datasources.DELETE("/:id", configHandler.DeleteDataSource)
			datasources.PATCH("/:id/toggle", configHandler.ToggleDataSource)
		}

		// Channel routes (protected)
		channels := v1.Group("/channels")
		channels.Use(middleware.JWTAuth(jwtAuth))
		{
			channels.GET("", configHandler.ListChannels)
			channels.GET("/:id", configHandler.GetChannel)
			channels.POST("", configHandler.CreateChannel)
			channels.PUT("/:id", configHandler.UpdateChannel)
			channels.DELETE("/:id", configHandler.DeleteChannel)
			channels.PATCH("/:id/toggle", configHandler.ToggleChannel)
			channels.POST("/:id/test", configHandler.TestChannel)
		}

		// RouteRule routes (protected)
		routes := v1.Group("/routes")
		routes.Use(middleware.JWTAuth(jwtAuth))
		{
			routes.GET("", configHandler.ListRouteRules)
			routes.GET("/:id", configHandler.GetRouteRule)
			routes.POST("", configHandler.CreateRouteRule)
			routes.PUT("/:id", configHandler.UpdateRouteRule)
			routes.DELETE("/:id", configHandler.DeleteRouteRule)
			routes.POST("/reorder", configHandler.ReorderRouteRules)
		}

		// SilenceRule routes (protected)
		silences := v1.Group("/silences")
		silences.Use(middleware.JWTAuth(jwtAuth))
		{
			silences.GET("", configHandler.ListSilenceRules)
			silences.GET("/:id", configHandler.GetSilenceRule)
			silences.POST("", configHandler.CreateSilenceRule)
			silences.PUT("/:id", configHandler.UpdateSilenceRule)
			silences.DELETE("/:id", configHandler.DeleteSilenceRule)
			silences.POST("/from-alert/:alertId", configHandler.CreateSilenceFromAlert)
		}

		// OnDuty routes (protected)
		onduty := v1.Group("/onduty")
		onduty.Use(middleware.JWTAuth(jwtAuth))
		{
			onduty.GET("", configHandler.ListOnDuty)
			onduty.GET("/current", configHandler.CurrentOnDuty)
			onduty.GET("/:id", configHandler.GetOnDuty)
			onduty.POST("", configHandler.CreateOnDuty)
			onduty.PUT("/:id", configHandler.UpdateOnDuty)
			onduty.DELETE("/:id", configHandler.DeleteOnDuty)
		}

		// AI routes (protected)
		ai := v1.Group("/ai")
		ai.Use(middleware.JWTAuth(jwtAuth))
		{
			ai.POST("/chat", aiHandler.Chat)
			ai.GET("/suggestions/:alertId", aiHandler.Suggestions)
			ai.GET("/logs", aiHandler.ListLogs)
			ai.PATCH("/logs/:id/accuracy", aiHandler.MarkAccuracy)
			ai.DELETE("/logs/:id", aiHandler.DeleteLog)
			ai.GET("/silence-recommendations", aiHandler.ListRecommendations)
			ai.POST("/silence-recommendations/:id/adopt", aiHandler.AdoptRecommendation)
			ai.POST("/silence-recommendations/:id/ignore", aiHandler.IgnoreRecommendation)
		}
	}

	// Webhook routes
	webhook := r.Group("/webhook")
	{
		webhook.POST("/:source_name", webhookHandler.HandleWebhook)
		webhook.POST("/test-template", webhookHandler.TestInputTemplate)
	}

	// WebSocket routes
	r.GET("/ws/alerts", func(c *gin.Context) {
		wsHandler.HandleAlerts(c)
	})

	return r
}
