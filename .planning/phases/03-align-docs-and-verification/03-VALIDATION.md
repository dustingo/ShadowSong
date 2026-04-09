---
phase: 03-align-docs-and-verification
created: 2026-04-09
status: pending
---

# Phase 03 Validation Strategy

## Validation Architecture

Phase 03 is a docs/config/verification closure phase. Validation therefore combines:

1. Static grep checks over Phase 03-owned docs, config, and codebase-map files.
2. A backend non-AI runtime script: `scripts/verify_backend_no_ai.ps1`.
3. A frontend non-AI verification script added in this phase: `scripts/verify_frontend_no_ai.ps1`.
4. A final phase report in `03-VERIFICATION.md` that records actual command execution and results.

## Validation Matrix

| Check ID | Plan | Task | Requirement | Validation Type | Command / Evidence | Status |
|---------|------|------|-------------|-----------------|--------------------|--------|
| 03-01-01 | 03-01 | 1 | DATA-01 | static grep | `rg -n '游戏运维 AI 告警系统|AI Alert System|internal/ai|\.kiro/specs/ai-alert-system/design\.md|AI 集成' README.md docs/CODE_REVIEW.md frontend/index.html` | pending |
| 03-01-02 | 03-01 | 2 | DATA-01 | static grep | `rg -n 'OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|ai_alert_system|游戏运维 AI 告警系统|AI Alert System' .env README.md internal/handlers/config.go` | pending |
| 03-02-01 | 03-02 | 1 | DATA-02 | static grep | `rg -n 'internal/ai|internal/handlers/ai\.go|AIAssistant|AILog|SilenceRecommendation|OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|OpenAI-compatible|AI Assistant Flow' .planning/codebase/ARCHITECTURE.md .planning/codebase/STACK.md .planning/codebase/STRUCTURE.md .planning/codebase/INTEGRATIONS.md` | pending |
| 03-02-02 | 03-02 | 2 | DATA-02 | static grep | `rg -n 'NewAIHandler|AIAssistant|internal/ai/client\.go|internal/handlers/ai\.go|AILog|AI suggestion|AI interactions|AI assistant' .planning/codebase/CONVENTIONS.md .planning/codebase/CONCERNS.md .planning/codebase/TESTING.md` | pending |
| 03-03-01 | 03-03 | 1 | VER-01 | script | `pwsh -ExecutionPolicy Bypass -File scripts/verify_backend_no_ai.ps1` | pending |
| 03-03-02 | 03-03 | 1 | VER-02 | script | `pwsh -ExecutionPolicy Bypass -File scripts/verify_frontend_no_ai.ps1` | pending |
| 03-03-03 | 03-03 | 2 | DATA-01, VER-01, VER-02 | script chain + report | `pwsh -NoProfile -ExecutionPolicy Bypass -Command "& { ./scripts/verify_backend_no_ai.ps1; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; ./scripts/verify_frontend_no_ai.ps1; exit $LASTEXITCODE }"` and `03-VERIFICATION.md` | pending |

## Human Verification

- Confirm `README.md` reads naturally as a non-AI alert system and no longer points to deleted AI design docs.
- Confirm `.planning/codebase/*.md` reflects the current repo rather than deleted AI runtime pieces.

## Exit Criteria

- All static grep checks for Phase 03-owned files pass.
- Backend and frontend verification scripts both pass.
- `03-VERIFICATION.md` records actual command execution and result summaries.
