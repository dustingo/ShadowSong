# Phase 13: Harden Notification Delivery Path - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the auto-selected discuss outcomes.

**Date:** 2026-04-21
**Phase:** 13-harden-notification-delivery-path
**Areas discussed:** Failure Boundary, Logging Contract, Traceability Scope, Verification Boundary
**Mode:** auto

---

## Failure Boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Protect Async Entry | 在异步通知入口补 panic recover 和统一边界 | ✓ |
| Sender-Only Fixes | 只在各发送器内部零散补防护 | |
| Queue Rewrite | 直接改造成新异步架构 | |

**User's choice:** Auto-selected recommended option: 在真实异步入口建立批次级防护边界。
**Notes:** 当前风险来自裸 goroutine，因此 recover 需要靠近 `go h.processAlertNotifications(newAlerts)`。

---

## Logging Contract

| Option | Description | Selected |
|--------|-------------|----------|
| Stable Context Logging | 日志至少包含 alert/channel/stage 等稳定上下文 | ✓ |
| Raw Error Strings | 继续保留零散 `fmt.Printf` 风格 | |
| Full Observability Stack | 直接引入大型日志/指标平台 | |

**User's choice:** Auto-selected recommended option: 建立稳定可检索的最小日志契约。
**Notes:** 这满足当前 requirement，也避免把 phase 扩大成平台建设。

---

## Traceability Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal Failure Traceability | 关联告警、渠道和失败阶段，覆盖关键失败点 | ✓ |
| Persist Every Event | 建通知事件表和完整审计历史 | |
| Retry System First | 优先做重试和补偿 | |

**User's choice:** Auto-selected recommended option: 做最小必要追踪，而不是扩成新子系统。
**Notes:** 当前 roadmap 明确要求基础可靠性和可追踪性，不要求重试基础设施。

---

## Verification Boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Go Tests For Reliability Paths | 补 Go 测试验证 panic 和失败边界 | ✓ |
| Manual Inspection Only | 仅靠人工代码阅读验收 | |
| E2E Infra First | 先搭完整外部通知环境再验证 | |

**User's choice:** Auto-selected recommended option: 用现有 Go 测试补可靠性验证。
**Notes:** Phase 12 已接好 CI，新增后端测试可以直接进入门禁。

---

## the agent's Discretion

- 日志 helper 和注入 seam 的最小实现方式
- panic 测试和发送失败测试的替身设计

## Deferred Ideas

- 重试队列与持久化通知任务
- 分布式 trace 和集中可观测性平台
- 更全面的通知指标和 SLA 监控
