# Architecture Research: v1.4 Delivery Recovery and Production Hardening

**Scope:** 仅覆盖新能力如何接入现有架构
**Researched:** 2026-04-29
**Confidence:** HIGH

## Integration Summary

现有主链路已经是 `webhook -> Alert 持久化 -> Redis stream -> async route match -> channel send`，且 `trace_id`、重试日志、统一日志契约已经存在。v1.4 不应引入新服务、新队列或全局架构迁移，而应在现有 Gin + GORM + PostgreSQL + Redis 应用内增加一条“可持久化、可查询、可重放”的通知投递子链路。

推荐把新增能力收敛到两个新边界：

1. **Delivery Recovery 子域**：把每次“某条告警发往某个渠道”的投递结果持久化，并允许手工重试/重放。
2. **Operations Health 子域**：把 readiness、metrics、入口保护和历史/健康查询挂到现有 Gin 路由与前端控制台中。

核心原则：

- 不改现有 webhook 接入协议，不替换 Redis，不引入外部 MQ。
- 先做“旁路记录 + 只读可观测”，再做“人工补发/重试”。
- `/health` 保持轻量 liveness；新增独立 readiness 与 metrics，不把浅健康检查改成重探针。

## Recommended Architecture

### New Components

| Component | Type | Responsibility | Why here |
|-----------|------|----------------|----------|
| `internal/models/notification_delivery.go` | New model | 持久化单次投递记录，主键可用自增 ID，关联 `alert_id`、`trace_id`、`channel_id`、`route_rule_id`、状态、最后错误、最后发送时间、触发方式 | 现有 `Alert` 只记录告警真相，不适合塞入多渠道投递生命周期 |
| `internal/models/notification_delivery_attempt.go` | New model | 持久化每次 send attempt / replay attempt，记录 attempt 序号、结果、错误、耗时、HTTP 状态码（如可提取）、触发人/触发源 | 需要保留 bounded retry 细节，供历史查询、指标聚合、补发诊断使用 |
| `internal/delivery/service.go` | New focused domain service | 封装“创建 delivery 记录 -> 发送 -> 写 attempt -> 更新终态 -> replay/retry” | 这是唯一值得抽出来的小服务，避免 `WebhookHandler` 继续膨胀；不是全仓 service-layer 迁移 |
| `internal/handlers/delivery.go` | New handler | 提供 delivery 列表、详情、手工 retry/replay API | 与现有 `alert/config/user/webhook` handler 分层一致 |
| `internal/handlers/ops.go` | New handler | 提供 `/readyz`、运维健康汇总 API、必要的历史摘要 API | 把探针和控制台运维视图从业务 handler 中分离 |
| `internal/middleware/webhook_guard.go` | New middleware | webhook body size limit、基于 datasource 或 source 名称的限流、拒绝原因标准化日志 | 入口硬化应在路由层完成，不应散落在 `HandleWebhook` 内 |
| `frontend/src/pages/DeliveryHistory.tsx` | New page | 投递历史、失败过滤、详情抽屉、手工重试入口 | 新能力的核心 UI |
| `frontend/src/pages/OpsHealth.tsx` | New page | readiness/metrics 摘要、失败位置趋势、最近异常概览 | 对应“operations history/health surface” |
| `frontend/src/stores/deliveryStore.ts` | New store | 管理 delivery/history/replay 查询与操作状态 | 遵循现有 Zustand 分域模式 |

### Modified Components

| Component | Change | Integration point |
|-----------|--------|-------------------|
| `internal/handlers/webhook.go` | 发送通知前不再直接只依赖日志；改为调用 `delivery.Service` 创建/推进投递状态 | 保留现有 trace/log contract，新增 DB 持久化 side effect |
| `internal/notifier/notifier.go` | 允许返回更结构化的发送失败信息（至少可提取 HTTP status / retryable 判断） | attempt 持久化和 metrics 需要比纯字符串日志更稳定的字段 |
| `internal/router/router.go` | 新增 `delivery` API、`ops` API、`/readyz`、`/metrics`，并把 webhook guard middleware 只挂到 `/webhook/:source_name` | 保持现有路由组织，不影响已上线业务 API |
| `internal/database/postgres.go` | AutoMigrate 新 delivery 表 | 复用现有 GORM 迁移入口 |
| `internal/authz/capabilities.go` | 初期建议不新增角色模型；读历史复用 `view_config`，手工 replay/retry 复用 `manage_config` | 最小化权限面变化，避免把 v1.4 变成 RBAC 里程碑 |
| `frontend/src/api/client.ts` | 新增 delivery / ops API wrapper | 复用现有 axios 客户端与 401 处理 |
| `frontend/src/App.tsx` | 新增菜单与路由：`/deliveries`、`/ops-health` | 与当前控制台壳层集成 |
| `frontend/src/types/index.ts` | 新增 delivery、attempt、ops health DTO | 与现有 TS 类型真源保持一致 |

## Delivery Data Model

推荐两层模型，而不是只加一个“失败表”：

### `notification_deliveries`

一行代表一次“告警 x 渠道”的投递实体。

建议字段：

- `id`
- `alert_id`
- `trace_id`
- `source`
- `channel_id`
- `channel_name_snapshot`
- `channel_type_snapshot`
- `route_rule_id`、`route_rule_name_snapshot`
- `delivery_status`：`pending | sending | delivered | failed | replay_queued | replayed`
- `delivery_mode`：`rendered | default`
- `attempt_count`
- `last_attempt_at`
- `last_success_at`
- `last_error`
- `last_http_status`
- `trigger_kind`：`pipeline | manual_retry | manual_replay`
- `created_at` / `updated_at`

### `notification_delivery_attempts`

一行代表一次实际发送尝试。

建议字段：

- `id`
- `delivery_id`
- `attempt_number`
- `trigger_kind`
- `result`：`success | failure`
- `retryable`
- `error_message`
- `http_status`
- `duration_ms`
- `initiated_by_user_id` / `initiated_by_username`（人工动作时）
- `created_at`

这两个表的关系能同时支撑：

- 失败历史查询
- 每条告警多渠道投递诊断
- 手工重试后的完整闭环
- 指标聚合而不需要解析日志回填

## Data Flow Changes

### 1. Ingress Path

现状：

`POST /webhook/:source_name -> HandleWebhook -> datasource/api_key check -> parse -> persist Alert -> Redis -> async notify`

建议：

`POST /webhook/:source_name -> webhook guard middleware -> HandleWebhook -> datasource/api_key check -> parse -> persist Alert -> Redis -> async delivery service`

新增点：

- **body size limit** 在进入 handler 前拒绝超大 payload。
- **rate limiting** 在 datasource/source 维度执行，优先使用 Redis 计数，Redis 不可用时可退化为保守拒绝或固定窗口内存限流。
- guard 拒绝与 handler 拒绝都写入现有 canonical log contract，保持 `trace_id` 或生成 ingress-level request id。

### 2. Notification Path

现状：

`processAlertNotifications -> findMatchedChannels -> sendNotification -> sendChannelNotification -> retry loop -> log only`

建议：

`processAlertNotifications -> findMatchedChannels -> deliveryService.StartDelivery(alert, channel, routeRule) -> deliveryService.SendWithPersistence(...) -> notifier.SendToChannel`

发送时序：

1. 为每个 `alert + channel` 创建 `notification_delivery`
2. 每次 attempt 前后写 `notification_delivery_attempt`
3. 成功则更新 delivery 为 `delivered`
4. retry budget 用尽则更新 delivery 为 `failed`
5. 仍保留现有 `send_attempt` / `terminal_failure` / `send_notification` 日志，不以 DB 代替日志

### 3. Manual Retry / Replay Path

新增独立 API，不从 webhook 入口回灌：

`POST /api/v1/deliveries/:id/retry`

流程：

1. 鉴权通过后读取 delivery + alert 快照
2. 重新计算当前 channel 配置可用性
3. 复用 `delivery.Service`
4. 生成新的 attempt 记录
5. 更新 delivery 状态与 `trigger_kind=manual_retry`
6. 写现有 `AuditLog`

如果要求“按告警整体重放全部失败渠道”，再补：

`POST /api/v1/alerts/:id/replay-deliveries`

但应晚于单条 delivery retry，实现顺序更安全。

### 4. Health / Metrics Path

推荐拆三类：

- `/health`：保留现状，仅进程存活
- `/readyz`：新增深检查，至少覆盖 PostgreSQL、Redis、关键配置加载状态
- `/metrics`：新增 Prometheus-style 指标暴露

推荐 metrics 来源：

- ingress request total / rejected total / limited total
- alert persist success/failure
- delivery created total
- delivery attempts total（按 channel type、result、retryable）
- delivery terminal failures total
- manual retry total / success total / failure total
- readiness dependency status（如用 gauge）

不要让前端直接读取 `/metrics` 原文；控制台应走 `ops` API 聚合摘要。

## API Surface Recommendation

### Read APIs

- `GET /api/v1/deliveries`
- `GET /api/v1/deliveries/:id`
- `GET /api/v1/ops/health`
- `GET /api/v1/ops/history`

`/deliveries` 过滤维度建议至少包含：

- `alert_id`
- `trace_id`
- `channel_id`
- `status`
- `source`
- `trigger_kind`
- `created_from` / `created_to`

### Action APIs

- `POST /api/v1/deliveries/:id/retry`

可延后：

- `POST /api/v1/alerts/:id/replay-deliveries`
- `POST /api/v1/deliveries/:id/cancel`（当前架构无队列，不是刚需）

### Auth Recommendation

- 读历史/健康：`view_config`
- retry/replay：`manage_config`
- readiness/metrics：不走前端鉴权，但应定位为内网/探针入口；应用内不做复杂 UI token 例外逻辑

## Frontend Integration

推荐新增两个独立页面，而不是把运维面板塞进现有 Dashboard：

### `DeliveryHistory`

展示：

- delivery 列表
- 失败态筛选
- attempt 时间线
- 最近错误
- 手工 retry 按钮

依赖：

- `deliveryStore`
- `deliveryApi`
- 现有 `PermissionNotice`

### `OpsHealth`

展示：

- readiness 摘要
- 最近 delivery failure trend
- 渠道成功率/失败率
- ingress reject / rate-limit 计数
- 最近人工 retry 结果

依赖：

- `opsApi`
- 轻量轮询或手动刷新

## Recommended Build Order

1. **Schema + passive persistence seam**
   - 新增 delivery / attempt 模型、迁移、基础 repository/service
   - 先让现有发送链路“额外写记录”，不改变重试判定和发送顺序
   - 这是所有后续 API、UI、metrics 的真源

2. **Send-path integration without new operator actions**
   - `WebhookHandler` 改接 `delivery.Service`
   - 保持现有日志字段和 bounded retry 行为不变
   - 先验证“成功/失败/重试耗尽”三类状态都能落库

3. **Read-only APIs + health summary**
   - 上线 `GET /deliveries`、`GET /deliveries/:id`、`GET /ops/health`
   - 加 `/readyz`
   - 这一阶段就能先给维护者排障，不引入额外补发风险

4. **Metrics exposure**
   - 基于 delivery 表和 attempt 结果补指标
   - 只新增暴露，不改业务行为
   - 便于随后调 ingress 限流阈值

5. **Manual retry / replay**
   - 先做单条 delivery retry
   - 写 `AuditLog`
   - 再决定是否需要“按 alert 批量 replay”

6. **Ingress hardening**
   - 新增 webhook size limit
   - 新增 datasource 维度 rate limit
   - readiness 与 metrics 已先落地，便于观察误伤与阈值问题

7. **Frontend ops surfaces**
   - 先 `DeliveryHistory`
   - 后 `OpsHealth`
   - UI 永远消费后端聚合 API，不直接解析指标文本

## Dependency Order For Roadmap

```text
delivery schema
  -> send-path persistence
  -> read APIs
  -> metrics + ops health summary
  -> manual retry/replay API
  -> delivery history UI
  -> ops health UI

webhook guard middleware
  -> readiness/metrics visibility
  -> production rollout tuning
```

## Non-Goals / Avoid

- 不引入消息队列或独立 worker
- 不把全仓改造成 service/repository 大迁移
- 不把 `/health` 改成深探针
- 不把人工 replay 设计成复杂 job orchestration；当前同步触发 + 持久化 attempt 足够
- 不让前端直接依赖日志文本或 `/metrics` 原始格式

## Risks To Watch

| Risk | Why it matters | Mitigation |
|------|----------------|------------|
| 发送路径新增 DB 写入，拉长通知延迟 | v1.4 最敏感变更点 | delivery/attempt 写入保持单条轻量、必要索引最小化、先覆盖回归测试 |
| replay 误触发重复通知 | 人工补发天然有外部副作用 | 只对 failed delivery 开放默认 retry；显式确认；写 audit log |
| rate limit 配置不当误伤正常 webhook | 直接影响主链路可用性 | 先上 metrics/readiness，再逐步启用限流 |
| delivery 记录与现有日志脱节 | 排障真源分裂 | `trace_id`、`alert_id`、`channel_id` 必须贯穿日志和表 |

## Confidence Notes

- **HIGH**：现有代码已明确显示通知逻辑集中在 `internal/handlers/webhook.go`，这是 delivery persistence 的最佳挂载点。
- **HIGH**：现有 Gin 路由和 Zustand/API 模式足够承载新增 read/action surface，无需迁移框架。
- **MEDIUM**：若后续需要大规模批量 replay，当前同步应用内执行模型可能不足；但这不应成为 v1.4 首轮设计前提。
