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

func TestChannelHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	db.AutoMigrate(&models.Channel{}, &models.NotificationDelivery{})

	// Create test channel
	channel := models.Channel{Name: "test-channel", Type: "email"}
	db.Create(&channel)

	h := NewChannelHealthHandler(db)

	r := gin.New()
	r.GET("/api/v1/channels/:id/health", h.GetChannelHealth)

	req := httptest.NewRequest("GET", "/api/v1/channels/1/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
