# Phase 18: Establish Delivery Ledger - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-29T00:00:00+08:00
**Phase:** 18-establish-delivery-ledger
**Areas discussed:** 账本粒度, 不可变快照范围, Phase 18 可查看落点, 恢复语义边界

---

## 账本粒度

| Option | Description | Selected |
|--------|-------------|----------|
| `delivery` 主记录 + `attempts` 明细 | 支撑成功/失败全量历史、后续恢复、健康聚合与 attempt 级诊断 | ✓ |
| 只存最终结果 | 改动最小，但会丢失 attempt 细节与健康/恢复依据 | |
| 只存 attempt 明细 | 保留发送细节，但缺少稳定主记录与当前终态锚点 | |

**User's choice:** the agent selected the recommended option.
**Notes:** 现有链路在 `internal/handlers/webhook.go` 中只有日志与内存态发送，没有 durable ledger。后续 Phase 19/21 明显依赖稳定主记录与 attempt 证据链，因此选择两层模型最稳妥。

---

## 不可变快照范围

| Option | Description | Selected |
|--------|-------------|----------|
| 冻结 alert 关键字段 + route/channel 身份快照 + 渲染模式 + 实际发送内容 | 兼顾审计、原始 retry 和解释性，快照宽度可控 | ✓ |
| 只冻结 delivery intent，后续可重新渲染 | 实现更轻，但会削弱审计与原始发送可解释性 | |
| 连更多渠道配置/原始回执一起冻结 | 证据更全，但会显著扩大 schema 与敏感信息处理面 | |

**User's choice:** the agent selected the recommended option.
**Notes:** 用户已把 `retry` 定义为原始发送语义，因此仅冻结 intent 不够；同时本 phase 不宜把敏感配置和过重回执一并固化。

---

## Phase 18 可查看落点

| Option | Description | Selected |
|--------|-------------|----------|
| 先做持久化 + 最小只读 API/详情，完整历史页放 Phase 19 | 满足 success criteria，同时不挤占后续恢复/UI phase 边界 | ✓ |
| Phase 18 只落库，不给读接口 | 改动最小，但无法满足“维护者可以查看任一账本记录” | |
| Phase 18 直接带列表/详情 UI | 操作体验更完整，但会提前吞掉 Phase 19 的历史面 scope | |

**User's choice:** the agent selected the recommended option.
**Notes:** roadmap 明确要求维护者可查看账本记录，因此不能只落库；但完整历史页和搜索面更适合与恢复操作一起在下一 phase 收口。

---

## 恢复语义边界

| Option | Description | Selected |
|--------|-------------|----------|
| replay 原始投递意图/原始发送内容，不重新跑当前策略 | 语义最稳定，但 retry/replay 基本重合 | |
| replay 时重新按当前配置计算 | 把 replay 定义成“基于当前策略重放” | |
| `retry` 走原始发送，`replay` 走当前策略 | 明确区分两种恢复语义，兼顾原始重试与策略重放 | ✓ |

**User's choice:** `retry` 走原始发送，`replay` 走当前策略。
**Notes:** 这一定义会直接影响账本快照设计与后续恢复 API 语义。Phase 18 需保证账本同时支撑 deterministic retry 与 policy-based replay。

---

## the agent's Discretion

- 最小只读能力的具体 API 形态可在规划时按最小满足 success criteria 的原则决定。
- 具体表字段命名、状态枚举和最小索引集合留给 researcher/planner 结合现有代码约束细化。

## Deferred Ideas

- 批量恢复与复杂 replay preview
- 完整 delivery history UI 与 ops health surface
- 更重的快照/回执存储范围
