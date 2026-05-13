---
name: phase-19-02
description: Backend recovery audit model, retry/replay service methods, and protected POST API
metadata:
  type: spec
  source_phase: 19-enable-safe-recovery-operations
  source_plan: "02"
  milestone: v1.4
  status: completed
---

# Phase 19 Plan 02: Recovery Backend

## Context & Goals

Plan 19-01 delivered the read-only delivery history surface. This plan delivers the backend controlled recovery capability: structured recovery audit, strict permission boundaries, and two distinct execution paths — `retry` (reuses frozen payload) and `replay` (re-runs with current route/template).

**Goal:** Complete DELV-03, DELV-04, DELV-05 backend truth and API without allowing viewers to inherit write access from the read-only history page.

## Success Criteria

- Maintainers can only execute retry or replay on single failed deliveries
- Every manual recovery leaves a structured audit record including operator, reason, action type, time, result, and original delivery reference
- Retry and replay semantics are separated: retry reuses frozen payload; replay re-evaluates current route/template
- Viewer still only has read access; operator/admin can execute recovery

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| recovery audit model | `internal/models/delivery_recovery.go` | Recovery action source of truth and status enums |
| POST /api/v1/deliveries/:id/retry and /:id/replay handlers | `internal/handlers/delivery.go` | Recovery API endpoints |
| GET/POST separation in router | `internal/router/router.go` | GET: CapabilityViewConfig, POST: CapabilityProcessAlerts |

## Architecture

### Recovery Model

`delivery_recoveries` table fields:
- `original_delivery_id`
- `action`: `retry | replay`
- `reason`: operator-provided audit reason (required, non-empty)
- `actor_user_id`, `actor_username`, `actor_role`
- `status`: `pending | succeeded | failed | rejected`
- `result_delivery_id`
- `error_message`
- `requested_at`, `completed_at`

### Retry vs Replay Semantics

**RetryDelivery:**
- Reuses original delivery's `rendered_payload_snapshot`
- Uses original `channel_id` and channel identity
- Does NOT re-render or re-route
- If original channel deleted/disabled: returns controlled failure, writes recovery record
- New delivery created with `trigger_kind=retry`

**ReplayDelivery:**
- Reloads original delivery's `alert_id` → `models.Alert`
- Re-runs current route/template/channel decision logic
- Uses current configuration, not frozen snapshots
- New delivery created with `trigger_kind=replay`

### Duplicate Protection

- If `original_delivery_id` already has a `pending/in_progress` recovery, second request is rejected
- Recovery records are never overwritten — only new records created

### Authorization

| Endpoint | Protection |
|----------|------------|
| GET /api/v1/deliveries | `CapabilityViewConfig` |
| POST /api/v1/deliveries/:id/retry | `CapabilityProcessAlerts` |
| POST /api/v1/deliveries/:id/replay | `CapabilityProcessAlerts` |

## Implementation Tasks

### Task 1: Establish Recovery Audit Model and Implement Retry/Replay Executors in Delivery Service

**Files:** `internal/models/delivery_recovery.go`, `internal/models/delivery_recovery_test.go`, `internal/database/postgres.go`, `internal/delivery/service.go`, `internal/delivery/service_test.go`, `internal/handlers/webhook.go`

**Acceptance Criteria:**
- Recovery model includes at minimum: `original_delivery_id`, `action`, `reason`, `actor_user_id`, `actor_username`, `actor_role`, `status`, `result_delivery_id`, `error_message`, `requested_at`, `completed_at`; registered in GORM migrations
- `delivery.Service` exposes separate `RetryDelivery(...)` and `ReplayDelivery(...)` methods; they do NOT share the same path just with different action
- `RetryDelivery` only reuses original delivery's `rendered_payload_snapshot`, original `channel_id`, channel identity; `ReplayDelivery` must reload from `alerts` table and re-evaluate current route/template/channel
- Same `original_delivery_id` with existing `pending/in_progress` recovery is rejected
- New resulting delivery uses `TriggerKindRetry` or `TriggerKindReplay`, creates independent ledger record (not written back to original delivery attempts)
- `internal/delivery/service_test.go` covers at minimum four behaviors: failed delivery + retry; failed delivery + replay; delivered/pending delivery rejection; duplicate in-progress rejection

**Action:** Create independent recovery source of truth model (do not stuff manual recovery metadata into `AuditLog.Detail` text or original delivery attempt). Implement service layer recovery entry points that explicitly receive actor info and reason text — reason is required and limited to auditable strings.

`retry` path must follow Phase 18's bounded 3-attempt send, attempt ledger, and terminal failure logic, but title/content come from original delivery's `rendered_payload_snapshot` — no re-render or re-route. If original channel deleted/disabled, return controlled failure and write recovery record.

`replay` path must look up original delivery's `alert_id` → `models.Alert`, reuse existing route matching and template rendering logic from `WebhookHandler`, send via current configuration, then link resulting delivery with recovery record.

Extract shared execution helpers from `internal/handlers/webhook.go` if needed, but do not break existing webhook hot path or Phase 18 log/ledger contracts.

**Verification:** `go test ./internal/models ./internal/delivery -run "Test(NotificationDeliveryRecovery|DeliveryServiceRecovery)" -count=1`

---

### Task 2: Expose Protected Recovery POST API and Verify Read/Write Permission Separation

**Files:** `internal/handlers/delivery.go`, `internal/handlers/delivery_test.go`, `internal/router/router.go`

**Acceptance Criteria:**
- `internal/handlers/delivery.go` adds `POST /api/v1/deliveries/:id/retry` and `POST /api/v1/deliveries/:id/replay` handlers; request body includes at minimum `reason`
- POST handler returns structured result including at minimum: `recovery_id`, `action`, `status`, `original_delivery_id`, `result_delivery_id`, `error_message`; failures continue with `{ "error": "..." }`
- `internal/router/router.go` continues GET history with `CapabilityViewConfig`; recovery POST is separately under `JWTAuth + RequireCapability(authz.CapabilityProcessAlerts)` — does NOT inherit viewer's read-only permission
- `internal/handlers/delivery_test.go` covers 401, viewer 403, operator success, and 400 branches (missing reason, invalid delivery id, non-recoverable state)
- Successful recovery writes `AuditLog` or equivalent structured audit entry; audit detail traceable to `original_delivery_id`, `recovery_id`, `action`, `result_delivery_id`

**Action:** Add request DTOs and response DTOs in `internal/handlers/delivery.go`; `reason` field uses Gin binding as required. Handler reads principal from Gin context and calls service layer `RetryDelivery`/`ReplayDelivery` — no send logic in handler. Update `internal/router/router.go` for read/write separated `/deliveries` route layout: GET history still `view_config`, POST recovery explicitly under `process_alerts`. Add to `internal/handlers/delivery_test.go` authorization and behavior regression proving viewer still only queries history. Make audit writing a contract that lands on both handler success and failure, not only on HTTP 200.

**Verification:** `go test ./internal/handlers ./internal/router -run "Test(DeliveryHandlerRecovery|RouterDeliveriesAuthorization)" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-19-05 | E | recovery POST routes | mitigate | POST routes use `RequireCapability(authz.CapabilityProcessAlerts)` alone; GET history and recovery action capabilities remain separated |
| T-19-06 | R | recovery auditing | mitigate | Every recovery writes structured recovery record plus `AuditLog` entry; covers success, rejection, and execution failure |
| T-19-07 | T | retry vs replay semantics | mitigate | Service layer separates `RetryDelivery` / `ReplayDelivery`; retry only reuses frozen payload, replay must look up `alerts` and live route/template |
| T-19-08 | D | duplicate recovery | mitigate | Uniqueness or transaction protection on `original_delivery_id + pending/in_progress`; second concurrent request rejected |
| T-19-09 | I | recovery snapshots | mitigate | Resulting delivery continues to only save channel identity and rendered payload; no channel secret/config |

## Established Patterns

- **Pattern 1:** Recovery model is independent from original delivery and attempt — manual recovery metadata is NOT stuffed into AuditLog text or original delivery attempt
- **Pattern 2:** Service layer `RetryDelivery` / `ReplayDelivery` are explicitly separate methods with distinct semantics
- **Pattern 3:** Recovery reason is required, non-empty, and auditable — not optional or freeform beyond operational need

## Decisions

- `trigger_kind=pipeline` for original deliveries, `trigger_kind=retry` for retry recoveries, `trigger_kind=replay` for replay recoveries — each creates independent ledger record
- Recovery records are append-only; status transitions are `pending → succeeded | failed | rejected`
- `result_delivery_id` is null until recovery execution creates the resulting delivery

## Deviation Log

No deviations from plan recorded.
