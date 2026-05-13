---
name: phase-16-02
description: Lock Phase 16 logging contract in automated tests and verification documentation
metadata:
  type: spec
  source_phase: 16-standardize-alert-path-logging
  source_plan: "02"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 16 Plan 02: Alert Path Logging Verification

## Context & Goals

Lock the Phase 16 alert-path logging contract into automated tests and verification documentation, preventing Phase 14/15/16 observability contracts from reverting to maintenance-by-memory.

Purpose: Satisfy LOG-03 and沉淀 OBS-03 diagnostic path as repo source of truth. Output: Handler contract tests, 16-VERIFICATION.md, and regression commands with log field samples.

## Success Criteria

- Logging contract is明确定义 in automated tests; future changes cannot casually drift field names or move machine values back to message text
- Phase 16 leaves reusable verification record showing how to trace from terminal_failure or send_attempt back to webhook ingest lifecycle via trace_id
- Verification artifacts cover only webhook-to-notification logging contract; no claims of repo-wide logging unification or new logging platform adoption

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook logging contract tests | `internal/handlers/webhook_test.go` | Field-level logging contract regression tests |
| Phase 16 verification doc | `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md` | Phase 16 logging contract truth, sample fields, operator troubleshooting path |

## Architecture

### Expected Contract After 16-01

```
stage
trace_id
alert_id
fingerprint
source
channel_id
channel_name
channel_type
attempt
max_attempts
error
matched_channels
mode
```

### Required Stage Names (Must Remain Searchable)

```
ingest / persist / dedup / redis_publish / route_match / notification_entry / send_attempt / send_notification / terminal_failure / async_panic
```

## Implementation Tasks

### Task 1: Lock the standardized logging contract with field-level handler regressions

**Files:** `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "matched_channels|mode|channel_type" internal/handlers/webhook_test.go` finds explicit field-level assertions for Phase 16 contract
- At least one handler test asserts stage=route_match together with matched_channels=
- At least one handler test asserts stage=terminal_failure or stage=send_attempt together with mode=, attempt=, max_attempts=
- Tests use logger buffer seam in newWebhookTestHandler; no network dependencies or manual-only checks
- No frontend tests, websocket tests, or unrelated backend packages modified

**Action:** Extend internal/handlers/webhook_test.go so Phase 16 contract enforced at field level rather than coarse substring checks. Add or tighten assertions for canonical field names (matched_channels, mode, channel_type) while preserving Phase 14/15 guarantees around stage, trace_id, attempt, max_attempts, terminal_failure. Assert machine-readable values exist as fields, avoid over-coupling to exact message text.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Lifecycle.*|.*Route.*|.*SendNotification.*|.*Terminal.*|.*Panic.*)" -count=1`

---

### Task 2: Record Phase 16 verification truth and operator search path

**Files:** `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`

**Acceptance Criteria:**
- 16-VERIFICATION.md exists and names canonical Phase 16 field set including stage, trace_id, matched_channels, mode, retry-related fields
- Verification doc references both Phase 14 and Phase 15 verification artifacts as upstream carry-forward contract
- Doc explicitly states phase stays on log.Logger and does not introduce slog, JSON logs, websocket cleanup, or external observability infrastructure
- Doc includes at least one example search path from terminal_failure back to ingest using trace_id
- Final verification command is automated and backend-only; no manual-only gate

**Action:** Create 16-VERIFICATION.md after contract tests pass. Document exact automated commands, canonical field vocabulary, troubleshooting path from terminal_failure or send_attempt backward via trace_id to notification_entry, route_match, redis_publish, persist, ingest. Include representative field examples for LOG-03 without pasting full raw payloads.

**Verification:** `go test ./internal/handlers -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-16-06 | R | webhook_test.go contract assertions | mitigate | Add field-level assertions for matched_channels, mode, channel_type, trace_id, terminal_failure so contract drift becomes test-visible |
| T-16-07 | T | 16-VERIFICATION.md | mitigate | Write doc only after automated tests pass; tie directly to Phase 14/15 verification truth |
| T-16-08 | I | verification examples | mitigate | Keep examples limited to field names and bounded values; no raw payloads, secrets, or full alert bodies |
| T-16-09 | D | scope interpretation | mitigate | State explicitly in verification artifact that phase covers only webhook-to-notification logging |
| T-16-10 | S | future logging extensions | mitigate | Document canonical field vocabulary so later contributors extend same contract |

## Established Patterns

- **Pattern 1:** Contract enforced at field level, not substring smoke checks
- **Pattern 2:** Verification doc written only after automated tests pass
- **Pattern 3:** Troubleshooting path walks backward through Phase 14/15 lifecycle using trace_id

## Decisions

- Phase 16 covers only webhook-to-notification logging; does not imply repo-wide standardization
- Existing log.Logger seam preserved; no slog migration
- No JSON logs, websocket cleanup, or external observability infrastructure introduced

## Deviation Log

None — plan executed as written.
