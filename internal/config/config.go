package config

import (
	"log/slog"
	"os"
	"strconv"
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
	Port string
	Mode string
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
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
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
