# Phase 7: Lock Down Protected Operations - Research

**Researched:** 2026-04-12  
**Domain:** Alert action authorization, configuration write protection, and backend audit logging on the existing Go/Gin/GORM stack  
**Confidence:** HIGH

## User Constraints

### Locked Decisions
- Keep `admin` / `operator` / `viewer` role naming and reuse the Phase 5 capability matrix as the single backend policy source.
- `operator` must retain alert-processing ability, including acknowledge and quick silence.
- All configuration write operations must be restricted to `admin`.
- Audit logging for critical account and permission actions is required in this milestone and must be backend-authored, not frontend-reported.

### Deferred Ideas (Out of Scope)
- Fine-grained custom permission editor or dynamic RBAC.
- Full frontend permission-aware menu/button pruning across the whole app; that remains Phase 8.
- External SIEM/log pipeline integration. Phase 7 only needs an internal persistent audit record.

## Requirement Support

| ID | Description | Research Support |
|---|---|---|
| PERM-01 | Only `admin` can create/update/delete datasources, channels, route rules, silence rules, and on-duty config | Current router exposes these routes behind authentication only; capability middleware can be applied at route-group or endpoint level without changing stack |
| PERM-02 | `operator` can view config and process alerts, including acknowledge and quick silence | `CapabilityProcessAlerts` already exists and maps to `operator`; alert action endpoints should be explicitly guarded by it |
| PERM-03 | `viewer` is read-only and cannot ack, quick-silence, or mutate config | Current viewer denial needs router and regression coverage |
| PERM-04 | Unauthorized requests return explicit deny responses | `RequireCapability` already returns `403 insufficient permissions`; Phase 7 should standardize use instead of ad hoc auth gaps |
| AUDIT-01 | Critical account/config/security actions are logged | No audit-log model or handler seam exists yet; Phase 7 must introduce one |
| AUDIT-02 | Audit logs include actor, target, action, result, timestamp | A small persistent GORM model is enough for this phase |
| AUDIT-03 | Audit logs are backend-authored | Best seam is a backend helper/service called from handlers after allow/deny decisions |

## Current Code Reality

- `internal/authz/capabilities.go` already defines `view_alerts`, `process_alerts`, `view_config`, `manage_config`, and `manage_users`.
- `internal/middleware/authorize.go` already provides `RequireCapability`, but `router.go` only uses it for user-management routes.
- `internal/router/router.go` still mounts alert actions and every configuration write route behind `JWTAuth` only.
- `internal/handlers/alert.go` does not record who acknowledged an alert from request principal and creates quick-silence rules with `CreatedBy: "system"`.
- `internal/handlers/config.go` has no audit hook and many write handlers use permissive patterns with minimal post-write checks.
- There is no `AuditLog` model, audit repository, or tests for protected operation allow/deny matrices.

## Recommended Phase Shape

Phase 7 should stay backend-heavy and split into three plans:

1. **Permission surface hardening**
   - Apply capability middleware to alert action routes and config route groups.
   - Preserve read access where `operator` and `viewer` are allowed.
   - Explicitly separate view-vs-manage behavior in router wiring.

2. **Audit seam and handler integration**
   - Add a persistent `AuditLog` model plus a lightweight recorder helper.
   - Record successful and denied critical actions for user/account/config mutations.
   - Fix handler behavior that currently hard-codes `"system"` or ignores the acting principal.

3. **Regression matrix**
   - Expand router-level coverage for `admin` / `operator` / `viewer`.
   - Add focused handler tests for alert-processing permissions and audit persistence.
   - Make unauthorized behavior explicit at HTTP level.

## Architecture Guidance

### Pattern 1: Capability Middleware As The Primary Policy Surface
- Use `RequireCapability(authz.CapabilityProcessAlerts)` for `ack` and `quick-silence`.
- Use `RequireCapability(authz.CapabilityManageConfig)` for config writes and mutations.
- Leave config reads behind authenticated access plus `CapabilityViewConfig` if route-by-route consistency is needed.

### Pattern 2: Small Persistent Audit Log
- Add one GORM model such as:
  - `actor_user_id`
  - `actor_username`
  - `actor_role`
  - `action`
  - `target_type`
  - `target_id`
  - `result`
  - `detail`
  - `created_at`
- Keep the write path simple and local to backend handlers; do not build async infra in this phase.

### Pattern 3: Handler-Level Recorder Helper
- Prefer one helper called from handlers over copy-pasted `db.Create(&AuditLog{})`.
- The helper should tolerate missing principal only on public/system flows and make that explicit in the record.

## Risks And Pitfalls

### Pitfall 1: Route Guards Without Regression Coverage
- If Phase 7 only changes `router.go`, later handler or route edits can silently reopen access.
- Router tests must assert specific `401` / `403` outcomes per role.

### Pitfall 2: Audit Only On Success
- Logging only successful actions misses denied privilege attempts, which are part of the threat picture.
- Phase 7 should at minimum capture both successful and explicitly denied critical account/config operations.

### Pitfall 3: Hard-Coded Actor Identity
- Current quick-silence flows stamp `"system"` even when a real operator performs the action.
- Alert/config handlers must read the principal from middleware when present.

## Required Verification

- Router regression coverage for `admin` / `operator` / `viewer` on:
  - `/api/v1/alerts/:id/ack`
  - `/api/v1/alerts/:id/quick-silence`
  - representative config write endpoints such as datasource/channel/route/silence/onduty mutations
- Handler tests proving:
  - `operator` can ack and quick-silence
  - `viewer` cannot ack or quick-silence
  - config write actions by non-admin are denied
  - audit records are written with actor/target/action/result fields
- End-to-end backend test command should continue to fit the current `go test` workflow.

## Open Questions (Resolved For Planning)

1. **Should Phase 7 audit denied actions as well as successful ones?**  
   Resolution: Yes, for critical account/config operations. This directly supports `AUDIT-03` and strengthens incident review without requiring a second audit phase.

2. **Should alert-action auditing land in Phase 7 or Phase 8?**  
   Resolution: Phase 7 should at least cover backend audit logging for critical alert actions (`ack`, `quick silence`) because they are role-sensitive operations and already part of the protected-operation scope.

## Planning Outcome

- Plan 01 should own route/capability hardening.
- Plan 02 should own the audit model/recorder and handler integration.
- Plan 03 should own the regression matrix and verification surface.

This split keeps code changes coherent and avoids mixing policy wiring, persistence design, and test expansion in one patch.
