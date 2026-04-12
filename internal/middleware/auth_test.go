package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := auth.NewJWT(&config.SecurityConfig{
		JWTSecret:   "test-secret",
		TokenExpiry: time.Hour,
	})

	tests := []struct {
		name                string
		header              string
		tokenRole           string
		tokenUserID         uint
		tokenUsername       string
		setupUser           func(t *testing.T, db *gorm.DB, seeded *models.User)
		expectedStatus      int
		expectedBody        string
		expectPrincipal     bool
		expectedContextRole string
	}{
		{
			name:           "missing header",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"authorization header required"}`,
		},
		{
			name:           "invalid bearer format",
			header:         "Token abc",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid authorization header format"}`,
		},
		{
			name:                "valid supported role",
			tokenRole:           authz.RoleAdmin,
			expectedStatus:      http.StatusOK,
			expectPrincipal:     true,
			expectedContextRole: authz.RoleAdmin,
		},
		{
			name:      "rejects disabled account",
			tokenRole: authz.RoleViewer,
			setupUser: func(t *testing.T, db *gorm.DB, seeded *models.User) {
				seeded.Role = authz.RoleViewer
				disabledAt := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
				seeded.DisabledAt = &disabledAt
				require.NoError(t, db.Save(seeded).Error)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"account disabled"}`,
		},
		{
			name:      "rejects stale token after invalidation",
			tokenRole: authz.RoleAdmin,
			setupUser: func(t *testing.T, db *gorm.DB, seeded *models.User) {
				seeded.InvalidateTokens(time.Now().Add(time.Second))
				require.NoError(t, db.Save(seeded).Error)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"session expired"}`,
		},
		{
			name:      "forced reset account limited on normal route",
			tokenRole: authz.RoleAdmin,
			setupUser: func(t *testing.T, db *gorm.DB, seeded *models.User) {
				seeded.SetForcePasswordReset(true, time.Now().Add(-time.Second))
				require.NoError(t, db.Save(seeded).Error)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"password reset required"}`,
		},
		{
			name:           "rejects unsupported role",
			tokenUserID:    1,
			tokenUsername:  "alice",
			tokenRole:      "owner",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid or expired token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openMiddlewareTestDB(t)
			seeded := models.User{
				Username: "alice",
				Name:     "Alice",
				Role:     authz.RoleAdmin,
			}
			require.NoError(t, seeded.SetPassword("password"))
			require.NoError(t, db.Create(&seeded).Error)

			recorder := httptest.NewRecorder()
			_, router := gin.CreateTestContext(recorder)

			if tt.setupUser != nil {
				tt.setupUser(t, db, &seeded)
			}

			router.Use(JWTAuth(jwtAuth, db))
			router.GET("/", func(c *gin.Context) {
				principal, ok := GetPrincipal(c)
				if tt.expectPrincipal {
					require.True(t, ok)
					assert.Equal(t, seeded.ID, principal.UserID)
					assert.Equal(t, seeded.Username, principal.Username)
					assert.Equal(t, tt.expectedContextRole, principal.Role)
					assert.Equal(t, seeded.ID, GetUserID(c))
					assert.Equal(t, seeded.Username, GetUsername(c))
					assert.Equal(t, tt.expectedContextRole, GetUserRole(c))
				} else {
					assert.False(t, ok)
				}
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(AuthorizationHeader, tt.header)
			}
			if tt.tokenRole != "" {
				tokenUserID := tt.tokenUserID
				if tokenUserID == 0 {
					tokenUserID = seeded.ID
				}
				tokenUsername := tt.tokenUsername
				if tokenUsername == "" {
					tokenUsername = seeded.Username
				}

				token, err := jwtAuth.GenerateToken(tokenUserID, tokenUsername, tt.tokenRole)
				require.NoError(t, err)
				req.Header.Set(AuthorizationHeader, "Bearer "+token)
			}

			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func TestJWTAuthAllowsForcedResetRecoveryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtAuth := auth.NewJWT(&config.SecurityConfig{
		JWTSecret:   "test-secret",
		TokenExpiry: time.Hour,
	})
	db := openMiddlewareTestDB(t)

	user := models.User{
		Username:           "reset-user",
		Name:               "Reset User",
		Role:               authz.RoleViewer,
		ForcePasswordReset: true,
	}
	require.NoError(t, user.SetPassword("password"))
	require.NoError(t, db.Create(&user).Error)

	token, err := jwtAuth.GenerateToken(user.ID, user.Username, user.Role)
	require.NoError(t, err)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/users/me"},
		{method: http.MethodPatch, path: "/api/v1/users/me/profile"},
		{method: http.MethodPut, path: "/api/v1/users/me/password"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			_, router := gin.CreateTestContext(recorder)
			router.Use(JWTAuth(jwtAuth, db))
			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set(AuthorizationHeader, "Bearer "+token)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
		})
	}
}

func TestPrincipalCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	SetPrincipal(context, Principal{
		UserID:   9,
		Username: "carol",
		Role:     authz.RoleViewer,
	})

	principal, ok := GetPrincipal(context)
	require.True(t, ok)
	assert.Equal(t, uint(9), principal.UserID)
	assert.Equal(t, "carol", principal.Username)
	assert.Equal(t, authz.RoleViewer, principal.Role)
	assert.Equal(t, uint(9), GetUserID(context))
	assert.Equal(t, "carol", GetUsername(context))
	assert.Equal(t, authz.RoleViewer, GetUserRole(context))
}

func TestPrincipalMarshalsStableFields(t *testing.T) {
	principal := Principal{
		UserID:   11,
		Username: "dave",
		Role:     authz.RoleOperator,
	}

	payload, err := json.Marshal(principal)
	require.NoError(t, err)
	assert.JSONEq(t, `{"user_id":11,"username":"dave","role":"operator"}`, string(payload))
}

func openMiddlewareTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:middleware_auth?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	require.NoError(t, db.Exec("DELETE FROM users").Error)

	return db
}
