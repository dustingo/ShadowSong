---
phase: 18-establish-delivery-ledger
plan: 01
subsystem: database
tags: [postgres, gorm, delivery-ledger, notification]
requires:
  - phase: 15-harden-notification-retry-boundaries
    provides: bounded notification retry semantics and terminal failure contract
  - phase: 16-standardize-alert-path-logging
    provides: stable trace and attempt field vocabulary for ledger alignment
provides:
  - notification delivery and attempt ledger schema
  - immutable audit and retry snapshots for alert x channel sends
  - focused delivery service for ledger writes and preloaded reads
affects: [phase-19-safe-recovery-operations, phase-21-ops-visibility-surfaces, delivery-api]
tech-stack:
  added: []
  patterns: [gorm jsonb snapshots, append-only attempt history, focused delivery service]
key-files:
  created: [internal/models/notification_delivery.go, internal/models/notification_delivery_test.go, internal/delivery/service.go, internal/delivery/service_test.go]
  modified: [internal/database/postgres.go]
key-decisions:
  - "账本主记录只聚合终态与快照，attempt 明细通过独立表 append-only 保存。"
  - "渠道快照只冻结身份字段，不复制 config、secret 或 api_key。"
  - "delivery service 的终态写入采用先读后存，避免模型校验被局部 Updates 绕坏。"
patterns-established:
  - "Pattern 1: notification_deliveries 保存 alert/channel/route/rendered payload 冻结快照。"
  - "Pattern 2: notification_delivery_attempts 使用 delivery_id + attempt_number 唯一键和 BeforeUpdate 禁止改写。"
requirements-completed: [DELV-02, DELV-06]
duration: 8min
completed: 2026-04-30
---

# Phase 18 Plan 01: Establish Delivery Ledger Summary

**PostgreSQL 投递账本双表、不可变快照和聚焦 delivery service，为后续单条 retry/replay 与只读查询建立真源**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-30T01:51:30Z
- **Completed:** 2026-04-30T01:59:42Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- 固定 `notification_deliveries` 与 `notification_delivery_attempts` 双表 schema，并加入现有 PostgreSQL 迁移入口。
- 落地 alert/channel/route/rendered payload/final failure 的冻结快照契约，满足审计与后续恢复语义。
- 提供显式 delivery service，覆盖创建 delivery、追加 attempts、写成功/失败终态，以及预加载 attempts 的详情/列表读取。

## Task Commits

Each task was committed atomically:

1. **Task 1: 建立账本双表模型与迁移注册** - `78d70d2` (feat)
2. **Task 2: 实现聚焦 delivery service 作为账本写读真源** - `ab7f1ad` (feat)

## Files Created/Modified
- `internal/models/notification_delivery.go` - 定义账本主记录、attempt 明细、状态枚举和快照结构。
- `internal/models/notification_delivery_test.go` - 验证默认值、状态校验、attempt 唯一编号和 append-only 约束。
- `internal/delivery/service.go` - 实现账本创建、attempt 追加、终态聚合和预加载读取。
- `internal/delivery/service_test.go` - 覆盖成功链路、重试耗尽失败终态、详情读取和列表预加载。
- `internal/database/postgres.go` - 把 ledger 双表纳入现有 GORM 迁移列表。

## Decisions Made

- 使用独立快照结构类型表达审计边界，避免把渠道敏感配置混入持久化账本。
- 将 `final_failure_summary` 保持为结构化 JSON，便于后续 recovery/API 直接消费。
- 让 service 负责状态机写入，而不是把跨表副作用塞进 GORM hook。

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] 为 Attempts 关系补充显式外键声明**
- **Found during:** Task 1 (建立账本双表模型与迁移注册)
- **Issue:** `NotificationDelivery.Attempts` 缺少明确关联键，GORM 无法迁移或解析模型。
- **Fix:** 为 `Attempts` 增加 `foreignKey:DeliveryID;references:ID`，保证双表关系可迁移且可预加载。
- **Files modified:** `internal/models/notification_delivery.go`
- **Verification:** `go test ./internal/models -count=1`
- **Committed in:** `78d70d2`

**2. [Rule 3 - Blocking] 修正终态聚合写入路径并隔离 sqlite 内存测试库**
- **Found during:** Task 2 (实现聚焦 delivery service 作为账本写读真源)
- **Issue:** 局部 `Updates` 会触发模型校验但缺失完整字段；共享 sqlite memory DSN 让不同测试串库。
- **Fix:** `MarkDelivered/MarkFailed` 改为先读后存；测试数据库改用 `t.Name()` 隔离 DSN。
- **Files modified:** `internal/delivery/service.go`, `internal/delivery/service_test.go`
- **Verification:** `go test ./internal/delivery -count=1` and `go test ./internal/models ./internal/delivery -count=1`
- **Committed in:** `ab7f1ad`

---

**Total deviations:** 2 auto-fixed (2 blocking)
**Impact on plan:** 两项修复都属于当前任务达成所必需的正确性修正，没有扩大 phase 范围。

## Issues Encountered

- 无额外遗留问题。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 发送热路径可以直接接入该 service，把现有重试和终态失败落到 PostgreSQL 真源。
- 后续只读 API 或恢复操作可直接复用 `GetDeliveryByID` / `ListDeliveries` 和结构化失败摘要。

## Self-Check: PASSED

---
*Phase: 18-establish-delivery-ledger*
*Completed: 2026-04-30*
