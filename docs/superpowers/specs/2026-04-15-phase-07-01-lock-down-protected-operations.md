---
name: phase-07-01
description: Alert action and config mutation capability guards with HTTP 401/403 enforcement
metadata:
  type: spec
  source_phase: 07-lock-down-protected-operations
  source_plan: "01"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 07 Plan 01: Protected Route Capability Guards

## Context & Goals

Lock alert action and config interfaces down to the explicit capability matrix, establishing the Phase 7 backend permission surface.

Purpose: establish real access boundaries first, preventing audit logs from recording actions that should never have been allowed in the first place.
Output: route capability guards, role matrix tests, explicit 401/403 rejection results.

## Success Criteria

- Only `admin` can execute configuration write operations.
- `operator` can acknowledge and quick-silence alerts, but `viewer` cannot.
- Unauthorized requests receive explicit `403` from the backend, not just login-state pass-through.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Protected router routes | `internal/router/router.go` | Alert actions and config writes wired to explicit capability guards |
| HTTP allow/deny matrix | `internal/router/router_test.go` | HTTP allow/deny matrix for protected operations |

## Architecture

### Protected Route Configuration

Alert routes:
- `GET /alerts`, `/alerts/stats`, `/alerts/active`, `/alerts/:id` - authenticated access (alert viewing)
- `POST /alerts/:id/ack` - requires `CapabilityProcessAlerts`
- `POST /alerts/:id/quick-silence` - requires `CapabilityProcessAlerts`

Config routes:
- Read operations - `admin`, `operator`, `viewer` with `CapabilityViewConfig`
- Write operations (create/update/delete/toggle/test) - `admin` only with `CapabilityManageConfig`

### Key Decisions

- Alert action endpoints protected by `RequireCapability(authz.CapabilityProcessAlerts)`
- Config mutation endpoints protected by `RequireCapability(authz.CapabilityManageConfig)`
- Router tests assert exact HTTP status codes: 401 for unauthenticated, 403 for unauthorized, actual handler result for authorized

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| authenticated principal -> protected routes | Logged-in users may still attempt actions outside their role |
| route guard -> handler logic | Handler behavior must only execute after the correct capability check |

## Implementation Tasks

### Task 1: Wire alert-action endpoints to process-alert capability and keep read routes readable

**Files:** `internal/router/router.go`, `internal/router/router_test.go`

**Acceptance Criteria:**
- `internal/router/router.go` contains `RequireCapability(authz.CapabilityProcessAlerts)` on both alert action endpoints
- `internal/router/router_test.go` asserts viewer-denied and operator-allowed outcomes for `ack` and `quick-silence`
- Unauthorized requests still return `401`

**Action:** Update `internal/router/router.go` so alert read routes stay available to authenticated roles with alert-view ability, while `POST /alerts/:id/ack` and `POST /alerts/:id/quick-silence` require `middleware.RequireCapability(authz.CapabilityProcessAlerts)`. Add router tests proving `admin` and `operator` receive current handler result, while `viewer` receives `403 {"error":"insufficient permissions"}` and unauthenticated callers receive `401`.

**Verification:** `go test ./internal/router ./internal/middleware ./internal/authz -run "Test(Router|Capability)" -count=1 -timeout 60s`

---

### Task 2: Restrict configuration mutations to manage-config capability without breaking read access

**Files:** `internal/router/router.go`, `internal/router/router_test.go`, `internal/authz/capabilities.go`, `internal/authz/capabilities_test.go`

**Acceptance Criteria:**
- `internal/router/router.go` applies `RequireCapability(authz.CapabilityManageConfig)` to all config write endpoints
- `internal/authz/capabilities_test.go` still shows only `admin` can `manage_config`
- Router tests assert `403` for non-admin config writes and non-403 for admin writes on representative endpoints

**Action:** Use existing config route groups to differentiate reads from writes. Keep list/get routes readable for `admin`, `operator`, and `viewer`. Protect every mutating config route with `middleware.RequireCapability(authz.CapabilityManageConfig)`, including datasource/channel/route-rule/silence-rule/on-duty operations. Extend router and capability tests so the matrix proves `operator` and `viewer` cannot mutate config while `admin` can.

**Verification:** `go test ./internal/router ./internal/authz -run "Test(Router|Capability)" -count=1 -timeout 60s`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-07-01 | E | alert action routes | mitigate | Require `CapabilityProcessAlerts` for ack/quick-silence |
| T-07-02 | E | config mutation routes | mitigate | Require `CapabilityManageConfig` on every write endpoint |
| T-07-03 | R | inconsistent deny behavior | mitigate | Lock 401/403 responses with router tests |

## Established Patterns

- **Pattern 1:** Capability guards at router level before handler execution
- **Pattern 2:** Router tests assert exact HTTP status codes for auth outcomes
- **Pattern 3:** Matrix-based authorization vs role string comparisons

## Decisions

- Alert actions limited to roles with process-alert ability
- Config writes limited to `admin`
- Viewer remains read-only at backend permission surface

## Deviation Log

None
