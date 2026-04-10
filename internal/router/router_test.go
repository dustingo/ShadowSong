package router

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSetup_RoutesWithoutAIRuntime(t *testing.T) {
	tests := []struct {
		name          string
		requiredPaths []string
		forbiddenPath string
	}{
		{
			name: "keeps core routes and removes ai routes",
			requiredPaths: []string{
				"/health",
				"/api/v1/alerts",
				"/api/v1/alerts/stats",
				"/api/v1/datasources/preview",
				"/webhook/test-template",
			},
			forbiddenPath: "/api/v1/ai",
		},
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Security: config.SecurityConfig{
			JWTSecret:   "test-secret",
			TokenExpiry: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Setup(nil, nil, cfg)
			routes := r.Routes()

			registered := make(map[string]bool, len(routes))
			for _, route := range routes {
				registered[route.Path] = true
			}

			for _, path := range tt.requiredPaths {
				assert.Truef(t, registered[path], "expected route %s to be registered", path)
			}

			for path := range registered {
				assert.NotContains(t, path, tt.forbiddenPath)
			}
		})
	}
}
