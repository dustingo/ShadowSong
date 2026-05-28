# S1: Alert Resolved Handling + Notification Throttling

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the alert lifecycle by handling recovery events from Alertmanager, and protect notification channels from storm traffic with per-channel rate limiting.

**Architecture:** Resolved handling inserts a branch in `webhook.go` after fingerprint generation — when `status == "resolved"`, find the matching active alert, update it, and send a recovery notification through the existing routing pipeline. Throttling adds an in-memory sliding-window rate limiter at the `sendChannelNotification` entry point, gated by a `rate_limit` field on the `Channel` model.

**Tech Stack:** Go, GORM, Gin, PrimeReact, TypeScript

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `internal/handlers/webhook.go` | Add resolved alert branch + throttle integration |
| Modify | `internal/handlers/webhook_test.go` | Tests for resolved handling + throttle |
| Create | `internal/notifier/throttle.go` | Per-channel sliding window rate limiter |
| Create | `internal/notifier/throttle_test.go` | Throttle unit tests |
| Modify | `internal/models/models.go` | Add `RateLimit` field to `Channel` |
| Modify | `internal/models/notification_delivery.go` | Add `DeliveryStatusThrottled` constant |
| Modify | `internal/escalation/checker.go` | Integrate throttle check |
| Modify | `frontend/src/components/SeverityBadge.tsx` | Add `resolved` status badge |
| Modify | `frontend/src/types/index.ts` | Add `rate_limit` to `Channel` |
| Modify | `frontend/src/pages/Channels.tsx` | Add rate_limit input to channel form |
| Modify | `frontend/src/pages/Deliveries.tsx` | Add "throttled" status filter |

---

## Task 1: Add `DeliveryStatusThrottled` constant

**Files:**
- Modify: `internal/models/notification_delivery.go:20-26`

- [ ] **Step 1: Add the constant and register it in the valid map**

In `internal/models/notification_delivery.go`, add after line 20 (`DeliveryStatusFailed = "failed"`):

```go
DeliveryStatusThrottled = "throttled"
```

Add to `validDeliveryStatuses` map:

```go
var validDeliveryStatuses = map[string]struct{}{
	DeliveryStatusPending:    {},
	DeliveryStatusDelivered:  {},
	DeliveryStatusFailed:     {},
	DeliveryStatusThrottled:  {},
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd D:\goproject\shadowsongAI && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/models/notification_delivery.go
git commit -m "feat(models): add DeliveryStatusThrottled constant"
```

---

## Task 2: Add `RateLimit` field to Channel model

**Files:**
- Modify: `internal/models/models.go` — Channel struct

- [ ] **Step 1: Add field to Channel struct**

In `internal/models/models.go`, in the `Channel` struct, add after the `Enabled` field:

```go
RateLimit int  `gorm:"default:0" json:"rate_limit"` // max notifications per minute, 0 = unlimited
```

- [ ] **Step 2: Verify compilation**

Run: `cd D:\goproject\shadowsongAI && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/models/models.go
git commit -m "feat(models): add rate_limit field to Channel"
```

---

## Task 3: Create per-channel throttle limiter

**Files:**
- Create: `internal/notifier/throttle.go`
- Create: `internal/notifier/throttle_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/notifier/throttle_test.go`:

```go
package notifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThrottleAllowUnderLimit(t *testing.T) {
	th := NewChannelThrottle()
	// limit 2 per minute
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
	// channel 1 at limit, channel 2 at limit
	assert.False(t, th.Allow(1, 1))
	assert.False(t, th.Allow(2, 1))
	// channel 3 is independent
	assert.True(t, th.Allow(3, 1))
}

func TestThrottleWindowExpiry(t *testing.T) {
	th := NewChannelThrottle()
	th.Allow(1, 1)
	assert.False(t, th.Allow(1, 1))

	// Manually expire the entry by setting timestamp in the past
	th.mu.Lock()
	th.buckets[1].timestamps[0] = time.Now().Add(-61 * time.Second)
	th.mu.Unlock()

	assert.True(t, th.Allow(1, 1))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/notifier/ -run TestThrottle -v`
Expected: FAIL — `NewChannelThrottle` and `Allow` undefined

- [ ] **Step 3: Implement the throttle**

Create `internal/notifier/throttle.go`:

```go
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
// It is safe for concurrent use.
type ChannelThrottle struct {
	mu      sync.Mutex
	buckets map[uint]*channelBucket
}

// NewChannelThrottle creates a new ChannelThrottle instance.
func NewChannelThrottle() *ChannelThrottle {
	return &ChannelThrottle{
		buckets: make(map[uint]*channelBucket),
	}
}

// Allow reports whether a send to the given channel is allowed under the limit.
// A limit of 0 means unlimited.
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

	// Prune expired timestamps
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/notifier/ -run TestThrottle -v`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/notifier/throttle.go internal/notifier/throttle_test.go
git commit -m "feat(notifier): add per-channel sliding-window throttle"
```

---

## Task 4: Integrate throttle into sendChannelNotification

**Files:**
- Modify: `internal/handlers/webhook.go` — `sendChannelNotification` method

- [ ] **Step 1: Add throttle field to WebhookHandler struct**

In `internal/handlers/webhook.go`, add to the `WebhookHandler` struct:

```go
throttle *notifier.ChannelThrottle
```

In `NewWebhookHandler`, initialize it:

```go
throttle: notifier.NewChannelThrottle(),
```

- [ ] **Step 2: Write the failing test**

In `internal/handlers/webhook_test.go`, add:

```go
func TestSendChannelNotificationThrottled(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	// Create channel with rate_limit = 1
	channel := models.Channel{
		Name:      "test-channel",
		Type:      "webhook",
		Config:    datatypes.JSON(`{"url":"http://localhost:1","method":"POST"}`),
		Enabled:   true,
		RateLimit: 1,
	}
	require.NoError(t, db.Create(&channel).Error)

	alert := models.Alert{
		AlertID:     "throttle-test-alert",
		Source:      "test",
		AlertName:   "TestAlert",
		Severity:    "P1",
		Message:     "test",
		Fingerprint: "fp-throttle",
		Status:      "firing",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
	}
	require.NoError(t, db.Create(&alert).Error)

	// First send should succeed (or at least not be throttled)
	handler.sendChannelNotification(&alert, &channel, nil, "title", "content", models.DeliveryModeDefault)

	// Second send should be throttled
	handler.sendChannelNotification(&alert, &channel, nil, "title", "content", models.DeliveryModeDefault)

	// Check that one delivery record is throttled
	var deliveries []models.NotificationDelivery
	db.Where("alert_id = ?", "throttle-test-alert").Find(&deliveries)
	require.Len(t, deliveries, 2)

	statuses := make(map[string]int)
	for _, d := range deliveries {
		statuses[d.DeliveryStatus]++
	}
	assert.Equal(t, 1, statuses[models.DeliveryStatusThrottled])
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -run TestSendChannelNotificationThrottled -v`
Expected: FAIL — throttle not yet integrated

- [ ] **Step 4: Implement throttle check in sendChannelNotification**

In `internal/handlers/webhook.go`, at the top of `sendChannelNotification`, add after the function signature:

```go
// Throttle check
if !h.throttle.Allow(channel.ID, channel.RateLimit) {
	h.logAlertEvent("throttled", h.traceFieldsForNotification(alert, channel), "notification throttled by rate limit")
	// Record a throttled delivery
	h.startNotificationDeliveryThrottled(alert, channel, routeRule, title, content, mode)
	return
}
```

Add the helper method:

```go
func (h *WebhookHandler) startNotificationDeliveryThrottled(
	alert *models.Alert,
	channel *models.Channel,
	routeRule *models.RouteRule,
	title, content, mode string,
) {
	deliverySvc := h.deliveryService
	if deliverySvc == nil {
		return
	}
	_, err := deliverySvc.StartDelivery(context.Background(), delivery.StartDeliveryInput{
		Alert:         alert,
		Channel:       channel,
		RouteRule:     routeRule,
		DeliveryMode:  mode,
		TriggerKind:   models.TriggerKindPipeline,
		RenderedTitle: title,
		RenderedBody:  content,
	})
	if err != nil {
		h.logger.Printf("throttled delivery record failed: %v", err)
		return
	}
	// Mark as throttled immediately
	// The StartDelivery creates a pending record; we need to update its status
	// For simplicity, we'll create with throttled status by updating right after
	var latest models.NotificationDelivery
	if err := h.db.Where("alert_id = ? AND channel_id = ?", alert.AlertID, channel.ID).
		Order("id DESC").First(&latest).Error; err == nil {
		latest.DeliveryStatus = models.DeliveryStatusThrottled
		h.db.Save(&latest)
	}
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -run TestSendChannelNotificationThrottled -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/handlers/webhook.go internal/handlers/webhook_test.go
git commit -m "feat(webhook): integrate per-channel throttle into notification pipeline"
```

---

## Task 5: Add resolved alert handling

**Files:**
- Modify: `internal/handlers/webhook.go` — add `handleResolvedAlert` and wire it
- Modify: `internal/handlers/webhook_test.go` — test resolved flow

- [ ] **Step 1: Write the failing tests**

In `internal/handlers/webhook_test.go`, add:

```go
func TestHandleResolvedAlertUpdatesFiringAlert(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	require.NoError(t, db.Create(&models.DataSource{
		Name:              "resolved-source",
		DisplayName:       "Resolved Source",
		APIKey:            "key",
		DeduplicateWindow: 3600,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "resolved-source",
			"status": "{{.event_type}}"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		GroupByLabels:  datatypes.JSON(`["alert_name","severity","source"]`),
		Enabled:        true,
	}).Error)

	// Create an existing firing alert
	require.NoError(t, db.Create(&models.Alert{
		AlertID:     "existing-firing",
		Source:      "resolved-source",
		AlertName:   "CPUHigh",
		Severity:    "P1",
		Message:     "cpu high",
		Fingerprint: "resolved-source|CPUHigh|P1",
		Status:      "firing",
		TriggerTime: time.Now().Add(-5 * time.Minute),
		ReceivedAt:  time.Now().Add(-5 * time.Minute),
	}).Error)

	// Send a resolved webhook
	payload := []map[string]interface{}{
		{
			"external_id": "existing-firing",
			"alert_name":  "CPUHigh",
			"severity":    "P1",
			"message":     "cpu high",
			"event_type":  "resolved",
		},
	}

	recorder := performWebhookRequest(t, handler, "resolved-source", "key", payload)
	require.Equal(t, http.StatusOK, recorder.Code)

	// Verify the existing alert is now resolved
	var alert models.Alert
	require.NoError(t, db.First(&alert, "alert_id = ?", "existing-firing").Error)
	assert.Equal(t, "resolved", alert.Status)
}

func TestHandleResolvedAlertNoMatchCreatesNothing(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	require.NoError(t, db.Create(&models.DataSource{
		Name:              "resolved-no-match",
		DisplayName:       "Resolved No Match",
		APIKey:            "key",
		DeduplicateWindow: 3600,
		InputTemplate: `{
			"alert_id": "{{.external_id}}",
			"alert_name": "{{.alert_name}}",
			"severity": "{{.severity}}",
			"message": "{{.message}}",
			"source": "resolved-no-match",
			"status": "{{.event_type}}"
		}`,
		OutputTemplate: `{"title":"{{.alert_name}}","content":"{{.message}}"}`,
		GroupByLabels:  datatypes.JSON(`["alert_name","severity","source"]`),
		Enabled:        true,
	}).Error)

	payload := []map[string]interface{}{
		{
			"external_id": "never-existed",
			"alert_name":  "DiskFull",
			"severity":    "P0",
			"message":     "disk full",
			"event_type":  "resolved",
		},
	}

	recorder := performWebhookRequest(t, handler, "resolved-no-match", "key", payload)
	require.Equal(t, http.StatusOK, recorder.Code)

	// No new alert should be created
	var count int64
	db.Model(&models.Alert{}).Where("alert_id = ?", "never-existed").Count(&count)
	assert.Equal(t, int64(0), count)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -run TestHandleResolved -v`
Expected: FAIL — resolved handling not yet implemented

- [ ] **Step 3: Implement handleResolvedAlert**

In `internal/handlers/webhook.go`, add the new method:

```go
// handleResolvedAlert processes a resolved alert by finding the matching
// active alert (firing/acked/silenced) and transitioning it to resolved.
// Returns true if the alert was handled (matched or discarded), false to
// continue with normal pipeline processing.
func (h *WebhookHandler) handleResolvedAlert(alert models.Alert, traceID string) bool {
	// Find matching active alert by fingerprint
	var existing models.Alert
	err := h.db.Where("fingerprint = ? AND status IN ?", alert.Fingerprint, []string{"firing", "acked", "silenced"}).
		Order("trigger_time DESC").
		First(&existing).Error

	if err != nil {
		// No matching active alert — discard silently
		h.logTraceStage("resolved_discarded", map[string]string{
			"trace_id":    traceID,
			"fingerprint": alert.Fingerprint,
		}, "no active alert found for resolved event")
		return true
	}

	// Transition to resolved
	existing.Status = "resolved"
	if err := h.db.Save(&existing).Error; err != nil {
		h.logger.Printf("failed to mark alert resolved: alert_id=%s error=%v", existing.AlertID, err)
		return true
	}

	h.logTraceStage("resolved", map[string]string{
		"trace_id":      traceID,
		"alert_id":      existing.AlertID,
		"fingerprint":   existing.Fingerprint,
		"prev_status":   existing.Status,
	}, "alert marked as resolved")

	// Send recovery notification through routing pipeline
	h.asyncRunner()(func() {
		h.processAlertNotificationsAsync([]models.Alert{existing})
	})

	return true
}
```

- [ ] **Step 4: Wire resolved branch into HandleWebhook**

In `internal/handlers/webhook.go`, inside the `for _, alertData := range alerts` loop, after step 6 (fingerprint generation, line ~173) and before step 7 (dedup logic), add:

```go
// 6.5 Handle resolved events
if alert.Status == "resolved" {
	if h.handleResolvedAlert(alert, traceID) {
		results = append(results, alert)
		continue
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -run TestHandleResolved -v`
Expected: PASS (both tests)

- [ ] **Step 6: Run all existing tests to check for regressions**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -v`
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/handlers/webhook.go internal/handlers/webhook_test.go
git commit -m "feat(webhook): handle resolved alerts by transitioning matching active alerts"
```

---

## Task 6: Add resolved status badge to frontend

**Files:**
- Modify: `frontend/src/components/SeverityBadge.tsx`

- [ ] **Step 1: Add a StatusBadge component**

The existing `SeverityBadge.tsx` handles severity (P0-P3). Add a new exported component for alert status in the same file:

```tsx
interface StatusBadgeProps {
  status: string
}

const statusConfig: Record<string, { label: string; severity: string }> = {
  firing: { label: '告警中', severity: 'danger' },
  acked: { label: '已确认', severity: 'warning' },
  silenced: { label: '已静默', severity: 'secondary' },
  resolved: { label: '已恢复', severity: 'success' },
  pending: { label: '待处理', severity: 'info' },
  deduplicated: { label: '已去重', severity: 'secondary' },
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const config = statusConfig[status]
  if (!config) {
    return <Tag value={status} severity="secondary" />
  }
  return <Tag value={config.label} severity={config.severity} />
}
```

- [ ] **Step 2: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/SeverityBadge.tsx
git commit -m "feat(ui): add StatusBadge component with resolved state"
```

---

## Task 7: Frontend — add rate_limit to Channel types and form

**Files:**
- Modify: `frontend/src/types/index.ts` — add `rate_limit` to Channel
- Modify: `frontend/src/pages/Channels.tsx` — add rate_limit input
- Modify: `frontend/src/pages/Deliveries.tsx` — add throttled filter option

- [ ] **Step 1: Add rate_limit to Channel interface**

In `frontend/src/types/index.ts`, in the `Channel` interface, add:

```ts
rate_limit?: number
```

- [ ] **Step 2: Add rate_limit input to Channels form**

In `frontend/src/pages/Channels.tsx`, in the channel edit/create dialog form, add after the `enabled` toggle:

```tsx
<div className="field">
  <label htmlFor="rate_limit">发送频率限制（每分钟，0=不限）</label>
  <InputNumber
    id="rate_limit"
    value={formValues.config?.rate_limit ?? 0}
    min={0}
    onChange={(e) =>
      setFormValues({
        ...formValues,
        config: { ...formValues.config, rate_limit: e.value ?? 0 },
      })
    }
  />
</div>
```

- [ ] **Step 3: Add "throttled" to Deliveries filter**

In `frontend/src/pages/Deliveries.tsx`, find the delivery status dropdown options and add:

```ts
{ label: '已限流', value: 'throttled' },
```

- [ ] **Step 4: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/pages/Channels.tsx frontend/src/pages/Deliveries.tsx
git commit -m "feat(ui): add rate_limit to channel form and throttled status to deliveries"
```

---

## Task 8: Integrate throttle into escalation checker

**Files:**
- Modify: `internal/escalation/checker.go`

- [ ] **Step 1: Add throttle field to Checker**

In `internal/escalation/checker.go`, add to the `Checker` struct:

```go
throttle *notifier.ChannelThrottle
```

In `NewChecker`, initialize:

```go
throttle: notifier.NewChannelThrottle(),
```

- [ ] **Step 2: Add throttle check before sending**

In `sendEscalationNotification`, before calling `c.sendToChannel`, add:

```go
if !c.throttle.Allow(channel.ID, channel.RateLimit) {
	log.Printf("escalation: throttled channel %d (%s)", channel.ID, channel.Name)
	return
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd D:\goproject\shadowsongAI && go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/escalation/checker.go
git commit -m "feat(escalation): integrate throttle check"
```

---

## Task 9: Final integration test

- [ ] **Step 1: Run all backend tests**

Run: `cd D:\goproject\shadowsongAI && go test ./... -v 2>&1 | Select-Object -Last 30`
Expected: all PASS

- [ ] **Step 2: Run frontend type check**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Final commit if any fixups needed**

```bash
git add -A
git commit -m "chore: s1 final integration fixes"
```
