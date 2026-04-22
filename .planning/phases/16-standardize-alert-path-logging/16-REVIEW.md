---
phase: 16-standardize-alert-path-logging
reviewed: 2026-04-22T01:47:02Z
depth: standard
files_reviewed: 2
files_reviewed_list:
  - internal/handlers/webhook.go
  - internal/handlers/webhook_test.go
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
status: issues_found
---

# Phase 16: Code Review Report

**Reviewed:** 2026-04-22T01:47:02Z
**Depth:** standard
**Files Reviewed:** 2
**Status:** issues_found

## Summary

Reviewed the Phase 16 webhook logging-contract changes in `internal/handlers/webhook.go` and the related regression tests in `internal/handlers/webhook_test.go`. The new canonical writer and the added test coverage are directionally correct, and `go test ./internal/handlers -count=1` passes locally, but two contract issues remain: the panic-recovery path still drops correlation fields, and the supposedly machine-readable `key=value` format is not escaped or encoded, so fields with spaces are not reliably parseable.

## Warnings

### WR-01: `async_panic` logs lose the alert trace context

**File:** `internal/handlers/webhook.go:843-847`
**Issue:** `processAlertNotificationsAsync` recovers panics with `h.logNotification("async_panic", nil, nil, ...)`, which emits no `trace_id`, `alert_id`, `fingerprint`, or channel fields. That breaks the stated Phase 16/OBS-03 goal of tracing a failed notification back through the webhook lifecycle. A sender panic is most likely to happen while handling a specific alert/channel pair, but the recovery record cannot be correlated to that alert anymore.

The current test only checks for `stage=async_panic` and the panic text, so this regression is not caught (`internal/handlers/webhook_test.go:609`).

**Fix:**
```go
func (h *WebhookHandler) processAlertNotifications(alerts []models.Alert) {
	for _, alert := range alerts {
		currentAlert := alert
		func() {
			defer func() {
				if r := recover(); r != nil {
					h.logAlertEvent(
						"async_panic",
						h.traceFieldsForAlert(currentAlert),
						"recovered panic=%v stack=%s",
						r,
						string(debug.Stack()),
					)
				}
			}()

			matchedChannels := h.findMatchedChannels(currentAlert, rules)
			// ...
		}()
	}
}
```

Also add a test that asserts `trace_id=` and `alert_id=` are present on the `async_panic` line.

### WR-02: Canonical log fields are not safely encoded, so spaces break the machine-readable contract

**File:** `internal/handlers/webhook.go:792-805`
**Issue:** `logAlertEvent` formats fields as raw `key=value` tokens separated by spaces, without quoting or escaping values. This makes the contract ambiguous as soon as a value contains whitespace. That already applies to common values in these paths:

- `error` values like `"record not found"` or `"send failed: ... status: 503"`
- `channel_name` if operators create names with spaces
- any future string field promoted into the envelope

Once a value contains spaces, simple token-based parsing only captures the first word. The helper test parser has the same limitation (`internal/handlers/webhook_test.go:1212-1221`), so the suite currently masks this issue instead of detecting it.

**Fix:**
```go
func appendField(parts []string, key, value string) []string {
	if value == "" {
		return parts
	}
	return append(parts, fmt.Sprintf("%s=%q", key, value))
}
```

Then use quoted values consistently in `logAlertEvent`, or switch the writer to a stable encoder such as JSON for the field payload while preserving the existing stage taxonomy. Add a regression test with a space-containing `channel_name` or `error` string and assert the parsed value survives intact.

---

_Reviewed: 2026-04-22T01:47:02Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
