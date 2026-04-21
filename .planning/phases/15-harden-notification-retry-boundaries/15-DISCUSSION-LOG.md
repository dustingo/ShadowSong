# Phase 15: Harden Notification Retry Boundaries - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-21
**Phase:** 15-harden-notification-retry-boundaries
**Areas discussed:** Retry Boundary, Failure Landing Zone, Retry Policy Shape, Attempt Context, Retry Window

---

## Retry Boundary

| Option | Description | Selected |
|--------|-------------|----------|
| 只重试“发送阶段”的瞬时失败 | 模板渲染失败、数据源缺失、渠道配置错误等确定性错误不重试；只对网络错误、5xx、超时等瞬时失败做有界重试 | ✓ |
| 发送失败都先进统一重试判断 | 在统一入口分类瞬时/确定性错误，再决定是否重试 | |
| 所有失败都按同一策略重试 | 不区分错误类型，全部重试 | |

**User's choice:** 只重试发送阶段的瞬时失败  
**Notes:** 用户明确要求 Phase 15 不把确定性失败纳入重试范围。

---

## Failure Landing Zone

| Option | Description | Selected |
|--------|-------------|----------|
| 先落日志 | 最终失败先以稳定日志字段落地，作为排障真源 | ✓ |
| 落日志，并写回现有持久化对象 | 在现有模型上补最小失败状态/时间/错误摘要 | |
| 落日志，并写入独立失败记录表 | 新增独立失败记录表作为最终落点 | |

**User's choice:** 先落日志  
**Notes:** 用户明确不希望在 Phase 15 引入新的持久化失败记录。

---

## Retry Policy Shape

| Option | Description | Selected |
|--------|-------------|----------|
| 所有渠道先用同一套固定重试策略 | `feishu`、`dingtalk`、`wecom`、`webhook` 先统一策略 | ✓ |
| 默认策略，但允许按渠道类型分支 | 允许不同渠道类型使用不同重试策略 | |
| 为每个渠道开放独立配置 | 每个渠道可配置自己的重试行为 | |

**User's choice:** 所有渠道先用同一套固定重试策略  
**Notes:** 用户优先选择统一行为和较低复杂度，不在本 phase 开放差异化配置。

---

## Attempt Context

| Option | Description | Selected |
|--------|-------------|----------|
| 最小集 | `trace_id`、`alert_id`、`channel_id`、`attempt`、`max_attempts`、`error` | ✓ |
| 扩展集 | 在最小集基础上增加 `channel_name`、`channel_type`、`source`、`fingerprint`、阶段名 | |
| 自定义字段集 | 用户自定义一套固定字段 | |

**User's choice:** 最小集  
**Notes:** 用户要求 attempt-level 新约定保持克制，避免提前扩成 Phase 16 的日志统一工程。

---

## Retry Window

| Option | Description | Selected |
|--------|-------------|----------|
| 短窗口内完成 | 重试都在当前异步 goroutine 里快速完成，不引入明显长时后台处理 | ✓ |
| 允许明显拉长的后台重试窗口 | 在当前 goroutine 内拉长重试时间，形成更长后台处理窗口 | |

**User's choice:** 短窗口内完成  
**Notes:** 用户明确不希望 Phase 15 把当前异步发送演化成准队列/长时后台调度。

---

## the agent's Discretion

- 固定重试策略的具体次数、间隔和退避算法由研究/规划阶段决定
- 瞬时失败的分类规则、日志 stage 名称和 helper 封装粒度由后续 planner 决定

## Deferred Ideas

- 每渠道独立重试配置
- 独立通知失败记录表
- 长时后台重试窗口或队列化调度
- 前端通知投递状态展示
