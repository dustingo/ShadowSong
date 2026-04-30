---
phase: 19-enable-safe-recovery-operations
plan: 02
subsystem: api
tags: [gin, gorm, postgres, delivery-ledger, retry, replay, audit]
requires:
  - phase: 18-establish-delivery-ledger
    provides: delivery ledger schema, delivery service, and protected read-only delivery history APIs
provides:
  - structured delivery recovery audit model and migration
  - separate retry and replay execution paths for failed deliveries
  - protected POST recovery APIs with process_alerts capability enforcement
affects: [phase-20-harden-ingress-and-runtime-readiness, phase-21-ops-visibility-surfaces, delivery-api, audit-log]
tech-stack:
  added: []
  patterns: [structured recovery truth model, bounded recovery re-send execution, read-write capability split on shared route prefix]
key-files:
  created: [internal/models/delivery_recovery.go, internal/models/delivery_recovery_test.go]
  modified: [internal/database/postgres.go, internal/delivery/service.go, internal/delivery/service_test.go, internal/handlers/delivery.go, internal/handlers/delivery_test.go, internal/router/router.go]
key-decisions:
  - "人工恢复单独建模为 notification_delivery_recoveries，而不是把元数据塞进原 delivery attempt 或 AuditLog 文本。"
  - "retry 只复用冻结 payload 与原 channel identity；replay 必须回查 live alert、route 和模板。"
  - "GET /deliveries 保留 view_config，POST retry or replay 单独要求 process_alerts，避免 viewer 继承写权限。"
patterns-established:
  - "Pattern 1: Recovery service 先创建 recovery 记录，再执行 bounded send，并在成功、失败、拒绝三种结果下写回终态。"
  - "Pattern 2: Delivery handler 无自定义发送逻辑，只负责 principal 提取、service 调用、结构化响应和 AuditLog 落点。"
requirements-completed: [DELV-03, DELV-04, DELV-05]
duration: 29min
completed: 2026-05-01
---

# Phase 19 Plan 02: Enable Safe Recovery Operations Summary

**单条失败通知现在支持受控 retry 或 replay，保留结构化 recovery 真源、结果关联和受保护的恢复 API**

## Performance

- **Duration:** 29 min
- **Started:** 2026-04-30T23:39:00+08:00
- **Completed:** 2026-05-01T00:07:23+08:00
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- 新增 `notification_delivery_recoveries` 真源模型，记录操作者、原因、动作类型、状态、结果 delivery 和完成时间。
- 在 delivery service 中实现分离的 `RetryDelivery` 和 `ReplayDelivery` 路径，保留三次 bounded send、attempt 账本和 terminal failure 契约。
- 暴露 `POST /api/v1/deliveries/:id/retry` 与 `/:id/replay`，并用 `process_alerts` 与 `view_config` 做读写分离授权。

## Task Commits

Each task was committed atomically:

1. **Task 1: 建立恢复审计模型并在 delivery service 中实现 retry/replay 执行器** - `a3bcf73` (feat)
2. **Task 2: 暴露受保护的 recovery POST API 并验证读写权限分离** - `385507e` (feat)

## Files Created/Modified

- `internal/models/delivery_recovery.go` - 定义 recovery 结构化真源模型、状态枚举和审计字段校验。
- `internal/models/delivery_recovery_test.go` - 覆盖 recovery 默认值、动作合法性和 reason 校验。
- `internal/database/postgres.go` - 将 recovery 表注册到现有 PostgreSQL 迁移列表。
- `internal/delivery/service.go` - 实现 retry/replay 恢复入口、recovery 状态机、live route replay 和 bounded resend 执行。
- `internal/delivery/service_test.go` - 覆盖 retry 冻结 payload、replay live route、非失败拒绝和重复恢复拒绝。
- `internal/handlers/delivery.go` - 新增 retry/replay POST handler、结构化响应和 AuditLog 记录。
- `internal/handlers/delivery_test.go` - 覆盖 401、viewer 403、operator 成功、400 分支和可审计失败。
- `internal/router/router.go` - 将 `/deliveries` 调整为 JWT 共享组下的读写 capability 分流。

## Decisions Made

- 使用独立 recovery 表保存人工恢复链路，保证“恢复前置失败”同样有可查询真源。
- 对重复的 pending/in_progress recovery 直接创建 rejected recovery 记录，而不是静默忽略第二次请求。
- 将 replay 缺少原始 alert、channel disabled 等情况作为受控业务失败返回 400，同时照常写 recovery 和审计记录。

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Recovery 失败状态最初会随事务回滚一起丢失**
- **Found during:** Task 1 (建立恢复审计模型并在 delivery service 中实现 retry/replay 执行器)
- **Issue:** service 在同一事务里先写 recovery 失败状态，再直接返回业务错误，导致 rejected 或 failed recovery 被回滚。
- **Fix:** 调整恢复事务流程，先提交 recovery 终态，再把业务错误带出事务返回调用方。
- **Files modified:** `internal/delivery/service.go`
- **Verification:** `go test ./internal/models ./internal/delivery -run "Test(NotificationDeliveryRecovery|DeliveryServiceRecovery)" -count=1`
- **Committed in:** `a3bcf73`

**2. [Rule 3 - Blocking] Recovery handler 测试最初会尝试真实 webhook 外呼**
- **Found during:** Task 2 (暴露受保护的 recovery POST API 并验证读写权限分离)
- **Issue:** retry 成功用例默认使用真实 `webhook` sender，测试环境会触发外部请求并导致不稳定失败。
- **Fix:** 在 handler 测试中改用本地 `httptest` webhook server，稳定验证成功恢复和审计写入。
- **Files modified:** `internal/handlers/delivery_test.go`
- **Verification:** `go test ./internal/handlers ./internal/router -run "Test(DeliveryHandlerRecovery|RouterDeliveriesAuthorization)" -count=1`
- **Committed in:** `385507e`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** 两项修正都直接服务于恢复链路正确性与测试可执行性，没有扩大计划范围。

## Issues Encountered

- handler 初版把 replay 查不到原始 alert 的业务失败映射成了 `404`；已改为保留 recovery 或审计记录并以 `400` 作为受控失败响应。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 20 可以在现有 recovery 真源上继续加入口防护、readiness 与运行时安全收口，不需要再补恢复账本基础设施。
- Phase 21 可以直接复用 delivery history 和 recovery 关联信息做运维聚合视图、失败证据联查和渠道健康摘要。

## Self-Check: PASSED

---
*Phase: 19-enable-safe-recovery-operations*
*Completed: 2026-05-01*
