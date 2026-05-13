---
name: phase-13-01
description: Harden async notification path with panic recovery and contextual failure handling
metadata:
  type: spec
  source_phase: 13-harden-notification-delivery-path
  source_plan: "01"
  milestone: v1.2
  status: completed
  completed: 2026-04-21
---

# Phase 13 Plan 01: Notification Path Hardening

## Context & Goals

Ensure webhook-triggered notification goroutines fail safely, preserve the current alert pipeline semantics, and emit enough context to debug failure points without redesigning the architecture.

Purpose: Ensure webhook-triggered notification goroutines fail safely, preserve the current alert pipeline semantics, and emit enough context to debug failure points without redesigning the architecture.
Output: A panic-protected notification execution boundary plus contextual failure handling around template rendering and channel delivery.

## Success Criteria

- Async notification processing cannot crash the process path when a notification goroutine panics
- Notification send boundaries preserve alert and channel context through failures
- Existing webhook ingest, routing, and send flow semantics remain intact after hardening

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| panic-protected async notification boundary and contextual logging path | `internal/handlers/webhook.go` | panic-protected async notification boundary and contextual logging path |
| notification send failure contract usable by handler-level trace logging | `internal/notifier/notifier.go` | notification send failure contract usable by handler-level trace logging |

## Architecture

### Key Architectural Decisions

- Panic-recover boundary for notification processing that logs failure context instead of letting goroutine die silently
- Send path explicit about stage boundaries: route lookup, datasource/template resolution, render fallback, and channel delivery
- Minimal injection seam or helper functions for testability without architecture redesign
- Notifier errors suitable for handler-level trace logging rather than opaque failures

## Implementation Tasks

### Task 1: Add panic recovery and explicit failure boundaries around async notification processing

**Files:** `internal/handlers/webhook.go`, `internal/notifier/notifier.go`

**Acceptance Criteria:**
- `internal/handlers/webhook.go` contains a panic recovery boundary for async notification processing
- notification failures can be logged with alert and channel context at the handler boundary
- existing notification routing and default-fallback behavior remain present

**Action:** Refactor the async notification entrypoint in `internal/handlers/webhook.go` so notification processing executes behind a panic-recover boundary that logs the failure context instead of letting the goroutine die silently. Keep the existing webhook ingest, route matching, template rendering, and send semantics, but make the send path explicit about stage boundaries such as route lookup, datasource/template resolution, render fallback, and channel delivery. If needed, introduce a minimal injection seam or helper functions so panic and send-failure paths can be tested without redesigning the handler architecture. In `internal/notifier/notifier.go`, preserve the sender API but ensure downstream errors are suitable for handler-level trace logging rather than opaque failures.

**Verification:** `go test ./internal/handlers ./internal/notifier`

## Security Considerations

None

## Established Patterns

None

## Decisions

- Panic-recover boundary instead of letting goroutine die silently
- Explicit stage boundaries for failure traceability
- Preserved existing webhook, route matching, template rendering, and send semantics

## Deviation Log

None
