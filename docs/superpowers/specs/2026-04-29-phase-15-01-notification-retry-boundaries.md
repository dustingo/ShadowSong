---
name: phase-15-01
description: Add bounded retry loop for transient send failures and establish retryability classification
metadata:
  type: spec
  source_phase: 15-harden-notification-retry-boundaries
  source_plan: "01"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 15 Plan 01: Notification Retry Boundaries

## Context & Goals

Add fixed, bounded, short-window retry for transient send-stage failures within the existing WebhookHandler -> sendNotification -> notifier.SendToChannel path, and establish retryability classification for send failures.

Purpose: Per D-01, D-02, D-05, D-07, D-09, D-10 and NTFY-01/NTFY-02/NTFY-03, upgrade notification reliability from first-failure-visible to transient-failure-recoverable-with-diagnosis. Output: Retry orchestration, unified retryability classifier, attempt-level logging, and automated tests covering retry success, non-retryable failure, and retry exhaustion boundaries.

## Success Criteria

- Transient send failures are retried within one bounded goroutine window using fixed shared policy across all channel types
- Retry exhaustion leaves explicit terminal_failure log landing zone (not just scattered attempt failures)
- Non-retryable failures (template render, datasource lookup, channel config errors) do not consume retry budget
- Every send attempt logs minimum stable contract: trace_id, alert_id, channel_id, attempt, max_attempts, error

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook handler with retry logic | `internal/handlers/webhook.go` | Send-stage retry orchestration, attempt-level logs, terminal failure boundary |
| Notifier with retryability classifier | `internal/notifier/notifier.go` | Unified retryable/terminal error classification |
| Webhook retry tests | `internal/handlers/webhook_test.go` | Short-window retry and attempt-level logging tests |
| Notifier retry tests | `internal/notifier/notifier_test.go` | Error classification contract tests |

## Architecture

### Key Architectural Decisions

- **Shared retry policy:** One fixed policy for all channel types, not per-channel branching
- **Short bounded window:** Retries remain inside current async goroutine; one webhook cannot become long-lived background work
- **Attempt-level logging minimum contract:** trace_id, alert_id, channel_id, attempt, max_attempts, error on every attempt
- **Terminal failure landing zone:** Explicit terminal_failure log from sendNotification after retry budget exhausted
- **Classification scope:** Only send-stage failures from SendToChannel classified; datasource/render/config failures stay terminal

### Minimum Attempt-Log Contract

```
trace_id
alert_id
channel_id
attempt
max_attempts
error
```

## Implementation Tasks

### Task 1: Add a shared retryability classifier for send-stage failures only

**Files:** `internal/notifier/notifier.go`, `internal/notifier/notifier_test.go`, `internal/handlers/webhook.go`

**Acceptance Criteria:**
- `rg -n "Retryable|retryable|Classif" internal/notifier/notifier.go` finds one shared retry classification path
- `rg -n "unsupported type|sender init failed|send failed" internal/notifier/notifier.go` shows wrapped channel-context errors distinguished by failure stage
- Focused tests prove transient send failures are retryable while unsupported type or config/init failures are not
- No per-channel retry policy struct, database field, env-driven retry config, or frontend/API surface introduced

**Action:** Per D-01, D-02, D-05, D-10 and NTFY-01, add small shared helper in internal/notifier/notifier.go that classifies wrapped sender errors into retryable vs terminal. Keep classification scoped to send-stage failures from SendToChannel; do not classify datasource lookup, template render failure, or invalid channel configuration as retryable.

**Verification:** `go test ./internal/notifier -run "Test(.*Retry|.*Retryable|.*Transient|.*Unsupported)" -count=1`

---

### Task 2: Apply bounded retries and minimum attempt-level logging in the async notification path

**Files:** `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`, `internal/notifier/notifier.go`

**Acceptance Criteria:**
- `rg -n "attempt|max_attempts" internal/handlers/webhook.go` shows explicit attempt-level log fields in send path
- `rg -n "terminal_failure|final_failure" internal/handlers/webhook.go` finds explicit retry-exhausted log branch in sendNotification
- Focused handler tests prove one transient failure causes more than one send attempt and can succeed before retry cap
- Focused handler tests prove non-retryable errors stop after attempt 1
- Focused handler tests prove retry exhaustion emits one explicit final-failure marker carrying trace_id, alert_id, channel_id, attempt, max_attempts, error
- Implementation stays inside WebhookHandler async notification flow; no queueing, durable delivery records, or per-channel retry config

**Action:** Per D-01, D-03, D-05, D-07, D-08, D-09, D-10 and NTFY-01/NTFY-02/NTFY-03, change sendNotification so actual channel send is wrapped in one fixed retry loop shared by all supported channel types. Keep retry policy bounded and short-window inside current async goroutine. When retry budget exhausted, emit one explicit terminal_failure log from sendNotification containing minimum attempt-log contract.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Retry|.*Attempt|.*Transient|.*NonRetryable)" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-15-01 | T | webhook retry loop | mitigate | Retry only after notifier classification says failure is transient send-stage; keep datasource/render/config failures out of loop |
| T-15-02 | D | async goroutine retry window | mitigate | Small fixed retry count and short delay; cannot turn one webhook into long-lived background work |
| T-15-03 | R | attempt-level diagnosis | mitigate | Log every attempt with trace_id, alert_id, channel_id, attempt, max_attempts, error |
| T-15-04 | S | retry policy ownership | mitigate | One shared policy for all channel types; no channel-specific ad hoc branches silently diverging |
| T-15-05 | I | attempt logs | accept | Existing context fields like channel_name, channel_type, source may remain if already logged; new fields stay within minimum contract |

## Established Patterns

- **Pattern 1:** Transient send-stage failures are retried; deterministic failures (config, render, datasource) are terminal
- **Pattern 2:** Retry budget is fixed and bounded; exhaustion produces explicit terminal_failure log as sole durable landing zone
- **Pattern 3:** Every attempt emits minimum stable logging contract; operators can prove whether retries happened

## Decisions

- Retry policy is shared across all channel types, not per-channel configurable
- Retry window is bounded short-duration inside single async goroutine
- Terminal failure lands in logs only; no durable delivery record, queue, or per-channel policy support added
- Classification helper exposed as simple exported or package-level function without introducing service layer or interface tree

## Deviation Log

None — plan executed as written.
