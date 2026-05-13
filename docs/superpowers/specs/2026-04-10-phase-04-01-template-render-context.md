---
name: phase-04-01
description: Expand output template render context to include raw event data
metadata:
  type: spec
  source_phase: 04-enable-raw-event-passthrough-in-notification-templates
  source_plan: "01"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 04 Plan 01: Output Template Render Context Expansion

## Context & Goals

Phase 04 enables raw event passthrough in notification templates. This plan (04-01) locks the backend contract for notification template rendering so output templates can access both normalized alert fields and the original webhook payload.

**Goal:** Satisfy TMPL-01 and TMPL-03 — add raw event passthrough without breaking existing templates using only standard fields.

## Success Criteria

- User-authored output templates can keep using existing standard fields like `{{.alert_name}}` and `{{.severity}}` without edits
- User-authored output templates can reference raw webhook JSON through one stable top-level variable
- Notification rendering falls back predictably when raw payload decoding fails or fields are missing

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Output template context builder exposing standard fields plus raw event data | `internal/handlers/webhook.go` | Stable render contract |
| Regression coverage for legacy templates and nested raw-field rendering | `internal/handlers/webhook_test.go` | Compatibility + raw access tests |

## Architecture

### Render Context Contract

**Standard top-level keys (backward compatible):**
- `alert_name`, `severity`, `message`, `source`, `status`, `trigger_time`, `labels`

**New raw-event variable:**
- Stable top-level `event` (or equivalent) decoded from `alert.Raw`
- Allows nested payload data access via `{{index .event "annotations"}}` or dot-safe map access

### Decode Failure Handling

- Malformed or empty `alert.Raw` does not break standard-field rendering
- Raw context degrades to empty/nil rather than panicking

### Key Principles

- Legacy templates work unchanged
- Raw event passthrough is additive, not replacing standard fields
- Reuse existing helper funcs (`toJson`, `default`)
- Narrow scope for new helpers — avoid broad DSL redesign

## Implementation Tasks

### Task 1: Expand Output-Template Render Context Contract

**Files:** `internal/handlers/webhook.go`

**Acceptance Criteria:**
- `renderNotification` accepts legacy templates unchanged
- `renderNotification` exposes documented raw-event variable for arbitrary JSON payload fields
- Notification rendering handles missing raw data without returning malformed output for standard-field-only templates

**Action:** Refactor `renderNotification` in `internal/handlers/webhook.go` to build a formal render context for `output_template`:
- Keep current top-level standard keys for backward compatibility (per TMPL-03)
- Add stable top-level raw-event object (e.g., `event`) decoded from `alert.Raw` (per TMPL-01)
- Expose nested normalized object (`alert`) only if it helps readability
- Reuse existing helpers (`toJson`, `default`) and add only narrowly scoped helpers for nested map safety
- Update default output templates only if they can demonstrate new contract without breaking old templates

**Verification:** `go test ./internal/handlers -run TestWebhook -count=1`

---

### Task 2: Add Focused Regression Tests for Compatibility and Nested Raw Access

**Files:** `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- Tests fail if legacy top-level fields disappear from output-template context
- Tests fail if nested raw payload fields are not reachable from stable passthrough variable
- Tests fail if nil or missing raw payload data causes template execution errors in compatibility scenarios

**Action:** Create or extend `internal/handlers/webhook_test.go` with table-driven tests around `renderNotification`:
- TMPL-03 backward compatibility explicitly
- TMPL-01 nested raw JSON access using realistic webhook payloads (`annotations.runbook`, `summary`, custom top-level fields)
- Decode-failure cases: missing raw payload keys do not crash rendering, fall back through helper/default behavior

**Verification:** `go test ./internal/handlers -run TestWebhook -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-04-01 | T | webhook.go | mitigate | Decode `alert.Raw` and `alert.Labels` defensively, preserve empty maps on failure, keep template helper surface narrow |
| T-04-02 | D | webhook.go | mitigate | Cover malformed/missing raw payload cases in handler tests so bad payloads cannot panic notification rendering |
| T-04-03 | I | outbound notification content | accept | Raw passthrough is explicit user-authored template feature; exposure controlled by template authors, not silently auto-included |

## Decisions

- `event` (or equivalent) is the stable top-level variable for raw webhook JSON
- Standard top-level keys remain unchanged for backward compatibility
- Decode failures degrade gracefully to empty/nil

## Deviation Log

None — plan executed as written.