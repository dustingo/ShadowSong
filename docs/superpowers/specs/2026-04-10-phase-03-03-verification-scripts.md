---
name: phase-03-03
description: Establish backend/frontend no-AI verification scripts and produce verification evidence
metadata:
  type: spec
  source_phase: 03-align-docs-and-verification
  source_plan: "03"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 03 Plan 03: Verification Scripts and Evidence

## Context & Goals

Plan 03-02 refreshed codebase maps. This plan establishes and executes explicit non-AI verification paths for Phase 3 closure, producing evidence that both backend and frontend work in non-AI state.

**Goal:** Complete VER-01 and VER-02 — produce repeatable verification assets proving non-AI version works.

## Success Criteria

- Backend has a repeatable non-AI verification path executed at Phase 3 end
- Frontend has an explicit non-AI verification path proving build/static checks pass without new test framework
- Phase 3 has final verification evidence proving both backend and frontend work in non-AI state

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| No-AI backend closed-loop verification entry | `scripts/verify_backend_no_ai.ps1` | Real service startup + API verification |
| No-AI frontend build + residual scan verification entry | `scripts/verify_frontend_no_ai.ps1` | `pnpm build` + AI grep scan |
| Phase 3 final verification evidence | `.planning/phases/03-align-docs-and-verification/03-VERIFICATION.md` | Command output evidence |

## Architecture

### Verification Scripts

**scripts/verify_backend_no_ai.ps1:**
- Updated to use "non-AI alert system" semantics (not AI-branded DB names)
- Validates no AI API/closed-loop endpoints

**scripts/verify_frontend_no_ai.ps1 (new):**
- Frontend deps prereq check
- `pnpm build`
- AI residual grep scan on `frontend/` and `frontend/index.html`
- Returns non-zero exit code on failure

### Execution Order

1. `scripts/verify_backend_no_ai.ps1`
2. `scripts/verify_frontend_no_ai.ps1`
3. Both must pass for Phase 3 closure

### Verification Report

`03-VERIFICATION.md` records:
- Backend command, result, key evidence summary
- Frontend command, result, key evidence summary
- Requirement coverage
- Acceptance of non-blocking residual risks

## Implementation Tasks

### Task 1: Solidify Backend and Frontend Non-AI Verification Entry Points

**Files:** `scripts/verify_backend_no_ai.ps1`, `scripts/verify_frontend_no_ai.ps1`

**Acceptance Criteria:**
- Two runnable verification scripts exist
- Backend script continues validating no-AI API/closed-loop
- Frontend script successfully runs `pnpm build` and AI residual scan
- New script fails with non-zero exit code on failure

**Action:** Keep existing backend PowerShell closed-loop script main path, but update AI-branded DB name descriptions and AI-only config instructions to "non-AI alert system" semantics. Create new `scripts/verify_frontend_no_ai.ps1` executing frontend deps prereq check, `pnpm build`, and AI residual grep scan on `frontend/` and `frontend/index.html`, returning non-zero exit on failure. Do not introduce new frontend test framework — verification uses existing build gate and residual scan.

**Verification:** `powershell -ExecutionPolicy Bypass -File scripts/verify_frontend_no_ai.ps1` returns 0

---

### Task 2: Execute Verification and Produce Phase 3 Evidence Report

**Files:** `.planning/phases/03-align-docs-and-verification/03-VERIFICATION.md`

**Acceptance Criteria:**
- `03-VERIFICATION.md` exists and records at least one passed non-AI verification path for both backend and frontend, with execution time and result summary
- Both scripts execute successfully in series

**Action:** Execute backend and frontend non-AI verification commands. Record commands, results, key evidence, and requirement coverage in `03-VERIFICATION.md`. Report must clearly record backend command, frontend command, pass/fail, key output summary, and still-accepted non-blocking residual risks. If command fails, fix script first or document blocking cause — do not fabricate passing conclusions.

**Verification:** Both scripts pass in series; `03-VERIFICATION.md` contains actual execution evidence

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-03-07 | T | verify_frontend_no_ai.ps1 | mitigate | Script must execute real `pnpm build` and AI residual scan, return non-zero on failure — no placeholder verification |
| T-03-08 | R | 03-VERIFICATION.md | mitigate | Report only records actually executed commands and result summaries — no unexecuted passing conclusions |
| T-03-09 | D | verify_backend_no_ai.ps1 | accept | Backend script depends on local Docker/DB/Redis; if env unavailable can block execution, but must be explicitly recorded in report, not degraded to fake verification |

## Decisions

- Backend script updated to non-AI semantics, preserving existing verification logic
- Frontend script is new — no new test framework, uses existing build gate and grep scan

## Deviation Log

None — plan executed as written.