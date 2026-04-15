package router

import (
	"bytes"
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
	alert := models.Alert{
		AlertID:     "alert-1",
		Source:      "system",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
		Status:      "firing",
	}
	require.NoError(t, db.Create(&alert).Error)
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
	viewer := &models.User{
		Username: "viewer",
		Name:     "Viewer",
		Role:     authz.RoleViewer,
	}
	require.NoError(t, viewer.SetPassword("password"))
	require.NoError(t, db.Create(viewer).Error)
	passwordUser := &models.User{
		Username: "password-user",
		Name:     "Password User",
		Role:     authz.RoleOperator,
	}
	require.NoError(t, passwordUser.SetPassword("password"))
	require.NoError(t, db.Create(passwordUser).Error)
	disabledUser := &models.User{
		Username: "disabled-user",
		Name:     "Disabled User",
		Role:     authz.RoleOperator,
	}
	require.NoError(t, disabledUser.SetPassword("password"))
	disabledUser.SetDisabled(true, time.Now().Add(-time.Second))
	require.NoError(t, db.Create(disabledUser).Error)
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
	viewerToken, err := jwtAuth.GenerateToken(viewer.ID, viewer.Username, authz.RoleViewer)
	require.NoError(t, err)
	passwordUserToken, err := jwtAuth.GenerateToken(passwordUser.ID, passwordUser.Username, authz.RoleOperator)
	require.NoError(t, err)
	disabledToken, err := jwtAuth.GenerateToken(disabledUser.ID, disabledUser.Username, authz.RoleOperator)
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
		body           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{name: "health remains public", method: http.MethodGet, path: "/health", expectedStatus: http.StatusOK},
		{name: "logout remains public", method: http.MethodPost, path: "/api/v1/auth/logout", expectedStatus: http.StatusOK},
		{name: "user list requires auth", method: http.MethodGet, path: "/api/v1/users", expectedStatus: http.StatusUnauthorized},
		{name: "operator denied user management capability", method: http.MethodGet, path: "/api/v1/users", token: operatorToken, expectedStatus: http.StatusForbidden},
		{name: "admin allowed user management capability", method: http.MethodGet, path: "/api/v1/users", token: adminToken, expectedStatus: http.StatusOK},
		{name: "disabled user rejected before capability-protected user management route", method: http.MethodGet, path: "/api/v1/users", token: disabledToken, expectedStatus: http.StatusUnauthorized, expectedBody: `{"error":"account disabled"}`},
		{name: "operator can read self context", method: http.MethodGet, path: "/api/v1/users/me", token: operatorToken, expectedStatus: http.StatusOK},
		{name: "operator can update own password", method: http.MethodPut, path: "/api/v1/users/me/password", body: `{"password":"new-password"}`, token: passwordUserToken, expectedStatus: http.StatusOK},
		{name: "forced reset blocks capability-protected alert ack route", method: http.MethodPost, path: "/api/v1/alerts/alert-1/ack", body: `{"comment":"ack"}`, token: forcedResetToken, expectedStatus: http.StatusUnauthorized, expectedBody: `{"error":"password reset required"}`},
		{name: "forced reset still allows self context", method: http.MethodGet, path: "/api/v1/users/me", token: forcedResetToken, expectedStatus: http.StatusOK},
		{name: "forced reset still allows self-service password change", method: http.MethodPut, path: "/api/v1/users/me/password", body: `{"password":"rotated-password"}`, token: forcedResetToken, expectedStatus: http.StatusOK},
		{name: "alerts ack requires auth", method: http.MethodPost, path: "/api/v1/alerts/alert-1/ack", body: `{"comment":"ack"}`, expectedStatus: http.StatusUnauthorized},
		{name: "operator can ack alerts", method: http.MethodPost, path: "/api/v1/alerts/alert-1/ack", body: `{"comment":"ack"}`, token: operatorToken, expectedStatus: http.StatusOK},
		{name: "viewer denied ack alerts", method: http.MethodPost, path: "/api/v1/alerts/alert-1/ack", body: `{"comment":"ack"}`, token: viewerToken, expectedStatus: http.StatusForbidden, expectedBody: `{"error":"insufficient permissions"}`},
		{name: "operator can quick silence alerts", method: http.MethodPost, path: "/api/v1/alerts/alert-1/quick-silence", body: `{"duration":3600}`, token: operatorToken, expectedStatus: http.StatusOK},
		{name: "viewer denied quick silence alerts", method: http.MethodPost, path: "/api/v1/alerts/alert-1/quick-silence", body: `{"duration":3600}`, token: viewerToken, expectedStatus: http.StatusForbidden, expectedBody: `{"error":"insufficient permissions"}`},
		{name: "viewer can read config", method: http.MethodGet, path: "/api/v1/datasources", token: viewerToken, expectedStatus: http.StatusOK},
		{name: "operator denied config writes", method: http.MethodPost, path: "/api/v1/datasources", body: `{"name":"ops-ds","display_name":"Ops DS","input_template":"{{ . }}","output_template":"{{ . }}"}`, token: operatorToken, expectedStatus: http.StatusForbidden, expectedBody: `{"error":"insufficient permissions"}`},
		{name: "viewer denied config writes", method: http.MethodPost, path: "/api/v1/datasources", body: `{"name":"viewer-ds","display_name":"Viewer DS","input_template":"{{ . }}","output_template":"{{ . }}"}`, token: viewerToken, expectedStatus: http.StatusForbidden, expectedBody: `{"error":"insufficient permissions"}`},
		{name: "admin can create datasource", method: http.MethodPost, path: "/api/v1/datasources", body: `{"name":"admin-ds","display_name":"Admin DS","input_template":"{{ . }}","output_template":"{{ . }}"}`, token: adminToken, expectedStatus: http.StatusOK},
		{name: "operator denied on-duty writes", method: http.MethodPost, path: "/api/v1/onduty", body: `{"user_id":"u1","user_name":"ops","channel_id":1,"start_time":"2026-04-12T00:00:00Z","end_time":"2026-04-12T08:00:00Z"}`, token: operatorToken, expectedStatus: http.StatusForbidden, expectedBody: `{"error":"insufficient permissions"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			router.ServeHTTP(recorder, req)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func newRouterTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:router_authz?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Alert{},
		&models.AuditLog{},
		&models.DataSource{},
		&models.Channel{},
		&models.RouteRule{},
		&models.SilenceRule{},
		&models.OnDuty{},
	))
	require.NoError(t, db.Exec("DELETE FROM users").Error)
	require.NoError(t, db.Exec("DELETE FROM alerts").Error)
	require.NoError(t, db.Exec("DELETE FROM audit_logs").Error)
	require.NoError(t, db.Exec("DELETE FROM data_sources").Error)
	require.NoError(t, db.Exec("DELETE FROM channels").Error)
	require.NoError(t, db.Exec("DELETE FROM route_rules").Error)
	require.NoError(t, db.Exec("DELETE FROM silence_rules").Error)
	require.NoError(t, db.Exec("DELETE FROM on_duties").Error)

	return db
}
