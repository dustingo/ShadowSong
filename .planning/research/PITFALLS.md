# Domain Pitfalls

**Domain:** Delivery recovery, ingress protection, and operations visibility for a brownfield alert pipeline
**Researched:** 2026-04-29
**Overall confidence:** MEDIUM-HIGH

This research is specific to the proposed v1.4 scope: persistent delivery records, manual replay/retry tooling, ingress hardening, readiness checks, metrics, and an operations history/health surface. Recommendations below combine current official docs with brownfield inference from this repository's existing webhook -> persist -> dedup -> route -> notify pipeline.

## Critical Pitfalls

### Pitfall 1: Mutable delivery records make replay non-deterministic
**What goes wrong:** Stored delivery records point to live alert, channel, template, and route state instead of preserving the exact delivery contract used for the original attempt.
**Why it happens:** Teams add a `delivery_logs` table for observability, then later try to use the same rows for replay without freezing payload, rendered content, receiver metadata, attempt number, and failure class.
**Consequences:** Manual replay sends something different from the original notification, operators cannot explain what changed, and audit history stops being trustworthy.
**Warning signs:** Replay code re-renders from current templates/config only; records store foreign keys and error text but not immutable request snapshot.
**Prevention:** Introduce a delivery ledger first. Persist immutable delivery intent plus normalized attempt/result fields before any replay UI/API ships.
**Absorb in roadmap:** **Phase 1 - Delivery record contract and persistence.**

### Pitfall 2: Replay bypasses dedup, silence, and routing rules accidentally
**What goes wrong:** Manual replay enters the wrong point in the pipeline: either too early and re-triggers grouping/dedup/silence unexpectedly, or too late and bypasses safety checks that should still apply.
**Why it happens:** Existing systems already have routing and bounded retry, so replay is added as "just call send again" without deciding whether replay means "repeat the original delivery" or "re-run policy evaluation."
**Consequences:** Duplicate notifications, silently skipped replays, or policy drift between original and replayed sends.
**Warning signs:** No written replay semantics; code paths mix "resend exact attempt" and "re-evaluate current config."
**Prevention:** Define two explicit modes if needed, but default roadmap should prefer one safe mode first: `replay original delivery intent`. Make "re-evaluate current rules" a later explicit feature, not an accidental side effect.
**Absorb in roadmap:** **Phase 1** for semantics, **Phase 2 - Replay workflow and operator tooling** for enforcement.

### Pitfall 3: Recovery APIs become a privileged side-effect backdoor
**What goes wrong:** Replay/retry endpoints are exposed to operators without hard capability checks, reason capture, or audit logging because the team treats them as "ops convenience."
**Why it happens:** Brownfield systems often already have alert action endpoints, and replay looks similar enough that teams copy an existing handler but forget that replay can trigger external messages again.
**Consequences:** Unauthorized or unexplained outbound notifications, weak incident accountability, and a new abuse path more dangerous than normal alert ack/silence actions.
**Warning signs:** Frontend hides buttons, but backend lacks dedicated capability; replay requests do not require actor/reason/target metadata; no audit link between manual action and resulting attempts.
**Prevention:** Add a dedicated replay capability, mandatory audit record, and operator-visible reason field before exposing any replay button or API to the UI.
**Absorb in roadmap:** **Phase 2 - Replay workflow and operator tooling.**

### Pitfall 4: Persistence is added only on failure paths
**What goes wrong:** Only failed deliveries get stored because the initial driver is "we need replay for failures."
**Why it happens:** This seems cheaper at first, but it produces an incomplete operational history and makes success-rate metrics impossible to trust.
**Consequences:** Operators cannot compare failed vs successful attempts, health surfaces lie, and replay tooling lacks context such as prior success after bounded retry.
**Warning signs:** Schema or handler names are failure-centric (`failed_notifications`, `retry_queue`) and successful attempts remain log-only.
**Prevention:** Persist all delivery attempts with normalized statuses. Failures are a filtered view, not a separate truth store.
**Absorb in roadmap:** **Phase 1.**

### Pitfall 5: Backfill or dual-write rollout locks the hot path
**What goes wrong:** Adding delivery persistence or history indexes causes long migrations, table rewrites, or slow synchronous writes on the live webhook/notification path.
**Why it happens:** Teams underestimate write amplification and add wide JSON columns, secondary indexes, and backfill jobs in one cutover.
**Consequences:** Higher ingest latency, slower notification fanout, lock contention, or rollback pressure during deployment.
**Warning signs:** Migration plan requires rewriting old rows before code rollout; hot tables get multiple new indexes immediately; replay/history queries depend on fresh secondary indexes from day one.
**Prevention:** Roll out additive schema first, then dual-write, then backfill selectively if truly needed. Keep first release query model narrow and index only access paths required for early ops workflows.
**Absorb in roadmap:** **Phase 1** for schema shape, **Phase 5 - Rollout, migration, and production validation** for cutover/backfill.

### Pitfall 6: Signature verification happens after body mutation
**What goes wrong:** Ingress signature checks fail intermittently or are disabled because verification runs after JSON binding, decompression, body reuse, or other request mutations.
**Why it happens:** Webhook handlers often bind directly into structs; Gin request bodies are consumable, and providers such as Stripe require verification against the raw body.
**Consequences:** Teams either reject valid traffic or silently fall back to insecure "accept without verification."
**Warning signs:** Handler calls `ShouldBindJSON` before verification; code reconstructs JSON string for HMAC; raw body is unavailable on signature failure debugging.
**Prevention:** Capture raw body once, verify first, then bind/parse. Keep provider-specific verification middleware close to ingress, not hidden deep in business logic.
**Absorb in roadmap:** **Phase 3 - Ingress protection and readiness.**

### Pitfall 7: Request-size and rate limits are bolted on with the wrong scope
**What goes wrong:** Limits are added globally or per-process with no source awareness, throttling trusted senders unpredictably, starving internal health routes, or behaving inconsistently across replicas.
**Why it happens:** "Add rate limiting" sounds simple, but brownfield systems usually have mixed traffic classes: public webhook ingress, authenticated UI/API traffic, health probes, and WebSocket endpoints.
**Consequences:** False positives during incident spikes, poor multi-instance behavior, and operators disabling protections because they break normal traffic.
**Warning signs:** Same limiter middleware sits on all routes; health/readiness gets limited; 429s spike during genuine alert storms.
**Prevention:** Separate policy by route class. Start with hard body-size caps and low-risk ingress-specific controls, then add source-aware rate limiting with explicit observability before tightening thresholds.
**Absorb in roadmap:** **Phase 3.**

### Pitfall 8: Readiness is confused with liveness or "process is up"
**What goes wrong:** The service reports ready even when PostgreSQL/Redis dependencies required for alert ingestion or notification recording are unavailable, or it reports unhealthy because a non-critical dependency blips.
**Why it happens:** Teams reuse a single `/health` endpoint for everything, or make readiness probe do deep expensive checks on every hit.
**Consequences:** Traffic is sent to instances that cannot complete core writes, or pods flap during transient dependency issues.
**Warning signs:** One endpoint serves human health, liveness, and readiness; no distinction between must-have and degraded dependencies; readiness does blocking deep checks every request.
**Prevention:** Split liveness and readiness semantics. Liveness should stay cheap and process-oriented. Readiness should reflect ability to complete the core alert path, with bounded dependency checks and clearly defined degraded states.
**Absorb in roadmap:** **Phase 3.**

### Pitfall 9: Metrics explode in cardinality and still miss the real questions
**What goes wrong:** New metrics label by `alert_id`, `trace_id`, raw channel name, user name, or error text, producing expensive high-cardinality series while still not answering "where are failures accumulating?"
**Why it happens:** Teams mirror log fields into metrics instead of defining a small contract around route, channel type, result class, and stage.
**Consequences:** Costly metrics, slow queries, and dashboards that still need logs for basic diagnosis.
**Warning signs:** Per-alert or per-trace labels; metrics created only on failure; dashboards rely on ad hoc label regex.
**Prevention:** Define the metrics contract before instrumenting. Use counters for attempts/results, histograms for send latency, gauges/timestamps for backlog or last-success indicators, and keep labels bounded.
**Absorb in roadmap:** **Phase 4 - Metrics and operations visibility.**

### Pitfall 10: Ops history UI ships before status semantics are stable
**What goes wrong:** The history page becomes a thin query over inconsistent backend states: alert state, delivery state, attempt state, replay state, and health state are mixed without a canonical model.
**Why it happens:** UI demand arrives early, but the backend has not yet settled on durable enums, pagination keys, filtering rules, or which timestamps matter.
**Consequences:** Operators lose trust in the page, frontend work gets redone, and requirements drift from "history" into "debug every edge case in one screen."
**Warning signs:** Mock fields like `status` or `result` mean different things across endpoints; frontend needs custom interpretation logic per row type.
**Prevention:** Finalize backend status taxonomy and query shapes before building the broad operations surface. Ship a narrow history API first, then add richer views.
**Absorb in roadmap:** **Phase 4.**

## Moderate Pitfalls

### Pitfall 11: Replay work extends the synchronous ingress path
**What goes wrong:** Delivery record creation, heavy serialization, or replay scheduling gets inserted into the request thread for inbound webhooks and pushes latency up.
**Prevention:** Keep ingress response fast and move heavy recovery work asynchronous. This aligns with common webhook guidance to acknowledge quickly and process asynchronously.
**Absorb in roadmap:** **Phase 1** for write-path design, **Phase 3** for ingress SLA validation.

### Pitfall 12: "Health surface" turns into an unbounded debug endpoint
**What goes wrong:** Operators ask for payload previews, last errors, traces, channel bodies, and config diffs all in one page/API. Queries become expensive and sensitive data spreads.
**Prevention:** Separate summary health from deep inspection. First release should expose counts, timestamps, and drill-down identifiers; detailed payload access should stay gated and intentionally scoped.
**Absorb in roadmap:** **Phase 4.**

### Pitfall 13: Failure taxonomy is too free-form to automate
**What goes wrong:** Errors are stored as arbitrary strings only, so product and ops cannot distinguish retryable network failures from permanent configuration errors.
**Prevention:** Add normalized failure classes alongside raw error detail from the first schema version.
**Absorb in roadmap:** **Phase 1.**

## Minor Pitfalls

### Pitfall 14: Metrics start only after the first failure
**What goes wrong:** Dashboards and alerts look empty in healthy periods because the system exports no zero/default series or no last-success timestamp.
**Prevention:** Predeclare core counters/gauges where practical and expose timestamps for "last success" / "last failure" style checks.
**Absorb in roadmap:** **Phase 4.**

### Pitfall 15: Replay tooling is built as bulk-first
**What goes wrong:** The first operator workflow tries to support batch replay, filtering, and mass actions before single-record safety is proven.
**Prevention:** Start with single-delivery replay plus clear confirmation and audit. Bulk replay should be a later milestone item.
**Absorb in roadmap:** **Phase 2.**

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Phase 1: Delivery record contract and persistence | Team stores only failures or mutable references | Persist all attempts with immutable delivery intent, normalized status, failure class, and attempt metadata |
| Phase 1: Delivery record contract and persistence | Schema/index rollout slows hot path | Use additive migration + dual-write first; defer backfill/index expansion until access patterns are proven |
| Phase 2: Replay workflow and operator tooling | Replay becomes an un-audited privileged endpoint | Add dedicated capability, actor reason, audit linkage, and single-record workflow before UI exposure |
| Phase 2: Replay workflow and operator tooling | Replay semantics drift between "resend" and "re-evaluate" | Freeze v1 behavior as replay of original delivery intent; defer policy re-evaluation |
| Phase 3: Ingress protection and readiness | Signature verification breaks after body parsing | Verify against raw body before JSON bind; keep raw body available for diagnostics |
| Phase 3: Ingress protection and readiness | Rate limit harms healthy traffic during alert storms | Scope controls per route/source class and instrument 413/429 outcomes before tightening |
| Phase 3: Ingress protection and readiness | Readiness says healthy when core dependencies are unavailable | Separate cheap liveness from core-path readiness with bounded dependency checks |
| Phase 4: Metrics and operations visibility | Metrics have high-cardinality labels and weak semantics | Define bounded labels and operator questions first; instrument to answer them directly |
| Phase 4: Metrics and operations visibility | History/health page becomes a misleading join of unstable statuses | Stabilize enums, filters, and timestamps in API contract before broad frontend surface |
| Phase 5: Rollout, migration, and production validation | Brownfield cutover hides regressions until production load | Validate dual-write, replay audit trail, readiness behavior, and metrics under failure injection before declaring done |

## Roadmap Prevention

Recommended sequencing so the roadmap does not bury production concerns:

1. **Phase 1 - Delivery record contract and persistence**
   Lock the data model first: immutable delivery intent, attempt rows, status enums, failure classes, and additive schema rollout. Do not start with UI.

2. **Phase 2 - Replay workflow and operator tooling**
   Add secured replay only after the ledger exists. Keep scope to single-record replay, explicit actor intent, and end-to-end auditability.

3. **Phase 3 - Ingress protection and readiness**
   Add raw-body signature verification, body-size caps, route-class-aware limiting, and proper readiness/liveness semantics before broadening ingress trust.

4. **Phase 4 - Metrics and operations visibility**
   Instrument against a stable contract, then expose a narrow health/history surface that answers specific operator questions instead of becoming a debug dump.

5. **Phase 5 - Rollout, migration, and production validation**
   Reserve explicit work for dual-write burn-in, backfill decisions, chaos/failure tests, and rollout checks. Do not hide this under "final polish."

## Sources

- Prometheus instrumentation best practices: labels, counters/gauges, timestamps, and missing-metric guidance. HIGH confidence. https://prometheus.io/docs/practices/instrumentation/
- Prometheus histogram guidance for latency metrics. HIGH confidence. https://prometheus.io/docs/practices/histograms/
- Prometheus Alertmanager concepts: deduplication, grouping, silences, and routing. HIGH confidence. https://prometheus.io/docs/alerting/latest/alertmanager/
- Kubernetes readiness/liveness/startup probe behavior and separation of concerns. HIGH confidence. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
- Stripe webhook docs: automatic retries, manual retries, duplicate delivery handling, async processing, and signature verification against raw body. HIGH confidence. https://docs.stripe.com/webhooks/signatures and https://docs.stripe.com/webhooks/signature
- Gin request-body limitations and body-consumption behavior. MEDIUM confidence for direct applicability to this codebase; official framework docs are current, but final middleware design remains project-specific. https://gin-gonic.com/zh-cn/docs/routing/upload-file/limit-bytes/ and https://gin-gonic.com/en/docs/examples/bind-body-into-dirrerent-structs/

## Confidence Notes

- **HIGH:** Webhook signature/raw-body handling, retry/duplicate expectations, probe semantics, and Prometheus metric-shape guidance are directly supported by official docs.
- **MEDIUM:** Brownfield rollout, replay semantics, audit boundary, and schema sequencing recommendations are design inferences from the current repository plus common alerting-system failure modes.
