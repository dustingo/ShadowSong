---
phase: 15-harden-notification-retry-boundaries
plan: 01
subsystem: api
tags: [go, webhook, notifier, retry, observability]
requires:
  - phase: 13-harden-notification-delivery-path
    provides: "panic-protected notification seam and deterministic handler/notifier test hooks"
  - phase: 14-establish-alert-trace-context
    provides: "trace_id propagation and notification lifecycle logging context"
provides:
  - "shared retryability classification for notifier send-stage failures"
  - "bounded async notification retries with attempt-level log fields"
  - "terminal failure landing zone after retry exhaustion"
affects: [webhook, notifier, notification-reliability, alert-observability]
tech-stack:
  added: []
  patterns: ["shared send-stage retry classification", "bounded async retry loop with stable attempt log fields"]
key-files:
  created: []
  modified: [internal/notifier/notifier.go, internal/notifier/notifier_test.go, internal/handlers/webhook.go, internal/handlers/webhook_test.go]
key-decisions:
  - "Retryability stays scoped to wrapped SendToChannel send-stage failures so datasource, template, and config errors remain terminal."
  - "All channel types share one fixed retry budget of three attempts with a short in-goroutine delay instead of channel-specific policies."
patterns-established:
  - "Notification attempts log trace_id, alert_id, channel_id, attempt, max_attempts, and error on every send try."
  - "Retry exhaustion lands in a dedicated terminal_failure stage rather than being inferred from scattered attempt logs."
requirements-completed: [NTFY-01, NTFY-02, NTFY-03]
duration: 5min
completed: 2026-04-21
---

# Phase 15 Plan 01: Harden Notification Retry Boundaries Summary

**Webhook notifications now classify transient send failures centrally, retry them inside one bounded async window, and emit attempt-level plus terminal-failure logs with stable trace fields**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-21T20:00:06+08:00
- **Completed:** 2026-04-21T20:04:58+08:00
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added a shared notifier helper that only marks transient `send failed` errors as retryable.
- Wrapped notification sends in one fixed retry loop shared by rendered and default notification paths.
- Logged every send attempt with stable trace and attempt fields, and added a dedicated `terminal_failure` landing zone after retry exhaustion.

## Task Commits

Each task was committed atomically through the TDD cycle:

1. **Task 1 RED: add failing retryability tests for notifier** - `467783b` (test)
2. **Task 1 GREEN: classify retryable send-stage notifier errors** - `a7ea007` (feat)
3. **Task 2 RED: add failing webhook retry boundary tests** - `b2bbc59` (test)
4. **Task 2 GREEN: add bounded notification retries and attempt logs** - `ff8b3ec` (feat)

## Files Created/Modified

- `internal/notifier/notifier.go` - adds the shared send-stage retryability classifier used by webhook notification retries.
- `internal/notifier/notifier_test.go` - proves transient send failures are retryable while unsupported/init/template/datasource failures remain terminal.
- `internal/handlers/webhook.go` - adds bounded retry orchestration, attempt-field logging, terminal failure logging, and a testable sleep seam.
- `internal/handlers/webhook_test.go` - verifies transient retry success, non-retryable early stop, and retry exhaustion logging.

## Decisions Made

- Kept retry classification in `internal/notifier` so all channel types reuse one decision path without introducing a service layer or per-channel policy objects.
- Applied the same bounded retry loop to both rendered and default notification sends, but only the final `SendToChannel` stage participates in retry decisions.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 15 retry boundaries are now explicit and test-covered.
- The next plan can build on the stable attempt-level fields and `terminal_failure` landing zone for broader logging standardization.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- Summary file exists at `.planning/phases/15-harden-notification-retry-boundaries/15-01-SUMMARY.md`.
- Task commit hashes `467783b`, `a7ea007`, `b2bbc59`, and `ff8b3ec` are present in git history.
