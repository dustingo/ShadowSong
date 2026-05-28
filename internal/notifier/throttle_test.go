package notifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThrottleAllowUnderLimit(t *testing.T) {
	th := NewChannelThrottle()
	assert.True(t, th.Allow(1, 2))
	assert.True(t, th.Allow(1, 2))
}

func TestThrottleBlockOverLimit(t *testing.T) {
	th := NewChannelThrottle()
	th.Allow(1, 2)
	th.Allow(1, 2)
	assert.False(t, th.Allow(1, 2))
}

func TestThrottleZeroMeansUnlimited(t *testing.T) {
	th := NewChannelThrottle()
	for i := 0; i < 100; i++ {
		assert.True(t, th.Allow(1, 0))
	}
}

func TestThrottlePerChannel(t *testing.T) {
	th := NewChannelThrottle()
	th.Allow(1, 1)
	th.Allow(2, 1)
	assert.False(t, th.Allow(1, 1))
	assert.False(t, th.Allow(2, 1))
	assert.True(t, th.Allow(3, 1))
}

func TestThrottleWindowExpiry(t *testing.T) {
	th := NewChannelThrottle()
	th.Allow(1, 1)
	assert.False(t, th.Allow(1, 1))
	th.mu.Lock()
	th.buckets[1].timestamps[0] = time.Now().Add(-61 * time.Second)
	th.mu.Unlock()
	assert.True(t, th.Allow(1, 1))
}
