package notifier

import (
	"errors"
	"testing"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestSendToChannel_UnsupportedTypeIncludesChannelContext(t *testing.T) {
	channel := &models.Channel{
		ID:     7,
		Name:   "broken-channel",
		Type:   "unknown",
		Config: datatypes.JSON(`{}`),
	}

	err := SendToChannel(channel, "title", "content")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "channel 7 (broken-channel)")
		assert.Contains(t, err.Error(), "unsupported type")
	}
}

func TestIsRetryableSendError_TransientSendFailuresAreRetryable(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "transport failure",
			err:  errors.New("channel 11 (ops-webhook) send failed: failed to send webhook notification: dial tcp timeout"),
		},
		{
			name: "upstream service unavailable",
			err:  errors.New("channel 11 (ops-webhook) send failed: webhook notification failed with status: 503"),
		},
		{
			name: "feishu upstream rate limit",
			err:  errors.New("channel 11 (ops-webhook) send failed: feishu notification failed, status: 429, body: busy"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, IsRetryableSendError(tc.err))
		})
	}
}

func TestIsRetryableSendError_DeterministicFailuresRemainTerminal(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "unsupported channel type",
			err:  errors.New("channel 7 (broken-channel) unsupported type: unknown"),
		},
		{
			name: "sender init failure",
			err:  errors.New("channel 9 (ops-feishu) sender init failed: feishu webhook_url is required"),
		},
		{
			name: "template render failure",
			err:  errors.New("template execute error: map has no entry for key \"labels\""),
		},
		{
			name: "datasource lookup failure",
			err:  errors.New("data source not found for source=prometheus"),
		},
		{
			name: "default notification fallback init failure",
			err:  errors.New("channel 10 (ops-webhook) send failed: failed to create webhook request: parse \"::\": missing protocol scheme"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, IsRetryableSendError(tc.err))
		})
	}
}
