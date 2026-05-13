---
name: phase-12-01
description: Add GitHub Actions CI workflow for backend tests and frontend quality gates
metadata:
  type: spec
  source_phase: 12-establish-automated-quality-gates
  source_plan: "01"
  milestone: v1.2
  status: completed
  completed: 2026-04-21
---

# Phase 12 Plan 01: CI Quality Gates

## Context & Goals

Turn the already-validated backend and frontend local checks into a repository-native CI workflow that blocks regressions before merge.

Purpose: Turn the already-validated backend and frontend local checks into a repository-native CI workflow that blocks regressions before merge.
Output: A GitHub Actions workflow that runs backend tests and frontend lint/test/build as explicit, traceable gates.

## Success Criteria

- Repository has an automated CI workflow that runs backend tests plus frontend lint, test, and build
- CI failures are attributable to a named step or job instead of a single opaque script blob
- The workflow reuses the repository's real commands rather than inventing a second build contract

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| GitHub Actions quality gate workflow | `.github/workflows/quality-gates.yml` | GitHub Actions quality gate workflow |

## Architecture

### Key Architectural Decisions

- Workflow runs on push and pull_request for main development branches
- Explicit jobs or clearly separated steps so failures are directly attributable
- Reuse current repository commands and toolchain versions from `go.mod`, README, and frontend package metadata
- Reasonable dependency caching without obscuring behavior

## Implementation Tasks

### Task 1: Add GitHub Actions workflow for backend tests and frontend quality gates

**Files:** `.github/workflows/quality-gates.yml`

**Acceptance Criteria:**
- `.github/workflows/quality-gates.yml` exists
- the workflow contains `go test ./...`
- the workflow contains `pnpm lint`
- the workflow contains `pnpm test -- --run`
- the workflow contains `pnpm build`
- backend and frontend checks are separated enough that failure origin is obvious

**Action:** Create `.github/workflows/quality-gates.yml` to run on push and pull_request for the main development branches. Use GitHub Actions with explicit jobs or clearly separated steps so failures are directly attributable. The workflow must cover `go test ./...` at repo root and `pnpm lint`, `pnpm test -- --run`, `pnpm build` in the frontend workspace. Reuse current repository commands and toolchain versions implied by `go.mod`, README, and frontend package metadata. Add reasonable dependency caching where it improves repeatability without obscuring behavior, but do not invent a parallel build script contract.

**Verification:** Get-Content `.github/workflows/quality-gates.yml`

## Security Considerations

None

## Established Patterns

None

## Decisions

- Explicit jobs/steps for failure attribution
- Reuse existing commands rather than inventing parallel build contract
- Reasonable dependency caching for repeatability

## Deviation Log

None
