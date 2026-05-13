---
name: phase-03-01
description: Align README, docs, env, and page title with non-AI product positioning
metadata:
  type: spec
  source_phase: 03-align-docs-and-verification
  source_plan: "01"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 03 Plan 01: Align Documentation and Config with Non-AI Product

## Context & Goals

Phase 03 aligns GSD documentation and configuration with the post-Phase 1/2 non-AI state. This plan (03-01) updates root-level user-facing documents, local environment config reference, and browser title to match "non-AI version" reality.

**Goal:** Complete DATA-01 — ensure README, review docs, and `.env` no longer describe the product as AI-based.

## Success Criteria

- README and user-visible text no longer describe the product as an AI alert system
- Local environment config reference no longer requires OPENAI or AI_* variables for normal startup
- Page title and system test text no longer describe product as AI alert system
- Root-level repo docs no longer guide users to obsolete AI design docs

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Non-AI project README, startup, and structure docs | `README.md` | Clean project description |
| Code review doc aligned with non-AI positioning | `docs/CODE_REVIEW.md` | Updated title and highlights |
| Non-AI local dev config baseline | `.env` | No AI_* variables |
| Browser title aligned to non-AI product | `frontend/index.html` | Non-AI `<title>` |
| Non-AI test notification text | `internal/handlers/config.go` | Neutral test message |

## Architecture

### Updated Files

**README.md:**
- Title, intro, project structure, API docs, startup instructions → non-AI alert system
- Remove references to `internal/ai/` and `.kiro/specs/ai-alert-system/design.md`

**docs/CODE_REVIEW.md:**
- Project name and "code highlights" → non-AI ops alert system

**frontend/index.html:**
- `<title>` → `游戏运维告警系统` (non-AI)

**.env:**
- Remove `OPENAI_API_KEY`, `OPENAI_API_BASE`, `AI_MODEL`, `AI_TIMEOUT`
- Change AI-branded DB name to neutral alert system naming
- README env var docs: only DB, Redis, port, JWT required

**internal/handlers/config.go:**
- Test notification title/content → `游戏运维告警系统` (non-AI)

## Implementation Tasks

### Task 1: Clean README and User-Facing Naming Entry Points

**Files:** `README.md`, `docs/CODE_REVIEW.md`, `frontend/index.html`

**Acceptance Criteria:**
- README, CODE_REVIEW, and index.html no longer contain AI-branded titles or descriptions
- `rg -n '游戏运维 AI 告警系统|AI Alert System|internal/ai|\.kiro/specs/ai-alert-system/design\.md|AI 集成' README.md docs/CODE_REVIEW.md frontend/index.html` returns 0 matches

**Action:** Update README title, intro, project structure, API docs entry, and startup instructions to non-AI alert system. Remove references to `internal/ai/` and AI design docs. Sync CODE_REVIEW project name and highlights. Check index.html `<title>` — if still contains `AI` or old product name, change to `游戏运维告警系统`. Do not rewrite history phase reports or create new AI replacement descriptions.

**Verification:** `rg -n '游戏运维 AI 告警系统|AI Alert System|internal/ai|\.kiro/specs/ai-alert-system/design\.md|AI 集成' README.md docs/CODE_REVIEW.md frontend/index.html` returns 0 matches

---

### Task 2: Close Local Environment Config and Test Text AI Residuals

**Files:** `.env`, `README.md`, `internal/handlers/config.go`

**Acceptance Criteria:**
- `.env` and `README.md` no longer require or show AI variables
- AI-branded DB name cleaned/replaced in these files
- `internal/handlers/config.go` test notification text no longer contains AI product name
- `rg -n 'OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|ai_alert_system|游戏运维 AI 告警系统|AI Alert System' .env README.md internal/handlers/config.go` returns 0 matches

**Action:** In `.env`: delete AI-specific vars and change AI-branded DB name/comments to neutral alert system naming. Update README env var section to clarify normal startup only needs DB, Redis, port, JWT config. Check `internal/handlers/config.go` test notification title/content — if still shows `AI Alert System` or `游戏运维 AI 告警系统`, change to `游戏运维告警系统`. Do not batch-rewrite entire `.env` — only handle AI-specific keys, branded naming, and test text residuals.

**Verification:** `rg -n 'OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|ai_alert_system|游戏运维 AI 告警系统|AI Alert System' .env README.md internal/handlers/config.go` returns 0 matches

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-03-01 | T | README.md | mitigate | Remove obsolete AI module, design doc, and config references to prevent misconfiguration |
| T-03-02 | I | .env | mitigate | Only remove AI-specific keys and close branded naming; do not spread or add sensitive values |
| T-03-03 | R | CODE_REVIEW.md | accept | Keep historical review conclusions; only fix document header and capability descriptions |

## Decisions

- Root-level docs updated to non-AI product description
- `.env` cleaned of AI-specific vars only — other vars untouched
- Test notification text changed to neutral `游戏运维告警系统`

## Deviation Log

None — plan executed as written.