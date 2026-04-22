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

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| webhook request handling -> alert-path event logs | Runtime processing state is transformed into operator-facing evidence. | trace identifiers, alert IDs, routing metadata, bounded error strings |
| alert/channel context -> shared event writer | Internal objects are flattened into a stable text log contract. | alert correlation fields, channel metadata |
| retry and route metadata -> machine-readable fields | Numeric and mode data must remain searchable without relying on free-text parsing. | attempt counts, retry budget, mode, matched channel count |
| implementation log contract -> regression tests | Tests become the enforcement layer for future logging changes. | canonical field names and parsing assumptions |
| verification document -> operator troubleshooting | Maintainers rely on docs to search the right fields and stages under pressure. | field vocabulary, stage taxonomy, traceback flow |

---

## Threat Register

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-16-01 | R | `internal/handlers/webhook.go` field contract | mitigate | Preserve `stage` and `trace_id` from Phases 14/15 and make failure-path reconstruction possible from the canonical writer. | open |
| T-16-02 | T | shared event writer migration | mitigate | Remove or delegate the parallel helper path so future edits cannot silently emit different field sets for the same event family. | closed |
| T-16-03 | I | alert-path logging payload | mitigate | Keep the canonical envelope limited to IDs, routing metadata, retry fields, and bounded error strings; do not add raw payload or secret material. | closed |
| T-16-04 | D | scope of LOG-02 cleanup | mitigate | Limit migration to webhook alert-path `Printf`-style sites so the phase does not expand into a repo-wide logging rewrite. | closed |
| T-16-05 | S | event taxonomy continuity | mitigate | Keep existing stage names instead of inventing a new event namespace. | closed |
| T-16-06 | R | `internal/handlers/webhook_test.go` | mitigate | Add field-level assertions for `matched_channels`, `mode`, `channel_type`, `trace_id`, and `terminal_failure` so contract drift becomes test-visible. | closed |
| T-16-07 | T | `16-VERIFICATION.md` | mitigate | Write the verification doc only after automated tests pass and tie it to Phase 14/15 verification truth. | closed |
| T-16-08 | I | verification examples | mitigate | Keep examples limited to field names and bounded values; do not include raw payloads, secrets, or full alert bodies. | closed |
| T-16-09 | D | scope interpretation | mitigate | State explicitly that Phase 16 covers only webhook-to-notification logging and does not imply repo-wide standardization. | closed |
| T-16-10 | S | future logging extensions | mitigate | Keep the canonical vocabulary safely extensible so later contributors do not inherit an incompatible parsing contract. | open |

*Status: open · closed*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Threat Evidence

### Open Threats

| Threat ID | Evidence | Missing |
|-----------|----------|---------|
| T-16-01 | `async_panic` still logs with `nil` alert/channel context in `internal/handlers/webhook.go`, so failed panic paths can lose `trace_id`, `alert_id`, and related correlation fields. This is also recorded in `16-VERIFICATION.md` as a blocking gap. | Attach alert/channel context to `async_panic` and add regression coverage that asserts correlation fields survive panic recovery. |
| T-16-10 | `logAlertEvent` still emits raw space-delimited `key=value` tokens, and `parseWebhookLogFields` still uses `strings.Fields`, so the contract is not safely extensible when values contain spaces. This remains a documented blocking gap in `16-VERIFICATION.md`. | Quote or encode field values in the canonical writer and add a regression proving values containing spaces remain parseable. |

### Closed Threats

| Threat ID | Evidence |
|-----------|----------|
| T-16-02 | `logAlertEvent` is the canonical writer and legacy helpers delegate into it instead of maintaining a parallel assembly path. |
| T-16-03 | The canonical envelope is limited to IDs, routing metadata, retry metadata, and bounded error strings; no raw payload logging contract was introduced. |
| T-16-04 | Implementation and verification artifacts remain scoped to `internal/handlers/webhook.go` and `internal/handlers/webhook_test.go`, not broader logging infrastructure. |
| T-16-05 | Existing stage names such as `ingest`, `persist`, `route_match`, `notification_entry`, `send_attempt`, `send_notification`, `terminal_failure`, and `async_panic` are preserved. |
| T-16-06 | Regression tests assert core canonical fields like `matched_channels`, `mode`, `channel_type`, `trace_id`, and `terminal_failure`. |
| T-16-07 | The verification artifact records executed automated checks and references earlier phase truth rather than speculative behavior. |
| T-16-08 | Verification examples remain at field vocabulary and troubleshooting-path level. |
| T-16-09 | The verification artifact explicitly states that Phase 16 does not claim repo-wide logging standardization. |

---

## Accepted Risks Log

No accepted risks.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-04-22 | 10 | 8 | 2 | Codex + gsd-security-auditor |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [ ] `threats_open: 0` confirmed
- [ ] `status: verified` set in frontmatter

**Approval:** blocked 2026-04-22
