---
name: phase-05-01
description: Canonical role constants and backend enforcement for admin/operator/viewer
metadata:
  type: spec
  source_phase: 05-normalize-role-model
  source_plan: "01"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 05 Plan 01: Normalize Role Model

## Context & Goals

Create one canonical backend role vocabulary and enforce it on user writes without renaming any existing role values.

Purpose: satisfy AUTHZ-01 while preserving the current login/session contract and giving later phases a single source of truth for role semantics.
Output: exported role helpers, model-level enforcement, and focused tests for invalid-role rejection.

## Success Criteria

- The backend accepts only `admin`, `operator`, or `viewer` when users are created or updated.
- Unsupported role strings cannot be persisted through model hooks or user-management handlers.
- Existing role names remain unchanged in API payloads and stored records.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Role constants and helpers | `internal/authz/roles.go` | Canonical role constants and supported-role helpers for backend reuse |
| Role unit coverage | `internal/authz/roles_test.go` | Direct unit coverage for the canonical role constants and helper behavior |
| User model enforcement | `internal/models/user.go` | User model hooks and validation that enforce the supported role set |
| User model tests | `internal/models/user_test.go` | Regression coverage for default role assignment and invalid-role rejection |
| Handler request tests | `internal/handlers/user_test.go` | Request-level coverage for create/update role validation behavior |

## Architecture

### Key Decisions

- Role validation centralized in `internal/authz/roles.go` as single source of truth
- Model hooks enforce supported roles before persistence
- Handler-level validation surfaces explicit 400 errors for invalid roles
- Supported roles: `admin`, `operator`, `viewer` (exact strings, no aliasing)

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| user-management request -> persisted `users.role` | Client-controlled JSON can attempt to store unsupported or privilege-escalating role values |
| model hook -> database | Invalid role strings can become durable state if validation is not centralized |

## Implementation Tasks

### Task 1: Define the canonical role contract for Phase 5 and later phases

**Files:** `internal/authz/roles.go`, `internal/authz/roles_test.go`

**Acceptance Criteria:**
- `internal/authz/roles.go` exports constants for `admin`, `operator`, and `viewer`.
- `internal/authz/roles_test.go` fails if supported roles or defaulting behavior drift from the canonical contract.
- No helper in this file maps supported roles to different stored values.

**Action:** Create `internal/authz/roles.go` as the canonical backend role contract per AUTHZ-01. Export string constants `RoleAdmin`, `RoleOperator`, and `RoleViewer`; export helpers such as `SupportedRoles()`, `IsSupportedRole(role string)`, and a normalization/default helper that only fills empty roles with `viewer` and does not alias or rename any other value.

**Verification:** `go test ./internal/authz -run TestRoles -count=1`

---

### Task 2: Enforce supported roles on model persistence and user write handlers

**Files:** `internal/models/user.go`, `internal/models/user_test.go`, `internal/handlers/user.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `go test` fails if an unsupported role can be saved through `User` model hooks.
- `go test` fails if `CreateUser` or `UpdateUser` accepts a role outside `admin|operator|viewer`.
- `rg "invalid role" internal/models/user.go internal/handlers/user.go` shows explicit rejection paths instead of silent fallback.

**Action:** Refactor `internal/models/user.go` to depend on `internal/authz/roles.go`. Add save-hook coverage that applies the default role and rejects unsupported roles on both create and update paths. Update `internal/handlers/user.go` so `CreateUser` and `UpdateUser` surface clear 400-level errors for invalid roles before hitting persistence.

**Verification:** `go test ./internal/models ./internal/handlers -run "TestUser|TestCreateUser|TestUpdateUser" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-05-01 | E | `internal/handlers/user.go` | mitigate | Reject unsupported role strings with explicit 400 responses before persistence and cover create/update cases with request tests |
| T-05-02 | T | `internal/models/user.go` | mitigate | Centralize role validation and defaulting in model hooks backed by tests so overlooked write paths cannot persist unsupported roles |
| T-05-03 | R | user write regression coverage | accept | This plan does not add audit logging yet; Phase 7 will handle durable audit trails after the role contract is stable |

## Established Patterns

- **Pattern 1:** Single source of truth for role constants in `internal/authz/roles.go`
- **Pattern 2:** Model hooks validate before persistence to prevent invalid state
- **Pattern 3:** Handler-level validation surfaces explicit errors without changing API contract

## Decisions

- Accepted strings remain exactly `admin`, `operator`, `viewer` per D-locked constraint
- Default role is `viewer` for empty input
- No custom roles, database migrations, or frontend-facing enum churn introduced
- Login response shape, JWT claim keys, and frontend storage keys unchanged

## Deviation Log

None
