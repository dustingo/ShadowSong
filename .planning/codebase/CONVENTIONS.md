# Coding Conventions

**Analysis Date:** 2026-04-09

## Naming Patterns

**Files:**
- Go backend files use lower-case package-oriented names such as `internal/handlers/alert.go`, `internal/database/postgres.go`, and `internal/middleware/auth.go`.
- React pages and components use PascalCase filenames such as `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Login.tsx`, and `frontend/src/components/SeverityBadge.tsx`.
- Frontend stores and API wrappers use camelCase filenames such as `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, and `frontend/src/api/client.ts`.
- Shared export barrels are named `index.ts` in `frontend/src/components/index.ts`, `frontend/src/pages/index.ts`, and `frontend/src/types/index.ts`.

**Functions:**
- Exported Go constructors and methods use PascalCase: `NewAIHandler` in `internal/handlers/ai.go`, `Setup` in `internal/router/router.go`, `Validate` and `BeforeCreate` in `internal/models/alert.go`.
- Unexported Go helpers use camelCase: `getEnvAsInt` and `getEnvAsDuration` in `internal/config/config.go`.
- React component functions use PascalCase when exported: `Alerts` in `frontend/src/pages/Alerts.tsx`, `Login` in `frontend/src/pages/Login.tsx`, `SeverityBadge` in `frontend/src/components/SeverityBadge.tsx`.
- Frontend event handlers and store actions use `handleX`/verb-first camelCase naming: `handleAckConfirm` in `frontend/src/pages/Alerts.tsx`, `fetchAlerts` and `quickSilence` in `frontend/src/stores/alertStore.ts`.

**Variables:**
- Backend local variables favor short descriptive names such as `cfg`, `db`, `req`, `user`, `alert`, and `recs` in `cmd/server/main.go`, `internal/handlers/user.go`, and `internal/handlers/ai.go`.
- Frontend local state uses descriptive camelCase names: `ackModalVisible`, `selectedAlert`, and `silenceDuration` in `frontend/src/pages/Alerts.tsx`.
- Boolean state is named explicitly with `is` or suffixes like `Loading`, `Visible`, and `Connected`, e.g. `loading` in `frontend/src/pages/Login.tsx`, `dataSourcesLoading` in `frontend/src/stores/configStore.ts`, and `wsConnected` in `frontend/src/stores/alertStore.ts`.

**Types:**
- Go struct types are singular domain nouns: `Alert`, `DataSource`, `RouteRule`, `SilenceRule`, `User`, `AILog` in `internal/models/alert.go`, `internal/models/models.go`, and `internal/models/user.go`.
- TypeScript interfaces are also singular domain nouns and mirror backend JSON shapes: `Alert`, `Channel`, `OnDuty`, and `User` in `frontend/src/types/index.ts`.
- Request/response DTOs are explicitly suffixed in backend auth flow: `LoginRequest` and `LoginResponse` in `internal/handlers/user.go`.

## Code Style

**Formatting:**
- Frontend formatting is enforced by Prettier in `frontend/.prettierrc`.
- Use no semicolons, single quotes, 2-space indentation, trailing commas where valid in ES5, `printWidth: 100`, and always include arrow-function parentheses, matching `frontend/.prettierrc`.
- Backend formatting follows standard Go formatting conventions; the code in `cmd/server/main.go` and `internal/handlers/user.go` is `gofmt`-style with tabs and grouped imports.

**Linting:**
- Frontend linting is configured in `frontend/.eslintrc.cjs`.
- Extend `eslint:recommended`, `plugin:@typescript-eslint/recommended`, and `plugin:react-hooks/recommended`.
- Treat `react-refresh/only-export-components` as a warning, `@typescript-eslint/no-explicit-any` as a warning, and allow intentionally unused arguments when prefixed with `_`.
- TypeScript strict mode is enabled in `frontend/tsconfig.json`, but `noUnusedLocals` and `noUnusedParameters` remain disabled. Code should still avoid dead state and unused props even when the compiler allows them.

## Import Organization

**Order:**
1. External libraries first, for example `react`, `antd`, `axios`, `gin`, and `gorm` in `frontend/src/pages/Alerts.tsx`, `frontend/src/api/client.ts`, and `internal/handlers/user.go`.
2. Internal project imports second, using relative frontend paths such as `../stores/alertStore` and module imports such as `github.com/game-ops/ai-alert-system/internal/models`.
3. Local CSS imports last in entry files, as in `frontend/src/main.tsx`.

**Path Aliases:**
- `@/*` is configured in `frontend/tsconfig.json`, but current code primarily uses relative imports such as `../types` and `./pages`.

## State Management

**Frontend:**
- Use Zustand for shared server-backed state in `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, and `frontend/src/stores/userStore.ts`.
- Keep page-specific UI state local with `useState`, such as modal visibility and form draft state in `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Channels.tsx`, and `frontend/src/pages/AIAssistant.tsx`.
- Trigger initial data fetches with `useEffect` in page components, for example `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Dashboard.tsx`, and `frontend/src/pages/OnDuty.tsx`.
- Persist auth state directly in `localStorage` inside the user store and API interceptors, as shown in `frontend/src/stores/userStore.ts` and `frontend/src/api/client.ts`.

**Backend:**
- Keep request state per-handler through struct receivers that hold shared dependencies, such as `UserHandler`, `AIHandler`, and `AlertHandler` in `internal/handlers/user.go`, `internal/handlers/ai.go`, and `internal/handlers/alert.go`.
- Store cross-request identity data in Gin context through middleware keys `user_id`, `username`, and `role` defined in `internal/middleware/auth.go`.
- Put persistent defaults and invariants in GORM hooks instead of controllers when the rule belongs to the model, as in `internal/models/alert.go` and `internal/models/models.go`.

## Error Handling

**Patterns:**
- Backend HTTP handlers return early on error and write JSON with an `error` field using `c.JSON(...)`, as in `internal/handlers/user.go`, `internal/handlers/alert.go`, `internal/handlers/config.go`, and `internal/handlers/ai.go`.
- Handler code commonly distinguishes not-found cases from generic database failures using `gorm.ErrRecordNotFound`, as in `internal/handlers/user.go` and `internal/handlers/alert.go`.
- Infrastructure and service packages return wrapped errors with `fmt.Errorf(... %w ...)`, as in `internal/ai/client.go`, `internal/database/postgres.go`, and `internal/database/redis.go`.
- Startup failures are treated as fatal in `cmd/server/main.go` and `internal/config/config.go`; use `log.Fatalf` or `os.Exit(1)` only for unrecoverable boot-time misconfiguration.
- Frontend shared stores either rethrow request errors after resetting loading state, as in `frontend/src/stores/alertStore.ts` and `frontend/src/stores/configStore.ts`, or log non-critical background refresh failures with `console.error`.
- Frontend pages convert API failures into Ant Design toast feedback with `message.error(...)` and success paths into `message.success(...)`, as in `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Login.tsx`, and `frontend/src/pages/Channels.tsx`.

## Validation

**Backend Validation:**
- Prefer model-owned validation methods for domain rules. `Validate()` exists on `Alert` in `internal/models/alert.go` and on `DataSource`, `Channel`, `RouteRule`, `SilenceRule`, `OnDuty`, and `AILog` in `internal/models/models.go`.
- Use GORM hooks such as `BeforeCreate` and `BeforeUpdate` to apply defaults and enforce validation before persistence, as in `internal/models/alert.go`, `internal/models/models.go`, and `internal/models/user.go`.
- Use Gin binding tags for required request payload fields on request-specific structs, such as `binding:"required"` on `LoginRequest` in `internal/handlers/user.go` and the inline chat input struct in `internal/handlers/ai.go`.
- Some update endpoints intentionally accept partial payloads without strict validation, for example `internal/handlers/config.go` and `internal/handlers/user.go`. Follow the existing partial-update pattern when modifying those endpoints.

**Frontend Validation:**
- Prefer Ant Design `Form.Item` rules for required UI validation, as in `frontend/src/pages/Login.tsx`, `frontend/src/pages/OnDuty.tsx`, `frontend/src/pages/RouteRules.tsx`, and `frontend/src/pages/Silences.tsx`.
- Keep form field names aligned to backend JSON contracts using snake_case names such as `alert_name_pattern`, `channel_id`, and `user_name` in `frontend/src/pages/Silences.tsx` and `frontend/src/pages/OnDuty.tsx`.
- The frontend does not use a shared schema validator such as Zod or Yup. Validation is currently split between AntD forms, ad hoc transformation logic in pages like `frontend/src/pages/DataSources.tsx`, and backend model validation.

## Logging

**Framework:** `log`, `log/slog`, and occasional `fmt.Println`/`fmt.Printf` on the backend; `console.error` in a few frontend stores.

**Patterns:**
- Use `log` or `slog` for process-level boot and infrastructure messages in `cmd/server/main.go`, `internal/database/postgres.go`, `internal/database/redis.go`, and `internal/config/config.go`.
- WebSocket and webhook flows print operational diagnostics directly in `internal/handlers/websocket.go` and `internal/handlers/webhook.go`.
- Frontend avoids structured logging; the only recurring pattern is `console.error(...)` for failed background refreshes in `frontend/src/stores/alertStore.ts`.

## Comments

**When to Comment:**
- Existing comments mainly mark sections, explain business intent, or annotate bilingual behavior, for example in `internal/handlers/ai.go`, `internal/router/router.go`, `internal/models/alert.go`, and `frontend/src/App.tsx`.
- Keep comments short and task-oriented. The repository does not use long explanatory blocks except for large prompt strings such as `systemPrompt` in `internal/handlers/ai.go`.

**JSDoc/TSDoc:**
- Not used in frontend TypeScript files.
- Go code uses line comments above exported types and methods rather than doc blocks, for example in `internal/models/models.go` and `internal/handlers/user.go`.

## Function Design

**Size:** Large handler files are accepted, especially `internal/handlers/config.go`, `internal/handlers/webhook.go`, and `frontend/src/pages/AIAssistant.tsx`. When adding code, follow the local grouping style inside the file instead of introducing a new abstraction pattern in one spot only.

**Parameters:**
- Backend methods prefer `(*Handler).Method(c *gin.Context)` signatures for HTTP work and pass dependencies through constructors like `NewUserHandler` and `NewAIHandler`.
- Frontend store actions and API methods use primitive IDs plus small object payloads, for example `ackAlert(id, comment)` in `frontend/src/stores/alertStore.ts` and `alertApi.ack(id, { comment })` in `frontend/src/api/client.ts`.

**Return Values:**
- Backend validation and service methods return `error` or typed values plus `error`, as in `internal/models/alert.go`, `internal/ai/client.go`, and `internal/notifier/notifier.go`.
- Frontend async actions usually return `Promise<void>` and update store state internally, as in `frontend/src/stores/configStore.ts`.

## Module Design

**Exports:**
- Frontend pages and components favor named exports from the implementation file and re-export through barrel files in `frontend/src/pages/index.ts` and `frontend/src/components/index.ts`.
- API wrappers export grouped objects such as `alertApi`, `channelApi`, `authApi`, and `aiApi` from `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`.
- Backend packages group related behavior by layer: `internal/handlers`, `internal/models`, `internal/database`, `internal/middleware`, and `internal/ai`.

**Barrel Files:**
- Used on the frontend for pages and components.
- Not used in the Go backend.

---

*Convention analysis: 2026-04-09*
