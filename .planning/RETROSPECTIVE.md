# Retrospective

本文件记录已发版里程碑的历史复盘，不是当前运行手册。当前维护者应先看 `README.md`、`.planning/PROJECT.md` 与 Phase 14-17 的 verification/runbook 真源；这里保留 `v1.0 AI Removal Complete` 作为 archive history。

## Milestone: v1.0 — AI Removal Complete

**Shipped:** 2026-04-10  
**Phases:** 4 | **Plans:** 12

### What Was Built

- 移除了后端 AI 配置、路由、运行时文件和持久化残留
- 移除了前端 AI 页面、入口、展示字段和调用链
- 对齐 README、环境基线、codebase map 和验证入口到非 AI 当前态
- 补齐了通知模板原始事件透传、产品内预览和端到端验证脚本

### What Worked

- 先拆除 AI 主链路，再清理文档和验证，回归面更可控
- PowerShell 自清理验证脚本非常适合当前 Windows/Go/React 工作流
- 模板契约和 UI 预览一起交付，避免功能可用但用户不会用

### What Was Inefficient

- `REQUIREMENTS.md` 在执行中没有持续同步，归档时仍保留旧状态
- `STATE.md` 的部分字段格式和 gsd-tools 预期不一致，需要手动收口
- 阶段结束后才做 milestone audit，会让里程碑归档前的状态修正成本变高

### Patterns Established

- 废弃能力下线后必须补脚本化验证，而不是只依赖代码审查
- 模板系统改动应同时交付共享渲染契约、UI 指导和端到端验证
- dirty worktree 下始终按文件精确提交，避免吞掉用户本地改动

### Key Lessons

- 用户能直接感知的字段契约必须显式命名，不能让用户猜内部映射
- 预览接口必须复用 live 渲染路径，否则说明文档会快速漂移
- 里程碑结束前需要一次 requirements/state 对账，否则归档文件会落后于真实执行结果

### Cost Observations

- 主要工作集中在 docs / fix / test 提交，阶段粒度清晰
- 真实验证脚本虽然耗时较长，但能有效提前暴露契约和路由问题

## Milestone: v1.3 — Notification Reliability and Observability

**Shipped:** 2026-04-29  
**Phases:** 4 | **Plans:** 11

### What Was Built

- 建立了从 webhook `ingest` 到通知发送失败路径的 `trace_id` 关联真源
- 为通知发送补齐了 bounded retry、attempt-level diagnostics 和 `terminal_failure` 落点
- 统一了 webhook alert-path 的 canonical `key=value` 日志契约，并修复空格值的可解析性
- 交付了维护者 alert-path runbook、Phase 17 truth artifacts，以及 v1.3 里程碑审计归档

### What Worked

- 先补 trace 真源，再补 retry 边界，最后做日志契约统一，这个 phase 顺序是对的
- 将 failure-path 约束写成 focused Go tests，比只写 verification 文档更稳
- 保持 brownfield 增量策略，没有引入新基础设施，也把主链路质量明显拉高

### What Was Inefficient

- 里程碑审计仍然是在 phase 全部完成后补做，closeout 时要额外对账 requirements、summaries 和 verification
- `gsd-tools milestone complete` 只能完成部分 closeout，`ROADMAP.md`、`PROJECT.md` 和 `STATE.md` 仍需要人工收口
- dirty worktree 下做 closeout 需要额外约束 staging 范围，否则很容易误带无关 planning 删除或本地实验改动

### Patterns Established

- 任何可靠性增强如果要可维护，必须同步交付 runbook，而不是只交付日志
- 文本型日志契约可以继续用，但必须配套 parse-safe encoding 和 field-level regressions
- 里程碑 closeout 前先做 milestone audit，可以显著降低归档时的文档修补成本

### Key Lessons

- 没有 durable retry 基础设施时，至少要把失败路径的 final landing zone 和 correlation story 做实
- 真源文档和历史归档必须分层，否则维护者会把旧 milestone 叙事误读成当前入口
- 对于历史运行时命名，先明确 deferred boundary，比在文档里模糊带过更安全

### Cost Observations

- 本轮以 docs / test / fix 混合提交为主，phase 颗粒度适合 brownfield 连续增强
- focused handler/notifier 回归测试成本低、回报高，是后续类似 phase 的默认验证方式

## Cross-Milestone Trends

- v1.0：以“先收主链路、再补证据”的顺序最稳
- v1.3：trace truth → retry boundary → log contract → maintainer runbook 的链式推进，比直接上“大而全 observability 改造”更适合当前仓库
- 当前维护叙事已转入可靠性、可观测性和文档真相分层；历史 AI 移除表述只作为复盘背景，不应覆盖当下的维护者入口
