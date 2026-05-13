---
name: phase-16-03
description: Preserve alert and channel context on async notification panic recovery
metadata:
  type: spec
  source_phase: 16-standardize-alert-path-logging
  source_plan: "03"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 16 Plan 03: Alert Path Async Panic Context Gap Closure

## Context & Goals

Close the Phase 16 OBS-03 / T-16-01 blocker where async_panic failure path loses alert and channel context, making failure notifications untraceable by trace_id back to webhook main path.

Purpose: Current processAlertNotificationsAsync passes nil alert and channel to canonical writer on panic recovery, causing failure notifications to lack trace_id and alert_id correlation. Output: Async panic logging with preserved alert/channel context, locked by handler regression tests.

## Success Criteria

- Async notification panic logs continue using stage=async_panic but retain trace_id, alert_id, fingerprint, source, and available channel fields
- Phase 13/14/15 established trace_id, send_attempt, terminal_failure, and existing stage taxonomy remain searchable without renaming
- Fix scoped to WebhookHandler async notification failure path; no new logging framework or repo-wide logging cleanup

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook handler with context-preserving async panic | `internal/handlers/webhook.go` | Async panic record logic with alert/channel context |
| Async panic regression tests | `internal/handlers/webhook_test.go` | Async panic related field regression tests |

## Architecture

### Current Failing Site

```go
h.logNotification("async_panic", nil, nil, "recovered panic=%v stack=%s", r, string(debug.Stack()))
```

### Required Correlation Vocabulary (Must Survive Unchanged)

```
stage
trace_id
alert_id
fingerprint
source
channel_id
channel_name
channel_type
async_panic
send_attempt
terminal_failure
```

## Implementation Tasks

### Task 1: Preserve alert and channel context on async panic recovery

**Files:** `internal/handlers/webhook.go`

**Acceptance Criteria:**
- `rg -n "async_panic" internal/handlers/webhook.go` shows panic recovery site no longer passes nil, nil into logging path
- `rg -n "trace_id|alert_id|fingerprint|source|channel_id|channel_name|channel_type" internal/handlers/webhook.go` confirms async panic code path routes real context into canonical field envelope
- No log/slog import, no new logging package, no edits outside internal/handlers/webhook.go

**Action:** Refactor processAlertNotificationsAsync so panic recovery closure has access to currently processed alert and current channel. Pass context into existing canonical writer for async_panic, preserving stage=async_panic and current message intent. Do not paper over gap by copying identifiers into free-text message only; correlation data must go through structured field envelope.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Panic.*|.*Trace.*)" -count=1`

---

### Task 2: Add regression coverage for async panic correlation fields

**Files:** `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "async_panic" internal/handlers/webhook_test.go` finds regression test asserting panic line itself, not just generic log presence
- `rg -n "trace_id|alert_id|fingerprint|source|channel_id|channel_name|channel_type" internal/handlers/webhook_test.go` shows explicit assertions tied to async_panic test case
- Focused panic/send notification test command passes without changing any frontend file or non-webhook backend package

**Action:** Extend existing panic regression in internal/handlers/webhook_test.go to capture async_panic line and assert correlation contract: minimum trace_id and alert_id, plus fingerprint, source, channel metadata when test drives channel-aware panic. Reuse handler in-memory logger seam and current helper style.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Panic.*|.*SendNotification.*)" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-16-01 | R | processAlertNotificationsAsync panic recovery | mitigate | Keep alert and channel context reachable inside recovery path; emit trace_id, alert_id, fingerprint, source, channel metadata through canonical writer |
| T-16-05 | S | alert-path stage taxonomy | mitigate | Preserve stage=async_panic and existing Phase 13-15 stage names; do not rename stages while fixing correlation gap |
| T-16-06 | R | handler regression coverage | mitigate | Add field-level tests on emitted async_panic line so future changes cannot silently regress correlation evidence |

## Established Patterns

- **Pattern 1:** Panic recovery preserves structured field envelope, not just free-text message
- **Pattern 2:** Regression tests fail if correlation envelope disappears from async_panic logs
- **Pattern 3:** Context available at panic time is preserved (alert always available; channel available when panic during channel processing)

## Decisions

- Fix scoped to async panic path in internal/handlers/webhook.go only
- No new logger package introduced; reuse existing canonical writer
- Stage name async_panic preserved; no renaming

## Deviation Log

None — plan executed as written.
