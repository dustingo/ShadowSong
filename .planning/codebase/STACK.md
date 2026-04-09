# Technology Stack

**Analysis Date:** 2026-04-09

## Languages

**Primary:**
- Go 1.25.0 - backend application and API server in `go.mod`, `cmd/server/main.go`, `internal/`
- TypeScript 5.3.x - frontend application in `frontend/package.json`, `frontend/src/`, `frontend/tsconfig.json`

**Secondary:**
- TSX / React JSX - frontend UI composition in `frontend/src/App.tsx` and `frontend/src/pages/`
- SQL via GORM models/migrations - relational persistence wired through `internal/database/postgres.go` and `internal/models/`
- YAML - container orchestration in `docker-compose.yml`
- Make - developer task runner in `Makefile`

## Runtime

**Environment:**
- Go runtime 1.25.0 from `go.mod`
- Node.js 18+ expected by project docs in `README.md`
- Browser runtime for the SPA built from `frontend/src/`

**Package Manager:**
- Go modules for backend dependency management in `go.mod`
- pnpm for frontend dependency management in `frontend/package.json`
- Lockfile: present in `go.sum` and `frontend/pnpm-lock.yaml`

## Frameworks

**Core:**
- Gin `v1.12.0` - HTTP server, routing, middleware, WebSocket entrypoints in `go.mod` and `internal/router/router.go`
- GORM `v1.31.1` - ORM, schema migration, model persistence in `go.mod` and `internal/database/postgres.go`
- React `^18.2.0` - frontend SPA in `frontend/package.json` and `frontend/src/App.tsx`
- React Router DOM `^6.21.1` - client-side routing in `frontend/package.json` and `frontend/src/App.tsx`
- Ant Design `^5.12.8` - component library in `frontend/package.json` and `frontend/src/`
- Zustand `^4.4.7` - frontend client state stores in `frontend/package.json` and `frontend/src/stores/`

**Testing:**
- Testify `v1.11.1` - Go assertions/helpers declared in `go.mod`
- Go built-in `go test` runner - invoked by `Makefile`
- No frontend test runner config detected under `frontend/`

**Build/Dev:**
- Vite `^5.0.11` - frontend dev server and production bundling in `frontend/package.json` and `frontend/vite.config.ts`
- TypeScript compiler `^5.3.3` - type-check/build step in `frontend/package.json`
- ESLint `^8.56.0` - frontend linting in `frontend/package.json`
- Prettier `^3.1.1` - frontend formatting in `frontend/package.json`
- Docker Compose - local Postgres/Redis provisioning in `docker-compose.yml`
- `godotenv` `v1.5.1` - `.env` loading at startup in `go.mod` and `cmd/server/main.go`
- Make - common dev/build targets in `Makefile`

## Key Dependencies

**Critical:**
- `github.com/gin-gonic/gin` `v1.12.0` - API surface and middleware execution in `internal/router/router.go`
- `gorm.io/gorm` `v1.31.1` - model persistence and migrations in `internal/database/postgres.go`
- `gorm.io/driver/postgres` `v1.6.0` - PostgreSQL driver in `internal/database/postgres.go`
- `github.com/redis/go-redis/v9` `v9.18.0` - Redis connectivity and stream publishing in `internal/database/redis.go` and `internal/handlers/webhook.go`
- `github.com/golang-jwt/jwt/v5` - JWT issuance and validation in `internal/auth/jwt.go`
- `axios` `^1.6.5` - frontend HTTP client layer in `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`

**Infrastructure:**
- `github.com/joho/godotenv` `v1.5.1` - local env loading in `cmd/server/main.go`
- `github.com/gorilla/websocket` `v1.5.3` - backend WebSocket support declared in `go.mod` and exposed via `internal/router/router.go`
- `@vitejs/plugin-react` `^4.2.1` - Vite React integration in `frontend/package.json` and `frontend/vite.config.ts`
- `echarts` `^5.4.3` and `echarts-for-react` `^3.0.2` - dashboard charting in `frontend/src/pages/Dashboard.tsx`
- `@monaco-editor/react` `^4.6.0` - in-browser editor component in `frontend/src/components/CodeEditor.tsx`
- `react-markdown` `^10.1.0` - AI/markdown rendering in `frontend/src/pages/AIAssistant.tsx` and `frontend/src/pages/Dashboard.tsx`
- `dayjs` `^1.11.19` - date formatting in `frontend/src/components/AlertCard.tsx` and multiple `frontend/src/pages/*.tsx`

## Configuration

**Environment:**
- Runtime configuration is loaded from process environment plus optional `.env` via `cmd/server/main.go`
- Central config parsing lives in `internal/config/config.go`
- Required secret: `JWT_SECRET` in `internal/config/config.go`
- Database vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` in `internal/config/config.go`
- Redis vars: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` in `internal/config/config.go`
- Server vars: `SERVER_PORT`, `SERVER_MODE` in `internal/config/config.go`
- AI vars: `OPENAI_API_KEY`, `OPENAI_API_BASE`, `AI_MODEL`, `AI_TIMEOUT` in `internal/config/config.go`
- Token expiry var: `TOKEN_EXPIRY` in `internal/config/config.go`
- `.env` file is present at project root; contents not inspected

**Build:**
- Backend build/run/test commands are defined in `Makefile`
- Frontend build pipeline is defined in `frontend/package.json`
- Frontend module resolution and aliasing are defined in `frontend/tsconfig.json` and `frontend/vite.config.ts`
- Dev proxying for `/api`, `/webhook`, and `/ws` is defined in `frontend/vite.config.ts`
- Local container runtime config is defined in `docker-compose.yml`

## Platform Requirements

**Development:**
- Go toolchain compatible with `go 1.25.0` in `go.mod`
- Node.js 18+ and pnpm per `README.md`
- Docker and Docker Compose for local Postgres/Redis per `README.md` and `docker-compose.yml`
- PostgreSQL 14 and Redis 7 containers for local services in `docker-compose.yml`

**Production:**
- Production hosting target is not explicitly defined
- Backend expects a reachable PostgreSQL instance, Redis instance, and environment-injected secrets/config from `internal/config/config.go`
- Frontend build output is generated by Vite, but deployment host is not specified in repository config

---

*Stack analysis: 2026-04-09*
