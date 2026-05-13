---
name: phase-01-01
description: Remove backend AI runtime assembly链路 - cut config, client, handler, and route dependencies
metadata:
  type: spec
  source_phase: 01-remove-backend-ai-runtime
  source_plan: "01-01"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 01 Plan 01: 移除后端 AI 运行时装配链路

## Context & Goals

移除后端 AI 运行时装配链路，先切断配置、客户端、Handler 和路由之间的直接依赖。

Purpose: 先满足 BEAI-01 与 BEAI-02，避免后续模型和验证任务建立在仍然存在的 AI 暴露面上。
Output: 不再包含 AI 运行时装配的配置/启动/路由代码，以及已删除的 AI 源文件。

## Success Criteria

- 服务启动在未提供任何 AI 环境变量时仍能进入常规初始化流程
- 认证 API 不再暴露任何 /api/v1/ai 路由
- 后端运行时不再构造 AI 客户端或 AI Handler

## Deliverables

| Artifact | Provides |
|----------|----------|
| `internal/config/config.go` | 无 AIConfig 的运行时配置结构 |
| `internal/router/router.go` | 不含 /api/v1/ai 的路由表 |
| `cmd/server/main.go` | 不依赖 AI 配置的启动路径 |

## Architecture

### Key Links

| From | To | Via | Pattern |
|------|----|-----|---------|
| `cmd/server/main.go` | `internal/config/config.go` | `config.Load()` | `config\.Load\(` |
| `internal/router/router.go` | `internal/handlers/ai.go` | `NewAIHandler` | `NewAIHandler` |

### Interfaces

From `internal/config/config.go`:
```go
type Config struct {
    Database DatabaseConfig
    Redis    RedisConfig
    Server   ServerConfig
    AI       AIConfig
    Security SecurityConfig
}

func Load() *Config
```

From `internal/router/router.go`:
```go
func Setup(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine
```

## Implementation Tasks

### Task 1: 移除启动配置中的 AI 专用结构与读取逻辑

**Files:** `internal/config/config.go`, `cmd/server/main.go`

**Action:** 删除 `Config` 上的 `AI` 字段、`AIConfig` 结构以及 `OPENAI_API_KEY`、`OPENAI_API_BASE`、`AI_MODEL`、`AI_TIMEOUT` 的读取逻辑；保留数据库、Redis、Server、Security 配置结构不变。不要顺手重构 `config.Load()` 的返回方式，也不要改变 `cmd/server/main.go` 的主流程顺序；仅把启动路径修正为在无 AI 配置时仍可继续初始化数据库、Redis、Gin 与路由。若 `cmd/server/main.go` 仅因传递整份 `cfg` 给路由而保留现有函数签名，则维持该签名，避免扩大改动面。

**Verification:**
```
rg -n "OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|type AIConfig|\bAI\s+AIConfig" internal/config cmd/server
```

**Acceptance Criteria:**
- `rg -n "OPENAI_API_KEY|OPENAI_API_BASE|AI_MODEL|AI_TIMEOUT|type AIConfig|\bAI\s+AIConfig" internal/config cmd/server` 无匹配
- `go test ./internal/config` 可执行通过

**Done:** 后端配置层不再声明或读取 AI 运行时配置，服务启动代码在无 AI 环境变量时仍只依赖数据库、Redis、JWT 与服务器配置。

### Task 2: 移除 AI Handler、客户端与受保护路由暴露面

**Files:** `internal/router/router.go`, `internal/handlers/ai.go`, `internal/ai/client.go`

**Action:** 在尊重当前脏工作树的前提下，仅删除 AI 相关装配：从 `internal/router/router.go` 中移除 `NewAIHandler` 初始化与 `/api/v1/ai` 路由组；删除 `internal/handlers/ai.go` 与 `internal/ai/client.go` 文件。不要动告警、配置、Webhook、WebSocket、用户与认证路由；这些组必须保持原有路径和中间件关系。若 `internal/router/router.go` 中存在用户未提交改动，合并删除操作时只去掉 AI 代码块，不回退其他差异。

**Verification:**
```powershell
if ((Test-Path 'internal/handlers/ai.go') -or (Test-Path 'internal/ai/client.go')) { exit 1 }
rg -n "NewAIHandler|Group\("/ai"|/api/v1/ai|type AIHandler" internal/router internal/handlers cmd
```

**Acceptance Criteria:**
- `Test-Path internal/handlers/ai.go` 返回 false
- `Test-Path internal/ai/client.go` 返回 false
- `rg -n "NewAIHandler|Group\("/ai"|/api/v1/ai|type AIHandler" internal/router internal/handlers cmd` 无匹配
- `go test ./internal/router` 可执行通过

**Done:** 认证 API 不再注册任何 AI 路径，后端代码中不再存在 AI Handler 或 AI 客户端运行时文件。

## Security Considerations

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| 环境变量 → 启动配置 | 非受信环境输入决定服务是否能启动 |
| 客户端 → 认证 API | 外部请求可能继续探测已下线的 AI 路由 |

### STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-01 | D | `internal/config/config.go` | mitigate | 删除 AI 配置强依赖，确保仅 `JWT_SECRET` 仍是必需启动项，避免因缺失 AI 变量导致拒绝服务 |
| T-01-02 | I | `internal/router/router.go` | mitigate | 彻底删除 `/api/v1/ai` 路由注册与 handler 装配，防止暴露已废弃接口面 |
| T-01-03 | T | `internal/router/router.go` | mitigate | 改动时仅移除 AI 代码块，不改动其他受保护路由与中间件顺序，避免清理 AI 时破坏鉴权链路 |

## Established Patterns

None documented in source plan.

## Decisions

None documented in source plan.

## Deviation Log

None
