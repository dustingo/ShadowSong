---
name: phase-09-03
description: Capability-only backend authz seam with retired RequireRole and strengthened security regressions
metadata:
  type: spec
  source_phase: 09-close-v1-1-milestone-tech-debt
  source_plan: "03"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 09 Plan 03: Backend Authz Seam Cleanup and Regression Strengthening

## Context & Goals

Close the real authorization implementation drift in the current tree: retire the unused `RequireRole` compatibility layer, unify to capability-first authz seam, and lock VER-04 critical security paths into automated regression.

Purpose: raw role check in user.go is no longer the problem, but the backend authz still retains an old role-based middleware seam and corresponding stale descriptions. Phase 9 should close this real drift point.
Output: capability-only middleware seam, updated security regression tests, codebase architecture note consistent with current implementation.

## Success Criteria

- Backend permission closure's public authz seam is only capability-first path; no unused role-based compatibility branch remains as a drift point.
- Disabled accounts, forced-reset limited mode, and critical allow/deny audit paths have explicit automated regression coverage after authz seam cleanup.
- Code and codebase documentation consistently describe current permission implementation as `JWTAuth + RequireCapability`, not mixed old `RequireRole` descriptions.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Cleaned authorize middleware | `internal/middleware/authorize.go` | Capability-only authorization middleware seam without stale role-based compatibility wrapper |
| Router security regression | `internal/router/router_test.go` | Route-level security regression coverage for forced-reset/disabled/capability-protected endpoints |
| User handler security regression | `internal/handlers/user_test.go` | Audit-log regression coverage for critical allowed and denied user-security actions |
| Updated architecture doc | `.planning/codebase/ARCHITECTURE.md` | Codebase map aligned with real capability-based backend authz path |

## Architecture

### Cleanup Scope

**Remove:**
- Unused `RequireRole(...)` compatibility wrapper from `internal/middleware/authorize.go`

**Keep:**
- `RequireCapability(...)` as single route-facing authorization entry point

### Regression Strengthening

**Router tests (VER-04 critical paths):**
- Disabled accounts rejected on representative protected routes after auth middleware and capability checks compose
- Forced-reset users limited to allowed self-service paths
- Capability-protected endpoints correctly allow/deny by capability, not role string

**Handler tests (VER-04 critical audit paths):**
- Denied self-disable or self-role-change emit expected audit log records
- Successful admin-managed security changes emit expected audit log records
- Successful self-service password change clearing forced-reset state emits expected audit log records

### Key Decisions

- No redesign of router policy or capability names
- Scope narrow to verification hardening around existing behavior
- Documentation updated to match actual tree, not desired state

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| router-facing authz API -> protected endpoints | Routes should expose one clear authorization seam so policy drift does not hide behind unused compatibility wrappers |
| auth middleware state -> security-sensitive handlers | Disabled-account denial, forced-reset limited mode, and audit logging must survive authz seam cleanup unchanged |

## Implementation Tasks

### Task 1: Retire the stale role-based compatibility middleware and document the capability-only backend seam

**Files:** `internal/middleware/authorize.go`, `internal/middleware/authorize_test.go`, `.planning/codebase/ARCHITECTURE.md`

**Acceptance Criteria:**
- Capability middleware still returns `401` without principal, `403` without capability, `200` for allowed principals
- No exported `RequireRole` compatibility shim remains in backend authz seam after cleanup
- `.planning/codebase/ARCHITECTURE.md` describes protected routes as using `RequireCapability`, not `RequireRole`

**Action:** Remove the unused `RequireRole(...)` compatibility wrapper from `internal/middleware/authorize.go`. Keep `RequireCapability(...)` as single route-facing entry point. Simplify `authorize_test.go` to assert surviving capability middleware contract only. Update ARCHITECTURE.md anywhere it still describes protected routes as using `RequireRole`.

**Verification:** `go test ./internal/middleware -count=1` and `rg -n '\bRequireRole\b' internal .planning/codebase/ARCHITECTURE.md` should find nothing

---

### Task 2: Strengthen backend regression coverage for VER-04-critical security paths at the capability boundary

**Files:** `internal/router/router_test.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `internal/router/router_test.go` contains explicit cases for disabled and forced-reset behavior on capability-protected endpoints
- `internal/handlers/user_test.go` contains explicit audit assertions for `user.disable`, `user.role_change`, `user.password_change`

**Action:** Expand existing backend regression suites so authz seam cleanup is tied to real VER-04 evidence. In `router_test.go`, add or tighten route-level cases proving disabled accounts and forced-reset users behave correctly on capability-protected endpoints. In `user_test.go`, make audit assertions explicit for key actions: denied self-disable/self-role-change, successful admin-managed security changes, successful self-service password change clearing forced-reset. Keep scope narrow to verification hardening; do not redesign handlers.

**Verification:** `go test ./internal/router ./internal/handlers ./internal/middleware -count=1 -run "Test(RouterCapabilityProtectedRoutes|AdminUpdateUser|UpdateOwnPassword)"` and `rg -n 'disabled|forced reset|user.disable|user.role_change|user.password_change' internal/router/router_test.go internal/handlers/user_test.go`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-09-06 | T | stale `RequireRole` compatibility path | mitigate | Remove unused role-based wrapper and leave capability middleware as single backend authz seam |
| T-09-07 | R | docs/code drift about how routes are actually protected | mitigate | Update authorize_test.go and ARCHITECTURE.md to describe only real `JWTAuth + RequireCapability` flow |
| T-09-08 | E | authz cleanup accidentally weakens disabled/forced-reset or audit-covered security behavior | mitigate | Expand router and handler regression coverage for disabled-user denial, forced-reset limits, and audit-log critical actions |

## Established Patterns

- **Pattern 1:** Single authorization seam (capability-only) prevents policy drift
- **Pattern 2:** Regression tests explicitly cover disabled/forced-reset security boundaries
- **Pattern 3:** Architecture docs match actual tree, not desired state

## Decisions

- Backend authz seam is capability-only, with no stale exported `RequireRole` compatibility layer left behind
- Route and handler regression coverage explicitly protects VER-04 critical paths during and after cleanup
- Codebase architecture notes no longer describe a role-based route-guard path that the current tree no longer uses

## Deviation Log

None
