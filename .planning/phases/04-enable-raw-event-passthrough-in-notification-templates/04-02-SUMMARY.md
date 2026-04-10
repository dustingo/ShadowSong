---
phase: 04-enable-raw-event-passthrough-in-notification-templates
plan: 02
subsystem: datasource-preview
tags: [backend, frontend, templates, preview]
requires:
  - phase: 04-enable-raw-event-passthrough-in-notification-templates
    plan: 01
    provides: "共享 output_template 渲染契约"
provides:
  - "认证后的 datasource 模板预览接口"
  - "在产品内可见的模板字段说明和预览抽屉"
  - "前后端共用同一套 live/preview 渲染逻辑"
affects: [backend, frontend, config-ui]
tech-stack:
  added: []
  patterns: ["preview 请求复用 live rendering path", "前端直接展示 context key 以降低模板试错成本"]
key-files:
  created:
    - ".planning/phases/04-enable-raw-event-passthrough-in-notification-templates/04-02-SUMMARY.md"
  modified:
    - "internal/handlers/config.go"
    - "internal/router/router.go"
    - "internal/router/router_test.go"
    - "frontend/src/api/client.ts"
    - "frontend/src/stores/configStore.ts"
    - "frontend/src/pages/DataSources.tsx"
    - "frontend/src/types/index.ts"
key-decisions:
  - "预览接口采用 `POST /api/v1/datasources/preview`，允许带 datasource_id，也允许直接带未保存模板做临时预览"
  - "预览返回 normalized alert、rendered title/content 和 context key 列表，明确标准字段与 `event` 字段契约"
  - "数据源编辑弹窗内直接提供模板契约说明和后端驱动预览，不再保留空壳测试抽屉"
patterns-established:
  - "模板能力上线时，产品内必须同步给出可执行预览和变量说明，避免用户反向猜后端结构"
requirements-completed: [TMPL-02]
duration: 25min
completed: 2026-04-10
---

# Phase 04 Plan 02 Summary

Wave 2 把 Phase 04 的后端模板契约真正接到了产品界面里，用户可以在数据源编辑页直接预览规范化结果和最终通知输出。

## Accomplishments

- 新增受保护的 datasource preview 接口，复用 `webhook.go` 的 input/output rendering 逻辑。
- 路由测试补充 `/api/v1/datasources/preview` 注册校验。
- 数据源编辑弹窗新增模板契约说明、样例 JSON、预览抽屉、context keys 展示和最终 title/content 结果。

## Verification

- `go test ./internal/handlers ./internal/router -run "TestWebhook|TestRouter" -count=1`
- `pnpm.cmd --dir frontend build`

## Residual Notes

- 预览接口当前返回第一条标准化结果；数组样例会取首个对象作为模板调试样本，已经足够覆盖当前产品预览用途。

