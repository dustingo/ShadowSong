# Project Research Summary

**Project:** 游戏运维告警系统
**Domain:** v1.4 Delivery Recovery and Production Hardening
**Researched:** 2026-04-29
**Confidence:** HIGH

## Executive Summary

这是一个已上线基础告警链路上的恢复与加固里程碑，不是架构重做。研究结论高度一致：继续沿用现有 `Go + Gin + GORM + PostgreSQL + Redis + React + Vite`，把“通知投递”从日志与内存副作用提升为 PostgreSQL 中可持久化、可查询、可审计的一等对象，再基于这个投递账本补手工重试、只读历史、readiness、指标和运维视图。

推荐做法是先锁定数据契约，再做发送链路双写接入，然后才开放运维读接口和人工恢复动作。入口侧只做最小但必要的硬化：原始请求体验签、body size limit、按来源限流、独立 `/readyz` 与 `/metrics`。不引入新队列、新 worker 平台、新前端数据层，也不把 Redis 变成投递历史真源。

主要风险不在技术选型，而在语义和顺序：如果先做 UI 或 replay 按钮、后补投递模型，历史会失真；如果 replay 复用错误入口，会绕过静默、去重或路由；如果指标与状态枚举不稳定，运维面会误导决策。路线图应强制以“投递账本优先、读多于写、单条恢复先于批量恢复、硬化与可观测分阶段落地”为主线。

## Key Findings

### Recommended Stack

本里程碑不做技术迁移。唯一明确值得新增的通用依赖是 `github.com/prometheus/client_golang` 用于 `/metrics` 暴露，以及 `golang.org/x/time/rate` 用于低增量限流。其余能力优先复用现有 PostgreSQL、GORM、Gin、Redis、React、Ant Design、ECharts。

**Core technologies:**
- `PostgreSQL 14`：投递真源与恢复账本，承载 `notification_deliveries` 和 `notification_delivery_attempts`
- `GORM 1.31.x`：迁移、事务、`FOR UPDATE SKIP LOCKED` 认领重试任务
- `Gin 1.12.x`：挂载 webhook guard、`/readyz`、`/metrics` 与新增 delivery/ops API
- `Redis 7`：继续做现有短时流与实时通道依赖，不承担投递历史真源
- `Prometheus client_golang`：最小化服务指标接入，支持入口、投递、readiness 指标
- `golang.org/x/time/rate`：应用内来源级限流，适合单实例或低复杂度场景
- `React + AntD + ECharts + Zustand`：复用现有控制台栈承载投递历史和健康概览，不新增 UI 基建

### Expected Features

**Must have (table stakes):**
- 持久化投递记录与尝试历史，覆盖成功与失败，而不是只存失败
- 明确的投递状态模型与失败分类，支持 retryable / terminal failure 判断
- 权限受控的单条手工 retry/replay，并记录操作者、原因与审计链路
- 可检索的投递历史，至少支持 `alert_id`、`trace_id`、`channel`、`status`、时间范围筛选
- 入口真实性校验、body size limit、按来源限流
- 独立 `/readyz` 与最小 Prometheus 指标
- 渠道健康 / 失败汇总视图

**Should have (differentiators):**
- 基于历史筛选结果的 replay
- 使用已存 payload / 模板快照做 replay preview
- 单告警统一投递时间线
- 入口拒绝原因分析与聚合

**Defer (v2+ / out of scope):**
- 新消息队列、作业平台或独立 worker 系统
- 无限自动重放、复杂 retry DSL、工作流引擎
- 批量 replay 先于单条安全模型
- 外部观测平台迁移或 incident command center
- 把 Redis 作为 delivery history 真源

### Architecture Approach

架构上只新增两个子域边界：`Delivery Recovery` 和 `Operations Health`。后端增加 delivery/attempt 模型、一个聚焦的 `internal/delivery/service.go`、delivery/ops handlers，以及 webhook guard middleware；前端增加 `DeliveryHistory`、`OpsHealth` 和对应 store/API。核心接入点仍在现有 `internal/handlers/webhook.go` 发送链路中，但先以“旁路持久化 + 不改变现有 bounded retry 行为”为第一阶段目标。

**Major components:**
1. `notification_deliveries` / `notification_delivery_attempts` — 投递真源、状态机、审计与历史查询基础
2. `internal/delivery/service.go` — 封装创建投递、写 attempt、发送、更新终态、手工 retry/replay
3. `internal/handlers/delivery.go` / `internal/handlers/ops.go` — 提供历史、详情、重试、健康汇总、readiness、metrics
4. `internal/middleware/webhook_guard.go` — 原始 body 校验、尺寸限制、来源级限流、拒绝原因标准化
5. `frontend/src/pages/DeliveryHistory.tsx` / `OpsHealth.tsx` — 运维控制台的历史与健康面

### Critical Pitfalls

1. **投递记录可变导致 replay 不可解释** — 首版就要冻结 delivery intent 和 attempt 元数据，不能只存外键与错误字符串
2. **replay 走错链路绕过去重/静默/路由** — v1.4 先明确为“重放原始投递意图”，不要隐式重新执行全量策略
3. **恢复 API 成为高危后门** — retry/replay 必须有独立权限、必填原因、审计记录和单条优先策略
4. **只持久化失败记录** — 会破坏成功率、历史查询和恢复判断；必须持久化全部 attempts
5. **热路径迁移或索引过重** — 先做 additive schema + dual-write，首版只加最小索引，避免锁住 live path

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Delivery Ledger Contract
**Rationale:** 这是所有恢复、指标、UI、审计能力的真源，必须先稳定数据契约。
**Delivers:** `notification_deliveries`、`notification_delivery_attempts`、状态枚举、失败分类、最小索引、双写准备。
**Addresses:** 持久化投递记录、状态模型、审计基础。
**Avoids:** 可变记录、只存失败、热路径大迁移。

### Phase 2: Send-Path Persistence Integration
**Rationale:** 先让现有发送链路额外落库，再判断真实吞吐、状态覆盖和日志对齐情况。
**Delivers:** `delivery.Service`、发送前后 attempt 持久化、成功/失败/重试耗尽终态更新、`trace_id` 贯通。
**Uses:** PostgreSQL、GORM 事务、现有 notifier/webhook handler。
**Implements:** Delivery Recovery 子域接入现有链路。

### Phase 3: Read APIs and Safe Operator Recovery
**Rationale:** 先开放只读面，再开放受控写动作；恢复必须建立在稳定账本之上。
**Delivers:** `GET /api/v1/deliveries`、`GET /api/v1/deliveries/:id`、单条 `POST /api/v1/deliveries/:id/retry`、审计落库。
**Addresses:** 可检索历史、单条人工恢复、运维排障。
**Avoids:** replay 成为未审计侧门、语义漂移、批量操作先行。

### Phase 4: Ingress Hardening and Service Readiness
**Rationale:** 入口保护要做，但要在账本和观测基础已有后再调阈值，避免误伤主链路。
**Delivers:** 原始 body 验签、body size limit、来源级限流、`/readyz`、拒绝原因标准化。
**Uses:** Gin middleware、`x/time/rate`、现有 datasource/source 配置。
**Implements:** Operations Health 的入口侧加固。

### Phase 5: Metrics and Ops Surfaces
**Rationale:** 指标和 UI 要基于稳定状态语义，不能反过来倒逼模型。
**Delivers:** `/metrics`、健康摘要 API、`DeliveryHistory` 页面、`OpsHealth` 页面、失败趋势与渠道健康汇总。
**Addresses:** delivery/ingress 可观测性、渠道健康、运维历史面。
**Avoids:** 高基数指标、历史页语义混乱、调试面无限膨胀。

### Phase Ordering Rationale

- 顺序严格依赖于“投递账本 -> 发送链路接入 -> 只读与单条恢复 -> 入口硬化 -> 运维面”；这是四份研究共同结论。
- 架构上先做后端真源与状态语义，再做前端页面，避免 UI 消费不稳定字段。
- 风险最高的是发送热路径和 replay 语义，因此前置双写验证与单条恢复，延后批量 replay 与高级 UX。

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3:** 若要把 `retry` 和 `replay` 分成不同语义，需在规划时补充更细的权限与行为定义
- **Phase 4:** 若部署是多实例，限流是否需要 Redis 共享计数要在规划时二次确认
- **Phase 5:** 若需要批量 replay、payload preview 或复杂健康图表，需单独做范围收敛

Phases with standard patterns (skip research-phase):
- **Phase 1:** PostgreSQL + GORM additive schema、attempt ledger、最小索引模式已足够明确
- **Phase 2:** 在现有 webhook/notifier 链路旁路持久化属于标准 brownfield 增量接入
- **Phase 4 `/readyz` and `/metrics` 部分:** Gin + Prometheus + 依赖探针是成熟固定模式

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | 结论集中，新增依赖极少，主要依据官方 Gin/GORM/PostgreSQL/Prometheus 文档 |
| Features | MEDIUM-HIGH | table stakes 与 defer 边界清晰，但批量 replay、preview 等增强项仍需产品取舍 |
| Architecture | HIGH | 接入点、边界与构建顺序在现有仓库结构上都很自然，研究结论一致 |
| Pitfalls | MEDIUM-HIGH | 大部分来自官方文档加 brownfield 推断，语义与 rollout 风险判断可信但仍需实现期验证 |

**Overall confidence:** HIGH

### Gaps to Address

- `retry` 与 `replay` 是否在 v1.4 同时暴露：建议先只承诺单条 `retry`，把 `replay` 作为显式后续选项
- 多实例部署现状是否成立：这会影响限流实现是否停留在进程内
- 是否需要冻结模板渲染快照或仅冻结 delivery intent：这决定首版 schema 宽度
- 首版健康页的数据来源是实时聚合还是读账本摘要：需要在性能和复杂度之间做一次规划决策

## Sources

### Primary (HIGH confidence)
- PostgreSQL docs — `SELECT ... FOR UPDATE ... SKIP LOCKED`
- GORM docs — locking、`db.DB()`、连接检查
- Gin docs/security guide — trusted proxies、rate limiting、request body limiting
- Prometheus docs — Go client instrumentation、metrics best practices、histograms
- Kubernetes docs — readiness/liveness probe separation
- Stripe / GitHub webhook docs — raw body signature verification、manual redelivery、duplicate delivery handling

### Secondary (MEDIUM confidence)
- Grafana Alerting docs — delivery status、state history、meta monitoring
- Prometheus Alertmanager docs — routing、dedup、silencing、monitoring baseline
- PagerDuty docs — orchestration limits、auditability expectations

---
*Research completed: 2026-04-29*
*Ready for roadmap: yes*
