# Phase 15 Verification

## Scope

Phase 15 verifies notification retry boundaries inside the existing `WebhookHandler -> sendNotification -> notifier.SendToChannel` path. It does not add a queue, durable delivery record, per-channel retry policy, frontend surfacing, or any new persistence surface for terminal failures.

## Executed Commands

```bash
go test ./internal/handlers -run "TestWebhookHandler(.*Retry.*|.*Terminal.*|.*Success.*)" -count=1
go test ./internal/notifier -count=1
go test ./... -count=1
```

## Verified Scenarios

### 1. First-attempt success

- Covered by `TestWebhookHandlerSendNotification_ImmediateSuccessSingleAttempt`.
- Confirms a successful send completes on attempt `1` without extra retries.
- Confirms the send path logs `stage=send_attempt` once and does not emit `stage=terminal_failure`.

### 2. Retry success

- Covered by `TestWebhookHandlerSendNotification_RetrySuccessAfterTransientFailures`.
- Confirms transient send-stage failures retry inside the same async goroutine window and succeed before the retry cap is exhausted.
- Confirms datasource lookup or render failures may still degrade to default notification content before the same send-stage retry boundary begins.
- Confirms attempt logs carry stable diagnosis fields while the terminal-failure branch stays absent on eventual success.

### 3. Retry exhausted

- Covered by `TestWebhookHandlerSendNotification_RetryExhaustLogsTerminalFailureWithoutPersistenceSideEffects`.
- Confirms repeated retryable send-stage failures end in one explicit `stage=terminal_failure` log.
- Confirms the retry budget is fixed at exactly three send attempts, with one `stage=send_attempt` log per attempt.
- Confirms retry exhaustion does not create any new durable delivery record or other database persistence side effect.

## Final Failure Contract

Terminal failure lands in logs only. The minimum final-failure diagnosis fields are:

- `trace_id`
- `alert_id`
- `channel_id`
- `attempt`
- `max_attempts`
- `error`

This phase treats the terminal failure log as the only allowed failure landing zone after retry exhaustion. No queue, retry table, Redis dead-letter record, or alert-side durable delivery artifact was introduced.

## Diagnostic Path Back To Phase 14

Use `trace_id` from the Phase 15 `terminal_failure` log to walk back through the Phase 14 lifecycle evidence:

1. Find the final failure log with `stage=terminal_failure`.
2. Use the same `trace_id` to locate earlier notification attempt and entry logs in Phase 15 and Phase 14.
3. Continue upstream to the Phase 14 lifecycle stages that establish the original alert path:
   - `notification_entry`
   - `route_match`
   - ingest / persist lifecycle stages recorded in Phase 14
4. Correlate the shared `trace_id` to confirm whether the alert was accepted, persisted, routed, and only failed at the bounded notification send stage.

Reference: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md` documents the Phase 14 trace/lifecycle evidence that Phase 15 terminal-failure diagnosis depends on.

## Notes

- Retryability remains scoped to send-stage failures classified by `internal/notifier.IsRetryableSendError`.
- Unsupported channel types and sender init failures remain terminal and do not consume multiple retry attempts.
- Datasource lookup and template/render failures can fall back to default notification content, after which the final `SendToChannel` stage still follows the same bounded retry contract.
- The final automated proof for this plan is `go test ./... -count=1`.
