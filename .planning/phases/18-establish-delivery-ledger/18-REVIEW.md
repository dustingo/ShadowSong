---
phase: 18-establish-delivery-ledger
reviewed: 2026-04-30T02:24:31Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - internal/models/notification_delivery.go
  - internal/models/notification_delivery_test.go
  - internal/delivery/service.go
  - internal/delivery/service_test.go
  - internal/handlers/webhook.go
  - internal/handlers/webhook_test.go
  - internal/handlers/delivery.go
  - internal/handlers/delivery_test.go
  - internal/router/router.go
findings:
  critical: 0
  warning: 2
  info: 1
  total: 3
status: issues_found
---

# Phase 18: Code Review Report

**Reviewed:** 2026-04-30T02:24:31Z
**Depth:** standard
**Files Reviewed:** 9
**Status:** issues_found

## Summary

Reviewed the Phase 18 delivery-ledger model, service, webhook integration, read API, and route wiring. The schema and handler surface are coherent, and the targeted tests pass, but the webhook integration currently treats ledger persistence as best-effort and drops structured HTTP status data from failed sends. Those two gaps break the phase's stated "ledger as source of truth" contract for some failure paths.

## Warnings

### WR-01: Ledger persistence failures are swallowed, allowing untracked sends

**File:** `internal/handlers/webhook.go:1115-1146`
**Issue:** `sendChannelNotification` logs `StartDelivery` failures and still continues sending. The same best-effort pattern is used for `RecordAttempt`, `MarkDelivered`, and `MarkFailed`, which only log persistence errors in `recordNotificationAttempt`, `markNotificationDelivered`, and `markNotificationFailed`. This means PostgreSQL can miss the delivery envelope, attempts, or terminal state while the notification is still sent, violating the phase contract that every `alert x channel` delivery has a PostgreSQL truth record.
**Fix:**
```go
deliveryRecord, err := h.startNotificationDelivery(alert, channel, routeRule, title, content, mode)
if err != nil {
  h.logAlertEvent("delivery_start", fields, "failed to persist delivery envelope")
  return
}

if err := h.recordNotificationAttempt(...); err != nil {
  h.markNotificationFailed(deliveryRecord, attempt, err, false, true)
  return
}
```
At minimum, fail closed when the envelope cannot be created. Prefer returning errors from the attempt/terminal write helpers as well, so the hot path cannot report success while the ledger is incomplete.

### WR-02: Failed delivery ledger entries never persist structured HTTP status

**File:** `internal/handlers/webhook.go:1188-1195`
**File:** `internal/handlers/webhook.go:1239-1245`
**Issue:** The webhook path records `ErrorMessage` and `Retryable`, but never sets `HTTPStatus` on either `NotificationDeliveryAttempt` or `FinalFailureSummary`. `internal/notifier/notifier.go:77-129` already parses trailing status codes to classify retryability, so 4xx/5xx failures like `status: 503` are currently reduced to plain strings in the ledger. That loses a key field that the schema and phase plan explicitly introduced for operator debugging and downstream recovery logic.
**Fix:**
```go
status := notificationHTTPStatus(sendErr)

_, err := h.deliveryService.RecordAttempt(..., delivery.RecordAttemptInput{
  HTTPStatus: status,
  // ...
})

err := h.deliveryService.MarkFailed(..., delivery.MarkFailedInput{
  HTTPStatus: status,
  // ...
})
```
Export a small helper from `internal/notifier` or add a local parser that reuses the same status extraction contract already used by `IsRetryableSendError`.

## Info

### IN-01: Model tests use a shared in-memory SQLite DSN

**File:** `internal/models/notification_delivery_test.go:111-114`
**Issue:** `newNotificationDeliveryTestDB` always opens `file:notification-delivery?mode=memory&cache=shared`. Unlike the delivery and webhook tests, this does not isolate state per test case, so future `t.Parallel()` use or additional tests in the package can leak rows across cases and create order-dependent failures.
**Fix:** Use a per-test DSN, for example `fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())`, matching the isolation approach already used in `internal/delivery/service_test.go` and `internal/handlers/webhook_test.go`.

---

_Reviewed: 2026-04-30T02:24:31Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
