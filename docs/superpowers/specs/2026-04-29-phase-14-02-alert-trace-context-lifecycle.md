---
name: phase-14-02
description: Complete webhook main-path lifecycle observability with stage-level trace_id coverage
metadata:
  type: spec
  source_phase: 14-establish-alert-trace-context
  source_plan: "02"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 14 Plan 02: Alert Trace Context Lifecycle Observability

## Context & Goals

Fill the gaps in webhook main-path lifecycle observability, ensuring trace_id covers ingest, persistence/dedup, Redis publish, route match, and notification entry — the five key troubleshooting checkpoints.

Purpose: Satisfy D-04 to D-09, D-12 and OBS-02, forming a reusable low-invasion observability contract for Phases 15/16 while explicitly handling Redis publish failure and dedup path. Output: Unified stage log helper/field conventions, handler tests covering success and failure branches, and executable regression verification commands.

## Success Criteria

- Operators can use the same trace_id to trace backward from terminal failure through ingest, persistence/dedup, Redis publish, route match, and notification entry logs
- Redis publish success and failure both leave stable, searchable lifecycle observability points — no silent results
- Dedup path explicitly records trace_id, fingerprint, and existing_alert_id instead of only leaving implicit database updates

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook handler with unified lifecycle helpers | `internal/handlers/webhook.go` | Unified stage observability helper and key-stage logs |
| Webhook lifecycle tests | `internal/handlers/webhook_test.go` | Lifecycle observability verification and Redis failure visibility tests |

## Architecture

### Key Architectural Decisions

- **Stage taxonomy preserved:** Existing stage names (ingest, persist, dedup, redis_publish, route_match, notification_entry) remain stable
- **Phase-local scope:** Changes limited to internal/handlers/webhook.go; no global logging rewrite, WebSocket handling, or frontend changes
- **Redis failure visibility:** Both XAdd success and failure explicitly logged with stream name and message id or error
- **Dedup context:** Dedup stage logs include trace_id, fingerprint, and existing_alert_id together
- **Payload protection:** New logs avoid dumping raw payload bodies while carrying enough identifiers for troubleshooting

### Expected Lifecycle Fields

```
stage
trace_id
alert_id
fingerprint
existing_alert_id
source
channel_id
channel_name
redis_stream
redis_message_id
```

## Implementation Tasks

### Task 1: Add stage-level observability for ingest, persistence, dedup, Redis, route match, and notification entry

**Files:** `internal/handlers/webhook.go`

**Acceptance Criteria:**
- `rg -n "stage=(ingest|persist|dedup|redis_publish|route_match|notification_entry)" internal/handlers/webhook.go` finds explicit lifecycle stages
- `rg -n "redis_message_id|redis_stream|XAdd" internal/handlers/webhook.go` shows Redis success/failure observability
- `rg -n "existing_alert_id.*fingerprint|fingerprint.*existing_alert_id" internal/handlers/webhook.go` confirms dedup stages include both fields with trace_id
- No new lifecycle log prints full alert.Raw or raw request bodies

**Action:** Per D-04, D-06, D-08, D-09, D-12 and OBS-02, extend existing handler-local logging into a small lifecycle helper. Emit explicit stages for request ingest, create success, dedup hit, Redis publish success/failure, route-rule result, and notification entry before channel send begins. Include trace_id on every stage and include fingerprint plus existing_alert_id on dedup stage. When Redis XAdd succeeds, capture returned message ID and stream name; when it fails, log the failure explicitly.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(LogsLifecycleStages|RedisPublishFailure|Dedup).*" -count=1`

---

### Task 2: Lock lifecycle observability with focused handler tests and regression commands

**Files:** `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "LogsLifecycleStages|RedisPublishFailure|DedupTrace" internal/handlers/webhook_test.go` finds dedicated lifecycle observability tests
- At least one test asserts the same trace_id appears in multiple stage logs, not just a single helper invocation
- At least one test asserts Redis publish failure generates a log entry carrying trace_id and Redis stage metadata
- At least one test asserts dedup logs carry existing_alert_id and fingerprint together

**Action:** Extend internal/handlers/webhook_test.go with deterministic sqlite/logger-based tests exercising full backend-only observability contract. Verify one trace_id can be found across ingest, persistence/dedup, route match, and notification entry logs, and Redis publish failures are logged instead of silently disappearing.

**Verification:** `go test ./internal/handlers -count=1 && go test ./... -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-14-06 | S | lifecycle logging helper | mitigate | Always source trace_id from server-generated alert/request context, never from inbound headers/body |
| T-14-07 | I | ingest/persist/dedup logs | mitigate | Restrict new observability fields to identifiers and bounded error context; do not log full raw payload JSON or alert.Raw |
| T-14-08 | R | Redis publish branch | mitigate | Capture both success and failure with trace_id, stream name, returned message id/error so Redis handoff cannot fail silently |
| T-14-09 | R | route/dedup lifecycle evidence | mitigate | Emit explicit stages for route miss/hit and dedup decisions so operators can prove whether alert entered routing/notification |
| T-14-10 | D | backend logging scope | accept | Global logging standardization deferred to Phase 16; Phase 14 limits changes to webhook path |

## Established Patterns

- **Pattern 1:** One trace_id follows an alert through all five lifecycle stages — operators can reconstruct full backend path from one ID
- **Pattern 2:** Redis publish is never fire-and-forget — both success and failure leave traceable evidence
- **Pattern 3:** Dedup decisions emit structured context (trace_id + fingerprint + existing_alert_id) rather than relying on database state inference

## Decisions

- Phase 14 intentionally limits scope to webhook path to avoid destabilizing unrelated flows
- Global logging standardization (field naming, output format) deferred to Phase 16
- Lifecycle helper is handler-local only; not introduced as a shared package or middleware
- No JSON logging migration; logs remain text-based key=value format

## Deviation Log

None — plan executed as written.
