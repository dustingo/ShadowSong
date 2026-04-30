---
phase: 18-establish-delivery-ledger
plan: 02
subsystem: api
tags: [webhook, delivery-ledger, notification, retry, testing]
requires:
  - phase: 15-harden-notification-retry-boundaries
    provides: bounded notification retry semantics and terminal failure logging contract
  - phase: 16-standardize-alert-path-logging
    provides: stable trace_id and send_attempt or terminal_failure field vocabulary
  - phase: 18-establish-delivery-ledger
    provides: delivery service and ledger schema from plan 01
provides:
  - webhook notification hot path writes delivery envelopes, attempts, and terminal outcomes
  - success, default fallback, and retry-exhausted failures persist canonical ledger truth alongside logs
  - webhook regression tests assert ledger status, attempt history, snapshots, and failure summaries
affects: [phase-19-safe-recovery-operations, phase-21-ops-visibility-surfaces, docs-alert-path-runbook]
tech-stack:
  added: []
  patterns: [hot-path ledger dual write, append-only attempt persistence, log-plus-db correlation via trace_id]
key-files:
  created: []
  modified: [internal/handlers/webhook.go, internal/handlers/webhook_test.go]
key-decisions:
  - "在 sendNotification 决定最终 mode 或 payload 后创建 delivery，避免 fallback 后快照与真实发送内容漂移。"
  - "保留 notificationMaxAttempts=3、send_attempt 和 terminal_failure 现有日志契约，账本只做旁路真源补充。"
  - "测试直接复用既有 retry 或 fallback 用例并追加 ledger 断言，避免把回归焦点从日志契约移走。"
patterns-established:
  - "Pattern 1: 每个 alert x channel 在真实发送前先 StartDelivery，再按每次 sender 调用追加 RecordAttempt。"
  - "Pattern 2: 非重试失败立即 MarkFailed，重试耗尽只在 terminal_failure 分支写失败终态。"
requirements-completed: [DELV-02, DELV-06]
duration: 9min
completed: 2026-04-30
---

# Phase 18 Plan 02: Establish Delivery Ledger Summary

**Webhook 通知热路径现在同步落库 delivery 账本，并为成功、default fallback 与 retry exhausted 结果保留可追溯终态**

## Performance

- **Duration:** 9 min
- **Started:** 2026-04-30T10:01:00+08:00
- **Completed:** 2026-04-30T10:09:45+08:00
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- 把 `delivery.Service` 接到 webhook 通知发送链路，在不改变三次 bounded retry 语义的前提下写入 delivery 主记录、attempt 明细和终态。
- 保留 `send_attempt`、`send_notification`、`terminal_failure` 既有日志名字与 `trace_id` 或 `attempt` 字段契约，让日志和数据库能交叉排障。
- 扩充 webhook 回归测试，覆盖成功发送、datasource lookup fallback 到 `default`、以及 retry budget exhausted 后失败终态的账本断言。

## Task Commits

Each task was committed atomically:

1. **Task 1: 把 delivery service 接入 webhook 通知热路径** - `5ad79ad` (feat)
2. **Task 2: 扩充 webhook 回归测试覆盖成功、fallback 和终态失败账本** - `8af8f3e` (test)

## Files Created/Modified

- `internal/handlers/webhook.go` - 注入 `deliveryService`，在发送前创建 delivery，在每次真实 attempt 后记录账本，并在成功或失败终态同步持久化。
- `internal/handlers/webhook_test.go` - 扩展 sqlite 迁移与 handler fixture，给成功、fallback、terminal failure 路径增加 delivery 或 attempt 账本断言。

## Decisions Made

- 在 `sendNotification` 拿到最终 title 或 content 与 mode 后再创建 ledger 快照，确保 fallback 模式下保存的是实际发送内容。
- 发送失败仍先遵循现有 `notifier.IsRetryableSendError` 判定和三次重试预算，再决定是否写终态失败，避免数据库逻辑反向驱动 retry 边界。
- 保持测试入口集中在 `internal/handlers/webhook_test.go`，用现有日志用例承载账本断言，保证 Phase 15/16 验证面的连续性。

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- 无额外阻塞；主要工作是把 route rule 与最终发送快照沿现有热路径带到 ledger service，同时保持旧测试调用方式兼容。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 19 可以直接基于 `notification_deliveries` 和 `notification_delivery_attempts` 继续实现单条 retry 或 replay 操作。
- Phase 21 可以复用同一账本真源做 delivery 历史查询与运维可视化，不需要再从日志反推终态。

## Self-Check: PASSED

---
*Phase: 18-establish-delivery-ledger*
*Completed: 2026-04-30*
