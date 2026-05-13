---
name: phase-06-02
description: Admin/self-service user route split with high-risk boundary protection
metadata:
  type: spec
  source_phase: 06-secure-user-administration
  source_plan: "02"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 06 Plan 02: Admin and Self-Service User API Split

## Context & Goals

Clearly separate "admin manages users" from "ordinary users maintain themselves" at the backend layer, and add high-risk boundary protection.

Purpose: satisfy USER-01 to USER-04, and connect USER-05/USER-06 account state control to specific management and self-service APIs.
Output: new admin/self-service user routes, clear DTOs, dangerous boundary protection, and complete regression tests.

## Success Criteria

- `admin` has independent user management endpoints: list users, create users, update other users, delete users, and manage account state.
- Regular users can only update their own profile and password; they cannot smuggle payload modifications to change roles or affect other users.
- Dangerous boundaries like self-disable, self-demotion, and stale session after security change all have explicit rejection and regression coverage.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Split user handlers | `internal/handlers/user.go` | Separated admin management API and self-service profile/password API |
| Router route groups | `internal/router/router.go` | Admin routes and self-service routes clearly grouped |
| Admin handler tests | `internal/handlers/user_test.go` | Admin and self-service flow regression tests |
| Router auth matrix tests | `internal/router/router_test.go` | Admin/self-service route authorization matrix |

## Architecture

### Route Split

Admin routes (require `CapabilityManageUsers`):
- `GET /api/v1/users` - list users with role and account-control state
- `POST /api/v1/users` - create user
- `PUT /api/v1/users/:id` - admin update user (role/profile/security flags)
- `DELETE /api/v1/users/:id` - delete user

Self-service routes (authenticated principal only):
- `GET /api/v1/users/me` - get current user
- `PUT /api/v1/users/me/profile` - update own name and email only
- `PUT /api/v1/users/me/password` - change own password

### Key Decisions

- Admin update uses dedicated DTOs with explicit binding of role/status/reset fields
- Self-service DTOs do not contain role or account-control fields
- Security-sensitive changes rotate per-user token invalidation cutoff
- Self-disable, self-demotion explicitly rejected with client-visible errors

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| authenticated principal->user-management handlers | Trusted identity may still attempt unauthorized peer or self-target mutations |
| admin payload->target user security state | Admin write paths can invalidate sessions or alter roles for another account |

## Implementation Tasks

### Task 1: Replace the mixed user write path with explicit admin-managed contracts

**Files:** `internal/handlers/user.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `internal/handlers/user.go` no longer relies on a single mixed `UpdateUser` payload for both admin and self-service flows
- Admin update DTO binds role/status/reset fields explicitly
- Tests cover admin allow, non-admin deny, self-disable reject, self-demotion reject, and admin-triggered token invalidation behavior

**Action:** Refactor `internal/handlers/user.go` so admin operations use dedicated DTOs. Keep `ListUsers`, `CreateUser`, and `DeleteUser` under `CapabilityManageUsers`. Introduce explicit admin update method that can modify `name`, `email`, `role`, `disabled_at`/enabled state, and `force_password_reset` for a target user. On admin security-sensitive changes, update per-user token invalidation cutoff. Block self-delete, add self-disable rejection, reject self-demotion from `admin` via admin-update path.

**Verification:** `go test ./internal/handlers -run "Test(Admin|CreateUser|DeleteUser|UpdateUser|Self)" -count=1 -timeout 60s`

---

### Task 2: Add self-service profile and password endpoints that cannot mutate role or peer accounts

**Files:** `internal/handlers/user.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `internal/handlers/user.go` exposes separate self-profile and self-password methods
- Self-service DTOs do not contain role or account-control fields
- Password success clears forced-reset state and refreshes token cutoff
- Tests prove a user cannot modify another user or elevate privileges through self-service endpoints

**Action:** Add dedicated self-service methods: keep `GetCurrentUser`, add `UpdateOwnProfile`, add `UpdateOwnPassword`. Restrict profile edits to safe fields (`name`, `email`) only. Move password changes into separate endpoint and request DTO. Ensure self-service code never binds or applies role/account-state fields and can operate while account is in forced-reset limited mode.

**Verification:** `go test ./internal/handlers -run "Test(SelfService|Password|Profile|ForcedReset)" -count=1 -timeout 60s`

---

### Task 3: Rewire router paths to the split contracts and lock the auth matrix with router tests

**Files:** `internal/router/router.go`, `internal/router/router_test.go`

**Acceptance Criteria:**
- `internal/router/router.go` contains distinct admin and self-service user routes
- No mixed write route remains for ordinary users to hit
- `internal/router/router_test.go` covers unauthenticated, operator/viewer, admin, disabled, and forced-reset route outcomes

**Action:** Update `internal/router/router.go` to express the split explicitly. Admin management routes remain under `/api/v1/users` with `RequireCapability(authz.CapabilityManageUsers)`. Self-service routes live under `/api/v1/users/me`, `/api/v1/users/me/profile`, and `/api/v1/users/me/password` and depend only on authenticated principals plus the 06-01 limited-mode rules. Remove the old mixed `PUT /users/:id` route. Expand router tests to cover admin allow/deny, self-service reachability, and stale-session/forced-reset behavior.

**Verification:** `go test ./internal/router ./internal/handlers ./internal/middleware -run "Test(Router|Admin|SelfService|Disabled|ForcedReset)" -count=1 -timeout 60s`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-06-05 | E | `internal/handlers/user.go` mixed update removal | mitigate | Split admin DTOs from self-service DTOs so non-admin payloads never bind role/security fields |
| T-06-06 | E | admin self-target mutations | mitigate | Explicitly reject self-disable and self-demotion in the admin update/delete handlers |
| T-06-07 | T | role/status changes | mitigate | When role/disable/reset/password actions succeed, rotate the per-user token invalidation cutoff from 06-01 |
| T-06-08 | R | route surface ambiguity | mitigate | Remove the old mixed write route and cover the new matrix in router tests with concrete status-code assertions |

## Established Patterns

- **Pattern 1:** Admin vs self-service DTO separation prevents field smuggling
- **Pattern 2:** Token invalidation cutoff rotation on security-sensitive changes
- **Pattern 3:** Explicit error messages for self-target mutations (`cannot disable yourself`, `cannot change your own role`)

## Decisions

- Admin and self-service write surfaces no longer ambiguous
- Non-admins cannot change other users, roles, or account-control fields
- Security-sensitive user changes invalidate stale sessions and are regression-covered

## Deviation Log

None
