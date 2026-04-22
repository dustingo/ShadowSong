# Phase 17: Clean Truth And Operational Docs - Context

**Gathered:** 2026-04-22
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只清理当前仍会误导维护者或新接手同学的仓库入口、运行说明和 v1.3 真相文档，并把通知可靠性与告警排障相关的现状沉淀为面向维护者的操作说明与回滚关注点。范围包含当前真源文档与低风险入口的命名/说明收口，但不包含 Go module path、内部 import 路径、数据库结构或业务主链路改造。

</domain>

<decisions>
## Implementation Decisions

### Truth Surface Scope
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

### the agent's Discretion
- 具体运维文档是落单个 runbook 还是“runbook + rollback/deferred summary”双文档结构，可由研究和计划阶段决定，只要满足维护者可直接执行。
- 低风险命名清理的最终文件清单和改名粒度，可由 planner 根据仓库扫描结果细化，只要不跨进高风险运行时契约。
- 历史归档文档中哪些地方只加 framing 注释、哪些地方需要正文改写，可由 planner 按影响面决定，但必须保持历史事实可追溯。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — 当前项目真相、Active requirement 和“只清理当前真源与低风险入口”的总约束
- `.planning/REQUIREMENTS.md` — Phase 17 对应的 `DOCS-01`、`DOCS-02`、`DOCS-03`
- `.planning/ROADMAP.md` — Phase 17 goal、plan slots 和 success criteria
- `.planning/STATE.md` — 当前已完成到 Phase 16，Phase 17 是 v1.3 的最后一个 planned phase
- `.planning/MILESTONES.md` — 已发布里程碑摘要，当前最容易暴露历史叙事混杂的维护入口之一
- `.planning/RETROSPECTIVE.md` — 里程碑复盘内容，需要与当前系统表述和历史定位保持一致

### Phase 14-16 Operational Truth
- `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md` — trace 真源、生命周期观测点与 scope 边界
- `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md` — trace/context 当前验证事实
- `.planning/phases/15-harden-notification-retry-boundaries/15-CONTEXT.md` — retry boundary、terminal failure 与短窗口重试约束
- `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md` — retry diagnostics 当前验证事实
- `.planning/phases/16-standardize-alert-path-logging/16-RESEARCH.md` — logging contract 研究结论与文档真相要求
- `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md` — parse-safe logging、`async_panic` traceability 和可复用排障证据
- `.planning/phases/16-standardize-alert-path-logging/16-SECURITY.md` — 与 logging/traceability 相关的威胁收口记录
- `.planning/phases/16-standardize-alert-path-logging/16-UAT.md` — 运维视角的 Phase 16 验证项，可作为 runbook 输入

### Current Entry Points And Low-Risk Naming Targets
- `README.md` — 当前仓库入口和开发/运行说明，需要继续弱化“AI 对照语境”并补维护者视角
- `scripts/verify_backend_no_ai.ps1` — 仍带历史命名的后端验证入口，也是低风险更名候选
- `scripts/verify_frontend_no_ai.ps1` — 前端无 AI 残留扫描入口，需要决定是保留历史含义还是升级为当前命名
- `.planning/codebase/STRUCTURE.md` — 当前代码结构图，仍显式引用 `verify_backend_no_ai.ps1` 和历史结构名
- `.planning/codebase/TESTING.md` — 当前测试/验证入口地图，仍用“non-AI verification” 描述部分真源命令

### Live Runtime And Repo Boundaries
- `README.md` §快速开始 / §模板预览与验证 / §工程质量门禁 — 维护者最先接触的运行入口
- `go.mod` — 高风险 module path 真源，本 phase 明确不改
- `internal/auth/jwt.go` — `Issuer: "ai-alert-system"` 的运行时命名边界，本 phase 明确不改
- `cmd/server/main.go` — 服务启动入口，作为“文档清理不碰运行主链路”的基准对照

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `README.md` 已有完整的启动、模板链路、验证脚本和质量门禁章节，适合在此基础上继续收口，而不是另起一份完全平行的入门文档。
- `.planning/phases/14-16/*-VERIFICATION.md`、`16-SECURITY.md`、`16-UAT.md` 已经沉淀了 trace、retry、logging 的当前行为证据，可直接抽成维护手册内容。
- `.planning/MILESTONES.md`、`.planning/PROJECT.md`、`.planning/ROADMAP.md` 已经承担不同层级的“真相文档”职责，适合通过明确分层来降低叙事混杂。

### Established Patterns
- 前几期已经明确采用 brownfield 小步增强，不做全仓高风险历史命名迁移；Phase 17 应延续“真源优先、风险分级”的策略。
- v1.2 和 v1.3 的 phase 文档已经把实现事实落在 verification/security/UAT 工件中，因此新的运维说明应围绕这些已有证据串联，而不是脱离 phase artifacts 单写。
- 仓库仍保留深层历史命名，例如 `github.com/game-ops/ai-alert-system`、`Issuer: "ai-alert-system"` 和若干测试/脚本名；此前这些被明确 deferred，说明本 phase 需要区分低风险与高风险命名。

### Integration Points
- `README.md`、`.planning/PROJECT.md`、`.planning/ROADMAP.md`、`.planning/REQUIREMENTS.md` 是当前真源叙事的主要入口。
- `.planning/MILESTONES.md` 与 `.planning/RETROSPECTIVE.md` 是最容易把“历史里程碑事实”和“当前系统现状”混写在一起的文档层。
- `scripts/verify_backend_no_ai.ps1`、`scripts/verify_frontend_no_ai.ps1`、`internal/router/router_test.go`、`internal/config/config_test.go` 等是低风险命名清理与验证入口同步更新的主要落点。

</code_context>

<specifics>
## Specific Ideas

- 默认执行方向是“当前真源 + 低风险入口一起清，完整维护手册一次补齐，高风险运行时命名继续 defer”。
- 运维文档应让维护者能从 `trace_id`、`send_attempt`、`terminal_failure`、`async_panic` 等既有观测点直接走完一条排障路径。
- 文档输出应兼顾 Phase 17 本身的收尾价值和 v1.3 里程碑归档前的复用价值，避免后面还要再补一轮同类说明。

</specifics>

<deferred>
## Deferred Ideas

- 全量迁移 `go.mod` module path 与所有 Go import 路径中的 `ai-alert-system`
- 修改 `internal/auth/jwt.go` 中的 `Issuer: "ai-alert-system"` 等可能影响运行时契约的深层命名
- 借文档清理顺带推动仓库结构大重组、目录改名或新文档站建设

None of the above belongs in Phase 17 scope.

</deferred>

---

*Phase: 17-clean-truth-and-operational-docs*
*Context gathered: 2026-04-22*
