---
name: alert-platform-v2-enhancements
description: Alert aggregation, escalation, delivery observability, and OnDuty removal
---

# Alert Platform v2 Enhancements Design

## Overview

Four changes to improve the alert platform from "functional" to "production-ready":

1. **Alert Aggregation** — Collapse repeated alerts by fingerprint, show "N occurrences" instead of N rows
2. **Escalation (Timeout Re-notification)** — Auto re-notify firing alerts that remain un-acked after N minutes
3. **Delivery Observability** — Embed delivery records inside alert detail view
4. **OnDuty Removal** — Delete the disconnected on-duty schedule feature

---

## 1. Alert Aggregation

### Problem

When the same alert fires repeatedly (e.g., a service flapping between firing→resolved→firing), the alert list shows multiple independent rows. During an alert storm, the list becomes unusable — operators cannot quickly identify which alerts are distinct issues vs. repeated occurrences of the same problem.

### Data Model

No schema changes. The existing `fingerprint` field on `alerts` already groups logically-identical alerts. The dedup logic merges within a single firing window, but across multiple firing windows (alert resolved then re-triggered), separate alert records are created with the same fingerprint.

### Backend Changes

#### New API endpoint

```
GET /api/v1/alerts/active?grouped=true
```

Response format when `grouped=true`:

```json
[
  {
    "fingerprint": "abc123...",
    "latest_alert": { /* full Alert object */ },
    "count": 5,
    "first_triggered_at": "2026-05-18T08:00:00Z",
    "last_triggered_at": "2026-05-18T12:30:00Z"
  }
]
```

When `grouped=false` (default): returns the existing flat array of Alert objects. No behavior change for existing callers.

#### Implementation

In `AlertHandler.Active()`:

1. If `grouped` query param is `"true"`:
   - Query all firing alerts as before.
   - Group by `fingerprint` in Go code.
   - For each group: pick the alert with the most recent `trigger_time` as `latest_alert`, compute `count`, `first_triggered_at` (min), `last_triggered_at` (max).
   - Return the grouped response.
2. If `grouped` is absent or `"false"`: return existing flat response.

#### Stats endpoint

`GET /api/v1/alerts/stats` should count **unique fingerprints** for the `firing` count, not total alert rows. This prevents a 5-occurrence alert from inflating the "active alerts" stat card to 5.

Update `getStatusCounts` in `internal/stats/queries.go`:

```sql
-- Current: counts all rows with status='firing'
-- New: counts distinct fingerprints with status='firing'
SELECT status, COUNT(DISTINCT fingerprint) as count FROM alerts GROUP BY status
```

### Frontend Changes

#### Alerts page (`Alerts.tsx`)

- Default to grouped mode. Fetch `GET /api/v1/alerts/active?grouped=true`.
- Each row shows: severity badge, alert name, source, "共 N 次" tag, last triggered time.
- Clicking a row expands to show the full history: all alert records with this fingerprint, sorted by trigger_time descending.
- The expander already exists — reuse it to show the occurrence list instead of just message+labels.

#### Dashboard page (`Dashboard.tsx`)

- Active alerts section also uses grouped mode.
- Each `AlertCard` shows occurrence count if > 1.

#### Alert detail / ack / silence

- Ack and quick-silence operate on the `latest_alert` (the currently-firing instance).
- When acking a grouped alert, the UI shows "确认此告警（含 N 次触发）" to make it clear what's being acked.

---

## 2. Escalation (Timeout Re-notification)

### Problem

An alert is **firing** and nobody acks it. It stays in `firing` forever (until CleanAlerts auto-resolves it after 1 hour). There is no mechanism to re-notify operators who may have missed the first notification.

### Data Model

#### Alert table additions

```go
LastNotifiedAt  *time.Time `gorm:"index" json:"last_notified_at"`  // timestamp of last notification sent
NotifyCount     int        `gorm:"default:0" json:"notify_count"`  // number of notifications sent for this alert
```

- `LastNotifiedAt`: Set to `now()` when the first notification is sent. Updated on each re-notification. Null means no notification has been sent yet.
- `NotifyCount`: Incremented on each notification (first + re-notifications). Starts at 0, set to 1 on first notification. Used to enforce `escalation_max_repeats`.

#### RouteRule table additions

```go
EscalationEnabled    bool   `gorm:"default:false" json:"escalation_enabled"`
EscalationTimeout    int    `gorm:"default:30" json:"escalation_timeout"`       // minutes
EscalationMaxRepeats int    `gorm:"default:3" json:"escalation_max_repeats"`    // max re-notification count
```

- `EscalationEnabled`: Whether timeout re-notification is active for this route rule.
- `EscalationTimeout`: How many minutes to wait before re-notifying. Default 30 minutes.
- `EscalationMaxRepeats`: Maximum number of re-notifications. Default 3 (so total notifications = 1 initial + 3 re-notifications = 4 max).

### Backend Changes

#### Webhook handler — set LastNotifiedAt on first notification

In `sendChannelNotification()` (webhook.go), after the first successful notification attempt:

```go
alert.LastNotifiedAt = &now
alert.NotifyCount = 1
h.db.Save(&alert)
```

Only set on the **first** notification for a new alert. Deduplicated alerts (which skip notification) should not update these fields.

#### New background task — EscalationChecker

Add a new goroutine in `cmd/server/main.go` that runs every 1 minute:

```go
func runEscalationChecker(db *gorm.DB, deliverySvc *delivery.Service, matcher *routing.Matcher, logger *log.Logger) {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        checkAndEscalate(db, deliverySvc, matcher, logger)
    }
}
```

`checkAndEscalate` logic:

1. Load all enabled RouteRules with `escalation_enabled = true` into a map keyed by ID for fast lookup.
2. Query alerts where:
   - `status = 'firing'`
   - `acked_at IS NULL`
   - `last_notified_at IS NOT NULL`
3. For each alert:
   - Re-run route matching via `Matcher.FindMatchedTargets()` to find the current channel targets and the matching RouteRule.
   - If the matched RouteRule has `escalation_enabled = true`:
     - Check if `now() - last_notified_at >= escalation_timeout (in minutes)`.
     - Check if `notify_count < escalation_max_repeats`.
     - If both conditions met:
       - Send notification via the delivery pipeline.
       - Update `last_notified_at = now()`, `notify_count++`.
       - Create a new `NotificationDelivery` record with `trigger_kind = "escalation"`.
   - If no route rule matches or escalation is disabled: skip.

#### TriggerKind addition

Add `"escalation"` to the valid `TriggerKind` values in `NotificationDeliveryAttempt`. Current values: `pipeline`, `retry`, `replay`. New: `escalation`. This is a string field validated in `RecordAttempt()` — add `"escalation"` to the allowed list.

### Frontend Changes

#### Alert list / detail

- Show `notify_count` next to the alert: "已通知 2 次" tag.
- If `notify_count >= escalation_max_repeats`: show "已达通知上限" tag in warning color.

#### Route rule edit form

- Add "超时重通知" section:
  - Toggle switch for `escalation_enabled`.
  - InputNumber for `escalation_timeout` (minutes, min 5, max 120, default 30).
  - InputNumber for `escalation_max_repeats` (min 1, max 10, default 3).
- Only shown when the route rule is enabled.

#### Dashboard stats

- Add a new stat card or expand existing: "待确认告警" — count of firing alerts with `acked_at IS NULL` and `notify_count > 0`.

---

## 3. Delivery Observability (Alert Detail Embedded)

### Problem

Operators cannot see whether a specific alert was successfully delivered, to which channels, and how many attempts were made. The delivery data exists in the backend but requires navigating to a separate page with manual filtering.

### Backend Changes

#### New API endpoint

```
GET /api/v1/alerts/:id/deliveries
```

Response:

```json
[
  {
    "id": 42,
    "channel_snapshot": { "id": 1, "name": "钉钉运维群", "type": "dingtalk", "enabled": true },
    "delivery_status": "delivered",
    "delivery_mode": "rendered",
    "attempt_count": 1,
    "last_success_at": "2026-05-18T10:05:00Z",
    "final_failure_summary": null,
    "trigger_kind": "pipeline",
    "created_at": "2026-05-18T10:04:55Z"
  }
]
```

Each delivery includes its attempts (preloaded, ordered by attempt_number).

#### Implementation

In `AlertHandler`:

```go
func (h *AlertHandler) AlertDeliveries(c *gin.Context) {
    alertID := c.Param("id")
    var deliveries []models.NotificationDelivery
    err := h.db.Where("alert_id = ?", alertID).
        Preload("Attempts", func(db *gorm.DB) *gorm.DB {
            return db.Order("attempt_number ASC")
        }).
        Order("created_at DESC").
        Find(&deliveries).Error
    // ... return as JSON
}
```

### Frontend Changes

#### Alerts page — expanded row detail

The existing row expander currently shows: message + labels.

Enhance it to show three sections:

1. **告警信息** (existing): message, labels, trigger_time, fingerprint
2. **投递记录** (new): table of deliveries for this alert
   - Columns: 渠道, 状态, 尝试次数, 通知时间, 触发类型
   - Status tags: 成功(green), 失败(red), 处理中(blue)
   - Failed deliveries show error message in red text
   - Trigger kind tags: 首次通知, 重试, 重放, 升级通知
3. **操作** (existing): ack/silence buttons (already present)

#### Data fetching

When a row is expanded, call `GET /api/v1/alerts/:id/deliveries` and cache the result in component state. Only fetch on first expand (lazy load).

---

## 4. OnDuty Removal

### Scope

Remove all code that references the OnDuty feature. The `on_duties` database table is NOT dropped (preserve production data), but no code reads or writes it.

### Backend removal

| File | What to remove |
|------|---------------|
| `internal/models/models.go` | `OnDuty` struct, `Validate()` method |
| `internal/handlers/config.go` | All 6 OnDuty handler methods (ListOnDuty, GetOnDuty, CreateOnDuty, UpdateOnDuty, DeleteOnDuty, CurrentOnDuty) |
| `internal/router/router.go` | `onduty` route group (6 routes) |
| `internal/database/postgres.go` | `&models.OnDuty{}` from migration list |

### Frontend removal

| File | What to remove |
|------|---------------|
| `frontend/src/pages/OnDuty.tsx` | Entire file (delete) |
| `frontend/src/pages/index.ts` | `OnDutyPage` export line |
| `frontend/src/App.tsx` | OnDuty import and `<Route path="/onduty">` element |
| `frontend/src/components/layout/AppSidebar.tsx` | `{ key: '/onduty', ... }` menu item |
| `frontend/src/types/index.ts` | `OnDuty` interface |
| `frontend/src/api/client.ts` | `onDutyApi` object and `OnDuty` type import |
| `frontend/src/stores/configStore.ts` | OnDuty state fields (onDutyList, currentOnDuty, onDutyLoading) and all OnDuty actions (fetchOnDuty, createOnDuty, updateOnDuty, deleteOnDuty) |

---

## Implementation Order

1. **OnDuty removal** — simplest, no dependencies, clears the way
2. **Alert aggregation** — changes how alerts are displayed, foundational for the other features
3. **Delivery observability** — builds on the existing delivery infrastructure, no new backend services needed
4. **Escalation** — most complex (new background task, new fields, new route rule config), depends on aggregation being done first (so escalation works on grouped alerts, not individual rows)

---

## Testing Strategy

- **OnDuty removal**: Verify no compile errors, no remaining references. Existing tests for other features still pass.
- **Aggregation**: Backend unit test for grouped query logic. Frontend test for grouped display.
- **Delivery observability**: Backend unit test for `AlertDeliveries` endpoint. Frontend test for expanded row with delivery table.
- **Escalation**: Backend unit test for `checkAndEscalate` logic (time thresholds, notify_count limits, route rule matching). Integration test for the background task. Frontend test for route rule escalation form fields.