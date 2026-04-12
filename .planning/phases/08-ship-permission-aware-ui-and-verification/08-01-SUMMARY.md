---
phase: 08-ship-permission-aware-ui-and-verification
plan: 01
subsystem: frontend-shell
tags: [frontend, authz, routes, menu, capabilities]
requirements-completed: [FEACL-01, FEACL-03]
completed: 2026-04-12
---

# Phase 08 Plan 01 Summary

## Accomplishments
- Added a shared frontend capability contract so menu, route gates, and page behavior now map to the same `admin` / `operator` / `viewer` permissions as the backend.
- Updated the main app shell to hide unauthorized menu entries, keep forced-password-reset users inside `/profile`, and show a stable forbidden notice instead of redirecting into confusing blank flows.
- Normalized client auth behavior so `401` clears local auth state and returns to `/login`, while `403` stays on-page for permission-aware UX handling.

## Key Files
- `frontend/src/authz/capabilities.ts`
- `frontend/src/App.tsx`
- `frontend/src/components/PermissionNotice.tsx`
- `frontend/src/components/index.ts`
- `frontend/src/api/client.ts`
- `frontend/src/api/auth.ts`
- `frontend/src/stores/userStore.ts`

## Verification
- `pnpm build`
- `pnpm test -- --run`

## Notes
- The frontend capability matrix intentionally mirrors backend Phase 7 capabilities: `view_alerts`, `process_alerts`, `view_config`, `manage_config`, `manage_users`.
- Forced-password-reset handling continues to reuse the existing login response shape; the UI only consumes `user.force_password_reset`.
