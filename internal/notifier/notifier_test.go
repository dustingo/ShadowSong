package notifier

import (
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
