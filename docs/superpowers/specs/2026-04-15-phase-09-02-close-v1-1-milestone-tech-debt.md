---
name: phase-09-02
description: Frontend test warning convergence for stable verification output
metadata:
  type: spec
  source_phase: 09-close-v1-1-milestone-tech-debt
  source_plan: "02"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 09 Plan 02: Frontend Warning Cleanup and Test Stabilization

## Context & Goals

Converge frontend verification noise so `pnpm test -- --run` continues to retain Phase 8 permission matrix coverage while removing non-blocking warnings that obscure readability.

Purpose: audit confirmed tests pass but output noise is high; Phase 9 needs verification signals to be stable and trustworthy, not real failures drowned in warnings.
Output: more stable frontend test harness, permission matrix test output with converged warnings.

## Success Criteria

- Frontend permission matrix tests continue to pass repeatably, with React Router future-flag and non-blocking `act(...)` noise significantly converged.
- Warning cleanup建立在真实测试 harness fixes之上, not simple global console muting.
- Phase 8 existing menu, alert action, and config page permission coverage does not degrade due to warning cleanup.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Router configuration aligned | `frontend/src/App.tsx` | Router configuration aligned with the test/runtime contract |
| Stable test environment | `frontend/src/test/setup.ts` | Stable jsdom test environment for Ant Design and router-driven permission tests |
| Shell permission regression tests | `frontend/src/App.test.tsx` | Shell/menu permission regression tests with reduced warning noise |
| Alert permission tests | `frontend/src/pages/Alerts.test.tsx` | Alert permission test coverage preserved after harness cleanup |
| Config permission tests | `frontend/src/pages/DataSources.test.tsx` | Config read-only/admin visibility coverage preserved after harness cleanup |

## Architecture

### Warning Targets

1. **React Router future-flag warning:** `React Router will begin wrapping state updates` - removed by using supported `future` configuration on rendered router
2. **Non-blocking `act(...)` warnings:** from permission-test path with known Ant Design Menu/layout async behavior - stabilized through deterministic setup or awaited assertions

### Constraints

- Do NOT fix by globally muting `console.warn`
- Only narrow filtering if warning is traced to unavoidable third-party internals after supported harness fixes, with exact-known benign message documented inline
- Keep existing role-matrix assertions intact

### Key Decisions

- Router-related warning noise removed at source (supported config), not suppressed
- Test harness updates make DOM settling deterministic
- Permission assertions remain unchanged

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| SPA runtime router config -> test runner | Test output should reflect the same router contract the app uses in production |
| jsdom harness -> Ant Design permission UI | Incomplete DOM shims or async handling can create noisy false positives that obscure real regressions |

## Implementation Tasks

### Task 1: Eliminate React Router future-flag warning at the app/test boundary

**Files:** `frontend/src/App.tsx`, `frontend/src/App.test.tsx`

**Acceptance Criteria:**
- Existing route and forbidden-state assertions still pass for viewer/admin/forced-reset users
- `pnpm test -- --run` output no longer includes React Router future-flag warnings

**Action:** Adjust app router/test boundary so React Router 6 future-flag warnings are removed at the source using supported `future` configuration on the router actually rendered. Keep test expectations aligned with production route contract. Do not fix by muting `console.warn` globally. Keep role-matrix assertions intact.

**Verification:** `cd frontend; $output = (pnpm test -- --run 2>&1 | Out-String); if ($output -match 'future flag' -or $output -match 'React Router will begin wrapping state updates') { exit 1 }`

---

### Task 2: Stabilize Ant Design permission tests so benign act warnings stop obscuring results

**Files:** `frontend/src/test/setup.ts`, `frontend/src/App.test.tsx`, `frontend/src/pages/Alerts.test.tsx`, `frontend/src/pages/DataSources.test.tsx`

**Acceptance Criteria:**
- App, Alerts, and DataSources permission tests still cover same visible/hidden outcomes after harness refactor
- `pnpm test -- --run` output no longer includes non-blocking `act(...)` warnings from known permission-test path

**Action:** Refine jsdom/test harness to match how Ant Design `Menu` and layout effects settle during existing permission tests. Prefer deterministic setup or awaited assertions over blanket console suppression. Update setup and three Phase 8 permission tests so render timing is explicit where needed, tests still prove viewer/operator/admin behavior, suite remains fast and repeatable. Only narrowly filter exact-known benign third-party messages if root-cause fixes prove impossible.

**Verification:** `cd frontend; $output = (pnpm test -- --run 2>&1 | Out-String); if ($output -match 'not wrapped in act') { exit 1 }` and `cd frontend && pnpm build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-09-03 | R | frontend warning-heavy verification output | mitigate | Remove warning sources through supported router config and deterministic test harness updates |
| T-09-04 | T | accidental coverage regression while cleaning warnings | mitigate | Keep existing App/Alerts/DataSources permission assertions and require both test and build commands to pass |
| T-09-05 | D | blanket console suppression hiding real failures | mitigate | Forbid global warning silencing; only allow narrowly scoped filtering if root-cause fixes prove impossible and message is exact-known benign third-party output |

## Established Patterns

- **Pattern 1:** Warning removal at source, not suppression
- **Pattern 2:** Deterministic test harness with explicit render timing
- **Pattern 3:** Coverage preserved through harness fixes, not overridden assertions

## Decisions

- Frontend role-matrix tests still pass
- Known React Router and `act(...)` warning noise removed or reduced
- No broad console muting introduced

## Deviation Log

None
