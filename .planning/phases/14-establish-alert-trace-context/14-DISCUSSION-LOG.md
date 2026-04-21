# Phase 14: Establish Alert Trace Context - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-21
**Phase:** 14-establish-alert-trace-context
**Areas discussed:** 关联标识方案, 观测覆盖范围, 生命周期打点粒度

---

## 关联标识方案

| Option | Description | Selected |
|--------|-------------|----------|
| 新增独立 trace 标识 | 建立独立 `trace_id`/trace context 作为一次 webhook 处理链路真源，与 `alert_id`、`fingerprint` 分工明确 | ✓ |
| 复用 `alert_id` / `fingerprint` | 不新增 trace 真源，仅使用现有字段拼接观察链路 | |
| 只做临时日志 request id | 不落稳定真源，只在日志中补充临时请求标识 | |

**User's choice:** 默认
**Notes:** 按推荐方案收敛为新增独立且稳定的 trace 标识，并要求后端存在明确真源。

---

## 观测覆盖范围

| Option | Description | Selected |
|--------|-------------|----------|
| 后端主链路 | 覆盖 `webhook -> 落库/去重 -> Redis -> 路由匹配 -> 通知入口`，不扩展到前端/WebSocket | ✓ |
| 端到端扩展到前端 | 同步把 trace 信息带到 WebSocket 和前端展示层 | |
| 最小闭环 | 只覆盖 webhook 和通知两端，不系统梳理中间阶段 | |

**User's choice:** 默认
**Notes:** 按推荐方案限制在后端主链路，避免 Phase 14 scope 膨胀到前端联动。

---

## 生命周期打点粒度

| Option | Description | Selected |
|--------|-------------|----------|
| 关键阶段结果 | 只记录接入、落库/去重、Redis、路由匹配、通知入口等关键阶段结果 | ✓ |
| 细粒度事件流 | 每个中间步骤都形成更细的事件记录 | |
| 仅失败打点 | 平时不记阶段结果，只在失败时额外记录 | |

**User's choice:** 默认
**Notes:** 按推荐方案只建立低侵入的关键阶段观测点，不做过细埋点。

---

## the agent's Discretion

- trace 字段具体命名、持久化方式和日志格式由后续研究与计划阶段细化。
- 关键观测点的 helper 组织方式由 planner 决定，只要满足最小侵入和可回溯。

## Deferred Ideas

- 将 trace context 扩展到 WebSocket / 前端可视化
- 引入集中式 tracing / metrics / logging 平台
- 构建细粒度事件流或审计表
