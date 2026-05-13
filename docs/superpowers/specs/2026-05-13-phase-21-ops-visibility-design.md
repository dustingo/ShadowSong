---
name: phase-21
description: Ship ops visibility surfaces with metrics, channel health, and ops health page
metadata:
  type: spec
  source_phase: 21-ship-ops-visibility-surfaces
  milestone: v1.4
  status: pending
---

# Phase 21: Ship Ops Visibility Surfaces

## Context & Goals

Phase 20 hardened ingress and runtime. Phase 21 delivers the visibility surfaces operators need to quickly diagnose notification failures and channel health.

**Goal:** Operators can quickly determine notification failure location, channel health status, and key failure evidence through metrics and ledger-aggregated views.

## Success Criteria

1. System exposes key runtime metrics: webhook ingest volume, notification send success rate, retry count, terminal failure count
2. Operators can view each notification channel's recent success rate, failure count, and latest failure reason for quick health assessment
3. First ops health page provides ledger-based aggregated summary view, not complex real-time dashboard
4. `terminal_failure` fallback default content path and `channel_lookup` secondary failure paths have stable regression verification, won't drift due to log text changes

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Metrics endpoint | `internal/handlers/metrics.go` | OPER-03: `/metrics` with key counters |
| Channel health API | `internal/handlers/channel_health.go` | OPER-02: per-channel health summary |
| Ops health page | `frontend/src/pages/OpsHealth.tsx` | OPER-05: ledger-based summary view |
| Terminal failure regression test | `internal/handlers/webhook_test.go` | DEBT-02: fallback content verification |
| Channel lookup regression test | `internal/handlers/webhook_test.go` | DEBT-03: field-level assertions |

## Architecture

### Metrics Endpoint (OPER-03)

**GET /api/v1/metrics** (protected by `CapabilityViewConfig`):
```json
{
  "webhook_ingest_total": 12345,
  "notification_send_success_total": 10000,
  "notification_send_failure_total": 500,
  "notification_retry_total": 1500,
  "notification_terminal_failure_total": 50,
  "period": "24h"
}
```

- Counters from `notification_deliveries` table aggregation
- Period: configurable, default 24h
- No external metrics infrastructure required

### Channel Health API (OPER-02)

**GET /api/v1/channels/:id/health** (protected by `CapabilityViewConfig`):
```json
{
  "channel_id": 1,
  "channel_name": "email-primary",
  "period": "24h",
  "total_deliveries": 1000,
  "successful": 950,
  "failed": 50,
  "success_rate": 0.95,
  "last_failure": {
    "delivery_id": 123,
    "error_message": "SMTP connection timeout",
    "failed_at": "2026-05-13T10:30:00Z"
  }
}
```

- Aggregation from `notification_deliveries` table
- Period: configurable, default 24h
- Single channel detail

### Ops Health Page (OPER-05)

**Frontend: `/ops-health` route:**
- Channel health summary table (success rate, recent failures)
- Recent terminal failures list (from delivery ledger)
- Link to delivery history for drill-down
- No complex real-time charts — ledger-based summary

### Regression Tests (DEBT-02, DEBT-03)

**DEBT-02: terminal_failure fallback content**
- Test that `terminal_failure` log contains expected fields when retry exhausted
- Test fallback default content when datasource template fails
- Field-level assertions, not substring matching

**DEBT-03: channel_lookup secondary failure**
- Test that `channel_lookup` failure path logs expected fields
- Field-level assertions on log output structure

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-21-01 | I | metrics endpoint | mitigate | Protected by `CapabilityViewConfig` |
| T-21-02 | I | channel health | mitigate | Protected by `CapabilityViewConfig` |
| T-21-03 | D | aggregation queries | mitigate | Use indexed queries, limit period |

## Decisions

- Metrics are DB-aggregated, not in-memory counters
- Period default 24h, configurable via query param
- Ops health page is read-only summary, no write actions
- Regression tests use field-level assertions, not substring matching

## Deviation Log

None — spec created from ROADMAP requirements.