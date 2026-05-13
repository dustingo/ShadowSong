package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestSizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		maxSize        int64
		contentLength  int64
		body           string
		expectedStatus int
	}{
		{
			name:           "request under limit",
			maxSize:        100,
			contentLength:  50,
			body:           strings.Repeat("a", 50),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "request over limit",
			maxSize:        100,
			contentLength:  200,
			body:           strings.Repeat("a", 200),
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(RequestSizeLimit(tt.maxSize))
			r.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Length", strconv.FormatInt(tt.contentLength, 10))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}