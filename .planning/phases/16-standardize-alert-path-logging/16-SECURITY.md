---
phase: 16
slug: standardize-alert-path-logging
status: blocked
threats_open: 2
asvs_level: 1
created: 2026-04-22
updated: 2026-04-22
---

# Phase 16 — Security

## Scope

This audit re-verifies only the declared Phase 16 threats against the current artifacts:

- `.planning/phases/16-standardize-alert-path-logging/16-01-PLAN.md`
- `.planning/phases/16-standardize-alert-path-logging/16-02-PLAN.md`
- `.planning/phases/16-standardize-alert-path-logging/16-01-SUMMARY.md`
- `.planning/phases/16-standardize-alert-path-logging/16-02-SUMMARY.md`
- `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`
- `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`
- `internal/handlers/webhook.go`
- `internal/handlers/webhook_test.go`

Implementation files were treated as read-only during this audit.

## Threat Verification

| Threat ID | Category | Disposition | Status | Evidence |
|-----------|----------|-------------|--------|----------|
| T-16-01 | R | mitigate | open | `internal/handlers/webhook.go:843-848` still logs `async_panic` through `logNotification("async_panic", nil, nil, ...)`, which drops `trace_id`, `alert_id`, and other alert correlation fields. The same gap is recorded in `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:20-28,57,83,98,109,117`. |
| T-16-02 | T | mitigate | closed | `internal/handlers/webhook.go:790-820` shows `logAlertEvent` as the canonical writer, while `logNotification` and `logTraceStage` now delegate into it instead of assembling fields independently. |
| T-16-03 | I | mitigate | closed | `internal/handlers/webhook.go:763-788,831-840,1027-1042,1067-1085` limits the canonical envelope to IDs, routing metadata, retry metadata, mode, and bounded error strings. No raw webhook payload or secret material was added to alert-path logs. `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:100` records `LOG-02` satisfied without payload expansion. |
| T-16-04 | D | mitigate | closed | Implementation changes remain in `internal/handlers/webhook.go` and `internal/handlers/webhook_test.go`. `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:59,100` explicitly keeps scope on webhook-to-notification logging rather than repo-wide standardization. |
| T-16-05 | S | mitigate | closed | Existing stage names remain in use at `internal/handlers/webhook.go:136,188,214,589,873,879,846,1027,1039,1072,1075,1080,1085`, and tests continue asserting searchable stages in `internal/handlers/webhook_test.go:289-295,609-610,694-701,746-756,1001-1020,1065-1072`. |
| T-16-06 | R | mitigate | closed | Field-level assertions exist in `internal/handlers/webhook_test.go:758-769,821-841,902-910,1011-1021,1065-1072` for `matched_channels`, `mode`, `channel_type`, `trace_id`, and `terminal_failure`. `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:77,101` also records this contract coverage. |
| T-16-07 | T | mitigate | closed | `.planning/phases/16-standardize-alert-path-logging/16-02-SUMMARY.md:68-72` states the verification artifact was written only after focused and broad `go test ./internal/handlers` runs passed. `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:88-93` records those automated commands and results, and `:56-57,76,98-100` ties the doc to Phase 14/15 trace and retry truth. |
| T-16-08 | I | mitigate | closed | `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md` documents field names, bounded values, and troubleshooting flow without including raw payloads, secrets, or full alert bodies; see `:58-59,67-69,75-85`. |
| T-16-09 | D | mitigate | closed | `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:59,67-69` explicitly states the phase stays within `internal/handlers/webhook.go` webhook-to-notification logging and does not claim repo-wide logging standardization. |
| T-16-10 | S | mitigate | open | `internal/handlers/webhook.go:792-810` still emits raw space-delimited `key=value` tokens, and `internal/handlers/webhook_test.go:1212-1221` still parses logs with `strings.Fields`, so the contract is not safely extensible when values contain spaces. This remains documented in `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md:31-38,55,67-68,84-85,99,111`. |

## Open Threats

| Threat ID | Category | Mitigation Expected | Files Searched |
|-----------|----------|---------------------|----------------|
| T-16-01 | R | Preserve failure-path reconstruction by keeping `async_panic` logs correlated with `trace_id`, `alert_id`, fingerprint, source, and channel context through the canonical writer. | `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`, `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`, `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md` |
| T-16-10 | S | Make the canonical vocabulary safely extensible by encoding or quoting field values and proving parser-safe handling for values containing spaces. | `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`, `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`, `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md` |

## Accepted Risks Log

None.

## Transfer Log

None.

## Unregistered Flags

None. Neither `16-01-SUMMARY.md` nor `16-02-SUMMARY.md` contains a `## Threat Flags` section, so there were no summary flags to map or record as unregistered.

## Audit Trail

| Audit Date | Threats Total | Closed | Open | Result |
|------------|---------------|--------|------|--------|
| 2026-04-22 | 10 | 8 | 2 | blocked |
| 2026-04-22 | 10 | 8 | 2 | blocked (re-audit) |
| 2026-04-22 | 10 | 8 | 2 | blocked (verify-open-threats) |
