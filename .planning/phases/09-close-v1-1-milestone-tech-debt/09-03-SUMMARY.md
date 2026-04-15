---
phase: 09-close-v1-1-milestone-tech-debt
plan: 03
subsystem: auth
tags: [jwt, gin, authz, capability, audit, testing]
requires:
  - phase: 05-ship-enterprise-role-baseline
    provides: capability matrix and JWT principal baseline
  - phase: 06-ship-admin-user-boundaries-and-self-service-security
    provides: disabled-account and forced-reset enforcement paths
  - phase: 07-close-backend-permission-gaps-and-audit-actions
    provides: capability-protected routes and persistent audit logs
provides:
  - capability-only backend authorization middleware seam
  - VER-04 regression coverage for disabled-account and forced-reset route behavior
  - explicit audit-log assertions for critical user-security actions
affects: [internal/middleware, internal/router, internal/handlers, verification, architecture-docs]
tech-stack:
  added: []
  patterns: [JWTAuth plus RequireCapability, route-level authz regression tests, explicit audit assertions]
key-files:
  created: []
  modified: [internal/middleware/authorize.go, internal/middleware/authorize_test.go, internal/router/router_test.go, internal/handlers/user_test.go, .planning/codebase/ARCHITECTURE.md]
key-decisions:
  - "Removed the unused RequireRole compatibility wrapper instead of preserving a dead authz seam."
  - "Anchored VER-04 to representative capability-protected routes and explicit audit-log contracts."
patterns-established:
  - "Protected backend routes are documented and tested as JWTAuth plus RequireCapability only."
  - "Security-sensitive handler regressions assert audit action, result, actor, target, and detail fields."
requirements-completed: [VER-04]
duration: 5min
completed: 2026-04-15
---

# Phase 9 Plan 3: Authz Seam Cleanup Summary

**Capability-only backend authz middleware with VER-04 regression coverage for disabled, forced-reset, and audited user-security paths**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-15T16:11:25+08:00
- **Completed:** 2026-04-15T16:15:56.6840793+08:00
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Removed the stale exported `RequireRole` compatibility wrapper so route-facing backend authz now exposes only `RequireCapability`.
- Updated the codebase architecture note to describe the real `JWTAuth + RequireCapability` seam instead of the retired role-based path.
- Strengthened VER-04 regressions with disabled-account and forced-reset coverage on capability-protected routes plus explicit audit assertions for denied and allowed user-security actions.

## Task Commits

Each task was committed atomically:

1. **Task 1: Retire the stale role-based compatibility middleware and document the capability-only backend seam** - `1096280` (fix)
2. **Task 2: Strengthen backend regression coverage for VER-04-critical security paths at the capability boundary** - `8ad854b` (test)

## Files Created/Modified

- `internal/middleware/authorize.go` - Removed the dead `RequireRole` shim and documented `RequireCapability` as the only route-facing authz seam.
- `internal/middleware/authorize_test.go` - Kept capability middleware coverage focused on unauthorized, forbidden, and allowed outcomes.
- `internal/router/router_test.go` - Added disabled-user and forced-reset cases against representative capability-protected routes.
- `internal/handlers/user_test.go` - Made audit assertions explicit for denied self-protection checks, allowed admin-managed security changes, and self-service password updates.
- `.planning/codebase/ARCHITECTURE.md` - Aligned the authenticated CRUD flow description with the actual backend middleware path.

## Decisions Made

- Removed the compatibility wrapper instead of keeping a dormant API surface, because the router already uses capability guards exclusively.
- Verified forced-reset and disabled-account behavior at protected route boundaries rather than only on generic authenticated endpoints.
- Treated audit-log fields as part of the regression contract for VER-04-critical user-security flows.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `rg` verification needed a longer timeout than the default shell invocation; rerunning with a wider timeout confirmed there are no `RequireRole` references left in `internal/` or `.planning/codebase/ARCHITECTURE.md`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Backend authz is documented and tested as capability-first only, so future permission changes have one clear seam to modify.
- VER-04 now has concrete backend evidence for disabled-account denial, forced-reset limited mode, and critical audit coverage.

## Self-Check: PASSED

- Found `.planning/phases/09-close-v1-1-milestone-tech-debt/09-03-SUMMARY.md`
- Found commit `1096280`
- Found commit `8ad854b`

---
*Phase: 09-close-v1-1-milestone-tech-debt*
*Completed: 2026-04-15*
