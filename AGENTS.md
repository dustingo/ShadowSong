<!-- GSD:project-start source:PROJECT.md -->
## Project

**游戏运维告警系统**

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。系统已经具备后端 API、前端控制台、Webhook 接入、通知路由、静默规则和值班管理能力。v1.0 已完成 AI 能力移除，当前重点转为在非 AI 基线上继续增强模板系统可用性与核心告警链路质量。

**Core Value:** 运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

### Constraints

- **Tech stack**: 维持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite 技术栈 — 本轮不做技术迁移
- **Brownfield**: 必须尊重当前仓库中的既有结构与未提交改动 — 避免误删或回退无关修改
- **Continuity**: AI 移除后核心告警流程仍需可用 — 不能破坏告警接入、展示、路由、静默和值班能力
- **Frontend compatibility**: 前端移除 AI 后路由、菜单、类型和 API 调用要保持自洽 — 不能留下断链入口或运行时错误
- **Documentation**: 项目名称、README、页面标题和测试文案需要反映“非 AI 告警系统”的现状 — 避免品牌和能力描述不一致
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.25.0 - backend application and API server in `go.mod`, `cmd/server/main.go`, `internal/`
- TypeScript 5.3.x - frontend application in `frontend/package.json`, `frontend/src/`, `frontend/tsconfig.json`
- TSX / React JSX - frontend UI composition in `frontend/src/App.tsx` and `frontend/src/pages/`
- SQL via GORM models/migrations - relational persistence wired through `internal/database/postgres.go` and `internal/models/`
- YAML - container orchestration in `docker-compose.yml`
- Make - developer task runner in `Makefile`
## Runtime
- Go runtime 1.25.0 from `go.mod`
- Node.js 18+ expected by project docs in `README.md`
- Browser runtime for the SPA built from `frontend/src/`
- Go modules for backend dependency management in `go.mod`
- pnpm for frontend dependency management in `frontend/package.json`
- Lockfile: present in `go.sum` and `frontend/pnpm-lock.yaml`
## Frameworks
- Gin `v1.12.0` - HTTP server, routing, middleware, WebSocket entrypoints in `go.mod` and `internal/router/router.go`
- GORM `v1.31.1` - ORM, schema migration, model persistence in `go.mod` and `internal/database/postgres.go`
- React `^18.2.0` - frontend SPA in `frontend/package.json` and `frontend/src/App.tsx`
- React Router DOM `^6.21.1` - client-side routing in `frontend/package.json` and `frontend/src/App.tsx`
- PrimeReact `^10.9.7` - component library in `frontend/package.json` and `frontend/src/`
- Zustand `^4.4.7` - frontend client state stores in `frontend/package.json` and `frontend/src/stores/`
- Testify `v1.11.1` - Go assertions/helpers declared in `go.mod`
- Go built-in `go test` runner - invoked by `Makefile`
- No frontend test runner config detected under `frontend/`
- Vite `^5.0.11` - frontend dev server and production bundling in `frontend/package.json` and `frontend/vite.config.ts`
- TypeScript compiler `^5.3.3` - type-check/build step in `frontend/package.json`
- ESLint `^8.56.0` - frontend linting in `frontend/package.json`
- Prettier `^3.1.1` - frontend formatting in `frontend/package.json`
- Docker Compose - local Postgres/Redis provisioning in `docker-compose.yml`
- `godotenv` `v1.5.1` - `.env` loading at startup in `go.mod` and `cmd/server/main.go`
- Make - common dev/build targets in `Makefile`
## Key Dependencies
- `github.com/gin-gonic/gin` `v1.12.0` - API surface and middleware execution in `internal/router/router.go`
- `gorm.io/gorm` `v1.31.1` - model persistence and migrations in `internal/database/postgres.go`
- `gorm.io/driver/postgres` `v1.6.0` - PostgreSQL driver in `internal/database/postgres.go`
- `github.com/redis/go-redis/v9` `v9.18.0` - Redis connectivity and stream publishing in `internal/database/redis.go` and `internal/handlers/webhook.go`
- `github.com/golang-jwt/jwt/v5` - JWT issuance and validation in `internal/auth/jwt.go`
- `axios` `^1.6.5` - frontend HTTP client layer in `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`
- `github.com/joho/godotenv` `v1.5.1` - local env loading in `cmd/server/main.go`
- `github.com/gorilla/websocket` `v1.5.3` - backend WebSocket support declared in `go.mod` and exposed via `internal/router/router.go`
- `@vitejs/plugin-react` `^4.2.1` - Vite React integration in `frontend/package.json` and `frontend/vite.config.ts`
- `echarts` `^5.4.3` and `echarts-for-react` `^3.0.2` - dashboard charting in `frontend/src/pages/Dashboard.tsx`
- `@monaco-editor/react` `^4.6.0` - in-browser editor component in `frontend/src/components/CodeEditor.tsx`
- `react-markdown` `^10.1.0` - markdown rendering dependency still present in `frontend/package.json`, but not part of the current primary UI flow
- `dayjs` `^1.11.19` - date formatting in `frontend/src/components/AlertCard.tsx` and multiple `frontend/src/pages/*.tsx`
## Configuration
- Runtime configuration is loaded from process environment plus optional `.env` via `cmd/server/main.go`
- Central config parsing lives in `internal/config/config.go`
- Required secret: `JWT_SECRET` in `internal/config/config.go`
- Database vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` in `internal/config/config.go`
- Redis vars: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` in `internal/config/config.go`
- Server vars: `SERVER_PORT`, `SERVER_MODE` in `internal/config/config.go`
- Token expiry var: `TOKEN_EXPIRY` in `internal/config/config.go`
- `.env` file is present at project root; contents not inspected
- Backend build/run/test commands are defined in `Makefile`
- Frontend build pipeline is defined in `frontend/package.json`
- Frontend module resolution and aliasing are defined in `frontend/tsconfig.json` and `frontend/vite.config.ts`
- Dev proxying for `/api`, `/webhook`, and `/ws` is defined in `frontend/vite.config.ts`
- Local container runtime config is defined in `docker-compose.yml`
## Platform Requirements
- Go toolchain compatible with `go 1.25.0` in `go.mod`
- Node.js 18+ and pnpm per `README.md`
- Docker and Docker Compose for local Postgres/Redis per `README.md` and `docker-compose.yml`
- PostgreSQL 14 and Redis 7 containers for local services in `docker-compose.yml`
- Production hosting target is not explicitly defined
- Backend expects a reachable PostgreSQL instance, Redis instance, and environment-injected secrets/config from `internal/config/config.go`
- Frontend build output is generated by Vite, but deployment host is not specified in repository config
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Go backend files use lower-case package-oriented names such as `internal/handlers/alert.go`, `internal/database/postgres.go`, and `internal/middleware/auth.go`.
- React pages and components use PascalCase filenames such as `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Login.tsx`, and `frontend/src/components/SeverityBadge.tsx`.
- Frontend stores and API wrappers use camelCase filenames such as `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, and `frontend/src/api/client.ts`.
- Shared export barrels are named `index.ts` in `frontend/src/components/index.ts`, `frontend/src/pages/index.ts`, and `frontend/src/types/index.ts`.
- Exported Go constructors and methods use PascalCase: `NewWebhookHandler` in `internal/handlers/webhook.go`, `Setup` in `internal/router/router.go`, `Validate` and `BeforeCreate` in `internal/models/alert.go`.
- Unexported Go helpers use camelCase: `getEnvAsInt` and `getEnvAsDuration` in `internal/config/config.go`.
- React component functions use PascalCase when exported: `Alerts` in `frontend/src/pages/Alerts.tsx`, `Login` in `frontend/src/pages/Login.tsx`, `SeverityBadge` in `frontend/src/components/SeverityBadge.tsx`.
- Frontend event handlers and store actions use `handleX`/verb-first camelCase naming: `handleAckConfirm` in `frontend/src/pages/Alerts.tsx`, `fetchAlerts` and `quickSilence` in `frontend/src/stores/alertStore.ts`.
- Backend local variables favor short descriptive names such as `cfg`, `db`, `req`, `user`, `alert`, and `rule` in `cmd/server/main.go`, `internal/handlers/user.go`, and `internal/handlers/config.go`.
- Frontend local state uses descriptive camelCase names: `ackModalVisible`, `selectedAlert`, and `silenceDuration` in `frontend/src/pages/Alerts.tsx`.
- Boolean state is named explicitly with `is` or suffixes like `Loading`, `Visible`, and `Connected`, e.g. `loading` in `frontend/src/pages/Login.tsx`, `dataSourcesLoading` in `frontend/src/stores/configStore.ts`, and `wsConnected` in `frontend/src/stores/alertStore.ts`.
- Go struct types are singular domain nouns: `Alert`, `DataSource`, `RouteRule`, `SilenceRule`, `User`, and `OnDuty` in `internal/models/alert.go`, `internal/models/models.go`, and `internal/models/user.go`.
- TypeScript interfaces are also singular domain nouns and mirror backend JSON shapes: `Alert`, `Channel`, `OnDuty`, and `User` in `frontend/src/types/index.ts`.
- Request/response DTOs are explicitly suffixed in backend auth flow: `LoginRequest` and `LoginResponse` in `internal/handlers/user.go`.
## Code Style
- Frontend formatting is enforced by Prettier in `frontend/.prettierrc`.
- Use no semicolons, single quotes, 2-space indentation, trailing commas where valid in ES5, `printWidth: 100`, and always include arrow-function parentheses, matching `frontend/.prettierrc`.
- Backend formatting follows standard Go formatting conventions; the code in `cmd/server/main.go` and `internal/handlers/user.go` is `gofmt`-style with tabs and grouped imports.
- Frontend linting is configured in `frontend/.eslintrc.cjs`.
- Extend `eslint:recommended`, `plugin:@typescript-eslint/recommended`, and `plugin:react-hooks/recommended`.
- Treat `react-refresh/only-export-components` as a warning, `@typescript-eslint/no-explicit-any` as a warning, and allow intentionally unused arguments when prefixed with `_`.
- TypeScript strict mode is enabled in `frontend/tsconfig.json`, but `noUnusedLocals` and `noUnusedParameters` remain disabled. Code should still avoid dead state and unused props even when the compiler allows them.
## Import Organization
- `@/*` is configured in `frontend/tsconfig.json`, but current code primarily uses relative imports such as `../types` and `./pages`.
## State Management
- Use Zustand for shared server-backed state in `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, and `frontend/src/stores/userStore.ts`.
- Keep page-specific UI state local with `useState`, such as modal visibility, form draft state, and template preview state in `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Channels.tsx`, and `frontend/src/pages/DataSources.tsx`.
- Trigger initial data fetches with `useEffect` in page components, for example `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Dashboard.tsx`, and `frontend/src/pages/OnDuty.tsx`.
- Persist auth state directly in `localStorage` inside the user store and API interceptors, as shown in `frontend/src/stores/userStore.ts` and `frontend/src/api/client.ts`.
- Keep request state per-handler through struct receivers that hold shared dependencies, such as `UserHandler`, `WebhookHandler`, `ConfigHandler`, and `AlertHandler` in `internal/handlers/user.go`, `internal/handlers/webhook.go`, `internal/handlers/config.go`, and `internal/handlers/alert.go`.
- Store cross-request identity data in Gin context through middleware keys `user_id`, `username`, and `role` defined in `internal/middleware/auth.go`.
- Put persistent defaults and invariants in GORM hooks instead of controllers when the rule belongs to the model, as in `internal/models/alert.go` and `internal/models/models.go`.
## Error Handling
- Backend HTTP handlers return early on error and write JSON with an `error` field using `c.JSON(...)`, as in `internal/handlers/user.go`, `internal/handlers/alert.go`, `internal/handlers/config.go`, and `internal/handlers/webhook.go`.
- Handler code commonly distinguishes not-found cases from generic database failures using `gorm.ErrRecordNotFound`, as in `internal/handlers/user.go` and `internal/handlers/alert.go`.
- Infrastructure and service packages return wrapped errors with `fmt.Errorf(... %w ...)`, as in `internal/database/postgres.go`, `internal/database/redis.go`, and `internal/notifier/notifier.go`.
- Startup failures are treated as fatal in `cmd/server/main.go` and `internal/config/config.go`; use `log.Fatalf` or `os.Exit(1)` only for unrecoverable boot-time misconfiguration.
- Frontend shared stores either rethrow request errors after resetting loading state, as in `frontend/src/stores/alertStore.ts` and `frontend/src/stores/configStore.ts`, or log non-critical background refresh failures with `console.error`.
- Frontend pages convert API failures into PrimeReact toast feedback with `toast.showError(...)` and success paths into `toast.showSuccess(...)`, as in `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/Dashboard.tsx`, and `frontend/src/pages/Channels.tsx`.
## Validation
- Prefer model-owned validation methods for domain rules. `Validate()` exists on `Alert` in `internal/models/alert.go` and on `DataSource`, `Channel`, `RouteRule`, `SilenceRule`, and `OnDuty` in `internal/models/models.go`.
- Use GORM hooks such as `BeforeCreate` and `BeforeUpdate` to apply defaults and enforce validation before persistence, as in `internal/models/alert.go`, `internal/models/models.go`, and `internal/models/user.go`.
- Use Gin binding tags for required request payload fields on request-specific structs, such as `binding:"required"` on `LoginRequest` in `internal/handlers/user.go` and webhook/template preview request structs in `internal/handlers/webhook.go`.
- Some update endpoints intentionally accept partial payloads without strict validation, for example `internal/handlers/config.go` and `internal/handlers/user.go`. Follow the existing partial-update pattern when modifying those endpoints.
- Prefer inline validation with conditional checks for required UI fields, as in `frontend/src/pages/Login.tsx`, `frontend/src/pages/RouteRules.tsx`, and `frontend/src/pages/Silences.tsx`.
- Keep form field names aligned to backend JSON contracts using snake_case names such as `alert_name_pattern`, `channel_id`, and `user_name` in `frontend/src/pages/Silences.tsx` and `frontend/src/pages/OnDuty.tsx`.
- The frontend does not use a shared schema validator such as Zod or Yup. Validation is currently split between inline component-level checks, ad hoc transformation logic in pages like `frontend/src/pages/DataSources.tsx`, and backend model validation.
## Logging
- Use `log` or `slog` for process-level boot and infrastructure messages in `cmd/server/main.go`, `internal/database/postgres.go`, `internal/database/redis.go`, and `internal/config/config.go`.
- WebSocket and webhook flows print operational diagnostics directly in `internal/handlers/websocket.go` and `internal/handlers/webhook.go`.
- Frontend avoids structured logging; the only recurring pattern is `console.error(...)` for failed background refreshes in `frontend/src/stores/alertStore.ts`.
## Comments
- Existing comments mainly mark sections, explain business intent, or annotate bilingual behavior, for example in `internal/handlers/webhook.go`, `internal/router/router.go`, `internal/models/alert.go`, and `frontend/src/App.tsx`.
- Keep comments short and task-oriented. The repository favors concise operational comments over long explanatory blocks, especially in routing, webhook, and model files.
- Not used in frontend TypeScript files.
- Go code uses line comments above exported types and methods rather than doc blocks, for example in `internal/models/models.go` and `internal/handlers/user.go`.
## Function Design
- Backend methods prefer `(*Handler).Method(c *gin.Context)` signatures for HTTP work and pass dependencies through constructors like `NewUserHandler`, `NewWebhookHandler`, and `NewConfigHandler`.
- Frontend store actions and API methods use primitive IDs plus small object payloads, for example `ackAlert(id, comment)` in `frontend/src/stores/alertStore.ts` and `alertApi.ack(id, { comment })` in `frontend/src/api/client.ts`.
- Backend validation and service methods return `error` or typed values plus `error`, as in `internal/models/alert.go`, `internal/database/postgres.go`, and `internal/notifier/notifier.go`.
- Frontend async actions usually return `Promise<void>` and update store state internally, as in `frontend/src/stores/configStore.ts`.
## Module Design
- Frontend pages and components favor named exports from the implementation file and re-export through barrel files in `frontend/src/pages/index.ts` and `frontend/src/components/index.ts`.
- API wrappers export grouped objects such as `alertApi`, `dataSourceApi`, `channelApi`, `routeRuleApi`, `silenceRuleApi`, and `onDutyApi` from `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`.
- Backend packages group related behavior by layer: `internal/handlers`, `internal/models`, `internal/database`, `internal/middleware`, and `internal/ai`.
- Used on the frontend for pages and components.
- Not used in the Go backend.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## Pattern Overview
- `cmd/server/main.go` is the single backend bootstrap and wires config, PostgreSQL, Redis, Gin, and routes in one composition root.
- `internal/router/router.go` defines the public surface area and instantiates handler objects directly from infrastructure dependencies instead of routing through a service layer.
- `frontend/src/App.tsx` owns client-side routing, auth gating, and shell layout; pages consume Zustand stores and `frontend/src/api/*.ts` wrappers.
## Layers
- Purpose: Start the process, load configuration, initialize infrastructure, and start Gin.
- Location: `cmd/server/main.go`
- Contains: `.env` loading, config loading, DB/Redis initialization, Gin mode selection, router startup.
- Depends on: `internal/config/config.go`, `internal/database/postgres.go`, `internal/database/redis.go`, `internal/router/router.go`
- Used by: `make run`, `go run cmd/server/main.go`, compiled server binaries.
- Purpose: Declare HTTP and WebSocket endpoints, attach middleware, and bind handlers to route groups.
- Location: `internal/router/router.go`
- Contains: `/health`, `/api/v1/*`, `/webhook/:source_name`, `/webhook/test-template`, `/ws/alerts`
- Depends on: `internal/handlers/*.go`, `internal/auth/jwt.go`, `internal/middleware/auth.go`
- Used by: `cmd/server/main.go`
- Purpose: Execute request-specific workflows and persist/query state directly with GORM.
- Location: `internal/handlers/alert.go`, `internal/handlers/config.go`, `internal/handlers/user.go`, `internal/handlers/webhook.go`, `internal/handlers/websocket.go`
- Contains: CRUD handlers, datasource preview/testing endpoints, webhook ingestion, WebSocket connection management.
- Depends on: `gorm.DB`, `redis.Client`, `internal/models/*.go`, `internal/notifier/notifier.go`
- Used by: `internal/router/router.go`
- Purpose: Define persisted entities, JSON shapes, validation hooks, and limited domain helpers.
- Location: `internal/models/alert.go`, `internal/models/models.go`, `internal/models/user.go`
- Contains: `Alert`, `User`, `DataSource`, `Channel`, `RouteRule`, `SilenceRule`, `OnDuty`
- Depends on: GORM and `gorm.io/datatypes`
- Used by: database initialization, all handlers, notification routing.
- Purpose: Encapsulate infrastructure clients and external protocol details.
- Location: `internal/database/postgres.go`, `internal/database/redis.go`, `internal/auth/jwt.go`, `internal/notifier/notifier.go`
- Contains: GORM connection setup and migration, Redis client creation, JWT issue/validation, per-channel notification senders.
- Depends on: env-derived config from `internal/config/config.go`
- Used by: `cmd/server/main.go`, `internal/router/router.go`, `internal/handlers/*.go`
- Purpose: Mount the SPA, provide PrimeReact locale, define the authenticated layout, and map routes to pages.
- Location: `frontend/src/main.tsx`, `frontend/src/App.tsx`
- Contains: React root, `ConfigProvider`, `BrowserRouter`, `RequireAuth`, side menu, header.
- Depends on: PrimeReact, React Router, Zustand user store, page exports.
- Used by: browser entrypoint `frontend/index.html`
- Purpose: Centralize HTTP access and client-side state transitions.
- Location: `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`, `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/stores/userStore.ts`
- Contains: Axios instances, auth/token interceptors, alert/config/user stores.
- Depends on: backend REST API and browser `localStorage`
- Used by: page components in `frontend/src/pages/*.tsx`
- Purpose: Render screens and bind UI actions to stores/API calls.
- Location: `frontend/src/pages/Dashboard.tsx`, `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/DataSources.tsx`, `frontend/src/pages/Channels.tsx`, `frontend/src/pages/RouteRules.tsx`, `frontend/src/pages/Silences.tsx`, `frontend/src/pages/OnDuty.tsx`, `frontend/src/pages/Login.tsx`
- Contains: route-level screens for every product area.
- Depends on: `frontend/src/stores/*.ts`, `frontend/src/api/*.ts`, `frontend/src/components/*.tsx`, `frontend/src/types/index.ts`
- Used by: `frontend/src/App.tsx`
## Data Flow
## State Management
- Persistent state lives in PostgreSQL tables defined by `internal/models/alert.go`, `internal/models/models.go`, and `internal/models/user.go`.
- Ephemeral stream state is written to Redis by `internal/handlers/webhook.go`.
- In-process WebSocket client state is held in memory inside `internal/handlers/websocket.go`.
- Authentication state is stored in `localStorage` and mirrored in `frontend/src/stores/userStore.ts`.
- Alert query state and dashboard summaries live in `frontend/src/stores/alertStore.ts`.
- Configuration CRUD collections live in `frontend/src/stores/configStore.ts`.
## Key Abstractions
- Purpose: Bundle infrastructure dependencies per endpoint area.
- Examples: `internal/handlers/AlertHandler`, `internal/handlers/ConfigHandler`, `internal/handlers/UserHandler`, `internal/handlers/WebhookHandler`, `internal/handlers/WSHandler`
- Pattern: Thin constructor plus methods bound directly as Gin handlers.
- Purpose: Drive the alert ingestion and notification pipeline from persisted configuration.
- Examples: `internal/models/models.go` types `DataSource`, `RouteRule`, and `Channel`
- Pattern: Runtime behavior is data-driven through templates, matcher JSON, priorities, and enabled flags.
- Purpose: Keep endpoint URLs and transport concerns out of page components.
- Examples: `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`
- Pattern: Singleton Axios clients with interceptors; exported domain-specific method objects.
- Purpose: Hold UI-visible collections and wrap async fetch/update workflows.
- Examples: `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/stores/userStore.ts`
- Pattern: One store per major concern; pages call store methods instead of mutating state locally for shared data.
## Entry Points
- Location: `cmd/server/main.go`
- Triggers: `go run cmd/server/main.go`, `make run`, built binaries.
- Responsibilities: Load env/config, initialize PostgreSQL and Redis, set Gin mode, assemble router, and listen on `SERVER_PORT`.
- Location: `frontend/src/main.tsx`
- Triggers: `pnpm dev`, `pnpm build`, loading `frontend/index.html`
- Responsibilities: Create the React root and install PrimeReact locale configuration.
- Location: `frontend/src/App.tsx`
- Triggers: Imported by `frontend/src/main.tsx`
- Responsibilities: Define page routing, auth guards, top-level layout, and navigation chrome.
- Location: `internal/router/router.go`
- Triggers: `POST /webhook/:source_name`
- Responsibilities: Admit external alert payloads into the internal alert model and notification pipeline via `internal/handlers/webhook.go`.
- Location: `internal/router/router.go`
- Triggers: `GET /ws/alerts`
- Responsibilities: Upgrade to WebSocket and stream firing-alert state through `internal/handlers/websocket.go`.
## Error Handling
- Gin handlers in `internal/handlers/*.go` use `c.ShouldBindJSON`, `gorm.ErrRecordNotFound`, and immediate `c.JSON(status, gin.H{"error": ...})` responses.
- Infrastructure bootstrap in `cmd/server/main.go`, `internal/database/postgres.go`, and `internal/database/redis.go` treats initialization failures as fatal startup errors.
- Background operations in `internal/handlers/webhook.go` and `internal/handlers/websocket.go` often log or print failures and continue instead of surfacing structured telemetry.
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
