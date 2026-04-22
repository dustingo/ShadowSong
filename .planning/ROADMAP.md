# Roadmap: 游戏运维告警系统

## Milestones

- ✅ **v1.0 AI Removal Complete** — Phases 1-4 (shipped 2026-04-10). Archive: `.planning/milestones/v1.0-ROADMAP.md`
- ✅ **v1.1 Enterprise Access Control** — Phases 5-9 (shipped 2026-04-15). Archive: `.planning/milestones/v1.1-ROADMAP.md`
- ✅ **v1.2 Alert Pipeline Hardening** — Phases 10-13 (shipped 2026-04-21). Archive: `.planning/milestones/v1.2-ROADMAP.md`
- 🚧 **v1.3 Notification Reliability and Observability** — Phases 14-17 (in progress)

## Overview

当前已启动 v1.3 `Notification Reliability and Observability`。本轮将在保持现有告警主流程和技术栈稳定的前提下，继续提升通知发送可靠性、告警链路可观测性、日志一致性以及历史命名/文档真相。

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

## v1.3 Notification Reliability and Observability (Phases 14-17)

### Phase 14: Establish Alert Trace Context

**Goal**: 为 webhook 接入到通知分发的主链路建立统一关联标识和关键生命周期观测点，先打通后续可靠性增强所依赖的排障基线。  
**Depends on**: Phase 13  
**Plans**: 4 plans

Plans:

- [x] 14-01: Generate and propagate stable correlation fields through webhook ingest and alert processing
- [x] 14-02: Verify alert lifecycle observability across ingest, persistence, routing, and notification entrypoints

**Success criteria:**
- 单次告警处理具备可跨阶段检索的统一关联字段
- 接入、持久化、路由和通知入口至少各有一个稳定观测点
- 验证记录能够演示如何从告警或通知日志回溯整条链路

### Phase 15: Harden Notification Retry Boundaries

**Goal**: 在现有异步通知实现内补齐有界重试、最终失败落点和尝试级上下文，降低瞬时失败导致的通知丢失风险。  
**Depends on**: Phase 14  
**Plans**: 2 plans

Plans:

- [x] 15-01: Add bounded retry behavior and attempt-level context to async notification delivery
- [x] 15-02: Verify terminal failure handling and retry diagnostics without changing the current stack

**Success criteria:**
- 瞬时失败通知不会在首次失败后直接结束
- 最终失败结果有明确日志或状态落点，不再只剩模糊报错
- 测试或验证文档覆盖首发成功、重试成功和重试耗尽三类场景

### Phase 16: Standardize Alert Path Logging

**Goal**: 统一告警主链路日志格式、字段命名和输出入口，减少散乱打印并强化可观测性契约。  
**Depends on**: Phase 15  
**Plans**: 2 plans

Plans:

- [x] 16-01-PLAN.md — Introduce one canonical alert-path event writer and migrate webhook call sites
- [x] 16-02-PLAN.md — Lock the logging contract in tests and capture Phase 16 verification truth (completed 2026-04-22)
- [x] 16-03-PLAN.md — Restore async panic correlation fields so failure logs remain traceable
- [x] 16-04-PLAN.md — Make canonical fields parse-safe for spaced values and refresh Phase 16 truth artifacts

**Success criteria:**
- 告警链路关键事件采用统一字段命名和格式
- 高风险路径中的 `fmt.Print*` 临时日志显著收口
- 仓库内存在可复用的日志样例或验证说明，后续扩展有据可循

### Phase 17: Clean Truth And Operational Docs

**Goal**: 继续清理历史命名和文档真相，并把 v1.3 的可靠性/可观测性行为沉淀到面向维护者的说明和验证产物。  
**Depends on**: Phase 16  
**Plans**: 3 plans

Plans:

- [ ] 17-01-PLAN.md — Rename low-risk verification entrypoints, tests, and repo-owned reference maps
- [ ] 17-02-PLAN.md — Refresh README/planning truth surfaces and clearly frame historical docs
- [ ] 17-03-PLAN.md — Create the maintainer alert-path runbook and Phase 17 truth artifacts

**Success criteria:**
- 仓库入口和阶段文档不再出现会误导当前系统定位的过期命名
- 可靠性与排障文档可直接指导维护者定位通知失败问题
- v1.3 的回滚关注点和 deferred items 被明确记录

## Next Step

- Use `/gsd-discuss-phase 17` to clarify Phase 17 implementation details, or `/gsd-plan-phase 17` to plan directly

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 10. Secure Realtime Alert Access | v1.2 | 2/2 | Complete | 2026-04-20 |
| 11. Restore Frontend Quality Baseline | v1.2 | 2/2 | Complete | 2026-04-20 |
| 12. Establish Automated Quality Gates | v1.2 | 2/2 | Complete | 2026-04-21 |
| 13. Harden Notification Delivery Path | v1.2 | 2/2 | Complete | 2026-04-21 |
| 14. Establish Alert Trace Context | v1.3 | 2/2 | Complete    | 2026-04-21 |
| 15. Harden Notification Retry Boundaries | v1.3 | 2/2 | Complete    | 2026-04-21 |
| 16. Standardize Alert Path Logging | v1.3 | 4/4 | Complete    | 2026-04-22 |
| 17. Clean Truth And Operational Docs | v1.3 | 0/3 | Planned | — |
