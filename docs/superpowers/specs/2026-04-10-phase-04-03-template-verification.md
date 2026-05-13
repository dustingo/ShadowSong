---
name: phase-04-03
description: Verification script and evidence for template passthrough feature
metadata:
  type: spec
  source_phase: 04-enable-raw-event-passthrough-in-notification-templates
  source_plan: "03"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 04 Plan 03: Verification Script and Evidence

## Context & Goals

Plan 04-02 added preview UI. This plan proves the raw-event passthrough feature works end-to-end and does not regress legacy templates, closing the phase with repeatable evidence.

**Goal:** Complete TMPL-02 and TMPL-03 — produce executable verification proving end-to-end passthrough and legacy template compatibility.

## Success Criteria

- There is an executable verification path proving a template can render raw webhook fields end-to-end
- There is an executable verification path proving legacy standard-field templates still work after the contract change
- Phase 04 closes with evidence, not just code changes

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Repeatable backend/frontend verification flow | `scripts/verify_template_passthrough.ps1` | Passthrough + compatibility |
| Phase 04 final verification evidence | `04-VERIFICATION.md` | Command outputs, pass/fail |

## Architecture

### Verification Script

**scripts/verify_template_passthrough.ps1** (new):
- Self-contained PowerShell style (same pattern as prior phases)
- Prepares minimal datasource/channel/route setup
- Proves two cases:
  1. Legacy standard-field template still renders and dispatches successfully
  2. Raw webhook field template references nested fields through new passthrough contract
- Reuses lightweight local HTTP listener pattern from Phase 01
- Script responsible for setup + cleanup

### Execution

```powershell
pwsh -ExecutionPolicy Bypass -File scripts/verify_template_passthrough.ps1
pnpm.cmd --dir frontend build
```

### Verification Report

`04-VERIFICATION.md` records:
- Commands executed, pass/fail status
- Observed notification proof points for legacy-template case
- Observed notification proof points for raw-field passthrough case
- Frontend build result

## Implementation Tasks

### Task 1: Script the Phase 04 Passthrough and Compatibility Verification Path

**Files:** `scripts/verify_template_passthrough.ps1`

**Acceptance Criteria:**
- Script asserts both legacy-template compatibility and raw-event passthrough in one run
- Script is self-contained: prepares test data, captures outbound notifications, cleans up
- Failures identify which path broke: normalization, output rendering, routing, or notification dispatch

**Action:** Create `scripts/verify_template_passthrough.ps1` using same self-contained PowerShell style as prior phases. Script should:
- Prepare minimal datasource/channel/route setup
- Prove legacy-template case: datasource output template using only standard fields still renders/dispatches
- Prove raw-event case: datasource output template referencing nested raw webhook fields through new stable passthrough contract renders/dispatches expected notification
- Reuse lightweight local HTTP listener pattern from Phase 01 for notification capture
- Script responsible for setup + cleanup

**Verification:** `pwsh -ExecutionPolicy Bypass -File scripts/verify_template_passthrough.ps1`

---

### Task 2: Run Verification and Record Evidence for the Phase

**Files:** `.planning/phases/04-enable-raw-event-passthrough-in-notification-templates/04-VERIFICATION.md`

**Acceptance Criteria:**
- `04-VERIFICATION.md` names commands executed and observed results
- Evidence includes one passing legacy-template notification and one passing raw-field notification
- Frontend build still passes after datasource UX changes

**Action:** Execute new passthrough verification script and existing frontend build gate. Record concise evidence in `04-VERIFICATION.md`:
- Commands run, pass/fail status
- Observed notification proof points for both cases
- Keep evidence concrete including exact raw field(s) validated in rendered notification output
- Do not modify unrelated verification assets unless new script genuinely needs a shared helper

**Verification:** `pwsh -ExecutionPolicy Bypass -File scripts/verify_template_passthrough.ps1 && pnpm.cmd --dir frontend build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-04-07 | R | verification evidence | mitigate | Write explicit command/result evidence to `04-VERIFICATION.md` so phase can be audited after execution |
| T-04-08 | D | verification script | mitigate | Keep script self-cleaning and reuse proven local listener pattern to avoid hanging Windows processes |
| T-04-09 | T | compatibility claim | mitigate | Assert both old-template and raw-template notification bodies in same scripted run |

## Decisions

- Script reuses proven Phase 01 pattern for notification capture
- Both legacy and new template cases verified in single script execution

## Deviation Log

None — plan executed as written.