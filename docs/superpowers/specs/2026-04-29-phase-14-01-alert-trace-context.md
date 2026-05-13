---
name: phase-14-01
description: Establish server-generated trace_id as durable source of truth across webhook ingestion, dedup, Redis handoff, and async notification entry
metadata:
  type: spec
  source_phase: 14-establish-alert-trace-context
  source_plan: "01"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 14 Plan 01: Alert Trace Context Establishment

## Context & Goals

Establish server-generated trace_id as the durable source of truth across webhook ingestion, persistence/dedup branching, Redis handoff, and async notification entry, serving as the foundation for Phases 15/16 reliability and logging unification work.

Purpose: Satisfy D-01, D-02, D-03, D-10, and OBS-01. Output: Alert model with persisted trace_id, webhook-level trace generation/propagation logic, and tests covering multi-alert requests and dedup semantics.

## Success Criteria

- Every webhook request generates exactly one server-side trace_id, never derived from caller-supplied fields
- Newly created Alert rows durably persist trace_id; downstream notification chain can query from alert source of truth
- Multiple new alerts from same webhook request share the same trace_id; alert_id and fingerprint semantics remain unchanged

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Alert model with TraceID | `internal/models/alert.go` | Alert persistence trace_id source of truth |
| Webhook trace handler | `internal/handlers/webhook.go` | trace_id generation, dedup/new split, Redis/notification propagation |
| Webhook trace tests | `internal/handlers/webhook_test.go` | trace generation and propagation baseline tests |

## Architecture

### Key Architectural Decisions

- **Server-generated only:** trace_id minted server-side after datasource auth passes; caller-supplied trace fields are ignored
- **Dedup branch preservation:** dedup path emits trace context with existing_alert_id and fingerprint without creating new row
- **Redis handoff:** Redis stream payload includes trace_id for newly created alerts
- **Async continuity:** processAlertNotificationsAsync receives alerts carrying TraceID intact
- **Scope boundary:** implementation stays inside existing handler/model architecture; no new service layer, middleware rewrite, or caller-supplied trace ingestion

## Implementation Tasks

### Task 1: Add Durable Trace contract to Alert and webhook request setup

**Files:** `internal/models/alert.go`, `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "TraceID\s+string" internal/models/alert.go` finds a persisted Alert field with an index tag
- `rg -n "newTrace|generateTrace|traceID :=" internal/handlers/webhook.go` shows one server-side trace creation site inside HandleWebhook flow
- A focused handler test proves one webhook request with multiple normalized alerts stores the same trace_id on each newly created row
- A focused handler test proves payload fields like trace_id or equivalent input keys do not override the stored trace_id

**Action:** Per D-01, D-02, D-03, D-10 and OBS-01, add a nullable indexed `TraceID string` field to `models.Alert` sized for a server-generated hex identifier. Introduce a small webhook-local helper that mints one trace_id only after datasource lookup and API-key validation succeed. Assign that trace_id to every newly created `Alert` before `db.Create`, without changing existing AlertID generation, Fingerprint generation, or dedup semantics.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(HandleWebhook|Trace)" -count=1`

---

### Task 2: Propagate trace identity through dedup, Redis handoff, and async notification entry

**Files:** `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "\"trace_id\"" internal/handlers/webhook.go` shows Redis payload propagation for new alerts
- `rg -n "existing_alert_id|fingerprint" internal/handlers/webhook.go` shows explicit dedup-path trace context
- Tests cover trace propagation into Redis handoff and verify `processAlertNotificationsAsync` receives alerts carrying TraceID
- No code path accepts inbound trace headers/body as authoritative trace truth

**Action:** Per D-04, D-06, D-08, thread trace_id through existing `HandleWebhook -> publishToRedis -> processAlertNotificationsAsync` seam. Update Redis stream payload to include trace_id for newly created alerts. On dedup branch, keep existing existing.AlertID and existing.Fingerprint behavior intact but record request-level trace context with existing_alert_id and fingerprint.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(HandleWebhook|PublishToRedis|ProcessAlertNotifications).*Trace" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-14-01 | S | webhook request ingest | mitigate | Mint trace_id server-side after datasource auth passes; ignore caller-supplied trace fields |
| T-14-02 | T | Alert persisted trace field | mitigate | Persist only handler-generated value; keep alert_id/fingerprint generation unchanged |
| T-14-03 | R | dedup branch | mitigate | Emit dedup-stage trace context with trace_id, fingerprint, and existing_alert_id |
| T-14-04 | I | webhook lifecycle logging | mitigate | Do not add raw alert.Raw or full request bodies to trace fields/logs |
| T-14-05 | A | Redis handoff | mitigate | Carry trace_id into Redis payloads for later stage correlation |

## Established Patterns

- **Pattern 1:** Server-generated trace_id only after auth validation — caller input never becomes trace authority
- **Pattern 2:** One trace_id per webhook request shared across all alerts from that request
- **Pattern 3:** Dedup path preserves existing_alert_id/fingerprint while exposing request-level trace context

## Decisions

- trace_id is sized for server-generated hex identifier, not caller-supplied string
- Dedup decisions emit trace context (trace_id + fingerprint + existing_alert_id) without creating new Alert row
- Async notification entry receives alerts with TraceID intact for downstream correlation
- No external tracing platform integration; scope limited to webhook-to-notification path

## Deviation Log

None — plan executed as written.
