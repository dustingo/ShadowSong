# Roadmap: 游戏运维告警系统

## Milestones

- ? **v1.0 AI Removal Complete** — Phases 1-4 (shipped 2026-04-10). Archive: `.planning/milestones/v1.0-ROADMAP.md`
- ? **v1.1 Enterprise Access Control** — Planned Phases 5-8. Focus: 企业级用户体系、权限收口与账号控制安全

## Current Status

当前里程碑为 `v1.1 Enterprise Access Control`。Phase 5 已完成，Phase 6 已完成规划，下一步进入用户管理边界与账号控制实现。

## Next Step

- `/gsd-execute-phase 6` — 执行 Phase 6 的账号禁用、强制改密、自助资料与管理员管人边界改造
- `/gsd-discuss-phase 7` — 在需要提前澄清审计日志与受保护操作矩阵时使用

## Overview

这个里程碑围绕“谁能看、谁能改、谁能管人、关键动作如何留痕”四条主线推进。先统一现有角色模型并把权限判定收口到后端，再完成用户管理边界、账号控制和配置接口加固，最后补上前端显隐与验证，确保系统从“登录即可大部分可改”演进到“最小权限、服务端强制控制、关键操作可审计”的企业级基线。

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 5: Normalize Role Model** - 统一现有角色语义并建立可复用的权限判定基线
- [ ] **Phase 6: Secure User Administration** - 收紧用户管理和个人资料修改边界，并加入账号禁用与强制改密
- [ ] **Phase 7: Lock Down Protected Operations** - 对配置与告警操作接口补齐角色校验，并落地关键操作审计
- [ ] **Phase 8: Ship Permission-Aware UI And Verification** - 让前端入口与按钮按权限收口，并补齐角色矩阵验证

## Phase Details

### Phase 5: Normalize Role Model
**Goal**: 在不破坏现有登录链路的前提下，统一 `admin`、`operator`、`viewer` 角色语义，并建立后端可复用的权限判定基线。
**Depends on**: Phase 4
**Requirements**: [AUTHZ-01, AUTHZ-02, AUTHZ-03]
**Success Criteria** (what must be TRUE):
  1. 用户模型、校验逻辑和鉴权上下文只接受 `admin`、`operator`、`viewer` 三种受支持角色。
  2. 现有账号在权限体系升级后继续可登录可鉴权，不需要承担角色改名迁移成本。
  3. 后端存在统一的权限判断方式，可被后续用户接口和配置接口复用。
**Plans**: 3 plans
**UI hint**: no

Plans:
- [x] 05-01-PLAN.md — Normalize role constants, validation, and token/session claims around existing role names
- [x] 05-02-PLAN.md — Audit and harden existing user/bootstrap data compatibility without role renaming
- [x] 05-03-PLAN.md — Introduce reusable authorization helpers and backend permission matrix tests

### Phase 6: Secure User Administration
**Goal**: 让用户管理明确区分“管理员管理用户”和“普通用户维护个人资料”，杜绝越权改其它用户或自提权，并补齐账号禁用与强制改密控制。
**Depends on**: Phase 5
**Requirements**: [USER-01, USER-02, USER-03, USER-04, USER-05, USER-06]
**Success Criteria** (what must be TRUE):
  1. `admin` 可以查看、创建、修改和删除用户，并安全地分配角色。
  2. 非 `admin` 不能修改其他用户资料、角色或删除其他用户。
  3. 普通用户仅能修改允许自助维护的个人字段与密码，不能通过接口提升自己的权限。
  4. 被禁用用户不能继续登录或继续使用既有会话访问受保护资源，强制改密用户在完成改密前不能继续正常使用系统。
  5. 用户管理关键边界具备明确错误返回和回归覆盖。
**Plans**: 3 plans
**UI hint**: yes

Plans:
- [ ] 06-01-PLAN.md — Add account state and session invalidation foundations for disabled and forced-reset users
- [ ] 06-02-PLAN.md — Split admin-managed user operations from self-service profile/password flows
- [ ] 06-03-PLAN.md — Expose the minimal frontend users/profile surfaces for Phase 6

### Phase 7: Lock Down Protected Operations
**Goal**: 对系统配置和运维操作接口补齐角色校验，并为关键安全动作建立后端审计，确保不同角色只能执行其职责范围内的动作。
**Depends on**: Phase 6
**Requirements**: [PERM-01, PERM-02, PERM-03, PERM-04, AUDIT-01, AUDIT-02, AUDIT-03]
**Success Criteria** (what must be TRUE):
  1. 只有 `admin` 可以修改数据源、渠道、路由规则、静默规则和值班配置。
  2. `operator` 可以查看配置并处理告警，包括确认和快速静默，但不能修改用户角色或系统配置。
  3. `viewer` 只能查看信息，不能确认告警、快速静默或做任何配置变更。
  4. 关键用户与权限操作具备后端审计日志，至少记录操作者、目标对象、动作和时间。
  5. 所有未授权请求都返回一致的拒绝结果并可被测试验证。
**Plans**: 3 plans
**UI hint**: yes

Plans:
- [ ] 07-01-PLAN.md — Apply role guards to alert action endpoints and configuration route groups
- [ ] 07-02-PLAN.md — Refactor handler-level authorization and add audit logging for critical account/config actions
- [ ] 07-03-PLAN.md — Add endpoint-level regression coverage for role allow/deny and audit cases

### Phase 8: Ship Permission-Aware UI And Verification
**Goal**: 让前端菜单、页面、按钮和提示与后端权限边界一致，并交付完整的角色矩阵验证与文档说明。
**Depends on**: Phase 7
**Requirements**: [FEACL-01, FEACL-02, FEACL-03, VER-01, VER-02, VER-03, VER-04]
**Success Criteria** (what must be TRUE):
  1. 不同角色登录后看到的菜单、页面入口和关键按钮与其后端权限一致。
  2. 用户访问无权操作的页面或动作时，会得到清晰一致的前端提示，而不是空白或误导性成功反馈。
  3. 至少有一组后端和一组前端验证覆盖三类角色的关键权限矩阵，以及账号禁用、强制改密和审计日志关键路径。
  4. 角色说明、默认行为和限制被记录到文档或验证说明中，便于后续维护。
**Plans**: 3 plans
**UI hint**: yes

Plans:
- [ ] 08-01-PLAN.md — Add frontend role-aware menu, route, and action visibility controls
- [ ] 08-02-PLAN.md — Normalize forbidden-state UX and role-specific page behavior
- [ ] 08-03-PLAN.md — Document the access model and add end-to-end verification evidence
