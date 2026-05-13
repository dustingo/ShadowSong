---
name: phase-08-02
description: Page-level alert/config/user permission gating with consistent forbidden feedback
metadata:
  type: spec
  source_phase: 08-ship-permission-aware-ui-and-verification
  source_plan: "02"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 08 Plan 02: Page-Level Permission Gating

## Context & Goals

Actually land Phase 8 permission awareness inside pages: within the same route, different roles can only see or reach actions they have permission for.

Purpose: shell and route alone cannot complete the least-privilege experience; page-internal actions must match Phase 7 backend capability boundaries.
Output: alert/config/user page action visibility closure, read-only notices, consistent forbidden feedback.

## Success Criteria

- `viewer` can see alert and config pages, but no longer see acknowledge, quick-silence, create, edit, delete, enable/disable, or test action entries they have no permission for.
- `operator` can process alerts on the alert page but are read-only on config pages, with clear consistent notices.
- `admin` retains existing config and user management operation surfaces, without losing entries due to permission visibility refactoring.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Alert page permission gating | `frontend/src/pages/Alerts.tsx` | Alert action buttons pruned by capability |
| Config page read-only behavior | `frontend/src/pages/DataSources.tsx` | Read-only config page behavior for non-admin roles |
| User page permission handling | `frontend/src/pages/Users.tsx` | Admin-only user-management surface with consistent forbidden feedback |

## Architecture

### Alert Page Behavior

| Role | View Table | Ack Button | Quick-Silence Button |
|------|------------|-------------|----------------------|
| admin | yes | yes | yes |
| operator | yes | yes (firing only) | yes |
| viewer | yes | no | no |

### Config Pages Behavior

| Role | View List/Detail | Create/Edit/Delete | Enable/Disable/Test |
|------|-------------------|--------------------|--------------------|
| admin | yes | yes | yes |
| operator | yes | no | no |
| viewer | yes | no | no |

### Read-Only Notice Copy

For non-admin on config pages: `当前角色可查看配置，但不能修改`

### Key Decisions

- Button/action visibility determined by capability helper at render time
- Backend `403` responses preserved as user-facing messages (not generic failure toasts)
- Modal entrypoints for forbidden actions cannot be opened by roles lacking capability

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| route-level access -> page-level actions | A readable route can still expose unsafe buttons if action-level gating is missing |
| backend 403 -> page feedback | Permission-denied calls must become consistent UX instead of generic failure toasts |

## Implementation Tasks

### Task 1: Prune alert-action UI so only processing roles can acknowledge or quick-silence

**Files:** `frontend/src/pages/Alerts.tsx`, `frontend/src/stores/alertStore.ts`

**Acceptance Criteria:**
- `frontend/src/pages/Alerts.tsx` references the shared capability helper and does not render ack/quick-silence buttons for `viewer`
- The file contains the read-only or forbidden guidance copy
- `frontend/src/stores/alertStore.ts` still exports `ackAlert` and `quickSilence` but does not discard backend `403` messages

**Action:** Update `frontend/src/pages/Alerts.tsx` so `确认` and `静默` action buttons only render for users who have `process_alerts` and alert is in `firing` state. When current role lacks capability, keep table readable but show compact read-only or permission note. Ensure modal entrypoints cannot be opened by `viewer`. Normalize failed `403` responses through same user-facing copy.

**Verification:** `cd frontend && pnpm build`

---

### Task 2: Convert config and user-management pages to permission-aware read-only/admin-only behavior

**Files:** `frontend/src/pages/DataSources.tsx`, `frontend/src/pages/Channels.tsx`, `frontend/src/pages/RouteRules.tsx`, `frontend/src/pages/Silences.tsx`, `frontend/src/pages/OnDuty.tsx`, `frontend/src/pages/Users.tsx`, `frontend/src/pages/Profile.tsx`, `frontend/src/stores/configStore.ts`

**Acceptance Criteria:**
- Each config page file contains either the shared permission notice or capability gating around create/edit/delete/toggle/test controls
- Non-admin roles no longer render config write buttons
- `frontend/src/pages/Users.tsx` still requires admin intent and contains explicit forbidden handling
- `frontend/src/pages/Profile.tsx` remains self-service only
- `frontend/src/stores/configStore.ts` does not swallow `403` details

**Action:** For config pages, keep list/detail viewing available to roles with `view_config`, but hide or disable every write control mapping to backend `manage_config`: create buttons, edit buttons, delete buttons, enable/disable toggles, test channel actions, reorder actions, save modal footers. Render shared read-only notice near top of each config page for `operator` and `viewer`. For `Users.tsx`, keep page admin-only and add forbidden or redirect-safe fallback. For `Profile.tsx`, preserve self-service editing without exposing admin-only fields. In stores, avoid burying permission errors so pages can show consistent messages when backend `403` occurs.

**Verification:** `cd frontend && pnpm build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-08-04 | E | alert actions in `Alerts.tsx` | mitigate | Hide processing controls for roles lacking `process_alerts` |
| T-08-05 | E | config write buttons in config pages | mitigate | Hide or disable all `manage_config` actions and show read-only notice |
| T-08-06 | R | ambiguous page-level denial feedback | mitigate | Preserve backend `403` details and normalize the UI copy across pages |

## Established Patterns

- **Pattern 1:** Capability gating at button/render level, not just route level
- **Pattern 2:** Shared read-only notice for consistent UX across all config pages
- **Pattern 3:** Backend `403` messages preserved as page-level feedback

## Decisions

- Page-level buttons and forms match the backend role matrix
- Read-only users get explicit guidance instead of silent no-ops
- Admin surfaces remain intact while operator/viewer surfaces stop advertising forbidden actions

## Deviation Log

None
