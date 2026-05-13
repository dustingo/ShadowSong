---
name: phase-08-03
description: Frontend permission tests, Phase 8 verification report, and traceability sync
metadata:
  type: spec
  source_phase: 08-ship-permission-aware-ui-and-verification
  source_plan: "03"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 08 Plan 03: Frontend Permission Tests and Phase Verification

## Context & Goals

Complete Phase 8's final evidence mile: runnable frontend role-matrix tests, cross-phase verification document, and synchronized tracking documents.

Purpose: requirements explicitly demand frontend verification and documentation; without automated frontend tests and final verification artifact, Phase 8 is just UI changes, not auditable delivery.
Output: frontend permission tests, phase verification report, requirements/roadmap/state synchronization.

## Success Criteria

- Project has at least one set of repeatable frontend permission matrix verifications, not just manual screenshots.
- Phase 7 backend matrix and audit evidence are explicitly referenced in Phase 8 verification document, satisfying cross-phase verification requirements.
- Requirements tracking, roadmap, and state files reflect Phase 8 planning completion.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Shell/menu/route tests | `frontend/src/App.test.tsx` | Shell/menu/route visibility matrix tests by role |
| Alert permission tests | `frontend/src/pages/Alerts.test.tsx` | Alert action visibility tests |
| Config permission tests | `frontend/src/pages/DataSources.test.tsx` | Config read-only vs admin write-surface tests |
| Phase 8 verification | `.planning/phases/08-ship-permission-aware-ui-and-verification/08-VERIFICATION.md` | Phase-owned verification evidence linking backend and frontend coverage |

## Architecture

### Test Coverage Requirements

**App.test.tsx:**
- Shell/menu visibility for `admin`, `operator`, `viewer`, forced-reset users
- Route behavior under different role states

**Alerts.test.tsx:**
- `viewer` cannot see `确认` and `静默` buttons
- Processing roles CAN see action buttons

**DataSources.test.tsx:**
- Non-admin roles get read-only config UI
- `admin` retains write controls

### Cross-Phase Evidence

Phase 8 verification document explicitly references:
- Phase 7 backend enforcement evidence (PERM-01-04, AUDIT-01-03)
- Phase 7 verification document
- Requirement mapping for VER-01 through VER-04 and FEACL-01 through FEACL-03

### Key Decisions

- Tests deterministic by mocking store and router state, not depending on live backend calls
- Verification document records exact commands run
- Requirements/roadmap/state synced to reflect Phase 8 completion

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| implemented frontend gating -> regression coverage | Future UI edits could silently re-open forbidden buttons without automated tests |
| cross-phase evidence -> milestone closure | Backend and frontend verification must be tied together rather than living in separate implicit histories |

## Implementation Tasks

### Task 1: Add a minimal frontend test runner and role-matrix permission tests

**Files:** `frontend/package.json`, `frontend/vite.config.ts`, `frontend/src/App.test.tsx`, `frontend/src/pages/Alerts.test.tsx`, `frontend/src/pages/DataSources.test.tsx`

**Acceptance Criteria:**
- `frontend/package.json` contains a frontend test script
- `frontend/vite.config.ts` contains test configuration for Vitest
- All three test files exist and each assert concrete role-specific visibility outcomes
- Frontend test command exits 0

**Action:** Add minimal frontend test setup compatible with existing Vite app, using Vitest and React Testing Library. Write `App.test.tsx` to cover shell/menu visibility for all roles and forced-reset states. Write `Alerts.test.tsx` to prove `viewer` cannot see action buttons. Write `DataSources.test.tsx` to prove non-admin read-only and admin write access. Keep tests deterministic by mocking store and router state.

**Verification:** `cd frontend && pnpm test -- --run`

---

### Task 2: Write the final Phase 8 verification and sync planning traceability

**Files:** `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`, `.planning/phases/08-ship-permission-aware-ui-and-verification/08-VERIFICATION.md`

**Acceptance Criteria:**
- `08-VERIFICATION.md` exists and contains VER-01 through VER-04 and FEACL-01 through FEACL-03
- File records exact frontend and backend test commands
- `REQUIREMENTS.md`, `ROADMAP.md`, and `STATE.md` reflect Phase 8 progress consistently

**Action:** After implementation and tests pass, update traceability so PERM-*, AUDIT-*, FEACL-*, and VER-* reflect completed phases accurately. Update ROADMAP and STATE so Phase 8 moves from planned to executed. Write verification document with exact commands run, explicit evidence lines for admin/operator/viewer frontend behavior, forced-reset routing, backend audit evidence reused from Phase 7, and requirement mapping.

**Verification:** `rg -n "VER-01|VER-02|VER-03|VER-04|FEACL-01|FEACL-02|FEACL-03|pnpm test|go test" .planning/phases/08-ship-permission-aware-ui-and-verification/08-VERIFICATION.md`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-08-07 | R | missing frontend regression coverage | mitigate | Add explicit role-matrix tests for shell, alerts, and config pages |
| T-08-08 | T | docs drift from implementation | mitigate | Write Phase 8 verification from exact commands and sync traceability files |

## Established Patterns

- **Pattern 1:** Deterministic tests using mocked stores and router state
- **Pattern 2:** Verification document records exact commands, not summary claims
- **Pattern 3:** Cross-phase evidence explicitly linked (Phase 7 -> Phase 8)

## Decisions

- Frontend permission behavior is regression-locked
- Phase 8 explicitly ties frontend evidence to Phase 7 backend enforcement and audit evidence
- Planning artifacts stay consistent with executed milestone state

## Deviation Log

None
