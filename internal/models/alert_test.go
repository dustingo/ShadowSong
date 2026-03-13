package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestAlert_Validation(t *testing.T) {
	tests := []struct {
		name    string
		alert   Alert
		wantErr bool
	}{
		{
			name: "valid alert",
			alert: Alert{
				AlertID:     "test-alert-1",
				Source:      "prometheus",
				AlertName:   "HighMemory",
				Severity:    "P0",
				Message:     "Memory usage is high",
				Labels:      datatypes.JSON(`{"host": "server-01"}`),
				Fingerprint: "abc123",
				TriggerTime: time.Now(),
				ReceivedAt:  time.Now(),
				Status:      "pending",
				Raw:         datatypes.JSON(`{"original": "data"}`),
			},
			wantErr: false,
		},
		{
			name: "invalid severity",
			alert: Alert{
				AlertID:     "test-alert-2",
				Source:      "prometheus",
				AlertName:   "HighMemory",
				Severity:    "INVALID",
				Message:     "Memory usage is high",
				Fingerprint: "abc123",
				TriggerTime: time.Now(),
				ReceivedAt:  time.Now(),
				Status:      "pending",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.alert.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAlert_IsValidSeverity(t *testing.T) {
	tests := []struct {
		severity string
		want     bool
	}{
		{"P0", true},
		{"P1", true},
		{"P2", true},
		{"P3", true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			alert := Alert{Severity: tt.severity}
			assert.Equal(t, tt.want, alert.IsValidSeverity())
		})
	}
}

func TestAlert_IsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"pending", true},
		{"firing", true},
		{"acked", true},
		{"silenced", true},
		{"resolved", true},
		{"deduplicated", true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			alert := Alert{Status: tt.status}
			assert.Equal(t, tt.want, alert.IsValidStatus())
		})
	}
}
