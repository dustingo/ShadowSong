---
phase: 05-normalize-role-model
verified: 2026-04-11T17:11:27Z
status: passed
score: 9/9 must-haves verified
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 8/9
  gaps_closed:
    - "Phase completion is blocked until the persisted-role audit runs against the target database and finds no unsupported `users.role` values."
  gaps_remaining: []
  regressions: []
---

# Phase 5: Normalize Role Model Verification Report

**Phase Goal:** 在不破坏现有登录链路的前提下，统一 `admin`、`operator`、`viewer` 角色语义，并建立后端可复用的权限判定基线。
**Verified:** 2026-04-11T17:11:27Z
**Status:** passed
**Re-verification:** Yes — after gap closure

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | 用户模型、校验逻辑和鉴权上下文只接受 `admin`、`operator`、`viewer` 三种受支持角色。 | ✓ VERIFIED | `internal/authz/roles.go`, `internal/models/user.go`, `internal/auth/jwt.go`, and `internal/middleware/auth.go` still centralize and enforce the three-role contract. |
| 2 | Unsupported role strings cannot be persisted through model hooks or user-management handlers. | ✓ VERIFIED | `internal/models/user.go` still rejects invalid roles in validation/hooks, and `internal/handlers/user.go` still rejects invalid create/update payloads before persistence. |
| 3 | 现有账号在权限体系升级后继续可登录可鉴权，不需要承担角色改名迁移成本。 | ✓ VERIFIED | `internal/auth/jwt.go` keeps `user_id` / `username` / `role` claim keys, login/refresh tests pass, and bootstrap/admin role naming remains unchanged. |
| 4 | Middleware rejects unsupported role claims instead of copying arbitrary strings into request context. | ✓ VERIFIED | `ValidateToken()` rejects unsupported role claims and `JWTAuth()` only sets context after `NewPrincipal()` succeeds. |
| 5 | There is one executable audit path to inspect current persisted `users.role` values before or during rollout. | ✓ VERIFIED | `cmd/roleaudit/main.go` still implements a read-only grouped audit over `users.role`, with command tests passing. |
| 6 | Phase completion is blocked until the persisted-role audit runs against the target database and finds no unsupported `users.role` values. | ✓ VERIFIED | `05-02-SUMMARY.md` now records a target-environment `go run ./cmd/roleaudit` execution on 2026-04-12, including output showing only `admin: 1` and an exit-0 pass message. |
| 7 | The backend can answer whether a principal may view alerts, process alerts, manage configuration, or manage users based on role. | ✓ VERIFIED | `internal/authz/capabilities.go` still provides the centralized role-to-capability matrix for alerts, config, and user-management decisions. |
| 8 | Route and handler code can depend on reusable capability helpers instead of ad hoc role string comparisons. | ✓ VERIFIED | `internal/middleware/authorize.go` and `internal/router/router.go` still expose and use reusable capability guards on admin-only user-management routes. |
| 9 | Later phases can add account disable, forced password reset, and capability-based guards without redesigning the Phase 5 authorization baseline. | ✓ VERIFIED | The principal abstraction in `internal/middleware/auth.go` and centralized policy helpers in `internal/authz/` remain available as extension seams. |

**Score:** 9/9 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/authz/roles.go` | Canonical role constants and helpers | ✓ VERIFIED | Exists, substantive, and reused by models, handlers, auth, and bootstrap code. |
| `internal/authz/roles_test.go` | Direct role contract coverage | ✓ VERIFIED | Regression coverage remains present and passing. |
| `internal/models/user.go` | Hook-backed model enforcement | ✓ VERIFIED | Keeps defaulting plus invalid-role rejection on create/update. |
| `internal/models/user_test.go` | Model regression coverage | ✓ VERIFIED | Covers default role and invalid-role rejection. |
| `internal/handlers/user_test.go` | Request-level role and auth compatibility coverage | ✓ VERIFIED | Covers create/update validation plus login/refresh compatibility. |
| `internal/auth/jwt.go` | Stable JWT claims contract with role validation | ✓ VERIFIED | Keeps claim keys stable and rejects unsupported roles. |
| `internal/middleware/auth.go` | Principal extraction with context compatibility | ✓ VERIFIED | Preserves `user_id`, `username`, and `role` keys via principal helpers. |
| `internal/middleware/auth_test.go` | Middleware regression coverage | ✓ VERIFIED | Covers valid extraction, missing header, and invalid-role rejection. |
| `cmd/roleaudit/main.go` | Read-only persisted-role audit command | ✓ VERIFIED | Uses grouped role query and session read-only setup. |
| `internal/database/postgres.go` | Canonical bootstrap admin role alignment | ✓ VERIFIED | Uses `authz.RoleAdmin` in default admin bootstrap path. |
| `internal/database/postgres_test.go` | Bootstrap admin regression coverage | ✓ VERIFIED | Guards canonical bootstrap role usage. |
| `cmd/roleaudit/main_test.go` | Audit command coverage | ✓ VERIFIED | Covers supported/unsupported datasets, query failure, and read-only behavior. |
| `internal/authz/capabilities.go` | Role-to-capability matrix | ✓ VERIFIED | Centralized capability matrix remains present and substantive. |
| `internal/authz/capabilities_test.go` | Capability matrix regression coverage | ✓ VERIFIED | Matrix tests remain present and passing. |
| `internal/middleware/authorize.go` | Reusable capability-based middleware | ✓ VERIFIED | Distinguishes missing principal from insufficient permissions. |
| `internal/router/router.go` | Capability-guarded baseline adoption | ✓ VERIFIED | User-management routes remain capability-wired. |
| `internal/router/router_test.go` | Router-level authz coverage | ✓ VERIFIED | Router regression coverage remains present and passing. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `internal/handlers/user.go` | `internal/authz/roles.go` | shared role validation before user persistence | ✓ WIRED | Manual source check confirms `DefaultRole` and `IsSupportedRole` calls in create/update paths. |
| `internal/models/user.go` | `internal/authz/roles.go` | model hook defaulting and validation | ✓ WIRED | Hooks and `Validate()` still call `authz.DefaultRole` and `authz.IsSupportedRole`. |
| `internal/auth/jwt.go` | `internal/middleware/auth.go` | `Claims.Role` validation before principal/context creation | ✓ WIRED | `JWTAuth()` depends on validated claims and then constructs a principal before setting context. |
| `internal/database/postgres.go` | `internal/authz/roles.go` | bootstrap default admin role constant | ✓ WIRED | Bootstrap path still assigns `authz.RoleAdmin`. |
| `cmd/roleaudit/main.go` | `users` table | grouped role query for rollout inspection | ✓ WIRED | `roleAuditQuery` still selects grouped `role` counts from `users`. |
| checkpoint rollout verification | `cmd/roleaudit/main.go` | required target-environment audit run before summary approval | ✓ WIRED | `05-02-SUMMARY.md` now records the concrete `go run ./cmd/roleaudit` rollout audit evidence and pass result. |
| `internal/middleware/authorize.go` | `internal/authz/capabilities.go` | capability predicates over the current request principal | ✓ WIRED | `RequireCapability()` calls `authz.Can(principal.Role, capability)`. |
| `internal/router/router.go` | `internal/middleware/authorize.go` | protected route groups and endpoint-level guards | ✓ WIRED | Admin-only user-management routes still use `RequireCapability(authz.CapabilityManageUsers)`. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `cmd/roleaudit/main.go` | `buckets` | `SELECT role, COUNT(*) FROM users GROUP BY role ORDER BY role` | Yes | ✓ FLOWING |
| `internal/middleware/auth.go` | `principal.Role` | JWT claims validated in `internal/auth/jwt.go` | Yes | ✓ FLOWING |
| `internal/router/router.go` | user-management authorization decision | `GetPrincipal()` -> `authz.Can(..., CapabilityManageUsers)` | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Phase 5 regression suites still pass | `go test ./internal/authz ./internal/models ./internal/handlers ./internal/auth ./internal/middleware ./internal/database ./internal/router ./cmd/roleaudit -run "TestRoles|TestUser|TestCreateUser|TestUpdateUser|TestJWT|TestJWTAuth|TestPrincipal|TestLogin|TestRefresh|TestCapability|TestRequireCapability|TestRouter|TestCreateDefaultAdminUser|TestRoleAudit|TestSetSessionReadOnly" -count=1` | All listed packages returned `ok` on re-verification. | ✓ PASS |
| Target persisted-role audit evidence is recorded | Review `05-02-SUMMARY.md` target rollout audit section | Summary now records a 2026-04-12 target-environment run with exit 0 and only supported role buckets. | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `AUTHZ-01` | `05-01-PLAN.md` | 系统使用统一且受约束的内置角色集 `admin`、`operator`、`viewer`，用户不能保存为未定义角色 | ✓ SATISFIED | Canonical roles plus model/handler enforcement remain present and tested. |
| `AUTHZ-02` | `05-02-PLAN.md` | 现有用户角色数据在权限体系升级后继续保持兼容，不因权限收口导致现有用户无法登录 | ✓ SATISFIED | Claim shape remains unchanged, login/refresh tests pass, and the recorded target-environment role audit evidence closes the rollout gate. |
| `AUTHZ-03` | `05-03-PLAN.md` | 服务端能基于角色矩阵判定用户对查看、处理告警、管理配置和管理用户等操作是否有权限 | ✓ SATISFIED | Capability matrix, middleware, and router wiring remain present and tested. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| `internal/handlers/user.go` | 208 | Raw `"admin"` comparison remains in `UpdateUser` self/other-user check | ⚠️ Warning | The reusable baseline is in place, but one brownfield branch still compares a raw role string directly. This does not block Phase 5 goal achievement. |

### Gaps Summary

The previously failing rollout gate is now closed by recorded target-environment audit evidence in `05-02-SUMMARY.md`, and the code/test baseline remains intact on re-verification. Phase 5 now satisfies the roadmap goal and all three requirements without introducing regressions in the login/session contract.

---

_Verified: 2026-04-11T17:11:27Z_
_Verifier: Claude (gsd-verifier)_
