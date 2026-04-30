---
phase: 19-enable-safe-recovery-operations
plan: 03
subsystem: ui
tags: [react, vite, antd, vitest, deliveries, recovery]
requires:
  - phase: 19-enable-safe-recovery-operations
    provides: read-only deliveries history page and protected delivery recovery POST APIs
provides:
  - delivery retry and replay actions on failed history rows for operator and admin users
  - reason-gated recovery modal with per-delivery loading and inline recovery feedback
  - delivery history regression coverage for viewer read-only and operator recovery flows
affects: [delivery-ui, phase-20-harden-ingress-and-runtime-readiness, phase-21-ops-visibility-surfaces]
tech-stack:
  added: []
  patterns: [capability-gated recovery actions, inline recovery audit feedback, per-row async action locking]
key-files:
  created: []
  modified: [frontend/src/authz/capabilities.ts, frontend/src/types/index.ts, frontend/src/api/client.ts, frontend/src/pages/Deliveries.tsx, frontend/src/pages/Deliveries.test.tsx]
key-decisions:
  - "恢复入口继续绑定 process_alerts 能力，并通过 canRecoverDeliveries 显式表达 UI 语义，避免历史页误用 view_config。"
  - "恢复原因由 Modal 表单强制必填，提交成功后立即刷新列表和当前详情，再用页面内结果块回显 recovery_id 与 resulting delivery。"
  - "同一 delivery 的恢复动作在请求进行中禁用重试和重放按钮，优先降低双击重复发送风险。"
patterns-established:
  - "Pattern 1: 失败 delivery 的写操作按钮只在具备 process_alerts 能力时渲染，并与 row 级 loading 状态绑定。"
  - "Pattern 2: 恢复提交后同时刷新列表与当前详情，页面内保留最新 recovery 结果块作为用户可见审计反馈。"
requirements-completed: [DELV-03, DELV-04, DELV-05]
duration: 17min
completed: 2026-05-01
---

# Phase 19 Plan 03: Enable Safe Recovery Operations Summary

**失败通知历史页现在支持受控 retry/replay、原因审计输入和 resulting delivery 结果回显，同时保持 viewer 全程只读**

## Performance

- **Duration:** 17 min
- **Started:** 2026-05-01T00:00:00+08:00
- **Completed:** 2026-05-01T00:17:19+08:00
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- 为前端补齐 delivery recovery 请求/响应类型、`retry/replay` API 合同和显式恢复能力判断。
- 在 `/deliveries` 页面接入失败记录的 `retry` / `replay` 按钮、原因弹窗、row 级防重入和结果反馈块。
- 回归测试覆盖 viewer 无恢复入口、operator 原因必填、成功恢复后的列表与详情刷新。

## Task Commits

Each task was committed atomically:

1. **Task 1: 扩展前端 recovery 合同与权限判断** - `2c08ed1` (feat)
2. **Task 2: 在 Deliveries 页面接入 retry/replay 交互、结果刷新与权限回归** - `bd7d981` (feat)

## Files Created/Modified

- `frontend/src/types/index.ts` - 新增 delivery recovery 请求和结果 DTO。
- `frontend/src/authz/capabilities.ts` - 新增 `canRecoverDeliveries`，显式复用 `process_alerts` 能力。
- `frontend/src/api/client.ts` - 新增 `deliveryApi.retry` 与 `deliveryApi.replay`。
- `frontend/src/pages/Deliveries.tsx` - 新增恢复弹窗、原因校验、row 级 loading、反馈结果块和刷新逻辑。
- `frontend/src/pages/Deliveries.test.tsx` - 覆盖 viewer 只读、operator 弹窗与成功恢复刷新回显。

## Decisions Made

- 恢复入口不复用 `view_config` 可见性，而是通过 `canRecoverDeliveries(user)` 显式绑定到 `process_alerts`，保证 viewer 在所有 UI 分支都只读。
- 恢复结果优先在页面内持久展示 `recovery_id`、`resulting_delivery_id` 和 `error_message`，而不是只依赖 toast。
- 不引入新的前端 store；`Deliveries` 页面本地状态已足够承载单页恢复交互和刷新闭环。

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] 收口恢复原因校验的未处理 Promise**
- **Found during:** Task 2 (在 Deliveries 页面接入 retry/replay 交互、结果刷新与权限回归)
- **Issue:** Modal 点击确认时，`validateFields()` 在原因为空时会抛出未处理拒绝，导致测试失败并暴露出真实错误路径。
- **Fix:** 在 `handleRecoverySubmit` 中显式捕获表单校验失败并直接返回，同时改用 `destroyOnHidden` 保持 Modal 生命周期与当前 Ant Design 约定一致。
- **Files modified:** `frontend/src/pages/Deliveries.tsx`
- **Verification:** `pnpm --dir frontend test -- --run src/pages/Deliveries.test.tsx`
- **Committed in:** `bd7d981`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** 修正项只服务于恢复交互的正确性和测试稳定性，没有扩展产品范围。

## Issues Encountered

- `Deliveries` 集成测试包含列表加载、详情刷新和 Modal 提交，Vitest 默认 5 秒超时过紧；测试中把该用例超时调到 10 秒以稳定覆盖完整恢复闭环。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 前端历史页已经具备单条恢复入口和结果反馈，后续 phase 可以直接复用这些 recovery 结果做更高层的运维聚合与健康摘要。
- 当前结果块已经把 `resulting_delivery_id` 暴露到 UI，后续如果要增加跳转或串联恢复审计详情，不需要再改 recovery API 合同。

## Self-Check: PASSED

---
*Phase: 19-enable-safe-recovery-operations*
*Completed: 2026-05-01*
