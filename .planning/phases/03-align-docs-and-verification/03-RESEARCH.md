---
phase: 03-align-docs-and-verification
researched: 2026-04-09
status: complete
source: inline-fallback
---

# Phase 03 Research: Align Docs And Verification

## Goal

为 Phase 3 规划提供实现边界与风险输入：清理剩余 AI 文案/配置引用、处理仍残留的 AI-only schema 或 data 痕迹，并把“非 AI 版本”的验证路径收敛成可执行产物。

## Current State

- Phase 1 已移除后端 AI runtime、AI 路由和主要模型/迁移残留，并已有 `scripts/verify_backend_no_ai.ps1` 与对应 verification 报告。
- Phase 2 已移除前端 AI 页面、导航入口、共享 API/types 里的 AI 合同，并通过 `pnpm build` 验证。
- Phase 3 目录此前不存在，说明文档/配置/验证收尾尚未拆成可执行 plans。

## Findings

### 1. User-facing docs and naming still expose AI branding

Observed files:
- `README.md`
  - Title still `游戏运维 AI 告警系统`
  - Intro still describes the product as an AI alert system
  - Structure tree still lists `internal/ai/`
  - API doc reference still points to `.kiro/specs/ai-alert-system/design.md`
- `docs/CODE_REVIEW.md`
  - Header still says `AI Alert System (shadowsongAI)`
  - Contains feature summary line `AI 集成: AI 分析和建议功能完整`
- `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/CONCERNS.md`, likely other codebase map files
  - Still describe deleted AI handlers, models, flows, and frontend pages as if they are active

Implication:
- DATA-01 cannot be satisfied until at least README and user-facing documentation are updated.
- If Phase 3 leaves `.planning/codebase/*` stale, future GSD runs will plan against incorrect architecture facts.

### 2. Runtime config/examples still require or advertise AI settings

Observed files:
- `.env`
  - Contains live-looking `OPENAI_*`, `AI_MODEL`, `AI_TIMEOUT`
  - `DB_NAME=ai_alert_system`
- `README.md`
  - Setup section still implies copying an example env file, but current repo state should no longer require AI vars for normal setup

Implication:
- DATA-01 likely needs a scoped env/config cleanup task.
- Because `.env` is a local file and may contain user-specific secrets, Phase 3 should treat it carefully: update only if roadmap/project intent expects local repo truth to reflect non-AI setup, and avoid broad secret churn outside the documented AI-specific keys and naming.

### 3. Some residual AI/data references are now historical or tooling-only

Observed files:
- `.planning/codebase/ARCHITECTURE.md`
  - Still references `internal/ai/client.go`, `AIHandler`, `AILog`, `SilenceRecommendation`, `frontend/src/pages/AIAssistant.tsx`
- `.planning/codebase/CONCERNS.md`
  - Still mentions AI logging / suggestions concerns
- `AGENTS.md`
  - Project instructions intentionally preserve historical stack/context including removed AI paths and env vars

Implication:
- There are two categories:
  1. **Must clean in Phase 3:** files that describe the current product/runtime and will mislead users or future planning (`README.md`, codebase maps, docs reports, config examples, page titles/test copy if any remain).
  2. **Probably preserve:** instruction/history files whose purpose is to explain the removal program itself (for example `AGENTS.md`, `PROJECT.md`, phase reports). These mention AI because the project goal is removing it.
- DATA-02 should focus on active code/runtime/data references, not rewriting historical planning evidence.

### 4. Verification gaps are mostly packaging/documentation gaps now

Reliable existing verification artifacts:
- Backend:
  - `scripts/verify_backend_no_ai.ps1`
  - `.planning/phases/01-remove-backend-ai-runtime/01-VERIFICATION.md`
- Frontend:
  - `pnpm build` in `frontend/`
  - `.planning/phases/02-remove-frontend-ai-surfaces/02-VERIFICATION.md`

Observed weakness:
- Frontend verification exists operationally but is not yet formalized as a dedicated Phase 3 verification asset/command doc.
- No frontend automated test runner exists; current strongest proof remains build + grep/report.

Implication:
- VER-02 can likely be satisfied without introducing a new test framework, by codifying a repeatable frontend verification path (for example script or documented command + expected output) and running it.
- Phase 3 should avoid expanding into large frontend test infrastructure unless the planner finds a tiny, low-risk path.

### 5. Natural execution slices have low overlap

Recommended slices:

1. **03-01 Docs / naming / config references**
   - `README.md`
   - `.env` or env-related docs
   - `docs/CODE_REVIEW.md`
   - possibly other user-facing copy docs

2. **03-02 Codebase map / residual AI-only schema-data references**
   - `.planning/codebase/ARCHITECTURE.md`
   - `.planning/codebase/CONCERNS.md`
   - possibly `STACK.md`, `STRUCTURE.md`, `INTEGRATIONS.md`, `CONVENTIONS.md` if they still claim AI runtime is present
   - any remaining active code/data references discovered by targeted grep

3. **03-03 Verification packaging and execution**
   - add/update a frontend verification script or documented runner
   - possibly refresh backend verification docs/commands for the final non-AI milestone state
   - produce Phase 3 verification evidence

This split keeps docs/config, architecture-map cleanup, and verification execution mostly separate.

## Risks

- `.env` is a real local file, not just an example. Editing it must stay tightly scoped to AI-only keys and AI-branded DB naming if the phase chooses to normalize it.
- `.planning/codebase/*` files are inputs to future planning; stale content there is high leverage and should be treated as productively important docs, not optional cleanup.
- `AGENTS.md` contains deliberate historical references to removed AI code paths. Treat it as instruction context unless the plan explicitly decides otherwise.
- Existing unrelated user changes remain in the worktree, so plans should use narrow file ownership and avoid broad scripted rewrites.

## Recommended Planning Direction

- Keep Phase 3 to three plans aligned with roadmap bullets:
  - `03-01` for README / naming / env-config references
  - `03-02` for codebase-map cleanup plus any remaining AI-only schema/data reference isolation
  - `03-03` for final non-AI verification path(s) and execution evidence
- Require every plan to preserve historical phase reports and project-goal documents that intentionally mention AI removal.
- Prefer grep-verifiable acceptance criteria and concrete commands because this phase is mostly doc/config/verification work.

## Validation Architecture

For this phase, the strongest automated checks are:

- Targeted grep over docs/config/codebase-map files for `AI`, `OPENAI`, `internal/ai`, `AIAssistant`, `ai_alert_system`, and stale path references where those strings are supposed to be removed.
- Backend verification via `scripts/verify_backend_no_ai.ps1` or a documented subset if full rerun is too heavy.
- Frontend verification via `pnpm build` in `frontend/`.
- Final phase verification should cross-check that:
  - user-facing docs no longer market the product as AI,
  - setup/config docs no longer require AI vars for normal use,
  - active codebase maps no longer describe deleted AI runtime pieces as current architecture,
  - backend and frontend each have at least one explicit non-AI verification path.
