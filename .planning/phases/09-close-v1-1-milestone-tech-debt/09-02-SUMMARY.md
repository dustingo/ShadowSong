---
phase: 09-close-v1-1-milestone-tech-debt
plan: 02
subsystem: testing
tags: [react, react-router, vitest, antd, jsdom]
requires:
  - phase: 08-ship-permission-aware-ui-and-verification
    provides: Phase 8 permission-aware App, Alerts, and DataSources coverage reused by this cleanup
provides:
  - Router future flags aligned between runtime and tests
  - Stable jsdom timing for Ant Design permission tests
  - Cleaner permission test output without React Router, act, or disconnected form warnings
affects: [frontend, verification, permission-tests]
tech-stack:
  added: []
  patterns:
    - Await Ant Design layout work explicitly in permission tests
    - Keep router future configuration on the actual BrowserRouter contract
key-files:
  created: [.planning/phases/09-close-v1-1-milestone-tech-debt/09-02-SUMMARY.md]
  modified:
    - frontend/src/App.tsx
    - frontend/src/test/setup.ts
    - frontend/src/App.test.tsx
    - frontend/src/pages/Alerts.test.tsx
    - frontend/src/pages/DataSources.test.tsx
    - frontend/src/pages/DataSources.tsx
key-decisions:
  - Keep warning cleanup rooted in router/test-harness fixes instead of global console silencing
  - Apply a minimal runtime fix in DataSources modal rendering when test cleanup exposed a real Ant Design form warning
patterns-established:
  - Permission tests should render, then await a short Ant Design settle step before asserting menu or table controls
  - Test output cleanup should come from supported config or real component fixes, not blanket log suppression
requirements-completed: [VER-02]
duration: 19min
completed: 2026-04-15
---

# Phase 09 Plan 02: Warning Cleanup Summary

**React Router future flags, Ant Design settle timing, and DataSources modal warnings cleaned up so Phase 8 permission tests stay readable and repeatable**

## Performance

- **Duration:** 19 min
- **Started:** 2026-04-15T08:07:00Z
- **Completed:** 2026-04-15T08:26:00Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Removed React Router v7 future-flag warnings by configuring the real `BrowserRouter` used by `App`
- Made App, Alerts, and DataSources permission tests wait for Ant Design layout work deterministically inside jsdom
- Eliminated disconnected `useForm` and deprecated `addonAfter` warning noise from the DataSources permission path without broad console muting

## Task Commits

Each task was committed atomically:

1. **Task 1: Eliminate React Router future-flag warning at the app/test boundary** - `b3dbe3a` (test), `a0cae6f` (feat)
2. **Task 2: Stabilize Ant Design permission tests so benign act warnings stop obscuring results** - `9e2796b` (fix)

_Note: Task 1 followed a TDD split with separate test and implementation commits._

## Files Created/Modified

- `frontend/src/App.tsx` - opts the production/test router into the supported React Router future flags
- `frontend/src/test/setup.ts` - adds deterministic browser shims for Ant Design timing in jsdom
- `frontend/src/App.test.tsx` - keeps route permission assertions while awaiting router/layout settle
- `frontend/src/pages/Alerts.test.tsx` - preserves viewer/operator alert action coverage with explicit async settling
- `frontend/src/pages/DataSources.test.tsx` - preserves viewer/admin config visibility coverage with explicit async settling
- `frontend/src/pages/DataSources.tsx` - keeps the hidden editor form connected and removes deprecated input adornment usage that surfaced during verification cleanup

## Decisions Made

- Kept the React Router warning fix in `App.tsx` so runtime and test contracts remain identical
- Chose explicit async settling in tests over global console filtering to avoid hiding regressions
- Fixed the DataSources warning at the component source once `forceRender` exposed a real hidden-form/deprecated-prop issue

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Cleaned DataSources runtime warnings exposed by the test harness**
- **Found during:** Task 2 (Stabilize Ant Design permission tests so benign act warnings stop obscuring results)
- **Issue:** `DataSources` emitted a disconnected `useForm` warning, and the minimal modal fix exposed deprecated `Input addonAfter` warnings during permission tests
- **Fix:** Added `forceRender` on the editor modal and replaced deprecated numeric input adornments with `Space.Compact`-based suffix UI
- **Files modified:** `frontend/src/pages/DataSources.tsx`
- **Verification:** `pnpm test -- --run` output no longer contains `useForm`, `addonAfter`, or `act(...)` warning noise; `pnpm build` passes
- **Committed in:** `9e2796b`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary adjacent cleanup to achieve the plan's warning-free verification target. No unrelated scope expansion.

## Issues Encountered

- Task 2 was briefly interrupted mid-edit; the remaining test harness refactor and verification were resumed and completed in the same plan window

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Frontend permission regression tests now produce high-signal output and are ready for future verification work
- No blocker remains for closing this plan; `STATE.md` and `ROADMAP.md` were intentionally left untouched per execution constraint

## Known Stubs

None.

## Self-Check: PASSED

- Summary file exists at `.planning/phases/09-close-v1-1-milestone-tech-debt/09-02-SUMMARY.md`
- Referenced commits exist: `b3dbe3a`, `a0cae6f`, `9e2796b`
