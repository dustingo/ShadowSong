---
name: phase-06-01
description: Account disable, force-password-reset, and per-user token invalidation backend
metadata:
  type: spec
  source_phase: 06-secure-user-administration
  source_plan: "01"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 06 Plan 01: Account Control State and JWT Re-validation

## Context & Goals

Without changing the current JWT contract, establish the backend foundation for account disable, forced password reset, and immediate session invalidation for Phase 6.

Purpose: satisfy USER-05, USER-06, and provide a unified account state source of truth for subsequent admin operations and frontend restricted mode.
Output: persistent user security state fields, request-time account state re-validation, login/refresh fail-closed regression tests.

## Success Criteria

- Disabled accounts cannot log in and cannot continue using old JWTs to access protected endpoints.
- Accounts marked for forced password reset receive sessions that can only complete password change or logout, not continue normal system use.
- Existing JWT claim keys, login response shape, and principal context keys remain compatible.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Account control fields | `internal/models/user.go` | Account disable, force-password-reset, and token invalidation baseline fields and helper methods |
| Account model tests | `internal/models/user_test.go` | disabled, forced-reset, stale token regression coverage |
| JWT validation middleware | `internal/middleware/auth.go` | JWT validation with database-backed account state re-check |
| Auth middleware tests | `internal/middleware/auth_test.go` | disabled, forced-reset, stale token regression coverage |

## Architecture

### New User Model Fields

- `DisabledAt *time.Time` - nullable disabled timestamp for account disable
- `ForcePasswordReset bool` - forced password reset flag
- `TokenInvalidBefore *time.Time` - per-user token freshness cutoff for invalidating already-issued JWTs

### Key Decisions

- JWT claim keys unchanged (`user_id`, `username`, `role`)
- Login still returns `{token, user, expire_at}`
- Middleware performs database lookup on every protected request to re-check account state
- Forced-reset users limited to narrow allowlist path (password reset, logout, self-context)

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| client->auth handlers | Untrusted credentials and bearer tokens enter login/refresh flows |
| JWT claims->database user state | Signed token data must be reconciled against current persisted account state |

## Implementation Tasks

### Task 1: Add persistent account-control fields and model helpers

**Files:** `internal/models/user.go`, `internal/models/user_test.go`

**Acceptance Criteria:**
- `internal/models/user.go` contains persistent fields for disabled and force-reset state plus a token invalidation cutoff
- `internal/models/user_test.go` has table-driven coverage for active vs disabled vs forced-reset/stale-token helper behavior
- Existing role validation tests still pass

**Action:** Extend `models.User` with nullable disabled timestamp, boolean forced-reset flag, and per-user token freshness cutoff. Keep existing role validation and password hashing intact. Add model-level helpers `IsDisabled`, freshness comparison, and password/security-state update helpers. Use nullable/safe defaults so GORM `AutoMigrate` remains rollout-safe.

**Verification:** `go test ./internal/models -run "TestUser" -count=1 -timeout 60s`

---

### Task 2: Re-check current user state during login, refresh, and protected-request auth

**Files:** `internal/auth/jwt.go`, `internal/auth/jwt_test.go`, `internal/middleware/auth.go`, `internal/middleware/auth_test.go`, `internal/handlers/user.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `internal/middleware/auth.go` performs a database lookup before `SetPrincipal`
- Protected requests with a pre-cutoff token return `401`
- `internal/handlers/user.go` rejects disabled login and refresh while keeping the existing response envelope
- Tests explicitly cover disabled accounts, stale sessions, and forced-reset limited mode

**Action:** Keep Phase 5 contracts stable. Update `UserHandler.Login` and `UserHandler.RefreshToken` to load current `User` row and fail closed for disabled users. For forced-reset users, preserve successful login but expose reset-required state on returned user payload. Refactor `middleware.JWTAuth` to load current user from PostgreSQL on every protected request, reject disabled users, reject tokens older than stored invalidation cutoff, and block forced-reset users from all protected routes except narrow Phase 6 allowlist path.

**Verification:** `go test ./internal/auth ./internal/middleware ./internal/handlers -run "Test(Login|Refresh|JWT|Auth)" -count=1 -timeout 60s`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-06-01 | E | `internal/middleware/auth.go` | mitigate | Load current `models.User` on every protected request and reject disabled or stale-token sessions before principal creation |
| T-06-02 | E | `internal/handlers/user.go` login/refresh | mitigate | Fail closed for disabled accounts and refuse refresh when account state no longer permits normal use |
| T-06-03 | T | `internal/auth/jwt.go` + per-user cutoff | mitigate | Compare JWT `IssuedAt` against persisted invalidation timestamp so old tokens cannot outlive password/security actions |
| T-06-04 | D | forced-reset limited mode | mitigate | Allow only explicit self-password/self-context/logout paths so users can recover without keeping wider access |

## Established Patterns

- **Pattern 1:** Database-backed account state re-validation on every protected request
- **Pattern 2:** Nullable fields with safe zero/default semantics for rollout safety
- **Pattern 3:** Fail-closed for disabled accounts while preserving login contract for forced-reset

## Decisions

- JWT claim keys and request principal compatibility from Phase 5 remain intact
- Forced-reset limited mode allows password reset, logout, and self-context only
- Token invalidation uses `IssuedAt` comparison against persisted cutoff timestamp

## Deviation Log

None
