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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name:           "rejects unsupported role",
			tokenRole:      "owner",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid or expired token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			_, router := gin.CreateTestContext(recorder)

			router.Use(JWTAuth(jwtAuth))
			router.GET("/", func(c *gin.Context) {
				principal, ok := GetPrincipal(c)
				if tt.expectPrincipal {
					require.True(t, ok)
					assert.Equal(t, uint(7), principal.UserID)
					assert.Equal(t, "alice", principal.Username)
					assert.Equal(t, tt.expectedContextRole, principal.Role)
					assert.Equal(t, uint(7), GetUserID(c))
					assert.Equal(t, "alice", GetUsername(c))
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
				token, err := jwtAuth.GenerateToken(7, "alice", tt.tokenRole)
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
