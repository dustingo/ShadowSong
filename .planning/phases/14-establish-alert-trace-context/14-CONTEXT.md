# Phase 14: Establish Alert Trace Context - Context

**Gathered:** 2026-04-21
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只为告警后端主链路建立统一 trace context 和关键生命周期观测点，让运维可以从 webhook 接入一路追到落库、路由匹配、Redis 发布和通知入口。范围不包含通知重试策略、完整日志体系重构、前端展示联动、WebSocket 实时消费改造或外部可观测性平台接入。

</domain>

<decisions>
## Implementation Decisions

### Trace Identity Strategy
- **D-01:** Phase 14 不只复用现有 `alert_id` 或 `fingerprint` 作为链路观测真源，而是要建立独立且稳定的 trace 标识，用于表达“一次 webhook 处理链路”的上下文。
- **D-02:** 这个 trace 标识必须在后端有明确真源，优先落在 `Alert` 相关持久化数据或可稳定重建的字段体系里，不能只存在于某一条临时日志输出中。
- **D-03:** 现有 `alert_id` 继续承担单条告警标识，`fingerprint` 继续承担去重/聚合语义；新 trace context 不能混淆这两类既有职责。

### Observability Scope
- **D-04:** Phase 14 的观测覆盖范围先限定在后端主链路：`webhook 接入 -> 告警渲染/落库/去重 -> Redis 发布 -> 路由匹配 -> 通知入口`。
- **D-05:** 本 phase 不要求把 trace context 继续暴露到 WebSocket 推送、前端 store 或页面展示；这些如果要做，应放到后续 phase 单独决策。
- **D-06:** 生命周期观测点至少要让运维能回答三类问题：这次 webhook 是否被接收、告警是否真正入库/被去重、后续是否进入路由和通知处理。

### Event Granularity
- **D-07:** Phase 14 只记录关键阶段结果，不做每个内部细节的细粒度事件流；目标是先建立稳定、可检索、低侵入的排障基线。
- **D-08:** 关键阶段应优先覆盖接入开始、告警创建/去重结果、Redis 发布结果、路由匹配结果和通知入口开始，而不是把模板函数内部或每个微小分支都打成事件。
- **D-09:** 允许在关键失败点补充更具体的上下文字段，但不把本 phase 扩展成通用审计事件系统或分布式 tracing 框架。

### Compatibility And Rollout
- **D-10:** Phase 14 必须保持现有告警接入、展示、路由、静默和值班能力不变；trace context 是增强排障能力，不是改变业务语义。
- **D-11:** 当前项目继续坚持 brownfield 小步增强，不引入消息队列、OpenTelemetry、Prometheus 或外部日志平台来完成本 phase 目标。
- **D-12:** 验证应证明从一条告警或通知相关日志可以回溯同一次处理链路中的至少多个关键阶段，而不是只证明字段“被写进某个 struct”。

### the agent's Discretion
- trace 字段的具体命名、落库方式和在日志中的序列化格式可由后续研究/计划阶段决定，只要满足“稳定真源、后端可回溯、最小侵入”。
- 关键观测点最终通过日志 helper、模型字段扩展还是 handler 内聚合辅助函数实现，可由 planner 决定，但应顺着现有 `WebhookHandler` 和 `Alert` 模型结构演进。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — v1.3 目标、brownfield 约束和“先在现有应用内补齐关联标识”的里程碑决定
- `.planning/REQUIREMENTS.md` — Phase 14 对应的 `OBS-01`、`OBS-02`
- `.planning/ROADMAP.md` — Phase 14 的 goal、plans 和 success criteria
- `.planning/STATE.md` — 当前阶段位置与后续 Phase 15/16 依赖关系

### Prior Phase Decisions
- `.planning/phases/13-harden-notification-delivery-path/13-CONTEXT.md` — 已锁定“最小必要上下文、暂不引入新平台”的 Phase 13 追踪边界
- `.planning/phases/10-secure-realtime-alert-access/10-CONTEXT.md` — 近期 phase 的 context 写法、最小侵入原则和 WebSocket 暂不扩 scope 的先例
- `.planning/phases/12-establish-automated-quality-gates/12-CONTEXT.md` — 文档真相、低风险演进和验证边界的近期约束

### Architecture And Conventions
- `.planning/codebase/ARCHITECTURE.md` — webhook/Redis/notification/WebSocket 当前数据流和未接入事实
- `.planning/codebase/CONVENTIONS.md` — handler、日志、错误处理和测试补充约定
- `.planning/codebase/STACK.md` — 当前技术栈和“不做技术迁移”的边界

### Live Code To Reuse
- `internal/handlers/webhook.go` — webhook 主链路、Redis 发布、路由匹配、通知入口和现有通知日志 helper
- `internal/handlers/webhook_test.go` — 当前通知链路测试 seam 和日志断言基础
- `internal/models/alert.go` — 告警持久化模型与可承载 trace 真源的核心实体
- `internal/handlers/websocket.go` — 当前 WebSocket 仅用于实时消费，作为本 phase 明确不纳入的对照边界
- `internal/router/router.go` — `/webhook/:source_name`、`/ws/alerts` 当前入口挂载方式

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `WebhookHandler.HandleWebhook` 已经串起接入、模板渲染、去重、落库、Redis 发布和异步通知入口，是添加 trace context 传播的主干位置。
- `WebhookHandler.logNotification` 已经具备阶段化日志 helper 和告警/渠道字段拼装能力，可作为后续统一 trace 日志约定的切入点。
- `internal/handlers/webhook_test.go` 已有日志输出断言和替身 seam，可用于补充 trace 生命周期验证。

### Established Patterns
- 当前系统没有 service layer，告警主链路直接在 handler 中编排数据库、Redis 和通知发送；Phase 14 应沿现有结构做局部增强，而不是借机重构架构。
- `alert_id`、`fingerprint`、`status` 已经分别承担业务标识、聚合去重和状态职责，因此新的 trace context 需要明确区分语义。
- 后端运行时日志目前同时存在 `log.Printf` 与 `fmt.Printf` 风格，说明本 phase 应把观测增强优先落在关键主链路，而不是一次性统一全仓日志实现。

### Integration Points
- `HandleWebhook` 中读取原始 body、遍历 normalized alerts 和调用 `publishToRedis` / `processAlertNotificationsAsync` 的链路，是 trace 生成和传播的首要集成点。
- `publishToRedis` 当前只写入 `alert_id`、`source`、`fingerprint` 等字段，若要让后续消费者可关联，需要在这里明确 trace 传播策略。
- `processAlertNotifications` 与 `findMatchedChannels` / `sendNotification` 是 route match 和通知入口生命周期观测点的主要落点。

</code_context>

<specifics>
## Specific Ideas

- 默认实现方向是“后端先打通 trace 真源和关键阶段回溯”，不是“用户界面可视化 trace”。
- 如果需要在模型层新增字段，优先让它成为后续 Phase 15 重试与 Phase 16 日志统一都能复用的真源，而不是一次性临时字段。
- 如果单次 webhook 会产生多条新告警，trace context 需要能表达“同一批处理链路”与“单条告警”的关系，避免后续日志只剩孤立的 alert 视角。

</specifics>

<deferred>
## Deferred Ideas

- 将 trace context 暴露到 WebSocket 消息或前端页面，形成端到端可视化链路
- 引入集中式 tracing / metrics 平台，例如 OpenTelemetry、Prometheus 或外部日志系统
- 把关键阶段观测扩展成更细粒度的事件流、审计表或通用事件总线

None of the above belongs in Phase 14 scope.

</deferred>

---

*Phase: 14-establish-alert-trace-context*
*Context gathered: 2026-04-21*
