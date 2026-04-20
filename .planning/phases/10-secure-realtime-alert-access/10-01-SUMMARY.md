---
phase: 10-secure-realtime-alert-access
plan: 01
subsystem: api
tags: [websocket, jwt, gin, authz, origin]
requires:
  - phase: 09-close-v1-1-milestone-tech-debt
    provides: "stable capability-based auth baseline and updated planning truth"
provides:
  - "Config-driven websocket origin allowlist"
  - "JWT-backed websocket handshake validation"
  - "Backend allow/deny coverage for /ws/alerts"
affects: [dashboard, realtime, security, phase-11, phase-12]
tech-stack:
  added: []
  patterns: ["websocket handshake reuses existing JWT/user-state validation", "origin allowlist loaded from server config"]
key-files:
  created: []
  modified: [internal/config/config.go, internal/middleware/auth.go, internal/handlers/websocket.go, internal/router/router.go, internal/router/router_test.go]
key-decisions:
  - "WebSocket auth reuses existing JWT and account-state validation rather than adding a second token type"
  - "Allowed origins are config-driven with localhost defaults instead of handler hardcoding"
patterns-established:
  - "Realtime routes should share REST auth truth through middleware.AuthenticateToken"
  - "WebSocket allow/deny behavior gets explicit router-level handshake tests"
requirements-completed: [RTAL-01, RTAL-02, RTAL-03]
duration: 35min
completed: 2026-04-20
---

# Phase 10: Secure Realtime Alert Access Summary

**Config-driven websocket origin enforcement with JWT-backed handshake validation and backend allow/deny regression coverage**

## Performance

- **Duration:** 35 min
- **Started:** 2026-04-20T17:41:00+08:00
- **Completed:** 2026-04-20T17:57:00+08:00
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- `/ws/alerts` no longer accepts anonymous websocket handshakes
- WebSocket origin checks moved from `return true` to an explicit allowlist sourced from server config
- Backend tests now cover missing token, blocked origin, and valid realtime handshake cases

## Task Commits

Each task was committed atomically:

1. **Task 1: Add config-driven websocket origin policy and authenticated route wiring** - `1fb2a7f` (fix)
2. **Task 2: Add backend websocket allow and deny coverage** - `1fb2a7f` (fix)

## Files Created/Modified
- `internal/config/config.go` - adds `AllowedOrigins` and `ALLOWED_ORIGINS` parsing
- `internal/middleware/auth.go` - extracts reusable token authentication for websocket handshakes
- `internal/handlers/websocket.go` - enforces token and origin checks before websocket upgrade
- `internal/router/router.go` - wires `/ws/alerts` directly to the secured websocket handler
- `internal/router/router_test.go` - verifies websocket allow/deny behavior

## Decisions Made
- Reused the existing JWT/user-state validation path instead of inventing a websocket-only auth contract
- Required a query-string token for the browser websocket handshake because it is browser-compatible and minimal to adopt
- Treated empty origin as disallowed to close the browser-exposed access surface by default

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `internal/router` tests initially hit sqlite in-memory locking because multiple tests shared the same DSN; this was resolved by isolating the DSN per test name in `newRouterTestDB`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Backend realtime security contract is now explicit: allowed origin + JWT token required
- Frontend can now adopt the authenticated websocket URL shape without reopening backend security questions

---
*Phase: 10-secure-realtime-alert-access*
*Completed: 2026-04-20*
