# Phase 18: Establish Delivery Ledger - Context

**Gathered:** 2026-04-29
**Status:** Ready for planning

<domain>
## Phase Boundary

为现有通知发送链路建立 PostgreSQL 持久化投递账本，覆盖每条通知的投递实体、attempt 历史、最终结果与审计所需不可变快照，并提供满足维护者查看单条账本记录的最小读能力。该 phase 不负责完整人工恢复操作、批量恢复、完整历史页或复杂运维健康面。

</domain>

<decisions>
## Implementation Decisions

### Ledger Data Model
- **D-01:** 账本采用两层模型：`notification_deliveries` 作为每条 `alert x channel` 的主记录，`notification_delivery_attempts` 作为 append-only attempt 明细。
- **D-02:** 必须持久化成功与失败两类投递，而不是只存最终失败；失败视图只是账本上的一个筛选结果，不单独建“失败表”真源。
- **D-03:** 现有应用内 bounded 3 次即时重试继续保留，Phase 18 先在发送热路径旁路落库，不引入 MQ、外部 worker 或 workflow engine。

### Snapshot Contract
- **D-04:** 每条 delivery 记录冻结审计与原始重试所需的不可变快照，至少包括：告警关键字段、渠道身份快照、路由身份快照、发送模式、最终实际发送内容，以及终态失败摘要。
- **D-05:** 首版快照优先保证“可审计、可解释、可支撑原始 retry”，不把完整渠道密钥、全量运行时配置或重型原始回执作为必存范围。

### Recovery Semantics Boundary
- **D-06:** `retry` 语义锁定为沿原始发送语义再次执行，基于账本记录与冻结快照完成，不重新走当前策略。
- **D-07:** `replay` 语义锁定为重新走当前策略，而不是复用原始 route/template 结果；因此 `replay` 的安全动作与完整 API/审计闭环继续属于 Phase 19。
- **D-08:** Phase 18 只需要把账本设计成足以同时支撑上述两种未来语义，不在本 phase 内开放人工恢复入口。

### Read Surface
- **D-09:** 为满足“维护者可以查看任一通知投递账本记录”，Phase 18 提供最小只读能力即可，优先是后端详情/查询 API 或等价只读入口，不强制在本 phase 同时交付完整前端历史页。
- **D-10:** 完整历史列表、搜索筛选、运维健康聚合页和人工恢复操作继续放在 Phase 19/21，避免 Phase 18 与后续 phase 边界重叠。

### the agent's Discretion
- 具体表字段命名、状态枚举细节、最小索引集合和 GORM 迁移组织方式可由后续 researcher/planner 在不违背上述契约的前提下决定。
- 最小读能力是先做单条详情 API 还是带有限过滤的只读列表 API，可按最小满足 success criteria 与改动面控制原则裁定。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase Scope And Requirements
- `.planning/ROADMAP.md` — Phase 18 milestone boundary, success criteria, and dependency split with Phases 19-21.
- `.planning/REQUIREMENTS.md` — `DELV-01`, `DELV-02`, and `DELV-06` requirement truth for persistent delivery records, durable final failures, and immutable replay-supporting snapshots.
- `.planning/PROJECT.md` — brownfield constraints, no-migration posture, and current v1.4 milestone intent.
- `.planning/STATE.md` — accumulated milestone context, especially single-item recovery direction and immutable snapshot expectation.

### Existing Runtime Truth
- `docs/alert-path-operations-runbook.md` — current verified alert-path guarantees for `trace_id`, `send_attempt`, `send_notification`, and `terminal_failure` that Phase 18 must preserve.
- `internal/handlers/webhook.go` — current webhook -> route -> notify path, bounded retry loop, and notification logging contract that the ledger must integrate with.
- `internal/handlers/webhook_test.go` — verified retry-boundary and `terminal_failure` behavior that Phase 18 must keep green while adding persistence.
- `internal/database/postgres.go` — current migration/bootstrap entrypoint where additive ledger schema should be introduced.
- `internal/models/models.go` — existing model/style conventions and current persistent domain boundaries.

### v1.4 Research
- `.planning/research/SUMMARY.md` — recommended phase ordering, ledger-first rollout, and deferred-scope guidance for retry/replay and ops surfaces.
- `.planning/research/ARCHITECTURE.md` — proposed delivery/attempt model, integration seam in `WebhookHandler`, and read-surface split across later phases.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/handlers/webhook.go`: already owns route match, notification send entry, trace propagation, and bounded retry loop; this is the natural ledger integration seam.
- `internal/database/postgres.go`: already performs additive GORM migrations table-by-table, so new ledger tables can be introduced without a new migration framework.
- `internal/models/alert.go` and `internal/models/models.go`: establish current model validation/default patterns and singular domain noun naming.
- `frontend/src/pages/*.tsx` plus existing Ant Design table/drawer patterns: reusable later for delivery history/detail surfaces, but not required to be fully consumed in Phase 18.

### Established Patterns
- Backend request logic is handler-centric with focused helper methods rather than a global service layer; introducing one narrow `delivery` domain service is acceptable, but a whole-repo refactor is not.
- Persistent truth belongs in PostgreSQL via GORM, while Redis remains ephemeral transport/state; delivery history should follow the same boundary.
- Current observability truth is a stable `key=value` log contract with shared `trace_id` and stage names; DB persistence must augment this, not replace it.
- Brownfield phases prefer additive schema and incremental hot-path changes backed by regression tests instead of broad architecture migration.

### Integration Points
- Notification ledger creation/update should hook around `processAlertNotifications`, `sendNotification`, and `sendChannelNotification`.
- Attempt/result persistence should share identifiers already present in logs: `trace_id`, `alert_id`, `channel_id`, and send mode.
- Minimal read capability should attach to existing authenticated Gin API groups and reuse current auth/capability middleware rather than introducing a new access model.

</code_context>

<specifics>
## Specific Ideas

- `retry` and `replay` intentionally have different semantics:
  - `retry` = 原始发送语义
  - `replay` = 当前策略重新执行
- Phase 18 should therefore optimize snapshots for auditability and deterministic retry, while leaving full replay execution flow to the next phase.
- The ledger should remain the recovery source of truth even though current alert-path logs stay searchable and verified.

</specifics>

<deferred>
## Deferred Ideas

- 完整 delivery history 页面、复杂搜索筛选和运维健康视图 — 留给 Phase 19/21。
- 人工 `retry/replay` API、原因录入、操作者审计闭环 — 留给 Phase 19。
- 批量恢复、复杂 replay preview、更多策略重算模式 — 继续 deferred，不纳入本 phase。

</deferred>

---

*Phase: 18-establish-delivery-ledger*
*Context gathered: 2026-04-29*
