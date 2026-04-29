# Milestone Stack Research: v1.4 Delivery Recovery and Production Hardening

**Project:** 游戏运维告警系统  
**Scope:** 只研究新增能力所需的栈变化，不重做现有主链路  
**Researched:** 2026-04-29  
**Overall confidence:** HIGH

## Recommendation

继续沿用现有 `Go + Gin + GORM + PostgreSQL + Redis + React + Vite`。这一里程碑不需要引入新队列、新作业平台、新观测平台或新的前端状态框架。核心做法是把“通知投递”从内存态提升为 PostgreSQL 中的持久化记录，并在应用内增加受控的 replay/retry、入口限流、readiness、Prometheus 指标与运维历史界面。

## Required Additions

### Backend Dependencies

| Addition | Version | Why this is needed | Integration point |
|----------|---------|--------------------|-------------------|
| `github.com/prometheus/client_golang` | `v1.22.x` | 官方 Prometheus Go client，直接暴露 `/metrics`，适合最小增量落地服务指标 | `cmd/server/main.go`, `internal/router/router.go`, webhook/notification handlers |
| `golang.org/x/time` | `latest compatible` (`rate` package; pkg.go.dev shows `v0.15.0`, not latest) | Gin 官方安全指南直接给出 `golang.org/x/time/rate` 作为简单限流中间件方案，适合 webhook/登录/高风险入口 | `internal/middleware/` 新增 rate-limit middleware，挂到 `/webhook/*` 与敏感管理接口 |

### Existing Stack to Reuse Harder

| Existing technology | Keep using it for | Concrete change |
|---------------------|-------------------|-----------------|
| PostgreSQL 14 | 投递真源、操作历史、重试调度、健康聚合 | 新增 `notification_deliveries` + `notification_delivery_attempts`（或等价主表+append-only attempts 表） |
| GORM 1.31.x | 迁移、事务、锁、批量状态迁移 | 用事务 + `clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}` 认领 retry/replay 任务 |
| Redis 7 / `go-redis/v9` | 保留为现有短时流/实时通道依赖，不做持久化真源 | readiness 里做 `Ping`；如未来多实例限流成为刚需，再复用 Redis 做 shared counters |
| Gin 1.12.x | 中间件编排、健康检查、指标挂载、入口硬化 | 新增 body-size limit、trusted proxies、rate limit、readiness、metrics routes |
| React + AntD + ECharts | 运维历史、健康总览、失败筛选、重放操作 UI | 复用现有表格/图表能力，不加新 UI 框架 |

## Data Model Changes

推荐把投递恢复拆成两个持久化层次，而不是继续只靠日志和内存重试：

| Table / record | Purpose | Key fields |
|----------------|---------|------------|
| `notification_deliveries` | 一条“应投递事件”的主记录，承载当前状态与下一步动作 | `id`, `alert_id`, `trace_id`, `channel_id`, `route_rule_id`, `status`, `retry_count`, `next_retry_at`, `last_error`, `last_http_status`, `last_attempt_at`, `resolved_at`, `created_at` |
| `notification_delivery_attempts` | append-only 尝试历史，支撑审计、运维查询、手工 replay 证据链 | `id`, `delivery_id`, `attempt_no`, `trigger_type` (`auto_retry`/`manual_replay`), `operator_user_id`, `request_snapshot`, `response_snapshot`, `error_text`, `status_code`, `duration_ms`, `created_at` |

### Storage choices

- 请求/响应快照、模板渲染上下文、provider 回执保存在 PostgreSQL `JSONB`，继续复用 GORM datatypes。
- `status`, `trigger_type`, `channel_type` 保持字符串枚举即可，不需要现在引入 PostgreSQL enum migration complexity。
- 必加索引：
  - `notification_deliveries(status, next_retry_at)`
  - `notification_deliveries(alert_id)`
  - `notification_deliveries(trace_id)`
  - `notification_deliveries(channel_id, created_at desc)`
  - `notification_delivery_attempts(delivery_id, created_at desc)`
- 若要限制重复 replay，可加业务幂等键，如 `delivery_key` 或 `(alert_id, channel_id, route_rule_id, attempt_group)` 唯一约束。

## Runtime Patterns

### 1. Persistent Retry / Replay

- 自动 retry 不再只靠 in-process bounded retry；内存重试保留为“快速瞬时补救”，但失败后必须落到 PostgreSQL。
- 后台恢复 worker 仍跑在现有 Go 进程内，不引入外部 job runner。
- 认领待处理记录时，用 PostgreSQL `FOR UPDATE ... SKIP LOCKED`；GORM 已支持 `Locking` 子句和 `SKIP LOCKED`。
- 手工 replay 通过受权 API 改写目标 delivery 状态并写一条 operator audit/attempt 记录，不直接伪造新告警。

### 2. Ingress Protection

- webhook 入口增加硬性 body size limit，直接用 `http.MaxBytesReader`，超限返回 `413`。
- 在 Gin 上显式配置 `SetTrustedProxies(...)`，否则 `ClientIP()` 可被伪造，限流和审计来源都不可靠。
- 限流先做应用层 middleware：
  - webhook: 按 `source + client IP` 或 `source + resolved client identity`
  - 登录/敏感管理操作: 单独更严格桶
- 如果当前部署仍是单实例，优先用 `x/time/rate` 本地 limiter；不要为了这一轮先引入网关/WAF。

### 3. Health and Readiness

- 保留现有轻量 `/health`。
- 新增 `/readyz`，显式检查：
  - PostgreSQL: `db.DB()` 后 `PingContext`
  - Redis: `client.Ping(ctx)`
  - 可选：notification worker goroutine 是否已启动、是否进入 fatal backoff
- `readiness` 失败应表示“实例暂不接流量”，不是“进程必须退出”。

### 4. Metrics

用 Prometheus exposition，不做新的 observability platform 扩张。

建议最小指标集：

| Metric | Type | Meaning |
|--------|------|---------|
| `webhook_requests_total` | counter | webhook 请求总量，按 source/status 分类 |
| `webhook_request_bytes` | histogram | webhook 体积分布，观察入口压力 |
| `notification_delivery_total` | counter | 创建的 delivery 数 |
| `notification_delivery_attempt_total` | counter | 尝试次数，按 channel/result/trigger_type 分类 |
| `notification_delivery_inflight` | gauge | 当前处理中 delivery |
| `notification_delivery_backlog` | gauge | `pending/retryable` backlog |
| `notification_delivery_latency_seconds` | histogram | 从告警持久化到投递完成/失败的耗时 |
| `notification_replay_total` | counter | 人工 replay 次数，按结果分类 |
| `readiness_failures_total` | counter | readiness 失败原因统计 |

## Frontend Additions

不建议新增前端基础设施。继续用现有依赖即可：

| Reuse | Use for |
|------|---------|
| Ant Design `Table`, `Drawer`, `Descriptions`, `Tag`, `Modal`, `Form` | delivery history、失败详情、手工 replay/retry |
| ECharts | channel success/failure trend、backlog、retry heatmap |
| 现有 Zustand stores | 新增 `deliveryStore` 或在告警相关 store 下扩展运维查询动作 |
| 现有 Vitest / Testing Library | 新页面与运维动作回归验证 |

推荐新增的产品面：

- `通知投递历史` 页面：按 alert、channel、trace_id、status、时间窗检索
- `通道健康` 页面：近 1h/24h 成功率、失败原因 TopN、backlog
- `delivery detail` 抽屉：attempt timeline、原始回执、operator replay 审计

## What NOT to Add

这些都不值得在本里程碑引入：

- 不加消息队列：`Kafka`、`RabbitMQ`、`NATS`、`SQS`
- 不加后台作业框架：`Asynq`、`Machinery`、`Temporal`
- 不加集中式观测平台迁移：Prometheus server/operator、Loki、ELK、Grafana 不是本仓库内交付前置
- 不加新的后端框架或 service mesh
- 不加新的前端数据层：`React Query`、`Redux Toolkit`
- 不把 Redis 变成 delivery history 真源；真源必须是 PostgreSQL
- 不为了 rate limit 先引入外部 API gateway/WAF，除非后续明确出现多实例共享限流需求

## Concrete Integration Plan

1. **PostgreSQL schema first**
   - delivery 主表 + attempts 历史表 + 必要索引
   - 保持与现有 `alerts`, `channels`, `route_rules`, `users` 的外键关系

2. **Notification sender refactor**
   - 发送前先建/更新 delivery
   - 每次尝试都写 attempt 记录
   - 内存 bounded retry 失败后把状态转成 `retryable`

3. **Recovery worker**
   - 应用内 goroutine 定时扫描 `retryable` deliveries
   - 事务内 `FOR UPDATE SKIP LOCKED` 认领批次

4. **Ingress hardening**
   - webhook body limit
   - trusted proxy config
   - rate-limit middleware

5. **Ops endpoints and UI**
   - `/readyz`, `/metrics`
   - delivery history / health summary APIs
   - React 运维页面

## Final Stack Call

这轮唯一明确值得新增的通用依赖是：

- `github.com/prometheus/client_golang`
- `golang.org/x/time`

其余能力优先通过现有 PostgreSQL、GORM、Gin、Redis、React、AntD、ECharts 落地。最重要的架构决定不是“再加什么基础设施”，而是把通知投递做成 PostgreSQL 中可查询、可重放、可度量的一等对象。

## Sources

- PostgreSQL `SELECT ... FOR UPDATE ... SKIP LOCKED`: https://www.postgresql.org/docs/18/sql-select.html (HIGH)
- GORM locking / `SKIP LOCKED`: https://gorm.io/docs/advanced_query.html (HIGH)
- GORM `db.DB()` and connection pooling for readiness checks: https://gorm.io/docs/connecting_to_the_database.html (HIGH)
- Gin security guide on rate limiting and trusted proxies: https://gin-gonic.com/en/docs/middleware/security-guide/ (HIGH)
- Gin body size limiting with `http.MaxBytesReader`: https://gin-gonic.com/zh-cn/docs/routing/upload-file/limit-bytes/ (HIGH)
- Prometheus Go instrumentation and `promhttp.Handler`: https://prometheus.io/docs/tutorials/instrumenting_http_server_in_go/ (HIGH)
- Prometheus Go client current stable release `v1.22.0` (2025-04-07): https://github.com/prometheus/client_golang/releases/tag/v1.22.0 (HIGH)
- `golang.org/x/time/rate` package docs: https://pkg.go.dev/golang.org/x/time/rate (MEDIUM; package docs show `v0.15.0` page, but also indicate a newer latest exists)
