# Codebase Concerns

**Analysis Date:** 2026-04-10

## Tech Debt

**Monolithic backend handlers:**
- Issue: CRUD, validation, persistence, templating, routing, deduplication, and notification dispatch are concentrated in a few oversized handlers instead of smaller services.
- Files: `internal/handlers/webhook.go`, `internal/handlers/config.go`, `internal/handlers/user.go`
- Impact: Small changes carry high regression risk because unrelated behaviors share the same functions and database coupling. The largest hotspot is `internal/handlers/webhook.go` at 879 lines; `internal/handlers/config.go` is 455 lines.
- Fix approach: Extract template rendering, deduplication, routing, and channel secret handling into dedicated packages under `internal/`. Keep HTTP handlers thin and focused on binding plus response formatting.

**Business logic duplicated across handlers and frontend state:**
- Issue: Alert acknowledgement, quick silence, and config refresh behavior are implemented ad hoc in multiple places instead of through a shared service contract.
- Files: `internal/handlers/alert.go`, `frontend/src/stores/alertStore.ts`, `frontend/src/stores/configStore.ts`, `frontend/src/pages/Dashboard.tsx`
- Impact: Backend behavior changes can silently desynchronize frontend optimistic updates and polling or websocket logic.
- Fix approach: Centralize alert mutations and config fetch flows behind explicit backend service methods and typed frontend API helpers with invariant checks.

**Dead or half-integrated features remain in active code paths:**
- Issue: The codebase exposes configuration for grouped notifications, but key execution paths are missing or unfinished.
- Files: `internal/models/models.go`, `internal/handlers/webhook.go`, `frontend/src/pages/DataSources.tsx`
- Impact: Operators can configure features that the system does not actually execute, which creates false confidence and support burden.
- Fix approach: Either remove the unused UI/API surface or finish the implementation before exposing the feature flags.

## Known Bugs

**Data source template testing UI is wired to non-existent backend endpoints:**
- Symptoms: The frontend defines `/datasources/:id/test-input` and `/datasources/:id/test-output` clients, but the backend only exposes `/webhook/test-template`.
- Files: `frontend/src/api/client.ts`, `frontend/src/pages/DataSources.tsx`, `internal/router/router.go`, `internal/handlers/webhook.go`
- Trigger: Any future call to `dataSourceApi.testInput` or `dataSourceApi.testOutput` returns 404. The current drawer button in `frontend/src/pages/DataSources.tsx` only reopens the drawer instead of invoking an API, so the feature is visibly unfinished.
- Workaround: Use `POST /webhook/test-template` directly. Do not rely on the data source test drawer as a working feature.

**Quick silence can report success even if silence rule persistence fails:**
- Symptoms: The alert status is saved first, then `h.db.Create(&silence)` is called without error handling.
- Files: `internal/handlers/alert.go`
- Trigger: Database write failure during quick silence creation leaves the alert marked `silenced` without a corresponding `SilenceRule`.
- Workaround: Verify the silence rule exists after using quick silence; manual correction may be required.

## Security Considerations

**Bootstrap admin password is printed to logs:**
- Risk: First-run credentials are emitted directly through `log.Printf("Default admin user created: admin / %s", pwd)`.
- Files: `internal/database/postgres.go`
- Current mitigation: The generated password is random and the user model hashes it before storage.
- Recommendations: Replace log emission with one-time secure delivery or require explicit admin bootstrap input. Never print credentials in application logs.

**Channel secrets are masked in list responses but fully exposed on detail fetch:**
- Risk: `ListChannels` masks configs through `maskChannelConfig`, while `GetChannel` intentionally returns the raw config for editing.
- Files: `internal/handlers/config.go`
- Current mitigation: The list endpoint hides common secret keys by returning `{"masked": true}`.
- Recommendations: Return redacted detail payloads plus explicit secret rotation/update fields. Avoid sending stored webhook secrets back to the browser.

**Data source API keys are retrievable and copyable in the browser:**
- Risk: Editing a data source performs `dataSourceApi.get(record.id)` and stores the returned `api_key` in local component state; the UI also exposes clipboard copy.
- Files: `frontend/src/pages/DataSources.tsx`, `internal/handlers/config.go`, `internal/models/models.go`
- Current mitigation: None beyond requiring authentication on config endpoints.
- Recommendations: Treat data source API keys as write-only secrets. Return placeholders and rotation controls instead of the stored secret value.

**JWT is stored in `localStorage`:**
- Risk: The frontend reads and writes the bearer token from `localStorage`, which is vulnerable to token theft if any XSS is introduced.
- Files: `frontend/src/api/client.ts`
- Current mitigation: None. Logout only deletes the local token.
- Recommendations: Move auth to `HttpOnly` secure cookies or introduce a hardened token storage/refresh strategy with CSRF protection.

**Fallback alerts can echo raw inbound payload data into persisted alert messages:**
- Risk: On template render failures, `createFallbackAlert` embeds truncated raw request content into `Alert.Message`.
- Files: `internal/handlers/webhook.go`
- Current mitigation: The message is truncated to 500 characters.
- Recommendations: Store a sanitized diagnostic code instead of raw payload fragments, especially if external systems may include secrets or PII.

## Performance Bottlenecks

**Webhook ingestion performs synchronous per-alert database lookups and writes:**
- Problem: Every inbound alert runs fingerprint generation, a deduplication lookup, potential update/save, insert, Redis publish, route matching, and per-channel notification work.
- Files: `internal/handlers/webhook.go`
- Cause: The handler loops alert-by-alert and performs multiple DB operations inside the loop rather than batching or offloading work.
- Improvement path: Move ingestion to a queue or worker model, batch deduplication checks when possible, and isolate notification fan-out from request latency.

**Alert statistics endpoint scales with many serial count queries:**
- Problem: `Stats` issues separate counts for totals, status buckets, each severity, and 24 hourly trend slices.
- Files: `internal/handlers/alert.go`
- Cause: The endpoint performs many sequential queries instead of aggregating in SQL.
- Improvement path: Use grouped SQL queries or pre-aggregated metrics. Cache dashboard stats if near-real-time precision is not required.

**Dashboard duplicates real-time and polling traffic:**
- Problem: The frontend opens a websocket and also polls `fetchActiveAlerts` plus `fetchStats` every 10 seconds.
- Files: `frontend/src/pages/Dashboard.tsx`, `frontend/src/stores/alertStore.ts`
- Cause: Real-time delivery is not trusted as the single source of truth, likely because websocket integration is incomplete.
- Improvement path: Make websocket updates authoritative for active alerts and use targeted refreshes instead of unconditional interval polling.

## Fragile Areas

**Webhook matching uses a custom pseudo-regex helper:**
- Files: `internal/handlers/webhook.go`
- Why fragile: `regexpMatch` only simulates prefix, suffix, and contains behavior. Route rule authors may expect real regex semantics from `LabelMatcher.Pattern`, but many patterns will behave incorrectly.
- Safe modification: Replace the helper with the Go `regexp` package behind validation and pre-compilation. Add tests for exact, prefix, suffix, alternation, and escaped patterns before changing behavior.
- Test coverage: Handler tests still do not cover route matching, fingerprinting, template rendering, or deduplication.

**WebSocket stack is only partially wired:**
- Files: `internal/handlers/websocket.go`, `internal/router/router.go`, `frontend/src/pages/Dashboard.tsx`
- Why fragile: `BroadcastAlert` exists but has no callers, so the dashboard relies on polling for freshness. The heartbeat goroutine writes pings until failure without explicit cancellation.
- Safe modification: Introduce a single publisher path from alert creation to websocket fan-out and add lifecycle management for per-connection goroutines.
- Test coverage: No websocket tests or frontend connection-state tests are present.

**Config CRUD writes trust bound input too broadly:**
- Files: `internal/handlers/config.go`
- Why fragile: Multiple update handlers call `ShouldBindJSON` and then overwrite model fields directly with limited validation or no error handling around `Save` or `Delete`.
- Safe modification: Add request DTOs per endpoint, validate every mutable field, and fail closed when optional payload fields are omitted.
- Test coverage: No tests cover config mutation, secret masking, route reorder, silence creation, or on-duty scheduling.

## Scaling Limits

**Single-process in-memory websocket registry:**
- Current capacity: One process holds all active websocket clients in `map[*websocket.Conn]bool`.
- Limit: Horizontal scaling breaks broadcast visibility because connections and broadcasts are not shared across instances.
- Scaling path: Back websocket fan-out with Redis pub/sub or a broker so alerts can be broadcast across replicas.

**Notification execution is best-effort and unbounded per request burst:**
- Current capacity: New alerts spawn asynchronous notification processing directly from the request handler.
- Limit: Large inbound batches can create many downstream HTTP sends without backpressure, retries, or worker limits.
- Scaling path: Push notification jobs to a queue with bounded worker concurrency, retry policy, and delivery metrics.

## Historical Cleanup Context

- Phase 1 and Phase 2 removed backend and frontend AI runtime surfaces. Remaining references to that work belong in phase summaries and verification reports, not in the active codebase map.
- This concern map intentionally tracks only current non-AI runtime risks. Historical AI removal details should be treated as migration context rather than present architecture.

## Missing Critical Features

**Grouped alert aggregation is exposed but not implemented:**
- Problem: `DataSource` includes `group_enabled` and `group_window`, and the frontend lets operators configure them, but webhook processing only implements deduplication and never aggregates grouped notifications.
- Blocks: Operators cannot rely on the advertised group-aggregation workflow to reduce noise.

**Secret rotation and audit controls are absent:**
- Problem: API keys, webhook configs, and channel secrets can be viewed or edited, but there is no rotation history, no reveal audit, and no permission separation for secret access.
- Blocks: Safer operational handoff and compliance-friendly secret management.

## Test Coverage Gaps

**Handler and integration behavior is still under-tested:**
- What's not tested: Authentication flows, user authorization, webhook ingestion, deduplication, routing, websocket delivery, notifier behavior, and config CRUD.
- Files: `internal/handlers/*.go`, `internal/router/router.go`, `internal/notifier/notifier.go`, `internal/middleware/auth.go`, `internal/auth/jwt.go`
- Risk: High-risk operational paths can regress without detection; current committed coverage is still narrow compared with the runtime surface.
- Priority: High

**Frontend has no automated test setup:**
- What's not tested: Stores, pages, token handling, websocket reconnection, and form flows.
- Files: `frontend/package.json`, `frontend/src/pages/*.tsx`, `frontend/src/stores/*.ts`, `frontend/src/api/client.ts`
- Risk: UI regressions and API contract mismatches can ship unnoticed. `frontend/package.json` has no `test` script.
- Priority: High

**Security-sensitive secret handling is unverified:**
- What's not tested: Masking behavior in channel lists, raw secret exposure on detail fetch, data source API key retrieval, and fallback alert content sanitization.
- Files: `internal/handlers/config.go`, `internal/handlers/webhook.go`, `frontend/src/pages/DataSources.tsx`
- Risk: Secret leakage can persist as an accidental feature because no tests lock down the expected redaction behavior.
- Priority: High

---

*Concerns audit: 2026-04-10*
