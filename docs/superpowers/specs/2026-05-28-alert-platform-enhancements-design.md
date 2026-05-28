---
title: "Alert Platform Enhancements — Resolved Handling, Throttling, UX"
date: 2026-05-28
status: draft
scope: backend + frontend
---

# Alert Platform Enhancements

## Overview

Six incremental improvements to the ShadowSong alert platform, organized into three delivery stages (S1 → S3). Each feature is independently deployable.

| # | Feature | Stage | Scope |
|---|---------|-------|-------|
| 1 | Alert Resolved Handling | S1 | backend + frontend |
| 2 | Notification Storm Throttling | S1 | backend |
| 3 | Alert List Quick Filters | S2 | frontend |
| 4 | Batch Ack / Silence | S2 | backend + frontend |
| 5 | Data Retention Policy | S2 | backend |
| 6 | Notification Template Preview | S3 | backend + frontend |

---

## S1 — Core Pipeline Completeness

### 1. Alert Resolved Handling

**Problem:** When Alertmanager sends a webhook with `status: "resolved"`, the system does not transition the corresponding `firing` alert to `resolved` or send recovery notifications. The functional spec (§340) explicitly requires this.

**Current behavior:**
- `webhook.go` normalizes incoming status via `mapStatus()`, which maps `"resolved"` to `"resolved"`
- `renderAlert()` produces an alert with `status = "resolved"`
- The dedup logic finds a matching alert by fingerprint — but only deduplicates `firing` alerts
- A `resolved` alert goes through the normal pipeline and gets treated as a new alert entry
- `CleanAlerts()` forcibly marks old `firing` alerts as `resolved` by timeout, not by actual recovery signal

**Design:**

After `renderAlert()` produces an alert with `status == "resolved"`:

1. **Find matching active alert** — query `alerts` table for `fingerprint = ? AND status IN ('firing', 'acked', 'silenced')` ordered by `trigger_time DESC LIMIT 1`
2. **If found:**
   - Update the existing alert: `status = "resolved"`, `updated_at = now`
   - Send recovery notification through the routing pipeline (reuse `matcher.FindMatchedTargets` + `sendChannelNotification`)
   - The notification template receives `status = "resolved"` so output templates can render recovery-specific messaging
   - Do NOT create a new alert row
3. **If not found:**
   - Log and discard silently (the alert was already resolved or never existed)
   - Do NOT create a new alert row

**Files changed:**
- `internal/handlers/webhook.go` — add `handleResolvedAlert()` function, call it when `alert.Status == "resolved"` after `renderAlert()`
- `internal/handlers/webhook_test.go` — test cases: resolved with matching alert, resolved with no match, resolved updates acked alert

**Frontend:**
- `frontend/src/components/SeverityBadge.tsx` — add `resolved` variant (green badge)
- No other changes needed; the alert list already shows `status` field

### 2. Notification Storm Throttling

**Problem:** When dozens of alerts fire simultaneously (e.g., a game server rack goes down), every alert triggers a notification to every matched channel. Feishu/DingTalk will throttle or ban the webhook URL. The functional spec (§161) mentions alert storm protection.

**Current behavior:**
- `sendChannelNotification()` in `webhook.go` sends immediately for every alert
- Webhook ingress has per-source rate limiting (`middleware/rate_limit.go`), but this limits incoming webhooks, not outgoing notifications
- No protection at the channel send level

**Design:**

Add a per-channel token bucket rate limiter in the notification pipeline.

**Channel model change (`internal/models/models.go`):**
```go
type Channel struct {
    // ... existing fields ...
    RateLimit int `gorm:"default:0" json:"rate_limit"` // max sends per minute, 0 = unlimited
}
```

**Rate limiter (`internal/notifier/throttle.go`):**
- Simple in-memory sliding window per channel ID
- Structure: `map[uint]*channelBucket` with mutex
- Each bucket tracks timestamps of sends within the last 60 seconds
- `Allow(channelID uint, limit int) bool` — returns false if limit exceeded
- On startup, the limiter is empty; no persistence needed

**Integration point — `webhook.go` `sendChannelNotification()`:**
- Before calling `notifier.SendToChannel()`, check `throttle.Allow(channel.ID, channel.RateLimit)`
- If throttled:
  - Create a `NotificationDelivery` record with `delivery_status = "throttled"` (new status constant)
  - Log the throttling event
  - Do NOT call `notifier.SendToChannel()`
- The `NotificationDelivery` record allows operators to see throttled sends in the delivery list

**New delivery status constant (`internal/models/notification_delivery.go`):**
```go
DeliveryStatusThrottled = "throttled"
```
Add to `validDeliveryStatuses` map.

**Files changed:**
- `internal/models/models.go` — add `RateLimit` field to `Channel`
- `internal/notifier/throttle.go` — new file, in-memory per-channel rate limiter
- `internal/notifier/throttle_test.go` — new file, unit tests
- `internal/models/notification_delivery.go` — add `DeliveryStatusThrottled`
- `internal/handlers/webhook.go` — integrate throttle check before sending
- `internal/escalation/checker.go` — integrate same throttle check

**Frontend:**
- `frontend/src/pages/Channels.tsx` — add `rate_limit` field to channel edit form (number input, 0 = unlimited)
- `frontend/src/types/index.ts` — add `rate_limit` to `Channel` interface
- `frontend/src/pages/Deliveries.tsx` — add "throttled" to status filter options

---

## S2 — Daily Operations UX

### 3. Alert List Quick Filters

**Problem:** The Alerts page shows all alerts in a flat list. Filtering by status requires typing in the status dropdown. Common daily operations (check what's currently firing, review recent resolutions) require manual filtering each time.

**Design:**

Add a row of status tabs above the existing alert table.

**Frontend changes (`frontend/src/pages/Alerts.tsx`):**
- Add a `TabView` (PrimeReact) above the DataTable with tabs:
  - "活跃" (firing) — default
  - "已确认" (acked)
  - "已静默" (silenced)
  - "已恢复" (resolved)
  - "全部"
- Clicking a tab calls `alertStore.setFilters({ status: selectedStatus })` (or clears status for "全部")
- Tab labels show count badges (reuse `fetchStats()` data)
- The existing advanced filters (severity, source, time range, label selector) remain below the tabs

**No backend changes needed** — `GET /api/v1/alerts?status=firing` already works.

### 4. Batch Ack / Silence

**Problem:** During incidents, operators need to acknowledge or silence multiple alerts at once. Currently each alert requires individual action.

**Backend — new endpoints:**

`POST /api/v1/alerts/batch-ack`
```json
{
  "alert_ids": ["alert-1", "alert-2", "alert-3"],
  "comment": "batch ack during incident"
}
```
- Validate: each `alert_id` exists and `CanBeAcked()`
- Update all matching alerts in a single DB transaction
- Record audit log per alert (reuse existing `recordAudit`)
- Return `{ "updated": 3, "skipped": 1, "errors": [...] }`

`POST /api/v1/alerts/batch-silence`
```json
{
  "alert_ids": ["alert-1", "alert-2"],
  "duration": 3600
}
```
- Validate: each `alert_id` exists
- Update status to `silenced` for all matching alerts
- Create a single `SilenceRule` covering all silenced alert names (reuse `QuickSilence` pattern)
- Return same shape as batch-ack

**Router (`internal/router/router.go`):**
```go
alerts.POST("/batch-ack", middleware.RequireCapability(authz.CapabilityProcessAlerts), alertHandler.BatchAck)
alerts.POST("/batch-silence", middleware.RequireCapability(authz.CapabilityProcessAlerts), alertHandler.BatchSilence)
```

**Frontend (`frontend/src/pages/Alerts.tsx`):**
- Add `rowSelection` to DataTable (checkbox column)
- When selection is non-empty, show a floating action bar with:
  - "批量确认" button → opens comment modal → calls `batchAck(ids, comment)`
  - "批量静默" button → calls `batchSilence(ids, 3600)`
- Add `batchAck` and `batchSilence` to `alertStore.ts` and `alertApi` in `client.ts`

**Files changed:**
- `internal/handlers/alert.go` — add `BatchAck()` and `BatchSilence()` methods
- `internal/handlers/alert_test.go` — test batch endpoints
- `internal/router/router.go` — register new routes
- `frontend/src/pages/Alerts.tsx` — multi-select UI + action bar
- `frontend/src/stores/alertStore.ts` — batch actions
- `frontend/src/api/client.ts` — batch API calls
- `frontend/src/types/index.ts` — batch request/response types

### 5. Data Retention Policy

**Problem:** Alerts, deliveries, and delivery attempts accumulate indefinitely. Over months, PostgreSQL will grow unbounded.

**Design:**

Add a background cleanup goroutine started from `cmd/server/main.go`.

**Configuration (`internal/config/config.go`):**
```go
// In ServerConfig:
RetentionDays int // from ALERT_RETENTION_DAYS env, default 30
```

**Cleanup logic (`internal/retention/cleanup.go` — new package):**
- `Run(db *gorm.DB, retentionDays int, interval time.Duration, stop <-chan struct{})`
- Every `interval` (default 1 hour), execute:
  1. `DELETE FROM alerts WHERE created_at < NOW() - INTERVAL '$retentionDays days'`
  2. `DELETE FROM notification_deliveries WHERE created_at < NOW() - INTERVAL '$retentionDays days'`
  3. `DELETE FROM notification_delivery_attempts WHERE created_at < NOW() - INTERVAL '$retentionDays days'`
  4. `DELETE FROM notification_delivery_recoveries WHERE created_at < NOW() - INTERVAL '$retentionDays days'`
- Use raw SQL with `db.Exec()` for performance (avoid loading ORM objects)
- Log deleted row counts: `"retention: cleaned alerts=%d deliveries=%d attempts=%d"`
- Skip if `retentionDays <= 0` (disabled)

**Startup wiring (`cmd/server/main.go`):**
```go
if cfg.Server.RetentionDays > 0 {
    stop := make(chan struct{})
    go retention.Run(db, cfg.Server.RetentionDays, 1*time.Hour, stop)
}
```

**Files changed:**
- `internal/config/config.go` — add `RetentionDays` to `ServerConfig`
- `internal/retention/cleanup.go` — new file
- `internal/retention/cleanup_test.go` — new file
- `cmd/server/main.go` — wire retention goroutine

**No frontend changes needed.**

---

## S3 — Later Iteration

### 6. Notification Template Preview

**Problem:** Operators configure channel templates but can only verify them by triggering a real alert and checking the notification. There's no way to preview what the final message looks like before saving.

**Backend — new endpoint:**

`POST /api/v1/channels/:id/preview`
```json
{
  "alert_id": "some-alert-id"  // optional; if omitted, use a synthetic sample alert
}
```

Response:
```json
{
  "title": "[P1] HighMemory on game-server-01",
  "content": "Memory usage exceeded 90%\n\nLabels: host=game-server-01, zone=east-1",
  "channel_type": "feishu"
}
```

Logic:
1. Load channel by ID
2. Load alert (by `alert_id` or generate synthetic sample)
3. Load the route rule that would match this alert (or use a default)
4. Call `template.Renderer.RenderAlert()` with the channel's output template
5. Return rendered title + content without sending

**Frontend (`frontend/src/pages/Channels.tsx`):**
- Add "预览" button next to "测试发送"
- Clicking opens a Dialog showing the rendered title and content
- Uses a recent alert from the system (or a sample payload) as input

**Files changed:**
- `internal/handlers/config.go` — add `PreviewChannel()` method
- `internal/router/router.go` — register route
- `frontend/src/pages/Channels.tsx` — preview button + dialog
- `frontend/src/api/client.ts` — preview API call

---

## Non-Goals (Explicitly Excluded)

- HMAC webhook signature verification
- Multi-instance WebSocket (Redis Pub/Sub broadcast)
- These are not needed for the current deployment model

## Dependencies Between Features

```
S1.1 (Resolved) ──no deps──> can ship independently
S1.2 (Throttle) ──no deps──> can ship independently
S2.3 (Filters)  ──no deps──> frontend only
S2.4 (Batch)    ──no deps──> can ship independently
S2.5 (Retention)──no deps──> can ship independently
S3.6 (Preview)  ──no deps──> can ship independently
```

All six features are independent and can be developed and shipped in any order within their stage.
