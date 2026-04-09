# Roadmap: 游戏运维告警系统

## Overview

当前里程碑的目标不是扩展新能力，而是把现有告警系统中与 AI 相关的运行时、界面、配置和表述完整移除，同时验证核心告警链路保持可用。路线按后端下线、前端清理、文档与验证收口三个阶段推进，确保每一阶段都能独立交付并降低回归风险。

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Remove Backend AI Runtime** - 下线后端 AI 路由、处理器、客户端和直接依赖
- [x] **Phase 2: Remove Frontend AI Surfaces** - 清除前端 AI 页面、导航、展示字段和调用链
- [ ] **Phase 3: Align Docs And Verification** - 清理剩余配置与文案并补足移除后的验证

## Phase Details

### Phase 1: Remove Backend AI Runtime
**Goal**: 后端不再暴露 AI 功能，且告警核心 API 与启动流程在无 AI 配置下继续正常工作。
**Depends on**: Nothing (first phase)
**Requirements**: [BEAI-01, BEAI-02, BEAI-03]
**Success Criteria** (what must be TRUE):
  1. Server startup no longer depends on AI config or `internal/ai` runtime wiring.
  2. Authenticated API no longer exposes `/api/v1/ai` routes or related handler registrations.
  3. Alert listing, stats, ack, quick silence, webhook ingestion, and notification paths still work after the AI removal changes.
**Plans**: 3 plans
**UI hint**: no

Plans:
- [x] 01-01-PLAN.md — Remove AI config, client, router, and handler wiring
- [x] 01-02-PLAN.md — Remove AI-only backend models, migrations, and runtime wording
- [x] 01-03-PLAN.md — Add non-AI backend regression tests and smoke verification

### Phase 2: Remove Frontend AI Surfaces
**Goal**: 前端不再暴露任何 AI 页面、入口、字段或调用链，应用导航与核心页面保持可用。
**Depends on**: Phase 1
**Requirements**: [FEAI-01, FEAI-02, FEAI-03]
**Success Criteria** (what must be TRUE):
  1. Navigation and routes no longer include the AI assistant page or related icons and labels.
  2. Dashboard, alerts, and shared components no longer render AI summaries, root-cause text, suggestions, or AI-trigger buttons.
  3. Frontend API clients and TypeScript types no longer include AI endpoints or AI-only fields.
**Plans**: 3 plans
**UI hint**: yes

Plans:
- [x] 02-01-PLAN.md — Remove AI page, route, and navigation entry points
- [x] 02-02-PLAN.md — Remove AI data rendering from alerts and dashboard flows
- [x] 02-03-PLAN.md — Remove AI API client and type dependencies plus frontend build verification

### Phase 3: Align Docs And Verification
**Goal**: 项目文档、品牌文案、环境配置和基本验证与“无 AI 版本”的实际状态保持一致。
**Depends on**: Phase 2
**Requirements**: [DATA-01, DATA-02, VER-01, VER-02]
**Success Criteria** (what must be TRUE):
  1. README, page title, and user-facing copy no longer describe the product as an AI alert system.
  2. Environment/config documentation no longer requires AI-related variables for normal setup.
  3. Backend and frontend each have at least one verification path confirming the non-AI build still works.
**Plans**: 3 plans
**UI hint**: yes

Plans:
- [ ] 03-01-PLAN.md — Clean README, user-facing naming, and local env/config references
- [ ] 03-02-PLAN.md — Refresh codebase maps and isolate remaining AI-only schema/data references
- [ ] 03-03-PLAN.md — Add explicit non-AI backend/frontend verification paths and final evidence

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Remove Backend AI Runtime | 3/3 | Complete | 2026-04-09 |
| 2. Remove Frontend AI Surfaces | 3/3 | Complete | 2026-04-09 |
| 3. Align Docs And Verification | 0/TBD | Not started | - |
