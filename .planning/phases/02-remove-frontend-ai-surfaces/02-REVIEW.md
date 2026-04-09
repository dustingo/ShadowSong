---
phase: 02-remove-frontend-ai-surfaces
reviewed: 2026-04-09T11:15:00Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - frontend/src/App.tsx
  - frontend/src/pages/index.ts
  - frontend/src/components/AlertCard.tsx
  - frontend/src/pages/Dashboard.tsx
  - frontend/src/pages/Alerts.tsx
  - frontend/src/api/client.ts
  - frontend/src/types/index.ts
  - frontend/src/pages/Login.tsx
  - frontend/index.html
findings:
  critical: 1
  warning: 1
  info: 2
  total: 4
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-04-09T11:15:00Z
**Depth:** standard
**Files Reviewed:** 9
**Status:** issues_found

## Summary

Phase 02 successfully removed the visible AI route, menu entry, dashboard AI workflow, and shared `aiApi` / alert AI fields from the reviewed frontend surface. I also re-ran the existing frontend gates: `pnpm build` passed, and `pnpm lint` failed with warnings, including warnings inside the scoped files.

No direct functional regression from the AI-removal diff was found in the reviewed files. The main issue is a security-sensitive credential disclosure still present in the touched login screen, plus residual hook/lint debt and missing regression coverage around the removed AI entry points.

## Critical Issues

### CR-01: Login page still exposes a default administrator credential

**File:** `frontend/src/pages/Login.tsx:109-112`
**Issue:** The login screen renders `admin / admin123` directly in the production UI. This is a hardcoded credential disclosure in shipped frontend code. If that seeded account still exists in any deployed environment, the Phase 02 branding update leaves a trivial authentication path visible to every user.
**Fix:**
```tsx
{import.meta.env.DEV && (
  <div style={{ marginTop: 24, textAlign: 'center', color: '#999', fontSize: 12 }}>
    <Space direction="vertical" size={0}>
      <span>本地开发默认账户: admin / admin123</span>
    </Space>
  </div>
)}
```
Remove the credential entirely from production builds and rotate/remove the seeded account outside the client bundle.

## Warnings

### WR-01: Scoped files still fail the repo lint gate because of stale-effect warnings

**File:** `frontend/src/pages/Dashboard.tsx:42-120`, `frontend/src/pages/Alerts.tsx:48-50`
**Issue:** Both touched pages suppress `react-hooks/exhaustive-deps` implicitly by omitting store actions from `useEffect` dependencies. `pnpm lint` currently fails on these warnings. That means Phase 02 is build-green but not lint-clean, and future refactors to Zustand action identities can leave dashboard refresh or alert-list bootstrapping bound to stale closures.
**Fix:**
```tsx
useEffect(() => {
  fetchAlerts()
}, [fetchAlerts])
```
Apply the same pattern to `fetchActiveAlerts`, `fetchStats`, and `setWsConnected`, or wrap the effect logic with a stable event pattern the project adopts consistently.

## Info

### IN-01: Unused import left behind in the touched shell file

**File:** `frontend/src/App.tsx:25`
**Issue:** `authApi` is still imported but unused. This is harmless at runtime, but it is one of the scoped lint warnings and indicates the Phase 02 cleanup did not fully remove now-dead auth-shell references.
**Fix:** Delete the unused `authApi` import.

### IN-02: No automated regression test covers the removed AI frontend surface

**File:** `frontend/src/App.tsx:30-38`, `frontend/src/App.tsx:169-178`, `frontend/src/api/client.ts:43-179`
**Issue:** The AI route/menu/API surface is now removed, but the repo has no frontend test runner or scoped regression test to assert that `/ai` stays unreachable, the menu does not reintroduce an AI entry, and `aiApi` is not exported again. Current verification relies on manual review plus `pnpm build`.
**Fix:** Add a lightweight frontend test setup (for example Vitest + React Testing Library) and cover:
```ts
expect(screen.queryByText('AI 助手')).toBeNull()
expect(routes).not.toContain('/ai')
expect('aiApi' in clientModule).toBe(false)
```

---

_Reviewed: 2026-04-09T11:15:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
