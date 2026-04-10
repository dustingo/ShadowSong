# Phase 04: Enable Raw Event Passthrough In Notification Templates - Context

**Gathered:** 2026-04-10
**Status:** Ready for planning
**Source:** Post-milestone gap capture from user discussion

<domain>
## Phase Boundary

This phase extends the webhook/template system so notification output templates can access both:

- the existing standardized alert fields (`alert_name`, `severity`, `message`, `source`, `status`, `trigger_time`, etc.)
- the original inbound webhook payload, including arbitrary nested JSON fields

The phase must keep the current alert ingestion, routing, and notification path working for existing data sources and channels.
</domain>

<decisions>
## Implementation Decisions

### Locked Decisions

- The user requirement is “自由透传任意原始 JSON 字段到最终通知模板”.
- The implementation must be user-friendly, not a hidden internal-only capability.
- Existing standard-field templates must remain compatible after the change.
- The system should not force users to understand backend `Alert` internals just to render notifications.
- The output-template context should expose both standardized fields and raw-event data in a stable way.

### the agent's Discretion

- Whether raw payload access is provided as `raw`, `event`, `payload`, or a similar top-level template variable.
- Whether to preserve additional derived context besides raw payload, such as normalized annotations or selected extracted values.
- Whether template helper functions need to be expanded for nested lookups and safe defaults.
- Whether a UI-side template test/example improvement is needed in the same phase or as a tightly coupled follow-up plan inside the phase.
</decisions>

<specifics>
## Specific Ideas

- Current behavior: `input_template` can shape inbound JSON into the internal alert model, but `output_template` currently receives only a narrow render context.
- Current pain point from user testing: fields like top-level `summary`, `description`, `value`, nested `annotations.runbook`, or custom payload fields cannot be referenced freely in notification templates unless they are first collapsed into the standard alert model.
- Target experience: users can keep writing standard fields like `{{.alert_name}}`, while also being able to reference raw-event data through a clear, documented context path.
</specifics>

<canonical_refs>
## Canonical References

### Webhook Rendering
- `internal/handlers/webhook.go` — inbound template rendering, alert normalization, notification render context, and datasource default templates

### Notification Senders
- `internal/notifier/notifier.go` — downstream notification sender contracts and channel config expectations

### Data Models
- `internal/models/models.go` — datasource model containing `input_template` and `output_template`
- `internal/models/alert.go` — standardized persisted alert model

### Frontend Configuration UI
- `frontend/src/pages/DataSources.tsx` — datasource editing and template entry UX
- `frontend/src/api/client.ts` — datasource/config APIs
- `frontend/src/stores/configStore.ts` — datasource state flow

### Recent Product Context
- `.planning/PROJECT.md`
- `.planning/ROADMAP.md`
</canonical_refs>

<deferred>
## Deferred Ideas

- General-purpose rule-engine redesign
- Full notification-template DSL redesign beyond what is required for raw payload passthrough
- Unrelated frontend config UX cleanup outside datasource template editing/testing
</deferred>

---

*Phase: 04-enable-raw-event-passthrough-in-notification-templates*
*Context gathered: 2026-04-10*
