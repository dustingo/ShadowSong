package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad_CurrentBaselineEnv(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "loads minimum non ai config",
			env: map[string]string{
				"JWT_SECRET":   "test-secret",
				"DB_HOST":      "127.0.0.1",
				"DB_PORT":      "5433",
				"DB_USER":      "tester",
				"DB_PASSWORD":  "secret",
				"DB_NAME":      "alerts_test",
				"DB_SSLMODE":   "disable",
				"REDIS_HOST":   "127.0.0.1",
				"REDIS_PORT":   "6380",
				"REDIS_DB":     "2",
				"SERVER_PORT":  "18080",
				"SERVER_MODE":  "release",
				"TOKEN_EXPIRY": "2h",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{
				"JWT_SECRET",
				"DB_HOST",
				"DB_PORT",
				"DB_USER",
				"DB_PASSWORD",
				"DB_NAME",
				"DB_SSLMODE",
				"REDIS_HOST",
				"REDIS_PORT",
				"REDIS_PASSWORD",
				"REDIS_DB",
				"SERVER_PORT",
				"SERVER_MODE",
				"TOKEN_EXPIRY",
				"OPENAI_API_KEY",
				"OPENAI_API_BASE",
				"AI_MODEL",
				"AI_TIMEOUT",
			} {
				t.Setenv(key, "")
			}

			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			cfg := Load()

			assert.NotNil(t, cfg)
			assert.Equal(t, "127.0.0.1", cfg.Database.Host)
			assert.Equal(t, 5433, cfg.Database.Port)
			assert.Equal(t, "tester", cfg.Database.User)
			assert.Equal(t, "secret", cfg.Database.Password)
			assert.Equal(t, "alerts_test", cfg.Database.DBName)
			assert.Equal(t, "disable", cfg.Database.SSLMode)
			assert.Equal(t, "127.0.0.1", cfg.Redis.Host)
			assert.Equal(t, 6380, cfg.Redis.Port)
			assert.Equal(t, 2, cfg.Redis.DB)
			assert.Equal(t, "18080", cfg.Server.Port)
			assert.Equal(t, "release", cfg.Server.Mode)
			assert.Equal(t, "test-secret", cfg.Security.JWTSecret)
			assert.Equal(t, 2*time.Hour, cfg.Security.TokenExpiry)
			assert.Empty(t, getenvUnsafe("OPENAI_API_KEY"))
			assert.Empty(t, getenvUnsafe("OPENAI_API_BASE"))
			assert.Empty(t, getenvUnsafe("AI_MODEL"))
			assert.Empty(t, getenvUnsafe("AI_TIMEOUT"))
		})
	}
}

func getenvUnsafe(key string) string {
	return getEnv(key, "")
}
