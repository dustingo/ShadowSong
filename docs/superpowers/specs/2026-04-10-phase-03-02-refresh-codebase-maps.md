---
name: phase-03-02
description: Refresh codebase maps to reflect non-AI runtime state
metadata:
  type: spec
  source_phase: 03-align-docs-and-verification
  source_plan: "02"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 03 Plan 02: Refresh Codebase Maps

## Context & Goals

Plan 03-01 updated user-facing docs. This plan refreshes GSD codebase maps so future planning inputs only describe the current non-AI alert system's valid runtime paths and real constraints.

**Goal:** Complete DATA-02 — remove or decouple AI-only schema/data/references from current runtime paths.

## Success Criteria

- GSD codebase maps no longer treat deleted AI processors, pages, models, and integrations as active architecture
- Remaining historical AI mentions are explicitly scoped as historical context, not current runtime paths
- AI-only schema/data/references are decoupled from current runtime paths and won't mislead future planning

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| Architecture description aligned with current state | `.planning/codebase/ARCHITECTURE.md` | System layers, data flow, entry points |
| Tech stack without AI runtime deps | `.planning/codebase/STACK.md` | No OpenAI integration |
| Structure aligned with actual directory layout | `.planning/codebase/STRUCTURE.md` | Current file org |
| External dependencies without OpenAI | `.planning/codebase/INTEGRATIONS.md` | No AI integration |
| Conventions not using AI pages/processors as examples | `.planning/codebase/CONVENTIONS.md` | Current patterns |
| Risk register without deleted AI runtime | `.planning/codebase/CONCERNS.md` | Decoupled risks |
| Test docs without AI logging as current subject | `.planning/codebase/TESTING.md` | Current test approach |

## Architecture

### Map Refresh Strategy

**ARCHITECTURE.md, STACK.md, STRUCTURE.md, INTEGRATIONS.md:**
- Remove current-state descriptions of `internal/ai`, `internal/handlers/ai.go`, `frontend/src/pages/AIAssistant.tsx`, `AILog`, `SilenceRecommendation`, OpenAI config, AI assistant data flows
- Replace with current non-AI architecture, real external integrations, actual directory structure

**CONVENTIONS.md, CONCERNS.md, TESTING.md:**
- Change AI page, AI processor, AI client, AI logging examples to current-phase-aligned descriptions
- If AI mentions required, explicitly label as "historical cleanup context" not current code paths
- Focus on `NewAIHandler`, `AIAssistant.tsx`, `internal/ai/client.go`, `AILog` as current examples

### Not Modified

- `.planning/phases/**` historical artifacts
- `AGENTS.md`

## Implementation Tasks

### Task 1: Refresh Runtime, Structure, and Integration Facts in Codebase Maps

**Files:** `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STACK.md`, `.planning/codebase/STRUCTURE.md`, `.planning/codebase/INTEGRATIONS.md`

**Acceptance Criteria:**
- Four codebase maps no longer present AI handler, AI page, AI model, OpenAI integration, or AI assistant flow as current facts
- `rg -n 'internal/ai|internal/handlers/ai\.go|AIAssistant|AILog|SilenceRecommendation|OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|OpenAI-compatible|AI Assistant Flow' .planning/codebase/ARCHITECTURE.md .planning/codebase/STACK.md .planning/codebase/STRUCTURE.md .planning/codebase/INTEGRATIONS.md` returns 0 matches

**Action:** Based on completed Phase 1/2 facts, rewrite entries still treating AI components as current runtime. Change narratives to current non-AI architecture, real external integrations, actual directory structure. Do not rewrite `.planning/phases/**` historical artifacts or `AGENTS.md`. Do not wrap historical AI goals as current features.

**Verification:** `rg -n 'internal/ai|internal/handlers/ai\.go|AIAssistant|AILog|SilenceRecommendation|OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|OpenAI-compatible|AI Assistant Flow' .planning/codebase/ARCHITECTURE.md .planning/codebase/STACK.md .planning/codebase/STRUCTURE.md .planning/codebase/INTEGRATIONS.md` returns 0 matches

---

### Task 2: Close AI-Only Residuals in Conventions, Concerns, and Testing Maps

**Files:** `.planning/codebase/CONVENTIONS.md`, `.planning/codebase/CONCERNS.md`, `.planning/codebase/TESTING.md`

**Acceptance Criteria:**
- Three supporting maps no longer use deleted AI code as current examples, current risks, or current test subjects
- `rg -n 'NewAIHandler|AIAssistant|internal/ai/client\.go|internal/handlers/ai\.go|AILog|AI suggestion|AI interactions|AI assistant' .planning/codebase/CONVENTIONS.md .planning/codebase/CONCERNS.md .planning/codebase/TESTING.md` returns 0 matches

**Action:** Change conventions, risks, and testing maps still using AI pages, AI processors, AI client, AI logging as current examples to current-phase-aligned descriptions. If AI mentions required, explicitly mark as "historical cleanup context" not current code paths. Do not rewrite history spec docs or `.planning/phases/**` verification reports.

**Verification:** `rg -n 'NewAIHandler|AIAssistant|internal/ai/client\.go|internal/handlers/ai\.go|AILog|AI suggestion|AI interactions|AI assistant' .planning/codebase/CONVENTIONS.md .planning/codebase/CONCERNS.md .planning/codebase/TESTING.md` returns 0 matches

## Security Considerations

| Threat ID | Category | Component | Disposition | Mitigation |
|-----------|----------|-----------|-------------|------------|
| T-03-04 | T | ARCHITECTURE.md | mitigate | Remove current-state descriptions of deleted AI runtime, pages, models, data flows to prevent future plans from being driven by erroneous inputs |
| T-03-05 | D | INTEGRATIONS.md | mitigate | Remove OpenAI integration and AI config descriptions to prevent executors from treating non-existent third-party dependencies as requirements |
| T-03-06 | R | CONCERNS.md | accept | Allow minimal historical risk context, but must explicitly mark as historical to avoid confusion with current risks |

## Decisions

- Codebase maps updated to reflect actual non-AI architecture
- Historical AI artifacts clearly marked as context, not current paths

## Deviation Log

None — plan executed as written.