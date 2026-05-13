---
name: phase-16-04
description: Make canonical writer safely parseable for values containing spaces
metadata:
  type: spec
  source_phase: 16-standardize-alert-path-logging
  source_plan: "04"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 16 Plan 04: Alert Path Logging Parse Fix

## Context & Goals

Close the Phase 16 LOG-01 / T-16-10 blocker where the canonical writer outputs raw space-delimited key=value tokens that become ambiguous when values contain spaces, and existing tests using strings.Fields掩盖 the problem.

Purpose: Current canonical writer unifies field names but output is bare key=value space-separated; values with spaces are parsed ambiguously. Output: Safely parseable alert-path field serialization rules, matching new test parser and regression tests, plus updated 16-VERIFICATION.md, 16-SECURITY.md, 16-UAT.md.

## Success Criteria

- Canonical alert-path writer continues using stable field names but values with spaces are safely parseable, no longer truncated by strings.Fields
- Webhook logging tests parse fields using same output contract, proving error and channel_name values containing spaces are preserved intact
- Phase 16 verification, security, and UAT documents aligned with fixed implementation and automated commands
- Fix scoped to internal/handlers/webhook.go alert main-path and phase-local truth artifacts; no JSON logging migration or full-repo logging cleanup

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook handler with safe serialization | `internal/handlers/webhook.go` | Safely parseable canonical field serialization |
| Webhook test parser | `internal/handlers/webhook_test.go` | Log parser and serialization regression tests handling space-containing values |
| Phase 16 verification doc (updated) | `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md` | Post-fix verification truth |
| Phase 16 security doc (updated) | `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md` | Post-fix threat closure record |
| Phase 16 UAT doc (updated) | `.planning/phases/16-standardize-alert-path-logging/16-UAT.md` | UAT truth aligned with actual evidence |

## Architecture

### Current Blocker Contract

```
Raw space-delimited key=value tokens are ambiguous when values contain spaces.
Tests currently parse via strings.Fields, which loses value tails and hides the ambiguity.
```

### Key Architectural Decisions

- **Quoted or deterministic encoding:** Accept quoted/escaped text values or equally deterministic field encoding
- **key=value searchability preserved:** Overall format remains text-based and key=value-searchable
- **No key renaming:** Existing key names stay stable
- **No JSON migration:** Alert-path logs remain text-based

## Implementation Tasks

### Task 1: Make the canonical writer safely parseable for values containing spaces

**Files:** `internal/handlers/webhook.go`, `internal/handlers/webhook_test.go`

**Acceptance Criteria:**
- `rg -n "strings\.Fields" internal/handlers/webhook_test.go` returns no matches for canonical field parser
- `rg -n "parseWebhookLogFields" internal/handlers/webhook_test.go` shows parser now understands quoted/escaped or deterministic field encoding
- `rg -n "error|channel_name" internal/handlers/webhook_test.go` finds at least one regression proving value with spaces round-trips through serialization and parsing
- `rg -n "logAlertEvent" internal/handlers/webhook.go` still shows one canonical writer; no JSON logger migration

**Action:** Change Phase 16 canonical writer to emit fields deterministically parseable even when values contain spaces. Keep overall contract text-based and key=value-searchable. In webhook_test.go, replace parseWebhookLogFields to decode same serialization rule writer emits, with regression coverage for space-containing value.

**Verification:** `go test ./internal/handlers -run "TestWebhookHandler(.*Logging.*|.*SendNotification.*|.*Terminal.*|.*Panic.*)" -count=1`

---

### Task 2: Refresh verification, security, and UAT truth after the gap fixes

**Files:** `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`, `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`, `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`

**Acceptance Criteria:**
- 16-VERIFICATION.md no longer lists async_panic correlation gap or parse-ambiguity gap as open blockers
- 16-SECURITY.md marks T-16-01 and T-16-10 closed with evidence tied to current files/commands
- 16-UAT.md updated from earlier inaccurate state; results match rerun evidence after fixes
- Documented commands are same ones actually run; go test ./internal/handlers -count=1 passes

**Action:** After Task 1 passes and 16-03 completed, rerun focused and broad handler commands to prove blockers closed. Update Phase 16 truth artifacts to match evidence. 16-VERIFICATION.md marks previously failed truths as verified. 16-SECURITY.md closes T-16-01 and T-16-10 with concrete evidence lines. 16-UAT.md corrected to align with post-fix reality.

**Verification:** `go test ./internal/handlers -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-16-10 | S | canonical writer serialization | mitigate | Encode or quote field values so space-containing data remains unambiguous while preserving key=value searchability |
| T-16-01 | R | Phase 16 truth artifacts | mitigate | Update verification/security/UAT only after passing commands confirm async panic correlation fix from 16-03 |
| T-16-03 | I | logged field values | mitigate | Keep format fix limited to encoding existing bounded fields; no raw payloads or secrets |
| T-16-09 | D | phase scope | mitigate | Restrict edits to webhook.go, webhook_test.go, Phase 16 truth files; not pretext for repo-wide logging cleanup |

## Established Patterns

- **Pattern 1:** Safe serialization preserves key=value searchability while handling space-containing values
- **Pattern 2:** Test parser decodes same serialization rule as writer; tests fail if ambiguity reappears
- **Pattern 3:** Truth artifacts updated only after passing evidence; no stale false positives

## Decisions

- Space-containing values now safely parseable without ambiguity
- Key names unchanged; no JSON migration
- Fix scoped to webhook alert-path logging only

## Deviation Log

None — plan executed as written.
