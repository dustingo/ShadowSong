# Phase 6: Secure User Administration - Research

**Researched:** 2026-04-12
**Domain:** User administration boundaries, account disable, and forced password reset on the existing Go/Gin/GORM + React stack [VERIFIED: .planning/ROADMAP.md][VERIFIED: internal/handlers/user.go][VERIFIED: frontend/src/App.tsx]
**Confidence:** HIGH

## User Constraints

### Locked Decisions
- Reuse the Phase 5 principal/capability baseline; do not redesign authz from scratch [VERIFIED: user prompt][VERIFIED: .planning/phases/05-normalize-role-model/05-03-SUMMARY.md]
- Separate admin-managed fields from self-service fields in API design and handler structure [VERIFIED: user prompt]
- Use a minimal but robust account-disable and force-password-reset model under the current stateless JWT flow [VERIFIED: user prompt][VERIFIED: internal/auth/jwt.go]
- Include only the smallest frontend/admin workflow surface needed for Phase 6; do not drift into full Phase 8 permission-aware UX [VERIFIED: user prompt][VERIFIED: .planning/ROADMAP.md]
- Call out compatibility risks and the tests that must exist [VERIFIED: user prompt]

### Claude's Discretion
- Choose the concrete route split, user-state fields, and middleware seam as long as they preserve Phase 5 capability middleware and satisfy `USER-01` through `USER-06` [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/authz/capabilities.go]

### Deferred Ideas (OUT OF SCOPE)
- Full permission-aware menu and page gating across the app belongs to Phase 8 [VERIFIED: .planning/ROADMAP.md]
- Full security audit logging belongs to Phase 7, though Phase 6 should leave clean seams for it [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: .planning/ROADMAP.md]

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| USER-01 | `admin` can view the user list and roles [VERIFIED: .planning/REQUIREMENTS.md] | Keep `CapabilityManageUsers` as the backend gate and include role plus account-status fields in the admin list [VERIFIED: internal/authz/capabilities.go][ASSUMED] |
| USER-02 | `admin` can create users, edit other users, and assign roles [VERIFIED: .planning/REQUIREMENTS.md] | Use admin-only DTOs and routes for role/status changes [VERIFIED: internal/handlers/user.go][ASSUMED] |
| USER-03 | Non-admins cannot edit other users, roles, or delete other users [VERIFIED: .planning/REQUIREMENTS.md] | Replace the mixed `PUT /users/:id` path with explicit admin and self-service paths [VERIFIED: internal/handlers/user.go] |
| USER-04 | Ordinary users can update only safe profile fields and password [VERIFIED: .planning/REQUIREMENTS.md] | Restrict self-service to `name`, `email`, and password-only flows [VERIFIED: internal/handlers/user.go][ASSUMED] |
| USER-05 | Disabled users cannot log in or keep using existing sessions [VERIFIED: .planning/REQUIREMENTS.md] | Add persistent disabled state and re-check current user state in middleware on every protected request [VERIFIED: internal/middleware/auth.go][CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html] |
| USER-06 | Forced-reset users cannot continue normal use until password change completes [VERIFIED: .planning/REQUIREMENTS.md] | Add a force-reset flag plus token freshness cutoff and allow only a narrow password-reset path until completion [VERIFIED: internal/auth/jwt.go][CITED: https://www.rfc-editor.org/rfc/rfc7519][CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html] |
</phase_requirements>

## Project Constraints (from AGENTS.md)

- Keep the existing Go + Gin + GORM + PostgreSQL + Redis + React + Vite stack; no migration [VERIFIED: AGENTS.md]
- Respect brownfield state; the worktree is currently dirty outside this phase [VERIFIED: AGENTS.md][VERIFIED: `git status --short`]
- Do not break the non-AI alerting baseline [VERIFIED: AGENTS.md][VERIFIED: .planning/PROJECT.md]
- Backend enforcement remains the source of truth; frontend hiding is only supporting UX [VERIFIED: .planning/PROJECT.md][VERIFIED: .planning/phases/05-normalize-role-model/05-03-SUMMARY.md]
- No `CLAUDE.md` was present in the repo root during this session [VERIFIED: `if (Test-Path .\\CLAUDE.md) { Get-Content .\\CLAUDE.md }`]

## Summary

Phase 6 should be planned as a boundary-hardening phase on top of Phase 5, not as a new auth design. The backend already has canonical roles, a `Principal`, and `CapabilityManageUsers`, but user administration and self-service still share one mixed `PUT /users/:id` handler, and JWT-authenticated requests still trust token claims without checking current account state in PostgreSQL [VERIFIED: internal/authz/roles.go][VERIFIED: internal/authz/capabilities.go][VERIFIED: internal/middleware/auth.go][VERIFIED: internal/handlers/user.go].

The minimum robust design under the current stateless JWT flow is: split admin-managed and self-service endpoints, add persistent account-control fields to `models.User`, and extend authenticated request processing to load the current user record and reject requests when the account is disabled or the token predates the latest security cutoff [VERIFIED: internal/models/user.go][VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go][CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html][CITED: https://www.rfc-editor.org/rfc/rfc7519].

**Primary recommendation:** Reuse Phase 5 capability middleware, replace the mixed user update flow with explicit admin and self-service APIs, add disable/reset state to `User`, and make JWT-authenticated requests re-check current user state in PostgreSQL on every protected request [VERIFIED: internal/middleware/auth.go][VERIFIED: internal/authz/capabilities.go][VERIFIED: internal/database/postgres.go][CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html].

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | `1.25.0` [VERIFIED: `go version`] | Backend handlers, middleware, and model changes [VERIFIED: cmd/server/main.go][VERIFIED: internal/] | Current auth/user tests already pass on it [VERIFIED: `go test ./internal/auth ./internal/middleware ./internal/handlers ./internal/router ./internal/models -count=1`] |
| Gin | `v1.12.0` [VERIFIED: go.mod] | Route grouping and auth middleware composition [VERIFIED: internal/router/router.go] | Phase 5 already established capability middleware here [VERIFIED: internal/middleware/authorize.go] |
| GORM | `v1.31.1` [VERIFIED: go.mod] | Persist user-state columns and migrate `models.User` [VERIFIED: internal/models/user.go][VERIFIED: internal/database/postgres.go] | The app already migrates tables through GORM at startup [VERIFIED: internal/database/postgres.go] |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` [VERIFIED: go.mod] | Existing JWT claims already include `iat` through `RegisteredClaims` [VERIFIED: internal/auth/jwt.go] | Phase 6 can build freshness checks around existing claims instead of replacing auth [VERIFIED: internal/auth/jwt.go][CITED: https://www.rfc-editor.org/rfc/rfc7519] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| React | repo `^18.2.0`; npm latest `19.2.5` on 2026-04-09 [VERIFIED: frontend/package.json][VERIFIED: `npm view react version time.modified`] | New admin/profile routes [VERIFIED: frontend/src/App.tsx] | Stay on repo version; Phase 6 is not a framework upgrade [VERIFIED: frontend/package.json] |
| React Router DOM | repo `^6.21.1`; npm latest `7.14.0` on 2026-04-10 [VERIFIED: frontend/package.json][VERIFIED: `npm view react-router-dom version time.modified`] | Mount admin and profile pages [VERIFIED: frontend/src/App.tsx] | Reuse current route patterns [VERIFIED: frontend/src/App.tsx] |
| Ant Design | repo `^5.12.8`; npm latest `6.3.5` on 2026-03-30 [VERIFIED: frontend/package.json][VERIFIED: `npm view antd version time.modified`] | Table, modal, form, and feedback components [VERIFIED: frontend/src/pages/Channels.tsx][VERIFIED: frontend/src/pages/OnDuty.tsx] | Reuse the existing CRUD UI style [VERIFIED: frontend/src/pages/Channels.tsx] |
| Zustand | repo `^4.4.7`; npm latest `5.0.12` on 2026-03-16 [VERIFIED: frontend/package.json][VERIFIED: `npm view zustand version time.modified`] | Keep current-user and limited-mode state [VERIFIED: frontend/src/stores/userStore.ts] | Extend the existing store rather than adding a new one [VERIFIED: frontend/src/stores/userStore.ts] |
| Axios | repo `^1.6.5`; npm latest `1.15.0` on 2026-04-08 [VERIFIED: frontend/package.json][VERIFIED: `npm view axios version time.modified`] | Reuse current 401 handling and user APIs [VERIFIED: frontend/src/api/auth.ts][VERIFIED: frontend/src/api/client.ts] | Keep current interceptor behavior [VERIFIED: frontend/src/api/auth.ts][VERIFIED: frontend/src/api/client.ts] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| DB-backed request-time account-state checks [VERIFIED: internal/middleware/auth.go][VERIFIED: internal/auth/jwt.go] | Purely stateless JWT until token expiry [VERIFIED: internal/auth/jwt.go] | Cannot satisfy immediate disable / forced-reset invalidation for existing sessions [VERIFIED: .planning/REQUIREMENTS.md] |
| Explicit admin and self-service endpoints [VERIFIED: internal/router/router.go] | Keep one mixed `PUT /users/:id` handler [VERIFIED: internal/handlers/user.go] | Preserves the current field-boundary ambiguity [VERIFIED: internal/handlers/user.go] |
| PostgreSQL-backed user state [VERIFIED: internal/database/postgres.go] | Redis token blacklist [ASSUMED] | Adds moving parts the current codebase does not otherwise need for Phase 6 [ASSUMED] |

**Installation:**
```bash
go test ./internal/auth ./internal/middleware ./internal/handlers ./internal/router ./internal/models -count=1
cd frontend && pnpm build
```

**Version verification:** Frontend current-version checks were verified with `npm view` on 2026-04-12; backend package versions were verified from `go.mod` [VERIFIED: command output][VERIFIED: go.mod].

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── handlers/user.go       # split admin vs self-service methods and DTOs
├── middleware/auth.go     # JWT parse + DB-backed account-state validation
├── models/user.go         # disable/reset/freshness fields
frontend/src/
├── pages/Users.tsx        # admin-only user admin
├── pages/Profile.tsx      # self-service profile + password
├── api/auth.ts            # split admin/self-service calls
└── stores/userStore.ts    # current user + optional reset-required state
```

### Pattern 1: Split Admin-Managed And Self-Service Contracts
**What:** Use different endpoints and DTOs for admin-managed fields versus self-service fields [VERIFIED: internal/handlers/user.go].  
**When to use:** Immediately, because the current `UpdateUser` path mixes `name`, `email`, `role`, and `password` in one request body [VERIFIED: internal/handlers/user.go].  
**Use:**  
- `GET /api/v1/users` admin-only [VERIFIED: internal/router/router.go]  
- `POST /api/v1/users` admin-only [VERIFIED: internal/router/router.go]  
- `PATCH /api/v1/users/:id` admin-only for role/status/profile edits [ASSUMED]  
- `GET /api/v1/users/me` self-service [VERIFIED: internal/router/router.go]  
- `PATCH /api/v1/users/me/profile` self-service profile only [ASSUMED]  
- `PUT /api/v1/users/me/password` self-service password only [ASSUMED]

### Pattern 2: DB-Backed Request Principal Validation
**What:** Keep JWT signature validation, then load the current `User` row before accepting the request principal [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go].  
**When to use:** On every protected route, because disable and force-reset state must override previously issued tokens [VERIFIED: .planning/REQUIREMENTS.md].  
**Why:** OWASP requires server-side invalidation after high-risk events and privilege changes; current middleware is token-only and cannot do that by itself [CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html][VERIFIED: internal/middleware/auth.go].

### Pattern 3: Per-User Security Cutoff Timestamp
**What:** Store a per-user freshness cutoff such as `token_invalid_before` and compare it against JWT `iat` [CITED: https://www.rfc-editor.org/rfc/rfc7519][VERIFIED: internal/auth/jwt.go].  
**When to use:** On disable, admin-triggered forced reset, role change, and successful password change [ASSUMED].  
**Why:** This is the smallest server-side invalidation model that works with the current stateless JWT architecture [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/database/postgres.go].

### Anti-Patterns to Avoid
- Do not keep `PUT /users/:id` as the primary write path after Phase 6 [VERIFIED: internal/handlers/user.go].
- Do not rely on frontend hiding to protect role or status fields [VERIFIED: frontend/src/App.tsx][VERIFIED: .planning/PROJECT.md].
- Do not treat logout or local-storage clearing as session invalidation; the server must reject stale tokens [VERIFIED: frontend/src/api/auth.ts][VERIFIED: frontend/src/api/client.ts][CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html].

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Authorization baseline | A new role system or ad hoc handler-only role checks [VERIFIED: internal/handlers/user.go] | Phase 5 `Principal` + `CapabilityManageUsers` + capability middleware [VERIFIED: internal/middleware/auth.go][VERIFIED: internal/authz/capabilities.go] | Phase 5 already created the seam this phase depends on [VERIFIED: .planning/phases/05-normalize-role-model/05-03-SUMMARY.md] |
| Session invalidation | Route-by-route custom “is this user valid now?” logic [ASSUMED] | One account-state validation step in auth middleware [VERIFIED: internal/middleware/auth.go][ASSUMED] | Disable and force-reset must apply uniformly [VERIFIED: .planning/REQUIREMENTS.md] |
| Self-service boundary | Filtering forbidden JSON fields after binding the full user model [VERIFIED: internal/handlers/user.go] | Dedicated admin/profile/password DTOs [VERIFIED: internal/handlers/user.go][ASSUMED] | Separate DTOs make privilege boundaries and tests much clearer [ASSUMED] |

**Key insight:** The hard part is not CRUD. It is making existing JWT sessions obey account-state changes immediately [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go][CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html].

## Common Pitfalls

### Pitfall 1: Token Validity Detached From User State
**What goes wrong:** A disabled user or forced-reset user continues using a previously issued JWT until expiry [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go].  
**Why it happens:** Current middleware validates the token and constructs a principal without loading the user record [VERIFIED: internal/middleware/auth.go].  
**How to avoid:** Load the current user and reject disabled or stale-token states before `SetPrincipal` [ASSUMED].  

### Pitfall 2: Self-Service Route Still Accepts Role Or Account-Control Fields
**What goes wrong:** A normal user can still try to send `role`, `disabled`, or `force_password_reset` on a generic update payload [VERIFIED: internal/handlers/user.go].  
**Why it happens:** The existing mixed update DTO includes `role` and `password` in the same body [VERIFIED: internal/handlers/user.go].  
**How to avoid:** Use separate endpoints and DTOs so those fields never bind on self-service routes [ASSUMED].  

### Pitfall 3: Forced-Reset Workflow Has No Allowed Escape Path
**What goes wrong:** The user is marked `force_password_reset`, every protected route fails, and there is no path to change the password [ASSUMED].  
**How to avoid:** Exempt the password-change route, logout, and optionally `GET /users/me` from normal-access blocking [ASSUMED].  

## Rollback / Compatibility Risks

- Existing JWTs currently remain valid until expiry because middleware is token-only; Phase 6 will intentionally change that by rejecting tokens for disabled or stale-security-state users [VERIFIED: internal/middleware/auth.go][ASSUMED].
- Existing frontend session storage only knows `token` and serialized `user`; add new user flags as optional fields first so old localStorage payloads do not crash the UI on first load [VERIFIED: frontend/src/stores/userStore.ts][VERIFIED: frontend/src/types/index.ts][ASSUMED].
- Startup migration currently relies on GORM `AutoMigrate` over `models.User`; new columns should be added as nullable or with safe defaults to avoid rollout breakage on existing rows [VERIFIED: internal/database/postgres.go][ASSUMED].
- The bootstrap admin flow already logs that the initial password must be changed; enforcing that through `force_password_reset` changes first-login behavior and needs explicit release-note coverage [VERIFIED: internal/database/postgres.go][ASSUMED].

## Required Verification

- Backend handler tests for admin update vs self-service update boundaries: self cannot edit another user, self cannot set `role`, `disabled`, or `force_password_reset`, admin can [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/handlers/user_test.go][ASSUMED].
- Middleware tests proving disabled users and stale tokens are rejected even when the JWT signature is otherwise valid [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/middleware/auth_test.go][ASSUMED].
- Login tests proving disabled users cannot log in, forced-reset users enter limited mode, and successful password change clears forced-reset state and invalidates older tokens [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/handlers/user_test.go][ASSUMED].
- Router tests proving only admin can access admin user-management routes, while self-service routes remain available to authenticated non-admin users [VERIFIED: internal/router/router_test.go][ASSUMED].
- Frontend build verification plus targeted UI tests or manual evidence for: admin user list/create/edit/disable/reset controls, self-service profile/password flow, and 401 handling after server-side session invalidation [VERIFIED: frontend/src/App.tsx][VERIFIED: frontend/src/api/auth.ts][ASSUMED].

## Code Examples

### Admin Route Group Reusing Capability Middleware
```go
// Source: internal/router/router.go + internal/authz/capabilities.go [VERIFIED: codebase grep]
users := v1.Group("/users")
users.Use(middleware.JWTAuth(jwtAuth, db))
{
	users.GET("", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.ListUsers)
	users.POST("", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.CreateUser)
	users.PATCH("/:id", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.AdminUpdateUser)

	users.GET("/me", userHandler.GetCurrentUser)
	users.PATCH("/me/profile", userHandler.UpdateOwnProfile)
	users.PUT("/me/password", userHandler.UpdateOwnPassword)
}
```

### Minimal User-State Additions
```go
// Source basis: internal/models/user.go + internal/auth/jwt.go [VERIFIED: codebase grep][ASSUMED]
type User struct {
	// existing fields...
	DisabledAt         *time.Time `json:"disabled_at,omitempty"`
	ForcePasswordReset bool       `json:"force_password_reset"`
	PasswordChangedAt  time.Time  `json:"password_changed_at"`
	TokenInvalidBefore *time.Time `json:"-"`
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| One mixed `PUT /users/:id` path [VERIFIED: internal/handlers/user.go] | Explicit admin and self-service routes [ASSUMED] | Recommended for Phase 6 on 2026-04-12 [VERIFIED: this research] | Removes the main field-boundary ambiguity [ASSUMED] |
| JWT validation based only on token signature/claims [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go] | JWT validation plus DB-backed current-user state check [ASSUMED] | Recommended for Phase 6 on 2026-04-12 [VERIFIED: this research] | Enables immediate disable / forced-reset enforcement [VERIFIED: .planning/REQUIREMENTS.md] |
| Bootstrap admin only logs “must change password” [VERIFIED: internal/database/postgres.go] | Bootstrap admin can be created with `force_password_reset=true` [ASSUMED] | Recommended for Phase 6 on 2026-04-12 [VERIFIED: this research] | Turns advisory text into an actual control [ASSUMED] |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Redis token blacklisting is unnecessary if middleware performs DB-backed account-state validation on every protected request [ASSUMED] | Standard Stack | Could underestimate future revocation/performance needs |
| A2 | `PATCH /users/:id`, `PATCH /users/me/profile`, and `PUT /users/me/password` are the cleanest route shapes here [ASSUMED] | Architecture Patterns | Low; planner can rename routes while preserving the split |
| A3 | Forced-reset users should be allowed only password-change/logout/self-context paths until completion [ASSUMED] | Common Pitfalls | Medium; UX and backend behavior must align |

## Open Questions

1. **Should forced-reset login return normal success with limited mode, or a special response that immediately drives the frontend to reset flow?**
   - What we know: Current login success shape is `{ token, user }`, and the frontend already depends on it [VERIFIED: internal/handlers/user.go][VERIFIED: frontend/src/api/auth.ts][VERIFIED: frontend/src/App.tsx]
   - Recommendation: Keep success shape stable and use `user.force_password_reset` to route into a password-reset-only UI [ASSUMED]

2. **Should Phase 6 block self-disable and self-demotion for admins?**
   - What we know: Self-delete is already blocked [VERIFIED: internal/handlers/user.go]
   - Recommendation: Block self-disable in Phase 6 and consider a “cannot remove the last admin” guard if multiple-admin support is expected [ASSUMED]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend implementation/tests | ✓ [VERIFIED: `go version`] | `go1.25.0` [VERIFIED: `go version`] | — |
| Node.js | Frontend build | ✓ [VERIFIED: `node --version`] | `v22.17.0` [VERIFIED: `node --version`] | — |
| pnpm | Frontend workflow | ✓ [VERIFIED: `pnpm --version`] | `10.28.2` [VERIFIED: `pnpm --version`] | npm could install, but repo expects pnpm [VERIFIED: frontend/package.json] |
| Docker | Local service stack if needed | ✓ [VERIFIED: `docker --version`] | `28.5.1` [VERIFIED: `docker --version`] | SQLite-backed tests already exist for handler coverage [VERIFIED: internal/handlers/user_test.go] |

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes [VERIFIED: internal/handlers/user.go][VERIFIED: internal/auth/jwt.go] | Reject disabled users at login and reject invalid account state during request auth [ASSUMED] |
| V3 Session Management | yes [VERIFIED: internal/auth/jwt.go][VERIFIED: frontend/src/api/auth.ts] | Server-side invalidation after password change and admin security actions through cutoff checks [CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html][CITED: https://www.rfc-editor.org/rfc/rfc7519] |
| V4 Access Control | yes [VERIFIED: .planning/REQUIREMENTS.md][VERIFIED: internal/authz/capabilities.go] | Reuse `CapabilityManageUsers` and keep self-service routes free of admin fields [VERIFIED: internal/authz/capabilities.go][ASSUMED] |
| V5 Input Validation | yes [VERIFIED: internal/models/user.go][VERIFIED: internal/handlers/user.go] | Separate DTOs and model validation for role/account-state invariants [VERIFIED: internal/models/user.go][ASSUMED] |
| V6 Cryptography | yes [VERIFIED: internal/models/user.go][VERIFIED: go.mod] | Keep bcrypt and `jwt/v5`; do not hand-roll [VERIFIED: internal/models/user.go][VERIFIED: internal/auth/jwt.go] |

### Known Threat Patterns for This Stack
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Self-privilege escalation through mixed update payloads [VERIFIED: internal/handlers/user.go] | Elevation of Privilege | Separate admin/self-service routes and DTOs [ASSUMED] |
| Disabled account continues using old JWT [VERIFIED: internal/auth/jwt.go][VERIFIED: internal/middleware/auth.go] | Elevation of Privilege | DB-backed state check plus token freshness cutoff [ASSUMED] |
| Forced-reset user continues normal app usage [VERIFIED: .planning/REQUIREMENTS.md] | Elevation of Privilege | Limited-mode middleware until password change completes [ASSUMED] |

## Sources

### Primary (HIGH confidence)
- `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/PROJECT.md`, `AGENTS.md` [VERIFIED: direct read]
- Phase 5 artifacts: `05-RESEARCH.md`, `05-01-SUMMARY.md`, `05-02-SUMMARY.md`, `05-03-SUMMARY.md` [VERIFIED: direct read]
- Backend files: `internal/models/user.go`, `internal/handlers/user.go`, `internal/auth/jwt.go`, `internal/middleware/auth.go`, `internal/middleware/authorize.go`, `internal/router/router.go`, `internal/database/postgres.go`, `internal/authz/*.go`, and relevant tests [VERIFIED: direct read]
- Frontend files: `frontend/src/api/auth.ts`, `frontend/src/stores/userStore.ts`, `frontend/src/App.tsx`, `frontend/src/types/index.ts` [VERIFIED: direct read]
- Local command checks: `go version`, `node --version`, `pnpm --version`, `docker --version`, `git status --short`, `npm view ...`, and targeted `go test ...` [VERIFIED: command output]
- OWASP Session Management Cheat Sheet [CITED: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html]
- RFC 7519 [CITED: https://www.rfc-editor.org/rfc/rfc7519]

### Secondary (MEDIUM confidence)
- None

### Tertiary (LOW confidence)
- None beyond assumptions explicitly logged above

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - verified from `go.mod`, `frontend/package.json`, `npm view`, and local command checks [VERIFIED: go.mod][VERIFIED: frontend/package.json][VERIFIED: command output]
- Architecture: HIGH - recommendations directly address concrete gaps in handlers, middleware, router, and frontend shell [VERIFIED: internal/handlers/user.go][VERIFIED: internal/middleware/auth.go][VERIFIED: internal/router/router.go][VERIFIED: frontend/src/App.tsx]
- Pitfalls: MEDIUM - risks are strongly indicated by code and requirements, but exact rollout behavior for new columns still needs implementation-time validation [VERIFIED: internal/handlers/user.go][VERIFIED: .planning/REQUIREMENTS.md][ASSUMED]

**Research date:** 2026-04-12
**Valid until:** 2026-05-12
