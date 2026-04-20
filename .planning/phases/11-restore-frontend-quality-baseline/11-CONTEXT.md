# Phase 11: Restore Frontend Quality Baseline - Context

**Gathered:** 2026-04-20
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只解决前端现有代码库里的 lint 红线和会持续制造噪音的高风险质量问题，让 `pnpm lint`、前端测试和生产构建重新回到可持续维护的 green 基线。范围不包含新增前端功能、重做页面交互、引入新的测试框架或接入 CI，这些分别属于后续 phase。

</domain>

<decisions>
## Implementation Decisions

### Cleanup Scope
- **D-01:** Phase 11 以“恢复当前前端质量基线”为目标，必须清零现有 lint error，并同步收口会长期制造噪音的 hook 依赖、未使用变量和关键页面 `any` 类型警告。
- **D-02:** 清理范围优先覆盖当前活跃页面与基础壳层，包括 `App.tsx`、告警页、配置页、用户页和共享类型；不把本 phase 扩展成大规模类型系统重构。

### Hook And State Stability
- **D-03:** 对 `react-hooks/exhaustive-deps` 的处理优先采用“补齐真实依赖并复用稳定 store action 引用”的方式，而不是用禁用规则或空依赖数组继续压 lint。
- **D-04:** 如果某些页面的初始化加载逻辑会因为依赖补齐而触发额外请求，应优先通过现有 Zustand action 形态做最小稳定化修正，而不是重写状态架构。

### Type Tightening Strategy
- **D-05:** 对明显的 `any` 噪音优先收口为 `unknown`、`React.Key`、表格 render 入参类型或已有业务类型，避免继续保留无边界的逃逸类型。
- **D-06:** 这轮只做“足够支撑 lint 和维护性”的类型补强；不为了消灭 warning 去引入与后端契约不一致的新前端领域模型。

### Verification Baseline
- **D-07:** Phase 11 完成后必须重新验证 `pnpm lint`、现有前端测试和 `pnpm build`，确保清理 lint 时没有引入新的运行时风险。
- **D-08:** 验证重点是保持现有告警、配置、用户与值班等主路径页面可构建、可测试，而不是新增视觉或交互验收项。

### the agent's Discretion
- 具体按文件拆分修复顺序可由后续计划阶段决定，只要先消除 error 再收口高风险 warning。
- 对 `any` 的具体替代类型可按上下文选最窄可维护方案，不要求一次把所有类型定义抽象到最完美。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — 当前里程碑约束、非 AI 产品命名和 brownfield 前提
- `.planning/REQUIREMENTS.md` — Phase 11 对应的 `FEQ-01` 到 `FEQ-03`
- `.planning/ROADMAP.md` — Phase 11 的目标、依赖与 success criteria
- `.planning/STATE.md` — 当前阶段位置与后续 Phase 12/13 的顺序关系

### Architecture And Conventions
- `AGENTS.md` — 当前仓库的 GSD 工作流约束与项目级说明
- `frontend/.eslintrc.cjs` — 前端 lint 规则、warning/error 门槛与 hooks 规则来源
- `frontend/package.json` — `lint`、`test`、`build` 命令真相
- `.planning/codebase/CONVENTIONS.md` — 前端命名、hooks、状态管理与错误处理约定
- `.planning/codebase/STACK.md` — React + Vite + Zustand + Ant Design 当前技术栈事实

### Live Code To Reuse
- `frontend/src/App.tsx` — 当前壳层菜单与 capability 入口，存在未使用 capability 变量
- `frontend/src/stores/configStore.ts` — 配置类页面依赖的 Zustand action 定义与稳定引用来源
- `frontend/src/stores/alertStore.ts` — 告警页和 Dashboard 的共享动作与日志模式
- `frontend/src/pages/Alerts.tsx` — 缺失 `fetchAlerts` 依赖和表格 render `any`
- `frontend/src/pages/DataSources.tsx` — 当前 lint error 所在页，也是模板预览与表单复杂度最高页面
- `frontend/src/pages/RouteRules.tsx` — 未使用变量、hook 依赖与排序逻辑的代表性页面
- `frontend/src/types/index.ts` — 多处 `any` 来源的共享类型入口

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `useConfigStore` / `useAlertStore`：页面大多直接从 Zustand store 取 action，修复 hooks 依赖时应优先复用这些现有 action，而不是改成新数据层。
- `getApiErrorMessage`：多个页面已统一用它处理接口报错，类型收口时可以围绕这个错误边界做最小改动。
- `frontend/src/types/index.ts`：已经是共享类型出口，适合补齐表格 render 和表单请求相关的窄类型。

### Established Patterns
- 页面初始化普遍在 `useEffect` 中直接调用 store action，这意味着 lint 修复要同时关注 action 引用稳定性和重复请求风险。
- 代码库允许 `@typescript-eslint/no-explicit-any` 作为 warning，但本 phase 已明确要求收口关键噪音，所以应优先处理活跃页面里的明显 `any`。
- 前端代码风格以最小包装和直接表单提交为主，适合做局部修正，不适合在本 phase 引入大规模抽象。

### Integration Points
- `App.tsx` 的 capability 菜单决定多个页面入口，未使用变量修复不能破坏现有权限感知导航。
- `DataSources.tsx`、`RouteRules.tsx`、`OnDuty.tsx`、`Silences.tsx`、`Users.tsx` 都依赖 `useConfigStore` 或 `useUserStore`，hooks 修复需要与 store API 保持一致。
- `frontend/package.json` 中的 `lint`、`test` 和 `build` 是 Phase 11 的主验证面，也会成为后续 Phase 12 CI 门禁的直接输入。

</code_context>

<specifics>
## Specific Ideas

- 当前 `pnpm lint` 的硬失败来自 `frontend/src/pages/DataSources.tsx`，因此执行顺序应先消除这里的 `no-useless-escape` error，再统一处理 warnings。
- 本 phase 追求的是“安静、可持续、可验证”的前端基线，不追求顺手重构出新的页面架构。
- 如果某些 warning 明显来自已经废弃或未接线的交互残留，优先删除死代码而不是强行保留。

</specifics>

<deferred>
## Deferred Ideas

- 为前端引入更严格的 schema 校验或统一表单模型
- 重构配置类页面的共享抽象和复用组件
- 新增更完整的前端测试体系或更细粒度测试覆盖要求

None of the above belongs in Phase 11 scope.

</deferred>

---

*Phase: 11-restore-frontend-quality-baseline*
*Context gathered: 2026-04-20*
