# Roadmap: 游戏运维告警系统

## Milestones

- ✅ **v1.0 AI Removal Complete** — Phases 1-4 (shipped 2026-04-10). Archive: `.planning/milestones/v1.0-ROADMAP.md`
- ✅ **v1.1 Enterprise Access Control** — Phases 5-9 (shipped 2026-04-15). Archive: `.planning/milestones/v1.1-ROADMAP.md`
- 🚧 **v1.2 Alert Pipeline Hardening** — Phases 10-13 (in progress)

## Overview

v1.2 围绕“实时告警访问面是否安全、前端质量门禁是否恢复、通知链路是否可可靠排障”三条主线推进。先收紧 WebSocket 订阅入口的服务端鉴权与来源校验，再修复前端 lint 红线并恢复本地质量基线，随后接入 CI 把测试、lint 与 build 变成自动门禁，最后补齐 webhook 异步通知的 panic 防护与失败追踪，确保系统在现有非 AI 基线之上继续稳定演进。

## Phases

**Phase Numbering:**
- Integer phases (10, 11, 12, 13): Planned milestone work
- Decimal phases (10.1, 10.2): Urgent insertions if later needed

<details>
<summary>✅ v1.0 AI Removal Complete (Phases 1-4) - SHIPPED 2026-04-10</summary>

See `.planning/milestones/v1.0-ROADMAP.md`

</details>

<details>
<summary>✅ v1.1 Enterprise Access Control (Phases 5-9) - SHIPPED 2026-04-15</summary>

See `.planning/milestones/v1.1-ROADMAP.md`

</details>

### 🚧 v1.2 Alert Pipeline Hardening (In Progress)

**Milestone Goal:** 在不改变现有技术栈和核心告警产品能力边界的前提下，加固实时告警访问安全、恢复工程质量门禁，并提升通知分发链路的基础可靠性。

- [x] **Phase 10: Secure Realtime Alert Access** - 收紧 WebSocket 告警流鉴权与来源限制
- [x] **Phase 11: Restore Frontend Quality Baseline** - 修复前端 lint 红线并收口关键页面质量问题
- [x] **Phase 12: Establish Automated Quality Gates** - 接入 CI 并同步收口文档与工程命名基线
- [ ] **Phase 13: Harden Notification Delivery Path** - 加固 webhook 异步通知的 panic 防护、失败日志与追踪链路

## Phase Details

### Phase 10: Secure Realtime Alert Access
**Goal**: 让实时告警 WebSocket 流只对经过后端授权的合法客户端开放，并对非法来源请求明确拒绝。  
**Depends on**: Phase 9  
**Requirements**: [RTAL-01, RTAL-02, RTAL-03]
**Success Criteria** (what must be TRUE):
  1. 未登录请求不能成功连接 `/ws/alerts`
  2. 来源不在允许列表中的连接请求会被拒绝
  3. 已授权用户仍能看到现有实时告警能力，不破坏当前主流程
**Plans**: 2 plans

Plans:
- [x] 10-01: Add JWT-backed WebSocket auth and reject unauthorized realtime sessions
- [x] 10-02: Make allowed origins configurable and verify allow/deny websocket cases

### Phase 11: Restore Frontend Quality Baseline
**Goal**: 清理前端 lint 红线并收口会持续制造噪音或潜在缺陷的关键质量问题，让前端本地质量基线恢复为 green。  
**Depends on**: Phase 10  
**Requirements**: [FEQ-01, FEQ-02, FEQ-03]
**Success Criteria** (what must be TRUE):
  1. `pnpm lint` 可以在当前仓库默认环境下通过
  2. 修复 lint 后前端测试与生产构建仍通过
  3. 高风险 hook 依赖和明显无效变量问题不再作为持续噪音存在
**Plans**: 2 plans

Plans:
- [x] 11-01: Fix frontend lint errors and remove obvious warning-level debt in critical pages
- [x] 11-02: Re-verify frontend test and build flows after lint cleanup

### Phase 12: Establish Automated Quality Gates
**Goal**: 把后端测试、前端 lint、前端测试和前端构建串成自动化门禁，并同步更新本轮真相文档与工程命名。  
**Depends on**: Phase 11  
**Requirements**: [CIV-01, CIV-02, CIV-03, DOCS-01, DOCS-02]
**Success Criteria** (what must be TRUE):
  1. 仓库存在可自动执行的 CI 工作流，覆盖约定的四类检查
  2. 失败步骤在 CI 输出中可直接定位
  3. README、planning 文档和工程命名与“非 AI 告警系统”现状保持一致
**Plans**: 2 plans

Plans:
- [x] 12-01: Add CI workflow for backend tests and frontend lint/test/build gates
- [x] 12-02: Align docs and naming with the v1.2 hardening milestone truth

### Phase 13: Harden Notification Delivery Path
**Goal**: 为 webhook 异步通知链路补上 panic 防护、失败日志和最小可追踪性，降低通知静默失败和排障盲区。  
**Depends on**: Phase 12  
**Requirements**: [NTFY-01, NTFY-02, NTFY-03]
**Success Criteria** (what must be TRUE):
  1. 通知异步处理即使发生 panic 也不会直接拖垮服务进程
  2. 通知失败日志能关联到告警或渠道上下文
  3. 通知链路加固后现有 webhook 入库、路由和发送主流程仍可工作
**Plans**: 2 plans

Plans:
- [ ] 13-01: Add panic recovery and failure boundaries around async notification processing
- [ ] 13-02: Improve notification failure logging and traceability verification

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 10. Secure Realtime Alert Access | v1.2 | 2/2 | Complete | 2026-04-20 |
| 11. Restore Frontend Quality Baseline | v1.2 | 2/2 | Complete | 2026-04-20 |
| 12. Establish Automated Quality Gates | v1.2 | 2/2 | Complete | 2026-04-21 |
| 13. Harden Notification Delivery Path | v1.2 | 0/2 | Not started | - |
