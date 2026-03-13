# Technology Stack

**Analysis Date:** 2026-03-13

## Languages

**Primary:**
- Go 1.25.0 - Backend API server, business logic
- TypeScript 5.3.3 - Frontend web application

**Secondary:**
- JavaScript/JSX - Frontend components

## Runtime

**Backend:**
- Go 1.25.0 - Standalone binary compilation

**Frontend:**
- Node.js 18+ (development)
- Browser (production)

**Package Manager:**
- pnpm 8+ (frontend)
- Go modules (backend)
- Lockfiles: `pnpm-lock.yaml`, `go.sum`

## Frameworks

**Backend:**
- Gin v1.12.0 - HTTP web framework
- GORM v1.31.1 - ORM for database operations
- gorilla/websocket v1.5.3 - WebSocket support

**Frontend:**
- React 18.2.0 - UI framework
- Vite 5.0.11 - Build tool and dev server
- React Router DOM 6.21.1 - Client-side routing
- Ant Design 5.12.8 - UI component library
- Zustand 4.4.7 - State management

**Testing:**
- testify 1.11.1 - Backend testing assertions

**Build/Dev:**
- TypeScript 5.3.3 - Type checking
- ESLint 8.56.0 - Code linting
- Prettier 3.1.1 - Code formatting

## Key Dependencies

**Backend Critical:**
- `github.com/gin-gonic/gin` v1.12.0 - HTTP router and middleware
- `gorm.io/gorm` v1.31.1 - Database ORM
- `gorm.io/driver/postgres` v1.6.0 - PostgreSQL driver
- `github.com/redis/go-redis/v9` v9.18.0 - Redis client
- `github.com/gorilla/websocket` v1.5.3 - WebSocket connections

**Frontend Critical:**
- `react` v18.2.0 - Core React library
- `antd` v5.12.8 - Enterprise UI components
- `@monaco-editor/react` v4.6.0 - Code editor component
- `axios` v1.6.5 - HTTP client
- `echarts` v5.4.3 - Charting library
- `react-markdown` v10.1.0 - Markdown rendering

## Configuration

**Environment:**
- `.env` file in root for Go backend
- Config via `os.Getenv()` in Go
- Environment variables documented in `internal/config/config.go`

**Key Configuration Options:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
- `SERVER_PORT`, `SERVER_MODE`
- `OPENAI_API_KEY`, `OPENAI_API_BASE`, `AI_MODEL`, `AI_TIMEOUT`
- `JWT_SECRET`, `TOKEN_EXPIRY`

**Frontend Build:**
- `vite.config.ts` - Vite configuration with proxy to backend
- `tsconfig.json` - TypeScript configuration with path alias `@/`
- Path alias: `@/*` maps to `src/*`

## Platform Requirements

**Development:**
- Node.js 18+ with pnpm
- Go 1.25+
- PostgreSQL 14+ (via Docker)
- Redis 7+ (via Docker)

**Production:**
- Go binary (compiled from `cmd/server/main.go`)
- PostgreSQL 14+ database
- Redis 7+ cache
- Nginx (optional, for reverse proxy)

---

*Stack analysis: 2026-03-13*
