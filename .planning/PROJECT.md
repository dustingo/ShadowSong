# 游戏运维告警系统

## What This Is

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。v1.0 已完成 AI 能力移除，并补齐了通知模板原始事件透传与产品内模板预览能力；当前里程碑重点转向企业级用户体系与权限分配，确保不同岗位只能执行被授权的操作。

## Core Value

运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## Requirements

### Validated

- [x] 告警系统核心链路已在无 AI 前提下可用，包括告警接入、查看、处理、路由、静默和值班能力
- [x] 通知模板已支持标准告警字段与原始事件字段透传，并具备模板预览能力
- [x] 系统已具备基础 JWT 登录、用户模型和角色字段，可作为权限体系增强的基础
- [x] 系统已统一 `admin`、`operator`、`viewer` 角色真源，并建立可复用的 principal/capability 鉴权基线（Validated in Phase 5）

### Active

- [ ] 防止非授权用户修改其它用户资料、角色或系统配置
- [ ] 让前端菜单、页面和操作按钮与后端权限校验保持一致
- [ ] 为权限矩阵补齐验证与回归测试，避免再出现“任意登录用户可改配置”的问题
- [ ] 为关键用户与权限操作补齐审计日志、账号禁用与强制改密能力

### Out of Scope

- 细粒度自定义权限编辑器或任意 RBAC DSL - 本轮先落地固定角色分级，优先消除高风险越权问题
- SSO、LDAP、OAuth 企业身份集成 - 当前重点是系统内权限隔离，不做外部身份平台接入
- 组织、部门、项目空间级多租户权限模型 - 本轮只覆盖单系统内的角色授权与操作边界

## Current State

- 已发版版本：`v1.0 AI Removal Complete`（2026-04-10）
- 当前能力：后端 API、前端控制台、Webhook 接入、通知路由、静默规则、值班管理、模板预览、原始事件字段透传、统一角色常量、JWT principal、capability matrix 鉴权基线
- 已验证路径：后端无 AI 闭环脚本、前端无 AI 构建/残留扫描、模板 passthrough 端到端验证脚本
- 最新阶段：Phase 5 已完成，角色模型与鉴权基线已落地，后续进入用户管理边界与账号控制强化
- 当前 roadmap/requirements 已归档到 `.planning/milestones/`

## Current Milestone: v1.1 Enterprise Access Control

**Goal:** 重新设计企业级用户体系和权限分配，确保 `admin`、`operator`、`viewer` 角色在后端和前端都按最小权限原则执行操作。

**Target features:**
- 在保留现有 `admin`、`operator`、`viewer` 角色命名的前提下，补齐清晰一致的企业权限分级
- 用户管理接口收敛为“管理员管人、普通用户仅能维护安全的个人资料”
- 配置类与运维类接口按角色做服务端强制鉴权，明确仅 `admin` 可修改系统配置，`operator` 可处理告警
- 前端菜单、页面入口、按钮和错误提示按权限感知，避免展示无权操作入口
- 为权限矩阵补齐自动化验证和角色回归测试
- 为关键用户与权限变更加入审计日志、账号禁用和强制改密能力

## Context

- 当前 `User` 模型已经包含 `role` 字段，角色命名为 `admin`、`operator`、`viewer`，本轮继续沿用该命名
- 路由层只有 `/api/v1/users` 的部分接口使用了 `RequireRole("admin")`，大量配置入口如 `datasources/channels/routes/silences/onduty` 目前只要求登录即可修改
- `UpdateUser` 已限制“非 admin 不能修改其他用户”，但系统缺少统一的权限矩阵与界面收口
- 前端当前没有用户管理页面，也没有按角色隐藏菜单和操作；所有已登录用户都能进入配置页面
- 当前用户体系尚未覆盖账号禁用、强制改密和关键操作审计等安全能力
- 本轮需要在 brownfield 代码库上演进，优先修复越权风险，再考虑体验收口与验证

## Constraints

- **Tech stack**: 维持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite 技术栈
- **Brownfield**: 必须尊重仓库中的既有结构与未提交改动
- **Continuity**: 不能破坏告警接入、展示、路由、静默和值班主流程
- **Security**: 权限控制必须以后端强制校验为准，不能只依赖前端隐藏入口
- **Compatibility**: 现有用户登录与历史角色数据需要保持兼容，避免权限收口后出现不可预期的角色退化或账号不可用

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 先移除 AI 运行时和前端入口，再做文档与验证收口 | 先锁定核心告警主链路，降低回归风险 | ? Good |
| 模板能力以“标准字段 + 原始事件上下文”正式契约提供 | 满足自由透传 JSON 字段的同时保持旧模板兼容 | ? Good |
| 模板能力进入产品时同步提供后端驱动预览与脚本化验证 | 降低模板作者试错成本，避免隐式契约 | ? Good |
| 沿用 `admin`、`operator`、`viewer` 三档固定角色，而不改名 | 减少历史数据和前后端合同迁移成本，优先解决真实越权问题 | ✓ Good |
| 权限控制以后端拒绝未授权请求作为最终防线，前端负责可见性与体验收口 | 防止“隐藏按钮但接口仍可调用”的伪权限控制 | ✓ Good |
| 配置类写权限仅授予 `admin`，`operator` 保留告警处理能力 | 配置变更风险高于日常值班处理，需要明确职责分层 | ✓ Good |
| 将审计日志、账号禁用和强制改密纳入本里程碑 | 这些能力直接关系到账户控制安全性，不能只做静态角色收口 | ✓ Good |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check -> still the right priority?
3. Audit Out of Scope -> reasons still valid?
4. Update Context with current state

<details>
<summary>v1.0 Historical Context</summary>

- v1.0 包含 4 个 phase、12 个 plans，完整归档见 `.planning/milestones/v1.0-ROADMAP.md`
- 里程碑摘要见 `.planning/MILESTONES.md`
- 历史 requirements 见 `.planning/milestones/v1.0-REQUIREMENTS.md`

</details>

---
*Last updated: 2026-04-12 after Phase 5 completion*
