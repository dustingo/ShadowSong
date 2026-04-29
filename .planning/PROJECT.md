# 游戏运维告警系统

## What This Is

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。v1.0 已完成 AI 能力移除，并补齐了通知模板原始事件透传与产品内模板预览能力；v1.1 已完成企业级用户体系、权限收口、审计与验证链路建设；v1.2 已完成告警链路安全与工程化加固；v1.3 已完成通知链路可靠性、告警链路可观测性、统一日志契约和维护者运行文档收口。当前代码库已经具备安全收口后的实时告警访问、前端质量基线、自动化质量门禁、bounded notification retry、trace-backed alert-path observability、canonical webhook logging，以及维护者 alert-path runbook。

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
- [x] 系统已为 webhook 接入到通知分发主链路建立服务端 trace_id 真源，并补齐 ingest / persist / dedup / Redis / route_match / notification_entry 生命周期观测点（Validated in Phase 14）
- [x] 系统已为通知发送链路补齐有界三次重试、最终失败日志落点和尝试级上下文字段，瞬时失败不再在首次发送后直接静默结束（Validated in Phase 15）
- [x] 系统已统一 webhook 告警主链路日志输出入口、字段命名与可解析格式，并补齐 `async_panic` 失败路径的关联字段保真（Validated in Phase 16）
- [x] 系统已完成维护者 alert-path runbook、当前真相文档收口，以及 v1.3 可靠性/可观测性工件归档准备（Validated in Phase 17）

### Active

- [ ] 定义下一里程碑的业务目标、requirements 和 phase 范围
- [ ] 评估是否将通知投递记录中心、补发入口或更复杂的 retry policy 提升为正式 milestone
- [ ] 决定运行时历史命名迁移是否值得单独立项，而不是继续停留在文档层 deferred 状态

### Out of Scope

- 细粒度自定义权限编辑器或任意 RBAC DSL - 本轮先落地固定角色分级，优先消除高风险越权问题
- SSO、LDAP、OAuth 企业身份集成 - 当前重点是系统内权限隔离，不做外部身份平台接入
- 组织、部门、项目空间级多租户权限模型 - 本轮只覆盖单系统内的角色授权与操作边界
- 引入新的消息队列、集中式可观测性平台或大规模技术栈迁移 - v1.3 已证明现有 Go 应用内仍可继续增量提升主链路质量

## Current State

- 已发版版本：`v1.0 AI Removal Complete`（2026-04-10）、`v1.1 Enterprise Access Control`（2026-04-15）、`v1.2 Alert Pipeline Hardening`（2026-04-21）、`v1.3 Notification Reliability and Observability`（2026-04-29）
- 当前能力：后端 API、前端控制台、Webhook 接入、通知路由、静默规则、值班管理、模板预览、原始事件字段透传、统一角色常量、JWT principal、capability matrix 鉴权基线、管理员用户管理页、自助资料页、账号禁用、强制改密与旧会话失效、配置写接口权限收口、告警动作权限收口、持久化审计日志、权限感知 UI、只读配置视图、角色矩阵验证文档、实时 WebSocket 访问控制、前端 green 质量基线、GitHub Actions 质量门禁、通知链路 panic recover 与失败上下文字段、有界三次重试、最终失败落点、webhook trace_id 持久化、Redis/通知入口 trace 传播与生命周期阶段日志、统一 webhook 告警主链路日志输出入口、parse-safe 字段序列化、`async_panic` 关联字段保真，以及维护者 alert-path runbook 与 v1.3 audit archive
- 已验证路径：后端告警主链路验证脚本、前端控制台基线验证脚本、模板 passthrough 端到端验证脚本、角色矩阵前后端验证、禁用用户/强制改密/审计日志关键安全路径验证、前端 lint/test/build 验证、`go test ./...` 全量回归、v1.1 里程碑审计 `23/23 requirements` 与 `5/5 flows`、v1.3 里程碑审计 `12/12 requirements` 与 `4/4 integration flows`、Phase 14 trace/context verification、Phase 15 notification retry boundaries verification、Phase 16 standardized alert-path logging verification、Phase 17 docs/runbook/security truth verification
- 当前焦点：v1.3 已完成归档；下一步是定义下一个 milestone，而不是继续复用 v1.3 requirements
- 历史路线图、requirements 与里程碑审计保存在 `.planning/milestones/`，新一轮执行将沿现有 phase 编号继续推进
- 暂缓迁移边界：`go.mod` module path 与 JWT issuer 仍是历史遗留运行时契约；v1.3 只做文档标注和维护者说明，不在本轮推动运行时重命名

## Next Milestone Goals

- 定义一个新的 milestone 目标，而不是让现有 backlog 继续悬空
- 决定下一轮是优先做通知投递可恢复能力，还是做更深一层的可观测性与运维面板建设
- 明确哪些技术债需要进入正式 phase，哪些仍保留为 deferred background items

## Context

- 当前 `User` 模型已经包含 `role` 字段，角色命名为 `admin`、`operator`、`viewer`，本轮继续沿用该命名
- 后端路由、handler 和 capability middleware 已统一按 capability matrix 收口，不再依赖零散原始角色判断
- 前端已具备用户管理页、个人资料页、权限感知菜单、按钮显隐和只读提示，角色体验与后端授权边界保持一致
- 当前用户体系已覆盖账号禁用、强制改密、旧会话失效和关键安全操作审计日志
- v1.1 里程碑已确认 23/23 requirements、5/5 phases、15/15 plans 完成，权限治理基线已经稳定
- v1.2 继续在 brownfield 代码库上演进，不做技术迁移，优先补齐安全边界、工程门禁和通知可靠性短板
- 当前仓库前后端测试可运行，WebSocket 访问面、前端质量基线、自动化门禁和通知链路基础可靠性均已收口；后续工作应转入新里程碑定义或遗留增强项筛选
- v1.3 证明了 brownfield 小步增强策略仍然有效：不引入消息队列、集中观测平台或技术栈迁移，也能显著提升主链路可靠性与排障效率
- 当前真源文档应把历史 AI 移除放在已发布里程碑背景中，而不是继续作为运行入口的主叙事
- `go.mod` 中的 module path 与 `internal/auth/jwt.go` 中的 JWT issuer 仍属历史遗留运行时命名，当前只做文档标注和维护者说明，不做契约迁移

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
| Phase 14 先把 trace 真源和生命周期观测点锁在 webhook 主链路，再继续 Phase 15/16 的重试与日志统一 | 先建立稳定关联字段和阶段证据，后续可靠性与日志格式化才有一致排障真源 | ✓ Good |
| Phase 16 继续沿用 text-based `key=value` 日志契约，并通过稳定字段名、parse-safe quoting 与回归测试收口 webhook 告警主链路 | 在不扩大成 JSON logging 或全仓库迁移的前提下提升检索性和可解析性，保持 brownfield 改动面可控 | ✓ Good |
| Phase 17 只清理维护者入口、runbook 与真相文档，不改运行时历史命名契约 | 避免把 doc cleanup 扩大成高风险 runtime rename，先把当前维护与排障真源稳定下来 | ✓ Good |

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

<details>
<summary>v1.3 Historical Context</summary>

- v1.3 包含 4 个 phase、11 个 plans，完整归档见 `.planning/milestones/v1.3-ROADMAP.md`
- 历史 requirements 见 `.planning/milestones/v1.3-REQUIREMENTS.md`
- 里程碑审计见 `.planning/milestones/v1.3-MILESTONE-AUDIT.md`

</details>

---
*Last updated: 2026-04-29 after v1.3 milestone completion*
