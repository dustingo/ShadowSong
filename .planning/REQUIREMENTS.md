# Requirements: v1.4 Delivery Recovery and Production Hardening

**Defined:** 2026-04-29
**Milestone:** v1.4 Delivery Recovery and Production Hardening
**Core Value:** 运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## v1 Requirements

### Delivery Recovery

- [ ] **DELV-01**: 维护者可以查看每次通知投递的持久化记录，包括告警、渠道、发送模式、尝试次数、最终结果和失败原因
- [ ] **DELV-02**: 通知在超过当前即时重试上限后，会把最终失败结果持久化保存，而不只停留在日志中
- [ ] **DELV-03**: 维护者可以针对单条失败通知执行受控的 `retry`
- [ ] **DELV-04**: 维护者可以针对单条失败通知执行受控的 `replay`
- [ ] **DELV-05**: 人工 `retry/replay` 会记录操作者、触发原因、触发时间、执行结果和关联原始投递记录，保持审计链路完整
- [ ] **DELV-06**: 首版投递记录会保存足够支撑审计和单条 replay 的不可变快照，而不是只依赖当前实时配置

### Ingress Hardening

- [ ] **INGR-01**: Webhook 入口会限制单次请求体大小，超限请求会被明确拒绝并留下可检索的失败记录
- [ ] **INGR-02**: Webhook 入口会对单数据源或单来源流量施加基础限流，避免瞬时洪峰直接压垮主服务
- [ ] **INGR-03**: Webhook 的签名或原始请求校验会在 JSON 绑定前处理原始 body，避免校验语义失真
- [ ] **INGR-04**: 生产环境会显式配置允许来源、JWT 密钥、数据库连接和 Redis 连接，不再依赖危险开发默认值
- [ ] **INGR-05**: 系统提供 readiness 检查，能够反映 PostgreSQL、Redis 和关键运行依赖是否可用，而不是只返回静态 `ok`

### Operations Visibility

- [ ] **OPER-01**: 维护者可以按时间、告警、渠道和结果查询通知投递历史，而不需要直接翻后端日志
- [ ] **OPER-02**: 维护者可以查看通知渠道最近的成功率、失败次数和最近失败原因，用于快速判断渠道健康状态
- [ ] **OPER-03**: 系统会暴露关键运行指标，包括 webhook 接入量、通知发送成功率、重试次数和最终失败次数
- [ ] **OPER-04**: 维护者可以从告警详情或运维页面直接跳转到关联的通知投递历史与失败证据
- [ ] **OPER-05**: 首版运维健康页提供基于账本的聚合摘要视图，而不是依赖复杂实时大盘

### Reliability Debt Closure

- [ ] **DEBT-01**: webhook dedup 更新路径会正确处理数据库保存失败，避免告警状态变化静默丢失
- [ ] **DEBT-02**: 通知 fallback 默认内容路径在重试耗尽时有专门回归验证，确保 `terminal_failure` 行为不漂移
- [ ] **DEBT-03**: `channel_lookup` 等次级失败路径采用字段级回归断言，而不是只依赖子串匹配

## v2 Requirements

### Delivery Platform

- **DELV-07**: 系统支持批量 `retry/replay` 失败通知，而不只限于单条操作
- **DELV-08**: 系统支持更复杂的通知退避策略、熔断或死信处理

### Observability Platform

- **OPER-06**: 系统接入更完整的集中式 metrics/trace 平台，而不只暴露本地 `/metrics`
- **OPER-07**: 系统提供更复杂的实时运维大盘和趋势分析视图

### Security Platform

- **INGR-06**: 系统支持面向多实例部署的一致性限流方案
- **SECU-01**: 系统支持更完整的企业身份集成，例如 OIDC/SSO

## Out of Scope

| Feature | Reason |
|---------|--------|
| 引入消息队列、Temporal、Asynq 或外部 workflow 平台 | 当前目标是在现有 Go + PostgreSQL + Redis 基线上补齐生产能力，而不是扩大成基础设施重构 |
| 全仓统一 JSON logging 改造 | v1.3 的 canonical webhook logging 已可用，本轮优先补恢复能力和生产防护 |
| 细粒度 RBAC / 多租户权限模型 | 与当前生产化短板相比优先级更低，会显著扩大 scope |
| 运行时历史命名全面迁移 | `go.mod` module path 和 JWT issuer 仍属单独契约迁移议题，不应混入本轮 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| DELV-01 | Phase 18 | Pending |
| DELV-02 | Phase 18 | Pending |
| DELV-03 | Phase 19 | Pending |
| DELV-04 | Phase 19 | Pending |
| DELV-05 | Phase 19 | Pending |
| DELV-06 | Phase 18 | Pending |
| INGR-01 | Phase 20 | Pending |
| INGR-02 | Phase 20 | Pending |
| INGR-03 | Phase 20 | Pending |
| INGR-04 | Phase 20 | Pending |
| INGR-05 | Phase 20 | Pending |
| OPER-01 | Phase 19 | Pending |
| OPER-02 | Phase 21 | Pending |
| OPER-03 | Phase 21 | Pending |
| OPER-04 | Phase 19 | Pending |
| OPER-05 | Phase 21 | Pending |
| DEBT-01 | Phase 20 | Pending |
| DEBT-02 | Phase 21 | Pending |
| DEBT-03 | Phase 21 | Pending |

**Coverage:**
- v1 requirements: 19 total
- Mapped to phases: 19
- Unmapped: 0

---
*Requirements defined: 2026-04-29*
*Last updated: 2026-04-29 after initial v1.4 definition*
