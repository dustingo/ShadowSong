# Phase 17: Clean Truth And Operational Docs - Research

**Researched:** 2026-04-22 [VERIFIED: system date]
**Domain:** Documentation truth surfaces, low-risk naming cleanup, and maintainer-facing alert-path operations guidance. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `.planning/REQUIREMENTS.md`]
**Confidence:** MEDIUM [VERIFIED: repo scan] [ASSUMED]

<user_constraints>
## User Constraints (from CONTEXT.md)

Verbatim copy from `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]

### Locked Decisions
- **D-01:** Phase 17 先收口当前会被维护者直接读取或执行的真源入口，包括 `README.md`、`.planning/PROJECT.md`、`.planning/ROADMAP.md`、`.planning/REQUIREMENTS.md`、`.planning/MILESTONES.md`、`.planning/RETROSPECTIVE.md` 以及 v1.3 相关 phase truth artifacts。
- **D-02:** 除了真源文档之外，本 phase 也要处理低风险但对外显性的入口，例如 `scripts/verify_*_no_ai.ps1`、测试名、说明文案和工程入口中仍然把当前系统表述绑定在“AI 移除对照语境”上的内容。
- **D-03:** 历史归档文档可以继续保留 v1.0 “AI Removal Complete” 的历史事实，但必须明确它们是归档背景，不应再被当前运行说明或维护入口当作现状主叙事。

### Operational Documentation Shape
- **D-04:** Phase 17 不是只补一页极简提示，而是要补出维护者能实际使用的完整运维文档，覆盖通知失败排查、trace/logging/retry 现状、常见验证命令和回滚关注点。
- **D-05:** 运维文档应优先服务“后端主链路故障排查”，把 Phase 14/15/16 已建立的 `trace_id`、lifecycle stages、retry diagnostics 和 logging contract 串成一条可执行排障路径，而不是重新介绍产品功能。
- **D-06:** v1.3 可靠性/可观测性文档需要把“如何验证当前行为”和“回滚时最容易丢失的保证”一起写清，便于后续 milestone 收尾和人工审查直接复用。

### Naming Cleanup Boundary
- **D-07:** Phase 17 不处理高风险的 `go.mod` module path、Go import 路径、JWT issuer、数据库名或其他可能牵动运行契约的深层历史命名。
- **D-08:** 本 phase 可以处理低风险命名，包括脚本文件名、测试名、README/规划文档中的对外表述、以及不会改变运行时契约的文案型标识。
- **D-09:** 对仍需保留的历史命名，要在文档中明确标注“历史遗留/暂缓迁移”的原因，避免维护者误以为这是当前推荐命名。

### Documentation Strategy
- **D-10:** 文档清理采用“当前真源优先、归档事实保留、历史大迁移延后”的策略，不做一次性全仓文案清洗。
- **D-11:** Phase 17 需要把 v1.3 的文档真源明确分层：运行入口与维护手册写当前现状；里程碑/归档文档保留历史过程；phase-local verification/security/UAT 文档继续作为实现事实证据。
- **D-12:** 如果同一事实已在 phase verification 或 security 文档中被验证，运维说明应链接和复用这些证据，而不是再造一套脱节叙述。

### Claude's Discretion
- 具体运维文档是落单个 runbook 还是“runbook + rollback/deferred summary”双文档结构，可由研究和计划阶段决定，只要满足维护者可直接执行。
- 低风险命名清理的最终文件清单和改名粒度，可由 planner 根据仓库扫描结果细化，只要不跨进高风险运行时契约。
- 历史归档文档中哪些地方只加 framing 注释、哪些地方需要正文改写，可由 planner 按影响面决定，但必须保持历史事实可追溯。

### Deferred Ideas (OUT OF SCOPE)
- 全量迁移 `go.mod` module path 与所有 Go import 路径中的 `ai-alert-system`
- 修改 `internal/auth/jwt.go` 中的 `Issuer: "ai-alert-system"` 等可能影响运行时契约的深层命名
- 借文档清理顺带推动仓库结构大重组、目录改名或新文档站建设

None of the above belongs in Phase 17 scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DOCS-01 | 仓库入口、运行说明和里程碑文档继续保持“非 AI 告警系统”的真实命名，不引回过期描述 [VERIFIED: `.planning/REQUIREMENTS.md`] | Truth-source docs and low-risk entrypoints are now classified into safe cleanup vs deferred runtime naming, with explicit target files and no-touch boundaries. [VERIFIED: `README.md`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/MILESTONES.md`] [VERIFIED: `.planning/RETROSPECTIVE.md`] [VERIFIED: `rg scan on README/.planning/scripts/internal`] |
| DOCS-02 | 与通知可靠性、告警排障相关的文档需要补充当前链路行为、失败诊断和回滚关注点 [VERIFIED: `.planning/REQUIREMENTS.md`] | Recommended output is a maintainer runbook anchored on Phase 14/15/16 verification, UAT, and security artifacts, with rollback guarantees and deferred items documented separately from historical narrative. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`] |
| DOCS-03 | v1.3 roadmap、phase 文档和验证记录需要准确反映新的可靠性与可观测性目标，作为后续执行真源 [VERIFIED: `.planning/REQUIREMENTS.md`] | Planner should treat v1.3 phase truth artifacts as evidence sources, refresh Phase 17-local truth artifacts, and keep roadmap/project/milestone framing aligned with v1.3 reliability and observability scope. [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/STATE.md`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
</phase_requirements>

## Summary

Phase 17 should be planned as a documentation-truth and low-risk naming phase, not as a runtime rename or platform migration. The repo already has strong operational evidence from Phases 14, 15, and 16, but the maintainer-facing entrypoints still mix current-state guidance with AI-removal framing, especially around verification script names, “non-AI” wording in current docs, and codebase/testing maps that still present old naming as the active operational vocabulary. [VERIFIED: `README.md`] [VERIFIED: `.planning/codebase/STRUCTURE.md`] [VERIFIED: `.planning/codebase/TESTING.md`] [VERIFIED: `scripts/verify_backend_no_ai.ps1`] [VERIFIED: `scripts/verify_frontend_no_ai.ps1`] [VERIFIED: `rg scan on README/.planning/scripts`] 

The most useful maintainer documentation is not a product overview. It is an operational runbook that starts from observable evidence already proven in code and tests: `trace_id` continuity from `ingest` through `notification_entry`, bounded retry behavior with `send_attempt` and `terminal_failure`, and failure-path correlation on `async_panic` with parse-safe `key=value` fields. That material already exists in phase-local verification, UAT, and security artifacts and should be lifted into one stable maintainer-facing guide rather than rewritten from memory. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`]

The phase boundary is clear: rename or rewrite low-risk docs, script names, and test names, but do not touch deep runtime identity such as the Go module path or JWT issuer. Those are confirmed historical names embedded in code contracts and must remain documented as deferred work, not silently “cleaned up” inside a documentation phase. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`] [VERIFIED: `rg scan on internal/cmd/go.mod`]

**Primary recommendation:** Plan Phase 17 as two coordinated outputs: one evergreen maintainer runbook under `docs/` plus one Phase 17 truth-artifact refresh across `README.md` and `.planning/*`, while keeping high-risk runtime naming explicitly deferred. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `Get-ChildItem docs`]

## Project Constraints (from AGENTS.md)

- Keep the existing Go + Gin + GORM + PostgreSQL + Redis + React + Vite stack; this phase must not introduce a stack migration. [VERIFIED: prompt `AGENTS.md instructions`]  
- Respect the brownfield repo and unrelated uncommitted changes; do not revert or widen scope into runtime refactors. [VERIFIED: prompt `AGENTS.md instructions`] [VERIFIED: `git status --short`]  
- Preserve the core alert flow after AI removal; docs may clarify runtime behavior but must not change ingress, display, routing, silence, or on-duty behavior. [VERIFIED: prompt `AGENTS.md instructions`]  
- Frontend-facing names, routes, menus, types, and API references must stay self-consistent when low-risk naming is changed. [VERIFIED: prompt `AGENTS.md instructions`]  
- Project naming in README, page titles, and test wording should reflect the non-AI alert-system reality, but deep runtime naming is explicitly out of scope for this phase. [VERIFIED: prompt `AGENTS.md instructions`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]

## Standard Stack

No new runtime or library dependency is needed for Phase 17. The standard implementation stack is the repo’s existing Markdown truth surfaces, PowerShell verification entrypoints, and Go test evidence. [VERIFIED: `.planning/config.json`] [VERIFIED: `README.md`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/ROADMAP.md`]

### Core
| Library / Tool | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Markdown docs in repo truth surfaces | repo-native | Update current entrypoint truth and maintainer instructions without changing runtime code. [VERIFIED: `README.md`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/ROADMAP.md`] | Existing planning workflow already treats these files as canonical truth. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| PowerShell verification scripts | repo-native | Keep low-risk executable verification entrypoints aligned with current naming and docs. [VERIFIED: `scripts/verify_backend_no_ai.ps1`] [VERIFIED: `scripts/verify_frontend_no_ai.ps1`] | They are already the maintainer-visible scripted verification surface. [VERIFIED: `README.md`] [VERIFIED: `.planning/codebase/STRUCTURE.md`] [VERIFIED: `.planning/codebase/TESTING.md`] |
| Go `go test` regressions | Go 1.25.0 in repo | Provide the reusable evidence behind trace, retry, and logging claims. [VERIFIED: `go.mod`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] | Existing phase truth already references these commands and planner should reuse that evidence instead of inventing manual-only claims. [VERIFIED: phase 14/15/16 verification docs] |

### Supporting
| Library / Tool | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `rg` / repo scan | local CLI | Find stale naming in docs, tests, scripts, and codebase maps. [VERIFIED: repo scans in this research] | Use for low-risk rename inventory and for keeping deferred runtime naming out of scope. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| `docs/` supplementary docs folder | repo-native | Hold the evergreen maintainer runbook without bloating README. [VERIFIED: `Get-ChildItem docs`] | Use for operational guidance that should outlive the phase-local artifact. [ASSUMED] |
| Phase-local verification/security/UAT artifacts | repo-native | Supply exact commands, field names, rollback guarantees, and deferred limitations. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`] | Use whenever docs need current operational truth, not historical narration. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Evergreen runbook in `docs/` plus short README links | Put all operational detail directly into `README.md` | This would overload the primary entrypoint and make rollback/deferred sections harder to maintain. `README.md` is better as an entry surface and pointer, not the full troubleshooting manual. [VERIFIED: `README.md` section layout] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| Phase-local truth refresh plus targeted current docs updates | Repo-wide text cleanup of all historical AI mentions | Context explicitly forbids a global migration and preserves archived historical facts. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |

**Installation:** No package installation is required for this phase. [VERIFIED: `.planning/config.json`] [VERIFIED: repo scan]

## Architecture Patterns

### Recommended Documentation Structure
```text
README.md                              # current entrypoint, quickstart, links to maintainer docs
docs/
└── alert-operations-runbook.md        # evergreen maintainer troubleshooting + rollback guide
.planning/
├── PROJECT.md                         # current project truth
├── ROADMAP.md                         # milestone/phase truth
├── MILESTONES.md                      # shipped milestone summaries with archive framing
├── RETROSPECTIVE.md                   # historical learning, explicitly framed as retrospective
└── phases/17-clean-truth-and-operational-docs/
    ├── 17-CONTEXT.md                  # locked decisions
    ├── 17-RESEARCH.md                 # this research
    └── 17-VERIFICATION.md             # phase-local proof that the docs and naming were aligned
```
[VERIFIED: `README.md`] [VERIFIED: `Get-ChildItem docs`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/MILESTONES.md`] [VERIFIED: `.planning/RETROSPECTIVE.md`] [ASSUMED]

### Pattern 1: Truth-Surface Layering
**What:** Keep current-state operational truth in `README.md`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, and the new maintainer runbook, while leaving historical facts in milestone archives and retrospective material. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/MILESTONES.md`] [VERIFIED: `.planning/RETROSPECTIVE.md`]
**When to use:** Use this whenever a doc is likely to be a maintainer’s first stop or a planner’s canonical source. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [ASSUMED]
**Example:**
```markdown
## 维护者入口

- 当前运行与验证入口见本文。
- 通知失败排查、trace/logging/retry 说明见 `docs/alert-operations-runbook.md`。
- 历史 AI 移除背景仅见 `.planning/milestones/` 与 retrospective 文档，不作为当前运行说明。
```
Source pattern derived from current truth-source separation requirements. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]

### Pattern 2: Evidence-First Operational Writing
**What:** Document operator actions by starting from concrete searchable fields and commands already proven in phase verification artifacts, not from free-form prose. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`]
**When to use:** Use this for retry diagnostics, trace lookups, logging field contracts, rollback notes, and maintainer checklists. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]
**Example:**
```markdown
1. 从 `stage=terminal_failure`、`stage=send_attempt` 或 `stage=async_panic` 开始检索。
2. 记录同一行的 `trace_id`、`alert_id`、`channel_id`、`attempt`、`max_attempts`。
3. 用相同 `trace_id` 回查 `notification_entry`、`route_match`、`redis_publish`、`persist`、`ingest`。
4. 如果只有 `send_attempt` 没有 `terminal_failure`，说明故障可能在重试窗口内恢复。 [ASSUMED]
```
Source pattern derived from verified trace/retry/logging evidence. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`]

### Pattern 3: Safe-Rename Boundary Table
**What:** Every rename candidate should be classified as either low-risk documentation/entrypoint cleanup or deferred runtime naming. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `rg scan on scripts/internal/go.mod`]
**When to use:** Use this before any file rename, test rename, or wording cleanup. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]
**Example:**
```markdown
| Target | Risk | Action |
|--------|------|--------|
| `scripts/verify_backend_no_ai.ps1` | low | rename/update references in same phase |
| `internal/router/router_test.go::TestSetup_RoutesWithoutAIRuntime` | low | rename test for current truth |
| `go.mod` module path | high | defer and document |
| `internal/auth/jwt.go` issuer | high | defer and document |
```
Source pattern derived from current scan and locked scope boundary. [VERIFIED: `scripts/verify_backend_no_ai.ps1`] [VERIFIED: `internal/router/router_test.go`] [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]

### Anti-Patterns to Avoid
- **Treating historical archive language as current-state truth:** `v1.0 AI Removal Complete` is valid milestone history, but it should not dominate current run instructions or maintainer entrypoints. [VERIFIED: `.planning/MILESTONES.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]
- **Cleaning runtime identity under a docs banner:** changing `go.mod` or JWT issuer here would exceed the phase boundary and create contract risk. [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]
- **Writing new operational prose without evidence links:** the repo already has command-backed verification truth for Phases 14-16, so planner should reuse it. [VERIFIED: phase 14/15/16 verification docs]
- **Keeping misleading supplemental docs unframed:** `docs/CODE_REVIEW.md` includes outdated and now-conflicting guidance such as recommending repo-wide `slog` migration despite Phase 16 explicitly not doing that. [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-01-PLAN.md`]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Maintainer troubleshooting narrative | A new prose-only explanation disconnected from code/tests | Reuse Phase 14/15/16 verification, UAT, and security artifacts as evidence sources | The repo already proves trace continuity, retry behavior, and logging contract with commands and field-level assertions. [VERIFIED: phase 14/15/16 verification docs] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`] |
| Truth-source cleanup | Global grep-and-replace of every historical AI mention | Target current truth surfaces and low-risk entrypoints only | Context explicitly preserves archive history and defers deep runtime naming. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| Rollback guidance | A hypothetical rollback design | Pull rollback concerns from guarantees already established in Phase 14-16 | Operators need to know what guarantees would be lost, not an invented rollback system. [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`] |
| Script rename risk handling | Silent file rename with no reference audit | Rename only after updating all in-repo references in README/planning/codebase maps/tests | The current repo still references `verify_backend_no_ai.ps1` and `verify_frontend_no_ai.ps1` in multiple maintainer-visible surfaces. [VERIFIED: `rg scan on README/.planning/scripts`] |

**Key insight:** Phase 17 should package already-verified truth, not invent new behavior or broaden scope into runtime migration. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `.planning/REQUIREMENTS.md`]

## Runtime State Inventory

This phase includes low-risk rename/cleanup work, so runtime-state categories were checked explicitly. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | No repo-visible database schema, model field, or migration key was found using `ai-alert-system` or `no_ai` as a persisted data key; current relevant persisted observability data is `trace_id` on alerts, which is part of the current contract and not a rename target. [VERIFIED: `internal/models/alert.go`] [VERIFIED: `rg scan on internal/cmd/go.mod`] | None for Phase 17 data migration. Keep runtime-name changes out of scope and document that stored alert data is not being renamed. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| Live service config | No repo-tracked GitHub workflow, compose file, or local docs reference was found that binds CI/runtime behavior to `verify_*_no_ai.ps1`; current references are documentation and codebase-map surfaces, not live deployment configuration. [VERIFIED: `.github/workflows` scan] [VERIFIED: `docker-compose.yml` scan] [VERIFIED: `rg scan on README/.planning/scripts`] | Update in-repo references only. If maintainers run these scripts from private automation outside git, that dependency must be confirmed before renaming file paths. [ASSUMED] |
| OS-registered state | `schtasks /query /fo LIST /v` scan found no scheduled-task registrations containing `ai-alert-system`, `shadowsongAI`, `verify_backend_no_ai`, or `verify_frontend_no_ai`. [VERIFIED: `schtasks scan 2026-04-22`] | None found on this machine. No OS re-registration task is required based on current evidence. [VERIFIED: `schtasks scan 2026-04-22`] |
| Secrets/env vars | Current environment scan did not reveal project env vars using the old naming, and repo config continues to require `JWT_SECRET`, `DB_*`, `REDIS_*`, and server vars rather than AI-specific runtime keys. `internal/config/config_test.go` still contains AI-env absence tests, but those are low-risk test naming candidates, not active runtime config. [VERIFIED: `Get-ChildItem Env:` scan 2026-04-22] [VERIFIED: `internal/config/config.go`] [VERIFIED: `internal/config/config_test.go`] | No secret migration identified. Rename test names and docs if desired, but do not rename real runtime env keys in this phase. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| Build artifacts | The only repo-visible filename artifacts carrying old naming are `scripts/verify_backend_no_ai.ps1` and `scripts/verify_frontend_no_ai.ps1`; no `.egg-info`, compiled binary, or other installed artifact carrying that name was found by file scan. [VERIFIED: `rg --files | rg \"no_ai|ai-alert-system|egg-info|dist/|build/|bin/|\\.exe$\"`] | Safe file rename candidate after reference audit. No package reinstall or artifact migration is currently indicated. [VERIFIED: `rg --files` scan] |

## Common Pitfalls

### Pitfall 1: Updating README But Leaving Planning Truth Behind
**What goes wrong:** The public entrypoint looks current, but `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/MILESTONES.md`, or `.planning/RETROSPECTIVE.md` still present mixed or stale framing. [VERIFIED: `README.md`] [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/MILESTONES.md`] [VERIFIED: `.planning/RETROSPECTIVE.md`]
**Why it happens:** These files have different roles and were updated across multiple milestones, so wording drift accumulates even when the code is current. [VERIFIED: phase and milestone doc timestamps/content] [ASSUMED]
**How to avoid:** Plan one explicit “truth-surface sweep” task across all current canonical docs, not just README. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]
**Warning signs:** `README.md` uses current framing while codebase maps or planning docs still advertise “non-AI verification” as the active operational identity. [VERIFIED: `README.md`] [VERIFIED: `.planning/codebase/STRUCTURE.md`] [VERIFIED: `.planning/codebase/TESTING.md`]

### Pitfall 2: Renaming Low-Risk Scripts Without Updating Their Entire Narrative Surface
**What goes wrong:** The file name changes, but README, testing maps, structure maps, or human instructions still point to the old script path. [VERIFIED: `README.md`] [VERIFIED: `.planning/codebase/STRUCTURE.md`] [VERIFIED: `.planning/codebase/TESTING.md`]
**Why it happens:** Script names are referenced in docs rather than in code wiring, so link drift is easy to miss. [VERIFIED: `rg scan on README/.planning/scripts`] [ASSUMED]
**How to avoid:** Treat script rename as a reference-audit task with mandatory `rg` verification before closing the plan. [VERIFIED: repo scan results]
**Warning signs:** `rg -n "verify_backend_no_ai|verify_frontend_no_ai" README.md .planning scripts` still returns mixed old/new paths after the rename. [VERIFIED: current `rg` scan]

### Pitfall 3: Treating Deep Runtime Naming As “Just Docs”
**What goes wrong:** Planner accidentally reaches into `go.mod`, import paths, JWT issuer, or other runtime contracts under the banner of “truth cleanup.” [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`]
**Why it happens:** The repo still visibly contains `ai-alert-system` in code, which can look like the same class of cleanup as script/test naming. [VERIFIED: `rg scan on internal/cmd/go.mod`] 
**How to avoid:** Keep a hard “deferred runtime naming” table in the new docs and in Phase 17 verification. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [ASSUMED]
**Warning signs:** Proposed edits include `module github.com/game-ops/ai-alert-system` or `Issuer: "ai-alert-system"`. [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`]

### Pitfall 4: Writing Operational Docs As Product Docs
**What goes wrong:** The doc explains features, but not how to debug a notification failure or what evidence to gather before rollback. [VERIFIED: phase requirements DOCS-02] [VERIFIED: `.planning/REQUIREMENTS.md`]
**Why it happens:** README-style writing defaults to setup and overview, while Phase 17 needs maintainer action paths. [VERIFIED: `README.md` section layout] [ASSUMED]
**How to avoid:** Organize the runbook around symptoms and search keys: `terminal_failure`, `send_attempt`, `async_panic`, `trace_id`, and upstream lifecycle stages. [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`]
**Warning signs:** The doc has no section on rollback concerns, no command list, and no trace-based search recipe. [VERIFIED: Phase 17 context requirements] [ASSUMED]

### Pitfall 5: Letting Supplemental Docs Contradict Current Truth
**What goes wrong:** A supplementary doc like `docs/CODE_REVIEW.md` keeps recommending actions that current phase decisions explicitly rejected or superseded. [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-01-PLAN.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`]
**Why it happens:** Supplemental docs are easier to forget because they are not always part of phase truth checks. [VERIFIED: `Get-ChildItem docs`] [ASSUMED]
**How to avoid:** Either refresh the doc to current truth, or clearly mark it as historical/advisory so it is not mistaken for active guidance. [VERIFIED: docs inventory] [ASSUMED]
**Warning signs:** The doc recommends repo-wide `slog` migration while Phase 16 explicitly kept the existing `log.Logger` seam and canonical `key=value` output. [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-01-PLAN.md`]

## Code Examples

Verified patterns from current repo truth:

### Current Maintainer Troubleshooting Flow
```markdown
1. Search logs for `stage=terminal_failure`, `stage=send_attempt`, or `stage=async_panic`.
2. Copy the `trace_id` from that line.
3. Search the same `trace_id` across `notification_entry`, `route_match`, `redis_publish`, `persist`, and `ingest`.
4. Compare `attempt` and `max_attempts` to decide whether the notification exhausted retries.
5. If rollback is being considered, confirm which guarantee would be lost:
   - bounded retries
   - explicit `terminal_failure`
   - parse-safe canonical fields
   - `async_panic` correlation
```
Source: Phase 14/15/16 evidence path. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-UAT.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`]

### Safe Rename Audit Command
```powershell
rg -n "verify_backend_no_ai|verify_frontend_no_ai|WithoutAI|WithoutAIRuntime|non-AI|无 AI" `
  README.md .planning scripts internal
```
Source: current repo scan method used in this research. [VERIFIED: repo scan commands 2026-04-22]

### Deferred Runtime Naming Table Pattern
```markdown
## Deferred Runtime Naming

| Item | Why Deferred |
|------|--------------|
| `go.mod` module path | import path and build contract risk |
| JWT issuer | token compatibility / runtime contract risk |
| deep import paths | broad refactor outside Phase 17 |
```
Source: locked phase boundary. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| “Non-AI” framing used as a continuing present-tense operational identity | Current docs should describe the alert system directly and reserve AI-removal language for historical archive context | v1.0 established the historical event; Phase 12 and Phase 17 continue current-truth cleanup. [VERIFIED: `.planning/MILESTONES.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/phases/12-establish-automated-quality-gates/12-CONTEXT.md`] | Maintainers stop reading the system primarily as a migration aftermath and start from the current operational baseline. [ASSUMED] |
| Free-text or scattered guidance for notification failures | Trace-based diagnosis anchored on `trace_id`, lifecycle stages, retry fields, and canonical logging contract | Established across Phases 14-16. [VERIFIED: phase 14/15/16 verification docs] | Phase 17 can produce a practical runbook without inventing new observability infrastructure. [VERIFIED: `.planning/REQUIREMENTS.md`] |
| Supplemental docs may drift independently | Supplemental docs should either be refreshed or clearly demoted if they contradict current truth | Needed now because `docs/CODE_REVIEW.md` includes outdated logging guidance. [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] | Reduces maintainer confusion about current logging and rollback expectations. [ASSUMED] |

**Deprecated/outdated:**
- Treating `scripts/verify_backend_no_ai.ps1` and `scripts/verify_frontend_no_ai.ps1` names as the preferred long-term maintainer vocabulary is outdated even if the scripts themselves still work today. [VERIFIED: `scripts/verify_backend_no_ai.ps1`] [VERIFIED: `scripts/verify_frontend_no_ai.ps1`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [ASSUMED]
- Treating repo-wide `slog` migration as the active recommendation for alert-path logging is outdated for current truth because Phase 16 explicitly kept the `log.Logger` seam and canonical text contract. [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-01-PLAN.md`] 

## Open Inventory: Safe Cleanup vs Deferred Runtime Naming

### Safe Low-Risk Cleanup Targets
- `README.md` current framing plus verification-entrypoint wording. [VERIFIED: `README.md`]
- `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`, `.planning/MILESTONES.md`, `.planning/RETROSPECTIVE.md` current-state wording and archive framing. [VERIFIED: corresponding files]
- `scripts/verify_backend_no_ai.ps1` and `scripts/verify_frontend_no_ai.ps1` file names and user-facing strings, after full in-repo reference updates. [VERIFIED: script files] [VERIFIED: `rg scan on README/.planning/scripts`]
- `internal/router/router_test.go::TestSetup_RoutesWithoutAIRuntime` and `internal/config/config_test.go::TestLoad_WithoutAIEnv` test naming. [VERIFIED: `internal/router/router_test.go`] [VERIFIED: `internal/config/config_test.go`]
- `.planning/codebase/STRUCTURE.md` and `.planning/codebase/TESTING.md` references to “non-AI verification” script names and current truth narrative. [VERIFIED: `.planning/codebase/STRUCTURE.md`] [VERIFIED: `.planning/codebase/TESTING.md`]
- `docs/CODE_REVIEW.md` framing or status marker because it currently contradicts current logging decisions. [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`]

### High-Risk Runtime Naming To Keep Deferred
- `go.mod` module path `github.com/game-ops/ai-alert-system`. [VERIFIED: `go.mod`]
- Go import paths under `internal/` and `cmd/` using that module path. [VERIFIED: `rg scan on internal/cmd/go.mod`]
- JWT issuer `ai-alert-system` in `internal/auth/jwt.go`. [VERIFIED: `internal/auth/jwt.go`]
- Any broader package/directory rename implied by a module-path migration. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [ASSUMED]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `docs/` is the best long-lived location for the maintainer runbook because it currently has low collision and will not confuse planner truth surfaces. [ASSUMED] | Architecture Patterns / Standard Stack | If wrong, planner may place the runbook in the wrong location and create another truth split. |
| A2 | Some maintainers may have out-of-repo automation or habits tied to the current `verify_*_no_ai.ps1` paths, so file renames should be treated as low-risk but still reference-audited. [ASSUMED] | Runtime State Inventory / Common Pitfalls | If wrong, planner may over-engineer compatibility or, conversely, break an undocumented local workflow. |
| A3 | Supplemental docs like `docs/CODE_REVIEW.md` are maintainer-visible enough that they should be refreshed or demoted during this phase. [ASSUMED] | Common Pitfalls / Safe Cleanup Inventory | If wrong, planner might spend effort on a file that is not actually part of the maintainer truth surface. |

## Open Questions

1. **Should Phase 17 rename the verification script files themselves, or only their displayed descriptions?** [VERIFIED: `scripts/verify_backend_no_ai.ps1`] [VERIFIED: `scripts/verify_frontend_no_ai.ps1`]
What we know: The file names are explicitly listed as low-risk rename candidates in context, and all current repo-visible references are doc/map references rather than runtime wiring. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `rg scan on README/.planning/scripts`]
What's unclear: Whether external human habits or private automation depend on the existing paths. [ASSUMED]
Recommendation: Planner should either rename files and update all in-repo references in the same wave, or keep file paths for now but rewrite visible descriptions and add a deferred note. [ASSUMED]

2. **Should `docs/CODE_REVIEW.md` be updated, archived, or left alone?** [VERIFIED: `docs/CODE_REVIEW.md`]
What we know: It is the only file currently under `docs/`, and it contains advice that conflicts with current Phase 16 truth, especially around repo-wide `slog` migration. [VERIFIED: `Get-ChildItem docs`] [VERIFIED: `docs/CODE_REVIEW.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`]
What's unclear: Whether the team considers it active guidance or historical reference. [ASSUMED]
Recommendation: Planner should decide explicitly instead of ignoring it; either refresh it, add a historical banner, or exclude it from current maintainer truth with a clear note. [ASSUMED]

## Security Domain

This phase is documentation-heavy, but it still touches security-relevant operator guidance because stale or contradictory docs can cause incorrect rollback, evidence loss, or accidental disclosure in examples. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`] [ASSUMED]

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Phase 17 should document, not modify, existing auth behavior. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| V3 Session Management | no | No session-contract change is in scope. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| V4 Access Control | no | This phase does not alter permission boundaries. [VERIFIED: `.planning/PROJECT.md`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| V5 Input Validation | no | Documentation updates do not add new user-input processing paths. [VERIFIED: phase scope docs] |
| V6 Cryptography | no | No cryptographic implementation or secret-handling change is planned. [VERIFIED: phase scope docs] |

### Known Threat Patterns for This Phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Maintainers follow stale docs and miss current rollback guarantees | Tampering / Repudiation | Write the runbook from current phase verification artifacts and link exact evidence. [VERIFIED: phase 14/15/16 verification docs] |
| Docs accidentally normalize high-risk runtime renames as safe cleanup | Denial of Service / Tampering | Keep an explicit deferred-runtime-naming table with `go.mod` and JWT issuer listed as no-touch items. [VERIFIED: `go.mod`] [VERIFIED: `internal/auth/jwt.go`] [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] |
| Operational examples leak too much data from logs | Information Disclosure | Reuse existing Phase 16 contract examples limited to IDs, routing metadata, retry fields, mode, and bounded error strings; do not paste raw payloads or secrets. [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md`] [VERIFIED: `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`] |

## Sources

### Primary (HIGH confidence)
- `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md` - locked decisions, scope boundary, target surfaces, and operational-doc expectations. [VERIFIED: file read]
- `.planning/REQUIREMENTS.md` - `DOCS-01`, `DOCS-02`, `DOCS-03` and v1.3 requirement mapping. [VERIFIED: file read]
- `.planning/STATE.md` - current milestone position and readiness for Phase 17 planning. [VERIFIED: file read]
- `README.md` - current entrypoint, quickstart, and verification entry wording. [VERIFIED: file read]
- `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/MILESTONES.md`, `.planning/RETROSPECTIVE.md` - current truth and archive framing surfaces. [VERIFIED: file read]
- `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md` - trace evidence and lifecycle stages. [VERIFIED: file read]
- `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md` - retry boundary, terminal failure, and command evidence. [VERIFIED: file read]
- `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`, `16-UAT.md`, `16-SECURITY.md` - current logging contract, operator tests, and rollback/security implications. [VERIFIED: file read]
- `go.mod` and `internal/auth/jwt.go` - deferred high-risk runtime naming facts. [VERIFIED: file read]

### Secondary (MEDIUM confidence)
- `docs/CODE_REVIEW.md` - supplemental maintainer-visible doc with outdated logging guidance. [VERIFIED: file read] [ASSUMED]
- `rg` scans across `README.md`, `.planning/`, `scripts/`, `internal/`, `.github/workflows`, and `docs/` - inventory of stale naming, live references, and no-touch runtime names. [VERIFIED: repo scans 2026-04-22]
- `schtasks /query` and `Get-ChildItem Env:` scans - runtime-state spot checks on this machine. [VERIFIED: shell scans 2026-04-22]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new dependency decision is needed and current implementation surfaces are repo-verified. [VERIFIED: `.planning/config.json`] [VERIFIED: repo files]
- Architecture: HIGH - the recommended doc layering follows explicit Phase 17 decisions and current truth-source layout. [VERIFIED: `.planning/phases/17-clean-truth-and-operational-docs/17-CONTEXT.md`] [VERIFIED: `README.md`] [VERIFIED: `.planning/*`]
- Pitfalls: MEDIUM - the repo evidence is strong, but some rename-impact and supplemental-doc visibility questions depend on human workflow outside git. [VERIFIED: repo scans] [ASSUMED]

**Research date:** 2026-04-22 [VERIFIED: system date]  
**Valid until:** 2026-05-22 for repo-local truth surfaces, or earlier if the team changes Phase 17 scope or performs external automation renames not visible in git. [VERIFIED: repo-local nature of sources] [ASSUMED]
