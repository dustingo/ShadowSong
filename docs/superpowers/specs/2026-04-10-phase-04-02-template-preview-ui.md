---
name: phase-04-02
description: Add datasource preview endpoint and frontend template guidance UI
metadata:
  type: spec
  source_phase: 04-enable-raw-event-passthrough-in-notification-templates
  source_plan: "02"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 04 Plan 02: Template Preview and UI Guidance

## Context & Goals

Plan 04-01 expanded the render context. This plan makes the new output-template contract usable by adding a first-class preview/testing path and clear in-UI guidance.

**Goal:** Satisfy TMPL-02 — make raw-event passthrough discoverable and safe to adopt, not just hidden backend capability.

## Success Criteria

- A user editing a data source can see how to reference standard fields and raw-event fields without reverse-engineering backend structs
- A user can run a template preview against sample webhook JSON before saving
- The preview path uses the same render contract as the live notification flow

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Authenticated datasource preview endpoint using backend render contract | `internal/handlers/config.go` | Preview API |
| Datasource preview route wiring | `internal/router/router.go` | Route registration |
| Template guidance plus preview UI | `frontend/src/pages/DataSources.tsx` | Preview tool in editor |

## Architecture

### Preview Endpoint

**POST /api/v1/datasources/preview** (authenticated):
- Request: `{ "datasource_id", "sample_webhook_json", "input_template_override?", "output_template_override?" }`
- Response: `{ "normalized_alert", "rendered_title", "rendered_content", "context_preview" }`
- Uses same contract as live notification rendering from `webhook.go`
- Allows preview of unsaved template changes

### Frontend Preview Flow

- Wire datasource preview endpoint into `client.ts` and `configStore.ts`
- DataSources page: template test drawer becomes real notification-template preview tool
- UI explains stable naming contract: standard top-level fields + raw event variable
- Inline examples for nested fields like `event.annotations.runbook`
- Sample JSON editor + rendered preview result

## Implementation Tasks

### Task 1: Add Authenticated Datasource Output-Preview Endpoint

**Files:** `internal/handlers/config.go`, `internal/router/router.go`

**Acceptance Criteria:**
- Route table includes one authenticated preview endpoint for template debugging
- Preview output produced by same contract as live notification rendering
- Error responses explain parse/template failures clearly for UI display

**Action:** Add protected datasource preview/test endpoint in `internal/router/router.go` and implement in `internal/handlers/config.go`:
- Accept datasource id + sample webhook JSON
- Use datasource `input_template` to normalize sample into `Alert`
- Render datasource `output_template` through Phase 04 backend contract from `webhook.go`
- Allow optional override strings for in-editor `input_template` and `output_template`
- Return structured JSON: normalized alert payload, rendered title/content, context preview/key list
- Do not introduce separate rendering pipeline — call shared helper(s) from `webhook.go`

**Verification:** `go test ./internal/handlers ./internal/router -run "TestWebhook|TestRouter" -count=1`

---

### Task 2: Turn Datasource Editor into Template Guide and Preview Flow

**Files:** `frontend/src/api/client.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/pages/DataSources.tsx`

**Acceptance Criteria:**
- Datasource page exposes concrete template-variable guidance instead of requiring users to infer backend field names
- Users can preview rendered notification output from sample webhook JSON before saving
- Frontend build passes without broken datasource imports, routes, or API contracts

**Action:** Wire new datasource preview endpoint into client and store. Rework `DataSources.tsx` so existing template test drawer becomes real notification-template preview tool:
- Explain stable naming contract explicitly: standard top-level fields remain available, raw event data via chosen top-level raw variable
- Include short inline examples for common nested fields like `event.annotations.runbook`
- Sample JSON editor + rendered preview result showing normalized alert + final title/content
- Keep within existing Ant Design patterns; do not broaden into unrelated datasource UX cleanup

**Verification:** `pnpm.cmd --dir frontend build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-04-04 | D | preview endpoint | mitigate | Reuse existing template helpers and return bounded structured output instead of arbitrary execution paths |
| T-04-05 | I | preview response | accept | Endpoint is JWT-protected config tooling for operators; returning normalized alert/context preview required for template authoring |
| T-04-06 | T | frontend editor | mitigate | Keep preview wired to shared backend contract so docs/examples cannot drift from live notification rendering |

## Decisions

- Preview endpoint shares rendering logic with live notification path — no duplicated pipeline
- UI guidance explicitly documents both standard fields and raw event variable

## Deviation Log

None — plan executed as written.