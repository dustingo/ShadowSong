---
name: phase-18-03
description: Minimal delivery ledger read-only API with JWT and capability protection
metadata:
  type: spec
  source_phase: 18-establish-delivery-ledger
  source_plan: "03"
  milestone: v1.4
  status: completed
  completed: 2026-04-30
---

# Phase 18 Plan 03: Delivery Read-Only API

## Context & Goals

This plan delivers the minimum read-only surface for the delivery ledger so operators can query delivery history via API rather than digging through backend logs. Single-item details return alert, channel, send mode, attempt count, final result, failure reason, attempt明细, and frozen snapshots.

**Goal:** Satisfy DELV-01 by providing a protected API for querying the ledger. No retry/replay, no full history UI, no health aggregation in this plan.

## Success Criteria

- Maintainers can query the delivery ledger via protected API without directly reading webhook logs
- Single-item detail returns alert, channel, send mode, attempt count, final result, failure reason, attempt明细, and frozen snapshots
- Minimal query surface supports filtering by alert, trace, channel, status, and time range

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| delivery list/detail read-only handler and response DTOs | `internal/handlers/delivery.go` | GET /api/v1/deliveries, GET /api/v1/deliveries/:id |
| JWT and capability-protected /api/v1/deliveries routes | `internal/router/router.go` | Auth protection via CapabilityViewConfig |
| read-only API, filtering, and authorization regression tests | `internal/handlers/delivery_test.go` | Test coverage for 401/403/200, list filtering, detail shape |

## Architecture

### API Endpoints

**GET /api/v1/deliveries** — list with filtering
- Query params: `alert_id`, `trace_id`, `channel_id`, `delivery_status`, `created_from`, `created_to`, `limit`, `offset`
- Response: paginated array + total count

**GET /api/v1/deliveries/:id** — single delivery detail
- Response: full delivery record including `alert_id`, `trace_id`, `channel_id`, `route_rule_id`, `delivery_status`, `delivery_mode`, `attempt_count`, `final_failure_summary`, four snapshot types, attempt明细 array

### Authorization

- Both endpoints protected by JWT auth + `CapabilityViewConfig`
- No new capability introduced
- No POST retry/replay in this plan

### Error Handling

- Binding/parameter errors: `400` with JSON `error` field
- Not found: `404`
- Database errors: `500`

## Implementation Tasks

### Task 1: Implement Minimal Delivery List and Detail Handler

**Files:** `internal/handlers/delivery.go`, `internal/handlers/delivery_test.go`

**Acceptance Criteria:**
- `rg -n "type DeliveryHandler struct|func NewDeliveryHandler|func \\(h \\*DeliveryHandler\\) List|func \\(h \\*DeliveryHandler\\) Get" internal/handlers/delivery.go` hits handler construction and both read-only entry points
- `rg -n "alert_id|trace_id|channel_id|delivery_status|created_from|created_to|limit|offset" internal/handlers/delivery.go` hits minimal filter param parsing
- `go test ./internal/handlers -run "TestDeliveryHandler" -count=1` passes

**Action:** Create `internal/handlers/delivery.go` injecting Plan 01's `delivery.Service`. Implement:
- `List`: reads minimal query params (alert_id, trace_id, channel_id, delivery_status, created_from, created_to, limit, offset), returns paginated array + total
- `Get`: returns single delivery by `:id` including all snapshots and attempt明细

Error style: binding errors `400`, not found `404`, DB errors `500`, all with JSON `error` field. Write tests first using sqlite to construct delivery/attempt data, verify list filtering, detail shape, and not-found branch.

**Verification:** `go test ./internal/handlers -run "TestDeliveryHandler" -count=1`

---

### Task 2: Wire Delivery Read-Only Routes into Existing Auth Group

**Files:** `internal/router/router.go`, `internal/handlers/delivery_test.go`

**Acceptance Criteria:**
- `rg -n "/deliveries|CapabilityViewConfig" internal/router/router.go` hits new route and authorization
- `go test ./internal/router ./internal/handlers -run "Test(DeliveryHandler|Router.*Deliveries.*)" -count=1` passes
- Unauthenticated request returns `401`, authenticated without `view_config` returns `403`, capable user gets `200` on both endpoints

**Action:** In `internal/router/router.go`, initialize `deliveryHandler := handlers.NewDeliveryHandler(...)` and add under existing `/api/v1`: `deliveries := v1.Group("/deliveries")`. Reuse `middleware.JWTAuth(jwtAuth, db)` and `middleware.RequireCapability(authz.CapabilityViewConfig)` for both GET endpoints. No new capability, no unauthenticated access, no POST retry/replay.

Add router-level or handler integration tests proving 401/403/200 authorization outcomes.

**Verification:** `go test ./internal/router ./internal/handlers -run "Test(DeliveryHandler|Router.*Deliveries.*)" -count=1`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-18-07 | I | `/api/v1/deliveries*` | mitigate (blocking) | Both endpoints under `JWTAuth + RequireCapability(CapabilityViewConfig)`; tests must cover 401/403/200 |
| T-18-08 | T | filter parsing | mitigate | Only fixed query params accepted; `limit/offset` have bounds; all queries via service/GORM, no raw SQL拼接 |
| T-18-09 | D | GET /api/v1/deliveries | mitigate | List endpoint limits pagination; no unbounded full-text search or complex aggregation in this phase |

## Established Patterns

- **Pattern 1:** Delivery query page uses URL query as filter truth source; alert deeplink directly targets `alert_id` condition
- **Pattern 2:** Detail evidence drawer uses single-item detail API; list only carries summary, does not guess failure details
- **Pattern 3:** Response only exposes attempts, final_failure_summary, and frozen snapshots — no channel secret/config展开

## Decisions

- No new capability introduced — delivery read API uses existing `CapabilityViewConfig`
- Detail response explicitly does NOT expand any channel secret/config
- List response uses backend-returned `total/limit/offset` pagination semantics, not client-side total guessing

## Deviation Log

None — plan executed as written.
