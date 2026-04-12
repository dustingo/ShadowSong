---
phase: 08-ship-permission-aware-ui-and-verification
plan: 02
subsystem: frontend-pages
tags: [frontend, pages, readonly, forbidden, alerts, config]
requirements-completed: [FEACL-02, FEACL-03]
completed: 2026-04-12
---

# Phase 08 Plan 02 Summary

## Accomplishments
- Updated alert handling UI so only users with `process_alerts` see acknowledge and quick-silence controls; read-only roles now get an explicit notice and row-level `只读` state.
- Locked configuration pages behind consistent read-only behavior: non-admin roles can inspect data sources, channels, route rules, silences, and on-duty records, but cannot see create/edit/delete/toggle controls.
- Reused a shared `PermissionNotice` and backend-derived API error messaging so forbidden or rejected actions surface a consistent explanation across pages.

## Key Files
- `frontend/src/pages/Alerts.tsx`
- `frontend/src/pages/DataSources.tsx`
- `frontend/src/pages/Channels.tsx`
- `frontend/src/pages/RouteRules.tsx`
- `frontend/src/pages/Silences.tsx`
- `frontend/src/pages/OnDuty.tsx`
- `frontend/src/pages/Users.tsx`

## Verification
- `pnpm build`

## Notes
- `operator` remains able to process alerts but is read-only across configuration pages.
- `viewer` stays read-only for both alerts and configuration surfaces.
