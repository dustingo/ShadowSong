# External Integrations

**Analysis Date:** 2026-04-09

## APIs & External Services

**AI Services:**
- OpenAI-compatible chat completion API - used for AI assistant chat and alert analysis in `internal/ai/client.go` and `internal/handlers/ai.go`
  - SDK/Client: custom `net/http` client in `internal/ai/client.go`
  - Auth: `OPENAI_API_KEY`
  - Endpoint base: `OPENAI_API_BASE` defaulting to `https://api.openai.com/v1` in `internal/config/config.go`
  - Model selection: `AI_MODEL` in `internal/config/config.go`

**Notification Services:**
- Feishu bot webhook - outbound channel delivery in `internal/notifier/notifier.go`
  - SDK/Client: custom `net/http` sender in `internal/notifier/notifier.go`
  - Auth: channel-level `secret` inside persisted channel config; not environment-based
- DingTalk bot webhook - outbound channel delivery in `internal/notifier/notifier.go`
  - SDK/Client: custom `net/http` sender in `internal/notifier/notifier.go`
  - Auth: channel-level `secret` inside persisted channel config; not environment-based
- WeCom webhook - outbound channel delivery in `internal/notifier/notifier.go`
  - SDK/Client: custom `net/http` sender in `internal/notifier/notifier.go`
  - Auth: webhook key embedded in `webhook_url` channel config
- Custom webhook target - generic outbound notification delivery in `internal/notifier/notifier.go`
  - SDK/Client: custom `net/http` sender in `internal/notifier/notifier.go`
  - Auth: optional per-channel headers stored in channel config

**Browser-to-Backend Transport:**
- REST API from frontend to backend - Axios clients in `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`
  - SDK/Client: `axios`
  - Auth: JWT bearer token from browser storage
- WebSocket alert stream - dashboard live updates in `frontend/src/pages/Dashboard.tsx` and backend route in `internal/router/router.go`
  - SDK/Client: browser `WebSocket`, backend Gorilla WebSocket dependency in `go.mod`
  - Auth: no explicit WebSocket auth layer detected at route definition in `internal/router/router.go`

## Data Storage

**Databases:**
- PostgreSQL via GORM
  - Connection: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` in `internal/config/config.go`
  - Client: `gorm.io/gorm` with `gorm.io/driver/postgres` in `internal/database/postgres.go`
  - Schema bootstrap: auto table creation/migration in `internal/database/postgres.go`
  - Local dev container: `postgres:14-alpine` in `docker-compose.yml`

**File Storage:**
- Local filesystem only for source/build artifacts
- No application-level object storage integration detected

**Caching:**
- Redis used as runtime message infrastructure in `internal/database/redis.go` and `internal/handlers/webhook.go`
  - Connection: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` in `internal/config/config.go`
  - Client: `github.com/redis/go-redis/v9`
  - Usage: Redis stream publishing for newly created alerts in `internal/handlers/webhook.go`
  - Local dev container: `redis:7-alpine` in `docker-compose.yml`

## Authentication & Identity

**Auth Provider:**
- Custom JWT auth
  - Implementation: token issuance and validation in `internal/auth/jwt.go`
  - Secret: `JWT_SECRET` required by `internal/config/config.go`
  - Token TTL: `TOKEN_EXPIRY` in `internal/config/config.go`
  - Protected routes: JWT middleware on `/api/v1/users`, `/alerts`, `/datasources`, `/channels`, `/routes`, `/silences`, `/onduty`, and `/ai` in `internal/router/router.go`
  - Frontend persistence: bearer token stored in browser `localStorage` in `frontend/src/api/client.ts`, `frontend/src/api/auth.ts`, and Zustand user store under `frontend/src/stores/`

**Ingress API Authentication:**
- Data-source webhook authentication uses per-source API keys stored on `models.DataSource` and validated from `X-API-Key` or `Authorization: Bearer ...` in `internal/handlers/webhook.go`

## Monitoring & Observability

**Error Tracking:**
- No external error tracking service detected

**Logs:**
- Standard library logging via `log` and structured `slog` in `cmd/server/main.go`, `internal/config/config.go`, `internal/database/postgres.go`, and `internal/database/redis.go`
- AI interactions are persisted as application records in `models.AILog` via `internal/handlers/ai.go`

## CI/CD & Deployment

**Hosting:**
- Hosting platform not detected
- Development infrastructure is containerized locally through `docker-compose.yml`

**CI Pipeline:**
- No GitHub Actions, GitLab CI, or other pipeline config detected in repository root

## Environment Configuration

**Required env vars:**
- `JWT_SECRET` is mandatory in `internal/config/config.go`
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` for PostgreSQL in `internal/config/config.go`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` for Redis in `internal/config/config.go`
- `SERVER_PORT`, `SERVER_MODE` for backend server mode in `internal/config/config.go`
- `OPENAI_API_KEY`, `OPENAI_API_BASE`, `AI_MODEL`, `AI_TIMEOUT` for AI calls in `internal/config/config.go`
- `TOKEN_EXPIRY` for JWT lifetime in `internal/config/config.go`

**Secrets location:**
- Root `.env` may provide local development config and is loaded by `github.com/joho/godotenv` in `cmd/server/main.go`
- Production secrets are expected from process environment; no secret manager integration is detected
- Channel webhook URLs, secrets, and custom headers are stored in persisted channel config handled by `internal/notifier/notifier.go` and masked by handlers in `internal/handlers/config.go`

## Webhooks & Callbacks

**Incoming:**
- `POST /webhook/:source_name` receives external alert payloads in `internal/router/router.go` and `internal/handlers/webhook.go`
- `POST /webhook/test-template` tests input template rendering in `internal/router/router.go` and `internal/handlers/webhook.go`

**Outgoing:**
- Feishu webhook POSTs from `FeishuSender.Send` in `internal/notifier/notifier.go`
- DingTalk webhook POSTs from `DingTalkSender.Send` in `internal/notifier/notifier.go`
- WeCom webhook POSTs from `WeComSender.Send` in `internal/notifier/notifier.go`
- Custom outbound webhook requests from `WebhookSender.Send` in `internal/notifier/notifier.go`

---

*Integration audit: 2026-04-09*
