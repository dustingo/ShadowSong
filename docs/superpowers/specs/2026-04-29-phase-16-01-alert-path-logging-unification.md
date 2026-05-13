---
name: phase-16-01
description: Consolidate split alert-path log helpers into single canonical writer with unified field contract
metadata:
  type: spec
  source_phase: 16-standardize-alert-path-logging
  source_plan: "01"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 16 Plan 01: Alert Path Logging Standardization

## Context & Goals

Consolidate the fragmented alert-path log helpers in internal/handlers/webhook.go into a single shared write path, and migrate high-risk call sites to the unified field contract.

Purpose: Without changing tech stack, existing log.Logger seam, or Phase 14/15 stage/trace_id/retry field semantics, complete one brownfield log contract standardization for operator troubleshooting. Output: Canonical alert-path event writer, unified base envelope, migrated webhook main-path log call sites, and handler tests covering key field contract.

## Success Criteria

- Webhook-to-notification key logs continue using existing stage names and trace_id continuity
- High-risk alert-path logs no longer drift between logNotification and logTraceStage helper paths
- Machine-readable values go into stable fields, not just free-text message content
- Operators can trace from terminal_failure or send_attempt logs back through ingest, persist, route_match, notification_entry using trace_id
- Phase scoped to internal/handlers/webhook.go main-path log contract; no new logging platform, slog migration, or websocket cleanup

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook handler with unified event writer | `internal/handlers/webhook.go` | Canonical alert-path event writer, base envelope helper, migrated high-risk call sites |
| Webhook logging contract tests | `internal/handlers/webhook_test.go` | Shared log contract and key field migration automated assertions |

## Architecture

### Key Architectural Decisions

- **One canonical writer:** Single event writer for webhook-to-notification operational logs, replacing split logNotification/logTraceStage pattern
- **Deterministic key=value text output:** Keep existing log.Logger seam; no JSON, slog, or new logging packages
- **Preserved phase contracts:** Phase 14/15 stage taxonomy (ingest, persist, dedup, redis_publish, route_match, notification_entry, send_attempt, terminal_failure) and correlation keys (trace_id, retry fields) remain intact
- **Field promotion:** Machine-readable values trapped in message text (matched_channels, mode) promoted to named fields
- **Phase-local scope:** No changes to websocket/bootstrap logging or unrelated packages

### Preserved Field and Stage Contract

```
stage
trace_id
alert_id
fingerprint
source
attempt
max_attempts
terminal_failure
ingest / persist / dedup / redis_publish / route_match / notification_entry / send_attempt / send_notification
```

### Field Drift Removed

- logNotification emits channel_type while traceFieldsForNotification does not
- matched_channels trapped in free-text message text
- mode trapped in free-text message text
- Two helper APIs choose field assembly differently

## Implementation Tasks

### Task 1: Introduce one canonical alert-path event writer and base envelope helpers

**Files:** `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "func \(h \*WebhookHandler\) (logAlertEvent|writeAlertEvent|alertPathEvent)" internal/handlers/webhook.go` finds one canonical writer
- `rg -n "channel_type|channel_name|channel_id|trace_id|alert_id|fingerprint|source" internal/handlers/webhook.go` shows shared helper logic
- `rg -n "func \(h \*WebhookHandler\) logNotification|func \(h \*WebhookHandler\) logTraceStage" internal/handlers/webhook.go` shows at most one active path; retained wrapper delegates to canonical writer
- Handler tests exercise shared writer contract through real handler behavior
- No new package under internal/observability, no log/slog import, no websocket/bootstrap logging changes

**Action:** Implement one shared event writer inside internal/handlers/webhook.go for webhook-to-notification operational logs. Replace split logNotification/logTraceStage field-assembly contract with single writer plus explicit helper(s) for canonical base envelope. Preserve existing stage taxonomy and Phase 14/15 correlation keys exactly.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Lifecycle.*|.*Trace.*|.*Redis.*|.*Logging.*)" -count=1`

---

### Task 2: Migrate high-risk webhook call sites and promote machine-readable values into named fields

**Files:** `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "matched_channels" internal/handlers/webhook.go internal/handlers/webhook_test.go` shows matched_channels asserted as named field in route-match logging
- `rg -n "mode" internal/handlers/webhook.go internal/handlers/webhook_test.go` shows mode emitted and asserted as named field on notification send-path logs
- `rg -n "channel_type" internal/handlers/webhook.go internal/handlers/webhook_test.go` shows channel-bearing logs include channel_type through shared contract
- `rg -n "logNotification\(|logTraceStage\(" internal/handlers/webhook.go` returns only canonical-writer delegation or migrated call sites
- Focused handler tests continue asserting stage=terminal_failure and trace_id= on retry-exhausted send paths
- No changes to internal/handlers/websocket.go, cmd/server/main.go, or any frontend file

**Action:** Move high-risk alert-path call sites (HandleWebhook, publishToRedis, processAlertNotificationsAsync, processAlertNotifications, findMatchedChannels, sendNotification, sendChannelNotification) onto shared writer. Promote matched_channels and mode into named fields. Ensure notification-context logs consistently include channel_type alongside channel_id and channel_name.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Route.*|.*SendNotification.*|.*Terminal.*|.*Panic.*|.*Redis.*)" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-16-01 | R | webhook field contract | mitigate | Preserve stage and trace_id; route all high-risk call sites through one writer for path reconstruction |
| T-16-02 | T | shared event writer migration | mitigate | Remove or delegate parallel helper path so future edits cannot silently emit different field sets |
| T-16-03 | I | alert-path logging payload | mitigate | Keep canonical envelope limited to IDs, routing metadata, retry fields, bounded error strings; no raw payload or secrets |
| T-16-04 | D | scope of LOG-02 cleanup | mitigate | Limit migration to webhook alert-path Printf-style sites; not whole-repo logging rewrite |
| T-16-05 | S | event taxonomy continuity | mitigate | Keep existing stage names instead of inventing new event namespace |

## Established Patterns

- **Pattern 1:** One canonical writer replaces multiple helper paths for same event family
- **Pattern 2:** Machine-readable values promoted from free-text message to named fields
- **Pattern 3:** Phase contracts (stage names, trace_id, retry fields) preserved exactly during refactoring

## Decisions

- Existing log.Logger seam preserved; no slog migration
- Stage names unchanged; no new event namespace introduced
- Field contract limited to IDs, routing metadata, retry fields, bounded error strings
- Scope strictly limited to webhook alert-path Printf-style sites in internal/handlers/webhook.go

## Deviation Log

None — plan executed as written.
