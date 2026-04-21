# Phase 13: Harden Notification Delivery Path - Context

**Gathered:** 2026-04-21
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只加固 webhook 入库后的异步通知链路：给异步处理补 panic recover、失败边界、稳定日志和最小可追踪上下文，减少静默失败与排障盲区。范围不包含重做通知架构、引入任务队列、改变现有路由/模板产品能力，或扩展为完整可观测性平台建设。

</domain>

<decisions>
## Implementation Decisions

### Failure Boundary
- **D-01:** 当前 `go h.processAlertNotifications(newAlerts)` 的异步通知入口必须具备 panic recover，不能让通知 goroutine 的 panic 直接威胁服务稳定性。
- **D-02:** panic 防护要落在真实异步边界附近，而不是只在更深层 sender 内零散补丁，确保整个通知处理批次都有兜底。

### Logging Contract
- **D-03:** 通知失败日志必须至少携带稳定的告警和渠道上下文，例如 `alert_id`、`source`、`channel_id`、`channel_name`、通知阶段或失败原因，不能只剩裸错误字符串。
- **D-04:** 当前 phase 优先建立“稳定可检索”的后端日志契约，不强制引入新的日志基础设施；可以沿用标准库日志，但不能继续主要依赖无结构的 `fmt.Printf` 诊断输出。

### Traceability Scope
- **D-05:** Phase 13 追踪目标是把关键失败点和具体告警/渠道关联起来，至少覆盖模板渲染失败、渠道发送失败、无匹配规则和异步批次 panic 等主要路径。
- **D-06:** 这轮追踪只做最小必要上下文，不扩展到持久化通知事件表、分布式 trace、消息重试队列或补偿机制。

### Verification Boundary
- **D-07:** Phase 13 必须补直接验证，证明 panic 不会把通知处理链路打崩，并证明失败日志/错误边界至少在代码层可观测、可定位。
- **D-08:** 验证要尽量靠现有 Go 测试落地，优先补 handler/notifier 级别的可靠性测试，而不是仅靠手工阅读代码宣布完成。

### the agent's Discretion
- 具体日志字段组织方式、helper 封装粒度和测试替身设计可由后续计划阶段决定，只要满足“最小侵入、稳定上下文、易验证”。
- 如需在 `WebhookHandler` 内新增小型注入 seam 以便测试 panic/失败边界，可做最小范围改造，但不改变现有业务路由语义。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — 当前里程碑约束和唯一剩余 Active requirement
- `.planning/REQUIREMENTS.md` — Phase 13 对应的 `NTFY-01` 到 `NTFY-03`
- `.planning/ROADMAP.md` — Phase 13 的目标、依赖和 success criteria
- `.planning/STATE.md` — 当前阶段位置与 v1.2 收尾状态

### Live Code To Reuse
- `internal/handlers/webhook.go` — webhook 入库、异步通知 goroutine、路由匹配与发送主链路
- `internal/notifier/notifier.go` — 各渠道 sender 的失败返回点和 HTTP 边界
- `internal/models/models.go` — `Channel`、`RouteRule`、`DataSource` 等通知链路涉及的数据模型
- `internal/handlers/webhook_test.go` — 当前 webhook / 通知模板相关测试基础

### Relevant Prior Verification
- `.planning/phases/12-establish-automated-quality-gates/12-VERIFICATION.md` — 当前 CI 和本地验证基线，说明新增 Go 测试会被自动门禁覆盖
- `.planning/phases/10-secure-realtime-alert-access/10-VERIFICATION.md` — 近期 phase 的验证写法和真相文档格式

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `WebhookHandler.processAlertNotifications` 已经是异步通知主入口，适合放置批次级 panic/日志边界。
- `WebhookHandler.sendNotification` / `sendDefaultNotification` 已经集中经过 `notifier.SendToChannel`，适合统一补充失败上下文。
- `notifier.SendToChannel` 已经把不同 channel type 聚合到一个入口，便于在不改产品能力的前提下统一发送错误契约。

### Established Patterns
- 当前通知链路大量使用 `fmt.Println` / `fmt.Printf` 做操作诊断，说明 Phase 13 需要把这些关键路径至少收口到更稳定的日志形式。
- sender 层当前主要用 `fmt.Errorf` 返回失败，说明链路已有错误传播基础，但缺少统一上下文补充与异步边界保护。
- webhook handler 仍是薄层直接编排数据库、模板和通知发送，没有 service layer；Phase 13 应顺着现有层次做局部加固，而不是借机重构架构。

### Integration Points
- `HandleWebhook` 中的 `go h.processAlertNotifications(newAlerts)` 是异步边界和 panic recover 的第一落点。
- `findMatchedChannels` 与 `sendNotification` 是“无匹配规则”和“具体渠道失败”上下文日志的主要来源。
- `internal/notifier/notifier.go` 中各 sender 的 HTTP 响应处理是补充稳定失败信息的下游边界。

</code_context>

<specifics>
## Specific Ideas

- 这期目标是“让通知失败可见、可追踪、不会直接炸掉异步链路”，不是把通知系统升级成完整任务平台。
- 如果测试需要替换实际发送器，优先引入最小 seam，而不是为了测试把 notifier 整体重写成复杂接口树。
- 当前 `fmt.Printf("Failed to send notification to channel %d: %v\n", ...)` 这类日志已经暴露了缺口：有部分上下文，但还不稳定也不完整，适合在本 phase 统一化。

</specifics>

<deferred>
## Deferred Ideas

- 通知重试队列、死信队列和持久化发送记录
- 结构化日志基础设施升级或集中日志平台接入
- 分布式追踪、通知指标面板和更细粒度 SLA 监控

None of the above belongs in Phase 13 scope.

</deferred>

---

*Phase: 13-harden-notification-delivery-path*
*Context gathered: 2026-04-21*
