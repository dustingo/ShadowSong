---
phase: 05-normalize-role-model
plan: 03
subsystem: auth
tags: [authz, gin, middleware, router, testing]
requires:
  - phase: 05-normalize-role-model
    provides: canonical backend role constants, JWT principal validation, and compatible request principal extraction from 05-01 and 05-02
provides:
  - centralized backend role-to-capability matrix for admin, operator, and viewer
  - reusable Gin capability middleware with explicit 401 and 403 behavior
  - baseline router adoption of capability guards for admin-only user management routes
affects: [phase-06-secure-user-administration, phase-07-lock-down-protected-operations, rollout-authz-baseline]
tech-stack:
  added: []
  patterns: [centralized capability matrix, principal-driven capability middleware, router-level authz regression coverage]
key-files:
  created: [internal/authz/capabilities.go, internal/authz/capabilities_test.go, internal/middleware/authorize.go, internal/middleware/authorize_test.go]
  modified: [internal/middleware/auth.go, internal/router/router.go, internal/router/router_test.go]
key-decisions:
  - "Added a small backend-only capability vocabulary, including view_config, so Phase 5 can distinguish read access from config mutation before broader Phase 7 hardening."
  - "Applied capability enforcement only to the already-admin-only user management surface, leaving wider config and alert route lockdown to later dedicated plans."
  - "Kept RequireRole as a compatibility helper while moving new policy entry points to RequireCapability."
patterns-established:
  - "Express backend authorization through authz capabilities rather than repeating raw role comparisons in routes."
  - "Capability middleware should fail closed with 401 for missing principal and 403 for insufficient permissions."
requirements-completed: [AUTHZ-03]
duration: 18min
completed: 2026-04-12
---

# Phase 05 Plan 03: Normalize Role Model Summary

**Centralized role-to-capability authorization matrix with Gin middleware adapters and capability-guarded user-management routes**

## Performance

- **Duration:** 18 min
- **Started:** 2026-04-11T16:34:38Z
- **Completed:** 2026-04-11T16:52:38Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Added a typed backend capability matrix that encodes allow and deny behavior for `admin`, `operator`, and `viewer`, with unsupported roles denied by default.
- Added reusable capability middleware that consumes the Phase 5 principal abstraction and distinguishes missing principal from insufficient permissions.
- Migrated the existing admin-only user-management routes to capability guards and added router regression coverage for public, unauthorized, forbidden, and allowed paths.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define and test the Phase 5 capability matrix** - `d694b0c` (feat)
2. **Task 2: Add capability-based middleware adapters and baseline router adoption** - `7bfb310` (feat)

## Files Created/Modified
- `internal/authz/capabilities.go` - defines the typed capability vocabulary and role-to-capability matrix.
- `internal/authz/capabilities_test.go` - locks the full allow and deny matrix for supported and unsupported roles.
- `internal/middleware/authorize.go` - adds reusable capability middleware and a compatibility role helper.
- `internal/middleware/authorize_test.go` - verifies `401`, `403`, and successful capability authorization paths.
- `internal/middleware/auth.go` - retains principal extraction while delegating authorization responsibilities to the new middleware file.
- `internal/router/router.go` - applies `CapabilityManageUsers` to the existing admin-only user-management endpoints.
- `internal/router/router_test.go` - verifies public routes remain available and user-management routes enforce the capability baseline.

## Decisions Made
- Added `CapabilityViewConfig` now because the Phase 5 requirement explicitly includes read-only config visibility, and that split will matter in Phase 7 when mutation routes are hardened.
- Scoped router adoption to the already-admin-only `/api/v1/users` management surface to prove the middleware baseline without prematurely locking down broader operational routes.
- Kept the compatibility `RequireRole` helper for brownfield callers, but treated `RequireCapability` as the new extension seam for future phases.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 6 can build on the shared capability helpers when separating admin-managed user actions from self-service profile updates.
- Phase 7 can reuse the same middleware and capability vocabulary to harden alert actions and configuration write paths without redesigning route guards.

## Self-Check: PASSED
- Found `.planning/phases/05-normalize-role-model/05-03-SUMMARY.md`
- Found commit `d694b0c`
- Found commit `7bfb310`

---
*Phase: 05-normalize-role-model*
*Completed: 2026-04-12*
