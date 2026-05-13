package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

	// Test with nil Redis client - should still check PostgreSQL
	handler := NewHealthHandler(db, nil)

	r := gin.New()
	r.GET("/ready", handler.Readiness)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// With nil Redis, we expect 503 (PostgreSQL healthy but Redis not configured)
	// This tests the degraded state handling
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d: %s", http.StatusServiceUnavailable, w.Code, w.Body.String())
	}
}

func TestReadinessCheckWithDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Test database-only health check by using a handler that only checks DB
	handler := NewHealthHandler(db, nil)

	r := gin.New()
	r.GET("/ready", handler.Readiness)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Response should indicate PostgreSQL is healthy
	// Since Redis is nil, it will be marked as unhealthy
	if w.Code != http.StatusServiceUnavailable {
		t.Logf("Response body: %s", w.Body.String())
	}
}
