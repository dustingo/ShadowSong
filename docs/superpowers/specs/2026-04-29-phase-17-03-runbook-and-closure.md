---
name: phase-17-03
description: Create evergreen alert-path operations runbook and Phase 17 verification/UAT/security truth
metadata:
  type: spec
  source_phase: 17-clean-truth-and-operational-docs
  source_plan: "03"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 17 Plan 03: Runbook and Phase Closure

## Context & Goals

Converge v1.3 verified trace, retry, and logging facts into one maintainer-executable operations runbook, and generate Phase 17 reusable verification/UAT/security truth artifacts.

Purpose: Satisfy D-04, D-05, D-06, D-11, D-12, giving maintainers one stable runbook while enabling direct Phase 17 reuse at v1.3 closure. Output: docs/alert-path-operations-runbook.md and 17-VERIFICATION.md / 17-UAT.md / 17-SECURITY.md.

## Success Criteria

- Maintainers can execute trace/logging/retry troubleshooting directly from one evergreen runbook without assembling multiple phase docs
- Current reliability and observability guarantees, common verification commands, rollback-sensitive points, and deferred runtime naming boundaries are all documented
- Phase 17 verification/UAT/security artifacts accurately reuse Phase 14-16 verified evidence, not inventing new runtime promises

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Alert-path operations runbook | `docs/alert-path-operations-runbook.md` | Maintainer alert main-path troubleshooting and rollback guidance |
| Phase 17 verification truth | `.planning/phases/17-clean-truth-and-operational-docs/17-VERIFICATION.md` | Phase 17 goal verification truth |
| Phase 17 UAT truth | `.planning/phases/17-clean-truth-and-operational-docs/17-UAT.md` | Maintainer perspective operation verification record |
| Phase 17 security truth | `.planning/phases/17-clean-truth-and-operational-docs/17-SECURITY.md` | Phase 17 docs/naming boundary threat closure |

## Architecture

### Key Architectural Decisions

- **Single evergreen runbook:** One maintainer-facing doc for alert-path operations referencing Phase 14-16 evidence
- **Evidence-cited claims:** Every behavioral claim cites or names its source Phase 14/15/16 truth artifact
- **Rollback-sensitive guarantees documented:** Explicit section on what breaks if Phases 14-16 changes reverted
- **Deferred runtime naming explicit:** go.mod module path and JWT issuer remain intentionally unchanged per D-07
- **No new runtime promises:** Artifacts reuse Phase 14-16 facts; do not claim unverified behavior

### Runbook Required Content

- Scope statement
- Current observability chain: trace_id, stage names (ingest to notification_entry), send_attempt, terminal_failure, async_panic
- Step-by-step notification failure troubleshooting path
- Exact verification commands maintainers should run now
- Rollback-sensitive guarantees from Phases 14-16
- Deferred-runtime-naming section for go.mod and JWT issuer

## Implementation Tasks

### Task 1: Create the evergreen maintainer alert-path operations runbook

**Files:** `docs/alert-path-operations-runbook.md`

**Acceptance Criteria:**
- docs/alert-path-operations-runbook.md exists and contains trace_id, terminal_failure, async_panic, rollback, deferred runtime naming
- Runbook names exact supporting evidence files: 14-VERIFICATION.md, 15-VERIFICATION.md, 16-VERIFICATION.md, 16-SECURITY.md, 16-UAT.md
- Runbook lists at least one current executable verification command including scripts/verify_backend_alert_flow.ps1 or focused go test ./internal/handlers commands

**Action:** Create docs/alert-path-operations-runbook.md as maintainer-facing evergreen runbook. Include scope statement, current guaranteed observability chain, step-by-step troubleshooting path for notification failure, exact verification commands, rollback-sensitive guarantees, and deferred-runtime-naming section. Every claim cites Phase 14/15/16 truth artifact; do not invent new runtime promises.

**Verification:** `pwsh -NoProfile -Command "if (-not (Test-Path 'docs/alert-path-operations-runbook.md')) { exit 1 }; rg -n 'trace_id|terminal_failure|async_panic|rollback|deferred runtime naming|14-VERIFICATION\.md|15-VERIFICATION\.md|16-VERIFICATION\.md|16-SECURITY\.md|16-UAT\.md|verify_backend_alert_flow\.ps1|go test ./internal/handlers' docs/alert-path-operations-runbook.md"`

---

### Task 2: Create Phase 17 verification, UAT, and security truth artifacts

**Files:** `.planning/phases/17-clean-truth-and-operational-docs/17-VERIFICATION.md`, `.planning/phases/17-clean-truth-and-operational-docs/17-UAT.md`, `.planning/phases/17-clean-truth-and-operational-docs/17-SECURITY.md`

**Acceptance Criteria:**
- All three files exist under .planning/phases/17-clean-truth-and-operational-docs/
- 17-VERIFICATION.md references DOCS-01, DOCS-02, DOCS-03 coverage and includes current command evidence
- 17-UAT.md includes maintainer test starting from terminal_failure or async_panic following the runbook
- 17-SECURITY.md explicitly mentions deferred runtime naming for go.mod and Issuer: "ai-alert-system"

**Action:** Author Phase 17 truth artifacts closing DOCS-02 and DOCS-03 with evidence. 17-VERIFICATION.md states observable truths for current doc surfaces, script naming, runbook usefulness, deferred runtime naming boundaries, and records exact grep/file/command evidence. 17-UAT.md walks maintainer through verifying renamed script entrypoints and using runbook to trace failure path. 17-SECURITY.md captures documentation-scope threat closure, noting go.mod and JWT issuer intentionally unchanged per D-07.

**Verification:** `pwsh -NoProfile -Command "if (-not (Test-Path '.planning/phases/17-clean-truth-and-operational-docs/17-VERIFICATION.md')) { exit 1 }; if (-not (Test-Path '.planning/phases/17-clean-truth-and-operational-docs/17-UAT.md')) { exit 1 }; if (-not (Test-Path '.planning/phases/17-clean-truth-and-operational-docs/17-SECURITY.md')) { exit 1 }; rg -n 'DOCS-0[123]|verify_backend_alert_flow|verify_frontend_console_baseline|terminal_failure|async_panic|go\.mod|Issuer: \"ai-alert-system\"|deferred' .planning/phases/17-clean-truth-and-operational-docs/17-VERIFICATION.md .planning/phases/17-clean-truth-and-operational-docs/17-UAT.md .planning/phases/17-clean-truth-and-operational-docs/17-SECURITY.md"`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-17-05 | R | docs/alert-path-operations-runbook.md | mitigate | Require runbook cite Phase 14-16 verification/UAT/security evidence for every operational claim |
| T-17-06 | I | runbook examples and truth artifacts | mitigate | Keep examples limited to field names, stage names, commands; exclude raw webhook payloads, secrets, private environment values |
| T-17-07 | D | maintainer rollback flow | mitigate | Add explicit rollback-sensitive guarantees section covering trace persistence, retry boundary, canonical logging fields, parse-safe quoting |
| T-17-08 | T | deferred runtime naming boundary | mitigate | Record in runbook and 17-SECURITY.md that go.mod and JWT issuer remain intentionally unchanged per D-07 |

## Established Patterns

- **Pattern 1:** Runbook cites Phase 14-16 evidence for every behavioral claim
- **Pattern 2:** Truth artifacts written from observed evidence, not speculation
- **Pattern 3:** Rollback-sensitive guarantees explicitly documented so maintainers know what breaks if reverted

## Decisions

- One runbook covers entire alert-path troubleshooting using Phase 14-16 evidence
- Phase 17 artifacts reuse existing Phase 14-16 verified evidence
- go.mod module path and JWT issuer remain unchanged per D-07
- No new runtime behavior promised during Phase 17

## Deviation Log

None — plan executed as written.
