---
name: phase-17-01
description: Rename low-risk verification entrypoints and eliminate AI-removal context from test naming
metadata:
  type: spec
  source_phase: 17-clean-truth-and-operational-docs
  source_plan: "01"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 17 Plan 01: Truth and Operational Docs - Verification Scripts

## Context & Goals

Complete low-risk verification entrypoint renaming and test naming cleanup first, closing executable scripts and codebase/testing maps to current baseline naming.

Purpose: Satisfy D-02, D-08, and comply with D-07 high-risk boundaries, so subsequent README/planning truth refresh and runbook can reference stable new verification entrypoints. Output: Renamed verification scripts, synchronized codebase/testing references, and low-risk test identifiers using current baseline naming.

## Success Criteria

- Repo-owned executable backend and frontend verification entrypoints use current naming
- All repo-owned references synchronized to new script paths
- Low-risk test names and codebase/testing entrypoint descriptions no longer frame current baseline as AI-removal对照语境
- Plan touches only low-risk naming surface; does not modify go.mod, Go imports, JWT issuer, or runtime contracts

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Backend alert flow verification script | `scripts/verify_backend_alert_flow.ps1` | Backend alert main-path verification entrypoint |
| Frontend console baseline script | `scripts/verify_frontend_console_baseline.ps1` | Frontend console build and residual scan entrypoint |

## Architecture

### Key Architectural Decisions

- **Script renaming:** verify_backend_no_ai.ps1 -> verify_backend_alert_flow.ps1, verify_frontend_no_ai.ps1 -> verify_frontend_console_baseline.ps1
- **Internal updates:** Temp log filenames, run IDs, step labels no longer contain no_ai or VerifyNoAI
- **Reference maps updated:** .planning/codebase/STRUCTURE.md and .planning/codebase/TESTING.md point to new paths
- **Test renames:** TestLoad_WithoutAIEnv and TestSetup_RoutesWithoutAIRuntime renamed to current baseline names
- **High-risk boundary:** go.mod, Go import paths, JWT issuer, DB names not touched

## Implementation Tasks

### Task 1: Rename low-risk verification entrypoints and eliminate repo-visible old script names

**Files:** `scripts/verify_backend_no_ai.ps1`, `scripts/verify_frontend_no_ai.ps1`, `scripts/verify_backend_alert_flow.ps1`, `scripts/verify_frontend_console_baseline.ps1`, `.planning/codebase/STRUCTURE.md`, `.planning/codebase/TESTING.md`, `internal/config/config_test.go`, `internal/router/router_test.go`

**Acceptance Criteria:**
- Test-Path scripts/verify_backend_alert_flow.ps1 and Test-Path scripts/verify_frontend_console_baseline.ps1 both return true
- `rg -n "verify_backend_no_ai|verify_frontend_no_ai|WithoutAI|WithoutAIRuntime|VerifyNoAI|verify_no_ai" .planning/codebase scripts internal/config internal/router` returns no matches
- `rg -n "verify_backend_alert_flow|verify_frontend_console_baseline|CurrentBaseline|AlertFlow" .planning/codebase scripts internal/config internal/router` returns expected new names

**Action:** Per D-02 and D-08, rename PowerShell verification entrypoints to current-state names. Update script internals so temp filenames, run IDs, step labels no longer contain no_ai or VerifyNoAI. Update repo-owned reference maps in .planning/codebase/STRUCTURE.md and .planning/codebase/TESTING.md to new paths. Rename low-risk test identifiers to current baseline names without changing assertions or runtime behavior.

**Verification:** `pwsh -NoProfile -Command "if (-not (Test-Path 'scripts/verify_backend_alert_flow.ps1')) { exit 1 }; if (-not (Test-Path 'scripts/verify_frontend_console_baseline.ps1')) { exit 1 }; rg -n 'verify_backend_no_ai|verify_frontend_no_ai|WithoutAI|WithoutAIRuntime|VerifyNoAI|verify_no_ai' .planning/codebase scripts internal/config internal/router; if ($LASTEXITCODE -eq 0) { exit 1 }; rg -n 'verify_backend_alert_flow|verify_frontend_console_baseline|CurrentBaseline|AlertFlow' .planning/codebase scripts internal/config internal/router"`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-17-01 | R | codebase/testing maps | mitigate | Replace repo-owned reference-map wording implying AI-removal framing; point executable references to renamed script files per D-02/D-08 |
| T-17-02 | T | script path references | mitigate | Rename script files and update in-repo references atomically so docs cannot point to stale executable names |
| T-17-03 | D | maintainer verification workflow | mitigate | Add grep-based acceptance checks proving old script names and test identifiers gone from maintainer-visible repo surfaces |
| T-17-04 | S | historical naming boundaries | mitigate | Explicitly document go.mod module path and JWT issuer as deferred runtime contracts per D-07/D-09 |

## Established Patterns

- **Pattern 1:** Script renaming done atomically with all in-repo references updated together
- **Pattern 2:** Low-risk test identifiers renamed without changing assertions or runtime behavior
- **Pattern 3:** High-risk runtime contracts (go.mod, JWT issuer) explicitly deferred

## Decisions

- Verification scripts now named for current baseline (alert_flow, console_baseline) rather than AI-removal context
- Test identifiers renamed to current baseline names
- go.mod module path and JWT issuer remain as deferred runtime contracts per D-07

## Deviation Log

None — plan executed as written.
