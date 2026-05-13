---
name: phase-01-03
description: Add no-AI automated regression verification proving backend startup and core alert闭环 still works
metadata:
  type: spec
  source_phase: 01-remove-backend-ai-runtime
  source_plan: "01-03"
  milestone: v1.0
  status: completed
  completed: 2026-04-10
---

# Phase 01 Plan 03: 无 AI 状态下的自动化回归验证

## Context & Goals

补上"无 AI"状态下的自动化回归验证，证明后端启动路径与核心告警闭环在移除 AI 后仍可实际跑通。

Purpose: 让 BEAI-03 有可重复执行的自动证明，而不是只依赖路由存在性或肉眼检查代码删除结果。
Output: 配置/路由回归测试，以及一个可直接执行的无 AI 后端闭环验证脚本。

## Success Criteria

- 无 AI 环境变量时后端仍可启动，且认证后 API 不再暴露 `/api/v1/ai`
- 带最小数据源/渠道/路由配置的真实 Webhook 接入会落库生成告警并触发通知分发
- 认证用户可以对该告警继续执行读取统计、确认和快速静默，不破坏核心告警闭环

## Deliverables

| Artifact | Provides |
|----------|----------|
| `internal/config/config_test.go` | 无 AI 配置依赖的配置加载回归测试 |
| `internal/router/router_test.go` | 路由表中无 AI 路由且保留核心路由的回归测试 |
| `scripts/verify_backend_no_ai.ps1` | 非 AI 后端最小配置闭环验证脚本 |

## Architecture

### Key Links

| From | To | Via | Pattern |
|------|----|-----|---------|
| `scripts/verify_backend_no_ai.ps1` | `cmd/server/main.go` | `go run cmd/server/main.go` | `go run cmd/server/main.go` |
| `internal/router/router_test.go` | `internal/router/router.go` | `router.Setup(...).Routes()` | `Routes\(\)` |
| `scripts/verify_backend_no_ai.ps1` | `/webhook/:source_name` | `Invoke-RestMethod + X-API-KEY` | `/webhook/` |
| `scripts/verify_backend_no_ai.ps1` | `/api/v1/alerts/:id/ack and /api/v1/alerts/:id/quick-silence` | Bearer token from `/api/v1/auth/login` | `/api/v1/auth/login\|/ack\|/quick-silence` |

### Interfaces

From `internal/config/config.go`:
```go
func Load() *Config
```

From `internal/router/router.go`:
```go
func Setup(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *gin.Engine
```

From `internal/handlers/webhook.go`:
```go
func (h *WebhookHandler) TestInputTemplate(c *gin.Context)
```

## Implementation Tasks

### Task 1: 为无 AI 配置和路由裁剪补充回归测试

**Files:** `internal/config/config_test.go`, `internal/router/router_test.go`

**Behavior:**
- Test 1: 仅设置数据库、Redis、Server、JWT 相关环境变量时，`config.Load()` 仍返回有效配置且不要求 AI 变量
- Test 2: `router.Setup(...).Routes()` 中包含 `/health`、`/api/v1/alerts`、`/api/v1/alerts/stats`、`/webhook/test-template`，但不包含任何 `/api/v1/ai` 路径

**Action:** 按现有 Go table-driven 测试风格新增 `internal/config/config_test.go` 与 `internal/router/router_test.go`。`router_test.go` 直接构造最小 `config.Config`、`nil` DB、`nil` Redis 调用 `Setup`，仅断言路由表，不触发真实数据库请求。不要把测试写成依赖外部数据库的集成测试；本任务目标是锁定"无 AI 配置"和"无 AI 路由暴露面"两个回归点。

**Verification:**
```
go test ./internal/config ./internal/router
```

**Acceptance Criteria:**
- `go test ./internal/config ./internal/router` 通过
- `internal/router/router_test.go` 明确断言 `/api/v1/ai` 缺失且核心路由存在

**Done:** 仓库内新增可重复执行的单元级回归测试，直接锁定无 AI 配置与无 AI 路由两个关键事实。

### Task 2: 增加无 AI 启动与核心告警闭环验证脚本

**Files:** `scripts/verify_backend_no_ai.ps1`

**Action:** 创建 `scripts/verify_backend_no_ai.ps1`，把它做成真实闭环脚本而不是纯探测脚本。脚本必须在执行时显式清空 `OPENAI_API_KEY`、`OPENAI_API_BASE`、`AI_MODEL`、`AI_TIMEOUT`，启动本地依赖（优先复用 `docker-compose.yml` 中的 Postgres/Redis），并在脚本内完成最小测试数据准备：创建或重置一组可登录管理员账号、一个启用的数据源（带 `APIKey`、`input_template`、`output_template`）、一个本地 webhook 渠道和一条匹配该 source+severity 的启用路由规则。为避免依赖外部系统，脚本应同时启动一个本地临时 HTTP 接收器（如 PowerShell `HttpListener`）作为通知渠道目标，并通过 `docker exec ... psql` 或等效的 CLI 可自动方式把测试数据写入 Postgres；不要要求人工预先建账号或手填配置。随后脚本启动 `go run cmd/server/main.go`，等待 `/health` 就绪后依次验证：`POST /api/v1/auth/login` 获取 Bearer token 成功；`POST /webhook/test-template` 对最小模板样例返回 200；`POST /webhook/:source_name` 携带正确 `X-API-KEY` 返回 200 且响应含已处理告警；认证后必须同时轮询并验证 `GET /api/v1/alerts` 与 `GET /api/v1/alerts/stats` 都返回 200，其中列表响应必须包含本次新接入的 webhook 告警，统计响应必须反映该告警带来的总数或状态计数变化；本地接收器捕获到至少一次通知请求，证明 notification dispatch 已发生；再对该告警依次调用 `POST /api/v1/alerts/:id/ack` 与 `POST /api/v1/alerts/:id/quick-silence`，并再次校验 `GET /api/v1/alerts/stats` 返回 200 且统计结果体现 ack / quick silence 后的计数或状态变化，同时列表接口仍可读取该告警；同时断言 `POST /api/v1/ai/chat` 为 404。脚本最后必须清理 server 进程、临时 HTTP 接收器以及脚本自己创建的测试数据，避免遗留后台进程或污染数据库。

**Verification:**
```
powershell -ExecutionPolicy Bypass -File scripts/verify_backend_no_ai.ps1
```

**Acceptance Criteria:**
- 脚本运行期间显式清空所有 AI 环境变量
- 脚本输出中包含 `/health=200`、`/api/v1/auth/login=200`、`/webhook/test-template=200`、`/webhook/{source}=200`、`/api/v1/alerts=200`、`/api/v1/alerts/stats=200`、`notification_dispatch=ok`、`/api/v1/alerts/{id}/ack=200`、`/api/v1/alerts/{id}/quick-silence=200`、`/api/v1/ai/chat=404`
- 脚本能证明 `GET /api/v1/alerts` 的列表响应包含本次新 webhook 接入生成的告警，并拿到用于 ack / quick silence 的 `alert_id`
- 脚本能证明 `GET /api/v1/alerts/stats` 在告警接入后反映对应总数或状态计数，且在 ack / quick silence 后再次返回 200 并反映计数或状态变化
- 脚本结束后不遗留 `go run cmd/server/main.go` 后台进程、本地通知接收器进程或测试数据

**Done:** 存在一条单命令可执行的自动化验证路径，能证明后端在无 AI 配置时仍可完成 webhook 接入、通知分发、ack、quick silence 的核心闭环，同时 `/api/v1/ai` 已不可达。

## Security Considerations

### Trust Boundaries

| Boundary | Description |
|----------|-------------|
| 脚本环境 → 服务启动 | 环境变量与依赖服务组合决定无 AI 启动是否真实可复现 |
| 未认证客户端 → API 路由 | 需要证明受保护核心路由仍在，而 AI 路由已不可达 |
| 测试输入 → 模板测试端点 | `/webhook/test-template` 接收未受信模板和样例数据 |
| Webhook 输入 → 告警落库/通知分发 | 未受信外部告警进入数据库并触发通知路由，是本 phase 最关键的运行链路 |

### STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-07 | S | `scripts/verify_backend_no_ai.ps1` | mitigate | 脚本显式清空 AI 环境变量并设置必要最小环境，避免误用本机残留变量导致假阳性 |
| T-01-08 | I | `internal/router/router_test.go` | mitigate | 用路由表断言核心路径仍存在且 `/api/v1/ai` 缺失，防止删路由或漏删路由都未被发现 |
| T-01-09 | D | `scripts/verify_backend_no_ai.ps1` | mitigate | 脚本在结束时停止临时 server 进程，避免留下占端口的后台服务影响后续运行 |
| T-01-10 | T | `scripts/verify_backend_no_ai.ps1` | mitigate | 用脚本内最小固定测试配置创建 datasource/channel/route/user，并在结束时清理，避免手工环境差异掩盖 webhook→notification 链路损坏 |

## Established Patterns

None documented in source plan.

## Decisions

None documented in source plan.

## Deviation Log

None
