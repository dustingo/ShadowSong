# Requirements: 游戏运维告警系统

**Defined:** 2026-04-20
**Core Value:** 运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## v1 Requirements

### Real-Time Alert Access

- [x] **RTAL-01**: 只有已登录且通过服务端鉴权的用户才能建立实时告警 WebSocket 连接
- [x] **RTAL-02**: WebSocket 连接会对来源域名执行显式校验，避免任意站点直接订阅告警流
- [x] **RTAL-03**: 未授权或来源不合法的实时连接请求会得到明确拒绝，而不是静默建立连接

### Frontend Quality Gate

- [x] **FEQ-01**: 前端代码库中的 lint error 必须清零，`pnpm lint` 在默认项目环境下可以通过
- [x] **FEQ-02**: 前端高风险的 hook 依赖和无效变量问题需要收口到可持续维护的状态，避免继续在关键页面积累质量债
- [x] **FEQ-03**: 前端生产构建和现有测试在修复 lint 后仍然能够通过，不引入新的运行时错误

### CI And Verification

- [ ] **CIV-01**: 仓库提供自动执行的 CI 流程，至少覆盖 `go test ./...`
- [ ] **CIV-02**: CI 同时覆盖前端 `pnpm lint`、`pnpm test -- --run` 和 `pnpm build`
- [ ] **CIV-03**: 质量门禁失败时能够明确暴露失败步骤，便于在合并前阻断回归

### Notification Reliability

- [ ] **NTFY-01**: Webhook 异步通知处理发生 panic 时不会直接把服务进程带崩
- [ ] **NTFY-02**: 通知发送失败时会留下结构化或至少稳定可检索的后端日志，便于定位失败原因
- [ ] **NTFY-03**: 通知链路关键失败点需要可追踪到具体告警或渠道上下文，而不是只有模糊报错

### Documentation Alignment

- [ ] **DOCS-01**: 项目文档、工程入口和命名继续保持“非 AI 告警系统”的真实表述，不引入过期 AI 名称或误导性说明
- [ ] **DOCS-02**: 新里程碑的 requirements、roadmap 和 state 文档应准确反映当前目标与执行顺序，能直接作为后续 phase 的真相来源

## v2 Requirements

### Access Governance Follow-up

- **ACL-01**: 支持更细粒度的权限点或授权继承，而不只依赖固定角色集合
- **ACL-02**: 支持 SSO / LDAP / OAuth 等企业身份源集成
- **ACL-03**: 支持关键管理操作审批流或更强治理能力

## Out of Scope

| Feature | Reason |
|---------|--------|
| 自定义 RBAC 编辑器 | 本轮优先修复现有安全和工程短板，不扩展到完整授权系统设计 |
| 外部身份源接入 | 当前目标是加固现有系统内认证与告警链路，不引入 SSO / LDAP / OAuth 集成复杂度 |
| 大规模通知任务队列重构 | 本轮只做基础可靠性增强与可观测性补齐，不扩展为全新的异步基础设施 |
| 技术栈迁移 | 约束明确要求保持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| RTAL-01 | Phase 10 | Complete |
| RTAL-02 | Phase 10 | Complete |
| RTAL-03 | Phase 10 | Complete |
| FEQ-01 | Phase 11 | Complete |
| FEQ-02 | Phase 11 | Complete |
| FEQ-03 | Phase 11 | Complete |
| CIV-01 | Phase 12 | Pending |
| CIV-02 | Phase 12 | Pending |
| CIV-03 | Phase 12 | Pending |
| NTFY-01 | Phase 13 | Pending |
| NTFY-02 | Phase 13 | Pending |
| NTFY-03 | Phase 13 | Pending |
| DOCS-01 | Phase 12 | Pending |
| DOCS-02 | Phase 12 | Pending |

**Coverage:**
- v1 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0

---
*Requirements defined: 2026-04-20*
*Last updated: 2026-04-20 after Phase 11 completion*
