---
name: phase-02-02
description: Remove AI rendering and actions from dashboard, alert cards, and alert list
metadata:
  type: spec
  source_phase: 02-remove-frontend-ai-surfaces
  source_plan: "02"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 02 Plan 02: Remove AI Rendering from Alert Pages

## Context & Goals

Plan 02-01 removed AI route entry. This plan removes AI-derived rendering and action chains from dashboard, alert list, and shared alert card while preserving operational alert workflow.

**Goal:** Fulfill FEAI-02 and eliminate page-level AI state usage before shared API/type contracts are deleted in Plan 03.

## Success Criteria

- Dashboard alert cards expose only operational actions that still exist after backend AI removal
- Alert list detail rows show operational alert data without AI summary, root-cause, or suggestion blocks
- Removing AI content does not break dashboard refresh, websocket, ack, or quick-silence behavior

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Shared alert card without AI action/render blocks | `frontend/src/components/AlertCard.tsx` | No `onAskAI`, no AI summary section |
| Dashboard without AI modal or aiApi flow | `frontend/src/pages/Dashboard.tsx` | Clean alert cards, existing operations |
| Alert detail rows without AI summary block | `frontend/src/pages/Alerts.tsx` | Operational fields only |

## Architecture

### Removal Scope

**AlertCard.tsx:**
- Remove `onAskAI` prop
- Remove AI summary/root-cause/suggestion sections
- Keep `onAck`, `onQuickSilence`, `showActions`

**Dashboard.tsx:**
- Remove AI modal/request flow
- Remove `aiApi` imports
- Remove `react-markdown` usage
- Preserve websocket, polling, stats, trend chart

**Alerts.tsx:**
- Remove AI analysis block from expanded rows
- Remove `ai_summary`, `ai_root_cause`, `ai_suggestions` rendering
- Keep search, filters, pagination, ack/silence modals

## Implementation Tasks

### Task 1: Strip AI Actions and Modal Flow from Dashboard-Active Alert Rendering

**Files:** `frontend/src/components/AlertCard.tsx`, `frontend/src/pages/Dashboard.tsx`

**Acceptance Criteria:**
- AlertCard no longer accepts or invokes `onAskAI`
- Dashboard no longer imports `aiApi` or `react-markdown`
- Active alert cards still support `确认` and `静默` with websocket/polling logic untouched

**Action:** Remove `问 AI` button, AI summary/root-cause/suggestion section, `onAskAI` prop, and dashboard AI modal/request flow. Preserve ack, quick-silence, websocket reconnect, stats cards, trend chart, and active-alert ordering.

**Verification:** `rg -n 'onAskAI|问 AI|AI 分析|AI 响应|aiApi|ReactMarkdown|handleAskAI|handleSendToAI|aiModalVisible|aiResponse' frontend/src/components/AlertCard.tsx frontend/src/pages/Dashboard.tsx` returns no matches

---

### Task 2: Remove AI Detail Fields from Alert-List Expanded Content

**Files:** `frontend/src/pages/Alerts.tsx`

**Acceptance Criteria:**
- Expanded rows retain message and labels
- Expanded rows no longer render `AI 分析` or read `record.ai_summary`
- Search/filter/table behavior remains unchanged

**Action:** Remove AI summary block entirely — no placeholder spacing or substitute text. Keep search filters, pagination, status tags, ack modal, and quick-silence modal unchanged.

**Verification:** `rg -n 'AI 分析|ai_summary|ai_root_cause|ai_suggestions' frontend/src/pages/Alerts.tsx` returns no matches

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-02-03 | D | Dashboard.tsx | mitigate | Remove dead `aiApi` calls instead of letting UI hit already-removed backend AI endpoints |
| T-02-04 | T | AlertCard.tsx | mitigate | Delete AI sections cleanly; preserve only `确认` and `静默` actions |
| T-02-05 | R | Alerts.tsx | accept | Alert detail rows become simpler-only; no audit-sensitive behavior changes |

## Decisions

- AI sections deleted cleanly — no placeholder or compatibility shims
- Operational alert handling (ack, silence, websocket) preserved exactly

## Deviation Log

None — plan executed as written.