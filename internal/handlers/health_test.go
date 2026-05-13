package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReadinessCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Create a mock Redis client (we'll test the handler logic, not actual Redis connection)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	handler := NewHealthHandler(db, redisClient)

	r := gin.New()
	r.GET("/ready", handler.Readiness)

	tests := []struct {
		name       string
		setupDB    bool
		wantStatus int
	}{
		{
			name:       "returns 200 when database is connected",
			setupDB:    true,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ready", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
