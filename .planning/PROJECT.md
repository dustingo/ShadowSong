# 游戏运维告警系统

## What This Is

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。系统已经具备后端 API、前端控制台、Webhook 接入、通知路由、静默规则和值班管理能力，当前重点是将已有的 AI 能力和相关配置从产品与代码中完整移除。

## Core Value

运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## Requirements

### Validated

- ✓ 用户可以通过 Webhook 接入外部告警并落库为统一告警模型 — existing
- ✓ 用户可以在前端查看告警列表、统计信息和实时更新 — existing
- ✓ 用户可以管理数据源、通知渠道、路由规则、静默规则和值班配置 — existing
- ✓ 用户可以通过认证后的 API 和前端界面执行日常告警处理操作 — existing
- ✓ 移除后端 AI 路由、处理器、客户端和环境配置，服务在无 AI 配置下仍可启动 — Validated in Phase 1
- ✓ AI 移除后的后端关键路径已有自动化验证，Webhook、通知分发、告警确认与快速静默闭环可跑通 — Validated in Phase 1
- ✓ 移除前端 AI 页面、入口、调用链和 AI 展示字段，界面不再残留 AI 功能 — Validated in Phase 2

### Active

- [ ] 清理 AI 相关数据库模型使用、文案、文档和品牌命名，使项目表述与实际能力一致

### Out of Scope

- 保留或重做 AI 助手能力 — 当前目标是完整移除，不做替代方案设计
- 大规模重构现有告警处理架构 — 本轮只做与 AI 移除直接相关的清理
- 新增非 AI 的高级分析能力 — 避免把“移除 AI”扩展成新功能开发

## Context

- 仓库是 brownfield 项目，已经有可运行的 Go 后端和 React 前端。
- 后端主要位于 `cmd/server/main.go`、`internal/router/router.go`、`internal/handlers/`、`internal/models/`、`internal/notifier/`。
- 前端主要位于 `frontend/src/App.tsx`、`frontend/src/pages/`、`frontend/src/components/`、`frontend/src/api/client.ts`、`frontend/src/types/index.ts`。
- 现有 AI 能力分散在 `internal/ai/client.go`、`internal/handlers/ai.go`、`internal/config/config.go`、`internal/models/alert.go`、`internal/models/models.go`、`frontend/src/pages/AIAssistant.tsx`、`frontend/src/components/AlertCard.tsx`、`frontend/src/pages/Dashboard.tsx`、`frontend/src/pages/Alerts.tsx` 等位置。
- `.planning/codebase/` 已完成代码库地图，可作为后续 phase 规划输入。
- 当前工作树存在用户未提交改动，初始化与后续规划必须避免覆盖这些改动。
- Phase 1 已完成：后端 AI 运行时、路由和主要持久化残留已移除。
- Phase 2 已完成：前端 AI 页面、导航入口、AI 展示字段、共享 API 合同和触达品牌文案已清理，并通过前端构建验证。

## Constraints

- **Tech stack**: 维持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite 技术栈 — 本轮不做技术迁移
- **Brownfield**: 必须尊重当前仓库中的既有结构与未提交改动 — 避免误删或回退无关修改
- **Continuity**: AI 移除后核心告警流程仍需可用 — 不能破坏告警接入、展示、路由、静默和值班能力
- **Frontend compatibility**: 前端移除 AI 后路由、菜单、类型和 API 调用要保持自洽 — 不能留下断链入口或运行时错误
- **Documentation**: 项目名称、README、页面标题和测试文案需要反映“非 AI 告警系统”的现状 — 避免品牌和能力描述不一致

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 先完成 GSD brownfield 初始化，再进入具体 phase 规划 | 当前仓库缺少 ROADMAP/STATE，无法直接执行 `/gsd-plan-phase` | ✓ Good |
| 当前主动工作定义为“移除项目中的 AI 能力，包括前端” | 这是用户当前明确提出的目标，后续 roadmap 与 plan 都围绕该目标展开 | ✓ Good |
| 保留现有告警主流程，AI 只做删减不做替代实现 | 降低范围扩张风险，保证本轮工作可控 | ✓ Good |
| 先完成后端 AI 下线与闭环验证，再继续前端清理 | 先锁定服务启动、Webhook、通知和告警处理主链路，减少后续前端阶段的回归面 | ✓ Good |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition**:
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone**:
1. Full review of all sections
2. Core Value check - still the right priority?
3. Audit Out of Scope - reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-09 after Phase 2 completion*
