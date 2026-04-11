---
phase: 05-normalize-role-model
plan: 02
subsystem: auth
tags: [jwt, gin, middleware, postgres, audit, testing]
requires:
  - phase: 05-normalize-role-model
    provides: canonical backend role constants, model-level role enforcement, and handler validation from 05-01
provides:
  - JWT claim validation that rejects unsupported persisted role strings before request context creation
  - reusable request principal extraction that preserves existing Gin auth context keys
  - read-only persisted-role audit CLI for rollout verification
  - bootstrap admin creation aligned to the canonical role constant
affects: [phase-06-secure-user-administration, phase-07-lock-down-protected-operations, rollout-authz-verification]
tech-stack:
  added: []
  patterns: [principal-first JWT middleware, fail-closed role-claim validation, read-only rollout audit command]
key-files:
  created: [internal/auth/jwt_test.go, internal/middleware/auth_test.go, internal/database/postgres_test.go, cmd/roleaudit/main.go, cmd/roleaudit/main_test.go]
  modified: [internal/auth/jwt.go, internal/middleware/auth.go, internal/handlers/user_test.go, internal/database/postgres.go]
key-decisions:
  - "Kept JWT claim keys and login/refresh response shapes unchanged while moving unsupported-role rejection into token validation."
  - "Added a lightweight middleware principal abstraction but continued populating user_id, username, and role Gin context keys for compatibility."
  - "Implemented role audit as a read-only grouped query command instead of reusing InitDB, so rollout checks do not migrate schema or create bootstrap users."
patterns-established:
  - "Validate supported roles before constructing any authenticated request principal."
  - "Operational audit tooling for live auth data should use dedicated read-only DB access rather than app bootstrap paths."
requirements-completed: [AUTHZ-02]
duration: 24min
completed: 2026-04-12
---

# Phase 05 Plan 02: Normalize Role Model Summary

**JWT principal hardening with compatible claim/session shape and a read-only persisted-role rollout audit command**

## Performance

- **Duration:** 24 min
- **Started:** 2026-04-11T16:13:12Z
- **Completed:** 2026-04-11T16:36:54Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- Hardened JWT validation and middleware principal extraction so unsupported role claims fail closed before authorization state reaches Gin context.
- Added regression coverage proving login and refresh still preserve the existing `token`/`user` contract and `user_id`/`username`/`role` claim keys.
- Added a dedicated `go run ./cmd/roleaudit` command that groups persisted `users.role` values, exits non-zero on unsupported buckets, and leaves rollout remediation outside this plan.

## Task Commits

Each task was committed atomically:

1. **Task 1: Validate JWT role claims and extract a reusable principal without breaking session shape** - `02568b4` (fix)
2. **Task 2: Audit persisted role values and align bootstrap assumptions to the canonical contract** - `938068f` (feat)

## Files Created/Modified
- `internal/auth/jwt.go` - rejects unsupported role claims during token validation so refresh and middleware both fail closed.
- `internal/middleware/auth.go` - adds a reusable `Principal` path while preserving `user_id`, `username`, and `role` context keys.
- `internal/auth/jwt_test.go` - locks claim-key compatibility and supported-role token validation.
- `internal/middleware/auth_test.go` - covers bearer parsing, principal extraction, invalid-role rejection, and getter compatibility.
- `internal/handlers/user_test.go` - proves login and refresh keep the existing response shape while invalid-role tokens are rejected.
- `internal/database/postgres.go` - aligns bootstrap admin creation with `authz.RoleAdmin`.
- `internal/database/postgres_test.go` - verifies bootstrap behavior and guards the canonical admin-role assignment.
- `cmd/roleaudit/main.go` - implements the read-only persisted-role audit CLI and deterministic exit behavior.
- `cmd/roleaudit/main_test.go` - covers supported-role success, unsupported-role failure, query failure, and grouped-query-only behavior.

## Decisions Made
- Kept compatibility at the contract edges: JWT claims still serialize as `user_id`, `username`, and `role`, and login/refresh response shapes remain unchanged.
- Put invalid-role rejection in `ValidateToken` so middleware and refresh share one fail-closed boundary instead of duplicating checks.
- Avoided `internal/database.InitDB` inside the audit CLI because that path migrates tables and can create bootstrap data, which would violate the rollout tool's read-only intent.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- The initial Task 1 RED run showed refresh still accepted unsupported role claims; this was resolved by validating supported roles during token validation rather than only in middleware.
- The initial Task 2 RED run showed bootstrap still used a raw `"admin"` string; this was resolved by switching to `authz.RoleAdmin` and locking it with a source-level regression assertion.

## Target Rollout Audit

Target-environment verification completed on 2026-04-12 with `go run ./cmd/roleaudit`.

Observed output:

```text
Persisted role counts:
- admin: 1
Audit passed: only supported roles found (admin, operator, viewer)
```

The command exited `0`, so the persisted-role compatibility gate for AUTHZ-02 is satisfied.

## User Setup Required

None - no external service configuration required inside the repo.

## Next Phase Readiness
- Phase 6 can build on the new `Principal` helper without breaking existing request-context consumers.
- Target-database persisted-role audit already passed, so Phase 5 no longer depends on an outstanding rollout compatibility check.

## Self-Check: PASSED
- Found `.planning/phases/05-normalize-role-model/05-02-SUMMARY.md`
- Found commit `02568b4`
- Found commit `938068f`

---
*Phase: 05-normalize-role-model*
*Completed: 2026-04-12*
