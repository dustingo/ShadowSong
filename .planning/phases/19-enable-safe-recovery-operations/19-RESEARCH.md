# Phase 19: Enable Safe Recovery Operations - Research

**Researched:** 2026-04-30  
**Domain:** Go + Gin + GORM + PostgreSQL 单条通知恢复与账本查询  
**Confidence:** HIGH

## User Constraints

- 未发现 `.planning/phases/19-enable-safe-recovery-operations/19-CONTEXT.md`；本研究以 `REQUIREMENTS.md`、`STATE.md`、`ROADMAP.md`、`AGENTS.md` 和 Phase 18 context/summary 为约束来源。[VERIFIED: .planning/phases/19-enable-safe-recovery-operations][VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: .planning/STATE.md][VERIFIED: .planning/ROADMAP.md][VERIFIED: AGENTS.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]

### Effective Locked Decisions

- 只做单条 `retry` 与单条 `replay`，不扩展批量恢复。[VERIFIED: .planning/STATE.md][VERIFIED: .planning/REQUIREMENTS.md]
- `retry` 语义是沿原始发送语义再次执行，不重新走当前策略。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- `replay` 语义是重新走当前策略，而不是复用原始 route/template 结果。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- 继续维持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite 技术栈，不做技术迁移。[VERIFIED: AGENTS.md]
- 不引入 MQ、Temporal、Asynq 或外部 workflow 平台。[VERIFIED: .planning/REQUIREMENTS.md]
- 账本已经是恢复真源，Phase 19 需要基于 Phase 18 的 `notification_deliveries` / `notification_delivery_attempts` 继续扩展，而不是另起一套历史模型。[VERIFIED: .planning/ROADMAP.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-01-SUMMARY.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-02-SUMMARY.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-03-SUMMARY.md]

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DELV-03 | 维护者可以针对单条失败通知执行受控的 `retry`。[VERIFIED: .planning/REQUIREMENTS.md] | 需要单条恢复写 API、原始语义发送执行器复用、失败可见性和权限边界。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/handlers/user.go] |
| DELV-04 | 维护者可以针对单条失败通知执行受控的 `replay`。[VERIFIED: .planning/REQUIREMENTS.md] | 需要基于当前 datasource/route/channel/template 重新决策的执行路径，并把新投递与原账本关联。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: internal/handlers/webhook.go] |
| DELV-05 | 人工 `retry/replay` 会记录操作者、触发原因、触发时间、执行结果和关联原始投递记录，保持审计链路完整。[VERIFIED: .planning/REQUIREMENTS.md] | 现有 `AuditLog` 可记录操作者与动作，但 Phase 19 还需要能把一次恢复请求和 resulting delivery 关联起来。[VERIFIED: internal/models/models.go][VERIFIED: internal/handlers/user.go][ASSUMED] |
| OPER-01 | 维护者可以按时间、告警、渠道和结果查询通知投递历史，而不需要直接翻后端日志。[VERIFIED: .planning/REQUIREMENTS.md] | Phase 18 已有最小列表/详情 API；Phase 19 需要把它变成可操作的历史面，并补足面向 UI 的筛选入口与失败证据展示。[VERIFIED: internal/handlers/delivery.go][VERIFIED: frontend/src/api/client.ts][VERIFIED: frontend/src/App.tsx] |
| OPER-04 | 维护者可以从告警详情或运维页面直接跳转到关联的通知投递历史与失败证据。[VERIFIED: .planning/REQUIREMENTS.md] | 当前前端没有 delivery 页面，也没有告警到账本的跳转入口，因此必须新增前端路由/页面和 alert-row deeplink。[VERIFIED: frontend/src/App.tsx][VERIFIED: frontend/src/pages/Alerts.tsx][VERIFIED: frontend/src/api/client.ts] |
</phase_requirements>

## Summary

Phase 19 不是“在现有列表旁边补两个按钮”这么简单；它至少包含三个实现切片：`1)` 把 Phase 18 的只读账本变成可筛选、可 deeplink 的运维历史面，`2)` 为单条失败 delivery 增加受控 `retry/replay` 写 API 与执行器，`3)` 为恢复动作补齐可审计链路和前端失败证据展示。[VERIFIED: .planning/ROADMAP.md][VERIFIED: internal/handlers/delivery.go][VERIFIED: frontend/src/App.tsx][VERIFIED: frontend/src/pages/Alerts.tsx]

现有代码最大的优势是发送热路径已经集中在 `WebhookHandler.sendNotification -> sendChannelNotification`，并且账本写入、attempt 记录、终态落库都在这一层完成。[VERIFIED: internal/handlers/webhook.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-02-SUMMARY.md] 这意味着 Phase 19 最稳妥的实现不是新起一条“人工恢复专用发送链路”，而是抽出一个可复用的通知执行器，让 webhook 热路径和恢复动作共同调用，同一套逻辑继续负责 `notificationMaxAttempts=3`、`IsRetryableSendError`、attempt 账本和最终结果记录。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/notifier/notifier.go][ASSUMED]

最大的规划风险有两个。第一，`retry` 和 `replay` 语义已经被上游锁定为两条不同路径，planner 不能把两者混成“重新发一次”。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] 第二，当前 ledger 只冻结了 `alert/channel/route/rendered payload/final failure`，没有冻结 channel secret/config，也没有把 `Alert.Raw` 存进 delivery snapshot；因此 `retry` 更适合定义为“对原 channel identity 重发冻结 payload”，而 `replay` 若要支持依赖 `.event` 的当前模板，就需要回查 `alerts` 表中的原始告警 `Raw` 字段，或明确接受 replay 上下文降级风险。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/models/alert.go][VERIFIED: internal/handlers/webhook.go][ASSUMED]

**Primary recommendation:** 把 Phase 19 规划成三段式交付：先交付 delivery 历史页和 alert->delivery deeplink，再交付后端恢复执行器与恢复审计模型，最后接前端 retry/replay 操作入口和结果反馈；其中 `GET` 历史继续走 `view_config`，`POST` 恢复动作复用 `process_alerts` 能力。[VERIFIED: internal/router/router.go][VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts][ASSUMED]

## Project Constraints (from AGENTS.md)

- 维持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite 技术栈。[VERIFIED: AGENTS.md]
- 尊重 brownfield 仓库与未提交改动，避免误删或回退无关修改。[VERIFIED: AGENTS.md]
- 不能破坏告警接入、展示、路由、静默和值班等现有主链路。[VERIFIED: AGENTS.md]
- 前端路由、菜单、类型和 API 调用必须保持自洽，不能留下断链入口或运行时错误。[VERIFIED: AGENTS.md]
- `workflow.nyquist_validation` 显式为 `false`，因此本研究故意省略 `## Validation Architecture` 章节。[VERIFIED: .planning/config.json]

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | `1.25.0`。[VERIFIED: go.mod][VERIFIED: go version] | 后端恢复 API、执行器和测试运行时。[VERIFIED: go.mod] | 当前仓库后端与现有 handler/test 都基于该 toolchain。[VERIFIED: AGENTS.md][VERIFIED: Makefile] |
| Gin | `v1.12.0`。[VERIFIED: go.mod] | 挂载 delivery 历史与恢复动作路由。[VERIFIED: internal/router/router.go] | 现有 API、JWT、capability middleware 都已固定在 Gin route group 上。[VERIFIED: internal/router/router.go] |
| GORM | `v1.31.1`。[VERIFIED: go.mod] | 继续承载 delivery/recovery 持久化、筛选和测试数据库模型。[VERIFIED: internal/delivery/service.go][VERIFIED: internal/models/notification_delivery.go] | Phase 18 的账本和现有 `AuditLog` 已经在 GORM 模型层稳定存在。[VERIFIED: internal/models/models.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-01-SUMMARY.md] |
| PostgreSQL | `14-alpine` 本地容器基线。[VERIFIED: docker-compose.yml] | 持久化 delivery 账本、恢复动作审计与关联关系。[VERIFIED: internal/database/postgres.go][VERIFIED: internal/models/notification_delivery.go] | Phase 19 需要强一致的查询与审计真源，继续用同库 additive schema 最稳妥。[VERIFIED: AGENTS.md][VERIFIED: .planning/ROADMAP.md] |
| React | `^18.2.0`。[VERIFIED: frontend/package.json] | 交付历史页、失败证据详情与恢复按钮入口。[VERIFIED: frontend/src/App.tsx][VERIFIED: frontend/src/pages/Alerts.tsx] | 当前前端路由、权限门控和页面结构都建立在 React 18 + React Router 上。[VERIFIED: frontend/src/App.tsx] |
| Ant Design | `^5.12.8`。[VERIFIED: frontend/package.json] | 表格筛选、Drawer/Modal、确认弹窗、权限提示。[VERIFIED: frontend/src/pages/Alerts.tsx] | 现有配置页与告警页已经统一使用 AntD table/form/modal 模式。[VERIFIED: frontend/src/pages/Alerts.tsx][VERIFIED: frontend/src/pages/Channels.tsx][VERIFIED: frontend/src/pages/RouteRules.tsx] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Zustand | `^4.4.7`。[VERIFIED: frontend/package.json] | 管理 delivery 历史列表状态和 recovery action loading/error 状态。[VERIFIED: frontend/src/stores/alertStore.ts] | 若 delivery 页面需要分页、筛选和跨组件刷新，沿用 store 模式；局部 modal state 仍放页面本地。[VERIFIED: AGENTS.md][VERIFIED: frontend/src/stores/alertStore.ts][ASSUMED] |
| Axios | `^1.6.5`。[VERIFIED: frontend/package.json] | 扩展 `/deliveries` 相关 GET/POST API 封装。[VERIFIED: frontend/src/api/client.ts] | 当前所有前端 API 都经由统一拦截器处理 401 和 token 注入。[VERIFIED: frontend/src/api/client.ts] |
| Vitest + Testing Library | `vitest ^1.6.1`、`@testing-library/react ^16.3.2`。[VERIFIED: frontend/package.json] | 覆盖 alert->delivery 跳转与权限可见性回归。[VERIFIED: frontend/src/pages/Alerts.test.tsx] | Phase 19 涉及新增前端交互，现有前端测试工具链可直接复用。[VERIFIED: frontend/package.json][VERIFIED: pnpm --dir frontend test -- --run Alerts.test.tsx] |
| Testify | `v1.11.1`。[VERIFIED: go.mod] | 覆盖恢复 handler/service 授权、账本关联和审计断言。[VERIFIED: internal/handlers/delivery_test.go] | 现有后端回归测试全部基于 testify 风格，继续复用最省改动。[VERIFIED: internal/handlers/delivery_test.go] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| 复用现有后端/前端栈 | 引入新 job/queue/recovery 平台 | 直接违反 milestone out-of-scope，并显著扩大实现面。[VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: AGENTS.md] |
| 在现有账本上追加恢复模型 | 只依赖 `AuditLog` 文本详情串联恢复 | 可以记人和时间，但很难稳定关联“原 delivery -> 恢复请求 -> resulting delivery”，对 UI 证据链也不友好。[VERIFIED: internal/models/models.go][VERIFIED: internal/handlers/user.go][ASSUMED] |
| 从 `WebhookHandler` 抽可复用发送执行器 | 手写第二套 recovery 发送循环 | 容易让 `retryable` 判定、3 次重试、attempt 落库和终态写法漂移。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/notifier/notifier.go][ASSUMED] |

**Installation:** 本 phase 研究未发现必须新增第三方依赖；优先在现有依赖集中完成。[VERIFIED: go.mod][VERIFIED: frontend/package.json][ASSUMED]

**Version verification:** 本研究推荐的库均来自仓库现有依赖声明和本机可执行环境，而不是新增包选择；因此版本以 `go.mod`、`frontend/package.json` 和本机命令探针为准。[VERIFIED: go.mod][VERIFIED: frontend/package.json][VERIFIED: go version][VERIFIED: node --version][VERIFIED: pnpm --version]

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── delivery/                    # 扩展：历史筛选、恢复执行、账本关联
├── handlers/
│   ├── delivery.go              # 扩展：GET 历史 + POST retry/replay
│   └── webhook.go               # 可能抽出共享发送执行器
├── models/
│   ├── notification_delivery.go
│   └── delivery_recovery.go     # 新增：恢复请求/结果关联 [ASSUMED]
frontend/src/
├── api/client.ts                # 扩展 deliveries API
├── pages/
│   ├── Alerts.tsx               # 新增 deeplink 入口
│   └── Deliveries.tsx           # 新增历史/恢复页面 [ASSUMED]
└── stores/
    └── deliveryStore.ts         # 若页面状态复杂则新增 [ASSUMED]
```

### Pattern 1: 读面与写面分离
**What:** `GET /api/v1/deliveries*` 保持历史查询与证据读取，`POST /api/v1/deliveries/:id/retry|replay` 专门承载恢复动作。[VERIFIED: internal/handlers/delivery.go][ASSUMED]  
**When to use:** 当同一 delivery 既要被 viewer/operator 查看，又只有 operator/admin 能执行恢复时。[VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts]  
**Why:** 当前 `/deliveries` 路由整组挂在 `CapabilityViewConfig` 下；如果直接在同组里加恢复动作而不拆 capability，会把危险写操作暴露给 viewer。[VERIFIED: internal/router/router.go][VERIFIED: internal/authz/capabilities.go]

### Pattern 2: 恢复请求与 resulting delivery 分开建模
**What:** 新增一条“恢复操作”记录，保存 `original_delivery_id`、`action`、`reason`、`actor`、`status`、`result_delivery_id`、`error_message`、时间戳；新发送结果继续落到 `notification_deliveries` / `notification_delivery_attempts`。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/models/models.go][ASSUMED]  
**When to use:** 单条恢复必须可审计，即使恢复在“尚未成功创建新 delivery”之前就失败，也要留下完整证据。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]  
**Why:** 单靠 `AuditLog.Detail` 文本不足以稳定支撑 UI 关联和后续查询，而把恢复元数据塞进新 delivery 又无法覆盖“恢复前置失败”的场景。[VERIFIED: internal/models/models.go][ASSUMED]

### Pattern 3: 抽取共享通知执行器
**What:** 把 `sendChannelNotification` 中“发送一次 logical delivery”的核心逻辑抽成可参数化 helper/service，让 webhook 热路径和人工恢复都走同一执行路径。[VERIFIED: internal/handlers/webhook.go][ASSUMED]  
**When to use:** `retry` 需要重发冻结 payload；`replay` 需要用当前策略重新选路后再执行同样的 bounded retry/attempt ledger 逻辑。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: internal/handlers/webhook.go]  
**Example:**
```go
// Source: internal/handlers/webhook.go + Phase 19 recommendation
type ExecuteDeliveryInput struct {
  Alert       *models.Alert
  Channel     *models.Channel
  RouteRule   *models.RouteRule
  Title       string
  Content     string
  Mode        string
  TriggerKind string
  OriginID    *uint
}

result, err := executor.Execute(ctx, ExecuteDeliveryInput{...})
```
Source: [VERIFIED: internal/handlers/webhook.go][ASSUMED]

### Pattern 4: Alert -> Delivery 深链而不是先做独立 alert detail 页
**What:** 在现有 `Alerts` 列表行或展开区提供“投递历史”入口，跳到 `/deliveries?alert_id=<alert_id>` 或 `/deliveries?trace_id=<trace_id>`。[VERIFIED: frontend/src/pages/Alerts.tsx][VERIFIED: frontend/src/App.tsx][ASSUMED]  
**When to use:** 当前前端没有独立 alert detail 路由，但 requirement 已要求能从告警详情或运维页面直接跳转到相关 delivery 证据。[VERIFIED: frontend/src/App.tsx][VERIFIED: .planning/REQUIREMENTS.md]  
**Why:** 这是满足 `OPER-04` 的最小 UI 方案，不必在 Phase 19 顺手扩一整套新的告警详情页。[VERIFIED: .planning/ROADMAP.md][ASSUMED]

### Anti-Patterns to Avoid
- **把 `retry` 与 `replay` 做成同一条代码路径，只是传不同按钮文案。** 这会直接违背上游语义边界。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- **把恢复结果 append 到原 delivery 的 attempts 上。** 原 delivery 是历史真相，人工恢复应产生新的 delivery 记录并与原记录关联，而不是改写历史。[VERIFIED: internal/models/notification_delivery.go][ASSUMED]
- **允许 viewer 继承 `/deliveries` 现有 `view_config` 后直接触发恢复。** 当前 capability 矩阵下 viewer 有 `view_config` 但没有 `process_alerts`。[VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts]
- **让前端自己决定 replay 语义或拼装恢复载荷。** 语义、权限、审计和失败判定都必须由后端掌控。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]

## Implementation Slices

### Slice 1: 历史查询与证据展示
- 复用现有 `DeliveryHandler.List/Get`，补前端 `deliveryApi`、`Deliveries` 页面和 deeplink 解析。[VERIFIED: internal/handlers/delivery.go][VERIFIED: frontend/src/api/client.ts][ASSUMED]
- 保持筛选边界聚焦在 `alert_id`、`trace_id`、`channel_id`、`delivery_status`、时间范围和分页；这已经足以支撑日志替代和 alert->history 跳转。[VERIFIED: internal/handlers/delivery.go][VERIFIED: .planning/REQUIREMENTS.md]
- 在 UI 中直接展示 `attempts`、`final_failure_summary`、`rendered_payload_snapshot` 作为失败证据，不要要求用户再翻日志。[VERIFIED: internal/handlers/delivery.go]

### Slice 2: 单条恢复后端
- 新增恢复操作模型和 POST API；动作输入至少包含 `reason`，动作输出至少包含 `recovery_status`、`result_delivery_id`、`error`。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]
- `retry` 读取原 delivery 的冻结 `rendered_payload_snapshot` 和原 `channel_id`，重发冻结 title/content，不重新 route/render。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- `replay` 读取原 `alert_id` 对应的 `alerts` 行与当前 datasource/route/channel 配置，重新 route/render 后执行发送。[VERIFIED: internal/models/alert.go][VERIFIED: internal/handlers/webhook.go][ASSUMED]

### Slice 3: 前端恢复入口
- 只在失败 delivery 上显示 `retry` / `replay` 按钮，并要求填写原因。[VERIFIED: .planning/ROADMAP.md][ASSUMED]
- 只对具备恢复能力的角色显示动作按钮；viewer 继续只读。[VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts][ASSUMED]
- 恢复完成后自动刷新原记录详情和新 resulting delivery 详情，形成闭环证据。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]

## API And UI Boundaries

### Backend API Boundary
- 后端负责资格校验：只有失败 delivery 可恢复，且必须存在有效原因文本。[VERIFIED: .planning/ROADMAP.md][ASSUMED]
- 后端负责语义区分：`retry` 走冻结 payload；`replay` 走当前策略。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- 后端负责写审计和结果关联：无论成功失败，都要记录 actor、reason、time、result、original delivery、result delivery。[VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/models/models.go][ASSUMED]

### Frontend UI Boundary
- 前端只负责筛选条件、deeplink、原因输入和结果展示，不负责构造业务语义。[VERIFIED: frontend/src/api/client.ts][ASSUMED]
- 新增运维页面比“把所有逻辑塞进 Alerts 页”更稳妥，因为当前 `Alerts.tsx` 已承担筛选、ack、quick silence 和权限提示；继续叠加 delivery 历史表会明显增肥组件。[VERIFIED: frontend/src/pages/Alerts.tsx][ASSUMED]
- `OPER-04` 的最小实现可以是 `Alerts` 行内“投递历史”按钮 + 新 `/deliveries` 页面，不必新建完整 alert detail route。[VERIFIED: frontend/src/App.tsx][VERIFIED: frontend/src/pages/Alerts.tsx][ASSUMED]

## Auth And Audit Constraints

### Authorization
- 当前 `GET /api/v1/deliveries*` 受 `CapabilityViewConfig` 保护，因此 `viewer`、`operator`、`admin` 都能读历史。[VERIFIED: internal/router/router.go][VERIFIED: internal/authz/capabilities.go]
- 当前仓库没有“恢复通知”专用 capability。[VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts]
- 最小变更建议：保留 GET 为 `view_config`，新增 POST recovery 用 `process_alerts`，从而允许 `operator/admin` 恢复、阻止 `viewer` 恢复。[VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts][ASSUMED]

### Auditing
- 现有 `AuditLog` 已能记录 `ActorUserID`、`ActorUsername`、`ActorRole`、`Action`、`TargetType`、`TargetID`、`Result`、`Detail` 和时间。[VERIFIED: internal/models/models.go]
- 现有 `recordAudit` helper 已从 Gin context 提取 principal 并写入数据库，适合继续用于 `delivery.retry` / `delivery.replay` 动作日志。[VERIFIED: internal/handlers/user.go]
- 仅写 `AuditLog` 还不足以替代恢复关联模型，因为 `Detail` 是自由文本，不利于稳定查询“原 delivery 的所有人工恢复及其 resulting delivery”。[VERIFIED: internal/models/models.go][ASSUMED]

## Dependency Risks

### Risk 1: `retry` 无法真正冻结 channel config
当前 `ChannelSnapshot` 只保存 `id/name/type/enabled`，没有保存 `config/secret/api_key`。[VERIFIED: internal/models/notification_delivery.go] 这意味着 `retry` 只能定位回原 channel identity，再用当前 live channel config 执行发送；若 channel 已删除、禁用或被改坏，恢复必须失败并留下审计，而不是偷偷改走别的 channel。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/notifier/notifier.go][ASSUMED]

### Risk 2: `replay` 可能依赖 `Alert.Raw`
当前模板渲染上下文会读取 `alert.Raw` 生成 `.event`，而 delivery snapshot 没有持久化 raw event。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/models/alert.go] 若某些 output template 依赖 `.event.*` 而不是 `.labels` / `.alert.*`，单靠 delivery snapshot 无法完整 replay；因此 Phase 19 规划必须明确 replay 要么回查 `alerts` 表原始告警，要么定义降级行为。[VERIFIED: internal/models/alert.go][VERIFIED: internal/handlers/webhook.go][ASSUMED]

### Risk 3: 双击或并发恢复会产生重复发送
当前仓库没有 recovery 锁或幂等键模型。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/models/models.go] 如果两个请求同时对同一失败 delivery 执行恢复，系统可能产生两条新 delivery 和两次真实通知；planner 应至少考虑“同一原 delivery 同时只允许一个 in-progress recovery”的约束。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]

### Risk 4: 当前前端没有 delivery 类型与路由
`frontend/src/types/index.ts` 没有 delivery 相关类型，`frontend/src/api/client.ts` 没有 delivery API，`frontend/src/App.tsx` 没有 `/deliveries` 路由。[VERIFIED: frontend/src/types/index.ts][VERIFIED: frontend/src/api/client.ts][VERIFIED: frontend/src/App.tsx] Phase 19 的 UI 面不是“补按钮”而是完整新增一个薄页面面。[VERIFIED: .planning/ROADMAP.md][ASSUMED]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 手工恢复审计 | 只靠日志或前端 toast 记录“谁点过按钮” | 继续写 `AuditLog`，并增加结构化恢复关联记录。[VERIFIED: internal/models/models.go][VERIFIED: internal/handlers/user.go][ASSUMED] | 日志和 UI 都不是稳定真源，不能满足 DELV-05。[VERIFIED: .planning/REQUIREMENTS.md] |
| 恢复发送逻辑 | 复制一份新的重试循环和 `notifier` 调用 | 抽取共享执行器，继续复用 `send_attempt` / `terminal_failure` 契约。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/notifier/notifier.go][ASSUMED] | 否则 Phase 15/18 已验证的行为很容易漂移。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-VERIFICATION.md] |
| 前端历史查询 | 在 `Alerts.tsx` 里再拼一张 delivery 大表 | 新建专用 `Deliveries` 页面并复用 AntD table/filter 模式。[VERIFIED: frontend/src/pages/Alerts.tsx][VERIFIED: frontend/src/App.tsx][ASSUMED] | 当前 Alerts 页已经承担 ack/silence 和筛选，继续堆叠会降低可维护性。[VERIFIED: frontend/src/pages/Alerts.tsx] |

**Key insight:** Phase 19 的新增复杂度主要来自“恢复动作的结构化关联与权限边界”，不是“如何把通知再发一次”。现有 notifier、webhook send loop 和 Phase 18 ledger 已经覆盖了后者的大部分机械动作。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/notifier/notifier.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-02-SUMMARY.md]

## Common Pitfalls

### Pitfall 1: 把人工恢复写回原 delivery
**What goes wrong:** 原始历史被污染，用户看不出“这次失败后又被人工重试了几次”。[VERIFIED: internal/models/notification_delivery.go][ASSUMED]  
**Why it happens:** 当前 attempt 表天然挂在 delivery 下，容易误以为人工恢复只是新 attempt。[VERIFIED: internal/models/notification_delivery.go]  
**How to avoid:** 原 delivery 保持只读；人工恢复总是创建新的 delivery，并通过恢复关联模型连回原记录。[VERIFIED: internal/models/notification_delivery.go][ASSUMED]  
**Warning signs:** 设计稿里 `POST /deliveries/:id/retry` 只写 `notification_delivery_attempts`，没有新的 delivery 或 recovery relation。[ASSUMED]

### Pitfall 2: `replay` 直接复用冻结 title/content
**What goes wrong:** 这会让 `replay` 退化成 `retry`，与锁定语义冲突。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Why it happens:** Phase 18 已经有 `RenderedPayloadSnapshot`，实现者会天然想直接复用。[VERIFIED: internal/models/notification_delivery.go]  
**How to avoid:** `replay` 必须重新走当前 datasource/template/route 逻辑；只有 `retry` 复用冻结 payload。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][ASSUMED]  
**Warning signs:** `retry` 和 `replay` 共用同一后端 helper，唯一差别只是 `action` 字符串。[ASSUMED]

### Pitfall 3: 沿用 `view_config` 给恢复 POST 授权
**What goes wrong:** `viewer` 也会获得真实发送能力。[VERIFIED: internal/authz/capabilities.go]  
**Why it happens:** 当前 `/deliveries` 整组就是 `view_config`，加新 POST 很方便。[VERIFIED: internal/router/router.go]  
**How to avoid:** 读写分组或按 route 单独挂 `RequireCapability(CapabilityProcessAlerts)`。[VERIFIED: internal/router/router.go][ASSUMED]  
**Warning signs:** 前端 viewer 在 delivery 页面也能看见 `retry` / `replay` 按钮。[VERIFIED: frontend/src/authz/capabilities.ts][ASSUMED]

### Pitfall 4: 认为 ledger snapshot 足以完整 replay 一切模板
**What goes wrong:** 依赖 `.event` 的模板会因为缺少 `Alert.Raw` 而丢上下文。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/models/alert.go]  
**Why it happens:** Phase 18 把 replay-supporting snapshot 说成“足够支撑 replay”，但当前代码的模板上下文并不只依赖 snapshot 字段。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: internal/handlers/webhook.go]  
**How to avoid:** 把 replay 实现显式设计为“回查原 alert 记录并以当前配置重跑”；查不到原 alert 时返回受控失败和审计。[VERIFIED: internal/models/alert.go][ASSUMED]  
**Warning signs:** replay 代码完全不查 `alerts` 表，也不定义 `.event` 缺失时的错误路径。[ASSUMED]

## Code Examples

Verified patterns from the current codebase:

### 现有 bounded retry + ledger write 契约
```go
for attempt := 1; attempt <= notificationMaxAttempts; attempt++ {
	startedAt := time.Now()
	err := sender(channel, title, content)
	h.recordNotificationAttempt(deliveryRecord, attempt, err, time.Since(startedAt), mode)

	if err == nil {
		h.markNotificationDelivered(deliveryRecord, attempt)
		return
	}

	if !notifier.IsRetryableSendError(err) {
		h.markNotificationFailed(deliveryRecord, attempt, err, false, false)
		return
	}

	if attempt == notificationMaxAttempts {
		h.markNotificationFailed(deliveryRecord, attempt, err, true, true)
		return
	}
}
```
Source: [VERIFIED: internal/handlers/webhook.go]

### 现有审计写入模式
```go
func recordAudit(db *gorm.DB, c *gin.Context, action, targetType, targetID, result, detail string) error {
	entry := models.AuditLog{
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Result:     result,
		Detail:     detail,
	}

	if principal, ok := middleware.GetPrincipal(c); ok {
		entry.ActorUserID = principal.UserID
		entry.ActorUsername = principal.Username
		entry.ActorRole = principal.Role
	}

	return db.Create(&entry).Error
}
```
Source: [VERIFIED: internal/handlers/user.go]

### 现有 delivery 列表筛选边界
```go
input.AlertID = c.Query("alert_id")
input.TraceID = c.Query("trace_id")
input.DeliveryStatus = c.Query("delivery_status")
// channel_id, created_from, created_to, limit, offset
```
Source: [VERIFIED: internal/handlers/delivery.go]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| 通知失败主要靠日志排查。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-VERIFICATION.md] | Phase 18 已把 delivery/attempt/final failure 持久化到 PostgreSQL。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-VERIFICATION.md] | 2026-04-30。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-VERIFICATION.md] | Phase 19 可以在账本上做真实查询和恢复，而不是解析日志。[VERIFIED: .planning/ROADMAP.md] |
| `/deliveries` 只有后端只读 API，没有 UI。[VERIFIED: internal/handlers/delivery.go][VERIFIED: frontend/src/App.tsx] | Phase 19 需要补前端历史页与 deeplink。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED] | 尚未开始。[VERIFIED: .planning/STATE.md] | `OPER-01` / `OPER-04` 目前还不满足。[VERIFIED: .planning/REQUIREMENTS.md] |
| 人工恢复动作不存在。[VERIFIED: internal/router/router.go][VERIFIED: frontend/src/api/client.ts] | Phase 19 将首次开放单条 `retry/replay`。[VERIFIED: .planning/ROADMAP.md] | 尚未开始。[VERIFIED: .planning/STATE.md] | 权限、审计、关联模型都需要第一次设计。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED] |

**Deprecated/outdated:**
- “只要有 delivery 列表 API，Phase 19 就只剩前端收口” 这个判断已经过时，因为 requirements 明确还要求受控恢复和完整审计闭环。[VERIFIED: .planning/REQUIREMENTS.md]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `POST` 恢复动作应复用 `CapabilityProcessAlerts`，而不是新增 capability。 | Summary / Auth And Audit Constraints | 如果产品要把“告警处理”和“通知恢复”权限拆开，planner 需要扩展 authz 与前端角色矩阵。 |
| A2 | 最小可用的审计关联模型应是单独 recovery 记录，而不是只写 `AuditLog`。 | Architecture Patterns / Dependency Risks | 如果决定只用 `AuditLog`，实现会更轻，但 UI 证据链和查询能力会变弱。 |
| A3 | `replay` 需要回查 `alerts` 表的原始 `Raw`，否则部分模板上下文不完整。 | Summary / Dependency Risks / Pitfalls | 如果当前所有模板都只依赖 snapshot 字段，则 replay 可以更轻；但一旦判断错误，会在生产恢复时触发模板缺字段。 |
| A4 | `OPER-04` 的最小实现是 alert 列表 deeplink 到新的 delivery 页面，而不是新增独立 alert detail route。 | Architecture Patterns / API And UI Boundaries | 如果用户坚持先做 alert detail 页，前端任务量会扩大并影响 phase 切片。 |

## Open Questions

1. **`retry` 是否应忽略当前 channel 的 disabled/deleted 状态？**  
   What we know: 当前 snapshot 不保存 channel config secret，只保存 channel identity；真实发送仍要依赖 live channel 记录。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: internal/notifier/notifier.go]  
   What's unclear: 业务是否接受“即便 channel 已禁用，也允许人工 retry 强制发送一次”。[ASSUMED]  
   Recommendation: 默认不要绕过当前禁用/缺失状态，直接受控失败并审计；这是更安全的运维默认值。[ASSUMED]

2. **是否需要把恢复动作设计成同步 HTTP 执行？**  
   What we know: 当前 webhook 发送本来就是同步执行三次尝试，单次恢复 scope 也被限制为 single-item。[VERIFIED: internal/handlers/webhook.go][VERIFIED: .planning/STATE.md]  
   What's unclear: 某些渠道在最坏情况下可能让请求时间接近超时上限。[VERIFIED: frontend/src/api/client.ts][VERIFIED: internal/notifier/notifier.go][ASSUMED]  
   Recommendation: v1 先同步执行，返回 resulting delivery 或 terminal error；不要为了单条恢复提前引入后台 job 系统。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | 后端实现与 `go test`。[VERIFIED: Makefile] | ✓。[VERIFIED: go version] | `go1.25.0`。[VERIFIED: go version] | — |
| Node.js | 前端页面与 Vitest。[VERIFIED: frontend/package.json] | ✓。[VERIFIED: node --version] | `v22.17.0`。[VERIFIED: node --version] | — |
| pnpm | 前端测试和构建脚本。[VERIFIED: frontend/package.json] | ✓。[VERIFIED: pnpm --version] | `10.28.2`。[VERIFIED: pnpm --version] | — |
| Docker | 启动本地 PostgreSQL/Redis。[VERIFIED: docker-compose.yml][VERIFIED: Makefile] | ✓。[VERIFIED: docker --version] | `28.5.1`。[VERIFIED: docker --version] | — |
| PostgreSQL service | 本地联调恢复 API 与完整后端启动。[VERIFIED: internal/config/config.go][VERIFIED: docker-compose.yml] | ✗ 当前 `5432` 未监听。[VERIFIED: TCP 127.0.0.1:5432 probe] | — | 用 Docker Compose 启动本地容器。[VERIFIED: Makefile][VERIFIED: docker-compose.yml] |
| Redis service | 完整后端启动与 webhook path。[VERIFIED: internal/database/redis.go][VERIFIED: docker-compose.yml] | ✗ 当前 `6379` 未监听。[VERIFIED: TCP 127.0.0.1:6379 probe] | — | 用 Docker Compose 启动本地容器；纯 handler/unit 测试可继续用 sqlite/假的 sender 隔离。[VERIFIED: docker-compose.yml][VERIFIED: internal/handlers/delivery_test.go] |
| `psql` CLI | 手工查库与 SQL 调试。[ASSUMED] | ✗。[VERIFIED: psql probe] | — | 用应用测试、容器内客户端或 GORM 测试夹具代替。[ASSUMED] |
| `redis-cli` | 手工探活与 stream 调试。[ASSUMED] | ✗。[VERIFIED: redis-cli probe] | — | 用容器内客户端或应用日志替代。[ASSUMED] |

**Missing dependencies with no fallback:**
- None.[VERIFIED: go version][VERIFIED: node --version][VERIFIED: pnpm --version][VERIFIED: docker --version]

**Missing dependencies with fallback:**
- 本机没有在 PATH 里暴露 `psql` / `redis-cli`，且本地 PostgreSQL/Redis 服务未启动；但 Docker 可用，后端单元测试也不依赖真实服务，因此不阻塞 planning。[VERIFIED: docker --version][VERIFIED: psql probe][VERIFIED: redis-cli probe][VERIFIED: TCP 127.0.0.1:5432 probe][VERIFIED: TCP 127.0.0.1:6379 probe]

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes。[VERIFIED: internal/router/router.go] | 继续复用 JWTAuth，恢复动作不能开匿名入口。[VERIFIED: internal/router/router.go] |
| V3 Session Management | no 新增会话机制。[VERIFIED: frontend/src/api/client.ts][ASSUMED] | 继续沿用现有 token/localStorage 方案。[VERIFIED: frontend/src/api/client.ts] |
| V4 Access Control | yes。[VERIFIED: internal/authz/capabilities.go] | 读写分离：GET 保留 `view_config`，POST recovery 走更强 capability。[VERIFIED: internal/router/router.go][ASSUMED] |
| V5 Input Validation | yes。[VERIFIED: internal/handlers/delivery.go][VERIFIED: internal/models/notification_delivery.go] | 校验 `delivery id`、原因长度、动作枚举、只允许失败 delivery 恢复。[VERIFIED: .planning/ROADMAP.md][ASSUMED] |
| V6 Cryptography | no 新增密码学实现。[VERIFIED: internal/notifier/notifier.go][ASSUMED] | 不复制 secret 到 ledger/recovery 审计模型。[VERIFIED: internal/models/notification_delivery.go] |

### Known Threat Patterns for Go + Gin + GORM Recovery Flows

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| viewer 越权触发恢复 | Elevation of Privilege | 恢复 POST 单独 capability；前后端都隐藏并强校验。[VERIFIED: internal/authz/capabilities.go][VERIFIED: frontend/src/authz/capabilities.ts][ASSUMED] |
| 人工恢复无证据链 | Repudiation | 每次恢复都写 `AuditLog` 和结构化 recovery relation。[VERIFIED: internal/models/models.go][ASSUMED] |
| 重复点击触发双发 | Tampering / DoS | 服务端拒绝同一原 delivery 的并发恢复，前端按钮请求中禁用。[ASSUMED] |
| 用当前 live route/template 冒充 retry | Integrity | 明确区分 `retry` 与 `replay` 的执行路径和 `trigger_kind`。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: internal/models/notification_delivery.go][ASSUMED] |
| 在账本或审计中泄露 channel secret | Information Disclosure | 继续只保存 channel identity，不把 `config/api_key/secret` 暴露给 delivery UI。[VERIFIED: internal/models/notification_delivery.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-SECURITY.md] |

## Verification Strategy

- 后端服务回归应覆盖四类真相：`1)` 只有失败 delivery 能恢复，`2)` `retry` 会创建新的 delivery 且 payload 与原记录冻结快照一致，`3)` `replay` 会重新 route/render 并把 resulting delivery 连回原记录，`4)` 无论成功失败都能留下 actor/reason/result 审计证据。[VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]
- 路由/鉴权回归应证明：viewer 仍可读历史但不能恢复，operator/admin 可恢复，未登录用户得到 `401`。[VERIFIED: internal/authz/capabilities.go][VERIFIED: internal/handlers/delivery_test.go][ASSUMED]
- 前端回归至少要覆盖：`Alerts` 页出现“投递历史”跳转、viewer 看不到恢复按钮、operator 能打开原因 modal 并收到结果反馈。[VERIFIED: frontend/src/pages/Alerts.test.tsx][ASSUMED]
- 当前基线已通过 `go test ./internal/handlers -run "Test(DeliveryHandler|RouterDeliveriesAuthorization)" -count=1` 与 `pnpm --dir frontend test -- --run Alerts.test.tsx`，说明 Phase 18 delivery 读面和 Alerts 页权限基线可作为 Phase 19 的稳定起点。[VERIFIED: go test ./internal/handlers -run "Test(DeliveryHandler|RouterDeliveriesAuthorization)" -count=1][VERIFIED: pnpm --dir frontend test -- --run Alerts.test.tsx]

## Sources

### Primary (HIGH confidence)
- `AGENTS.md` - 项目约束、技术栈、brownfield 和前端兼容性要求
- `.planning/REQUIREMENTS.md` - `DELV-03` / `DELV-04` / `DELV-05` / `OPER-01` / `OPER-04`
- `.planning/STATE.md` - Phase 19 当前定位与“单条恢复、不做批量”的里程碑上下文
- `.planning/ROADMAP.md` - Phase 19 goal、success criteria、UI hint
- `.planning/phases/18-establish-delivery-ledger/18-CONTEXT.md` - `retry` / `replay` 语义边界
- `.planning/phases/18-establish-delivery-ledger/18-01-SUMMARY.md` - ledger schema / snapshot / delivery service
- `.planning/phases/18-establish-delivery-ledger/18-02-SUMMARY.md` - webhook 热路径账本接入
- `.planning/phases/18-establish-delivery-ledger/18-03-SUMMARY.md` - delivery 只读 API 与 auth 接入
- `.planning/phases/18-establish-delivery-ledger/18-VERIFICATION.md` - Phase 18 已验证真相
- `.planning/phases/18-establish-delivery-ledger/18-SECURITY.md` - delivery snapshot 安全边界
- `.planning/config.json` - `workflow.nyquist_validation=false`
- `internal/models/notification_delivery.go` - delivery/attempt schema、snapshot 字段、`TriggerKind`
- `internal/models/alert.go` - `Alert.Raw`、`AlertID` 主键、template replay 相关上下文
- `internal/models/models.go` - `AuditLog`
- `internal/delivery/service.go` - delivery service 读写能力
- `internal/handlers/delivery.go` - 当前历史筛选与详情响应
- `internal/handlers/webhook.go` - 当前发送执行路径与 ledger 热路径
- `internal/handlers/user.go` - `recordAudit`
- `internal/notifier/notifier.go` - 发送器与 retryable error 判定
- `internal/router/router.go` - `/deliveries` 路由与 capability 保护
- `internal/authz/capabilities.go` / `frontend/src/authz/capabilities.ts` - 当前角色矩阵
- `frontend/src/App.tsx` - 当前前端路由
- `frontend/src/api/client.ts` - 当前 API 边界
- `frontend/src/pages/Alerts.tsx` / `frontend/src/pages/Alerts.test.tsx` - 当前告警页与权限测试
- `frontend/package.json` / `Makefile` / `docker-compose.yml` - 工具链、脚本、服务依赖
- `go version` / `node --version` / `pnpm --version` / `docker --version` - 本机环境探针
- `go test ./internal/handlers -run "Test(DeliveryHandler|RouterDeliveriesAuthorization)" -count=1` - 后端基线
- `pnpm --dir frontend test -- --run Alerts.test.tsx` - 前端基线

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - 全部来自仓库现状与本机环境探针，没有依赖外部过时假设。
- Architecture: HIGH - 主要结论直接建立在 Phase 18 context、现有 webhook 路径和 delivery API 上。
- Pitfalls: MEDIUM-HIGH - 大部分风险可从现有代码直接验证，少数关于 capability 复用和 recovery relation 的建议带 `[ASSUMED]` 标记。

**Research date:** 2026-04-30  
**Valid until:** 2026-05-30
