package notifier

import (
	"sync"
	"time"
)

const throttleWindow = 60 * time.Second

type channelBucket struct {
	timestamps []time.Time
}

// ChannelThrottle provides per-channel sliding-window rate limiting.
type ChannelThrottle struct {
	mu      sync.Mutex
	buckets map[uint]*channelBucket
}

// NewChannelThrottle creates a new ChannelThrottle.
func NewChannelThrottle() *ChannelThrottle {
	return &ChannelThrottle{
		buckets: make(map[uint]*channelBucket),
	}
}

// Allow reports whether a notification to the given channel is permitted.
// A limit <= 0 means unlimited.
func (ct *ChannelThrottle) Allow(channelID uint, limit int) bool {
	if limit <= 0 {
		return true
	}
	ct.mu.Lock()
	defer ct.mu.Unlock()
	bucket, exists := ct.buckets[channelID]
	if !exists {
		bucket = &channelBucket{}
		ct.buckets[channelID] = bucket
	}
	now := time.Now()
	cutoff := now.Add(-throttleWindow)
	valid := bucket.timestamps[:0]
	for _, ts := range bucket.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	bucket.timestamps = valid
	if len(bucket.timestamps) >= limit {
		return false
	}
	bucket.timestamps = append(bucket.timestamps, now)
	return true
}
