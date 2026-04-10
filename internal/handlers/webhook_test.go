package handlers

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

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
		})
	}
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
