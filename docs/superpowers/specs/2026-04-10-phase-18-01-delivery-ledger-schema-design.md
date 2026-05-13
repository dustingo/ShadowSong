---
name: phase-18-01
description: Establish delivery ledger dual-table schema and immutable snapshots
metadata:
  type: spec
  source_phase: 18-establish-delivery-ledger
  source_plan: "01"
  milestone: v1.4
  status: completed
  completed: 2026-04-30
---

# Phase 18 Plan 01: Delivery Ledger Schema and Service

## Context & Goals

Phase 18 establishes the delivery ledger as the source of truth for all notification deliveries. This plan (18-01) locks down the dual-table schema, immutable snapshot contracts, and the focused delivery service before the hot path integration and read-only API are built on top.

The goal is to satisfy requirements DELV-02 and DELV-06 by creating PostgreSQL-backed truth for every alert×channel delivery, with frozen snapshots sufficient to support future audit and single-item retry/replay without depending on runtime configuration state.

## Success Criteria

- System has a PostgreSQL source of truth for every alert×channel delivery, not just logs
- Ledger snapshots freeze alert, channel, route, delivery mode, final rendered content, and terminal failure summary — sufficient for audit and single-item retry/replay
- Attempt history is append-only明细; delivery main record only aggregates terminal state, never overwrites historical attempts

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| notification_deliveries + notification_delivery_attempts models | `internal/models/notification_delivery.go` | Dual-table schema, status enums, JSONB snapshot fields |
| delivery focused service | `internal/delivery/service.go` | Create delivery, record attempt, update delivered/failed, read details |
| ledger table migration registration | `internal/database/postgres.go` | Tables纳入现有 GORM 迁移入口 |

## Architecture

### Dual-Table Schema

**notification_deliveries** (main record):
- `alert_id`, `trace_id`, `channel_id`, `route_rule_id`
- `delivery_status`: `pending | delivered | failed`
- `delivery_mode`: `channel_default | force_email | force_sms | force_webhook | default`
- `attempt_count`, `trigger_kind`: `pipeline | retry | replay`
- `final_failure_summary`: JSONB with `error`, `is_retryable`, `stage`, `attempt_number`
- `alert_snapshot`, `channel_snapshot`, `route_snapshot`, `rendered_payload_snapshot`: JSONB frozen copies
- `last_attempt_at`, `last_success_at`

**notification_delivery_attempts** (append-only明细):
- `delivery_id`, `attempt_number`, `result`, `retryable`
- `error_message`, `http_status`, `duration_ms`, `trigger_kind`, `created_at`

### Snapshot Contract

Snapshots contain only identity and rendered content:
- Excluded: channel `APIKey`, secrets, full `Config` with sensitive values
- Included: alert identity fields, channel identity, route identity, rendered title/content

### Indexes

- `alert_id`, `trace_id`, `channel_id`, `delivery_status`, `created_at`
- Unique: `delivery_id + attempt_number`

## Implementation Tasks

### Task 1: Establish Ledger Dual-Table Models and Migration Registration

**Files:** `internal/models/notification_delivery.go`, `internal/database/postgres.go`, `internal/models/notification_delivery_test.go`

**Acceptance Criteria:**
- `rg -n "type NotificationDelivery struct|type NotificationDeliveryAttempt struct|AlertSnapshot|ChannelSnapshot|RouteSnapshot|RenderedPayloadSnapshot|FinalFailureSummary" internal/models/notification_delivery.go` hits all core fields
- `rg -n "NotificationDelivery|NotificationDeliveryAttempt" internal/database/postgres.go` shows both tables added to migration list
- `go test ./internal/models -count=1` passes

**Action:** Create `internal/models/notification_delivery.go` with `NotificationDelivery` and `NotificationDeliveryAttempt` models per D-01/D-02/D-04/D-05. Use `datatypes.JSON`/`jsonb` for snapshots; explicitly exclude channel secrets and API keys. Add minimal indexes. Register both tables in `internal/database/postgres.go`. Write tests first to verify status enums, default fields, and append-only attempt constraints.

**Verification:** `go test ./internal/models -count=1`

---

### Task 2: Implement Focused Delivery Service as Ledger Write/Read Source

**Files:** `internal/delivery/service.go`, `internal/delivery/service_test.go`

**Acceptance Criteria:**
- `rg -n "func \\(s \\*Service\\) (StartDelivery|RecordAttempt|MarkDelivered|MarkFailed|GetDeliveryByID|ListDeliveries)" internal/delivery/service.go` hits all service methods
- `rg -n "trigger_kind|delivery_mode|final_failure_summary|RenderedPayloadSnapshot" internal/delivery/service.go` hits snapshot and terminal field write logic
- `go test ./internal/delivery -count=1` passes

**Action:** Create `internal/delivery/service.go` with `Service { db *gorm.DB }` and explicit methods:
- `StartDelivery`: freezes AlertSnapshot, ChannelSnapshot, RouteSnapshot, RenderedPayloadSnapshot + delivery_mode/trigger_kind
- `RecordAttempt`: appends attempt明细 only, never overwrites history
- `MarkDelivered`: updates main record to success terminal state
- `MarkFailed`: updates to failure terminal state, writes final_failure_summary with retryable distinction
- `GetDeliveryByID` / `ListDeliveries`: returns aggregated objects with preloaded attempts

**Verification:** `go test ./internal/delivery -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-18-01 | I | snapshot fields | mitigate (blocking) | Snapshots only保存 alert key fields, channel identity, route identity, rendered content; explicitly exclude APIKey, channel secret, full Config values |
| T-18-02 | R | attempt history | mitigate (blocking) | Attempts use `delivery_id + attempt_number` unique constraint; main record only maintains aggregated terminal state |
| T-18-03 | T | GORM writes | mitigate | All writes via parameterized GORM; status enums validated at service layer |

## Established Patterns

- **Pattern 1:** Append-only attempt明细 with unique constraint prevents history overwrite
- **Pattern 2:** Snapshot frozen at StartDelivery time, not re-read at failure time
- **Pattern 3:** Delivery status is aggregated terminal state; attempts are individual明细

## Decisions

- Snapshot content does NOT include channel config secrets or API keys — only identity + rendered content
- delivery_mode is preserved from the original routing decision, not re-evaluated at retry time
- trigger_kind distinguishes `pipeline` (original), `retry` (reuses frozen payload), `replay` (re-routes with current config)

## Deviation Log

None — plan executed as written.
