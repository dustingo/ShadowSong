# Phase 17: Clean Truth And Operational Docs - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-22
**Phase:** 17-clean-truth-and-operational-docs
**Areas discussed:** Truth Surface Scope, Operational Documentation Shape, Naming Cleanup Boundary

---

## Truth Surface Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Current truth only | 只收口当前真源文档，如 `README.md`、`.planning/PROJECT.md`、`.planning/ROADMAP.md` 和 v1.3 phase artifacts | |
| Current truth + low-risk entry points | 除真源文档外，也处理低风险脚本、测试名和对外说明里仍会误导维护者的入口 | ✓ |
| Broad historical sweep | 把更多历史入口与旧文案一并做较大范围清理 | |

**User's choice:** 采用推荐默认方案：当前真源文档加低风险入口一起清。
**Notes:** 该选择延续了 Phase 12 已经锁定的“先清当前真源，历史大迁移延后”策略，但把脚本名和测试名等低风险入口纳入本 phase。

---

## Operational Documentation Shape

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal runbook | 只补一份精简维护说明，覆盖最常用命令与入口 | |
| Full operator docs | 补完整维护手册，串起 trace/logging/retry、验证命令与回滚关注点 | ✓ |
| Archive-only refresh | 只更新已有 phase verification / milestone 文档，不新增面向维护者的整合说明 | |

**User's choice:** 采用推荐默认方案：补完整运维文档。
**Notes:** 文档目标是让维护者能直接执行通知失败排查，而不是停留在 phase 级证据罗列。

---

## Naming Cleanup Boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Low-risk only | 不碰高风险 module/import/runtime 命名，只处理脚本、测试名和文案型入口 | ✓ |
| Mixed cleanup | 在低风险清理之外，尝试少量内部运行时命名调整 | |
| Deep rename | 推动 module path / import / issuer 等更深层历史命名迁移 | |

**User's choice:** 采用推荐默认方案：只清低风险命名，不碰高风险 module/import/runtime 边界。
**Notes:** 这与 `.planning/phases/12-establish-automated-quality-gates/12-CONTEXT.md` 中已 defer 的高风险历史命名迁移保持一致。

---

## the agent's Discretion

- 运维文档最终拆成单文档还是双文档结构，交由 planner 在不偏离“完整维护手册”目标的前提下细化。
- 低风险命名清理的精确文件清单与批次，交由研究/计划阶段基于实际扫描结果确定。

## Deferred Ideas

- Go module path / import path / runtime issuer 的深层历史命名迁移
- 文档站、知识库系统或更大规模的信息架构重做

