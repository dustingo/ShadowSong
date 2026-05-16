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

	deliveries, total, err := service.ListDeliveries(context.Background(), ListDeliveriesInput{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, deliveries, 2)
	assert.NotEmpty(t, deliveries[0].Attempts)
	assert.NotEmpty(t, deliveries[1].Attempts)
}

func TestDeliveryServiceRecoveryRetryUsesFrozenPayload(t *testing.T) {
	db := newDeliveryRecoveryTestDB(t)
	service, senderCalls := newRecoveryTestService(db)
	seedRecoveryData(t, db)

	original := seedFailedDeliveryRecord(t, db, recoveryDeliverySeed{
		alertID:        "alert-retry",
		traceID:        "trace-retry",
		channelID:      21,
		channelName:    "ops-webhook",
		channelType:    "webhook",
		channelEnabled: true,
		routeRuleID:    uintPtr(31),
		routeRuleName:  "route-original",
		renderedTitle:  "frozen-title",
		renderedBody:   "frozen-body",
		alertName:      "CPUHigh",
		source:         "prometheus",
		triggerKind:    models.TriggerKindPipeline,
	})

	result, err := service.RetryDelivery(context.Background(), RetryDeliveryInput{
		OriginalDeliveryID: original.ID,
		Reason:             "operator verified channel recovered",
		ActorUserID:        101,
		ActorUsername:      "operator-a",
		ActorRole:          "operator",
	})
	require.NoError(t, err)
	require.NotNil(t, result.ResultDelivery)
	assert.Equal(t, models.TriggerKindRetry, result.Action)
	assert.Equal(t, models.RecoveryStatusSucceeded, result.Recovery.Status)
	assert.Equal(t, original.ID, result.Recovery.OriginalDeliveryID)
	require.NotNil(t, result.Recovery.ResultDeliveryID)
	assert.Equal(t, result.ResultDelivery.ID, *result.Recovery.ResultDeliveryID)

	storedResult, err := service.GetDeliveryByID(context.Background(), result.ResultDelivery.ID)
	require.NoError(t, err)
	assert.Equal(t, models.DeliveryStatusDelivered, storedResult.DeliveryStatus)
	assert.Equal(t, "frozen-title", decodeRenderedPayload(t, storedResult.RenderedPayloadSnapshot).Title)
	assert.Equal(t, "frozen-body", decodeRenderedPayload(t, storedResult.RenderedPayloadSnapshot).Content)
	require.Len(t, storedResult.Attempts, 1)
	assert.Equal(t, models.TriggerKindRetry, storedResult.Attempts[0].TriggerKind)
	require.Len(t, *senderCalls, 1)
	assert.Equal(t, "frozen-title", (*senderCalls)[0].title)
	assert.Equal(t, "frozen-body", (*senderCalls)[0].content)
}

func TestDeliveryServiceRecoveryReplayUsesLiveRouteTemplate(t *testing.T) {
	db := newDeliveryRecoveryTestDB(t)
	service, senderCalls := newRecoveryTestService(db)
	seedRecoveryData(t, db)

	original := seedFailedDeliveryRecord(t, db, recoveryDeliverySeed{
		alertID:        "alert-replay",
		traceID:        "trace-replay",
		channelID:      21,
		channelName:    "ops-webhook",
		channelType:    "webhook",
		channelEnabled: true,
		routeRuleID:    uintPtr(31),
		routeRuleName:  "route-original",
		renderedTitle:  "stale-title",
		renderedBody:   "stale-body",
		alertName:      "CPUHigh",
		source:         "prometheus",
		triggerKind:    models.TriggerKindPipeline,
	})

	result, err := service.ReplayDelivery(context.Background(), ReplayDeliveryInput{
		OriginalDeliveryID: original.ID,
		Reason:             "reroute through current template",
		ActorUserID:        102,
		ActorUsername:      "operator-b",
		ActorRole:          "operator",
	})
	require.NoError(t, err)
	require.NotNil(t, result.ResultDelivery)
	assert.Equal(t, models.TriggerKindReplay, result.Action)
	assert.Equal(t, models.RecoveryStatusSucceeded, result.Recovery.Status)

	storedResult, err := service.GetDeliveryByID(context.Background(), result.ResultDelivery.ID)
	require.NoError(t, err)
	payload := decodeRenderedPayload(t, storedResult.RenderedPayloadSnapshot)
	assert.Equal(t, "[LIVE-P0] CPUHigh", payload.Title)
	assert.Contains(t, payload.Content, "route=live-route")
	assert.Contains(t, payload.Content, "event=critical")
	require.Len(t, storedResult.Attempts, 1)
	assert.Equal(t, models.TriggerKindReplay, storedResult.Attempts[0].TriggerKind)
	require.Len(t, *senderCalls, 1)
	assert.Equal(t, "[LIVE-P0] CPUHigh", (*senderCalls)[0].title)
}

func TestDeliveryServiceRecoveryRejectsNonFailedDelivery(t *testing.T) {
	db := newDeliveryRecoveryTestDB(t)
	service, _ := newRecoveryTestService(db)
	seedRecoveryData(t, db)

	original := seedFailedDeliveryRecord(t, db, recoveryDeliverySeed{
		alertID:        "alert-nonfailed",
		traceID:        "trace-nonfailed",
		channelID:      21,
		channelName:    "ops-webhook",
		channelType:    "webhook",
		channelEnabled: true,
		renderedTitle:  "title",
		renderedBody:   "body",
		alertName:      "CPUHigh",
		source:         "prometheus",
		triggerKind:    models.TriggerKindPipeline,
	})
	require.NoError(t, db.Model(&models.NotificationDelivery{}).Where("id = ?", original.ID).UpdateColumn("delivery_status", models.DeliveryStatusDelivered).Error)

	result, err := service.RetryDelivery(context.Background(), RetryDeliveryInput{
		OriginalDeliveryID: original.ID,
		Reason:             "should reject",
		ActorUserID:        103,
		ActorUsername:      "operator-c",
		ActorRole:          "operator",
	})
	require.Error(t, err)
	require.NotNil(t, result)
	assert.Equal(t, models.RecoveryStatusRejected, result.Recovery.Status)
	assert.Contains(t, result.Recovery.ErrorMessage, "only failed deliveries can be recovered")
	assertRecoveryCount(t, db, original.ID, 1)
}

func TestDeliveryServiceRecoveryRejectsDuplicatePendingRecovery(t *testing.T) {
	db := newDeliveryRecoveryTestDB(t)
	service, _ := newRecoveryTestService(db)
	seedRecoveryData(t, db)

	original := seedFailedDeliveryRecord(t, db, recoveryDeliverySeed{
		alertID:        "alert-duplicate",
		traceID:        "trace-duplicate",
		channelID:      21,
		channelName:    "ops-webhook",
		channelType:    "webhook",
		channelEnabled: true,
		renderedTitle:  "title",
		renderedBody:   "body",
		alertName:      "CPUHigh",
		source:         "prometheus",
		triggerKind:    models.TriggerKindPipeline,
	})
	require.NoError(t, db.Create(&models.NotificationDeliveryRecovery{
		OriginalDeliveryID: original.ID,
		Action:             models.TriggerKindRetry,
		Reason:             "existing in progress",
		ActorUserID:        104,
		ActorUsername:      "operator-d",
		ActorRole:          "operator",
		Status:             models.RecoveryStatusPending,
		RequestedAt:        time.Now(),
	}).Error)

	result, err := service.RetryDelivery(context.Background(), RetryDeliveryInput{
		OriginalDeliveryID: original.ID,
		Reason:             "duplicate request",
		ActorUserID:        104,
		ActorUsername:      "operator-d",
		ActorRole:          "operator",
	})
	require.Error(t, err)
	require.NotNil(t, result)
	assert.Equal(t, models.RecoveryStatusRejected, result.Recovery.Status)
	assert.Contains(t, result.Recovery.ErrorMessage, "recovery already in progress")
	assertRecoveryCount(t, db, original.ID, 2)
}

func newDeliveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.NotificationDelivery{}, &models.NotificationDeliveryAttempt{}))

	return db
}

func newDeliveryRecoveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + t.Name() + "-recovery?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Alert{},
		&models.DataSource{},
		&models.Channel{},
		&models.RouteRule{},
		&models.NotificationDelivery{},
		&models.NotificationDeliveryAttempt{},
		&models.NotificationDeliveryRecovery{},
	))

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

type recoverySenderCall struct {
	channelID uint
	title     string
	content   string
}

type recoveryDeliverySeed struct {
	alertID        string
	traceID        string
	channelID      uint
	channelName    string
	channelType    string
	channelEnabled bool
	routeRuleID    *uint
	routeRuleName  string
	renderedTitle  string
	renderedBody   string
	alertName      string
	source         string
	triggerKind    string
}

func newRecoveryTestService(db *gorm.DB) (*Service, *[]recoverySenderCall) {
	var senderCalls []recoverySenderCall
	service := NewService(db)
	service.sendToChannel = func(channel *models.Channel, title, content string, data map[string]interface{}) error {
		senderCalls = append(senderCalls, recoverySenderCall{
			channelID: channel.ID,
			title:     title,
			content:   content,
		})
		return nil
	}
	service.sleep = func(time.Duration) {}
	return service, &senderCalls
}

func seedRecoveryData(t *testing.T, db *gorm.DB) {
	t.Helper()

	alert := &models.Alert{
		AlertID:     "alert-replay",
		TraceID:     "trace-replay",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P0",
		Message:     "cpu critical",
		Status:      "firing",
		Fingerprint: "fp-replay",
		TriggerTime: time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC),
		ReceivedAt:  time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC),
		Raw:         datatypes.JSON(`{"severity":"critical","labels":{"instance":"game-01"}}`),
		Labels:      datatypes.JSON(`{"instance":"game-01"}`),
	}
	require.NoError(t, db.Create(alert).Error)

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  `{"alert_id":"{{.alert_id}}"}`,
		OutputTemplate: `{"title":"[LIVE-{{.severity}}] {{.alert_name}}","content":"route={{.route_name}}|event={{.severity_raw}}|message={{.message}}"}`,
		Enabled:        true,
		GroupByLabels:  datatypes.JSON("[]"),
	}).Error)

	require.NoError(t, db.Create(&models.Channel{
		ID:      21,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com/hook"}`),
		Enabled: true,
	}).Error)

	require.NoError(t, db.Create(&models.RouteRule{
		ID:            31,
		Name:          "live-route",
		Priority:      1,
		Sources:       datatypes.JSON(`["prometheus"]`),
		Severities:    datatypes.JSON(`["P0"]`),
		LabelMatchers: datatypes.JSON(`[]`),
		ChannelIDs:    datatypes.JSON(`[21]`),
		TimeRanges:    datatypes.JSON(`[]`),
		Enabled:       true,
	}).Error)
}

func seedFailedDeliveryRecord(t *testing.T, db *gorm.DB, seed recoveryDeliverySeed) *models.NotificationDelivery {
	t.Helper()

	routeSnapshot := datatypes.JSON([]byte("null"))
	if seed.routeRuleID != nil {
		routeSnapshot = mustJSONDelivery(t, models.RouteSnapshot{
			ID:         *seed.routeRuleID,
			Name:       seed.routeRuleName,
			Priority:   5,
			Enabled:    true,
			ChannelIDs: []uint{seed.channelID},
		})
	}

	deliveryRecord := &models.NotificationDelivery{
		AlertID:        seed.alertID,
		TraceID:        seed.traceID,
		ChannelID:      seed.channelID,
		RouteRuleID:    seed.routeRuleID,
		DeliveryStatus: models.DeliveryStatusFailed,
		DeliveryMode:   models.DeliveryModeRendered,
		AttemptCount:   1,
		AlertSnapshot: mustJSONDelivery(t, models.AlertSnapshot{
			AlertID:   seed.alertID,
			TraceID:   seed.traceID,
			Source:    seed.source,
			AlertName: seed.alertName,
			Severity:  "P0",
			Message:   "cpu critical",
			Status:    "firing",
		}),
		ChannelSnapshot: mustJSONDelivery(t, models.ChannelSnapshot{
			ID:      seed.channelID,
			Name:    seed.channelName,
			Type:    seed.channelType,
			Enabled: seed.channelEnabled,
		}),
		RouteSnapshot: routeSnapshot,
		RenderedPayloadSnapshot: mustJSONDelivery(t, models.RenderedPayloadSnapshot{
			Title:   seed.renderedTitle,
			Content: seed.renderedBody,
		}),
		FinalFailureSummary: mustJSONDelivery(t, models.FinalFailureSummary{
			Result:       models.AttemptResultFailed,
			Retryable:    true,
			ErrorMessage: "boom",
			AttemptCount: 1,
			TriggerKind:  seed.triggerKind,
		}),
	}
	require.NoError(t, db.Create(deliveryRecord).Error)
	require.NoError(t, db.Create(&models.NotificationDeliveryAttempt{
		DeliveryID:    deliveryRecord.ID,
		AttemptNumber: 1,
		Result:        models.AttemptResultFailed,
		Retryable:     true,
		ErrorMessage:  "boom",
		TriggerKind:   seed.triggerKind,
	}).Error)
	return deliveryRecord
}

func mustJSONDelivery(t *testing.T, value interface{}) datatypes.JSON {
	t.Helper()

	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return datatypes.JSON(raw)
}

func decodeRenderedPayload(t *testing.T, raw datatypes.JSON) models.RenderedPayloadSnapshot {
	t.Helper()

	var payload models.RenderedPayloadSnapshot
	require.NoError(t, json.Unmarshal(raw, &payload))
	return payload
}

func assertRecoveryCount(t *testing.T, db *gorm.DB, originalDeliveryID uint, expected int64) {
	t.Helper()

	var count int64
	require.NoError(t, db.Model(&models.NotificationDeliveryRecovery{}).
		Where("original_delivery_id = ?", originalDeliveryID).
		Count(&count).Error)
	assert.Equal(t, expected, count)
}

func uintPtr(v uint) *uint {
	return &v
}
