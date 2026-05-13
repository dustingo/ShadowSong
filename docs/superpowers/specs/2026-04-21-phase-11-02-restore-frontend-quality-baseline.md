---
name: phase-11-02
description: Re-verify frontend test and build health after lint cleanup
metadata:
  type: spec
  source_phase: 11-restore-frontend-quality-baseline
  source_plan: "02"
  milestone: v1.2
  status: completed
  completed: 2026-04-21
---

# Phase 11 Plan 02: Frontend Test and Build Verification

## Context & Goals

Prove the Phase 11 cleanup did not only satisfy ESLint, but also preserved the current frontend test flow and production build path that Phase 12 will later gate in CI.

Purpose: Prove the Phase 11 cleanup did not only satisfy ESLint, but also preserved the current frontend test flow and production build path that Phase 12 will later gate in CI.
Output: Passing frontend lint, test, and build verification, plus only the minimal follow-up fixes needed to restore those paths if cleanup exposed latent issues.

## Success Criteria

- Frontend lint cleanup does not break the existing test flow
- Frontend production build still succeeds after Phase 11 code changes
- Any fixes needed for test or build regressions remain within the current frontend baseline scope

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Verification command truth for lint, test, and build | `frontend/package.json` | Verification command truth for lint, test, and build |
| Representative frontend test coverage that must remain green after cleanup | `frontend/src/pages/Dashboard.test.tsx` | Representative frontend test coverage that must remain green after cleanup |

## Architecture

### Key Architectural Decisions

- Only minimal follow-up fixes within frontend baseline repair scope
- No new test runner, product surface changes, or CI setup in this phase

## Implementation Tasks

### Task 1: Re-run frontend lint, test, and build flows and fix only Phase 11 regressions

**Files:** `frontend/src/pages/Dashboard.test.tsx`, `frontend/src/test/setup.ts`, `frontend/vite.config.ts`, `frontend/tsconfig.json`

**Acceptance Criteria:**
- `pnpm lint` exits 0
- `pnpm test -- --run` exits 0
- `pnpm build` exits 0
- any follow-up changes remain within frontend baseline repair scope

**Action:** After the lint cleanup is complete, run the existing frontend verification commands in sequence: `pnpm lint`, `pnpm test -- --run`, and `pnpm build`. If any test or build step fails because the lint cleanup exposed a real regression, apply only the minimal follow-up fix needed to restore the existing baseline, keeping changes inside the current frontend scope and avoiding new feature work. Do not add a new test runner, change the product surface, or broaden into CI setup in this phase.

**Verification:** `pnpm lint && pnpm test -- --run && pnpm build`

## Security Considerations

None

## Established Patterns

None

## Decisions

- Only minimal follow-up fixes within frontend baseline repair scope
- No new test runner, product surface changes, or CI setup

## Deviation Log

None
