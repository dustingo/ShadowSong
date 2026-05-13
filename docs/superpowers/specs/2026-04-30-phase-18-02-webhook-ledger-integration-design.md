---
name: phase-18-02
description: Integrate delivery ledger into webhook notification hot path
metadata:
  type: spec
  source_phase: 18-establish-delivery-ledger
  source_plan: "02"
  milestone: v1.4
  status: completed
  completed: 2026-04-30
---

# Phase 18 Plan 02: Webhook Hot Path Ledger Integration

## Context & Goals

This plan connects the delivery ledger (established in Plan 18-01) to the existing webhook notification hot path. The existing bounded 3-attempt retry logic and Phase 14-17 logging contracts must remain unchanged. The ledger writes are additive — they do not replace existing logs, nor do they alter retry semantics.

**Goal:** Satisfy DELV-02 and DELV-06 by ensuring every real send attempt is written to the ledger, and notifications that exhaust retries leave a persistent failure record (not just logs).

## Success Criteria

- Existing bounded 3-attempt send retry is preserved; every real attempt is dual-written to the ledger
- Success, non-retryable failure, and retry-exhausted results all update delivery terminal state
- Ledger writes maintain consistency with existing `trace_id` / `send_attempt` / `terminal_failure` contracts
- Operators can cross-reference between logs and database for any delivery

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| webhook hot path with delivery service integration | `internal/handlers/webhook.go` | Ledger dual-write in send path |
| ledger dual-write and retry boundary regression tests | `internal/handlers/webhook_test.go` | Coverage for success, fallback, terminal failure ledger states |

## Architecture

### Integration Points

The ledger is integrated into `internal/handlers/webhook.go` at these points in the `processAlertNotifications` / `sendNotification` / `sendChannelNotification` chain:

1. **StartDelivery** — called when `alert × channel` enters send path; snapshots trace_id, alert_id, channel_id, route_rule_id (when available), delivery_mode
2. **RecordAttempt** — called after every real send attempt
3. **MarkDelivered** — called on success
4. **MarkFailed(retryable=false)** — called immediately on non-retryable failure
5. **MarkFailed(retryable=true, terminal=true)** — called in existing `terminal_failure` branch when `attempt == notificationMaxAttempts`

### Preserved Contracts

- `notificationMaxAttempts = 3` remains unchanged
- `notifier.IsRetryableSendError` determination unchanged
- `send_attempt` / `send_notification` / `terminal_failure` log names and fields unchanged
- No MQ, worker, or background replay introduced

## Implementation Tasks

### Task 1: Integrate Delivery Service into Webhook Notification Hot Path

**Files:** `internal/handlers/webhook.go`

**Acceptance Criteria:**
- `rg -n "deliveryService|StartDelivery|RecordAttempt|MarkDelivered|MarkFailed" internal/handlers/webhook.go` hits ledger integration code
- `rg -n "notificationMaxAttempts = 3|stage=terminal_failure|stage=send_attempt" docs/alert-path-operations-runbook.md internal/handlers/webhook.go` shows bounded retry contract preserved
- `go test ./internal/handlers -run "TestWebhookHandler(.*Delivery.*|.*Retry.*|.*Terminal.*)" -count=1` passes

**Action:** Add `deliveryService` dependency to `WebhookHandler`, initialized in `NewWebhookHandler` using Plan 01's `delivery.Service`. Integration points: StartDelivery before send, RecordAttempt after each real attempt, MarkDelivered on success, MarkFailed on non-retryable failure, MarkFailed in terminal_failure branch. Preserve all existing retry limits, error classification, and logging — ledger writes are additive.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Delivery.*|.*Retry.*|.*Terminal.*)" -count=1`

---

### Task 2: Expand Webhook Regression Tests for Success, Fallback, and Terminal Failure Ledger States

**Files:** `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "Delivery" internal/handlers/webhook_test.go` hits new ledger assertion test names
- `go test ./internal/handlers -run "TestWebhookHandler(.*Delivery.*|.*Fallback.*|.*Terminal.*|.*Retry.*)" -count=1` passes
- New tests assert at minimum: `attempt_count`, `delivery_status`, `delivery_mode`, `final_failure_summary`, `trace_id`, attempt count

**Action:** Add or refactor tests in `internal/handlers/webhook_test.go` using sqlite + GORM migrations including NotificationDelivery and NotificationDeliveryAttempt. Cover three minimal regressions:
1. First success: 1 delivery record, 1 success attempt, `delivery_status=delivered`
2. Fallback to default mode: ledger shows `delivery_mode=default` and `rendered_payload_snapshot` with actual sent content
3. Retry-exhausted: delivery terminal state is failure, 3 attempt records, `final_failure_summary` consistent with terminal_failure branch

Keep existing Phase 15/16 log assertions present — regression focus stays on log contracts.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Delivery.*|.*Fallback.*|.*Terminal.*|.*Retry.*)" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-18-04 | R | terminal failure path | mitigate (blocking) | In existing `terminal_failure` branch (`attempt == notificationMaxAttempts`), synchronously call `MarkFailed`; regression test asserts DB failure terminal state exists |
| T-18-05 | D | send path writes | mitigate | Ledger writes are lightweight synchronous persistence; no change to retry limits or log ordering |
| T-18-06 | T | snapshot creation | mitigate | Snapshots from pre-send title/content and route/channel identity, not re-read from current config after failure |

## Established Patterns

- **Pattern 1:** Ledger writes are synchronous and additive — they do not change hot path behavior or retry semantics
- **Pattern 2:** Snapshot is frozen before the send attempt, using the content that was actually sent (or attempted)
- **Pattern 3:** `trigger_kind=pipeline` for original deliveries; retry/replay use their own trigger kinds (defined in Phase 19)

## Decisions

- Ledger writes do NOT replace or duplicate existing logging — both paths exist independently
- The bounded retry count (`notificationMaxAttempts = 3`) is preserved exactly; ledger does not introduce a separate retry mechanism
- Fallback delivery mode is recorded in the ledger at StartDelivery time based on the routing decision made at that moment

## Deviation Log

None — plan executed as written.
