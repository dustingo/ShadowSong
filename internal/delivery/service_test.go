package delivery

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceStartDeliveryRecordAttemptAndMarkDelivered(t *testing.T) {
	db := newDeliveryTestDB(t)
	service := NewService(db)

	alert := &models.Alert{
		AlertID:     "alert-success",
		TraceID:     "trace-success",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Status:      "firing",
		Fingerprint: "fp-success",
		TriggerTime: time.Date(2026, 4, 30, 1, 0, 0, 0, time.UTC),
		Labels:      datatypes.JSON(`{"instance":"game-01"}`),
	}
	channel := &models.Channel{
		ID:      7,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com","secret":"keep-out"}`),
		Enabled: true,
	}
	route := &models.RouteRule{
		ID:         9,
		Name:       "p1-route",
		Priority:   10,
		ChannelIDs: datatypes.JSON(`[7]`),
		Enabled:    true,
	}

	delivery, err := service.StartDelivery(context.Background(), StartDeliveryInput{
		Alert:         alert,
		Channel:       channel,
		RouteRule:     route,
		DeliveryMode:  models.DeliveryModeRendered,
		TriggerKind:   models.TriggerKindPipeline,
		RenderedTitle: "CPUHigh",
		RenderedBody:  "cpu high on game-01",
	})
	require.NoError(t, err)
	require.NotZero(t, delivery.ID)
	assert.Equal(t, models.DeliveryStatusPending, delivery.DeliveryStatus)
	assert.Equal(t, models.DeliveryModeRendered, delivery.DeliveryMode)
	assertSnapshotExcludesSecrets(t, delivery.ChannelSnapshot)

	attempt, err := service.RecordAttempt(context.Background(), delivery.ID, RecordAttemptInput{
		AttemptNumber: 1,
		Result:        models.AttemptResultSuccess,
		Retryable:     false,
		DurationMS:    25,
		TriggerKind:   models.TriggerKindPipeline,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, attempt.AttemptNumber)

	err = service.MarkDelivered(context.Background(), delivery.ID, MarkDeliveredInput{
		AttemptCount: 1,
		DeliveredAt:  time.Date(2026, 4, 30, 1, 1, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	stored, err := service.GetDeliveryByID(context.Background(), delivery.ID)
	require.NoError(t, err)
	assert.Equal(t, models.DeliveryStatusDelivered, stored.DeliveryStatus)
	assert.Equal(t, 1, stored.AttemptCount)
	require.NotNil(t, stored.LastAttemptAt)
	require.NotNil(t, stored.LastSuccessAt)
	require.Len(t, stored.Attempts, 1)
	assert.Equal(t, models.AttemptResultSuccess, stored.Attempts[0].Result)
}

func TestServiceMarkFailedPersistsTerminalSummary(t *testing.T) {
	db := newDeliveryTestDB(t)
	service := NewService(db)

	delivery, err := service.StartDelivery(context.Background(), StartDeliveryInput{
		Alert: &models.Alert{
			AlertID:     "alert-failed",
			TraceID:     "trace-failed",
			Source:      "prometheus",
			AlertName:   "DiskHigh",
			Severity:    "P0",
			Message:     "disk high",
			Status:      "firing",
			Fingerprint: "fp-failed",
		},
		Channel: &models.Channel{
			ID:      8,
			Name:    "ops-feishu",
			Type:    "feishu",
			Config:  datatypes.JSON(`{"webhook_url":"https://example.com","api_key":"hidden"}`),
			Enabled: true,
		},
		DeliveryMode:  models.DeliveryModeDefault,
		TriggerKind:   models.TriggerKindPipeline,
		RenderedTitle: "[P0] DiskHigh",
		RenderedBody:  "disk high",
	})
	require.NoError(t, err)

	_, err = service.RecordAttempt(context.Background(), delivery.ID, RecordAttemptInput{
		AttemptNumber: 1,
		Result:        models.AttemptResultFailed,
		Retryable:     true,
		ErrorMessage:  "503 upstream",
		HTTPStatus:    intPtr(503),
		DurationMS:    30,
		TriggerKind:   models.TriggerKindPipeline,
	})
	require.NoError(t, err)

	_, err = service.RecordAttempt(context.Background(), delivery.ID, RecordAttemptInput{
		AttemptNumber: 2,
		Result:        models.AttemptResultFailed,
		Retryable:     true,
		ErrorMessage:  "503 upstream",
		HTTPStatus:    intPtr(503),
		DurationMS:    32,
		TriggerKind:   models.TriggerKindPipeline,
	})
	require.NoError(t, err)

	err = service.MarkFailed(context.Background(), delivery.ID, MarkFailedInput{
		AttemptCount: 2,
		FailedAt:     time.Date(2026, 4, 30, 2, 0, 0, 0, time.UTC),
		Result:       models.AttemptResultFailed,
		Retryable:    true,
		ErrorMessage: "retry budget exhausted",
		HTTPStatus:   intPtr(503),
		TriggerKind:  models.TriggerKindPipeline,
	})
	require.NoError(t, err)

	stored, err := service.GetDeliveryByID(context.Background(), delivery.ID)
	require.NoError(t, err)
	assert.Equal(t, models.DeliveryStatusFailed, stored.DeliveryStatus)
	assert.Equal(t, 2, stored.AttemptCount)
	require.NotNil(t, stored.LastAttemptAt)
	assert.Nil(t, stored.LastSuccessAt)
	require.Len(t, stored.Attempts, 2)

	var summary models.FinalFailureSummary
	require.NoError(t, json.Unmarshal(stored.FinalFailureSummary, &summary))
	assert.Equal(t, "retry budget exhausted", summary.ErrorMessage)
	assert.Equal(t, 2, summary.AttemptCount)
	assert.True(t, summary.Retryable)
}

func TestServiceListDeliveriesPreloadsAttempts(t *testing.T) {
	db := newDeliveryTestDB(t)
	service := NewService(db)

	first, err := service.StartDelivery(context.Background(), StartDeliveryInput{
		Alert: &models.Alert{
			AlertID:   "alert-list-1",
			TraceID:   "trace-list-1",
			Source:    "prometheus",
			AlertName: "CPUHigh",
			Severity:  "P1",
			Message:   "cpu high",
			Status:    "firing",
		},
		Channel: &models.Channel{
			ID:      11,
			Name:    "ops-webhook",
			Type:    "webhook",
			Enabled: true,
		},
		DeliveryMode:  models.DeliveryModeRendered,
		TriggerKind:   models.TriggerKindPipeline,
		RenderedTitle: "CPUHigh",
		RenderedBody:  "cpu high",
	})
	require.NoError(t, err)
	_, err = service.RecordAttempt(context.Background(), first.ID, RecordAttemptInput{
		AttemptNumber: 1,
		Result:        models.AttemptResultSuccess,
		TriggerKind:   models.TriggerKindPipeline,
	})
	require.NoError(t, err)

	second, err := service.StartDelivery(context.Background(), StartDeliveryInput{
		Alert: &models.Alert{
			AlertID:   "alert-list-2",
			TraceID:   "trace-list-2",
			Source:    "prometheus",
			AlertName: "MemoryHigh",
			Severity:  "P2",
			Message:   "memory high",
			Status:    "firing",
		},
		Channel: &models.Channel{
			ID:      12,
			Name:    "ops-dingtalk",
			Type:    "dingtalk",
			Enabled: true,
		},
		DeliveryMode:  models.DeliveryModeDefault,
		TriggerKind:   models.TriggerKindPipeline,
		RenderedTitle: "MemoryHigh",
		RenderedBody:  "memory high",
	})
	require.NoError(t, err)
	_, err = service.RecordAttempt(context.Background(), second.ID, RecordAttemptInput{
		AttemptNumber: 1,
		Result:        models.AttemptResultFailed,
		Retryable:     false,
		ErrorMessage:  "non retryable",
		TriggerKind:   models.TriggerKindPipeline,
	})
	require.NoError(t, err)

	deliveries, err := service.ListDeliveries(context.Background(), ListDeliveriesInput{
		Limit: 10,
	})
	require.NoError(t, err)
	require.Len(t, deliveries, 2)
	assert.NotEmpty(t, deliveries[0].Attempts)
	assert.NotEmpty(t, deliveries[1].Attempts)
}

func newDeliveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.NotificationDelivery{}, &models.NotificationDeliveryAttempt{}))

	return db
}

func assertSnapshotExcludesSecrets(t *testing.T, snapshot datatypes.JSON) {
	t.Helper()

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(snapshot, &decoded))
	assert.NotContains(t, decoded, "secret")
	assert.NotContains(t, decoded, "api_key")
	assert.NotContains(t, decoded, "config")
}

func intPtr(v int) *int {
	return &v
}
