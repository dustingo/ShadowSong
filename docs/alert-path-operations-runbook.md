# Alert Path Operations Runbook

## Scope

This is the maintainer-facing runbook for the current webhook-to-notification alert path in `v1.3 Notification Reliability and Observability`.

It only documents behavior already verified in:
- `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`
- `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`
- `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`
- `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`
- `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`

This runbook does not introduce new runtime guarantees, new observability platforms, or permission to rename live runtime contracts.

## Current Guarantees

### Trace Continuity

Evidence source: `14-VERIFICATION.md`

- Each webhook request gets a server-generated `trace_id`.
- The same `trace_id` is persisted on new alerts and reused through downstream logging.
- Maintainers can correlate one failing notification back to upstream webhook handling with the same `trace_id`.

### Stage Taxonomy

Evidence source: `14-VERIFICATION.md`, `16-VERIFICATION.md`

Current stage names that operators can search for in logs:
- `ingest`
- `persist`
- `dedup`
- `redis_publish`
- `route_match`
- `notification_entry`
- `datasource_lookup`
- `render_notification`
- `send_attempt`
- `send_notification`
- `terminal_failure`
- `async_panic`

### Retry Boundary

Evidence source: `15-VERIFICATION.md`

- Send-stage retries are bounded to 3 attempts.
- Retry-related lines keep stable fields including `trace_id`, `alert_id`, `channel_id`, `attempt`, and `max_attempts`.
- Exhausted retries end at `stage=terminal_failure`; non-retryable init/config-style failures are not silently reclassified as retryable sends.

### Logging Contract

Evidence source: `16-VERIFICATION.md`, `16-UAT.md`, `16-SECURITY.md`

- Alert-path logs use one canonical `key=value` contract in the webhook alert path.
- Whitespace-bearing values stay parse-safe through deterministic quoting.
- Failure-path lines such as `async_panic`, `send_attempt`, and `terminal_failure` keep the correlation envelope needed for backtracking.

## Verification Commands

Run the current executable entrypoints first:

```powershell
pwsh -ExecutionPolicy Bypass -File scripts/verify_backend_alert_flow.ps1
pwsh -ExecutionPolicy Bypass -File scripts/verify_frontend_console_baseline.ps1
```

Focused backend checks used by the verified artifacts:

```bash
go test ./internal/handlers -run "TestWebhookHandler(.*Trace.*|.*Redis.*|.*Retry.*|.*Terminal.*|.*Panic.*|.*Logging.*)" -count=1
go test ./internal/handlers -count=1
go test ./internal/notifier -run "Test(.*Retry|.*Retryable|.*Transient|.*Unsupported)" -count=1
```

If you are only checking config and route-baseline naming surfaces:

```bash
go test ./internal/config ./internal/router -count=1
```

## Troubleshooting Path

### Start From `terminal_failure`

Evidence source: `15-VERIFICATION.md`, `16-VERIFICATION.md`

1. Search for `stage=terminal_failure` and capture `trace_id`, `alert_id`, `channel_id`, `attempt`, and `max_attempts`.
2. Confirm the failure really exhausted the bounded retry window instead of being a one-off `send_attempt`.
3. Search the same `trace_id` for earlier `send_attempt` lines to confirm retry progression.
4. Search the same `trace_id` for `notification_entry` and `route_match` to confirm the alert reached routing and notification dispatch.
5. If needed, continue backward to `persist` and `ingest` to verify the alert was accepted and stored before delivery failed.

### Start From `async_panic`

Evidence source: `16-VERIFICATION.md`, `16-SECURITY.md`

1. Search for `stage=async_panic`.
2. Capture the same envelope fields: `trace_id`, `alert_id`, `channel_id`, `channel_name`, and `channel_type`.
3. Search backward by `trace_id` for `notification_entry`, `route_match`, `persist`, and `ingest`.
4. Treat `async_panic` as a rollback-sensitive failure path; if the line no longer carries upstream context, verify whether Phase 16 behavior was reverted.

### Start From Missing Delivery Without Obvious Failure

Evidence source: `14-VERIFICATION.md`, `15-VERIFICATION.md`

1. Search by `trace_id` or `alert_id` if available.
2. Confirm whether the flow reached `route_match`.
3. Confirm whether a `notification_entry` line exists.
4. If there is no `send_attempt`, inspect earlier `datasource_lookup` or `render_notification` behavior.
5. If there is a `send_attempt` but no success line, inspect for `terminal_failure`.

## Rollback-Sensitive Guarantees

Evidence source: `14-VERIFICATION.md`, `15-VERIFICATION.md`, `16-VERIFICATION.md`, `16-SECURITY.md`

Rollback review must preserve these guarantees:
- `trace_id` remains server-generated and reusable across the full alert path.
- The lifecycle stages from `ingest` through `notification_entry` remain searchable.
- Send-stage retries remain bounded and still land on `terminal_failure` when exhausted.
- Failure-path correlation remains intact for `async_panic`.
- Canonical alert-path logging remains parse-safe for values containing spaces.

If any rollback breaks one of the guarantees above, re-run:

```bash
go test ./internal/handlers -run "TestWebhookHandler(.*Trace.*|.*Retry.*|.*Terminal.*|.*Panic.*|.*Logging.*)" -count=1
```

and then re-run:

```powershell
pwsh -ExecutionPolicy Bypass -File scripts/verify_backend_alert_flow.ps1
```

## Deferred Runtime Naming

This section is intentionally explicit about deferred runtime naming.

- The `go.mod` module path remains `github.com/game-ops/ai-alert-system`.
- The JWT issuer remains `ai-alert-system`.

These names are historical/runtime contracts outside Phase 17. Documentation cleanup in this phase does not authorize changing them in code, config, or tokens. Any future migration must be planned as a separate runtime-contract phase.

## Maintainer Notes

- Use this runbook together with the Phase 14-16 verification artifacts when diagnosing regressions.
- Treat `docs/CODE_REVIEW.md` as historical background only; it is not the current operations guide.
- For current roadmap truth and verification entrypoints, prefer `README.md`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, and this runbook.
