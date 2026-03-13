# Architecture

**Analysis Date:** 2026-03-13

## Pattern Overview

**Overall:** Monolithic REST API with WebSocket support

**Key Characteristics:**
- Go backend using Gin web framework
- React frontend with Zustand state management
- PostgreSQL for persistent storage, Redis for caching/pub-sub
- JWT-based authentication
- OpenAI-compatible AI integration for alert analysis
- Webhook-based alert ingestion from external sources

## Layers

**Entry Point:**
- Location: `cmd/server/main.go`
- Responsibilities: Initialize config, database, Redis, router, and start HTTP server

**Router Layer:**
- Location: `internal/router/router.go`
- Purpose: Define all API routes and wire handlers
- Contains: Route definitions for /api/v1, /webhook, /ws endpoints

**Handler Layer:**
- Location: `internal/handlers/`
- Purpose: HTTP request handling and business logic orchestration
- Contains:
  - `alert.go` - Alert CRUD and operations (ack, silence)
  - `config.go` - DataSource, Channel, RouteRule, SilenceRule, OnDuty management
  - `user.go` - User authentication and management
  - `ai.go` - AI chat, suggestions, recommendations
  - `webhook.go` - External alert ingestion
  - `websocket.go` - Real-time alert streaming

**Model Layer:**
- Location: `internal/models/models.go`
- Purpose: Database entity definitions using GORM
- Contains: User, Alert, DataSource, Channel, RouteRule, SilenceRule, OnDuty, AILog, SilenceRecommendation

**Database Layer:**
- Location: `internal/database/`
- Contains: `postgres.go` - PostgreSQL connection and migrations, `redis.go` - Redis client

**Service Layer:**
- Location: `internal/ai/`, `internal/notifier/`, `internal/auth/`
- Contains:
  - `ai/client.go` - OpenAI-compatible AI API client
  - `notifier/notifier.go` - Notification senders (Feishu, DingTalk, WeCom, Webhook)
  - `auth/jwt.go` - JWT token handling
  - `middleware/auth.go` - JWT authentication middleware

**Config Layer:**
- Location: `internal/config/config.go`
- Purpose: Environment-based configuration loading

## Data Flow

**Alert Ingestion Flow:**
1. External system sends webhook to `/webhook/:source_name`
2. `webhook.go` validates and parses incoming alert
3. Applies deduplication via Redis
4. Applies grouping/aggregation rules
5. Creates Alert record in PostgreSQL
6. Matches against RouteRules to find notification channels
7. Sends notifications via `notifier.SendToChannel()`
8. Broadcasts to WebSocket clients via Redis pub/sub

**API Request Flow:**
1. Client sends HTTP request to `/api/v1/*`
2. Router matches route and applies middleware (CORS, JWT auth)
3. Handler processes request, interacts with database
4. Handler returns JSON response

**WebSocket Flow:**
1. Client connects to `/ws/alerts`
2. `websocket.go` upgrades connection
3. Subscribes to Redis pub/sub channel for new alerts
4. Pushes alerts to client in real-time

## Key Abstractions

**Alert Model:**
- Purpose: Represents a monitored alert event
- Examples: `internal/models/models.go` (Alert struct)
- Pattern: GORM model with JSON tags

**Channel Sender Interface:**
- Purpose: Abstract notification delivery
- Examples: `internal/notifier/notifier.go` (Sender interface)
- Pattern: Strategy pattern - FeishuSender, DingTalkSender, WeComSender, WebhookSender

**AI Client:**
- Purpose: Abstract AI API calls
- Examples: `internal/ai/client.go`
- Pattern: OpenAI-compatible API client supporting custom base URLs

## Entry Points

**Backend Server:**
- Location: `cmd/server/main.go`
- Triggers: Running `go run cmd/server/main.go` or `./shadowsongai.exe`
- Responsibilities: Initialize all dependencies, start Gin server on configured port

**Webhook Endpoint:**
- Location: `internal/router/router.go` - `webhook.POST("/:source_name", ...)`
- Triggers: External monitoring systems sending alerts
- Responsibilities: Receive, validate, process, route, and notify

**WebSocket Endpoint:**
- Location: `internal/router/router.go` - `r.GET("/ws/alerts", ...)`
- Triggers: Frontend clients connecting for real-time updates
- Responsibilities: Maintain connections, broadcast new alerts

## Error Handling

**Strategy:** Middleware-based error handling with Gin

**Patterns:**
- Handlers return HTTP status codes with JSON error messages
- Database errors wrapped with `fmt.Errorf` and returned as 500
- Validation errors return 400 with error details
- Not found returns 404
- Authentication failures return 401

## Cross-Cutting Concerns

**Logging:** Uses Go standard library `log/slog` for structured logging

**Validation:** Model-level validation via `Validate()` methods before database operations

**Authentication:** JWT tokens with role-based access control (admin, operator, viewer)

**CORS:** Hardcoded for localhost in development (router.go lines 17-30)

---

*Architecture analysis: 2026-03-13*
