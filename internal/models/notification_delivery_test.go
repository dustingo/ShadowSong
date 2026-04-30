package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotificationDeliveryDefaultsAndValidation(t *testing.T) {
	db := newNotificationDeliveryTestDB(t)

	delivery := NotificationDelivery{
		AlertID:                 "alert-1",
		TraceID:                 "trace-1",
		ChannelID:               7,
		RouteRuleID:             uintPtr(9),
		DeliveryMode:            DeliveryModeRendered,
		AlertSnapshot:           datatypes.JSON(`{"alert_id":"alert-1"}`),
		ChannelSnapshot:         datatypes.JSON(`{"id":7,"name":"ops-webhook","type":"webhook"}`),
		RouteSnapshot:           datatypes.JSON(`{"id":9,"name":"p1-route"}`),
		RenderedPayloadSnapshot: datatypes.JSON(`{"title":"CPUHigh","content":"cpu high"}`),
	}

	require.NoError(t, db.Create(&delivery).Error)

	assert.Equal(t, DeliveryStatusPending, delivery.DeliveryStatus)
	assert.Equal(t, 0, delivery.AttemptCount)
	assert.JSONEq(t, `{"alert_id":"alert-1"}`, string(delivery.AlertSnapshot))
	assert.NotZero(t, delivery.CreatedAt)
	assert.NotZero(t, delivery.UpdatedAt)
}

func TestNotificationDeliveryStatusValidation(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "pending", status: DeliveryStatusPending},
		{name: "delivered", status: DeliveryStatusDelivered},
		{name: "failed", status: DeliveryStatusFailed},
		{name: "invalid", status: "boom", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := NotificationDelivery{
				AlertID:                 "alert-1",
				TraceID:                 "trace-1",
				ChannelID:               7,
				DeliveryStatus:          tt.status,
				DeliveryMode:            DeliveryModeRendered,
				AlertSnapshot:           datatypes.JSON(`{"alert_id":"alert-1"}`),
				ChannelSnapshot:         datatypes.JSON(`{"id":7}`),
				RenderedPayloadSnapshot: datatypes.JSON(`{"title":"CPUHigh"}`),
			}

			err := delivery.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNotificationDeliveryAttemptUniqueNumberAndAppendOnly(t *testing.T) {
	db := newNotificationDeliveryTestDB(t)

	delivery := NotificationDelivery{
		AlertID:                 "alert-1",
		TraceID:                 "trace-1",
		ChannelID:               7,
		DeliveryMode:            DeliveryModeRendered,
		AlertSnapshot:           datatypes.JSON(`{"alert_id":"alert-1"}`),
		ChannelSnapshot:         datatypes.JSON(`{"id":7}`),
		RenderedPayloadSnapshot: datatypes.JSON(`{"title":"CPUHigh"}`),
	}
	require.NoError(t, db.Create(&delivery).Error)

	firstAttempt := NotificationDeliveryAttempt{
		DeliveryID:    delivery.ID,
		AttemptNumber: 1,
		Result:        AttemptResultFailed,
		Retryable:     true,
		ErrorMessage:  "temporary failure",
		TriggerKind:   TriggerKindPipeline,
	}
	require.NoError(t, db.Create(&firstAttempt).Error)

	duplicateAttempt := NotificationDeliveryAttempt{
		DeliveryID:    delivery.ID,
		AttemptNumber: 1,
		Result:        AttemptResultSuccess,
		TriggerKind:   TriggerKindPipeline,
	}
	assert.Error(t, db.Create(&duplicateAttempt).Error)

	firstAttempt.ErrorMessage = "mutated"
	assert.Error(t, db.Save(&firstAttempt).Error)
}

func newNotificationDeliveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:notification-delivery?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&NotificationDelivery{}, &NotificationDeliveryAttempt{}))

	return db
}

func uintPtr(v uint) *uint {
	return &v
}
