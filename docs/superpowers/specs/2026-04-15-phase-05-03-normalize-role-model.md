---
name: phase-05-03
description: Role-to-capability matrix and capability-based middleware adapters
metadata:
  type: spec
  source_phase: 05-normalize-role-model
  source_plan: "03"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 05 Plan 03: Capability Matrix and Authorization Middleware

## Context & Goals

Introduce the reusable backend authorization baseline for Phase 6 and Phase 7 by defining a role-to-capability matrix and wiring it into middleware-friendly adapters.

Purpose: satisfy AUTHZ-03 with reusable policy code now, without prematurely implementing account-disable or forced-password-reset behavior from later phases.
Output: capability matrix helpers, capability middleware, and tests for role-based allow/deny behavior.

## Success Criteria

- The backend can answer whether a principal may view alerts, process alerts, manage configuration, or manage users based on role.
- Route and handler code can depend on reusable capability helpers instead of ad hoc role string comparisons.
- Later phases can add account disable, forced password reset, and capability-based guards without redesigning the Phase 5 authorization baseline.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Capability matrix | `internal/authz/capabilities.go` | Capability matrix mapping `admin`, `operator`, and `viewer` to backend permissions |
| Matrix regression tests | `internal/authz/capabilities_test.go` | Matrix regression coverage for allow/deny behavior across all supported roles |
| Capability middleware | `internal/middleware/authorize.go` | Reusable capability-based Gin middleware adapters built on the principal abstraction |
| Baseline router adoption | `internal/router/router.go` | Baseline adoption of capability middleware where Phase 5 already has admin-only endpoints |
| Router regression tests | `internal/router/router_test.go` | Router-level regression coverage for capability-wired authz guards |

## Architecture

### Capability Matrix

| Role | view_alerts | process_alerts | view_config | manage_config | manage_users |
|------|-------------|----------------|-------------|---------------|--------------|
| admin | yes | yes | yes | yes | yes |
| operator | yes | yes | yes | no | no |
| viewer | yes | no | yes | no | no |

### Key Decisions

- Capabilities are typed strings: `view_alerts`, `process_alerts`, `view_config`, `manage_config`, `manage_users`
- Matrix implemented in code, not comments
- `RequireCapability(capability)` middleware adapter wraps the matrix check
- Extension seams kept for later Phase 6/7 work (typed capabilities, centralized)

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| authenticated principal -> capability check | Backend policy code decides whether a role may perform security-sensitive operations |
| router guard -> handler execution | Incorrect allow/deny behavior at middleware level can expose protected routes or break legitimate access |

## Implementation Tasks

### Task 1: Define and test the Phase 5 capability matrix

**Files:** `internal/authz/capabilities.go`, `internal/authz/capabilities_test.go`

**Acceptance Criteria:**
- The capability matrix encodes allow/deny behavior for all three supported roles.
- Unsupported roles evaluate to deny-by-default.
- Tests enumerate `admin`, `operator`, and `viewer` across the targeted capabilities from AUTHZ-03.

**Action:** Create `internal/authz/capabilities.go` with a small explicit capability vocabulary for the actions Phase 5 must normalize: viewing alerts/config data, processing alerts, managing config, and managing users. Implement the matrix in code. Add table-driven tests in `internal/authz/capabilities_test.go` that lock the full allow/deny matrix for all three supported roles.

**Verification:** `go test ./internal/authz -run "TestCapability" -count=1`

---

### Task 2: Add capability-based middleware adapters and baseline router adoption

**Files:** `internal/middleware/authorize.go`, `internal/router/router.go`, `internal/router/router_test.go`

**Acceptance Criteria:**
- `internal/middleware/authorize.go` exposes a capability-based guard reusable by later phases.
- `internal/router/router.go` contains at least one concrete capability-based route guard for current admin-only user operations.
- Tests fail if capability middleware stops distinguishing missing principal from insufficient permissions.

**Action:** Create `internal/middleware/authorize.go` with capability-based middleware adapters that consume the principal established in Plan 02 and the matrix from `internal/authz/capabilities.go`. Make `RequireRole(...)` a thin wrapper over the new baseline or start migrating already-admin-only user routes to `RequireCapability(authz.CapabilityManageUsers)`. Update `internal/router/router_test.go` to cover middleware registration and at least one allow/deny path for capability enforcement.

**Verification:** `go test ./internal/authz ./internal/middleware ./internal/router -run "TestCapability|TestRequireCapability|TestRouter" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-05-07 | E | `internal/authz/capabilities.go` | mitigate | Encode the allow/deny matrix explicitly in code and lock it with table-driven tests for all supported roles |
| T-05-08 | D | `internal/middleware/authorize.go` | mitigate | Return 401 for missing principal and 403 for denied capability, with direct middleware tests for both branches |
| T-05-09 | T | `internal/router/router.go` | mitigate | Adopt capability middleware on the already-admin-only route surface first so Phase 5 changes remain scoped while proving the baseline wiring |

## Established Patterns

- **Pattern 1:** Centralized typed capability constants prevent scattered string literals
- **Pattern 2:** Capability middleware returns 401 for missing principal, 403 for insufficient permissions
- **Pattern 3:** Matrix locked by table-driven tests for all supported roles

## Decisions

- Account disable, forced password reset, and audit logging not implemented in this plan (deferred to Phase 6/7)
- Capability vocabulary kept small and backend-oriented
- Only existing admin-only routes adopted in this plan; Phase 7 will handle config and alert mutation hardening

## Deviation Log

None
