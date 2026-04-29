# Feature Landscape

**Domain:** v1.4 Delivery Recovery and Production Hardening for a brownfield game-ops alert system
**Researched:** 2026-04-29
**Overall confidence:** MEDIUM-HIGH

## Scope Lens

Only covers the new milestone capabilities:
- persistent notification delivery records
- manual replay / retry tooling
- webhook and service-ingress hardening
- operations visibility for failure location and channel health

Existing alert ingest, routing, silences, on-duty, template preview, realtime feed, bounded retries, and trace/log baselines are assumed to already exist and should be reused, not redesigned.

## Table Stakes

Features production users will expect for this milestone. Missing them leaves recovery and hardening incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Persistent delivery record per notification attempt | Production alerting systems need a durable history of what was sent, to which channel, when, and with what result. | Med | Store alert ID, route/channel snapshot, attempt number, terminal status, timestamps, response code, error summary, trace ID. This is the foundation for every recovery and visibility feature. |
| Failure state model with clear retry/replay eligibility | Operators need to distinguish retrying, failed, delivered, and abandoned records before taking action. | Med | Keep the state machine small and auditable. Avoid open-ended workflow states. |
| Manual retry / replay for failed deliveries | Mature systems expose a way to recover failed notifications without fabricating a new alert. | Med | Start with single-record replay and controlled bulk replay by filter. Must be permission-gated and auditable. |
| Audit trail for operator-triggered recovery actions | Manual resend changes production behavior and must be attributable. | Low | Record who retried, when, why, and which delivery record or filter was targeted. |
| Searchable delivery history | Ops teams need to answer “was this ever sent?” and “which channel failed?” quickly. | Med | Filter by alert, channel, status, time range, source, and trace ID. |
| Ingress authentication / authenticity checks | Production webhook endpoints are expected to reject unauthenticated or tampered traffic. | Med | Reuse existing datasource config to add secret/token/signature verification rather than inventing a new trust model. |
| Request body size limits and malformed payload rejection | Ingress should fail fast on oversized or invalid input instead of pushing junk into the alert pipeline. | Low | Return explicit 4xx reasons such as unauthorized, invalid payload, too large, and rate limited. |
| Rate limiting on webhook/service ingress | Protects the system from burst abuse and accidental storm traffic. | Med | Prefer per-source or per-key limits over one global bucket so real sources stay isolated. |
| Readiness endpoint with dependency awareness | Production environments expect a machine-readable signal for whether the service can safely receive traffic. | Low | Differentiate liveness from readiness; readiness should reflect critical dependencies like PostgreSQL and Redis. |
| Delivery and ingress operational metrics | Teams need a fast way to see whether failures are at ingress, routing, or notification delivery. | Med | Minimum set: accepted/rejected ingress counts, retry counts, delivery latency, success/failure by channel, manual replay counts. |
| Channel health / failure summary view | Operators need a summarized view, not just raw logs, to spot broken integrations quickly. | Med | Build from persisted records and existing traces/logs instead of adding a new observability platform. |

## Useful Differentiators

Valuable for operator UX and incident speed, but can be phased after table stakes if scope gets tight.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Replay from filtered history | Lets operators recover a whole outage window or one broken channel without touching healthy records. | Med | Good second step after single-record replay. Needs strong confirmation UX and audit logging. |
| Replay preview using stored payload/template snapshot | Reduces fear of resending the wrong content after template or route config changed. | Med | Strong fit for this repo because template preview already exists. |
| Unified per-alert delivery timeline | Shows ingest, route match, each delivery attempt, and manual recovery in one view. | Med | High operator value, but depends on persistent delivery records first. |
| Rejection reason analytics for ingress | Helps distinguish bad clients from capacity or config issues. | Low | Aggregate 401/403/413/415/429-style outcomes by datasource or source key. |
| Temporary channel disable / maintenance guardrail during incident response | Prevents repeated retries to a known-broken destination while keeping alert ingest alive. | Med | Useful, but only if it fits existing channel model cleanly. |
| Saved operational filters / dashboards | Speeds up triage for common questions like “all Slack failures in the last hour.” | Low | Nice operational UX; not core to recovery correctness. |

## Anti-Features

These should stay out of scope for this milestone.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Queue or broker redesign for notification delivery | Violates the brownfield constraint and turns a recovery milestone into infrastructure migration. | Persist delivery records in PostgreSQL and keep the current in-process pipeline. |
| Fully automatic endless replay daemon | Risks duplicate paging, hides operator intent, and complicates idempotency before the recovery model is proven. | Start with bounded automatic retries plus explicit manual replay. |
| Highly programmable retry DSL or per-channel workflow engine | Too much surface area for one milestone and hard to validate safely. | Keep a fixed retry policy and expose a small set of operator recovery actions. |
| Retaining full raw payload bodies for rejected or unauthorized ingress forever | Creates security and storage risk, especially for bad traffic. | Keep normalized rejection metadata and capped/sanitized samples only when needed. |
| Cross-system incident command center or external observability platform integration | Expands scope from “make current system operable” to “build a broader ops platform.” | Build focused recovery/history views inside the current console. |
| Auto-remediation actions triggered by delivery failure | Couples alert transport failure to risky infrastructure mutation. | Limit this milestone to visibility and safe resend controls. |

## Feature Dependencies

```text
Persistent delivery record -> failure state model
Persistent delivery record -> searchable delivery history
Persistent delivery record -> manual retry / replay
Persistent delivery record -> channel health summary
Persistent delivery record + existing trace_id -> unified per-alert delivery timeline

Ingress authenticity checks -> trustworthy rejection metrics
Body size limits + rate limiting -> ingress hardening baseline
Ingress rejection classification -> rejection reason analytics

Readiness endpoint -> safer load balancer / orchestrator behavior
Metrics + history queries -> operations visibility dashboards
Audit trail -> safe operator replay / retry
```

## MVP Recommendation

Prioritize:
1. Persistent delivery records with a minimal, explicit status model
2. Permissioned manual retry / replay with audit logging
3. Ingress hardening baseline: authenticity checks, size limits, rate limits, readiness
4. Operations visibility: searchable delivery history plus summarized delivery/ingress metrics

Defer:
- Replay preview and filtered bulk replay until the core delivery ledger and permission model are stable
- Unified per-alert timeline until the stored delivery schema is proven adequate
- Saved dashboards/filters unless the UI scope remains small

## Brownfield Recommendations

- Reuse existing `trace_id`, bounded retry logs, and audit logging as the join points for new delivery records instead of inventing parallel identifiers.
- Prefer additive schema changes and new read APIs over rewriting the current notification path.
- Keep replay semantics explicit: replay creates a new delivery attempt linked to an old record, not silent mutation of history.
- Treat ingress hardening as policy + middleware work around existing webhook endpoints, not a new ingest subsystem.

## Sources

- Grafana docs: contact points and notification delivery status, alert monitoring, alert state history, meta-monitoring, webhook HMAC and replay protection
  - https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/
  - https://grafana.com/docs/grafana/latest/alerting/monitor-status/
  - https://grafana.com/docs/grafana/latest/alerting/monitor-status/view-alert-state-history/
  - https://grafana.com/docs/grafana/latest/alerting/set-up/meta-monitoring/
  - https://grafana.com/docs/grafana-cloud/alerting-and-irm/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/
- Prometheus Alertmanager docs: routing, grouping, silencing, HA, alert limits, and metrics baseline
  - https://prometheus.io/docs/alerting/latest/alertmanager/
- PagerDuty docs: Event Orchestration limits and dropped-event caveat, webhook subscriptions, auditability constraints
  - https://support.pagerduty.com/main/docs/event-orchestration
  - https://support.pagerduty.com/main/docs/webhooks
- GitHub webhook docs: failed delivery handling and manual/API redelivery patterns
  - https://docs.github.com/en/webhooks/testing-and-troubleshooting-webhooks/redelivering-webhooks
  - https://docs.github.com/en/enterprise-server%403.19/webhooks/using-webhooks/handling-failed-webhook-deliveries
