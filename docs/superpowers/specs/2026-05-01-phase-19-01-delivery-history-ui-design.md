---
name: phase-19-01
description: Delivery history query page with failure evidence drawer and alert deeplink
metadata:
  type: spec
  source_phase: 19-enable-safe-recovery-operations
  source_plan: "01"
  milestone: v1.4
  status: completed
  completed: 2026-05-01
---

# Phase 19 Plan 01: Delivery History UI

## Context & Goals

Phase 18 delivered the delivery ledger backend (dual-table schema, hot path integration, read-only API). Phase 19 Plan 01 delivers the frontend: a protected `/deliveries` history page with query-backed filters, failure evidence drawer, and alert→delivery deeplink.

This plan completes the read-only portion of OPER-01 and OPER-04, providing an operational UI container that Phase 19 subsequent recovery plans can extend with single-item retry/replay actions.

## Success Criteria

- Maintainers can filter notification delivery history by time, alert, channel, and result without reading backend logs
- Maintainers can jump from alert detail to associated delivery history via `alert_id` deeplink
- Maintainers can view failure evidence directly in the history page, including attempts, final failure summary, and frozen payloads — without needing backend logs

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| delivery history query page | `frontend/src/pages/Deliveries.tsx` | Filterable table, pagination, evidence drawer |
| delivery API contract | `frontend/src/api/client.ts` | `GET /api/v1/deliveries` list/get |
| protected `/deliveries` route and menu entry | `frontend/src/App.tsx` | `capabilityViewConfig` protection |
| delivery types | `frontend/src/types/index.ts` | Frontend DTOs for delivery, attempt, final_failure, snapshots |
| alert deeplink | `frontend/src/pages/Alerts.tsx` | "Delivery History" button linking to `/deliveries?alert_id=` |

## Architecture

### Page Structure

```
/deliveries
├── URL query drives initial filter state (alert_id, trace_id, channel_id, delivery_status, created_from, created_to)
├── Ant Design Card + Table
├── Filter form (alert_id, trace_id, channel_id, delivery_status, date range)
├── Pagination (backend total/limit/offset)
└── Drawer/Descriptions for failure evidence
```

### Evidence Drawer Contents

Displays backend-returned fields only:
- `attempts` array
- `final_failure_summary`
- `rendered_payload_snapshot`
- `channel_snapshot`
- `route_snapshot`

No channel secret/config展开.

### Authorization

- `/deliveries` route protected by `capabilityViewConfig`
- No recovery action buttons in this plan — only read evidence
- viewer/operator/admin all can view, none can write (yet)

## Implementation Tasks

### Task 1: Define Frontend Delivery Read Contract and Wire Protected Route and Deeplink

**Files:** `frontend/src/types/index.ts`, `frontend/src/api/client.ts`, `frontend/src/pages/index.ts`, `frontend/src/App.tsx`, `frontend/src/pages/Alerts.tsx`

**Acceptance Criteria:**
- `frontend/src/types/index.ts` exports `Delivery`, `DeliveryAttempt`, `DeliveryListResponse`, `DeliveryFilters`, `FinalFailureSummary` types matching backend `GET /api/v1/deliveries` / `GET /api/v1/deliveries/:id` with snake_case field names
- `frontend/src/api/client.ts` adds `deliveryApi.list(params)` and `deliveryApi.get(id)` — params only include backend-supported: `alert_id`, `trace_id`, `channel_id`, `delivery_status`, `created_from`, `created_to`, `limit`, `offset`
- `frontend/src/App.tsx` has protected `/deliveries` route via `capabilityViewConfig`, adds "Delivery History" to menu for `view_config` capability users
- `frontend/src/pages/Alerts.tsx` adds visible "Delivery History" entry per alert, linking to `/deliveries?alert_id=<record.alert_id>` — does not remove existing confirm/silence actions

**Action:** Define frontend types matching backend response DTOs. Add `deliveryApi` using existing axios interceptor pattern and `getApiErrorMessage`. Add `/deliveries` route and menu item in `App.tsx`, continuing to use `capabilityViewConfig`. In `Alerts.tsx`, add "Delivery History" deeplink in the action column or expanded content, using `alert_id` query param only (not requiring user to enter trace_id).

**Verification:** `pnpm --dir frontend test -- --run frontend/src/pages/Alerts.test.tsx`

---

### Task 2: Implement Filterable Delivery History Page and Failure Evidence Detail

**Files:** `frontend/src/pages/Deliveries.tsx`, `frontend/src/pages/Deliveries.test.tsx`, `frontend/src/pages/index.ts`

**Acceptance Criteria:**
- Page supports `alert_id`, `trace_id`, `channel_id`, `delivery_status`, `created_from`, `created_to` filters; URL query initializes page state
- Table pagination uses backend `total/limit/offset` semantics, not client-side total guessing
- Detail area directly displays `attempts`, `final_failure_summary`, `rendered_payload_snapshot`, `channel_snapshot`, `route_snapshot`
- `frontend/src/pages/Deliveries.test.tsx` covers at least two assertions: page load with `alert_id` query initial filter; viewer can see history evidence but has no recovery action buttons

**Action:** Create `frontend/src/pages/Deliveries.tsx` as standalone operations page using Ant Design `Card + Table + Drawer/Descriptions`. Filter options strictly match current GET API capabilities — no `result_delivery_id` or recovery record filtering. On first load, parse `location.search` and call `deliveryApi.list`. On row click, call `deliveryApi.get` to show attempts and final failure summary.

Create `frontend/src/pages/Deliveries.test.tsx` following existing Vitest + Testing Library patterns, mocking `deliveryApi` or page dependencies to verify deeplink initial filter and viewer read-only state.

**Verification:** `pnpm --dir frontend test -- --run frontend/src/pages/Alerts.test.tsx frontend/src/pages/Deliveries.test.tsx`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-19-01 | I | evidence display | mitigate | Only display backend-returned snapshot fields; do not expand channel `config`, secret, or any undownloaded sensitive fields |
| T-19-02 | T | query parsing | mitigate | Only construct backend-allowed filter params; time range统一转 RFC3339; pagination respects existing GET contract |
| T-19-03 | E | `/deliveries` route | mitigate | Route continues to be protected by `capabilityViewConfig`; this plan introduces no recovery POST endpoints, preventing viewers from gaining write access via history page |
| T-19-04 | R | alert→delivery deeplink | mitigate | Use explicit `alert_id` query jump and echo current filter conditions in page, allowing operator to verify which alert's delivery history they're viewing |

## Established Patterns

- **Pattern 1:** Delivery query page initializes and echoes filter state via URL search params, passing allowed filter params directly to `deliveryApi.list`
- **Pattern 2:** History detail uses single-item detail API for evidence drawer; list only carries summary, does not guess failure details
- **Pattern 3:** Failure evidence directly displays `attempts`, `final_failure_summary`, `rendered_payload_snapshot`, `channel_snapshot`, `route_snapshot` — no log lookup required

## Decisions

- Delivery history page continues to use `capabilityViewConfig` because this plan only satisfies read-only semantics for OPER-01 / OPER-04 — no recovery action buttons introduced
- Filter state uses URL query as truth source, ensuring alert page deeplink, pagination, and page refresh all maintain consistent state
- Failure evidence shows `attempts`, `final_failure_summary`, `rendered_payload_snapshot`, `channel_snapshot`, `route_snapshot` directly, preventing maintainers from needing to return to backend logs

## Deviation Log

### Auto-fixed Issues

**1. [Rule 1 - Bug] Alerts.tsx test adapter for new Router dependency**
- Found during: Task 1
- Issue: `Alerts.tsx` new `useNavigate` caused existing `Alerts.test.tsx` to crash without Router context
- Fix: Wrap test in `MemoryRouter`, add assertion for deeplink button visibility
- Files: `frontend/src/pages/Alerts.test.tsx`
- Verification: `pnpm --dir frontend test -- --run frontend/src/pages/Alerts.test.tsx`
- Committed in: `a2fd9e1`

**2. [Rule 3 - Blocking] Delivery read-only API response type unwrapping**
- Found during: Task 2
- Issue: Runtime response already unwrapped by axios interceptor to `data`, but TypeScript still typed `deliveryApi` return as `AxiosResponse`, causing page build failure
- Fix: Add `unwrapData` in `frontend/src/api/client.ts` for delivery read-only API minimum type seal
- Files: `frontend/src/api/client.ts`
- Verification: `pnpm --dir frontend build`
- Committed in: `84043d5`
