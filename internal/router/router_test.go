package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestRouterCapabilityProtectedRoutes(t *testing.T) {
	t.Parallel()

	db := newRouterTestDB(t)
	admin := &models.User{
		Username: "admin",
		Name:     "Admin",
		Role:     authz.RoleAdmin,
	}
	require.NoError(t, admin.SetPassword("password"))
	require.NoError(t, db.Create(admin).Error)
	operator := &models.User{
		Username: "operator",
		Name:     "Operator",
		Role:     authz.RoleOperator,
	}
	require.NoError(t, operator.SetPassword("password"))
	require.NoError(t, db.Create(operator).Error)
	resetUser := &models.User{
		Username: "reset-user",
		Name:     "Reset User",
		Role:     authz.RoleOperator,
	}
	require.NoError(t, resetUser.SetPassword("password"))
	require.NoError(t, db.Create(resetUser).Error)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode: "test",
		},
		Security: config.SecurityConfig{
			JWTSecret:   "test-secret",
			TokenExpiry: time.Hour,
		},
	}

	jwtAuth := auth.NewJWT(&cfg.Security)
	adminToken, err := jwtAuth.GenerateToken(admin.ID, admin.Username, authz.RoleAdmin)
	require.NoError(t, err)
	operatorToken, err := jwtAuth.GenerateToken(operator.ID, operator.Username, authz.RoleOperator)
	require.NoError(t, err)
	forcedResetToken, err := jwtAuth.GenerateToken(resetUser.ID, resetUser.Username, authz.RoleOperator)
	require.NoError(t, err)
	resetUser.SetForcePasswordReset(true, time.Now().Add(-time.Second))
	require.NoError(t, db.Save(resetUser).Error)

	router := Setup(db, nil, cfg)

	tests := []struct {
		name           string
		method         string
		path           string
		token          string
		expectedStatus int
	}{
		{name: "health remains public", method: http.MethodGet, path: "/health", expectedStatus: http.StatusOK},
		{name: "logout remains public", method: http.MethodPost, path: "/api/v1/auth/logout", expectedStatus: http.StatusOK},
		{name: "user list requires auth", method: http.MethodGet, path: "/api/v1/users", expectedStatus: http.StatusUnauthorized},
		{name: "operator denied user management capability", method: http.MethodGet, path: "/api/v1/users", token: operatorToken, expectedStatus: http.StatusForbidden},
		{name: "admin allowed user management capability", method: http.MethodGet, path: "/api/v1/users", token: adminToken, expectedStatus: http.StatusOK},
		{name: "operator can read self context", method: http.MethodGet, path: "/api/v1/users/me", token: operatorToken, expectedStatus: http.StatusOK},
		{name: "operator can update own password", method: http.MethodPut, path: "/api/v1/users/me/password", token: operatorToken, expectedStatus: http.StatusBadRequest},
		{name: "forced reset blocks normal alerts route", method: http.MethodGet, path: "/api/v1/alerts", token: forcedResetToken, expectedStatus: http.StatusUnauthorized},
		{name: "forced reset still allows self context", method: http.MethodGet, path: "/api/v1/users/me", token: forcedResetToken, expectedStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(recorder, req)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func newRouterTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:router_authz?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	require.NoError(t, db.Exec("DELETE FROM users").Error)

	return db
}
