---
name: phase-02-03
description: Remove AI API wrappers and shared alert types, pass frontend build gate
metadata:
  type: spec
  source_phase: 02-remove-frontend-ai-surfaces
  source_plan: "03"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 02 Plan 03: Remove AI Contracts and Pass Build Gate

## Context & Goals

Plan 02-02 removed AI rendering from pages. This plan removes remaining frontend AI contracts (API wrappers, shared types) and runs the production build gate to close Phase 2 with a self-consistent non-AI frontend.

**Goal:** Fulfill FEAI-03 and close the phase with a static regression check catching orphaned imports, fields, or routes.

## Success Criteria

- Frontend transport and shared types no longer define AI-only endpoints or fields
- Touched product title strings no longer market the frontend as an AI alert system
- Frontend still completes a production build after AI cleanup

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| API wrapper set without aiApi | `frontend/src/api/client.ts` | No `/ai/*` endpoints |
| Alert interface without ai_* fields | `frontend/src/types/index.ts` | Clean Alert type |
| Browser title aligned to non-AI product | `frontend/index.html` | Non-AI branding |

## Architecture

### Contract Removal

**frontend/src/api/client.ts:**
- Remove `aiApi` export
- Remove all `/ai/*` endpoint references

**frontend/src/types/index.ts:**
- Remove `ai_summary`, `ai_root_cause`, `ai_suggestions`, `ai_tags`, `ai_severity` from Alert interface

**frontend/index.html, frontend/src/pages/Login.tsx:**
- Update touched title strings to non-AI wording

### Build Verification

- `pnpm build` must pass in `frontend/`
- `pnpm lint` not required (has pre-existing failures out of scope)

## Implementation Tasks

### Task 1: Remove AI API Wrappers and Shared Alert AI Fields

**Files:** `frontend/src/api/client.ts`, `frontend/src/types/index.ts`

**Acceptance Criteria:**
- `frontend/src/api/client.ts` no longer exports `aiApi` or references `/ai/*` endpoints
- `frontend/src/types/index.ts` removes `ai_*` fields from `Alert`
- Repo-wide frontend grep shows no remaining AI field/endpoint references

**Action:** Delete `aiApi` group from `frontend/src/api/client.ts` and remove AI-only fields from shared `Alert` interface. Verify Plans 01-02 already removed call sites and render usage; clean any residual references. No compatibility shims — backend AI runtime is already gone per FEAI-03 requirement.

**Verification:** `rg -n 'aiApi|/ai/|ai_summary|ai_root_cause|ai_suggestions|ai_tags|ai_severity' frontend/src frontend/index.html` returns no matches

---

### Task 2: Clean Touched Frontend Branding Strings and Pass Build Gate

**Files:** `frontend/src/pages/Login.tsx`, `frontend/index.html`

**Acceptance Criteria:**
- Touched login/title strings no longer include `AI`
- No broad documentation or unrelated branding sweep
- `pnpm build` passes in `frontend/`

**Action:** Update only frontend strings directly touched by Phase 2 cleanup — login heading and browser title no longer describe product as AI alert system. Scope limited to `Login.tsx` and `index.html`; broader docs alignment belongs to Phase 3. Run `pnpm build` as required automated regression gate.

**Verification:** `pnpm build` (exit 0)

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-02-06 | D | client.ts | mitigate | Delete dead `/ai/*` wrappers; prove absence with repo-wide grep |
| T-02-07 | T | index.ts | mitigate | Remove AI-only fields from Alert only after Plans 01-02 clean render call sites |
| T-02-08 | I | Login.tsx, index.html | accept | Product title is low-risk informational copy; scoped string replacement only |

## Decisions

- No compatibility shims or placeholder fields — full contract removal per FEAI-03
- Build gate is primary regression proof; lint not required due to pre-existing failures

## Deviation Log

None — plan executed as written.