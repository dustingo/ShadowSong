package stats

import (
	"fmt"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}))
	return db
}

func createTestAlert(db *gorm.DB, alertID, severity, status string, triggerTime time.Time) error {
	alert := models.Alert{
		AlertID:     alertID,
		Source:      "test-source",
		AlertName:   "TestAlert",
		Severity:    severity,
		Message:     "Test message",
		Labels:      datatypes.JSON(`{"test": "label"}`),
		Fingerprint: "fp-" + alertID,
		TriggerTime: triggerTime,
		ReceivedAt:  time.Now(),
		Status:      status,
	}
	return db.Create(&alert).Error
}

func TestGetAlertStats_WithMultipleAlerts(t *testing.T) {
	db := setupTestDB(t)

	// Create alerts with different statuses and severities
	now := time.Now()

	// Firing alerts
	require.NoError(t, createTestAlert(db, "alert-1", "P0", "firing", now))
	require.NoError(t, createTestAlert(db, "alert-2", "P0", "firing", now))
	require.NoError(t, createTestAlert(db, "alert-3", "P1", "firing", now))
	require.NoError(t, createTestAlert(db, "alert-4", "P2", "firing", now))

	// Acked alerts
	require.NoError(t, createTestAlert(db, "alert-5", "P1", "acked", now))
	require.NoError(t, createTestAlert(db, "alert-6", "P2", "acked", now))

	// Silenced alerts
	require.NoError(t, createTestAlert(db, "alert-7", "P3", "silenced", now))

	// Resolved alerts (should not be counted in active stats)
	require.NoError(t, createTestAlert(db, "alert-8", "P0", "resolved", now))

	stats, err := GetAlertStats(db)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Verify total count (all alerts)
	assert.Equal(t, 8, stats.Total)

	// Verify status counts
	assert.Equal(t, 4, stats.Firing)
	assert.Equal(t, 2, stats.Acked)
	assert.Equal(t, 1, stats.Silenced)

	// Verify severity counts (firing only)
	assert.Equal(t, 2, stats.BySeverity["P0"])
	assert.Equal(t, 1, stats.BySeverity["P1"])
	assert.Equal(t, 1, stats.BySeverity["P2"])
	assert.Equal(t, 0, stats.BySeverity["P3"]) // P3 is silenced, not firing

	// Verify trend has 24 hours
	assert.Len(t, stats.Trend, 24)
}

func TestGetAlertStats_WithEmptyDatabase(t *testing.T) {
	db := setupTestDB(t)

	stats, err := GetAlertStats(db)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Verify all counts are zero
	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, 0, stats.Firing)
	assert.Equal(t, 0, stats.Acked)
	assert.Equal(t, 0, stats.Silenced)

	// Verify all severity levels have entries even if 0
	assert.Equal(t, 0, stats.BySeverity["P0"])
	assert.Equal(t, 0, stats.BySeverity["P1"])
	assert.Equal(t, 0, stats.BySeverity["P2"])
	assert.Equal(t, 0, stats.BySeverity["P3"])

	// Verify trend has 24 hours with zero counts
	assert.Len(t, stats.Trend, 24)
	for _, point := range stats.Trend {
		assert.Equal(t, 0, point.Count)
	}
}

func TestGetAlertStats_TrendHas24Hours(t *testing.T) {
	db := setupTestDB(t)

	// Create alerts at different hours
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	require.NoError(t, createTestAlert(db, "alert-now", "P0", "firing", now))
	require.NoError(t, createTestAlert(db, "alert-1h", "P1", "firing", hourAgo))
	require.NoError(t, createTestAlert(db, "alert-2h", "P2", "firing", twoHoursAgo))

	stats, err := GetAlertStats(db)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Verify trend has exactly 24 hours
	assert.Len(t, stats.Trend, 24)

	// Verify each trend point has valid time and count
	for _, point := range stats.Trend {
		assert.False(t, point.Time.IsZero())
		assert.GreaterOrEqual(t, point.Count, 0)
	}
}

func TestGetAlertStats_AllSeverityLevelsPresent(t *testing.T) {
	db := setupTestDB(t)

	// Create only P0 and P1 alerts
	now := time.Now()
	require.NoError(t, createTestAlert(db, "alert-p0", "P0", "firing", now))
	require.NoError(t, createTestAlert(db, "alert-p1", "P1", "firing", now))

	stats, err := GetAlertStats(db)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// All severity levels should be present
	assert.Contains(t, stats.BySeverity, "P0")
	assert.Contains(t, stats.BySeverity, "P1")
	assert.Contains(t, stats.BySeverity, "P2")
	assert.Contains(t, stats.BySeverity, "P3")

	// Verify values
	assert.Equal(t, 1, stats.BySeverity["P0"])
	assert.Equal(t, 1, stats.BySeverity["P1"])
	assert.Equal(t, 0, stats.BySeverity["P2"])
	assert.Equal(t, 0, stats.BySeverity["P3"])
}
