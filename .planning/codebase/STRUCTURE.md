# Codebase Structure

**Analysis Date:** 2026-04-10

## Directory Layout

```text
shadowsongAI/
├── cmd/                 # Go process entrypoints
│   └── server/          # HTTP server bootstrap
├── internal/            # Backend application code
│   ├── auth/            # JWT token utilities
│   ├── config/          # Environment-backed config loading
│   ├── database/        # PostgreSQL and Redis setup
│   ├── handlers/        # Gin handlers for API, webhook, and WebSocket endpoints
│   ├── middleware/      # Gin auth middleware
│   ├── models/          # GORM models and tests
│   ├── notifier/        # Feishu/DingTalk/WeCom/webhook senders
│   └── router/          # Route registration
├── frontend/            # Vite React SPA
│   ├── src/
│   │   ├── api/         # Axios clients and endpoint wrappers
│   │   ├── components/  # Shared presentational UI
│   │   ├── pages/       # Route-level screens
│   │   ├── stores/      # Zustand state containers
│   │   ├── types/       # Shared TypeScript interfaces
│   │   ├── App.tsx      # Route shell and auth guard
│   │   └── main.tsx     # Browser mount point
│   ├── dist/            # Generated frontend build output
│   ├── package.json     # Frontend dependencies and scripts
│   └── vite.config.ts   # Dev server proxy and alias config
├── docs/                # Ad hoc project documentation
├── scripts/             # Verification and helper scripts
├── .planning/           # Planning and codebase maps
├── docker-compose.yml   # Local Postgres and Redis services
├── go.mod               # Backend module definition
├── Makefile             # Common dev commands
└── README.md            # Project overview and quick start
```

## Directory Purposes

**`cmd/`:**
- Purpose: Hold executable entrypoints only.
- Contains: `cmd/server/main.go`
- Key files: `cmd/server/main.go`
- Guidance: Add new processes here only when introducing another binary. Keep reusable logic in `internal/`.

**`internal/`:**
- Purpose: Hold all non-exported backend application code.
- Contains: Config, infra setup, handlers, models, middleware, routing, integrations.
- Key files: `internal/router/router.go`, `internal/config/config.go`, `internal/database/postgres.go`
- Guidance: Place new backend business logic in a focused package under `internal/`; wire it into `cmd/server/main.go` through constructors instead of adding logic to the bootstrap.

**`internal/handlers/`:**
- Purpose: Own HTTP, webhook, and WebSocket endpoint behavior.
- Contains: `internal/handlers/alert.go`, `internal/handlers/config.go`, `internal/handlers/user.go`, `internal/handlers/webhook.go`, `internal/handlers/websocket.go`
- Key files: `internal/handlers/webhook.go`, `internal/handlers/config.go`
- Guidance: Add a new handler file here when exposing a new endpoint group. Keep route registration in `internal/router/router.go`.

**`internal/models/`:**
- Purpose: Centralize GORM-backed entities and model validation.
- Contains: `internal/models/alert.go`, `internal/models/models.go`, `internal/models/user.go`, `internal/models/alert_test.go`
- Key files: `internal/models/alert.go`, `internal/models/models.go`, `internal/models/user.go`
- Guidance: Add new persistent entities here, including hooks and validation helpers, before migrating them from `internal/database/postgres.go`.

**`frontend/src/pages/`:**
- Purpose: Route-level UI features.
- Contains: `frontend/src/pages/Dashboard.tsx`, `frontend/src/pages/Alerts.tsx`, `frontend/src/pages/DataSources.tsx`, `frontend/src/pages/Channels.tsx`, `frontend/src/pages/RouteRules.tsx`, `frontend/src/pages/Silences.tsx`, `frontend/src/pages/OnDuty.tsx`, `frontend/src/pages/Login.tsx`
- Key files: `frontend/src/pages/index.ts`, `frontend/src/pages/Dashboard.tsx`
- Guidance: Put new navigable screens here and export them through `frontend/src/pages/index.ts`.

**`frontend/src/components/`:**
- Purpose: Shared UI pieces reused by pages.
- Contains: `frontend/src/components/AlertCard.tsx`, `frontend/src/components/CodeEditor.tsx`, `frontend/src/components/SeverityBadge.tsx`
- Key files: `frontend/src/components/index.ts`
- Guidance: Put reusable presentational units here; keep page-specific orchestration in `frontend/src/pages/`.

**`frontend/src/api/`:**
- Purpose: Define all browser-to-backend HTTP contracts.
- Contains: `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`
- Key files: `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`
- Guidance: Add new REST methods here first, then consume them from stores or pages.

**`frontend/src/stores/`:**
- Purpose: Hold shared client state and async actions.
- Contains: `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/stores/userStore.ts`
- Key files: `frontend/src/stores/userStore.ts`
- Guidance: Add a new store when state must be shared across pages; avoid duplicating API orchestration in multiple screens.

**`frontend/dist/`:**
- Purpose: Store generated production assets.
- Generated: Yes
- Committed: Yes
- Guidance: Do not place source files here. Source belongs under `frontend/src/`.

## Key File Locations

**Entry Points:**
- `cmd/server/main.go`: Backend composition root and server startup.
- `frontend/src/main.tsx`: Frontend browser bootstrap.
- `frontend/src/App.tsx`: Client route map and auth gate.

**Configuration:**
- `internal/config/config.go`: Backend environment variable loading and typed config.
- `frontend/vite.config.ts`: Frontend alias and dev proxy configuration.
- `docker-compose.yml`: Local infrastructure topology for PostgreSQL and Redis.
- `Makefile`: Standard project run/build/test commands.

**Core Logic:**
- `internal/router/router.go`: HTTP surface area and dependency assembly for handlers.
- `internal/handlers/webhook.go`: Alert ingestion, normalization, deduplication, Redis publish, and notification routing.
- `internal/handlers/alert.go`: Alert queries, stats, and acknowledgement workflows.
- `internal/handlers/config.go`: CRUD for data sources, channels, route rules, silence rules, and on-duty schedules.
- `internal/notifier/notifier.go`: Channel-type-specific notification delivery.
- `frontend/src/stores/alertStore.ts`: Dashboard and alert list state.
- `frontend/src/stores/configStore.ts`: Configuration management state.

**Testing:**
- `internal/models/alert_test.go`: Existing backend model tests.
- `internal/config/config_test.go`: Config loading regression tests after AI runtime removal.
- `internal/router/router_test.go`: Route presence/absence regression tests.
- `internal/handlers/webhook_test.go`: Webhook severity normalization regression tests.
- `scripts/verify_backend_no_ai.ps1`: Backend non-AI verification script.

## Ownership Boundaries

**Backend HTTP boundary:**
- `internal/router/router.go` owns path registration.
- `internal/handlers/*.go` own request/response behavior.
- `internal/middleware/auth.go` owns request authentication checks.

**Backend data boundary:**
- `internal/models/*.go` own persisted schema and validation rules.
- `internal/database/postgres.go` owns DB connection creation and schema migration.
- `internal/database/redis.go` owns Redis connection creation only.

**Backend integration boundary:**
- `internal/notifier/notifier.go` owns outbound notification-channel protocols.
- `internal/handlers/webhook.go` is the only place where inbound webhook payloads are translated into internal alerts.
- `scripts/verify_backend_no_ai.ps1` owns the scripted end-to-end verification path for the retained backend flow.

**Frontend routing boundary:**
- `frontend/src/App.tsx` owns route registration and navigation shell.
- `frontend/src/pages/*.tsx` own route-specific UI behavior.

**Frontend state and transport boundary:**
- `frontend/src/api/*.ts` own HTTP request details.
- `frontend/src/stores/*.ts` own shared async state transitions.
- `frontend/src/types/index.ts` owns the frontend-facing data contracts.

## Naming Conventions

**Files:**
- Go backend files use lowercase package-oriented names such as `internal/handlers/webhook.go` and `internal/database/postgres.go`.
- React page and component files use PascalCase names such as `frontend/src/pages/DataSources.tsx` and `frontend/src/components/AlertCard.tsx`.
- Store and utility files use camelCase names such as `frontend/src/stores/userStore.ts` and `frontend/src/utils/formatter.ts`.
- Barrel exports use `index.ts` in `frontend/src/pages/index.ts` and `frontend/src/components/index.ts`.

**Directories:**
- Backend directories are lowercase nouns by concern: `internal/router`, `internal/models`, `internal/notifier`.
- Frontend feature directories are grouped by technical role: `frontend/src/pages`, `frontend/src/components`, `frontend/src/stores`, `frontend/src/api`, `frontend/src/types`.

## Where to Add New Code

**New Backend API Feature:**
- Primary code: `internal/handlers/` for the endpoint, `internal/models/` for schema additions, and `internal/router/router.go` for route registration.
- Supporting integrations: `internal/notifier/` or a new `internal/<package>/` directory if the feature introduces a new external boundary.
- Tests: Place backend tests next to the package, following `internal/models/alert_test.go`.

**New Frontend Screen:**
- Implementation: `frontend/src/pages/<FeatureName>.tsx`
- Route registration: `frontend/src/pages/index.ts` and `frontend/src/App.tsx`
- Shared page state: `frontend/src/stores/`
- Backend calls: `frontend/src/api/client.ts` or `frontend/src/api/auth.ts`

**New Shared Frontend Component:**
- Implementation: `frontend/src/components/<ComponentName>.tsx`
- Re-export: `frontend/src/components/index.ts` if the component should be part of the shared component surface.

**New Utility or Shared Types:**
- Shared frontend helpers: `frontend/src/utils/`
- Shared frontend interfaces: `frontend/src/types/index.ts`
- Shared backend helpers: a focused package under `internal/` rather than adding helpers to `cmd/`

## Special Directories

**`.planning/codebase/`:**
- Purpose: Machine-consumable repository maps for planning and execution.
- Generated: Yes
- Committed: Intended to be committed with planning artifacts.

**`frontend/dist/`:**
- Purpose: Vite build output.
- Generated: Yes
- Committed: Yes

**`.claude/` and `.kiro/`:**
- Purpose: Tooling and specification artifacts used outside the runtime application path.
- Generated: No
- Committed: Yes

**`docs/`:**
- Purpose: Human-authored supplementary documentation.
- Generated: No
- Committed: Yes

---

*Structure analysis: 2026-04-10*
