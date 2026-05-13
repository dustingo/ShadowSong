---
name: phase-17-02
description: Refresh README and planning truth surfaces to current operational narrative
metadata:
  type: spec
  source_phase: 17-clean-truth-and-operational-docs
  source_plan: "02"
  milestone: v1.3
  status: completed
  completed: 2026-04-29
---

# Phase 17 Plan 02: Docs and Planning Truth Refresh

## Context & Goals

Refresh README, planning truth sources, and historical supplementary docs to current operational narrative, ensuring maintainers see correct system positioning and deferred naming boundaries first.

Purpose: Satisfy D-01, D-03, D-09, D-10, D-11 and consume 17-01 output new script naming. Output: Updated README.md, .planning/* truth docs, and demoted docs/CODE_REVIEW.md as historical snapshot.

## Success Criteria

- Maintainers see current game-ops alert platform naming in README and planning truth entrypoints, not AI-removal context
- Historical archive facts and high-risk runtime historical naming explicitly marked as historical/deferred migration, not misread as current recommended state
- Supplemental docs do not incorrectly frame outdated docs/CODE_REVIEW.md as current operational guidance

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| README | `README.md` | Current repo operation and verification entrypoints |
| Project truth | `.planning/PROJECT.md` | Current project truth and constraints |
| Roadmap truth | `.planning/ROADMAP.md` | Phase 17 plan count and phase goal truth |
| Historical code review | `docs/CODE_REVIEW.md` | Explicitly marked as historical review snapshot |

## Architecture

### Key Architectural Decisions

- **Active entrypoints describe current system:** README updated to renamed script paths; system framed as ongoing game-ops alert platform
- **Deferred runtime naming noted:** go.mod module path and JWT issuer documented as historical/runtime contracts intentionally deferred per D-07
- **Historical framing for archives:** v1.0 AI Removal Complete preserved as historical shipped context; current entrypoints point to v1.3 reliability/observability truth
- **CODE_REVIEW.md demoted:** Historical-header framing, timestamp scope, pointers to current verification truth added

## Implementation Tasks

### Task 1: Refresh README and planning truth surfaces to current operational narrative

**Files:** `README.md`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`

**Acceptance Criteria:**
- README.md references only verify_backend_alert_flow.ps1 and verify_frontend_console_baseline.ps1 for current verification commands
- .planning/PROJECT.md, .planning/ROADMAP.md, .planning/REQUIREMENTS.md describe Phase 17 as truth-surface cleanup plus maintainer docs for v1.3, not runtime rename phase
- README.md or .planning/PROJECT.md explicitly states go.mod module path and JWT issuer remain deferred historical runtime names

**Action:** Per D-01, D-09, D-10, D-11, rewrite current-truth docs so active entrypoints describe system as ongoing game-ops alert platform. Update README.md verification section to renamed script paths; add note that deeper runtime names (go.mod module path, JWT issuer) are historical/runtime contracts deferred per D-07.

**Verification:** `pwsh -NoProfile -Command "rg -n 'verify_backend_alert_flow\.ps1|verify_frontend_console_baseline\.ps1' README.md; if ($LASTEXITCODE -ne 0) { exit 1 }; rg -n 'truth|运维|reliability|observability|maintainer|维护者' .planning/PROJECT.md .planning/ROADMAP.md .planning/REQUIREMENTS.md; if ($LASTEXITCODE -ne 0) { exit 1 }; rg -n 'go\.mod|JWT issuer|Issuer: \"ai-alert-system\"|历史遗留|暂缓迁移' README.md .planning/PROJECT.md .planning/ROADMAP.md"`

---

### Task 2: Reframe history-bearing docs and demote stale supplemental guidance

**Files:** `.planning/MILESTONES.md`, `.planning/RETROSPECTIVE.md`, `docs/CODE_REVIEW.md`

**Acceptance Criteria:**
- .planning/MILESTONES.md and .planning/RETROSPECTIVE.md preserve AI Removal Complete as history; current-entry wording points to v1.3 reliability/observability truth
- docs/CODE_REVIEW.md contains visible historical framing string (历史审查快照 or Historical)
- docs/CODE_REVIEW.md points readers to Phase 14-16 verification truth or Phase 17 runbook rather than presenting older advice as active guidance

**Action:** Per D-03, D-09, D-10, D-11, preserve v1.0 AI Removal Complete as historical shipped context but add framing it is archive history. In docs/CODE_REVIEW.md, add explicit historical-header framing, timestamp scope, pointers to current verification truth for Phases 14-16 plus Phase 17 runbook. Do not widen into new operational promises or runtime renames.

**Verification:** `pwsh -NoProfile -Command "rg -n 'AI Removal Complete' .planning/MILESTONES.md .planning/RETROSPECTIVE.md; if ($LASTEXITCODE -ne 0) { exit 1 }; rg -n '历史审查快照|Historical' docs/CODE_REVIEW.md; if ($LASTEXITCODE -ne 0) { exit 1 }; rg -n '14-VERIFICATION\.md|15-VERIFICATION\.md|16-VERIFICATION\.md|alert-path-operations-runbook' docs/CODE_REVIEW.md"`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-17-05 | R | README.md / .planning truth surfaces | mitigate | Replace current-entry wording implying AI-removal framing; point verification commands to renamed script files per D-01/D-02 |
| T-17-06 | S | .planning/MILESTONES.md / RETROSPECTIVE.md | mitigate | Keep historical milestone facts but add explicit archive framing so not treated as live operating narrative |
| T-17-07 | T | docs/CODE_REVIEW.md | mitigate | Add visible historical snapshot banner and links to current Phase 14-16 evidence plus Phase 17 runbook |
| T-17-08 | D | deferred runtime naming boundary | mitigate | State in current truth docs that go.mod and JWT issuer remain intentionally unchanged runtime contracts per D-07 |

## Established Patterns

- **Pattern 1:** Current truth surfaces describe live system; archive docs marked as historical
- **Pattern 2:** Deferred runtime contracts (go.mod, JWT issuer) explicitly documented as unchanged
- **Pattern 3:** Stale supplemental guidance demoted with pointers to current evidence

## Decisions

- README now describes game-ops alert platform, not AI-removal cleanup project
- Historical milestone/review language explicitly framed as archive, not active guidance
- Runtime contracts (go.mod, JWT issuer) intentionally deferred per D-07
- No new operational promises made during doc cleanup

## Deviation Log

None — plan executed as written.
