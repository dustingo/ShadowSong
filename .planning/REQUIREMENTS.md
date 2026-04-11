# Requirements: 游戏运维告警系统

**Defined:** 2026-04-11
**Core Value:** 运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## v1 Requirements

### Authorization Model

- [ ] **AUTHZ-01**: 系统使用统一且受约束的内置角色集 `admin`、`operator`、`viewer`，用户不能保存为未定义角色
- [ ] **AUTHZ-02**: 现有用户角色数据在权限体系升级后继续保持兼容，不因权限收口导致现有用户无法登录
- [ ] **AUTHZ-03**: 服务端能基于角色矩阵判定用户对查看、处理告警、管理配置和管理用户等操作是否有权限

### User Administration

- [ ] **USER-01**: `admin` 可以查看用户列表及其角色，明确当前谁具备管理权限
- [ ] **USER-02**: `admin` 可以创建用户、修改其他用户资料并分配角色
- [ ] **USER-03**: 非 `admin` 用户不能修改其他用户资料、角色或删除其他用户
- [ ] **USER-04**: 普通用户只能维护允许自助修改的个人信息与密码，不能提升自己的权限
- [ ] **USER-05**: `admin` 可以禁用账号，被禁用用户不能继续登录或使用现有会话访问受保护接口
- [ ] **USER-06**: 系统可以标记用户为“强制改密”，被标记用户在完成密码更新前不能继续以旧凭据正常使用系统

### Protected Operations

- [ ] **PERM-01**: 只有 `admin` 可以新增、修改、删除数据源、渠道、路由规则、静默规则和值班配置
- [ ] **PERM-02**: `operator` 可以查看配置与处理告警，包括确认和快速静默，但不能修改用户角色或系统配置
- [ ] **PERM-03**: `viewer` 只能查看告警、统计和配置结果，不能执行确认、快速静默或任何配置变更
- [ ] **PERM-04**: 所有未授权请求都会在后端返回明确的拒绝结果，而不是静默成功或落入默认行为

### Security Audit

- [ ] **AUDIT-01**: 用户创建、角色变更、账号禁用、账号删除和强制改密等关键安全操作都会记录审计日志
- [ ] **AUDIT-02**: 审计日志至少包含操作者、目标对象、动作类型、关键变更结果和时间信息
- [ ] **AUDIT-03**: 审计日志不能依赖前端上报，必须由后端在关键操作成功或拒绝时可靠落地

### Frontend Permission Awareness

- [ ] **FEACL-01**: 前端菜单、路由和页面入口会根据角色隐藏或禁用无权访问的功能
- [ ] **FEACL-02**: 前端在用户管理与配置管理界面中只展示当前角色允许的按钮和表单操作
- [ ] **FEACL-03**: 当用户访问无权执行的页面或操作时，前端会提供清晰的一致化权限提示

### Verification

- [ ] **VER-01**: 至少有一组后端验证覆盖 `admin`、`operator`、`viewer` 对关键接口的允许/拒绝矩阵
- [ ] **VER-02**: 至少有一组前端验证覆盖不同角色下的菜单、页面入口和关键操作显隐
- [ ] **VER-03**: 角色与权限的使用说明、默认行为和限制会同步更新到项目文档或测试文案中
- [ ] **VER-04**: 至少有一组验证覆盖账号禁用、强制改密和审计日志的关键安全路径

## v2 Requirements

### Access Control Follow-up

- **ACL-01**: 支持更细粒度的权限点配置，而不是固定角色集合
- **ACL-02**: 支持用户组、部门或值班班组级别的授权继承
- **ACL-03**: 支持关键管理操作审批流

## Out of Scope

| Feature | Reason |
|---------|--------|
| OAuth / SSO / LDAP 集成 | 本轮聚焦系统内角色权限隔离，不扩展到企业身份源对接 |
| 自定义角色编辑器 | 固定角色足以解决当前越权问题，避免第一轮设计过重 |
| 多租户项目空间权限 | 当前产品仍是单系统维度，先收敛全局角色边界 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTHZ-01 | Phase 5 | Complete |
| AUTHZ-02 | Phase 5 | Complete |
| AUTHZ-03 | Phase 5 | Complete |
| USER-01 | Phase 6 | Pending |
| USER-02 | Phase 6 | Pending |
| USER-03 | Phase 6 | Pending |
| USER-04 | Phase 6 | Pending |
| USER-05 | Phase 6 | Pending |
| USER-06 | Phase 6 | Pending |
| PERM-01 | Phase 7 | Pending |
| PERM-02 | Phase 7 | Pending |
| PERM-03 | Phase 7 | Pending |
| PERM-04 | Phase 7 | Pending |
| AUDIT-01 | Phase 7 | Pending |
| AUDIT-02 | Phase 7 | Pending |
| AUDIT-03 | Phase 7 | Pending |
| FEACL-01 | Phase 8 | Pending |
| FEACL-02 | Phase 8 | Pending |
| FEACL-03 | Phase 8 | Pending |
| VER-01 | Phase 8 | Pending |
| VER-02 | Phase 8 | Pending |
| VER-03 | Phase 8 | Pending |
| VER-04 | Phase 8 | Pending |

**Coverage:**
- v1 requirements: 23 total
- Mapped to phases: 23
- Unmapped: 0

---
*Requirements defined: 2026-04-11*
*Last updated: 2026-04-12 after Phase 5 completion*
