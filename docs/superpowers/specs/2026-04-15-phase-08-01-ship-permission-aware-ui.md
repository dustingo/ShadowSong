---
name: phase-08-01
description: Frontend permission shell with capability helper and 401/403 semantics
metadata:
  type: spec
  source_phase: 08-ship-permission-aware-ui-and-verification
  source_plan: "01"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 08 Plan 01: Permission-Aware Frontend Shell

## Context & Goals

Build a reusable frontend permission shell first, so App, menu, routes, and API error semantics all depend on a unified capability contract, instead of continuing with scattered role string checks.

Purpose: subsequent Phase 8 pages rely on unified permission contract; without the shell first, second wave would redo each page.
Output: frontend capability helper, shared permission notice component, permission-aware shell and 401/403 client semantics.

## Success Criteria

- Frontend has a single source of truth for permission decisions aligned with backend Phase 7 semantics, not scattered `role === 'admin'` checks throughout pages.
- Menu, routes, and session error handling distinguish 401 from 403, providing a unified permission notice seam for subsequent pages.
- Force-password-reset and admin-only `/users` behavior continue working under the new permission shell.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Frontend capability helper | `frontend/src/authz/capabilities.ts` | Frontend-owned role/capability contract matching backend Phase 7 semantics |
| Permission-aware app shell | `frontend/src/App.tsx` | Permission-aware shell, menu wiring, and route guarding |
| Shared permission notice | `frontend/src/components/PermissionNotice.tsx` | Shared read-only/forbidden feedback component |

## Architecture

### Frontend Capability Matrix

| Role | view_alerts | process_alerts | view_config | manage_config | manage_users |
|------|-------------|----------------|-------------|---------------|--------------|
| admin | yes | yes | yes | yes | yes |
| operator | yes | yes | yes | no | no |
| viewer | yes | no | yes | no | no |

### Exported Helpers

- `can(role, capability)` - check if role has capability
- `canUser(user, capability)` - check if user object has capability
- `isAdmin`, `isReadOnlyConfigUser`, `canProcessAlerts` - page-facing booleans

### 401 vs 403 Semantics

- `401` - logs out and redirects to `/login`
- `403` - preserves backend message (`insufficient permissions` or Chinese copy) for page-level display, no redirect

### Key Decisions

- Frontend capability matrix mirrors backend exactly
- Shared `PermissionNotice` component renders concise read-only or forbidden guidance
- Force-password-reset behavior preserved from Phase 6

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| backend capability matrix -> frontend capability helper | Client UX must reflect the same permission model Phase 7 already enforces on the server |
| API error status -> SPA behavior | `401` should end the session, while `403` should become a local permission message rather than a silent redirect |

## Implementation Tasks

### Task 1: Introduce a frontend capability matrix that mirrors Phase 7 backend semantics

**Files:** `frontend/src/authz/capabilities.ts`, `frontend/src/types/index.ts`, `frontend/src/stores/userStore.ts`

**Acceptance Criteria:**
- `frontend/src/authz/capabilities.ts` defines explicit capability strings and role matrix for `admin`, `operator`, `viewer`
- The file contains `manage_config` and `process_alerts`
- `frontend/src/stores/userStore.ts` still reads old localStorage payloads safely
- `frontend/src/types/index.ts` keeps role as `'admin' | 'operator' | 'viewer'`

**Action:** Create `frontend/src/authz/capabilities.ts` with concrete role/capability constants matching backend matrix exactly. Export `can(role, capability)`, `canUser(user, capability)`, and page-facing booleans. Update types and store for Phase 6 account-control fields without breaking older localStorage payloads.

**Verification:** `cd frontend && pnpm build`

---

### Task 2: Wire the app shell and API clients to the shared permission contract and forbidden semantics

**Files:** `frontend/src/App.tsx`, `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`, `frontend/src/components/PermissionNotice.tsx`, `frontend/src/components/index.ts`

**Acceptance Criteria:**
- `frontend/src/App.tsx` imports the shared capability helper and uses it for menu or route gating
- `frontend/src/components/PermissionNotice.tsx` exists and exports a reusable permission-feedback component
- `frontend/src/api/client.ts` and `frontend/src/api/auth.ts` still clear session on `401` but do not redirect on `403`
- `/users` remains protected while `/profile` and forced-reset behavior still exist in `frontend/src/App.tsx`

**Action:** Refactor `frontend/src/App.tsx` so menu construction and route decisions use the new capability helper instead of inline `role === 'admin'` checks. Preserve current routes but make shell compute access from capabilities. Add shared `PermissionNotice` component for read-only or forbidden guidance. Update API clients so `401` logs out and redirects to `/login`, while `403` preserves backend message for page-level display.

**Verification:** `cd frontend && pnpm build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-08-01 | E | scattered role checks in pages | mitigate | Introduce one frontend capability helper that mirrors backend semantics |
| T-08-02 | R | losing backend deny meaning in SPA | mitigate | Preserve `403` as page-displayable forbidden state instead of treating it like `401` |
| T-08-03 | D | stale localStorage payloads after user type expansion | mitigate | Keep account-control fields optional and parse old payloads safely |

## Established Patterns

- **Pattern 1:** Single capability helper as source of truth for frontend permission decisions
- **Pattern 2:** `401` logout/redirect vs `403` local message distinction at API client level
- **Pattern 3:** Shared `PermissionNotice` component for consistent forbidden UX

## Decisions

- Frontend has a single reusable permission contract
- App shell uses capabilities rather than ad hoc role string checks
- Client can distinguish forbidden UX from expired-session UX

## Deviation Log

None
