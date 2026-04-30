---
phase: 18-establish-delivery-ledger
plan: 03
subsystem: api
tags: [gin, jwt, delivery-ledger, gorm, testing]
requires:
  - phase: 18-establish-delivery-ledger
    provides: delivery ledger schema and delivery service from plan 01
provides:
  - protected read-only delivery list API
  - protected read-only delivery detail API
  - delivery filter and authorization regression coverage
affects: [phase-19-safe-recovery-operations, phase-21-ops-visibility-surfaces, delivery-api]
tech-stack:
  added: []
  patterns: [handler-to-service ledger reads, JWT-plus-capability protection, bounded delivery list filters]
key-files:
  created: [internal/handlers/delivery.go]
  modified: [internal/handlers/delivery_test.go, internal/router/router.go, internal/delivery/service.go, internal/delivery/service_test.go]
key-decisions:
  - "列表 total 仍由 delivery service 返回，避免 handler 绕开账本读真源直接查表。"
  - "delivery API 复用既有 CapabilityViewConfig，而不是新增权限模型。"
  - "详情响应直接回传冻结快照与 attempt 明细，先满足维护者排障，不扩成前端历史页。"
patterns-established:
  - "Pattern 1: 账本只读 handler 负责 query 参数边界校验，再把筛选条件交给 delivery service。"
  - "Pattern 2: 受保护只读运维接口沿用 JWTAuth + RequireCapability 的现有 Gin 组合。"
requirements-completed: [DELV-01]
duration: 9min
completed: 2026-04-30
---

# Phase 18 Plan 03: Establish Delivery Ledger Summary

**受保护的 delivery 列表与详情 API 现在可以返回冻结快照、attempt 历史和最终结果，维护者无需再翻 webhook 日志定位投递证据**

## Performance

- **Duration:** 9 min
- **Started:** 2026-04-30T10:10:00+08:00
- **Completed:** 2026-04-30T10:18:15+08:00
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- 新增 `DeliveryHandler`，提供 `GET /api/v1/deliveries` 与 `GET /api/v1/deliveries/:id` 两个最小只读入口。
- 列表读取支持 `alert_id`、`trace_id`、`channel_id`、`delivery_status`、`created_from`、`created_to`、`limit`、`offset` 的边界校验与筛选。
- 路由接入既有 `JWTAuth + RequireCapability(CapabilityViewConfig)` 体系，并补齐列表、详情和授权回归测试。

## Task Commits

Each task was committed atomically:

1. **Task 1: 实现最小 delivery 列表与详情 handler** - `27ff2ad` (feat)
2. **Task 2: 把 delivery 只读路由接入现有鉴权组** - `700d25b` (feat)

## Files Created/Modified
- `internal/handlers/delivery.go` - 实现 delivery 列表与详情 handler、响应 DTO、筛选参数解析和快照解码。
- `internal/handlers/delivery_test.go` - 覆盖列表筛选、详情结构、400/404，以及 JWT/capability 授权回归。
- `internal/router/router.go` - 挂载受保护的 `/api/v1/deliveries` 与 `/api/v1/deliveries/:id`。
- `internal/delivery/service.go` - 列表读取增加 `total` 统计，保持 handler 通过 service 访问账本真源。
- `internal/delivery/service_test.go` - 更新 service 列表测试以断言总数返回。

## Decisions Made

- 列表接口的 `total` 放在 service 层统计，避免为了分页元数据让 handler 直接依赖底层 GORM 查询。
- 详情接口直接返回四类冻结快照和 attempt 明细，满足 DELV-01 的可见性要求，不在本 plan 追加恢复动作。
- 路由权限沿用 `view_config`，保持与现有运维只读配置面的能力边界一致。

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] 为 delivery 列表补充 total 统计返回**
- **Found during:** Task 1 (实现最小 delivery 列表与详情 handler)
- **Issue:** 既有 `delivery.Service.ListDeliveries` 只返回切片，无法在不绕开 service 的前提下满足分页合同中的 `total`。
- **Fix:** 扩展 service 列表读取为 `deliveries + total` 返回，并同步更新 service 测试。
- **Files modified:** `internal/delivery/service.go`, `internal/delivery/service_test.go`
- **Verification:** `go test ./internal/handlers ./internal/router -run "Test(DeliveryHandler|Router.*Deliveries.*)" -count=1`
- **Committed in:** `27ff2ad`

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** 该修正仅用于保持 handler 继续通过 delivery service 读取账本，无额外 scope creep。

## Issues Encountered

- 现有 capability 矩阵里 `admin`、`operator`、`viewer` 都具备 `view_config`，所以真实 `/api/v1/deliveries` 路由无法构造“已鉴权但缺少 `view_config`”的 403 场景。测试保留了真实路由的 `401/200` 验证，并用同一 delivery handler + JWT 组合上的更严格 capability 链路证明 403 分支仍成立。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 19 可以直接在该只读 API 真源上继续追加单条 retry 或 replay 操作，而不必重新设计账本详情结构。
- Phase 21 可以复用相同 DTO 与 service 过滤面扩展运维历史页或聚合视图。

## Self-Check: PASSED

---
*Phase: 18-establish-delivery-ledger*
*Completed: 2026-04-30*
