---
name: phase-01-02
description: Remove backend AI-specific models, migrations, and runtime text while preserving alert/notification main链路
metadata:
  type: spec
  source_phase: 01-remove-backend-ai-runtime
  source_plan: "01-02"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 01 Plan 02: 移除后端 AI 专用模型与迁移项

## Context & Goals

移除仍挂在后端运行路径上的 AI 专用模型、迁移项和运行时文案，同时保证告警与通知主链路继续围绕非 AI 数据结构工作。

Purpose: 切断 AI 删除后的残余持久化耦合，避免数据库迁移和常规 API 仍携带 AI 结构。
Output: 精简后的 Alert/Model 定义、数据库迁移列表，以及不再引用 AI 品牌/能力的后端运行时文案。

## Success Criteria

- 后端持久化模型不再包含仅供 AI 使用的字段和表
- 数据库迁移不再创建 AI 日志或 AI 推荐相关表
- 保留下来的告警/配置/通知运行路径不引用已删除的 AI 结构

## Deliverables

| Artifact | Provides |
|----------|----------|
| `internal/models/alert.go` | 不含 AI 分析字段的 Alert 模型 |
| `internal/models/models.go` | 不含 AILog 与 SilenceRecommendation 的模型集合 |
| `internal/database/postgres.go` | 不迁移 AI 专用表的数据库初始化逻辑 |

## Architecture

### Key Links

| From | To | Via | Pattern |
|------|----|-----|---------|
| `internal/database/postgres.go` | `internal/models/alert.go` | GORM migration table list | `models\.Alert` |
| `internal/handlers/config.go` | `internal/notifier/notifier.go` | TestChannel message content | `TestChannel\|testContent` |

### Interfaces

From `internal/models/alert.go`:
```go
type Alert struct {
    AlertID     string
    Source      string
    AlertName   string
    Severity    string
    Message     string
    Labels      datatypes.JSON
    Fingerprint string
    TriggerTime time.Time
    ReceivedAt  time.Time
    Status      string
    Raw         datatypes.JSON
}
```

From `internal/database/postgres.go`:
```go
func InitDB(cfg *config.Config) (*gorm.DB, error)
```

## Implementation Tasks

### Task 1: 清理 Alert 模型与 AI 专用持久化实体

**Files:** `internal/models/alert.go`, `internal/models/models.go`, `internal/database/postgres.go`

**Action:** 从 `Alert` 中删除仅由 AI 功能产生和消费的字段：`AISummary`、`AIRootCause`、`AISeverity`、`AISuggestions`、`AITags`。从 `internal/models/models.go` 中删除 `AILog` 与 `SilenceRecommendation` 结构及其校验逻辑，并同步从 `internal/database/postgres.go` 的迁移表列表中移除这两个模型。不要动 `Alert` 的去重、确认、状态、原始载荷和时间字段；这些仍服务于 BEAI-03 的核心告警链路。

**Verification:**
```
rg -n "AISummary|AIRootCause|AISeverity|AISuggestions|AITags|type AILog|type SilenceRecommendation|models\.AILog|models\.SilenceRecommendation" internal/models internal/database
```

**Acceptance Criteria:**
- `rg -n "AISummary|AIRootCause|AISeverity|AISuggestions|AITags|type AILog|type SilenceRecommendation|models\.AILog|models\.SilenceRecommendation" internal/models internal/database` 无匹配
- `go test ./internal/models` 可执行通过

**Done:** 后端模型层与数据库迁移列表不再包含任何仅供 AI 使用的字段或表定义。

### Task 2: 清理保留后端路径中的 AI 运行时字样并做耦合回扫

**Files:** `internal/handlers/config.go`

**Action:** 把仍会通过后端 API 直接返回给用户的 AI 文案改为中性告警系统表述，例如 `TestChannel` 中的测试消息内容；随后回扫 `internal/handlers`、`internal/database`、`internal/models`，确保剩余运行时代码不再引用已删除的 AI 字段/实体。不要修改模块路径 `github.com/game-ops/ai-alert-system` 或 JWT issuer，这类品牌级调整留给后续文档/全局收口阶段处理。

**Verification:**
```
rg -n "AI Alert System|AILog|SilenceRecommendation|AISummary|AIRootCause|AISuggestions|AITags" internal/handlers internal/database internal/models
```

**Acceptance Criteria:**
- `internal/handlers/config.go` 中的测试通知文案不再包含 `AI Alert System`
- `rg -n "AILog|SilenceRecommendation|AISummary|AIRootCause|AISuggestions|AITags" internal/handlers internal/database internal/models` 无运行时匹配
- `go test ./...` 可执行通过

**Done:** 后端保留下来的常规 API 与通知路径不再携带 AI 专用文案或已删除模型引用，完整后端测试可通过。

## Security Considerations

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| 数据库迁移 → 已持久化 schema | 清理表/字段时可能破坏现有核心告警表的可用性 |
| 常规配置 API → 用户可见返回 | 后端运行时文案若残留 AI 表述会造成能力误导 |

### STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-04 | T | `internal/models/alert.go` | mitigate | 仅删除 AI 专用字段，保留告警状态、去重、确认和原始数据字段，避免误伤核心告警模型 |
| T-01-05 | D | `internal/database/postgres.go` | mitigate | 从迁移列表只移除 AI 专用实体，执行 `go test ./...` 验证剩余运行路径仍可编译和测试 |
| T-01-06 | R | `internal/handlers/config.go` | mitigate | 将用户可见测试消息改成中性表述，避免运维误以为系统仍提供 AI 能力 |

## Established Patterns

None documented in source plan.

## Decisions

None documented in source plan.

## Deviation Log

None
