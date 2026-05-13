---
name: phase-07-03
description: Backend role-matrix regression coverage and Phase 7 verification evidence
metadata:
  type: spec
  source_phase: 07-lock-down-protected-operations
  source_plan: "03"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 07 Plan 03: Role Matrix Regression Lock and Verification Report

## Context & Goals

Lock Phase 7 permission and audit results into a regression matrix, and produce verification evidence that Phase 8 can consume.

Purpose: prevent "this round added capability, next round new endpoint bypasses it", while leaving reviewable proof of the permission system.
Output: more complete backend role matrix tests, denial semantics verification, Phase 7 verification report.

## Success Criteria

- Backend has regression coverage for admin/operator/viewer allow/deny matrix on critical endpoints.
- Boundary between 401 and 403 is verifiable and consistent.
- Phase 7 audit and permission closure has explicit verification evidence.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Role matrix tests | `internal/router/router_test.go` | Protected-operation role matrix regression coverage |
| Phase 7 verification | `.planning/phases/07-lock-down-protected-operations/07-VERIFICATION.md` | Phase-level verification evidence |

## Architecture

### Coverage Requirements

At least one of each:
- Alert action endpoint (ack, quick-silence)
- Config write endpoint
- Config read endpoint
- User-management endpoint

For each:
- Unauthenticated requests return `401`
- Authenticated-but-disallowed roles return `403 {"error":"insufficient permissions"}`
- Allowed roles reach the current handler result

### Key Decisions

- Capability matrix tests aligned with router assertions so policy and route wiring cannot silently drift apart
- Verification docs must reflect actual commands and matrices, not vague summaries
- Verification artifact makes later Phase 8 work easier instead of rediscovering evidence

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| intended policy -> future code edits | Regression tests must prevent accidental re-opening of protected endpoints |
| executed checks -> milestone evidence | Verification docs must reflect actual commands and matrices, not vague summaries |

## Implementation Tasks

### Task 1: Expand backend role-matrix coverage for representative protected operations

**Files:** `internal/router/router_test.go`, `internal/middleware/authorize_test.go`, `internal/authz/capabilities_test.go`

**Acceptance Criteria:**
- `internal/router/router_test.go` contains representative `admin`/`operator`/`viewer` matrix cases across alert actions, config reads/writes, and user-management
- `internal/middleware/authorize_test.go` still asserts `403` for missing capability
- `internal/authz/capabilities_test.go` reflects intended matrix with `operator` able to process alerts but not manage config

**Action:** Extend the existing backend authorization tests so representative protected operations are covered for all three roles. Include at least one alert action endpoint, one config write endpoint, one config read endpoint, and one user-management endpoint. Assert exact outcomes: unauthenticated `401`, authenticated-but-disallowed `403 {"error":"insufficient permissions"}`, allowed roles reach current handler result. Keep capability matrix tests aligned with router assertions.

**Verification:** `go test ./internal/router ./internal/middleware ./internal/authz -count=1 -timeout 60s`

---

### Task 2: Write the Phase 7 verification evidence from executed checks and observed matrices

**Files:** `.planning/phases/07-lock-down-protected-operations/07-VERIFICATION.md`

**Acceptance Criteria:**
- `07-VERIFICATION.md` exists
- It names `PERM-01` through `PERM-04` and `AUDIT-01` through `AUDIT-03`
- It records exact `go test` command(s) run
- It explicitly states admin-only config writes, operator alert actions, viewer denial, and audit persistence evidence

**Action:** Create `.planning/phases/07-lock-down-protected-operations/07-VERIFICATION.md` after implementation. Summarize phase goal, exact automated commands executed, protected-operation matrix verified, and which Phase 7 requirements were satisfied. Include explicit evidence lines for admin-only config writes, operator alert processing, viewer read-only behavior, and backend-authored audit persistence.

**Verification:** `rg -n "PERM-01|PERM-02|PERM-03|PERM-04|AUDIT-01|AUDIT-02|AUDIT-03|go test" .planning/phases/07-lock-down-protected-operations/07-VERIFICATION.md`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-07-07 | R | missing regression coverage | mitigate | Lock representative matrices into router/capability tests |
| T-07-08 | T | permission drift between matrix and routes | mitigate | Cross-check capability tests and router tests together |

## Established Patterns

- **Pattern 1:** Router-level assertions for exact HTTP status codes (401 vs 403)
- **Pattern 2:** Verification artifact records exact commands run, not summary claims
- **Pattern 3:** Matrix tests aligned between capability tests and router tests

## Decisions

- Phase 7 ends with role-matrix regression coverage, not just implementation
- Verification artifact makes later Phase 8 work easier
- 401 vs 403 boundary is locked by tests

## Deviation Log

None
