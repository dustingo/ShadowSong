---
name: phase-11-01
description: Restore frontend lint baseline by fixing lint errors and reducing hook/any warning debt
metadata:
  type: spec
  source_phase: 11-restore-frontend-quality-baseline
  source_plan: "01"
  milestone: v1.2
  status: completed
  completed: 2026-04-21
---

# Phase 11 Plan 01: Frontend Lint Cleanup

## Context & Goals

Remove the current ESLint failure, eliminate the highest-value warning debt in active pages, and leave the React/Zustand codebase in a quieter, more maintainable state for later CI enforcement.

Purpose: Remove the current ESLint failure, eliminate the highest-value warning debt in active pages, and leave the React/Zustand codebase in a quieter, more maintainable state for later CI enforcement.
Output: Frontend source changes that make `pnpm lint` pass and reduce hook / unused-variable / `any` noise in the current phase scope.

## Success Criteria

- `pnpm lint` no longer fails on current frontend sources
- High-noise hook dependency warnings are fixed by aligning with real dependencies rather than suppressing lint
- Obvious unused variables and unbounded `any` usages in active pages are reduced to maintainable typed boundaries

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Removal of current lint errors and typed preview error handling | `frontend/src/pages/DataSources.tsx` | Removal of current lint errors and typed preview error handling |
| Permission shell cleanup without dead capability variables | `frontend/src/App.tsx` | Permission shell cleanup without dead capability variables |
| Narrowed shared frontend types for active flows | `frontend/src/types/index.ts` | Narrowed shared frontend types for active flows |

## Architecture

### Key Architectural Decisions

- Fix lint issues without changing user-facing product scope
- Resolve `react-hooks/exhaustive-deps` warnings by wiring effects to real Zustand actions, preferring minimal dependency-safe adjustments over lint suppression
- Narrow obvious `any` usages to `unknown`, table render argument types, or existing domain types

## Implementation Tasks

### Task 1: Remove current lint-blocking errors and warning debt in active frontend pages

**Files:** `frontend/src/App.tsx`, `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Channels.tsx`, `frontend/src/pages/DataSources.tsx`, `frontend/src/pages/Dashboard.tsx`, `frontend/src/pages/Login.tsx`, `frontend/src/pages/OnDuty.tsx`, `frontend/src/pages/RouteRules.tsx`, `frontend/src/pages/Silences.tsx`, `frontend/src/pages/Users.tsx`, `frontend/src/types/index.ts`

**Acceptance Criteria:**
- `pnpm lint` exits 0
- frontend/src/pages/DataSources.tsx no longer contains the current useless-escape failure
- frontend/src/App.tsx no longer keeps unused capability imports
- active pages no longer rely on empty dependency arrays where the effect uses fetched actions

**Action:** Fix the existing frontend lint findings in place without changing user-facing product scope. Remove dead imports and unused capability variables from the app shell and route/config pages. Resolve `react-hooks/exhaustive-deps` warnings by wiring effects to the real Zustand actions they use, preferring minimal dependency-safe adjustments over lint suppression. Narrow obvious `any` usages in active pages and shared types to `unknown`, table render argument types, or existing domain types where the code already knows the shape. In `frontend/src/pages/DataSources.tsx`, eliminate the current `no-useless-escape` lint errors and leave the template help text semantically unchanged.

**Verification:** `pnpm lint`

## Security Considerations

None

## Established Patterns

None

## Decisions

- Fix by wiring effects to real dependencies rather than suppressing lint
- Narrow `any` to `unknown` or existing domain types

## Deviation Log

None
