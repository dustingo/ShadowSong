---
name: phase-02-01
description: Remove frontend AI entry surface (route/menu/page deletion)
metadata:
  type: spec
  source_phase: 02-remove-frontend-ai-surfaces
  source_plan: "01"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 02 Plan 01: Remove Frontend AI Entry Surface

## Context & Goals

Phase 02 removes frontend AI surfaces after Phase 01 removed backend AI runtime. This plan (02-01) removes the frontend AI entry surface at the route/menu boundary — the `/ai` route, menu item, and `AIAssistant` page file.

**Goal:** Fulfill FEAI-01 by removing AI navigation and page entry points from the authenticated shell with minimal file changes.

## Success Criteria

- Authenticated users no longer see an AI menu item or `/ai` route in the frontend shell
- The app shell title matches non-AI alert-system positioning on screens touched by route cleanup

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Route table and sidebar menu without AI entry points | `frontend/src/App.tsx` | No `/ai` route, no AI menu item |
| Page barrel without AIAssistant export | `frontend/src/pages/index.ts` | Removed AIAssistant export |
| Deleted AI page | `frontend/src/pages/AIAssistant.tsx` | File removed |

## Architecture

### Removal Scope

- `frontend/src/pages/AIAssistant.tsx` — deleted entirely
- `frontend/src/pages/index.ts` — remove `AIAssistant` export
- `frontend/src/App.tsx` — remove `/ai` route, `RobotOutlined` import, `AI 助手` menu item, and brand title string

### Preserved

- Existing shell layout, navigation order, auth guard
- Non-AI routes (`/`, `/alerts`, `/channels`, `/config`, `/users`)

## Implementation Tasks

### Task 1: Remove AI Page Export Surface

**Files:** `frontend/src/pages/index.ts`, `frontend/src/pages/AIAssistant.tsx`

**Acceptance Criteria:**
- `frontend/src/pages/index.ts` no longer exports `AIAssistant`
- `frontend/src/pages/AIAssistant.tsx` is removed from the repo
- No replacement page or placeholder introduced

**Action:** Delete the `AIAssistant` barrel export and remove `frontend/src/pages/AIAssistant.tsx` entirely. No stub route or empty component replacement.

**Verification:** `rg -n 'AIAssistant' frontend/src/pages/index.ts` returns no matches

---

### Task 2: Remove Shell Route, Menu Item, Icon, and Touched AI Title Copy

**Files:** `frontend/src/App.tsx`

**Acceptance Criteria:**
- Sidebar no longer shows AI-related item
- Protected route table no longer registers `/ai`
- Unused `RobotOutlined` and `AIAssistant` imports removed
- Touched shell title string uses non-AI wording

**Action:** Remove `/ai` menu item, `RobotOutlined`, `AIAssistant` import, and `/ai` protected route registration. Change brand/title from `游戏运维 AI 告警系统` to non-AI wording per UI-SPEC. Keep existing shell layout, navigation order, and `RequireAuth` usage.

**Verification:** `rg -n '/ai|AIAssistant|RobotOutlined|AI 助手|游戏运维 AI 告警系统' frontend/src/App.tsx` returns no matches

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-02-01 | E | App.tsx route | mitigate | Remove `/ai` route and menu item together so no protected but unintended surface remains |
| T-02-02 | T | App.tsx title | mitigate | Limit shell edits to AI-specific strings; preserve route order and RequireAuth |

## Decisions

- Full deletion per FEAI-01 — no stub or placeholder
- Brand title changed from `游戏运维 AI 告警系统` to non-AI wording

## Deviation Log

None — plan executed as written.