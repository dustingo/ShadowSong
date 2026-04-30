# Roadmap: 游戏运维告警系统

## Milestones

- ✅ **v1.0 AI Removal Complete** — Phases 1-4 (shipped 2026-04-10). Archive: `.planning/milestones/v1.0-ROADMAP.md`
- ✅ **v1.1 Enterprise Access Control** — Phases 5-9 (shipped 2026-04-15). Archive: `.planning/milestones/v1.1-ROADMAP.md`
- ✅ **v1.2 Alert Pipeline Hardening** — Phases 10-13 (shipped 2026-04-21). Archive: `.planning/milestones/v1.2-ROADMAP.md`
- ✅ **v1.3 Notification Reliability and Observability** — Phases 14-17 (shipped 2026-04-29). Archive: `.planning/milestones/v1.3-ROADMAP.md`
- 🚧 **v1.4 Delivery Recovery and Production Hardening** — Phases 18-21 (planned)

## Overview

v1.4 围绕“失败可恢复、入口更安全、运维更可观测”的生产化基线展开，在不做技术栈迁移、不引入 MQ 或 workflow engine 的前提下，先把通知投递账本做成真源，再开放单条恢复动作、补入口硬化与 readiness，最后提供基于账本的指标与运维健康视图。

## Phases

**Phase Numbering:**
- Integer phases continue across milestones and never restart
- Decimal phases are reserved for urgent insertions if later needed

<details>
<summary>✅ v1.0 AI Removal Complete (Phases 1-4) - SHIPPED 2026-04-10</summary>

See `.planning/milestones/v1.0-ROADMAP.md`

</details>

<details>
<summary>✅ v1.1 Enterprise Access Control (Phases 5-9) - SHIPPED 2026-04-15</summary>

See `.planning/milestones/v1.1-ROADMAP.md`

</details>

<details>
<summary>✅ v1.2 Alert Pipeline Hardening (Phases 10-13) - SHIPPED 2026-04-21</summary>

See `.planning/milestones/v1.2-ROADMAP.md`

</details>

<details>
<summary>✅ v1.3 Notification Reliability and Observability (Phases 14-17) - SHIPPED 2026-04-29</summary>

See `.planning/milestones/v1.3-ROADMAP.md`

</details>

## v1.4 Delivery Recovery and Production Hardening (Phases 18-21)

**Milestone Goal:** 在保持现有告警主链路连续可用的前提下，为通知投递补齐可恢复账本、受控人工恢复、入口生产硬化，以及首版运维可观测面。

- [x] **Phase 18: Establish Delivery Ledger** - Persist notification deliveries, final failures, and immutable replay snapshots as the recovery source of truth. (completed 2026-04-30)
- [ ] **Phase 19: Enable Safe Recovery Operations** - Expose searchable delivery history and audited single-item retry/replay flows for operators.
- [ ] **Phase 20: Harden Ingress And Runtime Readiness** - Add webhook guardrails, production config enforcement, readiness checks, and dedup save-failure protection.
- [ ] **Phase 21: Ship Ops Visibility Surfaces** - Publish metrics, channel-health summaries, and the first ledger-backed operations health views.

## Phase Details

### Phase 18: Establish Delivery Ledger
**Goal**: 通知发送链路会把每次投递和最终结果持久化为稳定账本，并保存足够支撑审计与单条 replay 的不可变快照。  
**Depends on**: Phase 17  
**Requirements**: DELV-01, DELV-02, DELV-06  
**Success Criteria** (what must be TRUE):
  1. 维护者可以查看任一通知投递的账本记录，包含告警、渠道、发送模式、尝试次数、最终结果和失败原因。
  2. 超过当前即时重试上限的通知会留下持久化终态记录，而不是只存在于日志里。
  3. 单条投递记录保存的快照足以支撑后续审计和单条 replay，不依赖当时之外的实时配置状态。
**Plans**: 3 plans

Plans:
- [x] 18-01-PLAN.md - 锁定 delivery 双表 schema、不可变快照合同与聚焦 service 真源
- [x] 18-02-PLAN.md - 把账本接入 webhook 通知热路径并保留现有 retry/logging 契约
- [x] 18-03-PLAN.md - 交付最小 delivery 列表/详情只读 API 与现有鉴权接入

### Phase 19: Enable Safe Recovery Operations
**Goal**: 维护者可以基于稳定账本查询通知历史，并对单条失败通知执行受控、可审计的 retry 或 replay。  
**Depends on**: Phase 18  
**Requirements**: DELV-03, DELV-04, DELV-05, OPER-01, OPER-04  
**Success Criteria** (what must be TRUE):
  1. 维护者可以按时间、告警、渠道和结果筛选通知投递历史，而不需要直接翻后端日志。
  2. 维护者可以从告警详情或运维页面直接跳转到关联的通知投递历史和失败证据。
  3. 维护者可以仅针对单条失败通知触发 retry，并看到这次恢复动作的执行结果。
  4. 维护者可以仅针对单条失败通知触发 replay，并且每次人工恢复都会记录操作者、原因、时间、结果和关联原始投递记录。
**Plans**: 3 plans
**UI hint**: yes

Plans:
- [ ] 19-01-PLAN.md - 交付可筛选的 delivery 历史页、失败证据展示与 alert deeplink
- [ ] 19-02-PLAN.md - 交付后端单条 retry/replay、结构化恢复审计与读写权限分离
- [ ] 19-03-PLAN.md - 把 recovery 动作接入历史页并完成 operator 可操作、viewer 只读闭环

### Phase 20: Harden Ingress And Runtime Readiness
**Goal**: webhook 和服务运行入口具备明确的请求防护、生产配置收口和依赖就绪判断，入口异常与 dedup 保存失败不会再静默漂移。  
**Depends on**: Phase 19  
**Requirements**: INGR-01, INGR-02, INGR-03, INGR-04, INGR-05, DEBT-01  
**Success Criteria** (what must be TRUE):
  1. 超出 webhook 请求体大小限制的请求会被明确拒绝，并留下可检索的失败记录。
  2. 单数据源或单来源的瞬时洪峰会受到基础限流保护，不会直接把主服务压穿。
  3. webhook 的签名或原始请求校验基于原始 body 执行，避免 JSON 绑定后再校验带来的语义失真。
  4. 生产环境缺失允许来源、JWT 密钥、数据库或 Redis 连接配置时，系统不会继续依赖危险默认值启动。
  5. `/readyz` 能真实反映 PostgreSQL、Redis 和关键运行依赖是否可用，dedup 更新路径的数据库保存失败也会被正确处理而不是静默丢失。
**Plans**: TBD

### Phase 21: Ship Ops Visibility Surfaces
**Goal**: 维护者可以通过指标和账本聚合视图快速判断通知失败位置、渠道健康状态和关键失败证据，同时锁定次级失败路径回归契约。  
**Depends on**: Phase 20  
**Requirements**: OPER-02, OPER-03, OPER-05, DEBT-02, DEBT-03  
**Success Criteria** (what must be TRUE):
  1. 系统暴露关键运行指标，至少覆盖 webhook 接入量、通知发送成功率、重试次数和最终失败次数。
  2. 维护者可以查看各通知渠道最近的成功率、失败次数和最近失败原因，用于快速判断渠道健康状态。
  3. 首版运维健康页提供基于投递账本的聚合摘要视图，而不是依赖复杂实时大盘。
  4. `terminal_failure` 的 fallback 默认内容路径和 `channel_lookup` 等次级失败路径都有稳定回归验证，不会因日志文本漂移而失守。
**Plans**: TBD
**UI hint**: yes

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 18. Establish Delivery Ledger | v1.4 | 3/3 | Complete    | 2026-04-30 |
| 19. Enable Safe Recovery Operations | v1.4 | 0/TBD | Not started | - |
| 20. Harden Ingress And Runtime Readiness | v1.4 | 0/TBD | Not started | - |
| 21. Ship Ops Visibility Surfaces | v1.4 | 0/TBD | Not started | - |
