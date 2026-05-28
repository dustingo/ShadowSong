package retention

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRetentionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}, &models.NotificationDelivery{}, &models.NotificationDeliveryAttempt{}))
	return db
}

func TestCleanupDeletesOldRecords(t *testing.T) {
	db := newRetentionTestDB(t)

	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	require.NoError(t, db.Create(&models.Alert{
		AlertID:     "old-alert",
		Source:      "test",
		AlertName:   "Old",
		Severity:    "P1",
		Message:     "old",
		Fingerprint: "fp-old",
		Status:      "resolved",
		TriggerTime: oldTime,
		ReceivedAt:  oldTime,
	}).Error)

	require.NoError(t, db.Create(&models.Alert{
		AlertID:     "new-alert",
		Source:      "test",
		AlertName:   "New",
		Severity:    "P1",
		Message:     "new",
		Fingerprint: "fp-new",
		Status:      "firing",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
	}).Error)

	result := Cleanup(db, 30)

	assert.Equal(t, int64(1), result.AlertsDeleted)

	var count int64
	db.Model(&models.Alert{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestCleanupDisabledWhenZero(t *testing.T) {
	db := newRetentionTestDB(t)
	result := Cleanup(db, 0)
	assert.Equal(t, int64(0), result.AlertsDeleted)
}
