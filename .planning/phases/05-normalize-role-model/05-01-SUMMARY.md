---
phase: 05-normalize-role-model
plan: 01
subsystem: auth
tags: [roles, authz, gin, gorm, testing]
requires:
  - phase: 04-enable-raw-event-passthrough-in-notification-templates
    provides: stable non-AI alerting baseline for brownfield auth changes
provides:
  - canonical backend role constants and helpers
  - user model role defaulting and invalid-role rejection
  - request-level create and update validation for unsupported roles
affects: [phase-06-secure-user-administration, phase-07-lock-down-protected-operations]
tech-stack:
  added: [gorm.io/driver/sqlite]
  patterns: [shared backend role contract, model-hook-backed role enforcement, request-level role validation]
key-files:
  created: [internal/authz/roles.go, internal/authz/roles_test.go, internal/models/user_test.go, internal/handlers/user_test.go]
  modified: [internal/models/user.go, internal/handlers/user.go, go.mod]
key-decisions:
  - "Kept persisted and API role strings as admin/operator/viewer and centralized them in internal/authz."
  - "Applied viewer as the default only for empty roles; unsupported values are rejected instead of remapped."
  - "Validated invalid roles in handlers before persistence while keeping model hooks as the final guardrail."
patterns-established:
  - "Import authz constants/helpers instead of reusing raw role strings in backend write paths."
  - "Use model hooks plus focused request tests to enforce security-sensitive user invariants."
requirements-completed: [AUTHZ-01]
duration: 10min
completed: 2026-04-12
---

# Phase 05 Plan 01: Normalize Role Model Summary

**Canonical admin/operator/viewer backend role contract with model-hook and handler-level invalid-role rejection**

## Performance

- **Duration:** 10 min
- **Started:** 2026-04-11T16:17:00Z
- **Completed:** 2026-04-11T16:26:41Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Added `internal/authz` as the canonical backend role contract with exported constants and helper APIs.
- Moved user role defaulting and validation to shared authz helpers plus model hooks so unsupported roles cannot be persisted.
- Added focused request tests proving create and update endpoints reject unsupported roles and keep the default role as `viewer`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define the canonical role contract for Phase 5 and later phases** - `32a48d6` (feat)
2. **Task 2: Enforce supported roles on model persistence and user write handlers** - `cd468f5` (fix)

## Files Created/Modified
- `internal/authz/roles.go` - exports canonical role constants and supported-role helpers.
- `internal/authz/roles_test.go` - locks the supported role set and defaulting behavior with table-driven tests.
- `internal/models/user.go` - defaults empty roles to `viewer` and rejects unsupported roles through shared authz helpers and hooks.
- `internal/models/user_test.go` - covers validation defaults and invalid-role rejection at the model layer.
- `internal/handlers/user.go` - rejects unsupported create/update role inputs before persistence and uses a dedicated create request DTO.
- `internal/handlers/user_test.go` - verifies create and update handler behavior against an isolated in-memory database.
- `go.mod` - adds the SQLite driver required for request-level handler tests.

## Decisions Made
- Preserved the existing `admin`, `operator`, and `viewer` strings in storage and API payloads to avoid compatibility churn.
- Used handler-level validation for clear `400` responses and model-level validation as a fail-closed persistence guard.
- Introduced a dedicated create-user request payload with `password` rather than binding directly to the `User` model.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed create-user password binding so request validation can reach role checks**
- **Found during:** Task 2 (Enforce supported roles on model persistence and user write handlers)
- **Issue:** `CreateUser` bound JSON directly into `models.User`, but `PasswordHash` is tagged `json:"-"`, so incoming passwords never populated and the handler returned `password is required` before role validation.
- **Fix:** Added a dedicated `createUserRequest` DTO with a `password` field and mapped it into `models.User` before hashing and persistence.
- **Files modified:** `internal/handlers/user.go`, `internal/handlers/user_test.go`
- **Verification:** `go test ./internal/models ./internal/handlers -run "TestUser|TestCreateUser|TestUpdateUser" -count=1 -timeout 60s`
- **Committed in:** `cd468f5`

**2. [Rule 3 - Blocking] Added SQLite driver for isolated handler request tests**
- **Found during:** Task 2 (Enforce supported roles on model persistence and user write handlers)
- **Issue:** Request-level tests required a concrete GORM database, and the module was missing a direct SQLite driver requirement.
- **Fix:** Added `gorm.io/driver/sqlite` to `go.mod` and used isolated in-memory databases per test.
- **Files modified:** `go.mod`, `internal/handlers/user_test.go`
- **Verification:** `go test ./internal/authz ./internal/models ./internal/handlers -run "TestRoles|TestUser|TestCreateUser|TestUpdateUser" -count=1 -timeout 60s`
- **Committed in:** `cd468f5`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** Both fixes were required to make the planned validation and request-level tests actually executable. No scope creep.

## Issues Encountered
- The first handler test run revealed an existing create-user binding bug, which was fixed inline as part of Task 2.
- The initial SQLite test DSN reused a shared in-memory database across tests and caused a unique-key collision; the fixture was corrected to use a per-test DSN.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Later phases can import `internal/authz` instead of duplicating raw role literals.
- User-management and protected-operation hardening now have a stable backend role contract and regression coverage to build on.

## Self-Check: PASSED
- Found `.planning/phases/05-normalize-role-model/05-01-SUMMARY.md`
- Found commit `32a48d6`
- Found commit `cd468f5`

---
*Phase: 05-normalize-role-model*
*Completed: 2026-04-12*
