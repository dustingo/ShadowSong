# Architecture

**Analysis Date:** 2026-04-09

## Pattern Overview

**Overall:** Monorepo with a server-renderless split between a Go HTTP backend and a React SPA frontend.

**Key Characteristics:**
- `cmd/server/main.go` is the single backend bootstrap and wires config, PostgreSQL, Redis, Gin, and routes in one composition root.
- `internal/router/router.go` defines the public surface area and instantiates handler objects directly from infrastructure dependencies instead of routing through a service layer.
- `frontend/src/App.tsx` owns client-side routing, auth gating, and shell layout; pages consume Zustand stores and `frontend/src/api/*.ts` wrappers.

## Layers

**Backend Bootstrap Layer:**
- Purpose: Start the process, load configuration, initialize infrastructure, and start Gin.
- Location: `cmd/server/main.go`
- Contains: `.env` loading, config loading, DB/Redis initialization, Gin mode selection, router startup.
- Depends on: `internal/config/config.go`, `internal/database/postgres.go`, `internal/database/redis.go`, `internal/router/router.go`
- Used by: `make run`, `go run cmd/server/main.go`, compiled server binaries.

**Backend Routing Layer:**
- Purpose: Declare HTTP and WebSocket endpoints, attach middleware, and bind handlers to route groups.
- Location: `internal/router/router.go`
- Contains: `/health`, `/api/v1/*`, `/webhook/:source_name`, `/webhook/test-template`, `/ws/alerts`
- Depends on: `internal/handlers/*.go`, `internal/auth/jwt.go`, `internal/middleware/auth.go`
- Used by: `cmd/server/main.go`

**Backend Handler Layer:**
- Purpose: Execute request-specific workflows and persist/query state directly with GORM.
- Location: `internal/handlers/alert.go`, `internal/handlers/config.go`, `internal/handlers/ai.go`, `internal/handlers/user.go`, `internal/handlers/webhook.go`, `internal/handlers/websocket.go`
- Contains: CRUD handlers, AI chat endpoints, webhook ingestion, WebSocket connection management.
- Depends on: `gorm.DB`, `redis.Client`, `internal/models/*.go`, `internal/ai/client.go`, `internal/notifier/notifier.go`
- Used by: `internal/router/router.go`

**Backend Domain Model Layer:**
- Purpose: Define persisted entities, JSON shapes, validation hooks, and limited domain helpers.
- Location: `internal/models/alert.go`, `internal/models/models.go`, `internal/models/user.go`
- Contains: `Alert`, `User`, `DataSource`, `Channel`, `RouteRule`, `SilenceRule`, `OnDuty`, `AILog`, `SilenceRecommendation`
- Depends on: GORM and `gorm.io/datatypes`
- Used by: database initialization, all handlers, notification routing.

**Infrastructure Layer:**
- Purpose: Encapsulate infrastructure clients and external protocol details.
- Location: `internal/database/postgres.go`, `internal/database/redis.go`, `internal/auth/jwt.go`, `internal/ai/client.go`, `internal/notifier/notifier.go`
- Contains: GORM connection setup and migration, Redis client creation, JWT issue/validation, OpenAI-compatible HTTP client, per-channel notification senders.
- Depends on: env-derived config from `internal/config/config.go`
- Used by: `cmd/server/main.go`, `internal/router/router.go`, `internal/handlers/*.go`

**Frontend Application Shell Layer:**
- Purpose: Mount the SPA, provide Ant Design locale, define the authenticated layout, and map routes to pages.
- Location: `frontend/src/main.tsx`, `frontend/src/App.tsx`
- Contains: React root, `ConfigProvider`, `BrowserRouter`, `RequireAuth`, side menu, header.
- Depends on: Ant Design, React Router, Zustand user store, page exports.
- Used by: browser entrypoint `frontend/index.html`

**Frontend State and Transport Layer:**
- Purpose: Centralize HTTP access and client-side state transitions.
- Location: `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`, `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/stores/userStore.ts`
- Contains: Axios instances, auth/token interceptors, alert/config/user stores.
- Depends on: backend REST API and browser `localStorage`
- Used by: page components in `frontend/src/pages/*.tsx`

**Frontend Feature/Page Layer:**
- Purpose: Render screens and bind UI actions to stores/API calls.
- Location: `frontend/src/pages/Dashboard.tsx`, `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/DataSources.tsx`, `frontend/src/pages/Channels.tsx`, `frontend/src/pages/RouteRules.tsx`, `frontend/src/pages/Silences.tsx`, `frontend/src/pages/OnDuty.tsx`, `frontend/src/pages/AIAssistant.tsx`, `frontend/src/pages/Login.tsx`
- Contains: route-level screens for every product area.
- Depends on: `frontend/src/stores/*.ts`, `frontend/src/api/*.ts`, `frontend/src/components/*.tsx`, `frontend/src/types/index.ts`
- Used by: `frontend/src/App.tsx`

## Data Flow

**Authenticated CRUD Request Flow:**

1. The SPA route in `frontend/src/App.tsx` renders a page only through `RequireAuth`, which checks the token in `frontend/src/stores/userStore.ts`.
2. Page components call store actions in `frontend/src/stores/alertStore.ts` or `frontend/src/stores/configStore.ts`, or call auth functions in `frontend/src/api/auth.ts`.
3. Axios clients in `frontend/src/api/client.ts` and `frontend/src/api/auth.ts` attach the `Authorization` header from `localStorage` and send requests to `/api/v1/*`.
4. Gin groups in `internal/router/router.go` apply `middleware.JWTAuth` and optional `middleware.RequireRole` from `internal/middleware/auth.go`.
5. The selected handler in `internal/handlers/*.go` binds input, queries or mutates GORM models from `internal/models/*.go`, and returns JSON directly.

**Webhook Ingestion and Notification Flow:**

1. External systems post to `/webhook/:source_name`, handled by `internal/handlers/webhook.go`.
2. `WebhookHandler.HandleWebhook` loads the matching `DataSource` from `internal/models/models.go`, validates the API key, and reads the raw request body.
3. The handler normalizes payloads, renders alerts through the data source `InputTemplate`, generates a fingerprint, performs deduplication, and writes new alerts to PostgreSQL.
4. New alerts are published to the Redis stream `alerts:pending` through `publishToRedis` in `internal/handlers/webhook.go`.
5. `processAlertNotifications` loads enabled `RouteRule` rows, resolves matching `Channel` rows, renders the `OutputTemplate`, and dispatches through `internal/notifier/notifier.go`.

**Alert Dashboard Refresh Flow:**

1. Dashboard components call `fetchActiveAlerts` and `fetchStats` in `frontend/src/stores/alertStore.ts`.
2. `frontend/src/api/client.ts` requests `/api/v1/alerts/active` and `/api/v1/alerts/stats`.
3. `internal/handlers/alert.go` runs direct aggregate queries on `models.Alert` and returns summary JSON.
4. Store state updates drive cards, tables, and charts on `frontend/src/pages/Dashboard.tsx` and `frontend/src/pages/Alerts.tsx`.

**AI Assistant Flow:**

1. `frontend/src/pages/AIAssistant.tsx` sends chat requests through `aiApi.chat` in `frontend/src/api/client.ts`.
2. `internal/handlers/ai.go` calls `internal/ai/client.go` when `OPENAI_API_KEY` is present, otherwise emits a fallback message.
3. The handler persists the exchange as `models.AILog` and returns reply text plus canned follow-up suggestions.

**WebSocket Flow:**

1. The browser connects to `/ws/alerts`, proxied by `frontend/vite.config.ts` to the Go server.
2. `internal/handlers/websocket.go` upgrades the connection, sends an initial firing-alert snapshot, and maintains a client registry.
3. `WSHandler.BroadcastAlert` can push `new_alert` payloads to all connected clients through an in-memory channel.
4. No backend file currently calls `BroadcastAlert`, so the WebSocket path is structurally present but not integrated into the webhook ingestion path.

## State Management

**Backend State:**
- Persistent state lives in PostgreSQL tables defined by `internal/models/alert.go`, `internal/models/models.go`, and `internal/models/user.go`.
- Ephemeral stream state is written to Redis by `internal/handlers/webhook.go`.
- In-process WebSocket client state is held in memory inside `internal/handlers/websocket.go`.

**Frontend State:**
- Authentication state is stored in `localStorage` and mirrored in `frontend/src/stores/userStore.ts`.
- Alert query state and dashboard summaries live in `frontend/src/stores/alertStore.ts`.
- Configuration CRUD collections live in `frontend/src/stores/configStore.ts`.

## Key Abstractions

**Handler Structs:**
- Purpose: Bundle infrastructure dependencies per endpoint area.
- Examples: `internal/handlers/AlertHandler`, `internal/handlers/ConfigHandler`, `internal/handlers/AIHandler`, `internal/handlers/UserHandler`, `internal/handlers/WebhookHandler`, `internal/handlers/WSHandler`
- Pattern: Thin constructor plus methods bound directly as Gin handlers.

**Configuration Entity Trio:**
- Purpose: Drive the alert ingestion and notification pipeline from persisted configuration.
- Examples: `internal/models/models.go` types `DataSource`, `RouteRule`, and `Channel`
- Pattern: Runtime behavior is data-driven through templates, matcher JSON, priorities, and enabled flags.

**Axios API Modules:**
- Purpose: Keep endpoint URLs and transport concerns out of page components.
- Examples: `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`
- Pattern: Singleton Axios clients with interceptors; exported domain-specific method objects.

**Zustand Stores:**
- Purpose: Hold UI-visible collections and wrap async fetch/update workflows.
- Examples: `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/stores/userStore.ts`
- Pattern: One store per major concern; pages call store methods instead of mutating state locally for shared data.

## Entry Points

**Backend Process Entry Point:**
- Location: `cmd/server/main.go`
- Triggers: `go run cmd/server/main.go`, `make run`, built binaries.
- Responsibilities: Load env/config, initialize PostgreSQL and Redis, set Gin mode, assemble router, and listen on `SERVER_PORT`.

**Frontend Browser Entry Point:**
- Location: `frontend/src/main.tsx`
- Triggers: `pnpm dev`, `pnpm build`, loading `frontend/index.html`
- Responsibilities: Create the React root and install Ant Design locale configuration.

**Frontend App Routing Entry Point:**
- Location: `frontend/src/App.tsx`
- Triggers: Imported by `frontend/src/main.tsx`
- Responsibilities: Define page routing, auth guards, top-level layout, and navigation chrome.

**Webhook Integration Entry Point:**
- Location: `internal/router/router.go`
- Triggers: `POST /webhook/:source_name`
- Responsibilities: Admit external alert payloads into the internal alert model and notification pipeline via `internal/handlers/webhook.go`.

**Realtime Entry Point:**
- Location: `internal/router/router.go`
- Triggers: `GET /ws/alerts`
- Responsibilities: Upgrade to WebSocket and stream firing-alert state through `internal/handlers/websocket.go`.

## Error Handling

**Strategy:** Request handlers return JSON error bodies directly at the point of failure; there is no centralized error translation layer.

**Patterns:**
- Gin handlers in `internal/handlers/*.go` use `c.ShouldBindJSON`, `gorm.ErrRecordNotFound`, and immediate `c.JSON(status, gin.H{"error": ...})` responses.
- Infrastructure bootstrap in `cmd/server/main.go`, `internal/database/postgres.go`, and `internal/database/redis.go` treats initialization failures as fatal startup errors.
- Background operations in `internal/handlers/webhook.go` and `internal/handlers/websocket.go` often log or print failures and continue instead of surfacing structured telemetry.

## Cross-Cutting Concerns

**Logging:** Startup and infrastructure logging uses `log` in `cmd/server/main.go` and `internal/database/*.go`; config validation uses `slog` in `internal/config/config.go`; request-time async work often uses `fmt.Printf` in `internal/handlers/webhook.go`.

**Validation:** Input validation is split between Gin binding tags in handler-local structs, `Validate()` methods on models in `internal/models/*.go`, and GORM hooks such as `BeforeCreate` in `internal/models/alert.go` and `internal/models/models.go`.

**Authentication:** JWT creation and validation live in `internal/auth/jwt.go`; request enforcement lives in `internal/middleware/auth.go`; the frontend persists tokens in `frontend/src/stores/userStore.ts` and injects them through Axios interceptors in `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`.

**Configuration:** Runtime configuration is environment-driven through `internal/config/config.go`; frontend dev proxy settings live in `frontend/vite.config.ts`; build and run workflows live in `Makefile`.

---

*Architecture analysis: 2026-04-09*
