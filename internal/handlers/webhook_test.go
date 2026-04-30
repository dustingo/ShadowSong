package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/notifier"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWebhookHandlerHandleWebhookTraceSharedAcrossNewAlerts(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	require.NoError(t, db.Create(&models.DataSource{
		Name:              "trace-source",
		DisplayName:       "Trace Source",
		APIKey:            "key",
		DeduplicateWindow: 3600,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "trace-source",
			"status": "firing"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		GroupByLabels:  datatypes.JSON(`["alert_name","severity","source"]`),
		Enabled:        true,
	}).Error)

	payload := []map[string]interface{}{
		{
			"external_id": "external-1",
			"alert_name":  "CPUHigh",
			"severity":    "warning",
			"message":     "cpu high on game-01",
			"trace_id":    "caller-supplied-a",
		},
		{
			"external_id": "external-2",
			"alert_name":  "DiskHigh",
			"severity":    "warning",
			"message":     "disk high on game-02",
			"trace_id":    "caller-supplied-b",
		},
	}

	recorder := performWebhookRequest(t, handler, "trace-source", "key", payload)
	require.Equal(t, http.StatusOK, recorder.Code)

	var alerts []models.Alert
	require.NoError(t, db.Order("alert_id asc").Find(&alerts).Error)
	require.Len(t, alerts, 2)

	firstTraceID := alerts[0].TraceID
	require.NotEmpty(t, firstTraceID)
	assert.Equal(t, firstTraceID, alerts[1].TraceID)
	assert.NotEqual(t, "caller-supplied-a", alerts[0].TraceID)
	assert.NotEqual(t, "caller-supplied-b", alerts[1].TraceID)
	assert.Equal(t, "external-1", alerts[0].AlertID)
	assert.Equal(t, "external-2", alerts[1].AlertID)
	assert.Equal(t, computeCurrentFingerprintContract(handler, alerts[0]), alerts[0].Fingerprint)
	assert.Equal(t, computeCurrentFingerprintContract(handler, alerts[1]), alerts[1].Fingerprint)
}

func TestWebhookHandlerHandleWebhookTraceIgnoresCallerSuppliedValue(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	require.NoError(t, db.Create(&models.DataSource{
		Name:        "trace-ignore",
		DisplayName: "Trace Ignore",
		APIKey:      "key",
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"trace_id": "{{.trace_id}}",
			"source": "trace-ignore",
			"status": "firing"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		Enabled:        true,
	}).Error)

	payload := map[string]interface{}{
		"external_id": "external-3",
		"alert_name":  "MemoryHigh",
		"severity":    "critical",
		"message":     "memory high",
		"trace_id":    "caller-controlled-trace",
	}

	recorder := performWebhookRequest(t, handler, "trace-ignore", "key", payload)
	require.Equal(t, http.StatusOK, recorder.Code)

	var alert models.Alert
	require.NoError(t, db.First(&alert, "alert_id = ?", "external-3").Error)
	require.NotEmpty(t, alert.TraceID)
	assert.NotEqual(t, "caller-controlled-trace", alert.TraceID)
}

func TestWebhookHandlerHandleWebhookDedupLogsRequestTraceContext(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)

	require.NoError(t, db.Create(&models.DataSource{
		Name:               "trace-dedup",
		DisplayName:        "Trace Dedup",
		APIKey:             "key",
		DeduplicateWindow:  3600,
		DeduplicateEnabled: true,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "trace-dedup",
			"status": "firing"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		Enabled:        true,
	}).Error)

	payload := map[string]interface{}{
		"external_id": "repeat-1",
		"alert_name":  "CPUHigh",
		"severity":    "warning",
		"message":     "cpu high",
	}

	first := performWebhookRequest(t, handler, "trace-dedup", "key", payload)
	require.Equal(t, http.StatusOK, first.Code)
	logBuffer.Reset()

	second := performWebhookRequest(t, handler, "trace-dedup", "key", payload)
	require.Equal(t, http.StatusOK, second.Code)

	var alerts []models.Alert
	require.NoError(t, db.Find(&alerts).Error)
	require.Len(t, alerts, 1)
	assert.Equal(t, "repeat-1", alerts[0].AlertID)
	assert.Equal(t, 2, alerts[0].TriggerCount)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=dedup")
	assert.Contains(t, logOutput, "trace_id=")
	assert.Contains(t, logOutput, "existing_alert_id=repeat-1")
	assert.Contains(t, logOutput, "fingerprint="+alerts[0].Fingerprint)
}

func TestWebhookHandlerPublishToRedisTracePayload(t *testing.T) {
	handler, _ := newWebhookTestHandler(nil)
	var captured map[string]interface{}
	handler.redisXAdd = func(_ context.Context, args *redis.XAddArgs) *redis.StringCmd {
		captured = args.Values.(map[string]interface{})
		return redis.NewStringResult("1-0", nil)
	}

	handler.publishToRedis([]models.Alert{{
		AlertID:     "alert-redis",
		TraceID:     "trace-redis",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-1",
		Status:      "firing",
		TriggerTime: time.Unix(1710000000, 0),
	}})

	require.NotNil(t, captured)
	assert.Equal(t, "trace-redis", captured["trace_id"])
	assert.Equal(t, "alert-redis", captured["alert_id"])
}

func TestWebhookHandlerProcessAlertNotificationsTraceLogging(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		return fmt.Errorf("send exploded")
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)
	require.NoError(t, db.Create(&models.Channel{
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&models.RouteRule{
		Name:       "all",
		Priority:   1,
		ChannelIDs: datatypes.JSON(`[11]`),
		Enabled:    true,
	}).Error)

	handler.processAlertNotificationsAsync([]models.Alert{{
		AlertID:     "alert-trace",
		TraceID:     "trace-notify",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-notify",
		Status:      "firing",
	}})

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "trace_id=trace-notify")
	assert.Contains(t, logOutput, "alert_id=alert-trace")
}

func TestWebhookHandlerLogsLifecycleStages(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		return nil
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:              "lifecycle-source",
		DisplayName:       "Lifecycle Source",
		APIKey:            "key",
		DeduplicateWindow: 3600,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "lifecycle-source",
			"status": "firing"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		Enabled:        true,
	}).Error)
	require.NoError(t, db.Create(&models.Channel{
		ID:      9,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&models.RouteRule{
		Name:       "all",
		Priority:   1,
		Sources:    datatypes.JSON(`["lifecycle-source"]`),
		ChannelIDs: datatypes.JSON(`[9]`),
		Enabled:    true,
	}).Error)

	recorder := performWebhookRequest(t, handler, "lifecycle-source", "key", map[string]interface{}{
		"external_id": "lifecycle-1",
		"alert_name":  "CPUHigh",
		"severity":    "warning",
		"message":     "cpu high on game-01",
	})
	require.Equal(t, http.StatusOK, recorder.Code)

	var alert models.Alert
	require.NoError(t, db.First(&alert, "alert_id = ?", "lifecycle-1").Error)
	require.NotEmpty(t, alert.TraceID)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=ingest")
	assert.Contains(t, logOutput, "stage=persist")
	assert.Contains(t, logOutput, "stage=redis_publish")
	assert.Contains(t, logOutput, "stage=route_match")
	assert.Contains(t, logOutput, "stage=notification_entry")
	assert.GreaterOrEqual(t, strings.Count(logOutput, "trace_id="+alert.TraceID), 5)
	assert.Contains(t, logOutput, "redis_stream=alerts:pending")
	assert.Contains(t, logOutput, "redis_message_id=1-0")
}

func TestWebhookHandlerRedisPublishFailure(t *testing.T) {
	handler, logBuffer := newWebhookTestHandler(nil)
	handler.redisXAdd = func(_ context.Context, _ *redis.XAddArgs) *redis.StringCmd {
		return redis.NewStringResult("", fmt.Errorf("redis unavailable"))
	}

	handler.publishToRedis([]models.Alert{{
		AlertID:     "alert-redis-fail",
		TraceID:     "trace-redis-fail",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-redis-fail",
		Status:      "firing",
		TriggerTime: time.Unix(1710000000, 0),
	}})

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=redis_publish")
	assert.Contains(t, logOutput, "trace_id=trace-redis-fail")
	assert.Contains(t, logOutput, "redis_stream=alerts:pending")
	assert.Contains(t, logOutput, "failed err=redis unavailable")
}

func TestWebhookHandlerRedisPublishFailureLeavesLifecycleInspectable(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	handler.redisXAdd = func(_ context.Context, _ *redis.XAddArgs) *redis.StringCmd {
		return redis.NewStringResult("", fmt.Errorf("redis unavailable"))
	}
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		return nil
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:              "redis-failure-source",
		DisplayName:       "Redis Failure Source",
		APIKey:            "key",
		DeduplicateWindow: 3600,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "redis-failure-source",
			"status": "firing"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		Enabled:        true,
	}).Error)
	require.NoError(t, db.Create(&models.Channel{
		ID:      10,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&models.RouteRule{
		Name:       "all",
		Priority:   1,
		Sources:    datatypes.JSON(`["redis-failure-source"]`),
		ChannelIDs: datatypes.JSON(`[10]`),
		Enabled:    true,
	}).Error)

	recorder := performWebhookRequest(t, handler, "redis-failure-source", "key", map[string]interface{}{
		"external_id": "redis-failure-1",
		"alert_name":  "CPUHigh",
		"severity":    "warning",
		"message":     "cpu high on game-01",
	})
	require.Equal(t, http.StatusOK, recorder.Code)

	var alert models.Alert
	require.NoError(t, db.First(&alert, "alert_id = ?", "redis-failure-1").Error)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=redis_publish")
	assert.Contains(t, logOutput, "failed err=redis unavailable")
	assert.Contains(t, logOutput, "stage=route_match")
	assert.Contains(t, logOutput, "stage=notification_entry")
	assert.GreaterOrEqual(t, strings.Count(logOutput, "trace_id="+alert.TraceID), 4)
}

func TestWebhookHandlerDedupTrace(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)

	require.NoError(t, db.Create(&models.DataSource{
		Name:               "dedup-trace",
		DisplayName:        "Dedup Trace",
		APIKey:             "key",
		DeduplicateWindow:  3600,
		DeduplicateEnabled: true,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "dedup-trace",
			"status": "firing"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		Enabled:        true,
	}).Error)

	payload := map[string]interface{}{
		"external_id": "repeat-2",
		"alert_name":  "CPUHigh",
		"severity":    "warning",
		"message":     "cpu high",
	}

	first := performWebhookRequest(t, handler, "dedup-trace", "key", payload)
	require.Equal(t, http.StatusOK, first.Code)
	logBuffer.Reset()

	second := performWebhookRequest(t, handler, "dedup-trace", "key", payload)
	require.Equal(t, http.StatusOK, second.Code)

	var alert models.Alert
	require.NoError(t, db.First(&alert, "alert_id = ?", "repeat-2").Error)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=dedup")
	assert.Contains(t, logOutput, "trace_id=")
	assert.Contains(t, logOutput, "existing_alert_id=repeat-2")
	assert.Contains(t, logOutput, "fingerprint="+alert.Fingerprint)
}

func TestWebhookHandler_mapSeverity_KeepPLevels(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "keep p0",
			input: "P0",
			want:  "P0",
		},
		{
			name:  "keep p1",
			input: "P1",
			want:  "P1",
		},
		{
			name:  "keep p2 lowercase",
			input: "p2",
			want:  "P2",
		},
		{
			name:  "map warning text",
			input: "warning",
			want:  "P1",
		},
	}

	handler := &WebhookHandler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, handler.mapSeverity(tt.input))
		})
	}
}

func TestWebhookHandler_renderNotification(t *testing.T) {
	handler := &WebhookHandler{}
	alert := &models.Alert{
		AlertID:     "alert-1",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "CPU usage above 90%",
		Source:      "prometheus",
		Status:      "firing",
		TriggerTime: time.Date(2026, 4, 10, 7, 0, 0, 0, time.UTC),
		Labels:      datatypes.JSON(`{"instance":"game-01","env":"prod"}`),
		Raw:         datatypes.JSON(`{"summary":"cpu high","annotations":{"runbook":"https://runbook.local/cpu","owner":"ops"},"custom_field":"raw-value"}`),
	}

	tests := []struct {
		name        string
		template    string
		wantTitle   string
		wantContent string
		wantContext map[string]interface{}
	}{
		{
			name:        "keeps legacy top level fields",
			template:    `{"title":"[{{.severity}}] {{.alert_name}}","content":"{{.message}} from {{index .labels "instance"}}"}`,
			wantTitle:   "[P1] CPUHigh",
			wantContent: "CPU usage above 90% from game-01",
		},
		{
			name:        "exposes raw event payload",
			template:    `{"title":"{{default .event.summary .alert_name}}","content":"runbook={{.event.annotations.runbook}} owner={{.event.annotations.owner}} custom={{.event.custom_field}}"}`,
			wantTitle:   "cpu high",
			wantContent: "runbook=https://runbook.local/cpu owner=ops custom=raw-value",
		},
		{
			name:        "missing raw payload degrades safely",
			template:    `{"title":"{{.alert_name}}","content":"{{default .event.annotations.runbook "n/a"}}|{{default .message "missing"}}"}`,
			wantTitle:   "CPUHigh",
			wantContent: "n/a|CPU usage above 90%",
			wantContext: map[string]interface{}{"raw": "{}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAlert := *alert
			if tt.wantContext != nil {
				testAlert.Raw = datatypes.JSON(tt.wantContext["raw"].(string))
			}

			title, content, renderContext, err := handler.renderNotificationPreview(&testAlert, tt.template)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTitle, title)
			assert.Equal(t, tt.wantContent, content)
			assert.Contains(t, renderContext, "event")
			assert.Contains(t, renderContext, "alert")
			assert.Contains(t, renderContext, "alert_name")
			assert.Contains(t, renderContext, "labels")
			assert.Contains(t, renderContext, "severity_code")
			assert.Contains(t, renderContext, "severity_raw")
			assert.Equal(t, "P1", renderContext["severity_code"])
			assert.Equal(t, "P1", renderContext["severity"])
			if tt.name == "exposes raw event payload" {
				assert.Equal(t, "cpu high", renderContext["event"].(map[string]interface{})["summary"])
			}
		})
	}
}

func TestWebhookHandler_buildNotificationRenderContext_ExposesSeverityAliases(t *testing.T) {
	handler := &WebhookHandler{}
	alert := &models.Alert{
		AlertID:     "alert-2",
		AlertName:   "CPUHigh",
		Severity:    "P0",
		Message:     "CPU usage above 95%",
		Source:      "prometheus",
		Status:      "firing",
		TriggerTime: time.Date(2026, 4, 10, 7, 0, 0, 0, time.UTC),
		Labels:      datatypes.JSON(`{"instance":"game-02"}`),
		Raw:         datatypes.JSON(`{"severity":"critical","labels":{"severity":"critical"}}`),
	}

	renderContext := handler.buildNotificationRenderContext(alert)
	alertContext := renderContext["alert"].(map[string]interface{})

	assert.Equal(t, "P0", renderContext["severity"])
	assert.Equal(t, "P0", renderContext["severity_code"])
	assert.Equal(t, "critical", renderContext["severity_raw"])
	assert.Equal(t, "P0", alertContext["severity"])
	assert.Equal(t, "P0", alertContext["severity_code"])
	assert.Equal(t, "critical", alertContext["severity_raw"])
}

func TestMarshalRawAlertData_PrefersIndividualAlertObject(t *testing.T) {
	rawAlertBody := marshalRawAlertData(map[string]interface{}{
		"summary": "array item summary",
		"annotations": map[string]interface{}{
			"runbook": "https://runbook.local/array-item",
		},
	}, []byte(`[{"summary":"wrong"}]`))

	assert.JSONEq(t, `{"summary":"array item summary","annotations":{"runbook":"https://runbook.local/array-item"}}`, string(rawAlertBody))
	assert.Equal(t, "array item summary", decodeJSONMap(rawAlertBody)["summary"])
}

func TestWebhookHandlerProcessAlertNotificationsAsync_RecoversFromPanic(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		panic("boom")
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)
	require.NoError(t, db.Create(&models.Channel{
		ID:      11,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&models.RouteRule{
		Name:       "all",
		Priority:   1,
		ChannelIDs: datatypes.JSON(`[11]`),
		Enabled:    true,
	}).Error)

	handler.processAlertNotificationsAsync([]models.Alert{{
		AlertID:     "alert-panic",
		TraceID:     "trace-panic",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-panic",
		Status:      "firing",
	}})

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=async_panic")
	assert.Contains(t, logOutput, "recovered panic=boom")

	asyncPanicLine := findWebhookLogLine(logOutput, "stage=async_panic")
	require.NotEmpty(t, asyncPanicLine)
	assertWebhookLogFields(t, asyncPanicLine, map[string]string{
		"stage":        "async_panic",
		"trace_id":     "trace-panic",
		"alert_id":     "alert-panic",
		"fingerprint":  "fp-panic",
		"source":       "prometheus",
		"channel_id":   "11",
		"channel_name": "ops-webhook",
		"channel_type": "webhook",
	})
}

func TestWebhookHandlerSendNotification_LogsAlertAndChannelContext(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		return fmt.Errorf("send exploded")
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)

	alert := &models.Alert{
		AlertID:   "alert-log",
		Source:    "prometheus",
		AlertName: "CPUHigh",
		Severity:  "P1",
		Message:   "cpu high",
		Status:    "firing",
	}
	channel := &models.Channel{
		ID:      42,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "stage=send_notification")
	assert.Contains(t, logOutput, "alert_id=alert-log")
	assert.Contains(t, logOutput, "source=prometheus")
	assert.Contains(t, logOutput, "channel_id=42")
	assert.Contains(t, logOutput, "channel_name=ops-webhook")
	assert.Contains(t, logOutput, "failed to send rendered notification")
}

func TestWebhookHandlerSendNotification_ImmediateSuccessSingleAttempt(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	attempts := 0
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		attempts++
		return nil
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)

	alert := &models.Alert{
		AlertID:     "alert-immediate-success",
		TraceID:     "trace-immediate-success",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-immediate-success",
		Status:      "firing",
	}
	channel := &models.Channel{
		ID:      41,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Equal(t, 1, attempts)
	assert.Equal(t, 1, strings.Count(logOutput, "stage=send_attempt"))
	assert.Equal(t, 1, strings.Count(logOutput, "stage=send_notification"))
	assert.Contains(t, logOutput, "trace_id=trace-immediate-success")
	assert.Contains(t, logOutput, "alert_id=alert-immediate-success")
	assert.Contains(t, logOutput, "channel_id=41")
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "max_attempts=")
	assert.NotContains(t, logOutput, "stage=terminal_failure")

	deliveryRecord, attemptRecords := requireSingleDeliveryLedger(t, db, "alert-immediate-success")
	assert.Equal(t, "trace-immediate-success", deliveryRecord.TraceID)
	assert.Equal(t, channel.ID, deliveryRecord.ChannelID)
	assert.Equal(t, models.DeliveryModeRendered, deliveryRecord.DeliveryMode)
	assert.Equal(t, models.DeliveryStatusDelivered, deliveryRecord.DeliveryStatus)
	assert.Equal(t, 1, deliveryRecord.AttemptCount)
	assertDeliveryHasNoFailureSummary(t, deliveryRecord.FinalFailureSummary)
	renderedPayload := decodeRenderedPayloadSnapshot(t, deliveryRecord.RenderedPayloadSnapshot)
	assert.Equal(t, "CPUHigh", renderedPayload.Title)
	assert.Equal(t, "cpu high", renderedPayload.Content)
	require.Len(t, attemptRecords, 1)
	assert.Equal(t, 1, attemptRecords[0].AttemptNumber)
	assert.Equal(t, models.AttemptResultSuccess, attemptRecords[0].Result)
	assert.False(t, attemptRecords[0].Retryable)
	assert.Equal(t, models.TriggerKindPipeline, attemptRecords[0].TriggerKind)
}

func TestWebhookHandlerSendNotification_RetrySuccessAfterTransientFailures(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	attempts := 0
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("channel %d (%s) send failed: webhook notification failed with status: 503", channel.ID, channel.Name)
		}
		return nil
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)

	alert := &models.Alert{
		AlertID:     "alert-retry-success",
		TraceID:     "trace-retry-success",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-retry-success",
		Status:      "firing",
	}
	channel := &models.Channel{
		ID:      42,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Equal(t, 3, attempts)
	assert.GreaterOrEqual(t, strings.Count(logOutput, "stage=send_attempt"), 3)
	assert.Contains(t, logOutput, "trace_id=trace-retry-success")
	assert.Contains(t, logOutput, "alert_id=alert-retry-success")
	assert.Contains(t, logOutput, "channel_id=42")
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "attempt=2")
	assert.Contains(t, logOutput, "attempt=3")
	assert.Contains(t, logOutput, "max_attempts=")
	assert.Contains(t, logOutput, "error=")
	assert.Contains(t, logOutput, "stage=send_notification")
	assert.NotContains(t, logOutput, "stage=terminal_failure")

	sendAttemptLine := findWebhookLogLine(logOutput, "stage=send_attempt")
	sendAttemptFields := parseWebhookLogFields(sendAttemptLine)
	assert.Equal(t, "send_attempt", sendAttemptFields["stage"])
	assert.Equal(t, "trace-retry-success", sendAttemptFields["trace_id"])
	assert.Equal(t, "alert-retry-success", sendAttemptFields["alert_id"])
	assert.Equal(t, "42", sendAttemptFields["channel_id"])
	assert.Equal(t, "ops-webhook", sendAttemptFields["channel_name"])
	assert.Equal(t, "webhook", sendAttemptFields["channel_type"])
	assert.Equal(t, "rendered", sendAttemptFields["mode"])
	assert.Equal(t, "1", sendAttemptFields["attempt"])
	assert.Equal(t, fmt.Sprintf("%d", notificationMaxAttempts), sendAttemptFields["max_attempts"])
	assert.NotEmpty(t, sendAttemptFields["error"])
	assert.Contains(t, sendAttemptLine, "error=")
}

func TestWebhookHandlerSendNotification_DatasourceLookupFailureFallsBackIntoRetryBoundary(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	attempts := 0
	var titles []string
	var contents []string
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		attempts++
		titles = append(titles, title)
		contents = append(contents, content)
		if attempts < notificationMaxAttempts {
			return fmt.Errorf("channel %d (%s) send failed: webhook notification failed with status: 503", channel.ID, channel.Name)
		}
		return nil
	}

	alert := &models.Alert{
		AlertID:     "alert-datasource-fallback",
		TraceID:     "trace-datasource-fallback",
		Source:      "missing-source",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-datasource-fallback",
		Status:      "firing",
	}
	channel := &models.Channel{
		ID:      45,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Equal(t, notificationMaxAttempts, attempts)
	assert.Equal(t, notificationMaxAttempts, strings.Count(logOutput, "stage=send_attempt"))
	assert.Contains(t, logOutput, "stage=datasource_lookup")
	assert.Contains(t, logOutput, "mode=default")
	assert.Contains(t, logOutput, "trace_id=trace-datasource-fallback")
	assert.Contains(t, logOutput, "alert_id=alert-datasource-fallback")
	assert.Contains(t, logOutput, "channel_id=45")
	assert.Contains(t, titles, "[P1] CPUHigh")
	assert.Contains(t, contents, "cpu high")
	assert.NotContains(t, logOutput, "stage=terminal_failure")

	datasourceLookupLine := findWebhookLogLine(logOutput, "stage=datasource_lookup")
	datasourceLookupFields := parseWebhookLogFields(datasourceLookupLine)
	assert.Equal(t, "datasource_lookup", datasourceLookupFields["stage"])
	assert.Equal(t, "trace-datasource-fallback", datasourceLookupFields["trace_id"])
	assert.Equal(t, "alert-datasource-fallback", datasourceLookupFields["alert_id"])
	assert.Equal(t, "45", datasourceLookupFields["channel_id"])
	assert.Equal(t, "ops-webhook", datasourceLookupFields["channel_name"])
	assert.Equal(t, "webhook", datasourceLookupFields["channel_type"])
	assert.Equal(t, "default", datasourceLookupFields["mode"])

	sendAttemptLine := findWebhookLogLine(logOutput, "stage=send_attempt")
	sendAttemptFields := parseWebhookLogFields(sendAttemptLine)
	assert.Equal(t, "send_attempt", sendAttemptFields["stage"])
	assert.Equal(t, "trace-datasource-fallback", sendAttemptFields["trace_id"])
	assert.Equal(t, "alert-datasource-fallback", sendAttemptFields["alert_id"])
	assert.Equal(t, "45", sendAttemptFields["channel_id"])
	assert.Equal(t, "ops-webhook", sendAttemptFields["channel_name"])
	assert.Equal(t, "webhook", sendAttemptFields["channel_type"])
	assert.Equal(t, "default", sendAttemptFields["mode"])
	assert.Equal(t, "1", sendAttemptFields["attempt"])
	assert.Equal(t, fmt.Sprintf("%d", notificationMaxAttempts), sendAttemptFields["max_attempts"])
	assert.NotEmpty(t, sendAttemptFields["error"])
	assert.Contains(t, sendAttemptLine, "error=")

	deliveryRecord, attemptRecords := requireSingleDeliveryLedger(t, db, "alert-datasource-fallback")
	assert.Equal(t, "trace-datasource-fallback", deliveryRecord.TraceID)
	assert.Equal(t, models.DeliveryModeDefault, deliveryRecord.DeliveryMode)
	assert.Equal(t, models.DeliveryStatusDelivered, deliveryRecord.DeliveryStatus)
	assert.Equal(t, notificationMaxAttempts, deliveryRecord.AttemptCount)
	assertDeliveryHasNoFailureSummary(t, deliveryRecord.FinalFailureSummary)
	renderedPayload := decodeRenderedPayloadSnapshot(t, deliveryRecord.RenderedPayloadSnapshot)
	assert.Equal(t, "[P1] CPUHigh", renderedPayload.Title)
	assert.Equal(t, "cpu high", renderedPayload.Content)
	require.Len(t, attemptRecords, notificationMaxAttempts)
	assert.Equal(t, 1, attemptRecords[0].AttemptNumber)
	assert.Equal(t, models.AttemptResultFailed, attemptRecords[0].Result)
	assert.True(t, attemptRecords[0].Retryable)
	assert.Equal(t, notificationMaxAttempts, attemptRecords[len(attemptRecords)-1].AttemptNumber)
	assert.Equal(t, models.AttemptResultSuccess, attemptRecords[len(attemptRecords)-1].Result)
}

func TestWebhookHandlerSendNotification_RenderFailureFallsBackIntoRetryBoundary(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	attempts := 0
	var titles []string
	var contents []string
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		attempts++
		titles = append(titles, title)
		contents = append(contents, content)
		if attempts < notificationMaxAttempts {
			return fmt.Errorf("channel %d (%s) send failed: webhook notification failed with status: 503", channel.ID, channel.Name)
		}
		return nil
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{{`,
	}).Error)

	alert := &models.Alert{
		AlertID:     "alert-render-fallback",
		TraceID:     "trace-render-fallback",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-render-fallback",
		Status:      "firing",
	}
	channel := &models.Channel{
		ID:      46,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Equal(t, notificationMaxAttempts, attempts)
	assert.Equal(t, notificationMaxAttempts, strings.Count(logOutput, "stage=send_attempt"))
	assert.Contains(t, logOutput, "stage=render_notification")
	assert.Contains(t, logOutput, "mode=default")
	assert.Contains(t, logOutput, "trace_id=trace-render-fallback")
	assert.Contains(t, logOutput, "alert_id=alert-render-fallback")
	assert.Contains(t, logOutput, "channel_id=46")
	assert.Contains(t, titles, "[P1] CPUHigh")
	assert.Contains(t, contents, "cpu high")
	assert.NotContains(t, logOutput, "stage=terminal_failure")

	renderLine := findWebhookLogLine(logOutput, "stage=render_notification")
	renderFields := parseWebhookLogFields(renderLine)
	assert.Equal(t, "render_notification", renderFields["stage"])
	assert.Equal(t, "trace-render-fallback", renderFields["trace_id"])
	assert.Equal(t, "alert-render-fallback", renderFields["alert_id"])
	assert.Equal(t, "46", renderFields["channel_id"])
	assert.Equal(t, "ops-webhook", renderFields["channel_name"])
	assert.Equal(t, "webhook", renderFields["channel_type"])
	assert.Equal(t, "default", renderFields["mode"])
}

func TestWebhookHandlerSendNotification_NonRetryableFailureStopsAfterFirstAttempt(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	attempts := 0
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		attempts++
		return fmt.Errorf("channel %d (%s) sender init failed: webhook url is required", channel.ID, channel.Name)
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)

	alert := &models.Alert{
		AlertID:     "alert-non-retryable",
		TraceID:     "trace-non-retryable",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-non-retryable",
		Status:      "firing",
	}
	channel := &models.Channel{
		ID:      43,
		Name:    "ops-feishu",
		Type:    "feishu",
		Config:  datatypes.JSON(`{"webhook_url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Equal(t, 1, attempts)
	assert.Equal(t, 1, strings.Count(logOutput, "stage=send_attempt"))
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "max_attempts=")
	assert.Contains(t, logOutput, "error=")
	assert.Contains(t, logOutput, "stage=send_notification")
	assert.NotContains(t, logOutput, "stage=terminal_failure")
}

func TestWebhookHandlerSendNotification_RetryExhaustPersistsTerminalFailureLedger(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	attempts := 0
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		attempts++
		return fmt.Errorf("channel %d (%s) send failed: webhook notification failed with status: 503", channel.ID, channel.Name)
	}

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)

	alert := &models.Alert{
		AlertID:     "alert-terminal",
		TraceID:     "trace-terminal",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-terminal",
		Status:      "firing",
	}
	channel := &models.Channel{
		ID:      44,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}

	handler.sendNotification(alert, channel)

	logOutput := logBuffer.String()
	assert.Equal(t, notificationMaxAttempts, attempts)
	assert.Equal(t, notificationMaxAttempts, strings.Count(logOutput, "stage=send_attempt"))
	assert.Equal(t, 1, strings.Count(logOutput, "stage=terminal_failure"))
	assert.Contains(t, logOutput, "trace_id=trace-terminal")
	assert.Contains(t, logOutput, "alert_id=alert-terminal")
	assert.Contains(t, logOutput, "channel_id=44")
	assert.Contains(t, logOutput, "attempt=")
	assert.Contains(t, logOutput, "max_attempts=")
	assert.Contains(t, logOutput, "error=")

	terminalFailureLine := findWebhookLogLine(logOutput, "stage=terminal_failure")
	terminalFailureFields := parseWebhookLogFields(terminalFailureLine)
	assert.Equal(t, "terminal_failure", terminalFailureFields["stage"])
	assert.Equal(t, "trace-terminal", terminalFailureFields["trace_id"])
	assert.Equal(t, "alert-terminal", terminalFailureFields["alert_id"])
	assert.Equal(t, "44", terminalFailureFields["channel_id"])
	assert.Equal(t, "ops-webhook", terminalFailureFields["channel_name"])
	assert.Equal(t, "webhook", terminalFailureFields["channel_type"])
	assert.Equal(t, "rendered", terminalFailureFields["mode"])
	assert.Equal(t, fmt.Sprintf("%d", notificationMaxAttempts), terminalFailureFields["attempt"])
	assert.Equal(t, fmt.Sprintf("%d", notificationMaxAttempts), terminalFailureFields["max_attempts"])
	assert.NotEmpty(t, terminalFailureFields["error"])
	assert.Contains(t, terminalFailureLine, "error=")

	deliveryRecord, attemptRecords := requireSingleDeliveryLedger(t, db, "alert-terminal")
	assert.Equal(t, "trace-terminal", deliveryRecord.TraceID)
	assert.Equal(t, models.DeliveryModeRendered, deliveryRecord.DeliveryMode)
	assert.Equal(t, models.DeliveryStatusFailed, deliveryRecord.DeliveryStatus)
	assert.Equal(t, notificationMaxAttempts, deliveryRecord.AttemptCount)
	renderedPayload := decodeRenderedPayloadSnapshot(t, deliveryRecord.RenderedPayloadSnapshot)
	assert.Equal(t, "CPUHigh", renderedPayload.Title)
	assert.Equal(t, "cpu high", renderedPayload.Content)
	failureSummary := decodeFinalFailureSummary(t, deliveryRecord.FinalFailureSummary)
	assert.Equal(t, models.AttemptResultFailed, failureSummary.Result)
	assert.Equal(t, notificationMaxAttempts, failureSummary.AttemptCount)
	assert.True(t, failureSummary.Retryable)
	assert.Equal(t, models.TriggerKindPipeline, failureSummary.TriggerKind)
	assert.Contains(t, failureSummary.ErrorMessage, "retry budget exhausted")
	assert.Contains(t, failureSummary.ErrorMessage, "status: 503")
	require.Len(t, attemptRecords, notificationMaxAttempts)
	for idx, attemptRecord := range attemptRecords {
		assert.Equal(t, idx+1, attemptRecord.AttemptNumber)
		assert.Equal(t, models.AttemptResultFailed, attemptRecord.Result)
		assert.True(t, attemptRecord.Retryable)
		assert.Equal(t, models.TriggerKindPipeline, attemptRecord.TriggerKind)
	}
}

func TestWebhookHandlerProcessAlertNotifications_LogsMatchedChannelsAsStructuredField(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, logBuffer := newWebhookTestHandler(db)
	handler.sendToChannel = func(channel *models.Channel, title, content string) error {
		return nil
	}

	require.NoError(t, db.Create(&models.Channel{
		ID:      51,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&models.RouteRule{
		Name:       "all",
		Priority:   1,
		ChannelIDs: datatypes.JSON(`[51]`),
		Enabled:    true,
	}).Error)
	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		APIKey:         "key",
		InputTemplate:  "{}",
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
	}).Error)

	handler.processAlertNotifications([]models.Alert{{
		AlertID:     "alert-route-structured",
		TraceID:     "trace-route-structured",
		Source:      "prometheus",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "fp-route-structured",
		Status:      "firing",
	}})

	routeMatchLine := findWebhookLogLine(logBuffer.String(), "stage=route_match")
	routeMatchFields := parseWebhookLogFields(routeMatchLine)
	assert.Equal(t, "route_match", routeMatchFields["stage"])
	assert.Equal(t, "trace-route-structured", routeMatchFields["trace_id"])
	assert.Equal(t, "alert-route-structured", routeMatchFields["alert_id"])
	assert.Equal(t, "fp-route-structured", routeMatchFields["fingerprint"])
	assert.Equal(t, "prometheus", routeMatchFields["source"])
	assert.Equal(t, "1", routeMatchFields["matched_channels"])
}

func TestWebhookHandlerBaseAlertLogFields_IncludeAlertAndChannelContext(t *testing.T) {
	handler := &WebhookHandler{}
	alert := &models.Alert{
		TraceID:     "trace-1",
		AlertID:     "alert-1",
		Fingerprint: "fp-1",
		Source:      "prometheus",
	}
	channel := &models.Channel{
		ID:   7,
		Name: "ops-webhook",
		Type: "webhook",
	}

	fields := handler.baseAlertLogFields(alert, channel)

	assert.Equal(t, "trace-1", fields["trace_id"])
	assert.Equal(t, "alert-1", fields["alert_id"])
	assert.Equal(t, "fp-1", fields["fingerprint"])
	assert.Equal(t, "prometheus", fields["source"])
	assert.Equal(t, "7", fields["channel_id"])
	assert.Equal(t, "ops-webhook", fields["channel_name"])
	assert.Equal(t, "webhook", fields["channel_type"])
}

func TestWebhookHandlerLogAlertEvent_SkipsEmptyOptionalFieldsAndSortsStableKeys(t *testing.T) {
	buffer := &bytes.Buffer{}
	handler := &WebhookHandler{
		logger: log.New(buffer, "", 0),
	}

	handler.logAlertEvent("route_match", map[string]string{
		"trace_id":         "trace-1",
		"alert_id":         "alert-1",
		"fingerprint":      "fp-1",
		"source":           "prometheus",
		"matched_channels": "2",
		"channel_name":     "",
	}, "matched route rules")

	logLine := strings.TrimSpace(buffer.String())
	assert.Contains(t, logLine, "stage=route_match")
	assert.Contains(t, logLine, "trace_id=trace-1")
	assert.Contains(t, logLine, "alert_id=alert-1")
	assert.Contains(t, logLine, "fingerprint=fp-1")
	assert.Contains(t, logLine, "matched_channels=2")
	assert.NotContains(t, logLine, "channel_name=")

	assert.Less(t, strings.Index(logLine, "stage=route_match"), strings.Index(logLine, "alert_id=alert-1"))
	assert.Less(t, strings.Index(logLine, "alert_id=alert-1"), strings.Index(logLine, "fingerprint=fp-1"))
	assert.Less(t, strings.Index(logLine, "fingerprint=fp-1"), strings.Index(logLine, "matched_channels=2"))
	assert.Less(t, strings.Index(logLine, "matched_channels=2"), strings.Index(logLine, "source=prometheus"))
	assert.Less(t, strings.Index(logLine, "source=prometheus"), strings.Index(logLine, "trace_id=trace-1"))
}

func TestWebhookHandlerLogAlertEvent_PreservesSpaceContainingFieldValues(t *testing.T) {
	buffer := &bytes.Buffer{}
	handler := &WebhookHandler{
		logger: log.New(buffer, "", 0),
	}

	handler.logAlertEvent("send_attempt", map[string]string{
		"trace_id":     "trace-1",
		"alert_id":     "alert-1",
		"fingerprint":  "fp-1",
		"source":       "prometheus",
		"channel_id":   "7",
		"channel_name": "ops webhook primary",
		"error":        "dial tcp timeout exceeded",
	}, "notification attempt recorded")

	logLine := strings.TrimSpace(buffer.String())
	fields := parseWebhookLogFields(logLine)

	assert.Equal(t, "send_attempt", fields["stage"])
	assert.Equal(t, "ops webhook primary", fields["channel_name"])
	assert.Equal(t, "dial tcp timeout exceeded", fields["error"])
}

func newWebhookTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Alert{},
		&models.DataSource{},
		&models.Channel{},
		&models.RouteRule{},
		&models.NotificationDelivery{},
		&models.NotificationDeliveryAttempt{},
	))

	return db
}

func newWebhookTestHandler(db *gorm.DB) (*WebhookHandler, *bytes.Buffer) {
	buffer := &bytes.Buffer{}
	return &WebhookHandler{
		db:              db,
		deliveryService: delivery.NewService(db),
		logger:          log.New(buffer, "", 0),
		redisXAdd: func(_ context.Context, _ *redis.XAddArgs) *redis.StringCmd {
			return redis.NewStringResult("1-0", nil)
		},
		sendToChannel: notifier.SendToChannel,
		runAsync: func(fn func()) {
			fn()
		},
		sleep: func(time.Duration) {},
	}, buffer
}

func performWebhookRequest(
	t *testing.T,
	handler *WebhookHandler,
	sourceName string,
	apiKey string,
	payload interface{},
) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/webhook/"+sourceName, strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", apiKey)

	router := gin.New()
	router.POST("/webhook/:source_name", handler.HandleWebhook)
	router.ServeHTTP(recorder, request)

	return recorder
}

func computeCurrentFingerprintContract(handler *WebhookHandler, alert models.Alert) string {
	return handler.generateFingerprint(alert, nil)
}

func countWebhookNotificationState(t *testing.T, db *gorm.DB) map[string]int64 {
	t.Helper()

	counts := map[string]int64{}
	for name, model := range map[string]interface{}{
		"alerts":      &models.Alert{},
		"dataSources": &models.DataSource{},
		"channels":    &models.Channel{},
		"routeRules":  &models.RouteRule{},
	} {
		var count int64
		require.NoError(t, db.Model(model).Count(&count).Error)
		counts[name] = count
	}

	return counts
}

func requireSingleDeliveryLedger(
	t *testing.T,
	db *gorm.DB,
	alertID string,
) (models.NotificationDelivery, []models.NotificationDeliveryAttempt) {
	t.Helper()

	var deliveryRecord models.NotificationDelivery
	require.NoError(t, db.Where("alert_id = ?", alertID).First(&deliveryRecord).Error)

	var attemptRecords []models.NotificationDeliveryAttempt
	require.NoError(t, db.Where("delivery_id = ?", deliveryRecord.ID).Order("attempt_number asc").Find(&attemptRecords).Error)

	return deliveryRecord, attemptRecords
}

func decodeRenderedPayloadSnapshot(t *testing.T, raw datatypes.JSON) models.RenderedPayloadSnapshot {
	t.Helper()

	var snapshot models.RenderedPayloadSnapshot
	require.NoError(t, json.Unmarshal(raw, &snapshot))
	return snapshot
}

func decodeFinalFailureSummary(t *testing.T, raw datatypes.JSON) models.FinalFailureSummary {
	t.Helper()

	var summary models.FinalFailureSummary
	require.NoError(t, json.Unmarshal(raw, &summary))
	return summary
}

func assertDeliveryHasNoFailureSummary(t *testing.T, raw datatypes.JSON) {
	t.Helper()
	assert.True(t, len(raw) == 0 || string(raw) == "null")
}

func findWebhookLogLine(logOutput string, needle string) string {
	for _, line := range strings.Split(logOutput, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}

func parseWebhookLogFields(logLine string) map[string]string {
	fields := map[string]string{}
	index := 0

	for index < len(logLine) {
		for index < len(logLine) && logLine[index] == ' ' {
			index++
		}
		if index >= len(logLine) {
			break
		}

		keyStart := index
		for index < len(logLine) && isWebhookLogFieldKeyChar(logLine[index]) {
			index++
		}
		if keyStart == index || index >= len(logLine) || logLine[index] != '=' {
			break
		}

		key := logLine[keyStart:index]
		index++
		if index >= len(logLine) {
			break
		}

		value, nextIndex, ok := parseWebhookLogValue(logLine, index)
		if !ok {
			break
		}

		fields[key] = value
		index = nextIndex
	}
	return fields
}

func isWebhookLogFieldKeyChar(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func parseWebhookLogValue(logLine string, start int) (string, int, bool) {
	if logLine[start] == '"' {
		end := start + 1
		escaped := false
		for end < len(logLine) {
			ch := logLine[end]
			if escaped {
				escaped = false
				end++
				continue
			}
			if ch == '\\' {
				escaped = true
				end++
				continue
			}
			if ch == '"' {
				end++
				value, err := strconv.Unquote(logLine[start:end])
				if err != nil {
					return "", start, false
				}
				return value, end, true
			}
			end++
		}
		return "", start, false
	}

	end := start
	for end < len(logLine) && logLine[end] != ' ' {
		end++
	}
	if end == start {
		return "", start, false
	}
	return logLine[start:end], end, true
}

func assertWebhookLogFields(t *testing.T, logLine string, expected map[string]string) {
	t.Helper()

	fields := parseWebhookLogFields(logLine)
	for key, value := range expected {
		assert.Equal(t, value, fields[key], "field %s", key)
	}
}
