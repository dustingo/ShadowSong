# External Integrations

**Analysis Date:** 2026-03-13

## APIs & External Services

**AI/LLM:**
- OpenAI-compatible API - AI alert analysis
  - SDK: Native HTTP client in `internal/ai/client.go`
  - Env vars: `OPENAI_API_KEY`, `OPENAI_API_BASE` (defaults to OpenAI)
  - Supports both OpenAI and MiniMax (configurable via `OPENAI_API_BASE`)
  - Default model: GPT-4 (`AI_MODEL` env var)
  - Used for: Alert analysis, summarization, root cause identification

**Chinese Messaging Platforms:**

- **Feishu (Lark)** - Notification channel
  - Implementation: `internal/notifier/notifier.go` - `FeishuSender`
  - Config: `webhook_url`, `secret` in channel config
  - API: Custom webhook with signature verification

- **DingTalk** - Notification channel
  - Implementation: `internal/notifier/notifier.go` - `DingTalkSender`
  - Config: `webhook_url`, `secret` in channel config
  - API: Custom webhook

- **WeCom (WeChat Work)** - Notification channel
  - Implementation: `internal/notifier/notifier.go` - `WeComSender`
  - Config: `webhook_url` in channel config
  - API: Custom webhook

**Generic Webhook:**
- Custom webhook sender for external systems
  - Implementation: `internal/notifier/notifier.go` - `WebhookSender`
  - Config: `url`, `method`, `headers`, `template`

## Data Storage

**Databases:**
- **PostgreSQL 14+**
  - Connection: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
  - ORM: GORM
  - Client: `gorm.io/driver/postgres`
  - Models: User, Alert, DataSource, Channel, RouteRule, SilenceRule, OnDuty, AILog, SilenceRecommendation

**Cache/Message Queue:**
- **Redis 7+**
  - Connection: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
  - Client: `github.com/redis/go-redis/v9`
  - Used for: Alert stream processing (`alerts:pending` stream), session management

**File Storage:**
- Local filesystem only (no cloud storage integration detected)

## Authentication & Identity

**JWT Authentication:**
- Implementation: `internal/auth/jwt.go`
- Secret: `JWT_SECRET` environment variable (required)
- Token expiry: Configurable via `TOKEN_EXPIRY` (default 24 hours)
- Used for: API authentication, session management

**Data Source API Keys:**
- Per-datasource API key validation
- Headers: `X-API-Key` or `Authorization: Bearer <token>`
- Implemented in `internal/handlers/webhook.go`

## Monitoring & Observability

**Error Tracking:**
- Not detected (no Sentry, Bugsnag, or similar integration)

**Logs:**
- Go standard `log/slog` package
- GORM SQL logging in warning mode

## CI/CD & Deployment

**Container Orchestration:**
- Docker Compose for local development
- `docker-compose.yml` includes PostgreSQL and Redis services

**Deployment:**
- Standalone Go binary compilation
- Binary: `shadowsongai.exe` (Windows) or similar for target platform

## Environment Configuration

**Required env vars:**
- `JWT_SECRET` - Required, authentication secret
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` - Database connection
- `REDIS_HOST`, `REDIS_PORT` - Redis connection
- `SERVER_PORT` - HTTP server port (default 8080)
- `OPENAI_API_KEY` - AI service (optional but required for AI features)

**Optional env vars:**
- `DB_SSLMODE`, `REDIS_PASSWORD`, `REDIS_DB`
- `SERVER_MODE` - debug/release
- `OPENAI_API_BASE`, `AI_MODEL`, `AI_TIMEOUT`
- `TOKEN_EXPIRY`

**Secrets location:**
- `.env` file in project root (gitignored)

## Webhooks & Callbacks

**Incoming Webhooks:**
- Alert ingestion webhook: `/webhook/:source_name`
  - Accepts JSON alerts from monitoring systems
  - Supports Prometheus, AlertManager, and custom formats
  - Uses template system for input normalization
  - Per-source API key validation

**Outgoing:**
- Feishu webhook
- DingTalk webhook
- WeCom webhook
- Generic HTTP webhook

---

*Integration audit: 2026-03-13*
