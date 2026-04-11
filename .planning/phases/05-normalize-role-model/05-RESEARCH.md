# Phase 5: Normalize Role Model - Research

**Researched:** 2026-04-11
**Domain:** Backend role normalization, JWT compatibility, and reusable authorization baseline [VERIFIED: codebase grep]
**Confidence:** MEDIUM

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| AUTHZ-01 | 系统使用统一且受约束的内置角色集 `admin`、`operator`、`viewer`，用户不能保存为未定义角色 [VERIFIED: .planning/REQUIREMENTS.md] | Centralize exported role constants plus shared validation in `internal/models/user.go` and enforce them from model hooks and user handlers [VERIFIED: codebase grep] |
| AUTHZ-02 | 现有用户角色数据在权限体系升级后继续保持兼容，不因权限收口导致现有用户无法登录 [VERIFIED: .planning/REQUIREMENTS.md] | Keep DB role values and JWT claim shape unchanged, audit existing `users.role` data, and avoid renaming or remapping supported roles [VERIFIED: codebase grep] |
| AUTHZ-03 | 服务端能基于角色矩阵判定用户对查看、处理告警、管理配置和管理用户等操作是否有权限 [VERIFIED: .planning/REQUIREMENTS.md] | Introduce a reusable authorization helper package and middleware abstraction that can serve Phase 6 user APIs and Phase 7 protected operations [VERIFIED: .planning/ROADMAP.md][VERIFIED: codebase grep] |
</phase_requirements>

## Project Constraints (from AGENTS.md)

- Keep the existing Go + Gin + GORM + PostgreSQL + Redis + React + Vite stack; do not recommend a tech migration [VERIFIED: AGENTS.md]
- Treat the repository as brownfield and do not assume unrelated local modifications can be reverted [VERIFIED: AGENTS.md][VERIFIED: `git status --short`]
- Preserve the current non-AI alerting flows; Phase 5 must not break login or core alert processing continuity [VERIFIED: AGENTS.md]
- Keep the existing role names `admin`, `operator`, `viewer`; do not recommend a rename migration [VERIFIED: AGENTS.md][VERIFIED: .planning/STATE.md][VERIFIED: .planning/PROJECT.md]
- Backend enforcement is the source of truth for access control; frontend hiding is follow-up work in later phases [VERIFIED: .planning/PROJECT.md][VERIFIED: .planning/ROADMAP.md]

## Summary

Phase 5 should be planned as a backend normalization pass, not a feature expansion. The current code already uses the target role names in the `User` model, JWT claims, middleware context, frontend user types, and the bootstrap admin user, so the safest path is to preserve those strings and centralize their semantics instead of migrating data or renaming roles [VERIFIED: internal/models/user.go][VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go][VERIFIED: internal/database/postgres.go][VERIFIED: frontend/src/types/index.ts][VERIFIED: frontend/src/stores/userStore.ts].

The main gap is not naming; it is enforcement. Role validity currently exists as an unexported `validRoles` map in `internal/models/user.go`, but user writes do not consistently force that validation, JWT middleware only copies the role string into Gin context, and authorization is limited to route-level string checks for a few user endpoints [VERIFIED: internal/models/user.go][VERIFIED: internal/handlers/user.go][VERIFIED: internal/middleware/auth.go][VERIFIED: internal/router/router.go]. Planner tasks should therefore focus on one canonical role vocabulary, one reusable authorization surface, and one compatibility-preserving claim/session contract.

**Primary recommendation:** Keep `admin` / `operator` / `viewer` unchanged, add a shared role-and-permission module, enforce supported roles on write and on auth context creation, and keep JWT claim keys plus frontend-stored user shape stable so existing accounts and sessions do not pay a rename cost [VERIFIED: codebase grep].

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go `testing` | stdlib in Go `1.25.0` [VERIFIED: go.mod][VERIFIED: `go version`] | Unit and request tests for role validation, middleware, and handlers [VERIFIED: .planning/codebase/TESTING.md] | Existing backend tests already use standard `testing` patterns, so Phase 5 should extend that path instead of introducing another framework [VERIFIED: internal/models/alert_test.go][VERIFIED: internal/router/router_test.go] |
| Gin | `v1.12.0` [VERIFIED: go.mod] | Route groups and authorization middleware composition [VERIFIED: internal/router/router.go] | The router already mounts `JWTAuth` and `RequireRole(...)` at the group and endpoint level, so reusable authz should integrate there [VERIFIED: internal/router/router.go] |
| GORM | `v1.31.1` [VERIFIED: go.mod] | Persisting users and enforcing model hooks/defaults [VERIFIED: internal/database/postgres.go][VERIFIED: internal/models/user.go] | The codebase already uses model hooks for other domain invariants, which makes GORM hooks the least disruptive place to enforce role validity consistently [VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go] |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` [VERIFIED: go.mod] | Existing token issue/validate/refresh flow [VERIFIED: internal/auth/jwt.go] | Phase 5 can keep the current stateless JWT flow and normalize claim handling without altering the transport contract [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/handlers/user.go] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/stretchr/testify` | `v1.11.1` [VERIFIED: go.mod] | Table-driven assertions in new role and permission tests [VERIFIED: .planning/codebase/TESTING.md] | Use for model, middleware, and router assertion ergonomics because it is already committed in current tests [VERIFIED: internal/models/alert_test.go][VERIFIED: internal/router/router_test.go] |
| `net/http/httptest` | stdlib in Go `1.25.0` [VERIFIED: go.mod][ASSUMED] | Route or middleware request/response checks for allow/deny matrices [ASSUMED] | Use when Phase 5 needs request-level permission regression tests without a full integration harness [ASSUMED] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Shared role constants plus permission helpers in one backend package [VERIFIED: codebase grep] | Keep raw `"admin"` / `"operator"` / `"viewer"` string checks inside handlers and routes [VERIFIED: internal/router/router.go][VERIFIED: internal/handlers/user.go] | That preserves today’s sprawl and makes Phase 6/7 duplicate logic instead of reusing a baseline [VERIFIED: .planning/ROADMAP.md] |
| Stable JWT claim shape with stricter role validation [VERIFIED: internal/auth/jwt.go] | Change claim keys or token payload shape now [VERIFIED: internal/auth/jwt.go] | That increases session breakage risk without helping AUTHZ-01/02 because role names are already correct in the codebase [VERIFIED: internal/auth/jwt.go][VERIFIED: frontend/src/stores/userStore.ts] |

**Installation:**
```bash
go test ./...
```

**Version verification:** Phase 5 should stay on the repo’s committed backend stack rather than introducing new auth libraries. The relevant versions in scope are Gin `v1.12.0`, GORM `v1.31.1`, JWT `v5.3.1`, and Testify `v1.11.1` from `go.mod` [VERIFIED: go.mod].

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── auth/           # JWT issuance and claim validation
├── middleware/     # Gin auth/authz middleware and context helpers
├── models/         # User role constants, validation, and persistence hooks
└── authz/          # New reusable role/capability matrix helpers for Phase 5+
```

### Pattern 1: Canonical Role Vocabulary
**What:** Export one canonical role set from the backend and make every write/auth path depend on it instead of ad hoc string literals [VERIFIED: internal/models/user.go][VERIFIED: internal/handlers/user.go][VERIFIED: internal/router/router.go].  
**When to use:** Immediately for `User` validation, defaulting, JWT claim acceptance, route middleware, and later permission matrices in Phase 6 and 7 [VERIFIED: .planning/ROADMAP.md].  
**Example:**
```go
// Source basis: internal/models/user.go + internal/handlers/user.go [VERIFIED: codebase grep]
package authz

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

var supportedRoles = map[string]struct{}{
	RoleAdmin:    {},
	RoleOperator: {},
	RoleViewer:   {},
}

func IsSupportedRole(role string) bool {
	_, ok := supportedRoles[role]
	return ok
}
```

### Pattern 2: Validate On Write, Preserve On Read
**What:** Enforce supported roles whenever user records are created or updated, while preserving the current DB values and claim field names for already-supported accounts and sessions [VERIFIED: internal/models/user.go][VERIFIED: internal/handlers/user.go][VERIFIED: internal/auth/jwt.go].  
**When to use:** In `CreateUser`, `UpdateUser`, model hooks, and bootstrap admin creation; do not introduce a role rename migration because the accepted names already match the target vocabulary [VERIFIED: internal/database/postgres.go][VERIFIED: .planning/STATE.md].  
**Example:**
```go
// Source basis: internal/models/user.go [VERIFIED: codebase grep]
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = authz.RoleViewer
	}
	if !authz.IsSupportedRole(u.Role) {
		return errors.New("invalid role")
	}
	return nil
}
```

### Pattern 3: Principal-Oriented Authorization
**What:** Convert raw Gin context strings into a small principal abstraction so role checks become reusable and Phase 6 can later add account flags without redesigning every handler [VERIFIED: internal/middleware/auth.go][VERIFIED: .planning/ROADMAP.md].  
**When to use:** In middleware and helper functions that currently depend on `GetUserID` and `GetUserRole` plus string comparisons [VERIFIED: internal/middleware/auth.go][VERIFIED: internal/handlers/user.go].  
**Example:**
```go
// Source basis: internal/middleware/auth.go + Phase 6 roadmap dependency [VERIFIED: codebase grep][VERIFIED: .planning/ROADMAP.md]
type Principal struct {
	UserID   uint
	Username string
	Role     string
}

func (p Principal) CanManageUsers() bool {
	return p.Role == authz.RoleAdmin
}
```

### Anti-Patterns to Avoid
- **Role strings scattered across handlers and routes:** This already exists in `UpdateUser` and `router.Setup`; extending it will make later permission hardening slower and risk inconsistent policy behavior [VERIFIED: internal/handlers/user.go][VERIFIED: internal/router/router.go].
- **Changing JWT claim names during normalization:** The frontend stores both `token` and `user` in `localStorage`, so claim or response shape churn would create avoidable logout/refresh incompatibility [VERIFIED: frontend/src/stores/userStore.ts][VERIFIED: frontend/src/api/auth.ts].
- **Treating `User.Validate()` as sufficient enforcement by itself:** The model defines `Validate()`, but current user create/update flows do not call it explicitly and no `BeforeUpdate` hook exists on `User` today [VERIFIED: internal/models/user.go][VERIFIED: internal/handlers/user.go].

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Role vocabulary | Free-form role strings across handlers [VERIFIED: internal/handlers/user.go][VERIFIED: internal/router/router.go] | Exported constants and `IsSupportedRole` / matrix helpers in one backend package [VERIFIED: codebase grep] | Free-form strings make AUTHZ-01 hard to enforce consistently [VERIFIED: .planning/REQUIREMENTS.md] |
| Permission checks | Per-handler `if role == "admin"` branches only [VERIFIED: internal/handlers/user.go] | Reusable capability helpers or middleware wrappers [VERIFIED: .planning/ROADMAP.md][VERIFIED: internal/middleware/auth.go] | Phase 7 needs matrix-style reuse for config and alert operations, not one-off comparisons [VERIFIED: .planning/ROADMAP.md] |
| Session compatibility | Token or user payload remapping logic for renamed roles [VERIFIED: internal/auth/jwt.go][VERIFIED: frontend/src/stores/userStore.ts] | Keep `role` string stable and harden validation around the same values [VERIFIED: .planning/STATE.md][VERIFIED: .planning/PROJECT.md] | The target role names already match current names, so remapping adds risk without requirement value [VERIFIED: .planning/STATE.md] |

**Key insight:** Phase 5 is a normalization phase, so the hard part is removing duplicated semantics and preserving compatibility, not inventing a more complex authorization system [VERIFIED: .planning/ROADMAP.md][VERIFIED: codebase grep].

## Runtime State Inventory

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | PostgreSQL stores user roles on `models.User.Role` with default `'viewer'`; live row contents were not inspected in this session [VERIFIED: internal/models/user.go][ASSUMED] | Code edit: enforce supported roles on create/update/save. Deployment preflight: run a read-only audit for `users.role NOT IN ('admin','operator','viewer')`; only unsupported live rows would need manual data migration [ASSUMED] |
| Live service config | Active browser sessions cache `token` and serialized `user` in `localStorage`; the backend also refreshes JWTs from existing claims [VERIFIED: frontend/src/stores/userStore.ts][VERIFIED: frontend/src/api/auth.ts][VERIFIED: internal/handlers/user.go][VERIFIED: internal/auth/jwt.go] | Code edit: keep JWT claim keys and user response role field stable. No rename migration is needed if supported role values stay unchanged [VERIFIED: codebase grep] |
| OS-registered state | None found in the repository for role names or authz registrations [VERIFIED: codebase grep] | None |
| Secrets/env vars | `JWT_SECRET` and `TOKEN_EXPIRY` drive auth tokens, but no env var names encode specific roles [VERIFIED: internal/config/config.go] | None for role normalization; keep token signing/expiry behavior unchanged in Phase 5 [VERIFIED: internal/config/config.go] |
| Build artifacts | No compiled artifact or installed package naming depends on role strings in the repo [VERIFIED: codebase grep] | None |

## Common Pitfalls

### Pitfall 1: Enforcing Role Validity Only In Request DTO Logic
**What goes wrong:** A new write path or bootstrap path bypasses handler checks and persists an unsupported role anyway [VERIFIED: internal/handlers/user.go][VERIFIED: internal/database/postgres.go].  
**Why it happens:** `User.Validate()` exists, but current enforcement is not centralized in a save hook or shared validator used everywhere [VERIFIED: internal/models/user.go].  
**How to avoid:** Put role defaulting and validity checks behind exported shared helpers and a GORM save hook, then reuse the same helper in request-level validation for clearer errors [VERIFIED: internal/models/user.go][VERIFIED: internal/models/alert.go].  
**Warning signs:** `CreateUser`, `UpdateUser`, or bootstrap code can set `user.Role` directly without calling the shared validator [VERIFIED: internal/handlers/user.go][VERIFIED: internal/database/postgres.go].

### Pitfall 2: Breaking Existing Sessions While “Normalizing”
**What goes wrong:** Users get forced logout or refresh failures because token claims or frontend-stored user shape changed even though role names did not [VERIFIED: frontend/src/stores/userStore.ts][VERIFIED: frontend/src/api/auth.ts][VERIFIED: internal/auth/jwt.go].  
**Why it happens:** The frontend persists session state locally and the backend refresh flow regenerates tokens from existing claims [VERIFIED: frontend/src/stores/userStore.ts][VERIFIED: internal/auth/jwt.go].  
**How to avoid:** Keep the `role` claim name and role string values stable; only add stricter validation around allowed values [VERIFIED: internal/auth/jwt.go][VERIFIED: .planning/STATE.md].  
**Warning signs:** A plan mentions renaming claims, adding role aliases, or changing frontend user typing before Phase 8 [VERIFIED: frontend/src/types/index.ts][VERIFIED: .planning/ROADMAP.md].

### Pitfall 3: Solving Phase 5 With Endpoint-Specific `RequireRole` Calls Only
**What goes wrong:** Phase 5 “passes” by sprinkling more route guards, but Phase 6/7 still lack a reusable permission baseline [VERIFIED: internal/router/router.go][VERIFIED: .planning/ROADMAP.md].  
**Why it happens:** `RequireRole` currently accepts raw role strings and knows nothing about capability semantics like “can acknowledge alerts” or “can manage users” [VERIFIED: internal/middleware/auth.go].  
**How to avoid:** Add a thin role-capability matrix now and keep `RequireRole` as a convenience wrapper or adapter over that matrix [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: .planning/ROADMAP.md].  
**Warning signs:** Handler code starts adding more `currentUserRole == "admin"` checks instead of calling shared authz helpers [VERIFIED: internal/handlers/user.go].

## Code Examples

Verified patterns from current code and direct extensions of those patterns:

### Centralized Role Validation
```go
// Source basis: internal/models/user.go [VERIFIED: codebase grep]
func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Name == "" {
		return errors.New("name is required")
	}
	if !authz.IsSupportedRole(u.Role) {
		return errors.New("invalid role")
	}
	return nil
}
```

### Reusable Authorization Middleware Adapter
```go
// Source basis: internal/middleware/auth.go + internal/router/router.go [VERIFIED: codebase grep]
func RequireCapability(check func(Principal) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := MustPrincipal(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		if !check(principal) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Raw role string checks embedded in routes and handlers [VERIFIED: internal/router/router.go][VERIFIED: internal/handlers/user.go] | Shared constants plus capability helpers should become the Phase 5 baseline [VERIFIED: .planning/ROADMAP.md][VERIFIED: .planning/REQUIREMENTS.md] | Planned in Phase 5 on 2026-04-11 [VERIFIED: .planning/ROADMAP.md] | Makes Phase 6/7 reuse possible and reduces policy drift [VERIFIED: .planning/ROADMAP.md] |
| `User.Validate()` as a standalone method without save-hook enforcement [VERIFIED: internal/models/user.go] | Save-hook backed enforcement plus handler reuse should be the normalized model [VERIFIED: internal/models/alert.go][VERIFIED: internal/models/models.go] | Planned in Phase 5 on 2026-04-11 [VERIFIED: .planning/ROADMAP.md] | Prevents unsupported roles from being persisted through overlooked paths [VERIFIED: .planning/REQUIREMENTS.md] |
| JWT role claim copied through as a plain string [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go] | Keep the claim stable but convert it into a principal/capability abstraction in middleware [VERIFIED: internal/middleware/auth.go][VERIFIED: .planning/ROADMAP.md] | Planned in Phase 5 on 2026-04-11 [VERIFIED: .planning/ROADMAP.md] | Preserves session compatibility while improving reuse [VERIFIED: frontend/src/stores/userStore.ts] |

**Deprecated/outdated:**
- Direct new role checks such as `currentUserRole == "admin"` in handlers should be considered legacy once the shared authz baseline exists [VERIFIED: internal/handlers/user.go].

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Live `users` rows may contain unsupported role values because database contents were not inspected in this session [ASSUMED] | Runtime State Inventory | Deployment may need a manual cleanup step before strict enforcement |
| A2 | `net/http/httptest` is the preferred request-level test mechanism for new permission tests because no alternative harness is established [ASSUMED] | Standard Stack | Low; planner may instead keep tests at middleware/unit level |

## Open Questions

1. **Are there any live `users.role` values outside `admin` / `operator` / `viewer`?**
   - What we know: The schema and bootstrap user already use only those three names [VERIFIED: internal/models/user.go][VERIFIED: internal/database/postgres.go]
   - What's unclear: The live database contents were not inspected [ASSUMED]
   - Recommendation: Add a deployment preflight query before enabling strict write/read enforcement [ASSUMED]

2. **Should Phase 5 reject unsupported-role tokens immediately, or only reject them after loading the backing user record?**
   - What we know: Current middleware trusts the token claim role and sets it directly in Gin context [VERIFIED: internal/middleware/auth.go][VERIFIED: internal/auth/jwt.go]
   - What's unclear: Whether any live invalid-role tokens exist outside repo-controlled writes [ASSUMED]
   - Recommendation: Default to rejecting unsupported-role claims in middleware and document the behavior as part of the rollout check [ASSUMED]

## Environment Availability

No new external dependency is required for Phase 5 research or planning. The repository already has a working Go toolchain at `go1.25.0`, and current backend tests pass with `go test ./...` [VERIFIED: `go version`][VERIFIED: `go test ./...`].

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/handlers/user.go] | Existing JWT issue/validate/refresh flow should remain the auth transport baseline [VERIFIED: internal/auth/jwt.go] |
| V3 Session Management | yes [VERIFIED: frontend/src/stores/userStore.ts][VERIFIED: frontend/src/api/auth.ts] | Keep token and claim compatibility stable during normalization; do not silently widen accepted role values [VERIFIED: codebase grep] |
| V4 Access Control | yes [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/middleware/auth.go] | Shared backend capability matrix plus middleware/handler adapters [VERIFIED: .planning/ROADMAP.md] |
| V5 Input Validation | yes [VERIFIED: internal/models/user.go][VERIFIED: internal/handlers/user.go] | Centralized role validation on user writes and auth-context creation [VERIFIED: codebase grep] |
| V6 Cryptography | yes [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/models/user.go] | Continue using `jwt/v5` and bcrypt via `golang.org/x/crypto`; do not hand-roll token signing or password hashing [VERIFIED: go.mod][VERIFIED: internal/models/user.go] |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Role tampering through unsupported persisted values [VERIFIED: internal/models/user.go][ASSUMED] | Elevation of Privilege | Enforce supported-role validation in model hooks and user handlers [VERIFIED: codebase grep] |
| Over-broad authorization because routes only require login [VERIFIED: internal/router/router.go][VERIFIED: .planning/PROJECT.md] | Elevation of Privilege | Introduce shared role/capability checks that later phases can apply to config and alert actions [VERIFIED: .planning/ROADMAP.md] |
| Stale session privilege mismatch between token and current backend policy [VERIFIED: internal/auth/jwt.go][VERIFIED: frontend/src/stores/userStore.ts] | Spoofing / Elevation of Privilege | Keep claim shape stable, validate claim role strictly, and centralize principal construction in middleware [VERIFIED: internal/middleware/auth.go] |

## Sources

### Primary (HIGH confidence)
- Repository files via direct read: `.planning/REQUIREMENTS.md`, `.planning/STATE.md`, `.planning/ROADMAP.md`, `.planning/PROJECT.md`, `.planning/codebase/TESTING.md` [VERIFIED: codebase grep]
- Backend auth/user files: `internal/models/user.go`, `internal/auth/jwt.go`, `internal/middleware/auth.go`, `internal/handlers/user.go`, `internal/router/router.go`, `internal/database/postgres.go` [VERIFIED: codebase grep]
- Frontend session/type files: `frontend/src/stores/userStore.ts`, `frontend/src/api/auth.ts`, `frontend/src/api/client.ts`, `frontend/src/types/index.ts` [VERIFIED: codebase grep]
- Repo execution checks: `go version`, `go test ./...`, `git status --short` [VERIFIED: command output]

### Secondary (MEDIUM confidence)
- None

### Tertiary (LOW confidence)
- None beyond assumptions explicitly logged in `## Assumptions Log`

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all stack claims come from `go.mod`, current tests, and committed repo structure [VERIFIED: go.mod][VERIFIED: `.planning/codebase/TESTING.md`][VERIFIED: `go test ./...`]
- Architecture: HIGH - current auth flow, route guards, and session storage were verified directly in code [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go][VERIFIED: internal/router/router.go][VERIFIED: frontend/src/stores/userStore.ts]
- Pitfalls: MEDIUM - the gaps are visible in code, but live DB/session edge cases were not inspected directly [VERIFIED: codebase grep][ASSUMED]

**Research date:** 2026-04-11
**Valid until:** 2026-05-11
