package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireCapability(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		principal      *Principal
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing principal returns unauthorized",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
		{
			name: "principal without capability returns forbidden",
			principal: &Principal{
				UserID:   2,
				Username: "viewer",
				Role:     authz.RoleViewer,
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"insufficient permissions"}`,
		},
		{
			name: "principal with capability is allowed",
			principal: &Principal{
				UserID:   1,
				Username: "admin",
				Role:     authz.RoleAdmin,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"ok":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.principal != nil {
					SetPrincipal(c, *tt.principal)
				}
				c.Next()
			})
			router.GET("/", RequireCapability(authz.CapabilityManageUsers), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.JSONEq(t, tt.expectedBody, recorder.Body.String())
		})
	}
}
