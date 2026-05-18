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

func setupAlertDeliveryTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}, &models.NotificationDelivery{}, &models.NotificationDeliveryAttempt{}))
	return db
}

func TestAlertDeliveries_ReturnsDeliveriesWithAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAlertDeliveryTestDB(t)

	alertID := "alert-del-1"

	// Create a notification delivery
	delivery := models.NotificationDelivery{
		AlertID:        alertID,
		TraceID:        "trace-1",
		ChannelID:      1,
		DeliveryStatus: models.DeliveryStatusDelivered,
		DeliveryMode:   models.DeliveryModeDefault,
		AlertSnapshot:  datatypes.JSON(`{"alert_id":"alert-del-1"}`),
		ChannelSnapshot: datatypes.JSON(`{"id":1,"name":"test-channel","type":"webhook"}`),
		RenderedPayloadSnapshot: datatypes.JSON(`{"title":"Test","content":"body"}`),
	}
	require.NoError(t, db.Create(&delivery).Error)

	// Create an attempt for the delivery
	attempt := models.NotificationDeliveryAttempt{
		DeliveryID:    delivery.ID,
		AttemptNumber: 1,
		Result:        models.AttemptResultSuccess,
		Retryable:     false,
		DurationMS:    150,
		TriggerKind:   models.TriggerKindPipeline,
	}
	require.NoError(t, db.Create(&attempt).Error)

	handler := NewAlertHandler(db)

	r := gin.New()
	r.GET("/alerts/:id/deliveries", handler.AlertDeliveries)

	req := httptest.NewRequest("GET", "/alerts/alert-del-1/deliveries", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var deliveries []models.NotificationDelivery
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &deliveries))
	require.Len(t, deliveries, 1)
	assert.Equal(t, alertID, deliveries[0].AlertID)
	assert.Equal(t, models.DeliveryStatusDelivered, deliveries[0].DeliveryStatus)
	require.Len(t, deliveries[0].Attempts, 1)
	assert.Equal(t, 1, deliveries[0].Attempts[0].AttemptNumber)
	assert.Equal(t, models.AttemptResultSuccess, deliveries[0].Attempts[0].Result)
}

func TestAlertDeliveries_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupAlertDeliveryTestDB(t)

	handler := NewAlertHandler(db)

	r := gin.New()
	r.GET("/alerts/:id/deliveries", handler.AlertDeliveries)

	req := httptest.NewRequest("GET", "/alerts/nonexistent-alert/deliveries", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var deliveries []models.NotificationDelivery
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &deliveries))
	assert.Len(t, deliveries, 0)
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
