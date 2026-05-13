---
name: phase-15-02
description: Lock Phase 15 retry behavior and terminal failure evidence in automated tests and documentation
metadata:
  type: spec
  source_phase: 15-harden-notification-retry-boundaries
  source_plan: "02"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 15 Plan 02: Notification Retry Boundaries Verification

## Context & Goals

Verify Phase 15 retry boundaries, terminal failure landing zone, and diagnosis path, ensuring the "recoverable" and "terminal failure logs only" constraints are locked by automation and documentation truth.

Purpose: Per D-03, D-04, D-07, D-09 and roadmap, verify Phase 15 natural split of implementation vs. verification/diagnostics plan. Output: Automated tests for terminal failure and attempt diagnostics, Phase 15 verification report, and regression commands covering first-success, retry-success, and retry-exhausted scenarios.

## Success Criteria

- Retry exhaustion produces explicit final failure log with traceable evidence
- Verification covers three cases: first-attempt success, retry-then-success, retry-exhausted terminal failure
- Operators can trace from terminal failure log back to Phase 14 trace/lifecycle evidence
- Final failure lands in logs only; no durable delivery records or persistence side effects

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook terminal failure tests | `internal/handlers/webhook_test.go` | Terminal failure log and diagnostic trace-back tests |
| Notifier boundary regression tests | `internal/notifier/notifier_test.go` | Error classification and retry boundary supplementary regression tests |
| Phase 15 verification doc | `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md` | Phase 15 behavior verification and log diagnostics |

## Architecture

### Key Architectural Decisions

- **Terminal evidence contract:** Final failure leaves log entry with stage/trace marker, trace_id, alert_id, channel_id, attempt, max_attempts, error
- **Three scenario coverage:** Immediate success (no retries), transient failure then success (retries worked), retry exhaustion (terminal log)
- **Phase 14 trace linkage:** Terminal failure diagnosis connects back to Phase 14 trace/lifecycle evidence for backward tracing
- **No new persistence surface:** Verification work does not introduce database models, migrations, Redis streams, or durable failure artifacts

### Terminal Evidence Contract

```
stage or final-failure marker
trace_id
alert_id
channel_id
attempt
max_attempts
error
```

## Implementation Tasks

### Task 1: Lock terminal failure behavior and attempt diagnostics with focused regression tests

**Files:** `internal/handlers/webhook_test.go`, `internal/notifier/notifier_test.go`

**Acceptance Criteria:**
- `rg -n "Terminal|Final|RetryExhaust|ImmediateSuccess|RetrySuccess" internal/handlers/webhook_test.go` finds explicit scenario tests for three roadmap cases
- At least one handler test asserts final failure log includes trace_id, alert_id, channel_id, attempt, max_attempts, error
- At least one test proves non-retryable error still results in single send attempt
- No new database model, migration, Redis stream, or durable failure artifact introduced

**Action:** Per D-03, D-04, D-07, D-09 and NTFY-02/NTFY-03, extend focused Go tests to lock Phase 15 by behavior rather than implementation detail. Add one handler test for immediate success, one for transient failure then success, one for retry exhaustion asserting both attempt logs and explicit terminal failure log. Add notifier regression test for unsupported/config errors staying terminal.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Retry.*|.*Terminal.*|.*Success.*)" -count=1 && go test ./internal/notifier -count=1`

---

### Task 2: Record Phase 15 verification truth and operator-facing diagnostics

**Files:** `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`

**Acceptance Criteria:**
- Phase 15 verification doc exists and names three scenarios: first-attempt success, retry success, retry exhausted
- Doc states clearly terminal failure lands in logs only and names minimum final-failure fields
- Doc references Phase 14 trace-based observability as upstream troubleshooting path
- Final automated command is backend regression, not manual-only check

**Action:** Create 15-VERIFICATION.md as phase truth document after tests are green. Document exact commands used, three required scenarios, and diagnostic rule that final failure lands in logs only per D-03. Connect terminal failure diagnosis back to Phase 14 trace/lifecycle evidence.

**Verification:** `go test ./... -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-15-06 | R | terminal failure diagnosis | mitigate | Require explicit retry-exhausted tests and verification doc recording final-failure log contract |
| T-15-07 | T | verification drift | mitigate | Use focused automated tests plus `go test ./... -count=1` so contract enforced by CI-visible commands |
| T-15-08 | I | verification examples | mitigate | Document only bounded field names and troubleshooting steps; no raw payload bodies or secrets |
| T-15-09 | D | retry-scope creep | mitigate | Verification explicitly states no durable delivery record, queue, or per-channel config added |
| T-15-10 | S | cross-phase trace correlation | mitigate | Reference Phase 14 trace_id lifecycle path directly for operator trust |

## Established Patterns

- **Pattern 1:** Phase 15 verification locked by automated tests, not prose alone
- **Pattern 2:** Three-scenario coverage (immediate success, retry success, retry exhausted) proves full retry lifecycle
- **Pattern 3:** Terminal failure diagnosis path connects backward through Phase 14 trace evidence

## Decisions

- Verification artifacts define contract that future changes must preserve
- Final failure evidence is logs only; no queueing or durable delivery records introduced
- Phase 15 verification does not claim frontend surfacing or per-channel policy support

## Deviation Log

None — plan executed as written.
