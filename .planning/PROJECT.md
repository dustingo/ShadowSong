# 游戏运维告警系统

## What This Is

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。v1.0 已完成 AI 能力移除，并补齐了通知模板原始事件透传与产品内模板预览能力；v1.1 已完成企业级用户体系、权限收口、审计与验证链路建设；v1.2 已完成告警链路安全与工程化加固。当前代码库已经具备安全收口后的实时告警访问、前端质量基线、自动化质量门禁和基础通知可靠性，下一轮将聚焦通知链路可靠性、告警链路可观测性和工程真相继续收口。

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
- [x] 系统已完成实时告警 WebSocket 鉴权与来源限制收口，实时订阅入口不再对匿名或非法来源公开（Validated in Phase 10）
- [x] 系统已恢复前端本地质量基线，`pnpm lint`、`pnpm test -- --run` 与 `pnpm build` 均可通过（Validated in Phase 11）
- [x] 系统已建立 GitHub Actions 质量门禁，覆盖后端测试与前端 lint/test/build，并同步收口低风险工程命名与真相文档（Validated in Phase 12）
- [x] 系统已加固 webhook 异步通知链路，通知 goroutine panic 不再裸奔，失败日志具备基础可追踪性（Validated in Phase 13）

### Active

- [ ] 通知发送链路需要从“基础失败可排查”提升到“瞬时失败可恢复、最终失败有明确落点”的可靠性基线
- [ ] 告警从 webhook 接入到通知分发的关键阶段需要具备统一的关联标识和生命周期观测点，降低跨链路排障成本
- [ ] 后端告警链路日志需要统一格式和字段约定，减少 `fmt` 风格散乱日志带来的检索困难
- [ ] 历史命名和文档真相需要继续清理，确保仓库入口、运行说明和阶段文档反映当前非 AI 告警系统现状

### Out of Scope

- 细粒度自定义权限编辑器或任意 RBAC DSL - 本轮先落地固定角色分级，优先消除高风险越权问题
- SSO、LDAP、OAuth 企业身份集成 - 当前重点是系统内权限隔离，不做外部身份平台接入
- 组织、部门、项目空间级多租户权限模型 - 本轮只覆盖单系统内的角色授权与操作边界

## Current State

- 已发版版本：`v1.0 AI Removal Complete`（2026-04-10）、`v1.1 Enterprise Access Control`（2026-04-15）、`v1.2 Alert Pipeline Hardening`（2026-04-21）
- 当前能力：后端 API、前端控制台、Webhook 接入、通知路由、静默规则、值班管理、模板预览、原始事件字段透传、统一角色常量、JWT principal、capability matrix 鉴权基线、管理员用户管理页、自助资料页、账号禁用、强制改密与旧会话失效、配置写接口权限收口、告警动作权限收口、持久化审计日志、权限感知 UI、只读配置视图、角色矩阵验证文档、实时 WebSocket 访问控制、前端 green 质量基线、GitHub Actions 质量门禁、通知链路 panic recover 与失败上下文日志
- 已验证路径：后端无 AI 闭环脚本、前端无 AI 构建/残留扫描、模板 passthrough 端到端验证脚本、角色矩阵前后端验证、禁用用户/强制改密/审计日志关键安全路径验证、前端 lint/test/build 验证、`go test ./...` 全量回归、v1.1 里程碑审计 `23/23 requirements` 与 `5/5 flows`
- 最新阶段：v1.3 `Notification Reliability and Observability` 已启动，当前正在定义 requirements 与 roadmap
- 历史路线图与 requirement 基线保存在 `.planning/milestones/`，新一轮执行将沿现有 phase 编号继续推进

## Current Milestone: v1.3 Notification Reliability and Observability

**Goal:** 在不引入新技术栈和不破坏现有告警主流程的前提下，把通知发送链路提升到更可恢复、更易排障的工程基线，并继续补齐告警链路观测与文档真相。

**Target features:**
- 为通知发送补上有界重试、最终失败落点和更稳定的失败诊断信息
- 为 webhook 接入、告警落库、路由命中和通知发送建立统一关联标识与生命周期观测点
- 统一告警链路日志格式与字段约定，逐步替换高风险路径中的临时打印
- 继续清理历史命名和文档真相，确保“非 AI 告警系统”的工程表述保持一致

## Context

- 当前 `User` 模型已经包含 `role` 字段，角色命名为 `admin`、`operator`、`viewer`，本轮继续沿用该命名
- 后端路由、handler 和 capability middleware 已统一按 capability matrix 收口，不再依赖零散原始角色判断
- 前端已具备用户管理页、个人资料页、权限感知菜单、按钮显隐和只读提示，角色体验与后端授权边界保持一致
- 当前用户体系已覆盖账号禁用、强制改密、旧会话失效和关键安全操作审计日志
- v1.1 里程碑已确认 23/23 requirements、5/5 phases、15/15 plans 完成，权限治理基线已经稳定
- v1.2 继续在 brownfield 代码库上演进，不做技术迁移，优先补齐安全边界、工程门禁和通知可靠性短板
- 当前仓库前后端测试可运行，WebSocket 访问面、前端质量基线、自动化门禁和通知链路基础可靠性均已收口；后续工作应转入新里程碑定义或遗留增强项筛选
- v1.3 继续坚持 brownfield 小步增强，不引入消息队列、集中观测平台或技术栈迁移，而是在现有 Go 应用内先补齐可靠性与可观测性基线

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
| v1.3 先在现有应用内补齐通知重试、关联标识和日志统一，不引入新队列或外部观测平台 | 先以最小迁移成本提升主链路可靠性和排障效率，避免在 brownfield 阶段把问题扩大成基础设施改造 | ✓ Good |

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
*Last updated: 2026-04-21 after milestone v1.3 initialization*
