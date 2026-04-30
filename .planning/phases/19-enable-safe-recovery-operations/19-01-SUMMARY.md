---
phase: 19-enable-safe-recovery-operations
plan: 01
subsystem: ui
tags: [react, vite, antd, react-router, vitest, deliveries]
requires:
  - phase: 18-establish-delivery-ledger
    provides: protected read-only delivery list/detail API and ledger evidence schema
provides:
  - protected /deliveries history page with query-backed filters
  - alert-to-delivery deeplink via alert_id
  - failure evidence drawer showing attempts and frozen snapshots
affects: [phase-19-safe-recovery-operations, phase-21-ops-visibility-surfaces, delivery-ui]
tech-stack:
  added: []
  patterns: [query-string driven filter state, frontend ledger evidence drawer, typed axios unwrap for interceptor-backed reads]
key-files:
  created: [frontend/src/pages/Deliveries.tsx, frontend/src/pages/Deliveries.test.tsx]
  modified: [frontend/src/types/index.ts, frontend/src/api/client.ts, frontend/src/pages/Alerts.tsx, frontend/src/pages/Alerts.test.tsx, frontend/src/pages/index.ts, frontend/src/App.tsx]
key-decisions:
  - "继续沿用 capabilityViewConfig 保护 /deliveries，只交付只读账本查询面，不在本 plan 混入恢复写入口。"
  - "投递历史页使用 URL query 作为筛选真源，alert deeplink 直接落到 alert_id 条件，避免额外状态层。"
  - "详情证据只展示后端已下发的 attempts、final_failure_summary 和冻结快照，不展开任何 channel secret/config。"
patterns-established:
  - "Pattern 1: delivery 查询页通过 search params 初始化和回显筛选条件，再把允许的过滤参数直接传给 deliveryApi.list。"
  - "Pattern 2: 历史页详情使用单条 detail API 拉取证据 Drawer，列表只承担摘要，不猜测失败细节。"
requirements-completed: [OPER-01, OPER-04]
duration: 12min
completed: 2026-05-01
---

# Phase 19 Plan 01: Enable Safe Recovery Operations Summary

**受保护的通知投递历史页、失败证据 Drawer 和 alert_id deeplink 已落到前端，维护者可以直接从告警跳到账本并查看 attempts/冻结快照**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-30T23:48:00+08:00
- **Completed:** 2026-05-01T00:00:17+08:00
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- 为 `GET /api/v1/deliveries` / `GET /api/v1/deliveries/:id` 补齐前端 delivery DTO 与 API 合同。
- 新增受 `view_config` 保护的 `/deliveries` 页面、菜单入口，以及告警页到投递历史页的 `alert_id` deeplink。
- 历史页支持 URL 初始化筛选、后端 `total/limit/offset` 分页语义和失败证据 Drawer，并补齐 viewer 只读回归测试。

## Task Commits

Each task was committed atomically:

1. **Task 1: 定义前端 delivery 读面合同并接入受保护路由与 deeplink** - `a2fd9e1` (feat)
2. **Task 2: 实现可筛选的投递历史页与失败证据详情回归** - `84043d5` (feat)

## Files Created/Modified
- `frontend/src/types/index.ts` - 新增 delivery、attempt、final failure 和冻结快照的前端类型。
- `frontend/src/api/client.ts` - 新增 `deliveryApi.list/get`，并为拦截器拆包后的只读响应补最小类型收口。
- `frontend/src/App.tsx` - 挂载受保护的 `/deliveries` 路由并加入菜单入口。
- `frontend/src/pages/Alerts.tsx` - 在告警操作列增加“投递历史” deeplink。
- `frontend/src/pages/Alerts.test.tsx` - 适配 Router 上下文并覆盖 deeplink 按钮可见性。
- `frontend/src/pages/Deliveries.tsx` - 实现查询页、筛选表单、分页表格和失败证据 Drawer。
- `frontend/src/pages/Deliveries.test.tsx` - 覆盖 `alert_id` query 初始化和 viewer 只读证据展示。

## Decisions Made

- delivery 历史页继续走 `capabilityViewConfig`，因为本 plan 只满足 `OPER-01` / `OPER-04` 的只读语义，不引入任何恢复动作按钮。
- 筛选状态以 URL query 为真源，这样告警页 deeplink、分页和刷新后的页面回显都保持一致。
- 失败证据直接展示 `attempts`、`final_failure_summary`、`rendered_payload_snapshot`、`channel_snapshot` 和 `route_snapshot`，避免把维护者再推回后端日志。

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] 适配告警页测试到新的 Router 依赖**
- **Found during:** Task 1 (定义前端 delivery 读面合同并接入受保护路由与 deeplink)
- **Issue:** `Alerts.tsx` 新增 `useNavigate` 后，既有 `Alerts.test.tsx` 在无 Router 上下文下直接崩溃。
- **Fix:** 把测试包进 `MemoryRouter`，并补一条“投递历史”按钮可见性断言。
- **Files modified:** `frontend/src/pages/Alerts.test.tsx`
- **Verification:** `pnpm --dir frontend test -- --run src/pages/Alerts.test.tsx`
- **Committed in:** `a2fd9e1`

**2. [Rule 3 - Blocking] 为 delivery 只读 API 响应补最小类型拆包**
- **Found during:** Task 2 (实现可筛选的投递历史页与失败证据详情回归)
- **Issue:** 运行时响应已被 axios 拦截器拆成 `data`，但 TypeScript 仍把 `deliveryApi` 返回值视为 `AxiosResponse`，导致页面 build 失败。
- **Fix:** 在 `frontend/src/api/client.ts` 新增 `unwrapData`，只对 delivery 只读 API 做最小类型收口。
- **Files modified:** `frontend/src/api/client.ts`
- **Verification:** `pnpm --dir frontend build`
- **Committed in:** `84043d5`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** 两项修正都直接服务于 Task 1/2 的可运行性与验证闭环，没有引入额外产品范围。

## Issues Encountered

- 计划里的 Vitest 路径写法以仓库根目录为基准，但 `pnpm --dir frontend` 会把工作目录切到 `frontend/`。执行时改用 `src/pages/...` 相对路径完成验证，未改动仓库脚本。

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 19 后续计划可以在当前 `/deliveries` 页上追加单条 `retry/replay` 操作入口，而不必重新设计历史查询骨架。
- 失败证据展示已经锁在前端只读页面，后续只需把恢复结果和审计关联接入同一详情面即可。

## Self-Check: PASSED

---
*Phase: 19-enable-safe-recovery-operations*
*Completed: 2026-05-01*
