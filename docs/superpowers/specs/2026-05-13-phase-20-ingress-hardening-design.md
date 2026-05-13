---
name: phase-20
description: Harden webhook ingress and runtime readiness with request limits, rate limiting, and production config enforcement
metadata:
  type: spec
  source_phase: 20-harden-ingress-and-runtime-readiness
  milestone: v1.4
  status: pending
---

# Phase 20: Harden Ingress And Runtime Readiness

## Context & Goals

Phase 18-19 established the delivery ledger and recovery operations. Phase 20 hardens the webhook ingress and runtime readiness to prevent production incidents from malformed requests, traffic spikes, and missing configuration.

**Goal:** Webhook and service runtime entry points have explicit request protection, production config enforcement, and dependency readiness checks. Ingress anomalies and dedup save failures no longer silently drift.

## Success Criteria

1. Requests exceeding webhook body size limit are explicitly rejected with searchable failure records
2. Single datasource or source traffic spikes receive basic rate limiting protection, won't crush main service
3. Webhook signature/raw request validation executes on original body, avoiding JSON-binding semantic distortion
4. Production environment missing allowed origins, JWT secret, DB or Redis connection config refuses to start with dangerous defaults
5. `/readyz` truly reflects PostgreSQL, Redis, and key runtime dependencies availability; dedup update path DB save failures are properly handled, not silently lost

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Webhook body size limit middleware | `internal/middleware/request_limit.go` | INGR-01: size limit + rejection |
| Rate limiting middleware | `internal/middleware/rate_limit.go` | INGR-02: per-source rate limit |
| Raw body validation before JSON binding | `internal/handlers/webhook.go` | INGR-03: signature on raw body |
| Production config validation at startup | `cmd/server/main.go` | INGR-04: fail-fast on missing config |
| Readiness check endpoint | `internal/handlers/health.go` | INGR-05: `/readyz` with dependency checks |
| Dedup save failure handling | `internal/handlers/webhook.go` | DEBT-01: DB save failure not silent |

## Architecture

### Request Size Limit (INGR-01)

- Middleware checks `Content-Length` header before reading body
- Limit: configurable, default 1MB
- Oversized requests: return 413 with structured error JSON
- Log rejection with trace_id for searchability

### Rate Limiting (INGR-02)

- Per-source rate limit using Redis or in-memory sliding window
- Configurable: requests per minute per datasource
- Default: 1000 req/min per source
- Exceeded: return 429 with retry-after hint
- No distributed coordination required (single-instance baseline)

### Raw Body Validation (INGR-03)

- Read raw body first, store for signature/validation
- Only then bind to JSON struct
- Signature validation uses raw bytes, not re-serialized JSON

### Production Config Enforcement (INGR-04)

- At startup, validate required config:
  - `JWT_SECRET` (already required)
  - `ALLOWED_ORIGINS` (production must be explicit, not wildcard)
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
  - `REDIS_HOST`, `REDIS_PORT`
- If any missing in production mode (`SERVER_MODE=release`): exit with clear error message
- Development mode: allow defaults but log warning

### Readiness Check (INGR-05)

- `/readyz` endpoint checks:
  - PostgreSQL: ping DB connection
  - Redis: ping Redis connection
  - Return 200 if all healthy, 503 if any unhealthy
  - JSON response with component status

### Dedup Save Failure (DEBT-01)

- Current: `h.db.Save(&existing)` may fail silently
- Fix: wrap in error check, log failure with trace_id, continue without blocking webhook flow
- Failure is logged and traceable, not silently swallowed

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-20-01 | D | size limit | mitigate | Prevent oversized payloads from exhausting memory |
| T-20-02 | D | rate limit | mitigate | Prevent traffic spikes from crushing service |
| T-20-03 | T | raw body | mitigate | Signature validation on raw bytes, not re-serialized JSON |
| T-20-04 | I | config validation | mitigate | Production refuses dangerous defaults |
| T-20-05 | R | readiness | mitigate | `/readyz` reflects actual dependency state |

## Decisions

- Rate limiting is single-instance, in-memory or Redis-based
- Size limit default 1MB, configurable via env
- Production mode defined by `SERVER_MODE=release`
- Dedup save failure logged but doesn't block webhook flow

## Deviation Log

None — spec created from ROADMAP requirements.