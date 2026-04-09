# Phase 02: Remove Frontend AI Surfaces - Research

**Researched:** 2026-04-09
**Domain:** Existing React + Vite + Ant Design frontend cleanup
**Confidence:** HIGH

## Summary

Phase 2 is a brownfield frontend cleanup, not a redesign. The existing SPA already has a clear shell boundary in `frontend/src/App.tsx`, a route/page barrel in `frontend/src/pages/index.ts`, shared alert rendering in `frontend/src/components/AlertCard.tsx`, and centralized transport/types in `frontend/src/api/client.ts` and `frontend/src/types/index.ts`. The cleanest implementation path is to remove AI surfaces at those boundaries, then run a focused static verification pass so no orphaned route, type, or API reference survives. [VERIFIED: codebase grep 2026-04-09]

Phase 1 already removed backend AI runtime and confirmed `/api/v1/ai/chat` returns 404, so Phase 2 should treat any remaining AI client calls as dead code and remove them instead of attempting graceful fallbacks. Frontend behavior that must remain stable is the non-AI alert loop: navigation, dashboard stats/trend, alert acknowledgement, quick silence, and existing configuration pages. [VERIFIED: .planning/phases/01-remove-backend-ai-runtime/01-03-SUMMARY.md] [VERIFIED: .planning/ROADMAP.md]

**Primary recommendation:** Plan the work as three focused plans: shell/page entry cleanup, dashboard/alert rendering cleanup, and API/type/title cleanup with frontend build/lint verification. [VERIFIED: roadmap phase breakdown + current file ownership]

## Project Constraints (from AGENTS.md)

- Respect the existing Go + React + Vite + Ant Design stack; do not introduce a new UI library or technical migration. [VERIFIED: AGENTS.md]
- This is a brownfield repo with existing uncommitted changes; plans must avoid reverting or broad-touching unrelated files. [VERIFIED: AGENTS.md]
- Frontend cleanup must keep navigation, routes, menu, types, and API calls self-consistent so no runtime chain breaks remain. [VERIFIED: AGENTS.md]
- Documentation- and branding-related frontend text touched in this phase must reflect the non-AI alert system state. [VERIFIED: AGENTS.md]
- Follow existing frontend conventions: named exports, barrel exports, Zustand for shared state, Ant Design `message.*` feedback, Prettier formatting, and ESLint rules in `frontend/.eslintrc.cjs`. [VERIFIED: AGENTS.md] [VERIFIED: .planning/codebase/CONVENTIONS.md]

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| FEAI-01 | 前端导航、路由和页面中不再出现 AI 助手、AI 聊天、AI 日志或 AI 静默推荐入口 | `App.tsx`, `pages/index.ts`, `pages/AIAssistant.tsx`, and `@ant-design/icons` imports define all current UI entry points |
| FEAI-02 | 告警详情、列表和仪表盘中不再展示 AI 分析、AI 根因、AI 建议或“问 AI”操作 | `Dashboard.tsx`, `Alerts.tsx`, and `components/AlertCard.tsx` contain the rendering and actions that surface AI content |
| FEAI-03 | 前端 API 客户端、类型定义和状态使用中不再依赖 AI 相关请求或字段 | `api/client.ts`, `types/index.ts`, and page imports identify the remaining AI transport/type contracts |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| React | `^18.2.0` | Page/component runtime | Existing SPA is already structured around React function components and hooks. [VERIFIED: frontend/package.json] |
| React Router DOM | `^6.21.1` | Route table and auth gating | `/login`, `/`, `/alerts`, `/datasources`, `/channels`, `/routes`, `/silences`, `/onduty`, `/ai` are all wired in `App.tsx`; removing `/ai` should follow current router patterns. [VERIFIED: frontend/package.json] [VERIFIED: frontend/src/App.tsx] |
| Ant Design | `^5.12.8` | Shell, forms, cards, tables, feedback | Current screens rely on `Layout`, `Menu`, `Card`, `Alert`, `Modal`, `Form`, and `message`; cleanup should preserve these patterns rather than introduce replacements. [VERIFIED: frontend/package.json] [VERIFIED: frontend/src/App.tsx] [VERIFIED: frontend/src/pages/Dashboard.tsx] |
| Zustand | `^4.4.7` | Shared alert/config/user state | Shared alert flows already run through `useAlertStore`, so UI cleanup should not add new cross-page state layers. [VERIFIED: frontend/package.json] [VERIFIED: frontend/src/stores/alertStore.ts] |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| axios | `^1.6.5` | Typed API wrapper layer | Keep request removal centralized in `frontend/src/api/client.ts` and `frontend/src/api/auth.ts`. [VERIFIED: frontend/package.json] [VERIFIED: frontend/src/api/client.ts] |
| dayjs | `^1.11.19` | Timestamp formatting | Preserve current alert-card/list formatting while removing AI content. [VERIFIED: frontend/package.json] [VERIFIED: frontend/src/components/AlertCard.tsx] |
| echarts / echarts-for-react | `^5.4.3` / `^3.0.2` | Dashboard trend visualization | Dashboard chart remains in scope after AI removal; avoid coupling chart work to AI cleanup. [VERIFIED: frontend/package.json] [VERIFIED: frontend/src/pages/Dashboard.tsx] |
| TypeScript compiler | `^5.3.3` | Static contract verification | `pnpm build` is the strongest existing frontend gate because no dedicated frontend test runner is configured. [VERIFIED: frontend/package.json] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Existing Ant Design shell cleanup | UI redesign / shadcn migration | Rejected because Phase 2 is explicitly a removal phase and the approved `UI-SPEC` forbids redesign. [VERIFIED: 02-UI-SPEC.md] |
| Centralized API/type cleanup | Inline local hacks in pages | Rejected because dead AI types/endpoints would remain reachable and violate FEAI-03. [VERIFIED: frontend/src/api/client.ts] [VERIFIED: frontend/src/types/index.ts] |

## Architecture Patterns

### Recommended Project Structure

```text
frontend/src/
├── App.tsx                 # route shell + menu + product title
├── pages/                  # route-level pages, including Dashboard and AIAssistant
├── components/             # shared alert rendering such as AlertCard
├── api/client.ts           # grouped domain-specific API wrappers
├── stores/                 # Zustand state orchestration
└── types/index.ts          # frontend contracts mirrored from backend JSON
```

### Pattern 1: Remove at the boundary where the capability enters the UI

**What:** Delete AI entry points where they are introduced, not only where they are rendered.
**When to use:** Route/menu/page removal (`App.tsx`, `pages/index.ts`, `pages/AIAssistant.tsx`) and transport/type removal (`api/client.ts`, `types/index.ts`).
**Example:** Remove the `/ai` route and `AIAssistant` export in the same plan so the barrel and router cannot drift. [VERIFIED: frontend/src/App.tsx] [VERIFIED: frontend/src/pages/index.ts]

### Pattern 2: Shared render cleanup through reusable components first

**What:** Remove AI-only affordances from shared components such as `AlertCard` before or alongside page-level call sites.
**When to use:** Dashboard and alert-list cleanup where the same AI data/action appears in multiple screens.
**Example:** `AlertCard` currently owns the `问 AI` button and AI summary/root-cause/suggestion block; page code only wires callbacks into it. [VERIFIED: frontend/src/components/AlertCard.tsx]

### Pattern 3: Collapse layout instead of replacing removed content

**What:** When AI blocks are removed, preserve existing Ant Design spacing and success/error/empty-state patterns; do not substitute speculative filler content.
**When to use:** Dashboard modal removal, alert table row cleanup, shell label cleanup.
**Example:** The approved `UI-SPEC` requires the dashboard to remain anchored on connection warnings and active alerts after AI modal removal. [VERIFIED: 02-UI-SPEC.md]

### Anti-Patterns to Avoid

- **Page-only cleanup:** Removing `AIAssistant.tsx` but leaving route, menu, or barrel exports behind will create compile or runtime breakage. [VERIFIED: frontend/src/App.tsx] [VERIFIED: frontend/src/pages/index.ts]
- **Backend-fallback thinking:** The backend already removed AI runtime; leaving AI buttons that call `/api/v1/ai/*` will degrade into user-visible errors instead of graceful behavior. [VERIFIED: 01-03-SUMMARY.md]
- **Type drift:** Deleting page UI without deleting `Alert.ai_*` fields and `aiApi` methods leaves FEAI-03 incomplete and encourages future accidental re-use. [VERIFIED: frontend/src/types/index.ts] [VERIFIED: frontend/src/api/client.ts]

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Shared alert fetch/update state | Per-page ad hoc duplicated alert logic | Existing `useAlertStore` methods and `alertApi` wrappers | Current architecture already centralizes fetch, ack, quick silence, and WebSocket update flows there. [VERIFIED: frontend/src/stores/alertStore.ts] |
| UI feedback after cleanup | Custom banners or bespoke state wrappers | Existing Ant Design `message`, `Alert`, `Spin`, `Modal`, `Card` patterns | Preserves behavior and aligns with approved UI contract. [VERIFIED: frontend/src/pages/Dashboard.tsx] [VERIFIED: 02-UI-SPEC.md] |
| Route/menu registry | New dynamic navigation abstraction | Existing static `menuItems` array and explicit `<Route>` declarations | The current cleanup scope is small and explicit; abstraction would enlarge risk. [VERIFIED: frontend/src/App.tsx] |

**Key insight:** The safest Phase 2 work is subtraction along existing boundaries, not refactoring architecture while deleting AI. [VERIFIED: codebase structure + roadmap scope]

## Common Pitfalls

### Pitfall 1: Orphaned route or export after page deletion

**What goes wrong:** `AIAssistant.tsx` is deleted or ignored, but `App.tsx` and `pages/index.ts` still import/export it.
**Why it happens:** UI entry points are split across shell and barrel files.
**How to avoid:** Treat `App.tsx`, `pages/index.ts`, and `pages/AIAssistant.tsx` as one change cluster. [VERIFIED: frontend/src/App.tsx] [VERIFIED: frontend/src/pages/index.ts]
**Warning signs:** TypeScript import errors or unresolved symbol errors during `pnpm build`.

### Pitfall 2: Dashboard cleanup leaves dead local state and dependencies

**What goes wrong:** AI modal JSX is removed but `ReactMarkdown`, `aiApi`, `aiModalVisible`, `currentAlert`, or `aiResponse` state remains.
**Why it happens:** `Dashboard.tsx` mixes shell data, WebSocket behavior, and AI modal behavior in one file.
**How to avoid:** Remove imports, state, handlers, and modal markup together, then re-run type check. [VERIFIED: frontend/src/pages/Dashboard.tsx]
**Warning signs:** Unused state/import warnings, build failures, or dead buttons remaining in alert cards.

### Pitfall 3: Shared type cleanup breaks page assumptions

**What goes wrong:** `ai_*` fields are removed from `Alert` but page/component code still reads them.
**Why it happens:** AI fields are spread across `Alerts.tsx` and `AlertCard.tsx`, not isolated in one place.
**How to avoid:** Pair `types/index.ts` cleanup with grep-based call-site cleanup before concluding FEAI-03 is done. [VERIFIED: frontend/src/types/index.ts] [VERIFIED: grep results 2026-04-09]
**Warning signs:** TypeScript errors for missing `ai_summary`, `ai_root_cause`, or `ai_suggestions`.

### Pitfall 4: Branding cleanup deferred too late

**What goes wrong:** Route/page AI features are removed, but visible titles still say “AI 告警系统”.
**Why it happens:** Product title appears in multiple frontend files and can be missed if the phase is framed narrowly as route deletion.
**How to avoid:** Include touched-title cleanup inside the shell/API-type cleanup plan, at least for `frontend/src/App.tsx`, `frontend/src/pages/Login.tsx`, and `frontend/index.html`. [VERIFIED: grep results 2026-04-09]
**Warning signs:** Successful build with stale product copy still visible.

## Code Examples

Verified patterns from the current codebase:

### Route and menu ownership

```tsx
const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '告警大盘' },
  { key: '/alerts', icon: <AlertOutlined />, label: '告警管理' },
  { key: '/datasources', icon: <DatabaseOutlined />, label: '数据源' },
]
```

Source: `frontend/src/App.tsx` [VERIFIED: codebase read 2026-04-09]

### Shared alert action ownership

```tsx
<AlertCard
  key={alert.alert_id}
  alert={alert}
  showActions={true}
  onAck={handleAck}
  onQuickSilence={handleQuickSilence}
  onAskAI={handleAskAI}
/>
```

Source: `frontend/src/pages/Dashboard.tsx` [VERIFIED: codebase read 2026-04-09]

### API wrapper grouping

```ts
export const alertApi = {
  list: (params?: { page?: number }) => apiClient.get('/alerts', { params }),
  ack: (id: string, data: { comment: string }) => apiClient.post(`/alerts/${id}/ack`, data),
}
```

Source: `frontend/src/api/client.ts` [VERIFIED: codebase read 2026-04-09]

## Open Questions (RESOLVED)

1. **Should Phase 2 introduce a new UI design system while removing AI?**
   - RESOLVED: No. The approved `02-UI-SPEC.md` locks the phase to existing Ant Design patterns only. [VERIFIED: 02-UI-SPEC.md]

2. **Does frontend AI cleanup need backend compatibility shims?**
   - RESOLVED: No. Phase 1 already removed backend AI routes and verified `/api/v1/ai/chat=404`; frontend should delete the dead calls instead of shimming them. [VERIFIED: 01-03-SUMMARY.md]

3. **What is the strongest available verification gate for this phase?**
   - RESOLVED: `pnpm build` plus `pnpm lint` in `frontend/` because no frontend test runner is configured. [VERIFIED: frontend/package.json] [VERIFIED: .planning/codebase/STACK.md]

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | frontend build/lint | ✓ | `v22.17.0` | — |
| pnpm | frontend build/lint | ✓ | `10.28.2` | npm could install dependencies, but project standard is pnpm |

**Missing dependencies with no fallback:**
- None for planning. [VERIFIED: local CLI probe 2026-04-09]

**Missing dependencies with fallback:**
- None identified. [VERIFIED: local CLI probe 2026-04-09]

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | Preserve existing auth guard in `App.tsx` and token handling in `userStore` / axios interceptors |
| V3 Session Management | yes | Keep 401 redirect and `localStorage` token clearing behavior intact |
| V4 Access Control | yes | Do not bypass `RequireAuth`; only remove `/ai` from protected routes |
| V5 Input Validation | no | This phase does not add new user input surfaces |
| V6 Cryptography | no | No crypto changes in this frontend cleanup phase |

### Known Threat Patterns for this phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Removing a protected route but leaving a reachable menu/link target | Elevation of Privilege | Delete route registration and menu entry together in `App.tsx` |
| Frontend continues calling removed backend AI endpoints | Denial of Service | Delete `aiApi` wrapper and all call sites instead of retrying dead endpoints |
| Title/copy cleanup accidentally touches auth flow labels or navigation semantics | Tampering | Limit branding cleanup to AI-specific strings and preserve existing module names and auth controls |

## Sources

### Primary (HIGH confidence)

- Local codebase reads on 2026-04-09:
  - `frontend/src/App.tsx`
  - `frontend/src/pages/Dashboard.tsx`
  - `frontend/src/pages/Alerts.tsx`
  - `frontend/src/components/AlertCard.tsx`
  - `frontend/src/pages/AIAssistant.tsx`
  - `frontend/src/pages/index.ts`
  - `frontend/src/api/client.ts`
  - `frontend/src/types/index.ts`
  - `frontend/package.json`
  - `frontend/.eslintrc.cjs`
  - `frontend/.prettierrc`
- Planning artifacts:
  - `.planning/ROADMAP.md`
  - `.planning/REQUIREMENTS.md`
  - `.planning/STATE.md`
  - `.planning/phases/01-remove-backend-ai-runtime/01-01-SUMMARY.md`
  - `.planning/phases/01-remove-backend-ai-runtime/01-02-SUMMARY.md`
  - `.planning/phases/01-remove-backend-ai-runtime/01-03-SUMMARY.md`
  - `.planning/phases/02-remove-frontend-ai-surfaces/02-UI-SPEC.md`

### Secondary (MEDIUM confidence)

- None needed; the phase is codebase-driven and does not depend on external ecosystem choices.

### Tertiary (LOW confidence)

- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - entirely verified from `frontend/package.json` and current source files
- Architecture: HIGH - phase boundaries are explicit in current shell/page/component/api/type files
- Pitfalls: HIGH - derived from direct AI references still present in the codebase and prior Phase 1 outcomes

**Research date:** 2026-04-09
**Valid until:** 2026-05-09
