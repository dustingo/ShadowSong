---
name: phase-10-01
description: Secure backend websocket entrypoint with config-driven origin policy and JWT-backed handshake validation
metadata:
  type: spec
  source_phase: 10-secure-realtime-alert-access
  source_plan: "01"
  milestone: v1.2
  status: completed
  completed: 2026-04-21
---

# Phase 10 Plan 01: Secure Realtime Alert Access

## Context & Goals

Close the public `/ws/alerts` access surface without introducing a separate auth model or breaking current alert visibility flows.

Purpose: Close the public `/ws/alerts` access surface without introducing a separate auth model or breaking current alert visibility flows.
Output: Config-driven websocket origin policy, JWT-backed handshake validation, and backend tests covering allow/deny cases.

## Success Criteria

- Unauthenticated requests cannot open /ws/alerts
- Origins outside the configured allowlist are rejected during the websocket handshake
- Authenticated allowed-origin clients still receive the initial firing-alert snapshot

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| WebSocket auth and origin enforcement | `internal/handlers/websocket.go` | WebSocket auth and origin enforcement |
| Allowed websocket origin configuration | `internal/config/config.go` | Allowed websocket origin configuration |
| Allow and deny coverage for websocket access | `internal/router/router_test.go` | Allow and deny coverage for websocket access |

## Architecture

### Key Architectural Decisions

- `AllowedOrigins []string` added to `config.ServerConfig` populated from `ALLOWED_ORIGINS` environment variable, defaulting to localhost and 127.0.0.1 dev origins when unset
- Websocket handling no longer mounts as an unprotected top-level closure; constructed with access to config, JWT auth, and database-backed user validation
- Replaced package-level `CheckOrigin: return true` upgrader with handler-owned logic that allows only configured origins
- JWT validation on websocket handshake using concrete browser-compatible contract: read `token` from websocket URL query string

## Implementation Tasks

### Task 1: Add config-driven websocket origin policy and authenticated route wiring

**Files:** `internal/config/config.go`, `internal/router/router.go`, `internal/handlers/websocket.go`

**Acceptance Criteria:**
- internal/config/config.go contains `AllowedOrigins []string`
- internal/config/config.go contains `ALLOWED_ORIGINS`
- internal/handlers/websocket.go does not contain `return true // Allow all origins in development`
- internal/handlers/websocket.go contains `c.Query("token")`
- internal/router/router.go no longer exposes `/ws/alerts` as an unprotected inline closure

**Action:** Add `AllowedOrigins []string` to `config.ServerConfig` in `internal/config/config.go` and populate it from a comma-separated `ALLOWED_ORIGINS` environment variable, defaulting to localhost and 127.0.0.1 dev origins when unset. Update `router.Setup` so websocket handling no longer mounts as an unprotected top-level closure; instead, construct the websocket handler with access to config, JWT auth, and database-backed user validation. In `internal/handlers/websocket.go`, replace the package-level `CheckOrigin: return true` upgrader with handler-owned logic that allows only configured origins, permits empty origin only when explicitly treated as non-browser traffic, and rejects everything else. Require a JWT on websocket handshake using a concrete browser-compatible contract: read `token` from the websocket URL query string, validate it with the same `auth.JWT` and user-state rules used by `middleware.JWTAuth`, and reject disabled, force-reset, stale-session, malformed, missing, or unsupported-role users before upgrade. Keep the existing initial snapshot behavior and client registry intact.

**Verification:** `go test ./internal/router ./internal/middleware`

---

### Task 2: Add backend websocket allow and deny coverage

**Files:** `internal/router/router_test.go`

**Acceptance Criteria:**
- internal/router/router_test.go contains `/ws/alerts`
- internal/router/router_test.go contains a test for missing or invalid websocket token
- internal/router/router_test.go contains a test for rejected origin
- `go test ./internal/router` exits 0

**Action:** Extend `internal/router/router_test.go` with websocket-focused tests that exercise `/ws/alerts` using a real handshake request. Cover at least these cases: missing token returns unauthorized, invalid or unsupported origin is rejected during upgrade, and a valid token from an allowed origin upgrades successfully and receives the `init` payload when firing alerts exist. Reuse the existing sqlite test database pattern and `auth.NewJWT` helper instead of inventing new fixtures.

**Verification:** `go test ./internal/router -run WebSocket`

## Security Considerations

None

## Established Patterns

None

## Decisions

- Config-driven origin policy rather than hardcoded in handler comments
- JWT validation on websocket handshake using URL query string token
- Reuse existing sqlite test database pattern and auth helpers for test fixtures

## Deviation Log

None
