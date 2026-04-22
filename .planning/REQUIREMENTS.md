# Requirements: v1.3 Notification Reliability and Observability

**Status:** Drafted 2026-04-21
**Milestone:** v1.3 Notification Reliability and Observability
**Core Value:** 运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## Active Requirements

### Notification Reliability

- [x] **NTFY-01**: 通知发送遇到瞬时失败时会执行有界重试，而不是首次失败后直接结束
- [x] **NTFY-02**: 通知链路在超过重试上限后会留下明确的最终失败结果，便于值班人员确认不是静默丢失
- [x] **NTFY-03**: 每次通知尝试都会记录稳定的告警、渠道和尝试次数字段，便于比较首发与重试行为

### Alert Path Observability

- [x] **OBS-01**: Webhook 接入的单次告警处理会生成稳定的链路关联标识，并贯穿到后续关键处理阶段
- [x] **OBS-02**: 告警在接入、持久化、路由匹配和通知分发等关键阶段都会留下可检索的生命周期观测点
- [x] **OBS-03**: 运维排障时可以依据统一关联字段，从一条失败通知回溯到对应告警接入与处理上下文

### Logging Consistency

- [x] **LOG-01**: 告警主链路的后端日志使用统一字段命名和输出格式，避免同类事件字段漂移
- [x] **LOG-02**: 高风险链路中的临时 `fmt.Print*` 风格日志需要收口到统一日志入口，减少混杂输出
- [x] **LOG-03**: 新的日志约定需要在测试或验证文档中有明确样例，避免后续继续各写各的

### Naming And Documentation Truth

- [x] **DOCS-01**: 仓库入口、运行说明和里程碑文档继续保持当前游戏运维告警系统的真实命名，不把现状重新绑定到 AI 移除清理语境
- [x] **DOCS-02**: 与通知可靠性、告警排障相关的文档需要补充当前链路行为、失败诊断、维护者验证入口和回滚关注点
- [x] **DOCS-03**: v1.3 roadmap、phase 文档和验证记录需要准确反映可靠性/可观测性与真相分层目标，作为后续执行真源

## Future Requirements

- [ ] 持久化通知投递记录中心与独立重放/补发界面
- [ ] 基于 Prometheus / OpenTelemetry 的集中指标、trace 与告警面板
- [ ] 更复杂的通知重试策略，例如按渠道配置退避、熔断或死信队列
- [ ] Go module 路径、包名和所有历史文件名的全量改名迁移
- [ ] JWT issuer 与其他运行时历史命名的契约迁移

## Out of Scope

- 引入新的消息队列、任务调度器或外部 workflow 系统
- 引入新的可观测性平台或要求部署额外基础设施
- 改写现有告警接入、展示、路由、静默和值班主流程
- 借本轮文档清理顺带推动大规模技术栈或仓库结构迁移
- 在 Phase 17 中修改 `go.mod` module path、JWT issuer 或其他高风险运行时命名契约

## Traceability

| Requirement | Planned Phase |
|-------------|---------------|
| NTFY-01 | Phase 15 |
| NTFY-02 | Phase 15 |
| NTFY-03 | Phase 15 |
| OBS-01 | Phase 14 |
| OBS-02 | Phase 14 |
| OBS-03 | Phase 16 |
| LOG-01 | Phase 16 |
| LOG-02 | Phase 16 |
| LOG-03 | Phase 16 |
| DOCS-01 | Phase 17 |
| DOCS-02 | Phase 17 |
| DOCS-03 | Phase 17 |
