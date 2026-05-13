# GSD to Superpowers Migration Design

## Overview

Convert all existing GSD planning documents (`.planning/phases/*/XX-PLAN.md`) to the superpowers format (`docs/superpowers/specs/`).

## Goals

- Maintain all planning information while adopting the superpowers documentation structure
- Enable future planning to follow the superpowers workflow (brainstorming → writing-plans → execution)
- Preserve historical context from shipped phases (v1.0-v1.3)

## Scope

**All phases: 1-21**
- Phase 1-17: v1.0-v1.3 (shipped milestones)
- Phase 18-21: v1.4 (current/future work)

**Pilot scope (this phase):**
- Phase 18: Establish Delivery Ledger (plans 18-01, 18-02, 18-03)
- Phase 19: Enable Safe Recovery Operations (plans 19-01, 19-02, 19-03)

## Document Structure

### Source Format (GSD)
```
.planning/phases/<phase-name>/
├── XX-01-PLAN.md      # Frontmatter: phase, plan, subsystem, tags, provides, etc.
├── XX-01-SUMMARY.md   # Execution summary
├── XX-01-REVIEW.md    # Code review
├── XX-01-VERIFICATION.md
├── XX-RESEARCH.md
└── XX-CONTEXT.md
```

### Target Format (Superpowers)
```
docs/superpowers/
├── specs/
│   └── YYYY-MM-DD-phase-XX-YY-design.md  # One spec per plan
└── archive/
    └── .planning/  # Original docs moved here after conversion
```

## Field Mapping

| GSD Frontmatter | Superpowers Section |
|-----------------|---------------------|
| `phase`, `plan`, `subsystem` | Document header + filename |
| `objective` | Context & Goals |
| `must_haves.truths` | Success Criteria |
| `must_haves.artifacts` | Deliverables |
| `tasks` | Implementation Tasks |
| `threat_model` | Security Considerations |
| `patterns-established` | Established Patterns |
| `key-decisions` | Decisions |

## Spec Template

```markdown
---
name: phase-XX-YY
description: [one-line summary]
metadata:
  type: spec
  source_phase: [phase name]
  source_plan: [plan number]
  milestone: [v1.x]
  status: [draft|active|completed]
---

# Phase XX Plan YY: [Title]

## Context & Goals

[Why this matters, what problem it solves]

## Success Criteria

- [Criterion 1]
- [Criterion 2]

## Deliverables

| Artifact | Path | Provides |
|----------|------|----------|
| [Name] | [path] | [description] |

## Architecture

[Key architectural decisions]

## Implementation Tasks

### Task 1: [Name]
- Files: [list]
- Verification: [command]

### Task 2: [Name]
- Files: [list]
- Verification: [command]

## Security Considerations

[Threat model summary]

## Established Patterns

- [Pattern 1]
- [Pattern 2]

## Decisions

- [Decision 1]
- [Decision 2]

## Deviation Log

[Any deviations from the original plan]
```

## Migration Approach

### Phase 1: Pilot (Current)
1. Convert Phase 18 (3 plans) and Phase 19 (3 plans) as validation
2. User reviews output and provides feedback
3. Refine template if needed

### Phase 2: Batch Conversion
1. Archive original `.planning/phases/` to `docs/superpowers/archive/`
2. Use Agent to parallel-convert remaining phases (1-17, 20-21)
3. Create `docs/superpowers/MEMORY.md` as index

## Success Criteria

1. Phase 18-19 converted specs accurately reflect original plan intent
2. All required information preserved (no data loss)
3. Format validated by user
4. Remaining phases converted in batch
