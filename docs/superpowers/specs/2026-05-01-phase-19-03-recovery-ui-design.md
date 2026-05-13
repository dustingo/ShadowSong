---
name: phase-19-03
description: Frontend recovery actions integrated into delivery history page
metadata:
  type: spec
  source_phase: 19-enable-safe-recovery-operations
  source_plan: "03"
  milestone: v1.4
  status: completed
---

# Phase 19 Plan 03: Recovery UI

## Context & Goals

Plan 19-02 delivered the backend recovery capability (structured audit model, retry/replay service methods, protected POST API). This plan connects that capability to the frontend delivery history page, creating a closed loop where operators can take action and viewers remain read-only.

**Goal:** Complete DELV-03, DELV-04, DELV-05 user-facing portion — making single-item recovery truly usable, not just API-available.

## Success Criteria

- operator/admin can trigger retry or replay on single failed deliveries from the history page
- viewer continues to only read history evidence, cannot see or trigger recovery actions
- After recovery, page shows execution result and resulting delivery association

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| recovery buttons, reason input, result feedback, resulting delivery refresh | `frontend/src/pages/Deliveries.tsx` | Recovery modal, submission, result display |
| POST retry/replay frontend API contract | `frontend/src/api/client.ts` | `deliveryApi.retry`, `deliveryApi.replay` |
| frontend recovery capability check | `frontend/src/authz/capabilities.ts` | `canProcessAlerts` guard |

## Architecture

### Recovery UI Flow

1. Operator views failed delivery in history table
2. Clicks "Retry" or "Replay" button (visible only to `canProcessAlerts(user)` + `delivery_status === 'failed'`)
3. Modal prompts for required `reason` field
4. On submit: button disabled, loading state
5. On success: page refreshes original delivery detail, displays resulting `delivery_id`, `action`, `status`, `error_message`
6. On error: error message displayed, recovery record still queryable

### Permission Model

| Role | See History | See Failure Evidence | See Recovery Buttons | Execute Recovery |
|------|-------------|---------------------|---------------------|-----------------|
| viewer | Yes | Yes | No | No |
| operator | Yes | Yes | Yes (failed only) | Yes |
| admin | Yes | Yes | Yes (failed only) | Yes |

### No Channel Secret Expansion

Recovery UI, like the read-only history, only displays `channel_snapshot` (identity) — never `channel.config` or secrets.

## Implementation Tasks

### Task 1: Extend Frontend Recovery Contract and Capability Check

**Files:** `frontend/src/authz/capabilities.ts`, `frontend/src/types/index.ts`, `frontend/src/api/client.ts`

**Acceptance Criteria:**
- `frontend/src/types/index.ts` adds recovery request/response types with at minimum: `reason`, `recovery_id`, `action`, `status`, `original_delivery_id`, `result_delivery_id`, `error_message`
- `frontend/src/api/client.ts` adds `deliveryApi.retry(id, { reason })` and `deliveryApi.replay(id, { reason })`, using existing axios error handling
- `frontend/src/authz/capabilities.ts` exports reusable recovery capability check, semantically bound to `capabilityProcessAlerts` — not re-using `view_config`

**Action:** Extend frontend recovery types and API without changing existing viewer/operator/admin capability matrix. Recovery action check must reuse `capabilityProcessAlerts`, ensuring viewer never sees retry/replay entry in any UI branch. Keep request payload minimal — only sends `{ reason }`, does not let frontend decide business fields outside action semantics.

**Verification:** `pnpm --dir frontend test -- --run frontend/src/pages/Deliveries.test.tsx`

---

### Task 2: Integrate Retry/Replay Interaction, Result Refresh, and Permission Regression into Deliveries Page

**Files:** `frontend/src/pages/Deliveries.tsx`, `frontend/src/pages/Deliveries.test.tsx`

**Acceptance Criteria:**
- Buttons visible only when `delivery_status === 'failed'` AND user has `process_alerts` capability
- Clicking either action requires `reason` before submit; duplicate clicks disabled during submission; same record cannot be re-triggered before recovery completes
- After successful recovery: page refreshes original delivery detail; resulting `delivery_id`, `action`, `status`, `error_message` displayed in visible location (message/Alert/inline result block)
- `frontend/src/pages/Deliveries.test.tsx` covers at minimum three assertions: viewer cannot see recovery buttons; operator can open reason dialog; successful submission triggers list/detail refresh and result feedback

**Action:** Add recovery operation UI to `frontend/src/pages/Deliveries.tsx` — do NOT put buttons in `Alerts.tsx` to send POST directly. Use Ant Design `Modal` or `Form` to collect reason; on submit call `deliveryApi.retry` or `deliveryApi.replay`. Maintain per-row loading state to prevent double-click duplicate sends.

On successful submission, immediately refresh current detail and list; display resulting `delivery_id`, `action`, `status`, `error_message` via message/Alert/inline result block.

Viewer continues to only read history evidence. operator/admin have action entry.

Add `frontend/src/pages/Deliveries.test.tsx` verifying permission boundary, reason required, and successful refresh behaviors.

**Verification:** `pnpm --dir frontend test -- --run frontend/src/pages/Deliveries.test.tsx && pnpm --dir frontend build`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-19-10 | E | action rendering | mitigate | Buttons only rendered when `canProcessAlerts(user)` AND `delivery_status === 'failed'`; viewer always read-only |
| T-19-11 | R | reason capture | mitigate | Reason is required field submitted to backend and echoed in success/failure result; recovery id / result displayed avoiding "clicked but no record" disputes |
| T-19-12 | D | duplicate clicks | mitigate | Per-row loading state during recovery request; buttons disabled for that record until completion |
| T-19-13 | T | stale UI after recovery | mitigate | Successful submission forces refresh of list and detail; resulting delivery id or error_message displayed, preventing user from continuing based on stale state |

## Established Patterns

- **Pattern 1:** Recovery buttons are in `Deliveries.tsx` detail drawer/modal, not in `Alerts.tsx` action column
- **Pattern 2:** Per-row loading state prevents duplicate recovery submissions
- **Pattern 3:** Successful recovery triggers immediate list/detail refresh with result feedback, not just a success toast

## Decisions

- Recovery reason is always required and echoed back in the result — no anonymous recoveries
- Resulting delivery created by recovery uses same bounded 3-attempt send logic as original pipeline deliveries
- Viewer continues to use `CapabilityViewConfig`; operator/admin use `CapabilityProcessAlerts` for recovery — separation preserved at both frontend and backend

## Deviation Log

No deviations from plan recorded.
