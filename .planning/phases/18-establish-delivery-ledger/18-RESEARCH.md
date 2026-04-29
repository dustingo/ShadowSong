# Phase 18: Establish Delivery Ledger - Research

**Researched:** 2026-04-29  
**Domain:** Go + Gin + GORM + PostgreSQL 通知投递账本  
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions [VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
### Ledger Data Model
- **D-01:** 账本采用两层模型：`notification_deliveries` 作为每条 `alert x channel` 的主记录，`notification_delivery_attempts` 作为 append-only attempt 明细。
- **D-02:** 必须持久化成功与失败两类投递，而不是只存最终失败；失败视图只是账本上的一个筛选结果，不单独建“失败表”真源。
- **D-03:** 现有应用内 bounded 3 次即时重试继续保留，Phase 18 先在发送热路径旁路落库，不引入 MQ、外部 worker 或 workflow engine。

### Snapshot Contract
- **D-04:** 每条 delivery 记录冻结审计与原始重试所需的不可变快照，至少包括：告警关键字段、渠道身份快照、路由身份快照、发送模式、最终实际发送内容，以及终态失败摘要。
- **D-05:** 首版快照优先保证“可审计、可解释、可支撑原始 retry”，不把完整渠道密钥、全量运行时配置或重型原始回执作为必存范围。

### Recovery Semantics Boundary
- **D-06:** `retry` 语义锁定为沿原始发送语义再次执行，基于账本记录与冻结快照完成，不重新走当前策略。
- **D-07:** `replay` 语义锁定为重新走当前策略，而不是复用原始 route/template 结果；因此 `replay` 的安全动作与完整 API/审计闭环继续属于 Phase 19。
- **D-08:** Phase 18 只需要把账本设计成足以同时支撑上述两种未来语义，不在本 phase 内开放人工恢复入口。

### Read Surface
- **D-09:** 为满足“维护者可以查看任一通知投递账本记录”，Phase 18 提供最小只读能力即可，优先是后端详情/查询 API 或等价只读入口，不强制在本 phase 同时交付完整前端历史页。
- **D-10:** 完整历史列表、搜索筛选、运维健康聚合页和人工恢复操作继续放在 Phase 19/21，避免 Phase 18 与后续 phase 边界重叠。

### Claude's Discretion [VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- 具体表字段命名、状态枚举细节、最小索引集合和 GORM 迁移组织方式可由后续 researcher/planner 在不违背上述契约的前提下决定。
- 最小读能力是先做单条详情 API 还是带有限过滤的只读列表 API，可按最小满足 success criteria 与改动面控制原则裁定。

### Deferred Ideas (OUT OF SCOPE) [VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- 完整 delivery history 页面、复杂搜索筛选和运维健康视图 — 留给 Phase 19/21。
- 人工 `retry/replay` API、原因录入、操作者审计闭环 — 留给 Phase 19。
- 批量恢复、复杂 replay preview、更多策略重算模式 — 继续 deferred，不纳入本 phase。
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DELV-01 | 维护者可以查看每次通知投递的持久化记录，包括告警、渠道、发送模式、尝试次数、最终结果和失败原因 [VERIFIED: .planning/REQUIREMENTS.md] | 单条详情 API + `notification_deliveries`/`notification_delivery_attempts` 两层模型可直接覆盖该字段面。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: internal/router/router.go][CITED: https://gin-gonic.com/en/docs/middleware/using-middleware/] |
| DELV-02 | 通知在超过当前即时重试上限后，会把最终失败结果持久化保存，而不只停留在日志中 [VERIFIED: .planning/REQUIREMENTS.md] | 在 `sendChannelNotification` 的重试循环旁路写 attempt，并在现有 `terminal_failure` 分支同步更新 delivery 终态即可满足，不需要改动重试上限语义。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/handlers/webhook_test.go][VERIFIED: docs/alert-path-operations-runbook.md] |
| DELV-06 | 首版投递记录会保存足够支撑审计和单条 replay 的不可变快照，而不是只依赖当前实时配置 [VERIFIED: .planning/REQUIREMENTS.md] | 使用 PostgreSQL `jsonb`/`datatypes.JSON` 冻结 alert、channel、route、render result 和 failure summary 快照，避免后续配置漂移导致账本失真。[VERIFIED: internal/models/alert.go][VERIFIED: go.mod][CITED: https://gorm.io/docs/data_types.html][CITED: https://www.postgresql.org/docs/current/datatype-json.html] |
</phase_requirements>

## Summary

Phase 18 的实现核心不是“补一个失败表”，而是把当前只存在于 `internal/handlers/webhook.go` 发送热路径与日志中的投递生命周期提升为 PostgreSQL 真源。[VERIFIED: internal/handlers/webhook.go][VERIFIED: docs/alert-path-operations-runbook.md] 当前系统已经具备稳定的 `trace_id`、3 次 bounded retry、`send_attempt` / `send_notification` / `terminal_failure` 日志契约；账本必须增量接入这些已验证行为，而不是重写通知架构。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/handlers/webhook_test.go][VERIFIED: docs/alert-path-operations-runbook.md]

最稳妥的方案是遵循 phase context 已锁定的双表模型：`notification_deliveries` 记录每个 `alert x channel` 的主投递实体，`notification_delivery_attempts` 记录 append-only 尝试明细，并把不可变快照放在 delivery 主表中。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] 这样可以在不引入 MQ、worker 或新服务层改造的前提下，满足持久化成功/失败、审计解释、未来单条 `retry/replay` 三个方向。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: .planning/ROADMAP.md]

**Primary recommendation:** 使用新增的聚焦后端域服务封装账本写入，但把接入点保持在现有 `WebhookHandler -> sendNotification -> sendChannelNotification` 链路中，先交付单条详情只读 API，再把完整历史与人工恢复留给后续 phase。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/router/router.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][ASSUMED]

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.25.0 | 现有后端运行时与测试执行环境。[VERIFIED: go.mod][VERIFIED: go version] | 仓库已固定该 toolchain，Phase 18 不需要语言层迁移。[VERIFIED: AGENTS.md] |
| Gin | v1.12.0 | 继续承载 webhook 入口与新增 delivery 只读 API。[VERIFIED: go.mod][VERIFIED: go list -m github.com/gin-gonic/gin] | 当前所有 HTTP surface 都已在 Gin route groups 与 middleware 上实现；官方文档支持 group 级 middleware，适合把读 API 接到现有鉴权组中。[VERIFIED: internal/router/router.go][CITED: https://gin-gonic.com/en/docs/middleware/using-middleware/] |
| GORM | v1.31.1 | 定义 ledger 模型、迁移表结构、执行查询与更新。[VERIFIED: go.mod][VERIFIED: go list -m gorm.io/gorm] | 仓库已用 `Migrator` + `AutoMigrate` 做增量迁移；官方文档明确 `AutoMigrate`/index tag 适合当前这种 additive schema 扩展。[VERIFIED: internal/database/postgres.go][CITED: https://gorm.io/docs/migration.html][CITED: https://gorm.io/docs/indexes.html] |
| PostgreSQL | 14-alpine | 投递账本真源与不可变快照存储。[VERIFIED: docker-compose.yml] | 项目本地基线已经是 PostgreSQL 14；官方文档说明 `jsonb` 适合结构化快照与后续键路径查询。[VERIFIED: docker-compose.yml][CITED: https://www.postgresql.org/docs/current/datatype-json.html] |
| `gorm.io/datatypes` | v1.2.7 | 在 Go 模型中承载 `jsonb` 快照字段。[VERIFIED: go.mod] | 当前 `Alert.Raw`、`Labels`、`Channel.Config`、`RouteRule.*` 已大量使用 `datatypes.JSON`，复用该模式最符合仓库现状。[VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go][CITED: https://gorm.io/docs/data_types.html] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `gorm.io/driver/postgres` | v1.6.0 | 继续作为 GORM 的 PostgreSQL driver。[VERIFIED: go.mod][VERIFIED: go list -m gorm.io/driver/postgres] | 账本是同一数据库内的 additive schema，无需新 driver。[VERIFIED: internal/database/postgres.go] |
| `github.com/redis/go-redis/v9` | v9.18.0 | 保持现有 Redis stream 发布，不承担账本真源。[VERIFIED: go.mod][VERIFIED: go list -m github.com/redis/go-redis/v9] | Phase 18 只在通知热路径旁路落库；Redis 继续做现有 ephemeral transport。[VERIFIED: internal/handlers/webhook.go][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] |
| `github.com/stretchr/testify` | v1.11.1 | 承载新增 handler/model 回归测试断言。[VERIFIED: go.mod][VERIFIED: go list -m github.com/stretchr/testify] | 仓库现有关键回归均基于 testify；Phase 18 应沿用相同测试风格。[VERIFIED: internal/handlers/webhook_test.go] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| PostgreSQL 双表账本 | Redis stream / 临时内存状态 | 不能满足持久化审计与单条查询要求，且与当前 requirement 冲突。[VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] |
| 现有热路径旁路落库 | MQ / worker / workflow engine | 会扩大 phase 范围，直接违反 locked decision D-03 与 requirements out-of-scope。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][VERIFIED: .planning/REQUIREMENTS.md] |
| GORM additive migration | 独立迁移框架切换 | 当前仓库已稳定使用 `Migrator`；引入迁移框架会增加非目标变更面。[VERIFIED: internal/database/postgres.go][CITED: https://gorm.io/docs/migration.html] |

**Installation:** 本 phase 预计不需要新增第三方依赖；推荐直接复用仓库已锁定依赖集。[VERIFIED: go.mod][ASSUMED]

**Version verification:** `go list -m -json` 已确认当前仓库依赖版本与 upstream 发布日期：Gin `v1.12.0`（2026-02-28）、GORM `v1.31.1`（2025-11-02）、`gorm.io/driver/postgres v1.6.0`（2025-05-27）、`go-redis/v9 v9.18.0`（2026-02-16）、`testify v1.11.1`（2025-08-27）。[VERIFIED: go list -m github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/postgres github.com/redis/go-redis/v9 github.com/stretchr/testify]

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── delivery/                # 新增：账本写入/查询聚焦域服务
├── handlers/
│   ├── delivery.go          # 新增：最小只读 API
│   └── webhook.go           # 修改：接入 delivery service
├── models/
│   ├── notification_delivery.go
│   └── notification_delivery_attempt.go
└── router/
    └── router.go            # 新增 delivery 路由
```

### Pattern 1: Delivery 主记录 + Attempt 明细
**What:** 每个 `alert x channel` 创建一条 delivery 主记录，每次实际发送写一条 attempt，attempt 永不原地覆盖。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**When to use:** 当前 bounded retry、后续人工 retry、需要审计单次发送细节时都使用该模式。[VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: docs/alert-path-operations-runbook.md]  
**Example:**
```go
type NotificationDelivery struct {
  ID                     uint           `gorm:"primaryKey" json:"id"`
  AlertID                string         `gorm:"index;size:64;not null" json:"alert_id"`
  TraceID                string         `gorm:"index;size:64;not null" json:"trace_id"`
  ChannelID              uint           `gorm:"index;not null" json:"channel_id"`
  RouteRuleID            *uint          `gorm:"index" json:"route_rule_id,omitempty"`
  DeliveryStatus         string         `gorm:"index;size:32;not null" json:"delivery_status"`
  DeliveryMode           string         `gorm:"size:32;not null" json:"delivery_mode"`
  AttemptCount           int            `gorm:"not null;default:0" json:"attempt_count"`
  FinalFailureSummary    string         `gorm:"type:text" json:"final_failure_summary"`
  AlertSnapshot          datatypes.JSON `gorm:"type:jsonb;not null" json:"alert_snapshot"`
  ChannelSnapshot        datatypes.JSON `gorm:"type:jsonb;not null" json:"channel_snapshot"`
  RouteSnapshot          datatypes.JSON `gorm:"type:jsonb" json:"route_snapshot"`
  RenderedPayloadSnapshot datatypes.JSON `gorm:"type:jsonb;not null" json:"rendered_payload_snapshot"`
  CreatedAt              time.Time      `json:"created_at"`
  UpdatedAt              time.Time      `json:"updated_at"`
}
```
Source: [VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][CITED: https://gorm.io/docs/data_types.html][CITED: https://gorm.io/docs/indexes.html]

### Pattern 2: 在现有发送循环内旁路持久化
**What:** 不改 `notificationMaxAttempts = 3` 与 `notifier.IsRetryableSendError` 判定，只在 attempt 前后补 DB 持久化。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/handlers/webhook_test.go]  
**When to use:** Phase 18 的目标是 ledger-first，而不是行为重写。[VERIFIED: .planning/ROADMAP.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Example:**
```go
for attempt := 1; attempt <= notificationMaxAttempts; attempt++ {
  startedAt := time.Now()
  err := sender(channel, title, content)

  deliveryService.RecordAttempt(ctx, deliveryID, AttemptInput{
    AttemptNumber: attempt,
    Result:        resultFromErr(err),
    ErrorMessage:  errorString(err),
    DurationMS:    time.Since(startedAt).Milliseconds(),
  })

  if err == nil {
    deliveryService.MarkDelivered(ctx, deliveryID, attempt)
    return
  }

  if !notifier.IsRetryableSendError(err) {
    deliveryService.MarkFailed(ctx, deliveryID, attempt, err, false)
    return
  }

  if attempt == notificationMaxAttempts {
    deliveryService.MarkFailed(ctx, deliveryID, attempt, err, true)
    return
  }
}
```
Source: [VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/handlers/webhook_test.go][ASSUMED]

### Pattern 3: 最小只读面优先做单条详情
**What:** 在现有 JWT + capability 保护的 `/api/v1` 组下新增 `GET /api/v1/deliveries/:id`，返回主记录与 attempts 明细。[VERIFIED: internal/router/router.go][VERIFIED: internal/middleware/authorize.go]  
**When to use:** Context 已明确最小只读面即可，完整历史页和复杂筛选不在本 phase。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Example:**
```go
deliveries := v1.Group("/deliveries")
deliveries.Use(middleware.JWTAuth(jwtAuth, db))
{
  deliveries.GET("/:id",
    middleware.RequireCapability(authz.CapabilityViewConfig),
    deliveryHandler.Get,
  )
}
```
Source: [VERIFIED: internal/router/router.go][CITED: https://gin-gonic.com/en/docs/middleware/using-middleware/]

### Anti-Patterns to Avoid
- **只建“失败表”:** 会丢掉成功样本、attempt 序列与后续成功率统计能力，直接违背 D-02。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
- **把账本建立在日志解析之上:** 当前日志契约已稳定，但 requirement 要求的是持久化真源，不是可搜索日志替代品。[VERIFIED: docs/alert-path-operations-runbook.md][VERIFIED: .planning/REQUIREMENTS.md]
- **在 model hook 内做跨表账本写入:** GORM hooks 会随 create/update 自动触发并在错误时回滚事务，更适合字段默认值和模型内校验；跨实体投递流程状态机应放在显式 service 中，避免 alert 保存、副作用和发送尝试纠缠。[VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go][CITED: https://gorm.io/docs/hooks.html][ASSUMED]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON 快照存储 | 自定义字符串拼接或扁平字段爆炸 | PostgreSQL `jsonb` + `datatypes.JSON` | 官方文档支持结构化 JSONB 存储与后续路径查询；当前仓库也已在多个模型使用该模式。[VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go][CITED: https://gorm.io/docs/data_types.html][CITED: https://www.postgresql.org/docs/current/datatype-json.html] |
| 索引创建 | 手写零散 SQL 索引脚本作为首选 | GORM `index` / `uniqueIndex` tag + `Migrator` | 当前仓库迁移入口已基于 `Migrator`，官方文档说明索引标签可随 `AutoMigrate`/`CreateTable` 生效。[VERIFIED: internal/database/postgres.go][CITED: https://gorm.io/docs/indexes.html][CITED: https://gorm.io/docs/migration.html] |
| 权限控制 | 新增独立权限系统 | 复用 `CapabilityViewConfig` 保护最小读 API | 现有 capability matrix 已覆盖配置类只读能力，新增权限模型会扩大 scope。[VERIFIED: internal/authz/capabilities.go][VERIFIED: internal/router/router.go][ASSUMED] |

**Key insight:** Phase 18 的复杂度来自“语义冻结”，不是基础设施能力缺失；复用现有 PostgreSQL、GORM、Gin、authz seam 足以完成本 phase。[VERIFIED: internal/database/postgres.go][VERIFIED: internal/router/router.go][VERIFIED: .planning/ROADMAP.md]

## Common Pitfalls

### Pitfall 1: 只在终态失败时插入一行
**What goes wrong:** attempt 历史、成功投递样本和最终 `attempt_count` 都会失真。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Why it happens:** 容易把 Phase 18 误解为“补失败审计”，而不是“建立完整投递账本”。[VERIFIED: .planning/REQUIREMENTS.md]  
**How to avoid:** 每次发送都创建 delivery 主记录，并对每次真实 send 写 append-only attempt。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Warning signs:** 设计稿里只有 `final_error`、没有 `attempt_number` 或成功终态字段。[ASSUMED]

### Pitfall 2: 快照只存外键，不存冻结内容
**What goes wrong:** channel、route、template 后续变更后，账本无法解释“当时为什么这么发”。[VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Why it happens:** 误把 replay 支撑理解成“能重新查当前配置”。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**How to avoid:** 至少冻结 alert 关键字段、channel identity、route identity、mode、最终实际 title/content、终态失败摘要。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Warning signs:** schema 只有 `channel_id` / `route_rule_id`，没有任何 snapshot 字段。[ASSUMED]

### Pitfall 3: 把账本写入塞进 `Alert` hook
**What goes wrong:** 告警持久化、去重、异步通知、账本 side effect 之间边界会被混淆，失败回滚范围也不清晰。[VERIFIED: internal/handlers/webhook.go][CITED: https://gorm.io/docs/hooks.html]  
**Why it happens:** 当前模型已经用 `BeforeCreate/BeforeUpdate` 做校验，容易顺手扩展。[VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go]  
**How to avoid:** hook 继续只做模型级校验/默认值；delivery 生命周期写入放在显式 service/helper 中。[VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go][ASSUMED]  
**Warning signs:** `Alert.BeforeCreate` 或 `Channel.BeforeUpdate` 开始读取其他表或发送通知。[ASSUMED]

### Pitfall 4: 改动现有重试语义
**What goes wrong:** 会破坏 Phase 15 已验证的 bounded retry 与 `terminal_failure` 契约。[VERIFIED: docs/alert-path-operations-runbook.md][VERIFIED: internal/handlers/webhook_test.go]  
**Why it happens:** 在做账本时顺手重构 `sendChannelNotification` 的流程控制。[VERIFIED: internal/handlers/webhook.go]  
**How to avoid:** 先让数据库记录跟随现有循环，再考虑后续 phase 的人工恢复行为。[VERIFIED: .planning/ROADMAP.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]  
**Warning signs:** `notificationMaxAttempts`、`IsRetryableSendError`、`terminal_failure` 日志断言同时变化。[VERIFIED: internal/handlers/webhook.go][VERIFIED: internal/handlers/webhook_test.go]

## Code Examples

Verified patterns from official and local sources:

### PostgreSQL JSONB Snapshot Field
```go
type NotificationDelivery struct {
  AlertSnapshot   datatypes.JSON `gorm:"type:jsonb;not null" json:"alert_snapshot"`
  ChannelSnapshot datatypes.JSON `gorm:"type:jsonb;not null" json:"channel_snapshot"`
  RouteSnapshot   datatypes.JSON `gorm:"type:jsonb" json:"route_snapshot"`
}
```
Source: [VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go][CITED: https://gorm.io/docs/data_types.html]

### GORM Index Tags For Ledger Lookup Paths
```go
type NotificationDelivery struct {
  AlertID        string `gorm:"index;size:64;not null"`
  TraceID        string `gorm:"index;size:64;not null"`
  ChannelID      uint   `gorm:"index;not null"`
  DeliveryStatus string `gorm:"index;size:32;not null"`
  CreatedAt      time.Time `gorm:"index"`
}
```
Source: [CITED: https://gorm.io/docs/indexes.html][VERIFIED: internal/models/alert.go]

### Existing GORM Migration Seam
```go
tables := []interface{}{
  &models.User{},
  &models.Alert{},
  &models.DataSource{},
  &models.Channel{},
  &models.RouteRule{},
  &models.SilenceRule{},
  &models.AuditLog{},
  &models.OnDuty{},
}

for _, table := range tables {
  if !migrator.HasTable(table) {
    _ = migrator.CreateTable(table)
  } else {
    _ = migrator.AutoMigrate(table)
  }
}
```
Source: [VERIFIED: internal/database/postgres.go][CITED: https://gorm.io/docs/migration.html]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `terminal_failure` 只写日志 | 以日志为诊断面，同时补 PostgreSQL 账本真源 | Phase 18 目标，当前尚未实现。[VERIFIED: .planning/ROADMAP.md][VERIFIED: docs/alert-path-operations-runbook.md] | 运维不再需要把最终失败证据只寄托在日志检索上。[VERIFIED: .planning/REQUIREMENTS.md] |
| 只凭当前配置理解历史投递 | 用 immutable snapshot 冻结历史投递语义 | Phase 18 已被 requirement 锁定。[VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] | 后续单条 `retry/replay` 可以分别基于“原始语义”与“当前策略”做清晰分层。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] |
| Redis/日志承担大部分链路可见性 | PostgreSQL 账本成为恢复真源，Redis 保持 ephemeral | v1.4 路线图已明确。[VERIFIED: .planning/ROADMAP.md][VERIFIED: internal/handlers/webhook.go] | 为 Phase 19/21 的查询、审计和聚合提供稳定基础。[VERIFIED: .planning/ROADMAP.md] |

**Deprecated/outdated:**
- “先做前端历史页，再决定后端数据契约” 这条做法不适合本项目，因为 context 已经把最小只读面放在后端优先，而完整历史页明确 deferred。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | 最小只读能力优先做单条详情 API，比带过滤列表 API 更符合当前最小改动原则。 | Summary / Architecture Patterns | 如果 planner 更需要列表 API，任务拆分会调整，但账本 schema 与热路径接入基本不变。 |
| A2 | Phase 18 不需要新增第三方依赖即可完成。 | Standard Stack | 如果实现选择引入 repository helper、UUID 库或结构化错误库，安装与测试任务需要补充。 |
| A3 | 复用 `CapabilityViewConfig` 保护 delivery 只读 API 足够，不需要新增 capability。 | Don't Hand-Roll | 如果产品希望“可看配置但不可看投递账本”分离权限，router/authz 任务会扩大。 |

## Open Questions (RESOLVED)

1. **最小只读面到底是单条详情还是详情 + 极简列表？**
   What we know: context 明确“不强制完整前端历史页”，并允许 researcher/planner 在单条详情与有限过滤列表间裁定。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
   RESOLVED: 采用 `GET /api/v1/deliveries` + `GET /api/v1/deliveries/:id` 的最小后端只读面；列表只提供 `alert_id`、`trace_id`、`channel_id`、`delivery_status`、时间范围和分页这些极简过滤，不扩成完整历史页或复杂搜索。这样既满足“维护者可以查看任一通知投递记录”的 success criteria，也不吞掉 Phase 19 的完整历史面 scope。[VERIFIED: .planning/ROADMAP.md][VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]

2. **快照字段是否要拆成多个 JSONB 列，还是合并成一个 payload？**
   What we know: requirement 只锁定快照内容范围，没有锁定列布局。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md]
   RESOLVED: 快照拆成 `alert_snapshot`、`channel_snapshot`、`route_snapshot`、`rendered_payload_snapshot` 四类 `jsonb` 列；不合并成单个 blob，也不为了少量字段过早完全结构化成大量列。这样更利于后续审计解释、字段演进和最小读 API 输出，同时维持当前 phase 的 additive schema 复杂度可控。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md][ASSUMED]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | backend build/test | ✓ | 1.25.0 | — |
| Node.js | 前端类型契约或后续只读页联调 | ✓ | v22.17.0 | — |
| pnpm | 前端依赖安装/构建 | ✓ | 10.28.2 | — |
| Docker | 本地 PostgreSQL/Redis 容器 | ✓ | 28.5.1 | — |
| PostgreSQL CLI (`psql`) | 本地手工查账本/调试 SQL | ✗ | — | 使用 Docker 容器内 `psql` 或应用测试覆盖 |
| Redis CLI/server binary | 本地直接探活/调试 | ✗ | — | 使用 Docker `redis:7-alpine` 容器 |

**Missing dependencies with no fallback:**
- None.[VERIFIED: go version][VERIFIED: node --version][VERIFIED: pnpm --version][VERIFIED: docker --version]

**Missing dependencies with fallback:**
- `psql` 与 `redis-server` 未安装在宿主机 PATH，但仓库已有 Docker Compose 的 PostgreSQL 14 / Redis 7 服务定义，可作为本地联调替代。[VERIFIED: psql --version][VERIFIED: redis-server --version][VERIFIED: docker-compose.yml]

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | 复用现有 JWT + `middleware.JWTAuth` 保护 `/api/v1/deliveries/*`。[VERIFIED: internal/router/router.go] |
| V3 Session Management | no | 本 phase 不新增会话机制；继续沿用现有 JWT 流程。[VERIFIED: internal/router/router.go][ASSUMED] |
| V4 Access Control | yes | 复用 `RequireCapability` + `CapabilityViewConfig` 做最小只读授权。[VERIFIED: internal/middleware/authorize.go][VERIFIED: internal/authz/capabilities.go][ASSUMED] |
| V5 Input Validation | yes | 对 path/query 参数沿用当前 handler 解析和错误返回模式；对 snapshot/status 枚举在 model/service 中做显式校验。[VERIFIED: internal/handlers/alert.go][VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go][ASSUMED] |
| V6 Cryptography | no | Phase 18 不新增密码学实现，也不应把 channel secret 全量复制进快照。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] |

### Known Threat Patterns for Go + Gin + GORM Delivery Ledger

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| 未授权读取投递证据 | Information Disclosure | 所有 delivery 读 API 走现有 JWT 和 capability middleware。[VERIFIED: internal/router/router.go][VERIFIED: internal/middleware/authorize.go] |
| 用快照持久化敏感 secret | Information Disclosure | 仅冻结渠道身份、类型、发送内容、失败摘要，不复制 webhook secret/API key。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] |
| 手工拼接 SQL/JSON 导致查询风险 | Tampering | 继续使用 GORM 查询与 `datatypes.JSON`，不手写动态 SQL 字符串作为主路径。[VERIFIED: internal/database/postgres.go][VERIFIED: internal/models/alert.go][CITED: https://gorm.io/docs/data_types.html][ASSUMED] |
| 账本记录被更新覆盖导致审计链断裂 | Repudiation | 主 delivery 只更新聚合终态字段，attempt 明细 append-only，不原地改写历史 attempt。[VERIFIED: .planning/phases/18-establish-delivery-ledger/18-CONTEXT.md] |

## Sources

### Primary (HIGH confidence)
- `.planning/phases/18-establish-delivery-ledger/18-CONTEXT.md` - phase 边界、锁定决策、deferred scope
- `.planning/REQUIREMENTS.md` - `DELV-01` / `DELV-02` / `DELV-06` 真源
- `.planning/ROADMAP.md` - success criteria 与 phase 依赖
- `docs/alert-path-operations-runbook.md` - 当前 `trace_id` / retry / `terminal_failure` 契约
- `internal/handlers/webhook.go` - 通知热路径、重试循环、当前落点
- `internal/handlers/webhook_test.go` - 已验证重试和日志回归基线
- `internal/database/postgres.go` - 当前迁移接缝
- `internal/router/router.go` - 现有 API 分组与鉴权接缝
- `internal/authz/capabilities.go` / `internal/middleware/authorize.go` - 现有授权矩阵
- `internal/models/alert.go` / `internal/models/models.go` - 现有 JSON、hook、validation 模式
- `go list -m -json ...` - 当前 Go 依赖版本和 upstream 时间
- `go version` / `node --version` / `pnpm --version` / `docker --version` - 本机环境探针
- `docker-compose.yml` - PostgreSQL 14 / Redis 7 本地服务基线
- `go test ./internal/handlers -run "TestWebhookHandler(.*Retry.*|.*Terminal.*|.*Logging.*)" -count=1` - 关键通知回归当前可通过
- `go test ./internal/router ./internal/authz -count=1` - 路由与权限基线当前可通过
- https://gorm.io/docs/migration.html - `AutoMigrate` / `Migrator` 官方说明
- https://gorm.io/docs/indexes.html - GORM 索引标签官方说明
- https://gorm.io/docs/data_types.html - GORM 自定义数据类型与 PostgreSQL `JSONB`
- https://gorm.io/docs/hooks.html - GORM hooks 事务与回滚行为
- https://gin-gonic.com/en/docs/middleware/using-middleware/ - Gin group/per-route middleware 官方说明
- https://www.postgresql.org/docs/current/datatype-json.html - PostgreSQL `jsonb` 类型与索引说明

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - 全部基于当前仓库依赖、环境探针和官方文档交叉验证。
- Architecture: HIGH - 接入点、迁移方式和权限复用都直接来自现有代码结构与 locked decisions。
- Pitfalls: MEDIUM-HIGH - 大部分由现有代码/requirements 可直接推出，少量实现建议仍带明确 `[ASSUMED]` 标签。

**Research date:** 2026-04-29  
**Valid until:** 2026-05-29
