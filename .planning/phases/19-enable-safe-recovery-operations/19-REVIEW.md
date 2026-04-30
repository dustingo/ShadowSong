---
phase: 19-enable-safe-recovery-operations
reviewed: 2026-05-01T00:00:00Z
depth: standard
files_reviewed: 17
files_reviewed_list:
  - frontend/src/App.tsx
  - frontend/src/api/client.ts
  - frontend/src/authz/capabilities.ts
  - frontend/src/pages/Alerts.tsx
  - frontend/src/pages/Alerts.test.tsx
  - frontend/src/pages/Deliveries.tsx
  - frontend/src/pages/Deliveries.test.tsx
  - frontend/src/pages/index.ts
  - frontend/src/types/index.ts
  - internal/database/postgres.go
  - internal/delivery/service.go
  - internal/delivery/service_test.go
  - internal/handlers/delivery.go
  - internal/handlers/delivery_test.go
  - internal/models/delivery_recovery.go
  - internal/models/delivery_recovery_test.go
  - internal/router/router.go
findings:
  critical: 1
  warning: 2
  info: 0
  total: 3
status: issues_found
---
# Phase 19: Code Review Report

**Reviewed:** 2026-05-01T00:00:00Z
**Depth:** standard
**Files Reviewed:** 17
**Status:** issues_found

## Summary

Reviewed the delivery recovery backend, delivery UI, router wiring, and related tests. The main risks are in the recovery execution path: CORS origin validation is too permissive, duplicate recoveries can still slip through under concurrency, and retry/replay side effects are executed before the surrounding transaction is durably committed.

## Critical Issues

### CR-01: Prefix-based CORS allowlist accepts attacker-controlled origins

**File:** `internal/router/router.go:22-24`
**Issue:** The CORS check uses `strings.HasPrefix(origin, "http://127.0.0.1")` and `strings.HasPrefix(origin, "http://localhost")`. That also matches origins such as `http://localhost.attacker.tld`, so the server will reflect that hostile origin into `Access-Control-Allow-Origin` while also setting `Access-Control-Allow-Credentials: true`. Any browser holding a valid session/token for the app can then be driven into authenticated cross-origin requests from an attacker-controlled site.
**Fix:**
```go
origin := c.GetHeader("Origin")
if origin != "" {
	parsed, err := url.Parse(origin)
	if err == nil {
		host := parsed.Hostname()
		if (parsed.Scheme == "http" || parsed.Scheme == "https") &&
			(host == "localhost" || host == "127.0.0.1") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
}
```

## Warnings

### WR-01: Concurrent retry/replay requests can both pass the "active recovery" check

**File:** `internal/delivery/service.go:602-628`  
**Related File:** `internal/models/delivery_recovery.go:28-42`
**Issue:** `startRecoveryRecord` does a `COUNT(*)` for `pending`/`in_progress` rows and then inserts a new recovery record. Without a lock or database-enforced uniqueness, two concurrent requests can both observe `activeCount == 0` and both create runnable recovery rows. That defeats the "safe recovery" guard and can produce duplicate notifications.
**Fix:** Enforce single-active recovery in the database and make the code handle the conflict deterministically. For example, add a partial unique index on `original_delivery_id` for active statuses, or lock the original delivery row before checking/inserting:
```go
// migration example
CREATE UNIQUE INDEX uniq_active_delivery_recovery
ON notification_delivery_recoveries (original_delivery_id)
WHERE status IN ('pending', 'in_progress');
```

### WR-02: External send happens inside the database transaction, so commit failure can leave an untracked notification

**File:** `internal/delivery/service.go:385-442`  
**Related File:** `internal/delivery/service.go:462-531`  
**Related File:** `internal/delivery/service.go:752-819`
**Issue:** Both `RetryDelivery` and `ReplayDelivery` call `executeRecoveredDelivery` inside the transaction, and `executeRecoveredDelivery` performs the real `sendToChannel` call before the transaction commits. If any later `Save`/commit step fails, the API returns an error and the recovery row may roll back, but the notification may already have been delivered externally. Operators can then retry again, causing duplicate outbound sends with no durable audit trail for the first one.
**Fix:** Move side effects out of the DB transaction. A safe pattern is:
```go
// tx 1: reserve recovery + create pending delivery row, then commit
// external send outside transaction
// tx 2: persist attempt/result and finalize recovery status
```
Alternatively, use an outbox/job table and let a worker send only after the reservation transaction commits.

---

_Reviewed: 2026-05-01T00:00:00Z_  
_Reviewer: Claude (gsd-code-reviewer)_  
_Depth: standard_
