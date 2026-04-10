# Roadmap: 游戏运维告警系统

## Overview

当前里程碑先完成了 AI 运行时、界面、配置和表述的完整移除，并在收尾阶段补上通知模板的原始事件透传能力与产品内预览路径。整体路线已经完成，核心告警链路与模板增强能力均有自动化或脚本化验证。

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Remove Backend AI Runtime** - 下线后端 AI 路由、处理器、客户端和直接依赖
- [x] **Phase 2: Remove Frontend AI Surfaces** - 清除前端 AI 页面、导航、展示字段和调用链
- [x] **Phase 3: Align Docs And Verification** - 清理剩余配置与文案并补足移除后的验证

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
- [x] 03-01-PLAN.md — Clean README, user-facing naming, and local env/config references
- [x] 03-02-PLAN.md — Refresh codebase maps and isolate remaining AI-only schema/data references
- [x] 03-03-PLAN.md — Add explicit non-AI backend/frontend verification paths and final evidence

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Remove Backend AI Runtime | 3/3 | Complete | 2026-04-09 |
| 2. Remove Frontend AI Surfaces | 3/3 | Complete | 2026-04-09 |
| 3. Align Docs And Verification | 3/3 | Complete | 2026-04-10 |
| 4. Enable raw event passthrough in notification templates | 3/3 | Complete | 2026-04-10 |

### Phase 4: Enable raw event passthrough in notification templates

**Goal:** 在不破坏现有标准告警模型和通知链路的前提下，让输出模板可直接访问原始 webhook 事件字段，并提供对用户友好的模板上下文与示例。
**Requirements**: [TMPL-01, TMPL-02, TMPL-03]
**Depends on:** Phase 3
**Success Criteria** (what must be TRUE):
  1. 数据源 `output_template` 在保留现有标准字段的同时，可以访问原始事件字段或扩展上下文。
  2. 用户可以通过稳定、可预期的模板变量写法引用嵌套 JSON 字段，而不必反向推测后端内部 `Alert` 模型。
  3. 现有仅依赖标准字段的通知模板继续可用，且至少有一条验证路径证明原始字段透传后的通知内容完整。
**Plans:** 3 plans

Plans:
- [x] 04-01-PLAN.md — Lock the backend output-template contract for standard fields plus raw event passthrough
- [x] 04-02-PLAN.md — Add datasource preview APIs and in-product template guidance for the new contract
- [x] 04-03-PLAN.md — Script end-to-end passthrough verification and record Phase 04 evidence (completed 2026-04-10)
