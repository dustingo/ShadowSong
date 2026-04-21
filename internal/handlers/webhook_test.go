package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
		ChannelIDs: datatypes.JSON(`[1]`),
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
	assert.Contains(t, logOutput, "trace_id="+alert.TraceID)
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
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com"}`),
		Enabled: true,
	}).Error)
	require.NoError(t, db.Create(&models.RouteRule{
		Name:       "all",
		Priority:   1,
		ChannelIDs: datatypes.JSON(`[1]`),
		Enabled:    true,
	}).Error)

	handler.processAlertNotificationsAsync([]models.Alert{{
		AlertID:   "alert-panic",
		Source:    "prometheus",
		AlertName: "CPUHigh",
		Severity:  "P1",
		Message:   "cpu high",
		Status:    "firing",
	}})

	assert.Contains(t, logBuffer.String(), "stage=async_panic")
	assert.Contains(t, logBuffer.String(), "recovered panic=boom")
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

func newWebhookTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}, &models.DataSource{}, &models.Channel{}, &models.RouteRule{}))

	return db
}

func newWebhookTestHandler(db *gorm.DB) (*WebhookHandler, *bytes.Buffer) {
	buffer := &bytes.Buffer{}
	return &WebhookHandler{
		db:     db,
		logger: log.New(buffer, "", 0),
		redisXAdd: func(_ context.Context, _ *redis.XAddArgs) *redis.StringCmd {
			return redis.NewStringResult("1-0", nil)
		},
		sendToChannel: notifier.SendToChannel,
		runAsync: func(fn func()) {
			fn()
		},
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
