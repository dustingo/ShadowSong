package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
