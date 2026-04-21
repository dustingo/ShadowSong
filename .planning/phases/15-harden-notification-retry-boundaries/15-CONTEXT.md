# Phase 15: Harden Notification Retry Boundaries - Context

**Gathered:** 2026-04-21
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只在现有异步通知实现内部补齐“发送阶段”的有界重试、最终失败落点和尝试级上下文，降低瞬时失败导致的通知丢失风险。范围不包含消息队列、独立投递记录中心、按渠道自定义重试配置、前端失败展示或新的外部基础设施。

</domain>

<decisions>
## Implementation Decisions

### Retry Boundary
- **D-01:** Phase 15 只重试“发送阶段”的瞬时失败，不把模板渲染失败、数据源查找失败、渠道配置错误或其他确定性错误纳入重试。
- **D-02:** 发送失败需要先经过统一的可重试/不可重试判断，再决定是否进入重试，但最终行为以“仅瞬时失败可重试”为准。

### Final Failure Landing Zone
- **D-03:** 重试耗尽后的最终失败先落稳定日志，不在本 phase 引入新的持久化失败记录表，也不把最终失败写回新的持久化业务对象字段。
- **D-04:** 最终失败日志必须成为值班排障的主要落点，能够和已有 Phase 14 trace/lifecycle 日志衔接。

### Retry Policy Shape
- **D-05:** Phase 15 先让 `feishu`、`dingtalk`、`wecom`、`webhook` 四类渠道共用一套固定重试策略，不做按渠道类型差异化，也不开放每渠道独立配置。
- **D-06:** 固定重试策略的具体次数、间隔和退避方式可由研究/计划阶段决定，但必须是有界的，且适合当前 goroutine 内异步发送模型。

### Attempt-Level Observability
- **D-07:** 每次通知尝试的稳定日志字段采用最小集：`trace_id`、`alert_id`、`channel_id`、`attempt`、`max_attempts`、`error`。
- **D-08:** 本 phase 不要求额外扩展 `channel_name`、`channel_type`、`source`、`fingerprint` 之外的新 attempt 级字段集合；如已有上下文已包含，不必刻意移除，但新约定以最小集为准。

### Runtime Boundary
- **D-09:** 重试必须在当前异步 goroutine 的短窗口内完成，不引入明显拉长的后台重试窗口，也不把单次 webhook 的后台处理演化成准队列调度。
- **D-10:** 本 phase 继续坚持 brownfield 小步增强，沿现有 `WebhookHandler` -> `sendNotification` -> `notifier.SendToChannel` 主链路演进，不借机重构完整通知架构。

### the agent's Discretion
- 瞬时失败的错误分类标准、固定重试次数、短间隔/退避细节和 helper 封装粒度，可由研究与计划阶段决定，只要满足“有界、短窗口、全渠道统一策略”。
- 最终失败日志与尝试日志的具体 stage 名称、字段排序和测试替身设计，可由 planner 决定，但应复用 Phase 13/14 已有通知日志 seam 与 trace 上下文。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — v1.3 目标、brownfield 约束以及“不引入新队列/外部平台”的里程碑决定
- `.planning/REQUIREMENTS.md` — Phase 15 对应的 `NTFY-01`、`NTFY-02`、`NTFY-03`
- `.planning/ROADMAP.md` — Phase 15 的 goal、plans 和 success criteria
- `.planning/STATE.md` — 当前阶段位置、Phase 14 已完成事实和后续 Phase 16 依赖关系

### Prior Phase Decisions
- `.planning/phases/13-harden-notification-delivery-path/13-CONTEXT.md` — 已锁定通知链路的 panic recover、最小可追踪上下文和“不引入重试队列”的边界
- `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md` — 已锁定 trace 真源、生命周期观测点和主链路范围
- `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md` — 已验证 `trace_id`、lifecycle stage 和 Redis/notification trace 传播的当前真相

### Live Code To Reuse
- `internal/handlers/webhook.go` — 当前异步通知入口、通知日志 helper、trace 字段传播和 `sendNotification` 主链路
- `internal/handlers/webhook_test.go` — 当前通知入口 / lifecycle / send failure 测试 seam
- `internal/notifier/notifier.go` — 各渠道 sender 的 HTTP 边界、错误返回契约和全渠道聚合入口
- `internal/notifier/notifier_test.go` — notifier 层现有错误契约测试基础
- `internal/models/models.go` — `Channel`、`RouteRule`、`DataSource` 等通知链路关联模型

### Architecture And Conventions
- `.planning/codebase/ARCHITECTURE.md` — webhook -> Redis -> notification 当前数据流和无队列现状
- `.planning/codebase/CONVENTIONS.md` — handler / logging / error handling / 测试补充约定
- `.planning/codebase/STACK.md` — 当前技术栈与“不做技术迁移”的约束

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `WebhookHandler.processAlertNotificationsAsync` 已经是异步通知边界，适合在这里保持短窗口重试而不扩展到新调度体系。
- `WebhookHandler.logNotification` 与 `logTraceStage` 已经能输出 `trace_id`、`alert_id`、`channel_id` 等字段，可直接承接 attempt-level 和 terminal-failure 日志。
- `WebhookHandler.sendNotification` 已经把模板渲染、sender 调用和默认通知 fallback 聚合在一起，是加错误分类和重试判断的自然位置。
- `notifier.SendToChannel` 已经统一封装四类渠道 sender，便于在不做按渠道差异化配置的前提下挂统一重试策略。

### Established Patterns
- 当前发送链路是 `WebhookHandler` 直接编排，没有 service layer，也没有后台任务系统；Phase 15 应沿现有 handler/notifier seam 做局部增强。
- 当前 sender 基本通过返回 `error` 表达失败，说明已有错误传播基础，但还缺少瞬时/确定性失败分类与 attempt 级上下文。
- Phase 14 已建立 `trace_id` 真源和 `notification_entry` / `send_notification` 等阶段日志，因此 Phase 15 不需要重新设计 trace，只需要在重试与最终失败路径复用它。

### Integration Points
- `HandleWebhook` 中的 `h.processAlertNotificationsAsync(newAlerts)` 是整个重试窗口的上游入口。
- `processAlertNotifications` 遍历 alert/channel 的位置，决定 attempt context 需要绑定到“单 alert + 单 channel”发送单元。
- `sendNotification` 和 `sendDefaultNotification` 是最适合插入重试判定、尝试计数与最终失败日志的位置。
- `internal/notifier/notifier.go` 中各 sender 的 HTTP 错误与状态码，是研究“瞬时失败 vs 确定性失败”分类的下游事实来源。

</code_context>

<specifics>
## Specific Ideas

- 这期目标是“短窗口内补齐重试边界与失败落点”，不是把通知系统升级成持久化投递平台。
- 固定重试策略应优先服务于值班排障和行为可预测性，而不是追求每个渠道最优调参。
- 尝试级日志字段只要求最小集，避免在 Phase 15 抢跑 Phase 16 的日志统一范围。

</specifics>

<deferred>
## Deferred Ideas

- 为每个渠道开放独立重试策略或按渠道类型分叉策略
- 新增独立通知失败记录表、投递历史中心或补发/重放界面
- 把重试窗口扩展成明显长时后台处理或引入队列/调度器
- 在前端页面直接展示通知投递失败与重试状态

None of the above belongs in Phase 15 scope.

</deferred>

---

*Phase: 15-harden-notification-retry-boundaries*
*Context gathered: 2026-04-21*
