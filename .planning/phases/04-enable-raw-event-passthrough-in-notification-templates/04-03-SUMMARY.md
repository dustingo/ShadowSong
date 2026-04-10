---
phase: 04-enable-raw-event-passthrough-in-notification-templates
plan: 03
subsystem: verification
tags: [verification, powershell, notifications, passthrough]
requires:
  - phase: 04-enable-raw-event-passthrough-in-notification-templates
    plan: 01
    provides: "live notification render context with `event` passthrough"
  - phase: 04-enable-raw-event-passthrough-in-notification-templates
    plan: 02
    provides: "datasource preview UI and route wiring"
provides:
  - "Phase 04 legacy/raw template end-to-end verification script"
  - "Phase 04 verification evidence report"
affects: [verification, scripts, phase-closure]
tech-stack:
  added: [powershell]
  patterns: ["沿用 Phase 01 的自清理后端验证脚本模式", "同一条脚本同时断言 legacy 和 raw passthrough 两条通知路径"]
key-files:
  created:
    - "scripts/verify_template_passthrough.ps1"
    - ".planning/phases/04-enable-raw-event-passthrough-in-notification-templates/04-VERIFICATION.md"
    - ".planning/phases/04-enable-raw-event-passthrough-in-notification-templates/04-03-SUMMARY.md"
key-decisions:
  - "验证脚本直接播种两套 datasource：一套只走 legacy 标准字段模板，一套走 `event` 原始字段模板"
  - "通知捕获继续使用本地 PowerShell HttpListener，避免额外引入测试服务"
  - "verification 报告直接记录脚本和前端构建的实测结果，作为 phase closure 证据"
patterns-established:
  - "模板契约类 phase 收尾时必须同时提供单元回归和真实通知分发脚本证据"
requirements-completed: [TMPL-02, TMPL-03]
duration: 20min
completed: 2026-04-10
---

# Phase 04 Plan 03 Summary

Wave 3 为 Phase 04 留下了可重复执行的真实通知验证路径，并用实际命令结果证明 legacy 模板兼容和 raw event passthrough 都已打通。

## Accomplishments

- 新增 `scripts/verify_template_passthrough.ps1`，自动启动依赖、播种两套 datasource、捕获 webhook channel 通知并断言 legacy/raw 两条路径的标题和内容。
- 执行脚本并记录 `legacy_notification=ok`、`raw_notification=ok`、`phase04_passthrough_verification=passed`。
- 再次执行 `pnpm.cmd --dir frontend build`，确认 datasource 预览 UI 改动没有破坏前端构建。
- 产出 `04-VERIFICATION.md`，把本 phase 的 must-haves、证据和 requirement coverage 固化到 phase 目录。

## Verification

- `pwsh -ExecutionPolicy Bypass -File scripts/verify_template_passthrough.ps1`
- `pnpm.cmd --dir frontend build`

## Residual Notes

- 本轮构建仍有既有的大 chunk warning，但不影响 Phase 04 目标达成。
- `docker-compose.yml` 的历史 `version` warning 仍存在，属于既有环境噪音。

