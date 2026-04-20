# 游戏运维告警系统

## What This Is

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。v1.0 已完成 AI 能力移除，并补齐了通知模板原始事件透传与产品内模板预览能力；v1.1 已完成企业级用户体系、权限收口、审计与验证链路建设。当前里程碑为 v1.2 `Alert Pipeline Hardening`，重点转向告警链路安全加固、工程化门禁补齐与通知可靠性提升。

## Core Value

运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## Requirements

### Validated

- [x] 告警系统核心链路已在无 AI 前提下可用，包括告警接入、查看、处理、路由、静默和值班能力
- [x] 通知模板已支持标准告警字段与原始事件字段透传，并具备模板预览能力
- [x] 系统已具备基础 JWT 登录、用户模型和角色字段，可作为权限体系增强的基础
- [x] 系统已统一 `admin`、`operator`、`viewer` 角色真源，并建立可复用的 principal/capability 鉴权基线（Validated in Phase 5）
- [x] 系统已完成管理员管人与普通用户自助资料/密码边界拆分，并落地账号禁用、强制改密和旧会话失效控制（Validated in Phase 6）
- [x] 系统已对配置写接口与告警动作完成后端强制权限收口，并为关键用户/权限操作落地持久化审计日志（Validated in Phase 7）
- [x] 系统已完成权限感知 UI、只读/拒绝提示、角色矩阵验证文档与关键安全路径验证，23/23 个 v1 requirement 已完成（Validated in Phase 8）

### Active

- [ ] WebSocket 告警流必须具备鉴权与可配置来源限制，未授权请求不能直接订阅实时告警数据
- [ ] 前端质量门禁需要恢复为 green 状态，并建立自动执行的 CI 检查，覆盖后端测试、前端 lint、前端测试和前端构建
- [ ] Webhook 异步通知链路需要具备基础的 panic 防护、失败可观测性与可追踪日志，降低静默丢通知风险
- [ ] 项目文档、命名和工程入口应继续与“非 AI 告警系统”现状保持一致，避免后续里程碑在错误基线上迭代

### Out of Scope

- 细粒度自定义权限编辑器或任意 RBAC DSL - 本轮先落地固定角色分级，优先消除高风险越权问题
- SSO、LDAP、OAuth 企业身份集成 - 当前重点是系统内权限隔离，不做外部身份平台接入
- 组织、部门、项目空间级多租户权限模型 - 本轮只覆盖单系统内的角色授权与操作边界

## Current State

- 已发版版本：`v1.0 AI Removal Complete`（2026-04-10）、`v1.1 Enterprise Access Control`（2026-04-15）
- 当前能力：后端 API、前端控制台、Webhook 接入、通知路由、静默规则、值班管理、模板预览、原始事件字段透传、统一角色常量、JWT principal、capability matrix 鉴权基线、管理员用户管理页、自助资料页、账号禁用、强制改密与旧会话失效、配置写接口权限收口、告警动作权限收口、持久化审计日志、权限感知 UI、只读配置视图、角色矩阵验证文档
- 已验证路径：后端无 AI 闭环脚本、前端无 AI 构建/残留扫描、模板 passthrough 端到端验证脚本、角色矩阵前后端验证、禁用用户/强制改密/审计日志关键安全路径验证、v1.1 里程碑审计 `23/23 requirements` 与 `5/5 flows`
- 最新阶段：`v1.2 Alert Pipeline Hardening` 已启动，当前处于 requirements / roadmap 定义完成、等待 Phase 10 上下文收敛的状态
- 旧 milestone phase 目录已清理，历史路线图与 requirement 基线保存在 `.planning/milestones/`，新一轮执行将从 Phase 10 继续编号

## Next Milestone Goals

**v1.2 Alert Pipeline Hardening**
- 收紧 WebSocket 告警流的认证与来源控制，堵住实时告警流的公开访问面
- 收口前端 lint / build 质量问题，并建立 CI 质量门禁，防止后续回归静默进入主干
- 提升 webhook 异步通知的可靠性、失败日志与可追踪性，减少丢通知和排障盲区
- 统一本轮涉及的文档、命名与工程入口，确保项目继续沿“非 AI 告警系统”的真相演进

## Current Milestone: v1.2 Alert Pipeline Hardening

**Goal:** 在不改变现有技术栈和核心产品能力边界的前提下，加固告警实时链路安全性、补齐工程质量门禁，并提高通知分发链路的运行可靠性。

**Target features:**
- WebSocket 实时告警流鉴权与来源限制
- 前端 lint 修复与持续集成质量门禁
- Webhook 异步通知链路 panic 防护、失败日志与追踪改进
- 文档与命名对齐，保持“非 AI 告警系统”表述一致

## Context

- 当前 `User` 模型已经包含 `role` 字段，角色命名为 `admin`、`operator`、`viewer`，本轮继续沿用该命名
- 后端路由、handler 和 capability middleware 已统一按 capability matrix 收口，不再依赖零散原始角色判断
- 前端已具备用户管理页、个人资料页、权限感知菜单、按钮显隐和只读提示，角色体验与后端授权边界保持一致
- 当前用户体系已覆盖账号禁用、强制改密、旧会话失效和关键安全操作审计日志
- v1.1 里程碑已确认 23/23 requirements、5/5 phases、15/15 plans 完成，权限治理基线已经稳定
- v1.2 继续在 brownfield 代码库上演进，不做技术迁移，优先补齐安全边界、工程门禁和通知可靠性短板
- 当前仓库前后端测试可运行，但前端 lint 仍为红线，WebSocket 与通知链路存在应优先收口的运行风险

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

<details>
<summary>v1.1 Historical Context</summary>

- v1.1 包含 5 个 phase、15 个 plans，完整归档见 `.planning/milestones/v1.1-ROADMAP.md`
- 历史 requirements 见 `.planning/milestones/v1.1-REQUIREMENTS.md`
- 里程碑审计与收口记录见 `.planning/v1.1-MILESTONE-AUDIT.md`

</details>

---
*Last updated: 2026-04-20 after v1.2 milestone start*
