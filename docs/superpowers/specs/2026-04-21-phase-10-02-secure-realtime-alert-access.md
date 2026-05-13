---
name: phase-10-02
description: Update dashboard realtime client to send JWT token during websocket handshake
metadata:
  type: spec
  source_phase: 10-secure-realtime-alert-access
  source_plan: "02"
  milestone: v1.2
  status: completed
  completed: 2026-04-21
---

# Phase 10 Plan 02: Dashboard Realtime Client Auth

## Context & Goals

Keep the existing realtime UX working after backend hardening without redesigning the dashboard or removing the polling fallback.

Purpose: Keep the existing realtime UX working after backend hardening without redesigning the dashboard or removing the polling fallback.
Output: Dashboard websocket handshake includes the JWT token, and frontend tests cover the new connection contract.

## Success Criteria

- The dashboard websocket client sends authentication when connecting to /ws/alerts
- A successful websocket handshake still sets the UI connected state
- Connection failures caused by auth or origin rejection do not break the existing polling fallback

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Authenticated realtime dashboard client | `frontend/src/pages/Dashboard.tsx` | Authenticated realtime dashboard client |
| Frontend coverage for authenticated websocket URL and connection state behavior | `frontend/src/pages/Dashboard.test.tsx` | Frontend coverage for authenticated websocket URL and connection state behavior |

## Architecture

### Key Architectural Decisions

- JWT token sent via query string (`/ws/alerts?token=${encodeURIComponent(token)}`)
- Token pulled from existing user store or localStorage source of truth
- Missing token skips websocket connection setup, marks realtime as disconnected, relies on polling fallback

## Implementation Tasks

### Task 1: Send JWT during dashboard websocket handshake without changing the visible realtime UX

**Files:** `frontend/src/pages/Dashboard.tsx`

**Acceptance Criteria:**
- frontend/src/pages/Dashboard.tsx contains `ws/alerts?token=`
- frontend/src/pages/Dashboard.tsx still contains `setWsConnected(true)`
- frontend/src/pages/Dashboard.tsx still contains `setInterval(() => {`
- frontend/src/pages/Dashboard.tsx handles the missing-token path without calling `new WebSocket` blindly

**Action:** Update the websocket URL construction in `frontend/src/pages/Dashboard.tsx` so the browser connects to `/ws/alerts?token=${encodeURIComponent(token)}` using the same persisted auth token the SPA already stores for REST calls. Pull the token from the existing user store or its localStorage source of truth rather than introducing a second auth cache. Preserve the current reconnect behavior, polling fallback, `setWsConnected(true/false)`, and alert message handling. If no token exists, skip websocket connection setup, mark realtime as disconnected, and rely on the existing polling refresh path instead of throwing an error.

**Verification:** `pnpm test -- --run Dashboard`

---

### Task 2: Add frontend coverage for the authenticated websocket contract

**Files:** `frontend/src/pages/Dashboard.test.tsx`

**Acceptance Criteria:**
- frontend/src/pages/Dashboard.test.tsx exists
- frontend/src/pages/Dashboard.test.tsx contains `ws/alerts?token=`
- frontend/src/pages/Dashboard.test.tsx contains a case for missing token
- `pnpm test -- --run Dashboard` exits 0

**Action:** Create `frontend/src/pages/Dashboard.test.tsx` to verify that the dashboard builds a websocket URL containing `/ws/alerts?token=` when a token is present, updates the connected state on successful open, and avoids opening an unauthenticated websocket when no token is available. Mock the browser `WebSocket`, store state, and polling APIs in the same Vitest style already used elsewhere in `frontend/src/pages/*.test.tsx`.

**Verification:** `pnpm test -- --run Dashboard`

## Security Considerations

None

## Established Patterns

None

## Decisions

- JWT token via query string rather than headers (browser WebSocket limitation)
- Reuse existing user store for token rather than introducing second auth cache

## Deviation Log

None
