---
phase: 18-establish-delivery-ledger
verified: 2026-04-30T02:25:19Z
status: passed
score: 6/6 must-haves verified
overrides_applied: 0
---

# Phase 18: Establish Delivery Ledger Verification Report

**Phase Goal:** 通知发送链路会把每次投递和最终结果持久化为稳定账本，并保存足够支撑审计与单条 replay 的不可变快照。  
**Verified:** 2026-04-30T02:25:19Z  
**Status:** passed  
**Re-verification:** No - initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | 维护者可以查看任一通知投递的账本记录，包含告警、渠道、发送模式、尝试次数、最终结果和失败原因。 | ✓ VERIFIED | `internal/handlers/delivery.go` exposes list/detail read APIs, `internal/router/router.go:155-159` protects `GET /api/v1/deliveries` and `GET /api/v1/deliveries/:id`, and `TestDeliveryHandlerListFiltersAndPagination` plus `TestDeliveryHandlerGetReturnsDetailShape` assert returned fields. |
| 2 | 超过当前即时重试上限的通知会留下持久化终态记录，而不是只存在于日志里。 | ✓ VERIFIED | `internal/handlers/webhook.go:1115-1146` creates delivery records, appends attempts, and calls `markNotificationFailed(..., true, true)` in the `terminal_failure` branch. `TestWebhookHandlerSendNotification_RetryExhaustPersistsTerminalFailureLedger` asserts `delivery_status=failed`, `attempt_count=3`, and persisted `final_failure_summary`. |
| 3 | 单条投递记录保存的快照足以支撑后续审计和单条 replay，不依赖当时之外的实时配置状态。 | ✓ VERIFIED | `internal/models/notification_delivery.go` defines immutable `AlertSnapshot`, `ChannelSnapshot`, `RouteSnapshot`, `RenderedPayloadSnapshot`, and `FinalFailureSummary`. `internal/delivery/service.go:77-137` freezes those snapshots at `StartDelivery`, and `TestServiceStartDeliveryRecordAttemptAndMarkDelivered` plus `assertSnapshotExcludesSecrets` verify secrets/config are excluded. |
| 4 | 系统对每条实际 `alert x channel` 投递都有 PostgreSQL 真源，不再只靠日志推断。 | ✓ VERIFIED | `internal/database/postgres.go` migrates both ledger tables, `internal/handlers/webhook.go:1115-1129` writes delivery envelopes before sending and appends every real attempt, and webhook tests read persisted ledger rows with `requireSingleDeliveryLedger(...)`. |
| 5 | attempt 历史是 append-only 明细，delivery 主记录只聚合终态，不会覆盖历史尝试。 | ✓ VERIFIED | `internal/models/notification_delivery.go` enforces unique `delivery_id + attempt_number` and rejects `BeforeUpdate`; `internal/delivery/service.go` only `Create(attempt)` for attempt history and `Save(delivery)` for terminal aggregation. `TestNotificationDeliveryAttemptUniqueNumberAndAppendOnly` verifies duplicate attempt numbers and updates fail. |
| 6 | 现有 bounded 3 次发送重试与 `send_attempt` / `terminal_failure` 契约保持不变，同时同步写账本。 | ✓ VERIFIED | `internal/handlers/webhook.go:35` keeps `notificationMaxAttempts = 3`; `internal/handlers/webhook.go:1130` and `1146` keep `send_attempt` and `terminal_failure` logging stages. Webhook retry/fallback tests pass while also asserting ledger writes. |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/models/notification_delivery.go` | 双表模型、状态枚举、快照字段 | ✓ VERIFIED | Substantive model definitions, validation hooks, JSONB snapshots, and append-only attempt guard are present and used by service/tests. |
| `internal/delivery/service.go` | 账本写读真源 | ✓ VERIFIED | Implements `StartDelivery`, `RecordAttempt`, `MarkDelivered`, `MarkFailed`, `GetDeliveryByID`, and `ListDeliveries`; wired from webhook and read API handlers. |
| `internal/database/postgres.go` | 迁移注册 | ✓ VERIFIED | `NotificationDelivery` and `NotificationDeliveryAttempt` are included in the migration table list. |
| `internal/handlers/webhook.go` | 热路径写账本 | ✓ VERIFIED | Delivery service is injected and called across start, per-attempt, success, and terminal failure branches. |
| `internal/handlers/webhook_test.go` | 热路径回归测试 | ✓ VERIFIED | Tests cover immediate success, datasource/render fallback, and retry exhaustion with ledger assertions. |
| `internal/handlers/delivery.go` | 只读列表/详情 API | ✓ VERIFIED | Parses query filters, calls delivery service, and returns snapshots plus attempt history. |
| `internal/router/router.go` | 受保护的 `/deliveries` 路由 | ✓ VERIFIED | Routes are mounted under JWT auth plus `CapabilityViewConfig`. |
| `internal/handlers/delivery_test.go` | 列表/详情/授权回归测试 | ✓ VERIFIED | Covers list filters, detail shape, invalid input, 404, 401, and capability middleware behavior. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `internal/database/postgres.go` | `internal/models/notification_delivery.go` | Migrator registration | ✓ WIRED | `NotificationDelivery` and `NotificationDeliveryAttempt` are registered in the migration table slice. |
| `internal/delivery/service.go` | `notification_deliveries` / `notification_delivery_attempts` | GORM create/update/query | ✓ WIRED | `Create(delivery)` at line 143, `Create(attempt)` at line 172, `Save(delivery)` at lines 205 and 253, and `Model(&models.NotificationDelivery{})` plus `Count/Find` at lines 286-322. |
| `internal/handlers/webhook.go` | `internal/delivery/service.go` | `StartDelivery` / `RecordAttempt` / `MarkDelivered` / `MarkFailed` | ✓ WIRED | Calls occur in `sendChannelNotification` and helper methods at lines 1115-1251. |
| `internal/handlers/webhook.go` | retry/logging contract | `send_attempt` + `terminal_failure` preservation | ✓ WIRED | Existing logging stages and three-attempt loop remain intact while dual-writing to the ledger. |
| `internal/router/router.go` | `internal/handlers/delivery.go` | `GET /api/v1/deliveries*` | ✓ WIRED | `deliveries.GET("", deliveryHandler.List)` and `deliveries.GET("/:id", deliveryHandler.Get)` are mounted under the protected group. |
| `internal/handlers/delivery.go` | `internal/delivery/service.go` | `ListDeliveries` / `GetDeliveryByID` | ✓ WIRED | `h.service.ListDeliveries(...)` at line 87 and `h.service.GetDeliveryByID(...)` at line 112. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `internal/handlers/delivery.go` | `items`, `response` | `delivery.Service.ListDeliveries` / `GetDeliveryByID` | Yes - service queries `models.NotificationDelivery` via GORM with `Count`, `Find`, and `Preload("Attempts")` | ✓ FLOWING |
| `internal/handlers/webhook.go` | `deliveryRecord` and persisted attempts | `delivery.Service.StartDelivery` / `RecordAttempt` / `Mark*` | Yes - hot path persists rows before and after each sender invocation | ✓ FLOWING |
| `internal/delivery/service.go` | snapshot JSON fields | Runtime `Alert`, `Channel`, `RouteRule`, rendered title/body inputs | Yes - values are marshaled into JSONB at creation time and later returned by the read API | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Ledger schema and append-only model contract | `go test ./internal/models -count=1` | `ok` | ✓ PASS |
| Delivery service writes and reads terminal state with attempts | `go test ./internal/delivery -count=1` | `ok` | ✓ PASS |
| Webhook hot path persists success/fallback/terminal-failure ledger state | `go test ./internal/handlers -run 'TestWebhookHandler(.*Delivery.*|.*Fallback.*|.*Retry.*|.*Terminal.*)' -count=1` | `ok` | ✓ PASS |
| Protected delivery read API behaves for list/detail flows | `go test ./internal/handlers ./internal/router -run 'Test(DeliveryHandler|Router.*Deliveries.*)' -count=1` | `ok` | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| `DELV-01` | `18-03-PLAN.md` | 维护者可以查看每次通知投递的持久化记录，包括告警、渠道、发送模式、尝试次数、最终结果和失败原因 | ✓ SATISFIED | `internal/handlers/delivery.go` returns those fields; `internal/router/router.go` exposes protected GET endpoints; `internal/handlers/delivery_test.go` verifies list/detail shape. |
| `DELV-02` | `18-01-PLAN.md`, `18-02-PLAN.md` | 通知在超过当前即时重试上限后，会把最终失败结果持久化保存，而不只停留在日志中 | ✓ SATISFIED | `internal/handlers/webhook.go` marks terminal failure into ledger on retry exhaustion; `TestWebhookHandlerSendNotification_RetryExhaustPersistsTerminalFailureLedger` asserts persisted failure state. |
| `DELV-06` | `18-01-PLAN.md`, `18-02-PLAN.md` | 首版投递记录会保存足够支撑审计和单条 replay 的不可变快照，而不是只依赖当前实时配置 | ✓ SATISFIED | `internal/models/notification_delivery.go` defines immutable snapshot fields; `internal/delivery/service.go` freezes runtime alert/channel/route/rendered payload data; `assertSnapshotExcludesSecrets` confirms sensitive config is excluded. |

No orphaned Phase 18 requirements were found in `.planning/REQUIREMENTS.md`; all traced IDs (`DELV-01`, `DELV-02`, `DELV-06`) are claimed by phase plans and accounted for above.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| `internal/handlers/delivery_test.go` | 190-211 | 403 path is exercised through a stricter synthetic route (`CapabilityManageUsers`) rather than the real `/api/v1/deliveries` role matrix | ℹ️ Info | Middleware behavior is still validated, but the test does not prove any currently-defined role can hit a 403 on the real delivery route. |
| `internal/handlers/webhook.go` | 1120, 1202, 1219, 1251 | Ledger write-failure log branches exist without dedicated regression tests | ℹ️ Info | Hot-path persistence failure handling is present, but failure-path observability is not directly regression-tested. |

### Gaps Summary

No blocking gaps were found. Phase 18 achieves the delivery-ledger goal: the schema exists and migrates, the webhook hot path writes real delivery and attempt records while preserving retry/logging semantics, and operators can read persisted ledger history through protected APIs.

---

_Verified: 2026-04-30T02:25:19Z_  
_Verifier: Claude (gsd-verifier)_
