package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotificationDeliveryRecoveryDefaultsAndValidation(t *testing.T) {
	db := newNotificationDeliveryRecoveryTestDB(t)

	recovery := NotificationDeliveryRecovery{
		OriginalDeliveryID: 11,
		Action:             TriggerKindRetry,
		Reason:             "operator retried after webhook recovered",
		ActorUserID:        7,
		ActorUsername:      "operator-a",
		ActorRole:          "operator",
	}

	require.NoError(t, db.Create(&recovery).Error)
	assert.Equal(t, RecoveryStatusPending, recovery.Status)
	assert.NotZero(t, recovery.RequestedAt)
	assert.NotZero(t, recovery.CreatedAt)
}

func TestNotificationDeliveryRecoveryStatusValidation(t *testing.T) {
	testCases := []struct {
		name    string
		status  string
		action  string
		reason  string
		wantErr bool
	}{
		{name: "pending", status: RecoveryStatusPending, action: TriggerKindRetry, reason: "valid"},
		{name: "in progress", status: RecoveryStatusInProgress, action: TriggerKindReplay, reason: "valid"},
		{name: "succeeded", status: RecoveryStatusSucceeded, action: TriggerKindRetry, reason: "valid"},
		{name: "failed", status: RecoveryStatusFailed, action: TriggerKindReplay, reason: "valid"},
		{name: "rejected", status: RecoveryStatusRejected, action: TriggerKindRetry, reason: "valid"},
		{name: "invalid status", status: "boom", action: TriggerKindRetry, reason: "valid", wantErr: true},
		{name: "invalid action", status: RecoveryStatusPending, action: TriggerKindPipeline, reason: "valid", wantErr: true},
		{name: "missing reason", status: RecoveryStatusPending, action: TriggerKindRetry, reason: "   ", wantErr: true},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			recovery := NotificationDeliveryRecovery{
				OriginalDeliveryID: 5,
				Action:             tt.action,
				Reason:             tt.reason,
				ActorUserID:        1,
				ActorUsername:      "operator",
				ActorRole:          "operator",
				Status:             tt.status,
				RequestedAt:        time.Now(),
			}

			err := recovery.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func newNotificationDeliveryRecoveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:notification-delivery-recovery?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&NotificationDeliveryRecovery{}))

	return db
}
