package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	Security SecurityConfig
}
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type ServerConfig struct {
	Port           string
	Mode           string
	AllowedOrigins []string
}

type SecurityConfig struct {
	JWTSecret   string
	TokenExpiry time.Duration
}

func Load() *Config {
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	return &Config{
		Database: LoadDatabaseConfig(),
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8080"),
			Mode:           getEnv("SERVER_MODE", "debug"),
			AllowedOrigins: getEnvAsCSV("ALLOWED_ORIGINS", []string{"http://localhost:*", "http://127.0.0.1:*"}),
		},
		Security: SecurityConfig{
			JWTSecret:   jwtSecret,
			TokenExpiry: getEnvAsDuration("TOKEN_EXPIRY", 24*time.Hour),
		},
	}
}

func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "ai_alert_system"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsCSV(key string, defaultValue []string) []string {
	valueStr := strings.TrimSpace(os.Getenv(key))
	if valueStr == "" {
		return defaultValue
	}

	parts := strings.Split(valueStr, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return defaultValue
	}

	return values
}

// ValidateProductionConfig validates config for production requirements.
// In release mode, it enforces stricter security requirements.
func ValidateProductionConfig(cfg *Config) error {
	if cfg.Server.Mode == "release" {
		if len(cfg.Security.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in release mode, got %d", len(cfg.Security.JWTSecret))
		}
	}
	return nil
}
