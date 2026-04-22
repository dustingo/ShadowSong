# Phase 16: Standardize Alert Path Logging - Research

**Researched:** 2026-04-22
**Domain:** Backend alert-path logging conventions and operational observability
**Confidence:** HIGH

<user_constraints>
## User Constraints

No `16-CONTEXT.md` exists yet, so the planner should treat the following as the current planning constraints inferred from roadmap, requirements, project docs, and the completed Phase 14/15 artifacts. [VERIFIED: `.planning/phases/16-standardize-alert-path-logging` directory has no `16-CONTEXT.md`] [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/PROJECT.md`]

### Locked Decisions (inferred from current truth)

- Keep the existing Go + Gin + GORM + PostgreSQL + Redis + React + Vite stack; this phase is not a logging-platform migration. [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/PROJECT.md`]
- Stay brownfield and preserve the current alert ingest, routing, silence, on-duty, and notification flows while standardizing logging. [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/PROJECT.md`]
- Build on Phase 14 and Phase 15 instead of redefining them: `trace_id` is already the correlation truth, and retry logs already rely on `attempt`, `max_attempts`, and terminal-failure stages. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`]
- Do not introduce external observability infrastructure such as OpenTelemetry, Prometheus, or a centralized logging platform to satisfy this phase. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/PROJECT.md`]
- The output must leave behind reusable proof of the logging contract in tests or verification docs, not just implementation code. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/ROADMAP.md`]

### Claude's Discretion

- The exact helper shape may be a single `webhook.go` event writer or a tiny local package such as `internal/observability/alertlog.go`, as long as the call sites stay close to the current handler flow. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]
- The planner may keep the existing `log.Logger` seam or wrap it very lightly, but should not widen scope into a repo-wide logger rewrite. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [ASSUMED]
- Event-specific optional fields can be expanded beyond the Phase 15 minimum set if they improve consistency, provided the naming stays stable and machine-readable. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-CONTEXT.md`] [ASSUMED]

### Deferred Ideas (OUT OF SCOPE)

- Repo-wide logging unification across unrelated packages such as all WebSocket, bootstrap, and frontend code in the same phase. [VERIFIED: `.planning/ROADMAP.md`] [ASSUMED]
- Switching the entire backend to a new structured logging framework as a prerequisite for Phase 16. [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/PROJECT.md`] [ASSUMED]
- Introducing external log storage, metrics, traces, dashboards, or alerting infrastructure. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `.planning/PROJECT.md`]
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OBS-03 | 运维排障时可以依据统一关联字段，从一条失败通知回溯到对应告警接入与处理上下文。 [VERIFIED: `.planning/REQUIREMENTS.md`] | Keep `trace_id` and existing stage names stable, then remove field drift so failure logs, route-match logs, Redis handoff logs, and ingest/persist logs can all be searched with the same schema. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| LOG-01 | 告警主链路的后端日志使用统一字段命名和输出格式，避免同类事件字段漂移。 [VERIFIED: `.planning/REQUIREMENTS.md`] | Replace the split `logNotification` / `logTraceStage` contract with one canonical event writer and one shared field envelope for alert-path events. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| LOG-02 | 高风险链路中的临时 `fmt.Print*` 风格日志需要收口到统一日志入口，减少混杂输出。 [VERIFIED: `.planning/REQUIREMENTS.md`] | Literal `fmt.Print*` matches are absent in current runtime alert-path code, but the same operational risk remains in direct ad hoc `Printf`-style logging and duplicated field assembly; Phase 16 should scope LOG-02 to those print-style sites in the alert path. [VERIFIED: `rg -n \"fmt\\.Print|fmt\\.Printf|fmt\\.Println\" internal cmd .planning/phases/14-establish-alert-trace-context .planning/phases/15-harden-notification-retry-boundaries`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| LOG-03 | 新的日志约定需要在测试或验证文档中有明确样例，避免后续继续各写各的。 [VERIFIED: `.planning/REQUIREMENTS.md`] | Extend `internal/handlers/webhook_test.go` with contract-focused assertions and record the final field conventions in a phase verification artifact. [VERIFIED: `internal/handlers/webhook_test.go`] [VERIFIED: `.planning/ROADMAP.md`] [ASSUMED] |
</phase_requirements>

## Summary

Phase 16 should be planned as a brownfield contract pass over the existing webhook-to-notification path, not as a new logging architecture. The current code already has one injectable logger seam in `WebhookHandler`, and Phase 14/15 already established stable `stage` names plus `trace_id` and retry-specific fields. The real problem is that the schema is split across two helpers, some machine-readable values still live in free-text messages, and field sets drift depending on which helper a caller used. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`]

The cleanest planning target is one shared alert-path event writer that preserves the Phase 14/15 taxonomy, always emits a canonical base field set, and accepts event-specific optional fields in a deterministic order. That approach keeps tests stable, avoids new dependencies, and directly addresses `LOG-01/02/03` plus `OBS-03`. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `internal/handlers/webhook.go`] [CITED: https://pkg.go.dev/log] [ASSUMED]

One subtle but important finding is that literal `fmt.Print*` runtime usage is not the current problem in the alert path; the repo has already converged away from that. The real planning risk is ad hoc `Printf` logging with duplicated field assembly and machine data embedded in the message string. If the plan chases only literal `fmt.Print*`, it will under-scope the actual consistency work. [VERIFIED: `rg -n \"fmt\\.Print|fmt\\.Printf|fmt\\.Println\" internal cmd .planning/phases/14-establish-alert-trace-context .planning/phases/15-harden-notification-retry-boundaries`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Primary recommendation:** Keep the current `WebhookHandler` logger seam, replace the dual helper contract with one canonical alert-path event writer, preserve `stage` and `trace_id` from Phases 14/15, and lock the new schema with handler tests plus a short verification document. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [ASSUMED]

## Project Constraints (from AGENTS.md)

- Keep the current tech stack; do not recommend a platform or framework migration. [VERIFIED: `AGENTS.md`]
- Respect brownfield structure and any unrelated local changes. [VERIFIED: `AGENTS.md`]
- Do not break the core alert flow while changing logging behavior. [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/PROJECT.md`]
- Keep documentation and verification aligned with the non-AI alert-system truth. [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/PROJECT.md`]

## Standard Stack

### Core

| Library / Tool | Version | Purpose | Why Standard | Source |
|----------------|---------|---------|--------------|--------|
| Go `log` package | stdlib in Go 1.25.0 | Existing runtime logger seam for alert-path operational events. | `WebhookHandler` already injects `*log.Logger`, tests capture it directly, and the standard logger serializes writes across goroutines without adding dependency churn. | [VERIFIED: `go.mod`] [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [CITED: https://pkg.go.dev/log] |
| `WebhookHandler` local event seam | repo-local | The actual alert-path integration point for ingest, dedup, persist, Redis publish, route match, notification entry, send attempts, and terminal failure. | Phase 14 and 15 already centralized alert-path lifecycle and retry logging here, so Phase 16 can standardize the contract without moving the workflow elsewhere. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] |
| Testify + `go test` | Testify v1.11.1; `go test` from Go 1.25.0 | Lock logging-field and stage-contract behavior in backend tests. | The repo already uses focused handler tests with an in-memory logger buffer, and `go test ./internal/handlers` is green locally. | [VERIFIED: `go.mod`] [VERIFIED: `internal/handlers/webhook_test.go`] [VERIFIED: `go test ./internal/handlers -count=1`] |

### Supporting

| Library / Tool | Version | Purpose | When to Use | Source |
|----------------|---------|---------|-------------|--------|
| Go `log/slog` package | stdlib in Go 1.25.0 | Optional future bridge for structured attrs, grouped fields, and eventual handler-based formatting. | Use only if the planner needs a very small adapter for the existing alert-path seam; do not make a repo-wide slog migration the prerequisite for Phase 16. | [VERIFIED: `go.mod`] [CITED: https://pkg.go.dev/log/slog] [ASSUMED] |
| Existing phase verification docs | repo-local | Reusable format for documenting the final field contract and troubleshooting path. | Use a Phase 16 verification artifact to record sample log lines and the required search path from terminal failure back to ingest. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff | Source |
|------------|-----------|----------|--------|
| Keep `*log.Logger` seam and standardize one text `key=value` event writer | Immediate `log/slog` migration for the whole alert path | `slog` is a real structured-logging option and can emit text or JSON, but adopting it here would widen scope because current constructors and tests inject `*log.Logger`, and wrapper functions can also distort source attribution if handled poorly. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [CITED: https://pkg.go.dev/log/slog] |
| Preserve current Phase 14/15 `stage` taxonomy | Introduce a new top-level event key such as `event` and rename existing stages | Renaming all stage labels would create unnecessary churn in tests and prior verification docs; keeping `stage` stable while standardizing the surrounding fields is lower risk. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [ASSUMED] |
| Phase-local alert-path standardization | Whole-repo logging rewrite | A whole-repo pass would absorb WebSocket, bootstrap, and unrelated packages, which is larger than the roadmap’s “alert main path” goal for this phase. | [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/PROJECT.md`] [ASSUMED] |

**Installation:** No new dependency is required for the recommended path. [VERIFIED: `go.mod`] [ASSUMED]

**Version verification:** The recommended path stays on repo-pinned Go/tooling and existing code seams rather than adding new packages. Local environment confirms Go `1.25.0`, Node `22.17.0`, and pnpm `10.28.2` are available. [VERIFIED: `go version`] [VERIFIED: `node --version`] [VERIFIED: `pnpm --version`]

## Architecture Patterns

### Recommended Project Structure

```text
internal/
├── handlers/
│   ├── webhook.go              # Canonical alert-path event emission + call sites
│   └── webhook_test.go         # Contract tests for stages and fields
├── observability/
│   └── alertlog.go             # Optional tiny helper if webhook.go grows too large
.planning/phases/16-standardize-alert-path-logging/
└── 16-VERIFICATION.md          # Sample log lines and troubleshooting contract
```

This keeps the work adjacent to the existing alert path and avoids inventing a new service layer just for logs. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [ASSUMED]

### Pattern 1: One Event Writer, Many Call Sites

**What:** Collapse `logNotification` and `logTraceStage` into one shared event writer that always accepts `stage`, human message text, and a canonical field map or typed event struct. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**When to use:** Use for all alert-path operational events in `HandleWebhook`, `publishToRedis`, `processAlertNotifications`, `sendNotification`, and fallback branches. [VERIFIED: `internal/handlers/webhook.go`]

**Why:** The current split helper design already causes field drift: `logNotification` emits `channel_type`, but `traceFieldsForNotification` does not; some callers sort keys, while others hardcode field order; some event data is only present in the message text. [VERIFIED: `internal/handlers/webhook.go`]

### Pattern 2: Canonical Base Envelope Plus Event-Specific Extras

**What:** Preserve one stable base field envelope, then layer extra keys only when they are real data for that event. Recommended base fields for alert-path logs are `stage`, `trace_id`, `alert_id`, `fingerprint`, and `source`; optional fields include `channel_id`, `channel_name`, `channel_type`, `attempt`, `max_attempts`, `error`, `mode`, `matched_channels`, `redis_stream`, and `redis_message_id`. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [ASSUMED]

**When to use:** Apply the base envelope everywhere an alert object exists, and add optionals only when the branch truly has that context. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Why:** Phase 15 already made operators depend on a stable retry-field minimum, and Phase 14 already made them depend on `trace_id` plus stage continuity. Phase 16 should unify, not replace, those contracts. [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`]

### Pattern 3: Machine Data Goes in Fields, Not the Message

**What:** Move structured values such as `matched_channels` and `mode` into fields instead of embedding them only in `Printf` message text. Keep the free-text message for human interpretation only. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**When to use:** Apply whenever a message currently encodes operational state as `"matched_channels=%d"` or `"mode=%s"` rather than a dedicated field. [VERIFIED: `internal/handlers/webhook.go`]

**Why:** `slog`’s official model treats attributes as the machine-readable part of the record, and the current Phase 14/15 troubleshooting contract depends on field searchability rather than English message parsing. [CITED: https://pkg.go.dev/log/slog] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [ASSUMED]

### Anti-Patterns to Avoid

- **Keeping both helper contracts alive:** Leaving `logNotification` and `logTraceStage` as parallel APIs will preserve the same field drift Phase 16 is meant to remove. [VERIFIED: `internal/handlers/webhook.go`]
- **Changing stage labels gratuitously:** Renaming `ingest`, `persist`, `route_match`, `notification_entry`, `send_attempt`, or `terminal_failure` would break the continuity Phase 14/15 already proved. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`]
- **Treating LOG-02 as only a literal `fmt.Print*` grep cleanup:** Runtime alert-path code has no literal `fmt.Print*` matches today, so a grep-only plan would miss the real duplication problem. [VERIFIED: `rg -n \"fmt\\.Print|fmt\\.Printf|fmt\\.Println\" internal cmd .planning/phases/14-establish-alert-trace-context .planning/phases/15-harden-notification-retry-boundaries`] [ASSUMED]
- **Expanding to whole-repo logging modernization:** That would entangle WebSocket and unrelated packages with no roadmap requirement forcing it. [VERIFIED: `.planning/ROADMAP.md`] [ASSUMED]
- **Logging raw alert payloads or secrets as part of the contract:** The phase needs searchable operational context, not payload duplication or disclosure risk. [VERIFIED: `internal/models/alert.go`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why | Source |
|---------|-------------|-------------|-----|--------|
| Alert-path event emission | Ad hoc `Printf` calls that each choose their own fields | One shared event writer over the existing logger seam | The repo already has one injectable logger seam in `WebhookHandler`; reuse it instead of letting each branch rebuild field formatting. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] |
| Field serialization | Per-call hand-built string fragments with mixed ordering | Deterministic field merge + sorted output | `logTraceStage` already sorts keys, which makes tests and docs stable; Phase 16 should preserve that determinism across all alert-path logs. | [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| Structured logging rollout | A custom JSON encoder or a brand-new external logging dependency | Existing `log.Logger` seam now; optional `slog` only if a tiny adapter is truly needed later | The phase goal is contract consistency, not logging-infrastructure innovation. | [VERIFIED: `AGENTS.md`] [VERIFIED: `.planning/PROJECT.md`] [CITED: https://pkg.go.dev/log/slog] [ASSUMED] |
| Validation proof | Informal manual eyeballing only | Focused handler tests plus a verification doc with sample lines | `LOG-03` explicitly requires reusable proof so future edits do not drift back to one-off logging. | [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `internal/handlers/webhook_test.go`] [ASSUMED] |

**Key insight:** This phase does not need a “better logger” first; it needs a stable alert-path logging contract first. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

## Common Pitfalls

### Pitfall 1: Field Drift Survives Even After You “Refactor Logging”

**What goes wrong:** The code still emits different field sets for similar events because some callers use the old helper path and others use the new one. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Why it happens:** The current alert path already has two helper styles with overlapping but non-identical conventions. [VERIFIED: `internal/handlers/webhook.go`]

**How to avoid:** Make the plan explicitly remove or deprecate one helper path and migrate all Phase 16 call sites onto one event writer. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Warning signs:** `channel_type` appears on some logs but not retry logs; `matched_channels` still only appears in message text. [VERIFIED: `internal/handlers/webhook.go`]

### Pitfall 2: Machine-Readable State Is Hidden in Free Text

**What goes wrong:** Operators can see that something happened, but automated grep or future tooling cannot reliably extract counts, modes, or result metadata. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Why it happens:** Current messages like `"matched_channels=%d"` and `"mode=%s"` keep machine data in the formatted string rather than the field map. [VERIFIED: `internal/handlers/webhook.go`]

**How to avoid:** Make the event writer own both canonical fields and event-specific optionals, then keep message text secondary. [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Warning signs:** Tests assert substrings inside the message body instead of asserting named keys. [VERIFIED: `internal/handlers/webhook_test.go`] [ASSUMED]

### Pitfall 3: Under-Scoping LOG-02 to Literal `fmt.Print*`

**What goes wrong:** The plan reports success because no `fmt.Print*` calls remain, but high-risk operational logging is still duplicated and inconsistent. [VERIFIED: `rg -n \"fmt\\.Print|fmt\\.Printf|fmt\\.Println\" internal cmd .planning/phases/14-establish-alert-trace-context .planning/phases/15-harden-notification-retry-boundaries`] [ASSUMED]

**Why it happens:** The requirement is phrased around `fmt.Print*` style, but the live risk in this repo is ad hoc `Printf` patterns and split field assembly. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**How to avoid:** Treat LOG-02 as “print-style ad hoc operational logs in the alert path,” not just a literal grep exercise. [VERIFIED: `.planning/REQUIREMENTS.md`] [ASSUMED]

**Warning signs:** The diff removes zero runtime logging call sites yet claims the contract is standardized. [ASSUMED]

### Pitfall 4: Over-Scoping Into WebSocket or Global Bootstrap Logging

**What goes wrong:** Phase 16 turns into a general logging cleanup pass and loses momentum on the actual alert path. [VERIFIED: `.planning/ROADMAP.md`] [ASSUMED]

**Why it happens:** There are still direct `log.Printf` calls in `internal/handlers/websocket.go` and `cmd/server/main.go`, which can tempt a broader cleanup. [VERIFIED: `internal/handlers/websocket.go`] [VERIFIED: `cmd/server/main.go`]

**How to avoid:** Keep mandatory scope on the webhook-to-notification path, and only touch other packages if they directly block the alert-path contract. [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/PROJECT.md`] [ASSUMED]

**Warning signs:** The task list starts including WebSocket heartbeat logs or server startup formatting with no direct link to `OBS-03` or `LOG-01/02/03`. [VERIFIED: `.planning/REQUIREMENTS.md`] [ASSUMED]

### Pitfall 5: If You Introduce `slog`, Wrapper Source Info Becomes Wrong

**What goes wrong:** Source file and line attribution points to the wrapper helper, not the real call site. [CITED: https://pkg.go.dev/log/slog]

**Why it happens:** Official `slog` docs warn that naïve wrapper functions can distort source information unless they build records carefully. [CITED: https://pkg.go.dev/log/slog]

**How to avoid:** Prefer a minimal contract helper over a broad `slog` wrapper migration in this phase, or explicitly handle source attribution if the planner chooses `slog`. [CITED: https://pkg.go.dev/log/slog] [ASSUMED]

**Warning signs:** The plan proposes a generic `Infof/Warnf` wrapper layer as the main change instead of alert-path contract cleanup. [ASSUMED]

## Code Examples

Verified patterns adapted to this repo:

### Canonical Alert-Path Event Writer

```go
// Source inspiration:
// - current handler seam in internal/handlers/webhook.go
// - https://pkg.go.dev/log
func (h *WebhookHandler) logAlertEvent(stage string, fields map[string]string, msg string) {
  parts := []string{fmt.Sprintf("stage=%s", stage)}
  keys := make([]string, 0, len(fields))
  for key := range fields {
    if fields[key] == "" {
      continue
    }
    keys = append(keys, key)
  }
  sort.Strings(keys)
  for _, key := range keys {
    parts = append(parts, fmt.Sprintf("%s=%s", key, fields[key]))
  }
  h.notificationLogger().Printf("%s %s", strings.Join(parts, " "), msg)
}
```

This is the smallest change that matches the existing handler/test seam while making all alert-path logs deterministic and field-first. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [CITED: https://pkg.go.dev/log] [ASSUMED]

### Optional `slog` Bridge for a Future Adapter

```go
// Source: https://pkg.go.dev/log/slog
logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
logger = logger.With("trace_id", traceID, "alert_id", alertID)
logger.Info("notification send", "stage", "send_attempt", "attempt", attempt)
```

Official `slog` text output is already `key=value`, and `With` is designed for repeated common attributes. It is a viable future bridge, but still optional for Phase 16 planning. [CITED: https://pkg.go.dev/log/slog] [ASSUMED]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact | Source |
|--------------|------------------|--------------|--------|--------|
| Two helper styles (`logNotification` and `logTraceStage`) with overlapping field sets | One canonical alert-path event writer and one shared field contract | Planned for Phase 16 | Removes field drift and makes new alert-path logs harder to improvise incorrectly. | [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| Machine data partly embedded in message text | Machine data promoted to named fields; message remains human-oriented | Planned for Phase 16 | Improves grepability, docs, and future automation. | [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| Phase 14/15 proved trace and retry observability per event family | Phase 16 should unify those event families under one schema without renaming the stages | Required by current v1.3 roadmap | Preserves continuity while improving consistency. | [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] |

**Deprecated/outdated:**

- Treating literal `fmt.Print*` removal as sufficient proof of logging standardization is outdated for this repo’s current code state. [VERIFIED: `rg -n \"fmt\\.Print|fmt\\.Printf|fmt\\.Println\" internal cmd .planning/phases/14-establish-alert-trace-context .planning/phases/15-harden-notification-retry-boundaries`] [ASSUMED]
- Adding more alert-path logs through whichever helper is closest is no longer acceptable once Phase 16 lands. [VERIFIED: `.planning/REQUIREMENTS.md`] [ASSUMED]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Phase 16 should stay focused on the webhook-to-notification alert path and not include a required WebSocket/bootstrap logging cleanup. | User Constraints, Common Pitfalls | Planner may under-scope cross-package work if the user expects broader coverage. |
| A2 | Preserving `stage` as the canonical event key is preferable to introducing a new top-level event taxonomy. | Summary, Architecture Patterns | Planner may need a wider doc/test migration if a different key is mandated. |
| A3 | A tiny local helper package is acceptable if `webhook.go` would otherwise become too noisy. | User Constraints, Recommended Project Structure | Planner could choose the wrong seam and create avoidable churn. |
| A4 | A text `key=value` contract is the right near-term output format for this phase, rather than line-delimited JSON. | Summary, Code Examples | Planner may need to revisit downstream parsing expectations if JSON is explicitly desired later. |

## Open Questions (resolved for planning)

1. **Should Phase 16 include WebSocket or server bootstrap logging too?**  
   **Resolved for planning:** No, not by default. The roadmap and dependencies point to the alert main path established in Phases 14 and 15, which is centered on `internal/handlers/webhook.go`. Other logging sites are only secondary unless they directly block the alert-path contract. [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `.planning/STATE.md`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

2. **Should the phase change output format to JSON?**  
   **Resolved for planning:** No, not as a requirement. The phase goal is consistency, not a transport change. A deterministic `key=value` text contract fits the current test seam and existing docs while staying low-risk. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [CITED: https://pkg.go.dev/log/slog] [ASSUMED]

3. **Does LOG-02 still matter if there are no live `fmt.Print*` calls?**  
   **Resolved for planning:** Yes. In this repo, the live operational problem is broader “print-style” ad hoc logging, not literal `fmt.Print*`. The plan should explicitly say so to avoid a shallow grep-only implementation. [VERIFIED: `.planning/REQUIREMENTS.md`] [VERIFIED: `rg -n \"fmt\\.Print|fmt\\.Printf|fmt\\.Println\" internal cmd .planning/phases/14-establish-alert-trace-context .planning/phases/15-harden-notification-retry-boundaries`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control | Source |
|---------------|---------|------------------|--------|
| V2 Authentication | no | This phase does not change auth flows; only logging around alert-path operations. | [VERIFIED: `internal/router/router.go`] [VERIFIED: `.planning/ROADMAP.md`] |
| V3 Session Management | no | No JWT/session-state behavior changes are in scope. | [VERIFIED: `internal/router/router.go`] |
| V4 Access Control | no | The phase standardizes logs and does not alter capability checks or route guards. | [VERIFIED: `internal/middleware/auth.go`] [VERIFIED: `.planning/ROADMAP.md`] |
| V5 Input Validation | yes | Continue using server-generated `trace_id` and avoid logging caller-controlled payload fragments as canonical fields. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [ASSUMED] |
| V6 Cryptography | no | No new cryptographic behavior is introduced in this phase. | [VERIFIED: `.planning/ROADMAP.md`] |

### Known Threat Patterns for Go Alert-Path Logging

| Pattern | STRIDE | Standard Mitigation | Source |
|---------|--------|---------------------|--------|
| Correlation ambiguity | Repudiation | Preserve `trace_id` and stable stage names, then standardize surrounding fields so failures can be traced backward consistently. | [VERIFIED: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`] [VERIFIED: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`] [ASSUMED] |
| Raw payload leakage | Information Disclosure | Keep the contract focused on IDs, routing context, and bounded error data rather than full `alert.Raw` payloads. | [VERIFIED: `internal/models/alert.go`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED] |
| Operator misdiagnosis from field drift | Repudiation / Availability | Use one canonical event writer and contract tests so similar events cannot silently diverge in shape. | [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`] [ASSUMED] |

## Sources

### Primary (HIGH confidence)

- `.planning/REQUIREMENTS.md` - `OBS-03`, `LOG-01`, `LOG-02`, `LOG-03` requirements.
- `.planning/ROADMAP.md` - Phase 16 goal, plans, and success criteria.
- `.planning/PROJECT.md` - milestone constraints and v1.3 boundaries.
- `.planning/STATE.md` - current milestone position and Phase 14/15 continuity.
- `internal/handlers/webhook.go` - live alert-path logging helpers, fields, and call sites.
- `internal/handlers/webhook_test.go` - existing test seam and current assertions.
- `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md` - established stage and trace contract.
- `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md` - established retry and terminal-failure contract.
- `go.mod` - repo-pinned Go and dependency versions.
- `Makefile` - current backend test command.
- `go test ./internal/handlers -count=1` - local verification that handler tests are currently green.
- [Go `log` docs](https://pkg.go.dev/log) - standard logger behavior and concurrency guarantees.
- [Go `slog` docs](https://pkg.go.dev/log/slog) - structured logging model, text/json handlers, `With`, and wrapper caveats.

### Secondary (MEDIUM confidence)

- `internal/handlers/websocket.go` - evidence that broader repo logging cleanup exists but is not central to this phase.
- `cmd/server/main.go` - evidence that bootstrap logging exists outside the main alert path.

### Tertiary (LOW confidence)

- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - The recommended path stays on existing repo seams and official Go docs. [VERIFIED: `go.mod`] [VERIFIED: `internal/handlers/webhook.go`] [CITED: https://pkg.go.dev/log]
- Architecture: HIGH - The exact pain points and integration points are directly visible in the live alert-path code and tests. [VERIFIED: `internal/handlers/webhook.go`] [VERIFIED: `internal/handlers/webhook_test.go`]
- Pitfalls: MEDIUM - The failure modes are visible now, but some final scope decisions still depend on whether the user wants broader than webhook-path coverage. [VERIFIED: `.planning/ROADMAP.md`] [VERIFIED: `internal/handlers/webhook.go`] [ASSUMED]

**Research date:** 2026-04-22
**Valid until:** 2026-05-22
