---
phase: 01-remove-backend-ai-runtime
plan: 03
subsystem: verification
tags: [go, tests, powershell, regression, webhook]
requires:
  - phase: 01-remove-backend-ai-runtime
    provides: "后端 AI 运行时、路由和持久化残留已被移除，可对非 AI 主链路做回归验证"
provides:
  - "新增无 AI 配置与路由裁剪的回归测试"
  - "新增无 AI 后端 webhook->通知->ack->quick silence 闭环验证脚本"
  - "验证 `/api/v1/ai/chat` 已返回 404 且核心告警 API 仍可用"
affects: [backend, tests, verification]
tech-stack:
  added: [powershell]
  patterns: ["用轻量单测锁定配置/路由回归点", "用单命令脚本验证真实后端告警闭环"]
key-files:
  created: [".planning/phases/01-remove-backend-ai-runtime/01-03-SUMMARY.md", "scripts/verify_backend_no_ai.ps1"]
  modified: ["internal/config/config_test.go", "internal/router/router_test.go"]
key-decisions:
  - "保留 Wave 3 的单元级回归测试，脚本则专注验证真实 webhook、通知分发与 ack/quick silence 闭环"
  - "将临时通知接收器改为 PowerShell HttpListener，避免 go run 子进程在当前环境下不稳定导致假阴性"
patterns-established:
  - "闭环验证脚本显式清空 AI 环境变量，避免依赖本机残留配置产生假阳性"
  - "验证脚本必须自己启动依赖、准备最小数据、执行 API 校验并清理进程与测试数据"
requirements-completed: [BEAI-03]
duration: 40min
completed: 2026-04-09
---

# Phase 01 Plan 03: No-AI Backend Verification Summary

**新增单元回归测试与一条可直接执行的后端闭环脚本，证明移除 AI 后核心告警链路仍可实际跑通**

## Performance

- **Duration:** 40 min
- **Started:** 2026-04-09T08:25:00Z
- **Completed:** 2026-04-09T09:05:00Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- 新增 `internal/config/config_test.go`，锁定 `config.Load()` 在没有任何 AI 环境变量时仍可返回有效配置。
- 新增 `internal/router/router_test.go`，断言核心路由仍存在且 `/api/v1/ai` 已从路由表中移除。
- 新增 `scripts/verify_backend_no_ai.ps1`，自动启动依赖、创建最小测试数据、验证 webhook 接入、通知分发、告警读取、统计、ack、quick silence 与 `/api/v1/ai/chat=404`。

## Task Commits

Each task was committed atomically:

1. **Task 1: 为无 AI 配置和路由裁剪补充回归测试** - `ad3fa77` (test)
2. **Task 2: 增加无 AI 启动与核心告警闭环验证脚本** - `66ac0cb` (test)

## Files Created/Modified

- `internal/config/config_test.go` - 覆盖无 AI 环境变量场景下的配置加载回归。
- `internal/router/router_test.go` - 覆盖核心路由保留与 AI 路由移除回归。
- `scripts/verify_backend_no_ai.ps1` - 自动化验证非 AI 后端启动与核心告警闭环。
- `.planning/phases/01-remove-backend-ai-runtime/01-03-SUMMARY.md` - 记录本计划执行结果、提交和验证结论。

## Verification

- `go test ./internal/config ./internal/router` -> passed
- `./scripts/verify_backend_no_ai.ps1` -> passed
- 脚本输出确认：
  - `/health=200`
  - `/api/v1/auth/login=200`
  - `/webhook/test-template=200`
  - `/webhook/{source}=200`
  - `/api/v1/alerts=200`
  - `/api/v1/alerts/stats=200`
  - `notification_dispatch=ok`
  - `/api/v1/alerts/{id}/ack=200`
  - `/api/v1/alerts/{id}/quick-silence=200`
  - `/api/v1/ai/chat=404`

## Decisions Made

- 把临时通知接收器从 `go run` 辅助程序切换为 PowerShell `HttpListener`，减少 Windows 环境下的子进程不确定性。
- 保持脚本自清理职责，验证结束后停止 server/listener 进程并清除本次创建的数据库测试数据。

## Deviations from Plan

None - plan intent was preserved. The listener implementation was adjusted during execution to make the scripted verification reliable in this runtime.

## Issues Encountered

- 初版脚本中的临时通知接收器未稳定监听目标端口，导致 `notification_dispatch` 观察失败；改为 PowerShell `HttpListener` 后闭环验证通过。

## User Setup Required

None - script provisions the required local data and uses existing `docker-compose.yml` dependencies automatically.

## Next Phase Readiness

- 后端移除 AI 后的启动路径、路由裁剪和核心告警闭环已经有自动化证明，可进入后续前端与文档层清理。
- 当前仓库仍存在用户未提交改动与计划外未跟踪文件，后续阶段继续只处理各自计划范围内的文件。

## Self-Check: PASSED

- Summary file exists: `.planning/phases/01-remove-backend-ai-runtime/01-03-SUMMARY.md`
- Commit `ad3fa77` found in git history
- Commit `66ac0cb` found in git history
- Verification commands passed in the current working tree

---
*Phase: 01-remove-backend-ai-runtime*
*Completed: 2026-04-09*
