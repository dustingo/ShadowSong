---
name: phase-07-02
description: Persistent audit log model and handler integration for critical operations
metadata:
  type: spec
  source_phase: 07-lock-down-protected-operations
  source_plan: "02"
  milestone: v1.1
  status: completed
  completed: 2026-04-15
---

# Phase 07 Plan 02: Backend Audit Log Seam

## Context & Goals

Introduce a small but persistent backend audit seam for Phase 7, and wire critical account/config/alert actions into it.

Purpose: let critical security actions be both enforced by the backend and traceable afterward with "who did what to what, with what result".
Output: audit model, unified recording helper, critical handler wiring.

## Success Criteria

- Critical user, permission, config, and alert actions are recorded in the backend as audit logs.
- Audit logs carry at least actor, target, action, result, and timestamp.
- Handlers use a unified seam to write audit, not scattered inline insertions.

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Audit log model | `internal/models/models.go` | Persistent audit log model |
| Config handler audit | `internal/handlers/config.go` | Backend-authored audit entries for critical config operations |
| Alert handler audit | `internal/handlers/alert.go` | Real actor-based audit entries for ack and quick-silence |

## Architecture

### AuditLog Model Fields

- `actor_user_id` - user ID of acting principal
- `actor_username` - username of acting principal
- `actor_role` - role of acting principal
- `action` - action name (e.g., `alert.ack`, `config.datasource.create`)
- `target_type` - type of target entity
- `target_id` - ID of target entity
- `result` - `allowed` or `denied`
- `detail` - additional context
- `created_at` - timestamp

### Key Decisions

- One table, one simple helper/constructor pattern - no queues, external sinks, or schema complexity
- Actor identity from middleware principal (not frontend-supplied data)
- Backend-authored audit: reflects what backend actually decided, not client claims

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| handler allow/deny decision -> audit persistence | Audit must reflect what the backend actually decided, not client claims |
| middleware principal -> audit actor fields | Actor identity must come from validated request context where available |

## Implementation Tasks

### Task 1: Add a persistent audit-log model and migration-safe recorder seam

**Files:** `internal/models/models.go`, `internal/database/postgres.go`

**Acceptance Criteria:**
- `internal/models/models.go` defines `type AuditLog struct`
- `internal/database/postgres.go` migrates `models.AuditLog`
- Field names for actor/target/action/result/timestamp are grep-visible in model definition

**Action:** Add an `AuditLog` GORM model to `internal/models/models.go` with concrete persisted fields. Ensure `internal/database/postgres.go` includes the new model in `AutoMigrate`. Keep the phase scope pragmatic.

**Verification:** `go test ./internal/database ./internal/models -count=1 -timeout 60s`

---

### Task 2: Record critical alert, config, and user-security actions through one backend audit helper

**Files:** `internal/handlers/alert.go`, `internal/handlers/config.go`, `internal/handlers/user.go`, `internal/handlers/user_test.go`

**Acceptance Criteria:**
- `internal/handlers/alert.go` no longer hard-codes `"system"` for user-triggered quick silence actor attribution
- `internal/handlers/config.go` or a shared helper writes `AuditLog` rows for representative create/update/delete actions
- `internal/handlers/user.go` records critical security mutations
- Handler tests assert persisted audit rows with concrete `action`, `result`, and actor fields

**Action:** Create one reusable audit-recording seam inside handlers layer. Log successful and denied critical operations with explicit values: actor identity from middleware principal, target object type/id, action names, and result values. Replace hard-coded actor fallbacks like `CreatedBy: "system"` with actual acting username. Extend handler tests to verify audit rows are persisted with expected fields.

**Verification:** `go test ./internal/handlers ./internal/models ./internal/database -count=1 -timeout 60s`

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-07-04 | R | missing audit trail | mitigate | Persist one backend-owned audit record for critical allow/deny events |
| T-07-05 | T | fake actor identity | mitigate | Use middleware principal fields instead of frontend-supplied actor data |
| T-07-06 | I | silent security changes | mitigate | Record role/disable/force-reset/config mutations with explicit action/result metadata |

## Established Patterns

- **Pattern 1:** One backend-owned audit seam reusable by handlers, not handler-specific duplication
- **Pattern 2:** Actor identity from middleware principal, not client claims
- **Pattern 3:** Audit reflects actual backend decision (allow/deny), not client assumptions

## Decisions

- Critical operations are audited by the backend
- Audit rows carry enough actor/target/result context for incident review
- Audit seam is reusable by later phases

## Deviation Log

None
