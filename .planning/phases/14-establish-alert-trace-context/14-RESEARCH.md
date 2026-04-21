# Phase 14: Establish Alert Trace Context - Research

**Researched:** 2026-04-21
**Domain:** Backend alert-path trace context and lifecycle observability
**Confidence:** MEDIUM

<user_constraints>
Verbatim copy from `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`]

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Phase 14 不只复用现有 `alert_id` 或 `fingerprint` 作为链路观测真源，而是要建立独立且稳定的 trace 标识，用于表达“一次 webhook 处理链路”的上下文。
- **D-02:** 这个 trace 标识必须在后端有明确真源，优先落在 `Alert` 相关持久化数据或可稳定重建的字段体系里，不能只存在于某一条临时日志输出中。
- **D-03:** 现有 `alert_id` 继续承担单条告警标识，`fingerprint` 继续承担去重/聚合语义；新 trace context 不能混淆这两类既有职责。
- **D-04:** Phase 14 的观测覆盖范围先限定在后端主链路：`webhook 接入 -> 告警渲染/落库/去重 -> Redis 发布 -> 路由匹配 -> 通知入口`。
- **D-05:** 本 phase 不要求把 trace context 继续暴露到 WebSocket 推送、前端 store 或页面展示；这些如果要做，应放到后续 phase 单独决策。
- **D-06:** 生命周期观测点至少要让运维能回答三类问题：这次 webhook 是否被接收、告警是否真正入库/被去重、后续是否进入路由和通知处理。
- **D-07:** Phase 14 只记录关键阶段结果，不做每个内部细节的细粒度事件流；目标是先建立稳定、可检索、低侵入的排障基线。
- **D-08:** 关键阶段应优先覆盖接入开始、告警创建/去重结果、Redis 发布结果、路由匹配结果和通知入口开始，而不是把模板函数内部或每个微小分支都打成事件。
- **D-09:** 允许在关键失败点补充更具体的上下文字段，但不把本 phase 扩展成通用审计事件系统或分布式 tracing 框架。
- **D-10:** Phase 14 必须保持现有告警接入、展示、路由、静默和值班能力不变；trace context 是增强排障能力，不是改变业务语义。
- **D-11:** 当前项目继续坚持 brownfield 小步增强，不引入消息队列、OpenTelemetry、Prometheus 或外部日志平台来完成本 phase 目标。
- **D-12:** 验证应证明从一条告警或通知相关日志可以回溯同一次处理链路中的至少多个关键阶段，而不是只证明字段“被写进某个 struct”。

### Claude's Discretion
- trace 字段的具体命名、落库方式和在日志中的序列化格式可由后续研究/计划阶段决定，只要满足“稳定真源、后端可回溯、最小侵入”。
- 关键观测点最终通过日志 helper、模型字段扩展还是 handler 内聚合辅助函数实现，可由 planner 决定，但应顺着现有 `WebhookHandler` 和 `Alert` 模型结构演进。

### Deferred Ideas (OUT OF SCOPE)
- 将 trace context 暴露到 WebSocket 消息或前端页面，形成端到端可视化链路
- 引入集中式 tracing / metrics 平台，例如 OpenTelemetry、Prometheus 或外部日志系统
- 把关键阶段观测扩展成更细粒度的事件流、审计表或通用事件总线

None of the above belongs in Phase 14 scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OBS-01 | Webhook 接入的单次告警处理会生成稳定的链路关联标识，并贯穿到后续关键处理阶段。 [VERIFIED: `.planning/REQUIREMENTS.md`] | Use one server-generated `trace_id` per webhook request, persist it on newly created `Alert` rows, and propagate it into Redis payloads plus lifecycle logs. [ASSUMED] |
| OBS-02 | 告警在接入、持久化、路由匹配和通知分发等关键阶段都会留下可检索的生命周期观测点。 [VERIFIED: `.planning/REQUIREMENTS.md`] | Add a single phase-local stage logging helper around the existing `WebhookHandler` flow so ingest, dedup/create, Redis publish, route match, and notification entry all emit the same correlation fields. [ASSUMED] |
</phase_requirements>

## Summary

`internal/handlers/webhook.go` already owns the full backend chain for this phase: datasource/auth checks, JSON parsing, template rendering, fingerprint generation, dedup-or-create, Redis `XAdd`, and async notification entry. That makes `WebhookHandler` the correct integration spine for Phase 14, and it also means a service-layer refactor would be unnecessary scope expansion. [VERIFIED: `internal/handlers/webhook.go`]

The strongest low-intrusion design is to create one server-side `trace_id` per webhook request, attach it to every newly created `Alert`, and reuse the same value in all stage logs and Redis messages. That preserves the existing meaning of `alert_id` and `fingerprint`, gives the async notification path a durable correlation field, and stays within the current Go + Gin + GORM + Redis stack. [ASSUMED]

The main planning nuance is dedup. A deduplicated webhook request does not create a new alert row today, so `Alert.TraceID` alone cannot be the only observability mechanism for that path. The plan should explicitly log a dedup stage with `trace_id`, `fingerprint`, and `existing_alert_id`, while keeping row persistence for newly created alerts as the durable source of truth. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Primary recommendation:** Add a nullable indexed `trace_id` field to `models.Alert`, generate one secure server-side trace once per `HandleWebhook` call, and funnel all Phase 14 lifecycle emission through one helper in `internal/handlers/webhook.go`. [ASSUMED]

## Project Constraints (from AGENTS.md)

- Keep the existing Go + Gin + GORM + PostgreSQL + Redis + React + Vite stack; do not introduce a platform or framework migration. [VERIFIED: `AGENTS.md`]
- Treat this as brownfield work: preserve current repo structure and unrelated changes, and avoid broad refactors outside the alert path. [VERIFIED: `AGENTS.md`]
- Do not break ingest, alert listing, routing, silencing, or on-duty behavior while adding trace context. [VERIFIED: `AGENTS.md`]
- Do not expand this phase into frontend trace display, WebSocket scope changes, or external observability infrastructure. [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`]

## Standard Stack

### Core

| Library / Tool | Version | Purpose in Phase 14 | Why Standard | Source |
|----------------|---------|---------------------|--------------|--------|
| Go | 1.25.0 | Generate trace IDs, hold request-scoped values, and emit logs from the existing backend path. | The repo runtime is pinned to Go 1.25.0, and the standard library already provides `crypto/rand`, `context`, and `log.Logger`, so no new tracing dependency is required for this phase. | [VERIFIED: `go.mod`] [CITED: https://pkg.go.dev/crypto/rand] [CITED: https://pkg.go.dev/context@go1.25.5] [CITED: https://pkg.go.dev/log] |
| Gin | v1.12.0 | Keep the webhook request as the natural trace boundary and optionally stash request-local values during handler execution. | `HandleWebhook` already runs on `*gin.Context`, and Gin documents request-local `Set` / `Get` support without requiring a new middleware framework. | [VERIFIED: `go.mod`] [VERIFIED: `internal/router/router.go`] [CITED: https://pkg.go.dev/github.com/gin-gonic/gin] |
| GORM | v1.31.1 | Persist `trace_id` on `Alert` rows and rely on the existing migrator path to add the column and index. | The repo already migrates `models.Alert` through the database bootstrap path, and GORM documents `AutoMigrate` plus index/model tags for additive schema changes. | [VERIFIED: `go.mod`] [VERIFIED: `internal/database/postgres.go`] [CITED: https://gorm.io/docs/migration.html] [CITED: https://gorm.io/docs/models.html] [CITED: https://gorm.io/docs/indexes.html] |
| go-redis/v9 | v9.18.0 | Propagate `trace_id` into the Redis stream payload emitted after new alert creation. | `publishToRedis` already writes the cross-process handoff message, so adding the same trace field there extends correlation without changing architecture. | [VERIFIED: `go.mod`] [VERIFIED: `internal/handlers/webhook.go`] |

### Supporting

| Library / Tool | Version | Purpose in Phase 14 | When to Use | Source |
|----------------|---------|---------------------|-------------|--------|
| `log.Logger` | stdlib | Keep lifecycle logs serialized and compatible with the existing `WebhookHandler.logger` seam used in tests. | Use for the phase-local stage helper instead of introducing a new logger abstraction before Phase 16. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [CITED: https://pkg.go.dev/log] |
| Testify + `go test` | v1.11.1 | Assert lifecycle log content and propagation behavior in backend tests. | Use to extend `internal/handlers/webhook_test.go` with trace-field expectations and stage coverage. | [VERIFIED: `go.mod`] [VERIFIED: `internal/handlers/webhook_test.go`] [VERIFIED: `Makefile`] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff | Source |
|------------|-----------|----------|--------|
| Secure random hex generated from stdlib | `github.com/google/uuid` | `google/uuid` is already an indirect dependency, but stdlib `crypto/rand` avoids promoting another package into the phase surface and still produces server-generated, unpredictable IDs. | [VERIFIED: `go.mod`] [CITED: https://pkg.go.dev/crypto/rand] [CITED: https://pkg.go.dev/github.com/google/uuid] |
| Phase-local helper in `webhook.go` or a tiny local observability package | OpenTelemetry / external tracing stack | External tracing is explicitly out of scope for Phase 14 and would violate the locked decision to avoid new platforms. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] |
| `Alert.trace_id` plus lifecycle logs for dedup-only requests | Dedicated lifecycle event table | A dedicated table would cover dedup-only persistence better, but it is heavier than the locked “key stages only, low intrusion” boundary for this phase. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [ASSUMED] |

**Installation:** No new packages are required for the recommended path. [VERIFIED: `go.mod`] [ASSUMED]

**Version verification:** The recommended implementation uses repo-pinned dependencies already present in `go.mod`; no new package version lookup is required for the planning baseline. [VERIFIED: `go.mod`]

## Architecture Patterns

### Recommended Project Structure

```text
internal/
├── handlers/
│   ├── webhook.go        # Generate trace_id, emit stage logs, propagate into Redis + async notification
│   └── webhook_test.go   # Verify trace propagation and stage observability
├── models/
│   └── alert.go          # Add nullable indexed trace_id to persisted alert rows
└── observability/        # Optional tiny helper package if planner wants to avoid growing webhook.go further
    └── trace.go          # trace ID generation + lifecycle field formatting
```

This keeps Phase 14 inside the existing handler-and-model architecture, which is how the codebase already composes request workflows today. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/models/alert.go`] [ASSUMED]

### Pattern 1: One Trace Per Webhook Request

**What:** Generate a single `trace_id` after datasource/auth validation and before the alert normalization loop, then reuse it for every alert derived from that webhook body. [ASSUMED]

**When to use:** Use for all `POST /webhook/:source_name` requests, including array payloads that normalize into multiple alerts. [VERIFIED: `internal/router/router.go`] [VERIFIED: `internal/handlers/webhook.go`]

**Why:** The phase goal is “one webhook handling chain,” not “one alert row,” and the current code can emit multiple alerts from one request via `normalizeData`. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [VERIFIED: `internal/handlers/webhook.go`]

### Pattern 2: Durable Row Field Plus Log Correlation

**What:** Persist `trace_id` on newly created `Alert` rows and emit the same field in lifecycle logs for every stage, including dedup paths that do not create a new row. [ASSUMED]

**When to use:** Apply on create success, dedup hit, Redis publish result, route-match result, and notification-entry start. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [VERIFIED: `internal/handlers/webhook.go`]

**Why:** New alerts need a durable source of truth, but dedup-only requests still need a searchable trace even though they reuse an existing row. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

### Pattern 3: Explicit Propagation Across Async Boundaries

**What:** Pass trace data explicitly into helper functions and the async notification boundary instead of relying on ambient request state. [ASSUMED]

**When to use:** Apply to `publishToRedis`, `processAlertNotificationsAsync`, `processAlertNotifications`, `findMatchedChannels`, and `sendNotification`. [VERIFIED: `internal/handlers/webhook.go`]

**Why:** The standard library recommends using `context` for request-scoped values that transit APIs, and the current code already crosses a goroutine boundary after the HTTP handler returns. An explicit value or stored `Alert.TraceID` is safer than depending on `*gin.Context` after response work has been scheduled. [CITED: https://pkg.go.dev/context@go1.25.5] [VERIFIED: `internal/handlers/webhook.go`]

### Anti-Patterns to Avoid

- **Reusing `alert_id` or `fingerprint` as the trace key:** The phase context explicitly forbids collapsing those semantics into the new correlation field. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`]
- **Generating a new trace for each normalized alert:** That breaks the “single webhook chain” requirement for array payloads. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`]
- **Logging only success paths:** Redis `XAdd` failures and dedup hits are currently silent or under-instrumented, so only logging happy-path create/send events would miss the point of Phase 14. [VERIFIED: `internal/handlers/webhook.go`]
- **Pushing raw request bodies into every lifecycle log:** The code persists raw payloads on the alert already; duplicating full raw JSON into stage logs would raise noise and disclosure risk without improving correlation. [VERIFIED: `internal/models/alert.go`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why | Source |
|---------|-------------|-------------|-----|--------|
| Trace ID generation | `time.Now().UnixNano()` or caller-supplied IDs | `crypto/rand`-backed server-generated IDs | The phase needs a distinct trace identity that is not confused with current business IDs, and the standard library already provides cryptographically secure random bytes. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [CITED: https://pkg.go.dev/crypto/rand] |
| Schema change path | Manual one-off SQL migration plumbing for this single field | Existing GORM migrator / `AutoMigrate` flow | The database bootstrap already uses `Migrator().AutoMigrate(table)` for existing tables, which is the least disruptive path for an additive indexed column. | [VERIFIED: `internal/database/postgres.go`] [CITED: https://gorm.io/docs/migration.html] |
| Correlation transport | A new event bus or tracing platform | Existing `Alert` model field + Redis payload + phase-local logging helper | The locked phase boundary explicitly rejects infrastructure expansion and wants only key lifecycle observability. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] |
| Architecture cleanup | New service layer just for trace propagation | Existing `WebhookHandler` orchestration seam | The request flow already lives in the handler, and this phase should enhance it rather than reorganize the application. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/codebase/ARCHITECTURE.md`] |

**Key insight:** Phase 14 is an observability baseline, not a tracing-platform rollout. Reuse the current handler/model seams and reserve broader logging standardization for Phase 16. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [VERIFIED: `.planning/REQUIREMENTS.md`]

## Common Pitfalls

### Pitfall 1: Treating `Alert.TraceID` as sufficient for dedup

**What goes wrong:** A webhook that deduplicates against an existing alert leaves no new row to carry the current request’s trace. [VERIFIED: `internal/handlers/webhook.go`]

**Why it happens:** The dedup path increments and saves the existing alert, then continues without creating a new alert object for the current request. [VERIFIED: `internal/handlers/webhook.go`]

**How to avoid:** Emit an explicit dedup lifecycle log with `trace_id`, `existing_alert_id`, and `fingerprint`, and treat row persistence as the durable source only for newly created alerts. [ASSUMED]

**Warning signs:** You can query `trace_id` on created rows but cannot explain why a repeated webhook produced no new alert. [ASSUMED]

### Pitfall 2: Losing correlation at the async notification boundary

**What goes wrong:** Notification-side logs keep `alert_id` and channel fields but cannot be tied back to the webhook request that spawned them. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`]

**Why it happens:** `processAlertNotificationsAsync` receives `[]models.Alert`, but `logNotification` currently does not emit any trace field. [VERIFIED: `internal/handlers/webhook.go`]

**How to avoid:** Make `TraceID` part of the alert model and include it in the existing notification helper’s common field set. [ASSUMED]

**Warning signs:** Notification failure logs only show `alert_id`/`channel_id`, and operators still need multiple ad hoc searches to reconstruct one request path. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

### Pitfall 3: Ignoring Redis publish observability

**What goes wrong:** A request may create alerts successfully but fail to publish to Redis with no searchable trace-stage evidence. [VERIFIED: `internal/handlers/webhook.go`]

**Why it happens:** `publishToRedis` currently does not inspect or log the `XAdd` result. [VERIFIED: `internal/handlers/webhook.go`]

**How to avoid:** Capture `XAdd` success/failure per alert and emit a stage log with `trace_id`, `alert_id`, and stream name. [ASSUMED]

**Warning signs:** Operators can see stored alerts but cannot tell whether the downstream Redis handoff happened. [ASSUMED]

### Pitfall 4: Turning Phase 14 into a logging-system rewrite

**What goes wrong:** The plan balloons into global logger replacement, structured logging migration, or frontend trace display work. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [VERIFIED: `.planning/REQUIREMENTS.md`]

**Why it happens:** The codebase has mixed `log.Printf` and `fmt.Printf` habits, which makes “fix all logging now” tempting. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/websocket.go`] [VERIFIED: `cmd/server/main.go`]

**How to avoid:** Limit Phase 14 to the webhook path and one lifecycle helper; defer global field naming/output unification to Phase 16. [VERIFIED: `.planning/REQUIREMENTS.md`] [ASSUMED]

**Warning signs:** The plan starts touching unrelated handlers, WebSocket code, or frontend routes without Phase 14 requirements forcing it. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`]

## Code Examples

Verified patterns from official sources:

### Secure Trace ID Generation

```go
// Source: https://pkg.go.dev/crypto/rand
package trace

import (
  "crypto/rand"
  "encoding/hex"
)

func NewTraceID() string {
  b := make([]byte, 16)
  rand.Read(b)
  return hex.EncodeToString(b)
}
```

This fits the phase because it adds no new dependency and produces a server-generated ID that is independent from `alert_id` and `fingerprint`. [CITED: https://pkg.go.dev/crypto/rand] [ASSUMED]

### Typed Request-Scoped Context Accessor

```go
// Source: https://pkg.go.dev/context@go1.25.5
package trace

import "context"

type key struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
  return context.WithValue(ctx, key{}, traceID)
}

func FromContext(ctx context.Context) (string, bool) {
  traceID, ok := ctx.Value(key{}).(string)
  return traceID, ok
}
```

The important part is the typed key pattern; the planner can choose whether to use this directly or keep propagation as explicit function parameters. [CITED: https://pkg.go.dev/context@go1.25.5] [ASSUMED]

### Additive Indexed Model Field

```go
// Source: https://gorm.io/docs/models.html
type Alert struct {
  AlertID string `gorm:"primaryKey;type:varchar(64)" json:"alert_id"`
  TraceID string `gorm:"index;type:varchar(64)" json:"trace_id"`
}
```

GORM documents `type` and `index` tags for model fields, and the repo already migrates `models.Alert` through its bootstrap migrator path. [CITED: https://gorm.io/docs/models.html] [VERIFIED: `internal/database/postgres.go`] [ASSUMED]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact | Source |
|--------------|------------------|--------------|--------|--------|
| Ad hoc lifecycle logs without a request-level correlation field | One independent `trace_id` per webhook request carried through key stages | Required by v1.3 requirements drafted 2026-04-21 | Enables a single search key across ingest, persistence, Redis, routing, and notification entry. | [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] |
| Reusing business identifiers for observability | Keep `alert_id` and `fingerprint` semantics separate from trace semantics | Locked in Phase 14 context on 2026-04-21 | Avoids conflating row identity, dedup grouping, and request correlation. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] |
| Infrastructure-heavy tracing rollout | Minimal in-app trace helper plus persisted field and lifecycle logs | Locked for Phase 14 on 2026-04-21 | Preserves brownfield velocity and avoids new deployment dependencies. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [VERIFIED: `AGENTS.md`] |

**Deprecated/outdated:**

- Treating `alert_id` or `fingerprint` as the “good enough” correlation field for webhook-to-notification troubleshooting is no longer acceptable for Phase 14 planning. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Persisting `trace_id` on newly created alert rows plus stage logs for dedup-only requests is sufficient to satisfy Phase 14 without adding an event table. | Summary, Architecture Patterns, Common Pitfalls | Planner may under-scope persistence needs for dedup-only troubleshooting. |
| A2 | A small helper in `internal/handlers/webhook.go` or a tiny local `internal/observability/trace.go` package is acceptable without triggering a broader architecture refactor. | Recommended Project Structure | Planner may choose the wrong seam and create unnecessary churn. |
| A3 | A 16-byte random trace encoded to 32 hex chars is the right length/format for this phase’s stable correlation field. | Code Examples, Don't Hand-Roll | Planner may need to revisit column size or external search ergonomics. |

## Open Questions (RESOLVED)

1. **Should `trace_id` be returned in the webhook HTTP response body?**  
   **RESOLVED:** No. Phase 14 remains backend-only, and exposing `trace_id` in the HTTP response is deferred because the locked scope does not include client-facing trace UX. The plan should keep trace truth inside backend persistence, Redis propagation, and lifecycle logs only. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [ASSUMED]

2. **How durable must dedup-only request lookup be after logs age out?**  
   **RESOLVED:** Phase 14 accepts a log-first answer for dedup-only requests. Newly created alerts get durable `Alert.TraceID`; dedup requests must emit explicit lifecycle logs with `trace_id`, `fingerprint`, and `existing_alert_id`, but they do not require a new event table or long-term dedup trace store in this phase. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [ASSUMED]

3. **Should the planner standardize field names now or wait for Phase 16?**  
   **RESOLVED:** Lock only the minimum Phase 14 field contract now: `trace_id`, `stage`, `alert_id`, `fingerprint`, `existing_alert_id`, `source`, and Redis metadata where applicable. Broader logging-field standardization remains Phase 16 work. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [ASSUMED]

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control | Source |
|---------------|---------|------------------|--------|
| V2 Authentication | yes | Preserve existing datasource API-key validation and only create trace context after the request has passed datasource lookup and API-key checks. | [VERIFIED: `internal/handlers/webhook.go`] |
| V3 Session Management | no | This phase does not change JWT/session flows or browser auth state. | [VERIFIED: `internal/router/router.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] |
| V4 Access Control | no | This phase does not change capability-gated API routes or frontend authorization scope. | [VERIFIED: `internal/router/router.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] |
| V5 Input Validation | yes | Generate `trace_id` server-side, keep current JSON/template/model validation behavior, and do not trust caller-supplied correlation fields. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [ASSUMED] |
| V6 Cryptography | yes | Use the standard library `crypto/rand` for unpredictable trace IDs; do not use guessable clock-based values for this field. | [CITED: https://pkg.go.dev/crypto/rand] [ASSUMED] |

### Known Threat Patterns for Go + Gin + GORM + Redis Alert Path

| Pattern | STRIDE | Standard Mitigation | Source |
|---------|--------|---------------------|--------|
| Caller-supplied trace spoofing | Spoofing | Ignore inbound trace headers/body fields for Phase 14 and mint the trace on the server after auth succeeds. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md`] [ASSUMED] |
| Async correlation loss | Repudiation | Propagate `trace_id` explicitly into goroutines and Redis payloads instead of depending on ephemeral request-local state after the response path continues. | [VERIFIED: `internal/handlers/webhook.go`] [CITED: https://pkg.go.dev/context@go1.25.5] [ASSUMED] |
| Raw payload leakage in logs | Information Disclosure | Emit stable lifecycle fields and bounded error context, not full `alert.Raw` content, in new trace-stage logs. | [VERIFIED: `internal/models/alert.go`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| Silent Redis handoff failure | Repudiation / Availability | Check and log `XAdd` results with `trace_id` so operators can distinguish create success from publish failure. | [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |

## Sources

### Primary (HIGH confidence)

- `.planning/phases/14-establish-alert-trace-context/14-CONTEXT.md` - locked scope, trace semantics, out-of-scope boundaries.
- `.planning/REQUIREMENTS.md` - `OBS-01` and `OBS-02` requirement truth.
- `internal/handlers/webhook.go` - actual ingest, dedup, persistence, Redis, and notification code path.
- `internal/handlers/webhook_test.go` - existing notification log seam and backend test style.
- `internal/models/alert.go` - current alert persistence model and field semantics.
- `internal/database/postgres.go` - current migration path for additive schema changes.
- [Go `context` docs](https://pkg.go.dev/context@go1.25.5) - request-scoped value propagation guidance.
- [Go `crypto/rand` docs](https://pkg.go.dev/crypto/rand) - secure random ID generation.
- [Go `log` docs](https://pkg.go.dev/log) - logger behavior and goroutine-safe serialization.
- [Gin package docs](https://pkg.go.dev/github.com/gin-gonic/gin) - `Context` request-local storage surface.
- [GORM migration docs](https://gorm.io/docs/migration.html) - additive schema migration behavior.
- [GORM model docs](https://gorm.io/docs/models.html) - field tags such as `type` and `index`.
- [GORM index docs](https://gorm.io/docs/indexes.html) - index tag behavior during migration.

### Secondary (MEDIUM confidence)

- [google/uuid package docs](https://pkg.go.dev/github.com/google/uuid) - comparison point for ID generation alternatives.

### Tertiary (LOW confidence)

- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - The phase stays on repo-pinned dependencies and official library docs. [VERIFIED: `go.mod`] [CITED: https://pkg.go.dev/github.com/gin-gonic/gin] [CITED: https://gorm.io/docs/migration.html]
- Architecture: MEDIUM - The main seams are verified in code, but the exact split between `webhook.go` and an optional helper package is still a planning choice. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]
- Pitfalls: MEDIUM - The failure points are visible in current code, but the final mitigation details still depend on planner decisions. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Research date:** 2026-04-21
**Valid until:** 2026-05-21
