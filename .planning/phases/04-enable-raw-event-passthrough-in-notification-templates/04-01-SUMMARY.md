---
phase: 04-enable-raw-event-passthrough-in-notification-templates
plan: 01
subsystem: webhook-rendering
tags: [backend, templates, webhook, compatibility]
requires:
  - phase: 03-align-docs-and-verification
    provides: "现有 webhook 验证与非 AI 基线"
provides:
  - "兼容旧模板字段的 output_template 渲染上下文"
  - "通过 `event` 暴露原始 webhook JSON 的稳定模板契约"
  - "覆盖兼容性和原始字段透传的回归测试"
affects: [backend, templates, notifications]
tech-stack:
  added: []
  patterns: ["共享渲染上下文供 live send 和 preview 共用", "模板上下文向后兼容优先"]
key-files:
  created:
    - ".planning/phases/04-enable-raw-event-passthrough-in-notification-templates/04-01-SUMMARY.md"
  modified:
    - "internal/handlers/webhook.go"
    - "internal/handlers/webhook_test.go"
key-decisions:
  - "保留 `alert_name`、`severity`、`message`、`labels` 等原有顶层字段，避免现有 output_template 迁移"
  - "新增稳定顶层变量 `event` 承载 `alert.Raw` 解码后的原始 JSON，并额外提供嵌套 `alert` 便于可读性"
  - "当 `alert.Raw` 或 `alert.Labels` 解码失败时回退为空 map，确保标准字段模板仍可正常渲染"
patterns-established:
  - "notification preview 必须复用 live notification 的共享 render context，而不是复制一套前端近似逻辑"
requirements-completed: [TMPL-01, TMPL-03]
duration: 15min
completed: 2026-04-10
---

# Phase 04 Plan 01 Summary

Wave 1 固化了通知模板渲染契约：旧模板继续可用，同时新模板可以通过 `event` 直接访问原始 webhook JSON。

## Accomplishments

- 将 `renderNotification` 抽成共享渲染路径，统一构建包含标准字段、`labels`、`event`、`alert` 的模板上下文。
- 复用现有 `toJson`、`default`、`get`、`lookup` 等 helper，避免引入新的模板 DSL。
- 新增回归测试，覆盖旧字段兼容、嵌套原始字段访问和缺失原始数据时的安全降级。

## Verification

- `go test ./internal/handlers -run TestWebhook -count=1`

## Residual Notes

- 此 plan 只锁定后端渲染契约；数据源预览接口和前端引导在后续 Wave 2 完成。

