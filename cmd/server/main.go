package main

import (
	"log"
	"os"
	"time"

	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/database"
	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/escalation"
	"github.com/game-ops/ai-alert-system/internal/notifier"
	"github.com/game-ops/ai-alert-system/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// INGR-04: Validate production config at startup
	if err := config.ValidateProductionConfig(cfg); err != nil {
		log.Fatalf("Invalid production config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize notifier DB reference for email sender
	notifier.SetDB(db)

	// Initialize Redis
	redisClient, err := database.InitRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Setup router
	r := router.Setup(db, redisClient, cfg)

	// Start escalation background checker
	deliverySvc := delivery.NewService(db)
	escalationChecker := escalation.NewChecker(db, deliverySvc)
	go escalationChecker.Run(1*time.Minute, make(chan struct{}))

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
