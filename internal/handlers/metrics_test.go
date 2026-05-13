package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// Auto-migrate delivery tables
	db.AutoMigrate(&models.NotificationDelivery{})

	h := NewMetricsHandler(db)

	r := gin.New()
	r.GET("/api/v1/metrics", h.GetMetrics)

	req := httptest.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
