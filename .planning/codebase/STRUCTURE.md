# Codebase Structure

**Analysis Date:** 2026-03-13

## Directory Layout

```
D:\goproject\shadowsongAI/
├── cmd/                          # Application entry points
│   └── server/
│       └── main.go               # Main server entry point
├── internal/                     # Private application code
│   ├── ai/                       # AI client implementation
│   │   └── client.go
│   ├── auth/                     # Authentication (JWT)
│   │   └── jwt.go
│   ├── config/                   # Configuration loading
│   │   └── config.go
│   ├── database/                 # Database connections
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── handlers/                 # HTTP request handlers
│   │   ├── alert.go
│   │   ├── ai.go
│   │   ├── config.go
│   │   ├── user.go
│   │   ├── webhook.go
│   │   └── websocket.go
│   ├── middleware/               # HTTP middleware
│   │   └── auth.go
│   ├── models/                   # Data models
│   │   ├── models.go
│   │   ├── alert.go
│   │   ├── user.go
│   │   └── alert_test.go
│   ├── notifier/                 # Notification senders
│   │   └── notifier.go
│   └── router/                   # Route definitions
│       └── router.go
├── frontend/                     # React frontend application
│   ├── src/
│   │   ├── main.tsx              # Frontend entry point
│   │   ├── App.tsx               # Main app component with routing
│   │   ├── api/                  # API client layer
│   │   │   ├── client.ts         # Axios instance
│   │   │   └── auth.ts           # Auth endpoints
│   │   ├── components/           # Reusable UI components
│   │   │   ├── index.ts
│   │   │   ├── AlertCard.tsx
│   │   │   ├── CodeEditor.tsx
│   │   │   └── SeverityBadge.tsx
│   │   ├── pages/                # Page components
│   │   │   ├── index.ts
│   │   │   ├── Alerts.tsx
│   │   │   ├── Channels.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   ├── DataSources.tsx
│   │   │   ├── Login.tsx
│   │   │   ├── OnDuty.tsx
│   │   │   ├── RouteRules.tsx
│   │   │   ├── Silences.tsx
│   │   │   └── AIAssistant.tsx
│   │   ├── stores/               # Zustand state stores
│   │   │   ├── alertStore.ts
│   │   │   ├── configStore.ts
│   │   │   └── userStore.ts
│   │   ├── types/                # TypeScript type definitions
│   │   │   └── index.ts
│   │   ├── utils/                # Utility functions
│   │   │   └── formatter.ts
│   │   └── hooks/                # Custom React hooks (empty)
│   └── package.json
├── .env                          # Environment configuration
├── go.mod                        # Go module definition
├── go.sum                        # Go dependencies lock
├── docker-compose.yml            # Docker Compose config
└── Makefile                      # Build automation
```

## Directory Purposes

**cmd/server:**
- Purpose: Application entry point
- Contains: `main.go` - Server bootstrap

**internal/ai:**
- Purpose: AI API client implementation
- Contains: OpenAI-compatible client for alert analysis

**internal/auth:**
- Purpose: JWT token generation and validation
- Contains: JWT implementation

**internal/config:**
- Purpose: Configuration management
- Contains: Environment variable loading

**internal/database:**
- Purpose: Database connection management
- Contains: PostgreSQL and Redis initialization

**internal/handlers:**
- Purpose: HTTP request handling
- Contains: Business logic for each domain

**internal/middleware:**
- Purpose: HTTP middleware
- Contains: Authentication middleware

**internal/models:**
- Purpose: Data models
- Contains: GORM model definitions

**internal/notifier:**
- Purpose: Notification delivery
- Contains: Channel-specific senders

**internal/router:**
- Purpose: Route registration
- Contains: All API route definitions

**frontend/src:**
- Purpose: Frontend application source
- Contains: React components, pages, stores

## Key File Locations

**Entry Points:**
- `cmd/server/main.go` - Backend server entry
- `frontend/src/main.tsx` - Frontend entry

**Configuration:**
- `internal/config/config.go` - Backend config
- `.env` - Environment variables
- `docker-compose.yml` - Container orchestration

**Core Logic:**
- `internal/handlers/` - Business logic
- `internal/models/models.go` - Data models

**Frontend State:**
- `frontend/src/stores/` - Zustand stores

## Naming Conventions

**Backend:**
- Files: snake_case (e.g., `alert_handler.go`)
- Functions: PascalCase (e.g., `func List() {}`)
- Variables: camelCase (e.g., `alertHandler`)
- Types/Structs: PascalCase (e.g., `AlertHandler`)
- Packages: snake_case (e.g., `internal/handlers`)

**Frontend:**
- Files: PascalCase for components (e.g., `AlertCard.tsx`), camelCase for utilities (e.g., `formatter.ts`)
- Components: PascalCase (e.g., `function AlertCard()`)
- Hooks: camelCase starting with use (e.g., `useUserStore`)
- Variables: camelCase

## Where to Add New Code

**New Backend Feature:**
- Handler: `internal/handlers/`
- Model: `internal/models/models.go`
- Route: `internal/router/router.go`

**New Notification Channel:**
- Implementation: `internal/notifier/notifier.go`
- Add new sender type following the Sender interface pattern

**New Frontend Feature:**
- Page: `frontend/src/pages/`
- Component: `frontend/src/components/`
- API endpoint: `frontend/src/api/client.ts`
- State: `frontend/src/stores/`

**New API Endpoint:**
- Backend handler: `internal/handlers/`
- Route registration: `internal/router/router.go`
- Frontend API client: `frontend/src/api/client.ts`
- Frontend store: `frontend/src/stores/`

---

*Structure analysis: 2026-03-13*
