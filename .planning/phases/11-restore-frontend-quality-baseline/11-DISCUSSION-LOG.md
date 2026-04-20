# Phase 11: Restore Frontend Quality Baseline - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the auto-selected discuss outcomes.

**Date:** 2026-04-20
**Phase:** 11-restore-frontend-quality-baseline
**Areas discussed:** Cleanup Scope, Hook And State Stability, Type Tightening Strategy, Verification Baseline
**Mode:** auto

---

## Cleanup Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Baseline Recovery | 清零现有 lint error，并同步收口关键 warning 噪音，让当前代码库恢复 green 基线 | ✓ |
| Errors Only | 只修复会让 lint 失败的 error，其余 warning 继续保留 | |
| Broad Refactor | 借机做大范围页面和类型重构 | |

**User's choice:** Auto-selected recommended option: 以当前仓库的基线恢复为目标，同时处理 error 和关键 warning。
**Notes:** Phase 11 的 success criteria 明确要求 `pnpm lint` 通过，并收口高风险 hook 依赖和无效变量问题，因此不能只修 error。

---

## Hook And State Stability

| Option | Description | Selected |
|--------|-------------|----------|
| Honor Real Dependencies | 补齐真实依赖并维持现有 Zustand action 模式，不通过禁用 lint 规避问题 | ✓ |
| Disable Rule Per Page | 在页面局部压制 `exhaustive-deps` | |
| Rewrite Store Layer | 为了满足 hooks 规则重写状态管理方式 | |

**User's choice:** Auto-selected recommended option: 补齐真实依赖并保持最小改动。
**Notes:** 当前问题集中在初始化取数页面，最合理的处理是让依赖与现有 store action 契约对齐，而不是扩大到架构改造。

---

## Type Tightening Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Narrow Existing Types | 用 `unknown`、已有业务类型和表格入参类型收口 `any` | ✓ |
| Ignore Warning Debt | 保留 `any`，只追求 lint 通过 | |
| Full Type Redesign | 借此重建前端领域类型体系 | |

**User's choice:** Auto-selected recommended option: 用最窄、最可维护的现有类型替代明显 `any`。
**Notes:** 这符合“修质量基线而不改产品能力”的 phase 边界，也能为后续 CI 门禁提供更稳定输入。

---

## Verification Baseline

| Option | Description | Selected |
|--------|-------------|----------|
| Re-run Lint Test Build | lint 清理后重新验证 `pnpm lint`、前端 test 和 build | ✓ |
| Lint Only | 只检查 lint，不验证测试和构建 | |
| Add New Test Targets | 在本 phase 扩大验证矩阵 | |

**User's choice:** Auto-selected recommended option: 重新验证 lint、test、build 三条现有链路。
**Notes:** 这是 Phase 11 对 `FEQ-03` 的直接映射，也为 Phase 12 的 CI 接入提供本地先验。

---

## the agent's Discretion

- 先修复 `DataSources.tsx` error，再按页面批量处理 warning
- 对 `any` 的具体替代类型按局部上下文选择最窄方案

## Deferred Ideas

- 配置类页面的统一抽象重构
- 更系统的前端测试增强
- 更严格的共享表单和 schema 类型体系
