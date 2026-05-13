---
name: phase-06-03
description: Admin user management UI, self-service profile page, and forced-reset navigation
metadata:
  type: spec
  source_phase: 06-secure-user-administration
  source_plan: "03"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 06 Plan 03: Frontend User Administration and Self-Service UI

## Context & Goals

Expose the minimum frontend entry points for Phase 6: admin user management page, self-service profile/password page, and forced-reset navigation behavior under restricted mode.

Purpose: make the backend Phase 6 boundaries usable in UI, but not prematurely absorb Phase 8 global permission shell engineering.
Output: user type extension, admin user page, profile page, minimal routing/menu wiring, and build verification.

## Success Criteria

- Admin users have a minimal usable user management entry point after login; they can view and maintain users and account states.
- All logged-in users have a self-service profile/password entry; forced-reset accounts are funneled to that page.
- Frontend only adds user management and self-service surfaces required by Phase 6, not extending to Phase 8 global permission visibility/hiding.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Admin user page | `frontend/src/pages/Users.tsx` | Admin-only user list and edit entry point |
| Self-service profile page | `frontend/src/pages/Profile.tsx` | Self-service profile and password modification page |
| App routing and menu | `frontend/src/App.tsx` | Minimal routing/menu wiring and forced-reset redirect |
| Auth API extension | `frontend/src/api/auth.ts` | Admin/self-service user API encapsulation |

## Architecture

### Page Structure

```
/users (admin only)
├── User list with role and account-state columns
├── Create user modal/form
├── Edit user modal/form
├── Disable/enable controls
└── Force-password-reset controls

/profile (all authenticated users)
├── Profile form (name, email)
└── Password change form
```

### Forced-Reset Navigation

When `force_password_reset` is set:
- Redirect any route other than `/profile` (and `/login`) to `/profile`
- Logout remains available
- Present concise prompt that password update is required before returning to normal use

### Key Decisions

- `/users` route and menu visible only to `admin` role
- `/profile` route available to all authenticated users
- Phase 6 scoped strictly - no broader permission-matrix buttons for other modules
- Phase 8 menu pruning not implemented

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| backend user state->frontend local store | Server-provided role/reset flags drive route decisions and must not crash older persisted sessions |
| authenticated UI->admin user page | Frontend can improve UX but must not become the only enforcement boundary |

## Implementation Tasks

### Task 1: Extend frontend auth contracts and store state for account-control flags

**Files:** `frontend/src/api/auth.ts`, `frontend/src/stores/userStore.ts`, `frontend/src/types/index.ts`

**Acceptance Criteria:**
- `frontend/src/types/index.ts` defines optional Phase 6 account-state fields on `User`
- `frontend/src/api/auth.ts` exposes separate admin and self-service methods rather than one ambiguous update call
- `frontend/src/stores/userStore.ts` can parse stored users with or without the new flags

**Action:** Extend the shared `User` type with minimal optional Phase 6 fields returned by backend. In `frontend/src/api/auth.ts`, preserve existing login/logout/current-user contract while adding distinct admin and self-service calls. Update `useUserStore` so persisted user JSON tolerates old localStorage payloads.

**Verification:** `cd frontend; pnpm build`

---

### Task 2: Build the minimum admin Users page and self-service Profile page

**Files:** `frontend/src/pages/Users.tsx`, `frontend/src/pages/Profile.tsx`, `frontend/src/pages/index.ts`

**Acceptance Criteria:**
- `frontend/src/pages/Users.tsx` exists and calls only admin user APIs
- `frontend/src/pages/Profile.tsx` exists and does not expose role/account-state fields to ordinary users
- `frontend/src/pages/index.ts` re-exports both pages

**Action:** Create `frontend/src/pages/Users.tsx` as smallest Ant Design admin user-management surface: list users with role and account-state columns, create-user modal/form, edit-user modal/form, explicit controls for disable/enable and force-password-reset backed by new admin APIs. Create `frontend/src/pages/Profile.tsx` as self-service surface with profile form (`name`, `email`) and separate password form. Re-export both pages via `frontend/src/pages/index.ts`. Do not add broader permission-matrix buttons.

**Verification:** `cd frontend; pnpm build`

---

### Task 3: Wire minimal routes, menu entrypoints, and forced-reset-only navigation

**Files:** `frontend/src/App.tsx`, `frontend/src/stores/userStore.ts`, `frontend/src/api/auth.ts`

**Acceptance Criteria:**
- `frontend/src/App.tsx` contains `/users` and `/profile` routes, shows users entry only for `admin`
- Redirects forced-reset users away from normal routes to `/profile`
- No unrelated module menus or pages are reworked

**Action:** Update `frontend/src/App.tsx` with two new entrypoints: admin-only `/users` route/menu item and `/profile` route for all authenticated users. Keep existing authenticated shell intact. Add minimal route guarding based on stored user role for `/users`. Implement forced-reset limited mode: when current user carries `force_password_reset`, redirect any route other than `/profile` to `/profile`, keep logout available.

**Verification:** `cd frontend; pnpm build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-06-09 | E | `frontend/src/App.tsx` admin route gating | mitigate | Hide and block `/users` client-side for non-admins while relying on backend capability checks as final authority |
| T-06-10 | D | `frontend/src/stores/userStore.ts` old localStorage payloads | mitigate | Treat new Phase 6 flags as optional and parse old persisted user JSON safely |
| T-06-11 | E | forced-reset navigation | mitigate | Redirect reset-required users to `/profile` until password change clears the server flag |
| T-06-12 | T | admin page action wiring | accept | Backend Phase 6 routes remain the source of truth; client-side forms only mirror allowed actions |

## Established Patterns

- **Pattern 1:** Separate admin and self-service API methods prevent field smuggling
- **Pattern 2:** Forced-reset redirect at shell level before any other navigation
- **Pattern 3:** Account-state fields treated as optional for backward compatibility with old localStorage

## Decisions

- Frontend only adds Phase 6 required surfaces, not Phase 8 permission shell work
- Backend remains authoritative; frontend is UX improvement only
- No broader module menu rewrites

## Deviation Log

None
