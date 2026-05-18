package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAlertTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}))
	return db
}

func createFiringAlert(db *gorm.DB, alertID, severity, fingerprint string, triggerTime time.Time) error {
	alert := models.Alert{
		AlertID:     alertID,
		Source:      "test-source",
		AlertName:   "TestAlert",
		Severity:    severity,
		Message:     "Test message",
		Labels:      datatypes.JSON(`{"test": "label"}`),
		Fingerprint: fingerprint,
		TriggerTime: triggerTime,
		ReceivedAt:  time.Now(),
		Status:      "firing",
	}
	return db.Create(&alert).Error
}

func TestActiveAlerts_ReturnsFlatList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAlertTestDB(t)
	now := time.Now()

	require.NoError(t, createFiringAlert(db, "alert-1", "P0", "fp-a", now))
	require.NoError(t, createFiringAlert(db, "alert-2", "P1", "fp-b", now))

	handler := NewAlertHandler(db)

	r := gin.New()
	r.GET("/alerts/active", handler.Active)

	req := httptest.NewRequest("GET", "/alerts/active", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var alerts []models.Alert
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &alerts))
	assert.Len(t, alerts, 2)
}

func TestActiveAlerts_GroupedByFingerprint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAlertTestDB(t)
	now := time.Now()

	// Create 3 alerts with the same fingerprint
	require.NoError(t, createFiringAlert(db, "alert-a", "P0", "fp-same", now.Add(-2*time.Minute)))
	require.NoError(t, createFiringAlert(db, "alert-b", "P0", "fp-same", now.Add(-1*time.Minute)))
	require.NoError(t, createFiringAlert(db, "alert-c", "P0", "fp-same", now))

	// Create 1 alert with a different fingerprint
	require.NoError(t, createFiringAlert(db, "alert-d", "P1", "fp-other", now))

	handler := NewAlertHandler(db)

	r := gin.New()
	r.GET("/alerts/active", handler.Active)

	req := httptest.NewRequest("GET", "/alerts/active?grouped=true", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var grouped []GroupedActiveAlert
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &grouped))
	assert.Len(t, grouped, 2)

	// Find the group with fingerprint "fp-same"
	var sameGroup *GroupedActiveAlert
	for i := range grouped {
		if grouped[i].Fingerprint == "fp-same" {
			sameGroup = &grouped[i]
			break
		}
	}
	require.NotNil(t, sameGroup, "expected to find group with fingerprint fp-same")
	assert.Equal(t, 3, sameGroup.Count)
	assert.Equal(t, "alert-c", sameGroup.LatestAlert.AlertID, "latest alert should be the one with most recent trigger_time")
	assert.True(t, sameGroup.FirstTriggeredAt.Before(sameGroup.LastTriggeredAt) || sameGroup.FirstTriggeredAt.Equal(sameGroup.LastTriggeredAt))

	// Find the group with fingerprint "fp-other"
	var otherGroup *GroupedActiveAlert
	for i := range grouped {
		if grouped[i].Fingerprint == "fp-other" {
			otherGroup = &grouped[i]
			break
		}
	}
	require.NotNil(t, otherGroup, "expected to find group with fingerprint fp-other")
	assert.Equal(t, 1, otherGroup.Count)
	assert.Equal(t, "alert-d", otherGroup.LatestAlert.AlertID)
}

func TestActiveAlerts_GroupedFalseReturnsFlat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAlertTestDB(t)
	now := time.Now()

	require.NoError(t, createFiringAlert(db, "alert-1", "P0", "fp-a", now))
	require.NoError(t, createFiringAlert(db, "alert-2", "P0", "fp-a", now))

	handler := NewAlertHandler(db)

	r := gin.New()
	r.GET("/alerts/active", handler.Active)

	req := httptest.NewRequest("GET", "/alerts/active?grouped=false", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var alerts []models.Alert
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &alerts))
	assert.Len(t, alerts, 2)
}

func TestActiveAlerts_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAlertTestDB(t)

	handler := NewAlertHandler(db)

	r := gin.New()
	r.GET("/alerts/active", handler.Active)

	req := httptest.NewRequest("GET", "/alerts/active?grouped=true", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var grouped []GroupedActiveAlert
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &grouped))
	assert.Len(t, grouped, 0)
}
