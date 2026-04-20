# Phase 10: Secure Realtime Alert Access - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the auto-selected discuss outcomes.

**Date:** 2026-04-20
**Phase:** 10-secure-realtime-alert-access
**Areas discussed:** Authentication Boundary, Authorization Scope, Origin Policy, Compatibility And Rollout
**Mode:** auto

---

## Authentication Boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Reuse JWT | Reuse the existing Bearer token and user-state checks from REST auth | ✓ |
| Separate WS Token | Introduce a dedicated realtime token format | |
| Anonymous Snapshot | Allow unauthenticated read-only connection | |

**User's choice:** Auto-selected recommended option: reuse the existing JWT / principal validation path.
**Notes:** This matches current backend security truth in `internal/middleware/auth.go` and avoids introducing a second authentication contract.

---

## Authorization Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Authenticated Baseline | First require authenticated users only and keep scope aligned with existing alert-view semantics | ✓ |
| New Capability Matrix | Introduce a dedicated websocket permission matrix in this phase | |
| Role-Specific Stream Split | Different roles receive different realtime channels now | |

**User's choice:** Auto-selected recommended option: authenticated baseline aligned to current alert-view capability.
**Notes:** The phase requirement only asks for closing the public access surface, not for introducing new product semantics.

---

## Origin Policy

| Option | Description | Selected |
|--------|-------------|----------|
| Config Allowlist | Replace `CheckOrigin = true` with config-driven allowlist and dev-friendly localhost support | ✓ |
| Hardcode Local Only | Keep hardcoded localhost-only rules | |
| Disable Origin Check | Leave origin fully open and rely only on auth | |

**User's choice:** Auto-selected recommended option: config-driven allowlist.
**Notes:** This preserves local development while making production origin policy explicit and maintainable.

---

## Compatibility And Rollout

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal Frontend Change | Keep `/ws` path and existing state shape, only add what is required for authenticated handshake and tests | ✓ |
| Frontend Rewrite | Redesign the whole realtime client flow in this phase | |
| Delay Frontend Changes | Secure backend only and postpone client compatibility work | |

**User's choice:** Auto-selected recommended option: minimal frontend change.
**Notes:** The phase should reduce security risk without expanding into a broader realtime product redesign.

---

## the agent's Discretion

- Exact handshake token transport mechanism
- Exact config field names for origin allowlist

## Deferred Ideas

- Real-time broadcast integration with webhook ingestion
- Fine-grained realtime authorization and content shaping
- Cross-node or durable realtime delivery infrastructure
