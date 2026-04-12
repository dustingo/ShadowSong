---
phase: 08-ship-permission-aware-ui-and-verification
researched: 2026-04-12
status: complete
requirements: [FEACL-01, FEACL-02, FEACL-03, VER-01, VER-02, VER-03, VER-04]
---

# Phase 8 Research

## Objective

Answer: what needs to change so the frontend matches the Phase 7 backend capability matrix, gives stable permission feedback, and leaves behind verification evidence instead of relying on manual spot checks.

## Current State

### Existing Shell And Auth State
- `frontend/src/App.tsx` already has an admin-only `/users` route, a forced-reset redirect to `/profile`, and menu construction based on `user.role`.
- The shell does not expose a reusable capability model; role checks are still ad hoc (`user.role === 'admin'`).
- There is no client-side forbidden UX abstraction. Non-admin `/users` requests redirect to `/`, but page-level action denials in other modules still fall through to generic `message.error(...)`.

### Existing Page-Level Gaps
- `frontend/src/pages/Alerts.tsx` renders `确认` and `静默` actions for any alert in `firing` state. There is no viewer/operator differentiation in the page.
- Config pages such as `frontend/src/pages/DataSources.tsx` still render create/edit/delete/toggle controls unconditionally.
- `frontend/src/stores/configStore.ts` exposes write methods globally and does not carry any permission metadata.
- `frontend/src/api/client.ts` and `frontend/src/api/auth.ts` redirect on `401`, but `403` has no normalized handling or user-facing contract.

### Verification Baseline
- The frontend currently has no test runner in `frontend/package.json`.
- Backend verification exists from Phase 7, but the project requirements still need a consolidated Phase 8 verification artifact that names both backend and frontend role-matrix evidence.

## Planning Conclusions

### Conclusion 1: Add a frontend capability contract, not scattered role string checks
- Mirror the backend capability matrix in a small frontend helper so menu visibility, route decisions, and page actions use one source of truth.
- Keep roles as `admin` / `operator` / `viewer`; do not invent frontend-only pseudo-roles.

### Conclusion 2: Treat config pages as readable but selectively non-mutable
- `admin`: full config CRUD/toggle/test/reorder UI.
- `operator`: keep config pages accessible for viewing, but hide/disable all write actions and show a read-only notice.
- `viewer`: same config read-only behavior as operator, plus no alert acknowledge/quick-silence actions.

### Conclusion 3: Normalize forbidden UX at the client boundary
- Keep `401` behavior as-is: clear session and redirect to `/login`.
- Add a lightweight `403` normalization path that pages can surface consistently, e.g. shared helper/component + standard copy such as `当前角色无权执行该操作`.
- Prefer disabled/hidden controls plus explicit inline page notice over waiting for failing API calls.

### Conclusion 4: Phase 8 should establish a frontend test baseline
- Add a minimal Vite-compatible test setup, likely Vitest + React Testing Library, because `VER-02` needs repeatable evidence.
- Test at least:
  - shell/menu visibility by role
  - alert actions hidden for `viewer`
  - config write controls absent or disabled for non-admin
  - forced-reset/profile path behavior remains intact

### Conclusion 5: Verification should be phase-owned and explicit
- Reuse Phase 7 backend tests as evidence for `VER-01` and `VER-04`.
- Add Phase 8 frontend tests and one verification doc that names exact commands and role outcomes.
- Update requirements traceability and roadmap/state after planning and execution so the milestone closes cleanly.

## Risks To Address In Plans
- Duplicating capability logic in each page will drift from Phase 7; introduce one frontend capability helper first.
- Hiding buttons without page-level copy can make read-only pages feel broken; add explicit read-only/forbidden notices.
- Adding frontend tests without a stable test seam in `App.tsx` will be painful; plan should separate capability helpers and shell logic to make tests straightforward.

## Recommended Plan Shape

### Wave 1
- Introduce frontend capability helper, permission-aware menu/route shell, and shared forbidden/read-only UI primitives.

### Wave 2
- Apply page-level action pruning and consistent 403 messaging across alerts and all config/user-management surfaces.

### Wave 3
- Add frontend test runner and role-matrix tests, then write the phase verification and documentation evidence.
