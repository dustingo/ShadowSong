---
phase: 07-lock-down-protected-operations
plan: 01
subsystem: authz
tags: [router, capabilities, alerts, config, permissions]
requirements-completed: [PERM-01, PERM-02, PERM-03, PERM-04]
completed: 2026-04-12
---

# Phase 07 Plan 01 Summary

## Accomplishments
- Added explicit capability guards to alert action routes so only roles with `process_alerts` can acknowledge or quick-silence alerts.
- Restricted every configuration write endpoint to `manage_config`, while preserving authenticated read access for `admin`, `operator`, and `viewer`.
- Expanded router coverage to lock in representative allow/deny outcomes for alert actions, config reads and writes, user management, and 401 versus 403 boundaries.

## Key Files
- `internal/router/router.go`
- `internal/router/router_test.go`
- `internal/middleware/authorize.go`
- `internal/authz/capabilities.go`
- `internal/authz/capabilities_test.go`

## Verification
- `go test ./internal/router ./internal/middleware ./internal/authz -run "Test(Router|Capability)" -count=1 -timeout 60s`

## Notes
- `operator` can acknowledge and quick-silence alerts but cannot mutate datasource or on-duty configuration.
- `viewer` remains backend read-only and now receives consistent `403 {"error":"insufficient permissions"}` on protected writes.
