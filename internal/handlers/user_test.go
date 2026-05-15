package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateUserRejectsInvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := &UserHandler{db: db}

	req := newJSONRequest(t, http.MethodPost, "/users", map[string]any{
		"username": "alice",
		"name":     "Alice",
		"role":     "owner",
		"password": "plain-text-password",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.CreateUser(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"invalid role"}`, recorder.Body.String())
}

func TestCreateUserDefaultsEmptyRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := &UserHandler{db: db}

	req := newJSONRequest(t, http.MethodPost, "/users", map[string]any{
		"username": "alice",
		"name":     "Alice",
		"password": "plain-text-password",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.CreateUser(context)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var saved models.User
	require.NoError(t, db.First(&saved).Error)
	assert.Equal(t, authz.RoleViewer, saved.Role)
}

func TestLoginPreservesResponseContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	jwtAuth := newUserTestJWT()
	handler := NewUserHandler(db, jwtAuth)

	user := seedUser(t, db, "alice", authz.RoleOperator)

	req := newJSONRequest(t, http.MethodPost, "/auth/login", map[string]string{
		"username": "alice",
		"password": "plain-text-password",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.Login(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response LoginResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.NotEmpty(t, response.Token)
	require.NotNil(t, response.User)
	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, authz.RoleOperator, response.User.Role)
	assert.Equal(t, "", response.User.PasswordHash)
	assert.False(t, response.User.ForcePasswordReset)

	claims, err := jwtAuth.ValidateToken(response.Token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, authz.RoleOperator, claims.Role)
}

func TestLoginRejectsDisabledUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	user := seedUser(t, db, "disabled", authz.RoleViewer)
	now := time.Now()
	user.SetDisabled(true, now)
	require.NoError(t, db.Save(user).Error)

	req := newJSONRequest(t, http.MethodPost, "/auth/login", map[string]string{
		"username": user.Username,
		"password": "plain-text-password",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.Login(context)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.JSONEq(t, `{"error":"account disabled"}`, recorder.Body.String())
}

func TestLoginReturnsForcedResetState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	user := seedUser(t, db, "reset-me", authz.RoleViewer)
	user.SetForcePasswordReset(true, time.Now())
	require.NoError(t, db.Save(user).Error)

	req := newJSONRequest(t, http.MethodPost, "/auth/login", map[string]string{
		"username": user.Username,
		"password": "plain-text-password",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.Login(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response LoginResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.User)
	assert.True(t, response.User.ForcePasswordReset)
}

func TestRefreshTokenRejectsDisabledAndStaleSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	jwtAuth := newUserTestJWT()
	handler := NewUserHandler(db, jwtAuth)

	disabled := seedUser(t, db, "disabled", authz.RoleViewer)
	disabled.SetDisabled(true, time.Now())
	require.NoError(t, db.Save(disabled).Error)
	disabledToken, err := jwtAuth.GenerateToken(disabled.ID, disabled.Username, disabled.Role)
	require.NoError(t, err)

	stale := seedUser(t, db, "stale", authz.RoleViewer)
	staleToken, err := jwtAuth.GenerateToken(stale.ID, stale.Username, stale.Role)
	require.NoError(t, err)
	stale.InvalidateTokens(time.Now().Add(time.Second))
	require.NoError(t, db.Save(stale).Error)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "disabled user refresh denied",
			token:          disabledToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"account disabled"}`,
		},
		{
			name:           "stale token refresh denied",
			token:          staleToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"session expired"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = req

			handler.RefreshToken(context)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
		})
	}
}

func TestAdminUpdateUserRejectsInvalidRoleAndSelfProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	admin := seedUser(t, db, "admin", authz.RoleAdmin)
	target := seedUser(t, db, "target", authz.RoleViewer)

	tests := []struct {
		name           string
		targetID       uint
		currentUserID  uint
		body           map[string]any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "reject invalid role",
			targetID:       target.ID,
			currentUserID:  admin.ID,
			body:           map[string]any{"role": "owner"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid role"}`,
		},
		{
			name:           "reject self disable",
			targetID:       admin.ID,
			currentUserID:  admin.ID,
			body:           map[string]any{"disabled": true},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"cannot disable yourself"}`,
		},
		{
			name:           "reject self demotion",
			targetID:       admin.ID,
			currentUserID:  admin.ID,
			body:           map[string]any{"role": authz.RoleViewer},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"cannot change your own role"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newJSONRequest(t, http.MethodPatch, fmt.Sprintf("/users/%d", tt.targetID), tt.body)
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = req
			context.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", tt.targetID)}}
			context.Set(middleware.UserIDKey, tt.currentUserID)
			context.Set(middleware.UsernameKey, "admin")
			context.Set(middleware.RoleKey, authz.RoleAdmin)

			handler.AdminUpdateUser(context)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
		})
	}
}

func TestAdminUpdateUserInvalidatesTargetSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	admin := seedUser(t, db, "admin", authz.RoleAdmin)
	target := seedUser(t, db, "target", authz.RoleViewer)

	req := newJSONRequest(t, http.MethodPatch, fmt.Sprintf("/users/%d", target.ID), map[string]any{
		"role":                 authz.RoleOperator,
		"disabled":             true,
		"force_password_reset": true,
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req
	context.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", target.ID)}}
	context.Set(middleware.UserIDKey, admin.ID)
	context.Set(middleware.UsernameKey, admin.Username)
	context.Set(middleware.RoleKey, authz.RoleAdmin)

	handler.AdminUpdateUser(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var saved models.User
	require.NoError(t, db.First(&saved, target.ID).Error)
	assert.Equal(t, authz.RoleOperator, saved.Role)
	assert.True(t, saved.IsDisabled())
	assert.True(t, saved.ForcePasswordReset)
	assert.NotNil(t, saved.TokenInvalidBefore)

	audits := loadAuditLogs(t, db)
	require.Len(t, audits, 3)
	assert.Equal(t, "user.role_change", audits[0].Action)
	assert.Equal(t, "user.disable", audits[1].Action)
	assert.Equal(t, "user.force_password_reset", audits[2].Action)

	for _, audit := range audits {
		assert.Equal(t, auditResultAllowed, audit.Result)
		assert.Equal(t, admin.Username, audit.ActorUsername)
		assert.Equal(t, authz.RoleAdmin, audit.ActorRole)
		assert.Equal(t, fmt.Sprintf("%d", target.ID), audit.TargetID)
	}
	assert.Equal(t, "role=operator", audits[0].Detail)
	assert.Equal(t, "disabled=true", audits[1].Detail)
	assert.Equal(t, "force_password_reset=true", audits[2].Detail)
}

func TestUpdateOwnProfileRejectsAdminOnlyFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	user := seedUser(t, db, "profile", authz.RoleViewer)

	req := newJSONRequest(t, http.MethodPatch, "/users/me/profile", map[string]any{
		"name": "Updated Name",
		"role": authz.RoleAdmin,
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req
	context.Set(middleware.UserIDKey, user.ID)
	context.Set(middleware.UsernameKey, user.Username)
	context.Set(middleware.RoleKey, user.Role)

	handler.UpdateOwnProfile(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var saved models.User
	require.NoError(t, db.First(&saved, user.ID).Error)
	assert.Equal(t, authz.RoleViewer, saved.Role)
	assert.Equal(t, user.Name, saved.Name)
}

func TestUpdateOwnPasswordClearsForcedResetAndInvalidatesTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	user := seedUser(t, db, "password-user", authz.RoleViewer)
	user.SetForcePasswordReset(true, time.Now())
	require.NoError(t, db.Save(user).Error)

	req := newJSONRequest(t, http.MethodPut, "/users/me/password", map[string]any{
		"password": "new-password",
	})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req
	context.Set(middleware.UserIDKey, user.ID)
	context.Set(middleware.UsernameKey, user.Username)
	context.Set(middleware.RoleKey, user.Role)

	handler.UpdateOwnPassword(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var saved models.User
	require.NoError(t, db.First(&saved, user.ID).Error)
	assert.True(t, saved.CheckPassword("new-password"))
	assert.False(t, saved.ForcePasswordReset)
	assert.NotNil(t, saved.TokenInvalidBefore)

	var audit models.AuditLog
	require.NoError(t, db.Order("id DESC").First(&audit).Error)
	assert.Equal(t, "user.password_change", audit.Action)
	assert.Equal(t, auditResultAllowed, audit.Result)
	assert.Equal(t, user.Username, audit.ActorUsername)
	assert.Equal(t, authz.RoleViewer, audit.ActorRole)
	assert.Equal(t, fmt.Sprintf("%d", user.ID), audit.TargetID)
	assert.Equal(t, "self-service password update", audit.Detail)
}

func TestAdminUpdateUserDeniedWritesAuditLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	admin := seedUser(t, db, "admin", authz.RoleAdmin)

	tests := []struct {
		name           string
		body           map[string]any
		expectedAction string
		expectedDetail string
		expectedBody   string
	}{
		{
			name:           "self disable denied writes audit log",
			body:           map[string]any{"disabled": true},
			expectedAction: "user.disable",
			expectedDetail: "cannot disable yourself",
			expectedBody:   `{"error":"cannot disable yourself"}`,
		},
		{
			name:           "self role change denied writes audit log",
			body:           map[string]any{"role": authz.RoleViewer},
			expectedAction: "user.role_change",
			expectedDetail: "cannot change your own role",
			expectedBody:   `{"error":"cannot change your own role"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, db.Exec("DELETE FROM audit_logs").Error)

			req := newJSONRequest(t, http.MethodPatch, fmt.Sprintf("/users/%d", admin.ID), tt.body)
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = req
			context.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", admin.ID)}}
			context.Set(middleware.UserIDKey, admin.ID)
			context.Set(middleware.UsernameKey, admin.Username)
			context.Set(middleware.RoleKey, authz.RoleAdmin)

			handler.AdminUpdateUser(context)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())

			var audit models.AuditLog
			require.NoError(t, db.Order("id DESC").First(&audit).Error)
			assert.Equal(t, tt.expectedAction, audit.Action)
			assert.Equal(t, auditResultDenied, audit.Result)
			assert.Equal(t, tt.expectedDetail, audit.Detail)
			assert.Equal(t, admin.Username, audit.ActorUsername)
			assert.Equal(t, authz.RoleAdmin, audit.ActorRole)
			assert.Equal(t, fmt.Sprintf("%d", admin.ID), audit.TargetID)
		})
	}
}

func loadAuditLogs(t *testing.T, db *gorm.DB) []models.AuditLog {
	t.Helper()

	var audits []models.AuditLog
	require.NoError(t, db.Order("id ASC").Find(&audits).Error)
	return audits
}

func openUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate user model: %v", err)
	}
	if err := db.AutoMigrate(&models.AuditLog{}); err != nil {
		t.Fatalf("migrate user model: %v", err)
	}

	return db
}

func seedUser(t *testing.T, db *gorm.DB, username, role string) *models.User {
	t.Helper()

	user := &models.User{
		Username: username,
		Name:     username,
		Role:     role,
	}
	require.NoError(t, user.SetPassword("plain-text-password"))
	require.NoError(t, db.Create(user).Error)

	return user
}

func newJSONRequest(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestListAuditLogsReturnsPaginatedResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	for i := 0; i < 5; i++ {
		require.NoError(t, db.Create(&models.AuditLog{
			ActorUserID:   1,
			ActorUsername: "admin",
			ActorRole:     authz.RoleAdmin,
			Action:        "user.create",
			TargetType:    "user",
			TargetID:      fmt.Sprintf("%d", i+2),
			Result:        auditResultAllowed,
			Detail:        "role=viewer",
		}).Error)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/audit-logs?page=1&page_size=3", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.ListAuditLogs(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float64(5), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(3), response["page_size"])

	items := response["items"].([]any)
	assert.Len(t, items, 3)
}

func TestListAuditLogsFiltersByAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.create", TargetType: "user", TargetID: "2", Result: auditResultAllowed,
	}).Error)
	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.disable", TargetType: "user", TargetID: "2", Result: auditResultAllowed,
	}).Error)

	req := httptest.NewRequest(http.MethodGet, "/users/audit-logs?action=user.create", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.ListAuditLogs(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float64(1), response["total"])
}

func TestListAuditLogsFiltersByResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.create", TargetType: "user", TargetID: "2", Result: auditResultAllowed,
	}).Error)
	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.disable", TargetType: "user", TargetID: "2", Result: auditResultDenied,
	}).Error)

	req := httptest.NewRequest(http.MethodGet, "/users/audit-logs?result=denied", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.ListAuditLogs(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float64(1), response["total"])
}

func newUserTestJWT() *auth.JWT {
	return auth.NewJWT(&config.SecurityConfig{
		JWTSecret:   "test-secret",
		TokenExpiry: time.Hour,
	})
}
