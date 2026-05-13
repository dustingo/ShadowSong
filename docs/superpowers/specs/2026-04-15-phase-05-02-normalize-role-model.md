---
name: phase-05-02
description: JWT principal hardening and read-only persisted-role audit tool
metadata:
  type: spec
  source_phase: 05-normalize-role-model
  source_plan: "02"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 05 Plan 02: JWT Principal and Persisted Role Audit

## Context & Goals

Preserve login and session compatibility while hardening JWT principal extraction and giving rollout a concrete persisted-role audit path.

Purpose: satisfy AUTHZ-02 without changing claim keys, frontend localStorage behavior, or the existing role names.
Output: validated principal extraction, bootstrap compatibility alignment, and a read-only role-audit command.

## Success Criteria

- Existing users can continue logging in with the same role strings and JWT claim keys.
- Middleware rejects unsupported role claims instead of copying arbitrary strings into request context.
- There is one executable audit path to inspect current persisted `users.role` values before or during rollout.
- Phase completion is blocked until the persisted-role audit runs against the target database and finds no unsupported `users.role` values.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| JWT claims contract | `internal/auth/jwt.go` | Stable JWT claims contract with explicit role validation boundaries |
| Principal extraction | `internal/middleware/auth.go` | Principal extraction that preserves existing context keys while validating supported roles |
| Auth middleware tests | `internal/middleware/auth_test.go` | Regression coverage for bearer parsing, claim extraction, and invalid-role rejection |
| Handler regression tests | `internal/handlers/user_test.go` | Login and refresh regression coverage proving the existing handler contract remains compatible after JWT hardening |
| Role audit command | `cmd/roleaudit/main.go` | Read-only persisted-role audit command for rollout compatibility checks |
| Bootstrap admin alignment | `internal/database/postgres.go` | Bootstrap admin creation aligned with the canonical role constants |
| Bootstrap tests | `internal/database/postgres_test.go` | Regression coverage proving bootstrap admin creation uses the canonical role constant |
| Audit command tests | `cmd/roleaudit/main_test.go` | Command-level coverage for read-only audit output and supported vs unsupported exit-code behavior |

## Architecture

### Key Decisions

- JWT claim keys (`user_id`, `username`, `role`) remain unchanged for frontend compatibility
- Middleware validates role claim against canonical set before setting Gin context
- Bootstrap admin uses `authz.RoleAdmin` constant instead of raw string
- Audit command performs read-only grouped query and exits 0 for supported roles, non-zero for unsupported

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| JWT bearer token -> request principal | Untrusted token content becomes authorization state for protected routes |
| rollout audit CLI -> production-like database | Read-only operational tooling inspects live user role values and must not mutate data |

## Implementation Tasks

### Task 1: Validate JWT role claims and extract a reusable principal without breaking session shape

**Files:** `internal/auth/jwt.go`, `internal/middleware/auth.go`, `internal/middleware/auth_test.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `Claims` still serializes `user_id`, `username`, and `role`.
- Middleware has an explicit invalid-role rejection path before setting Gin auth context.
- `internal/handlers/user_test.go` proves login and refresh keep the existing response contract while invalid-role tokens are rejected.

**Action:** Keep the `Claims` JSON field names unchanged. Harden `internal/auth/jwt.go` and `internal/middleware/auth.go` so unsupported `role` claims cannot flow into request authorization. Add a reusable principal abstraction that captures `UserID`, `Username`, and `Role`, preserves existing Gin context keys, and becomes the hook for later Phase 6 account-status flags.

**Verification:** `go test ./internal/middleware ./internal/auth ./internal/handlers -run "TestJWT|TestJWTAuth|TestPrincipal|TestLogin|TestRefresh" -count=1`

---

### Task 2: Audit persisted role values and align bootstrap assumptions to the canonical contract

**Files:** `internal/database/postgres.go`, `internal/database/postgres_test.go`, `cmd/roleaudit/main.go`, `cmd/roleaudit/main_test.go`

**Acceptance Criteria:**
- `createDefaultAdminUser` uses the canonical admin role constant.
- `cmd/roleaudit/main_test.go` fails if supported-role datasets do not exit `0`.
- `cmd/roleaudit/main_test.go` fails if unsupported-role datasets do not exit non-zero and report offending role values.
- `cmd/roleaudit/main_test.go` fails if the audit command attempts any write behavior.

**Action:** Update `internal/database/postgres.go` so the default admin bootstrap path uses the canonical `authz.RoleAdmin` constant. Add a read-only audit utility at `cmd/roleaudit/main.go` that connects with the existing app database configuration pattern and prints grouped counts for `users.role`, explicitly surfacing any values outside `admin`, `operator`, `viewer`.

**Verification:** `go test ./internal/database ./cmd/roleaudit -run "TestCreateDefaultAdminUser|TestRoleAudit" -count=1`

---

### Task 3: Run the persisted-role audit against the target rollout database (Human Verification Gate)

**What Built:** Task 2 delivers a read-only audit command with deterministic exit behavior for supported vs unsupported persisted roles.

**How to Verify:**
1. Run `go run ./cmd/roleaudit` with the same database environment variables that the target deployment uses.
2. Confirm the command exits `0` and the output contains only supported buckets for `admin`, `operator`, and `viewer`.
3. If the command exits non-zero or prints any unsupported role values, stop phase completion, capture the offending role rows, remediate the data outside this phase, and rerun the audit until it exits `0`.
4. Only mark the plan complete after a clean rerun proves the persisted data is compatible with the narrowed role contract.

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-05-04 | S | `internal/middleware/auth.go` | mitigate | Validate claim role against the canonical set before populating Gin context and cover the rejection path in middleware tests |
| T-05-05 | E | JWT-authenticated routes | mitigate | Preserve claim keys but reject unsupported roles so stale or forged claims cannot gain undefined authorization behavior |
| T-05-06 | T | `cmd/roleaudit/main.go` | mitigate | Keep the audit command strictly read-only and make unsupported-role findings explicit with a non-zero exit code |
| T-05-10 | T | rollout database contents | mitigate | Block phase completion on a real audit run so unsupported persisted role rows cannot silently survive the contract tightening |

## Established Patterns

- **Pattern 1:** Principal abstraction preserves `user_id`, `username`, `role` context keys while enabling future extension
- **Pattern 2:** Read-only audit command surfaces unsupported values explicitly with non-zero exit code
- **Pattern 3:** Bootstrap code uses canonical constants rather than raw strings

## Decisions

- Phase completion gated on real database audit run returning exit 0
- No auto-remediation; unsupported live rows must be manually remediated per AUTHZ-02
- localStorage keys (`token`, `user`) and login response JSON shape unchanged

## Deviation Log

Human verification gate required before plan completion - audit must be run against target database.
