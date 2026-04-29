# Milestones

## v1.3 Notification Reliability and Observability (Shipped: 2026-04-29)

**Phases completed:** 4 phases, 11 plans, 21 tasks

**Key accomplishments:**

- Webhook requests now mint one server-side trace_id that persists on new alerts and survives dedup, Redis handoff, and async notification entry without changing alert_id or fingerprint semantics
- Webhook lifecycle logs now expose one searchable trace across ingest, persist or dedup, Redis handoff, route matching, and notification entry with explicit Redis outcome metadata, and the constructor remains nil-safe for router tests without Redis
- Webhook notifications now classify transient send failures centrally, retry them inside one bounded async window, and emit attempt-level plus terminal-failure logs with stable trace fields
- Phase 15 verification now locks the three retry outcomes and exact three-attempt exhaustion behavior in tests, and documents terminal failure as a log-only landing zone that operators can trace back through the Phase 14 lifecycle
- Canonical webhook alert-path logging now uses one shared key=value writer with stable trace/channel envelopes and structured `matched_channels` and `mode` fields across route and send failure paths
- Webhook logging contract is now pinned by field-level handler regressions and a phase verification artifact that shows how to walk a `trace_id` from `terminal_failure` back to `ingest`
- Webhook async panic recovery now keeps trace_id, alert_id, fingerprint, source, and concrete channel metadata on the emitted async_panic log line
- Webhook alert-path logs now preserve space-containing field values through quoted key=value serialization, parser-aligned regressions, and refreshed Phase 16 verification truth
- Low-risk verification entrypoints now use current alert-flow and console-baseline naming across scripts, tests, and repo-owned reference maps.
- README and planning truth surfaces now describe the live game-ops alert platform, while historical milestone and review docs are explicitly framed as archive context.
- Maintainers now have one evergreen alert-path runbook plus phase-local verification, UAT, and security artifacts grounded in the verified Phase 14-16 evidence chain.

---

当前维护者入口请以 `README.md`、`.planning/PROJECT.md`、`.planning/ROADMAP.md` 和 Phase 14-17 真相工件为准。本文件保留已发布里程碑事实，尤其是 `v1.0 AI Removal Complete` 的历史上下文，但这些标题不代表当前运行叙事或推荐命名。

## v1.0 AI Removal Complete (Shipped: 2026-04-10)

**Phases completed:** 4 phases, 12 plans, 18 tasks

**Key accomplishments:**

- Go 后端已移除 AI 配置读取、AI 路由装配与 AI 运行时文件，服务启动仅依赖常规告警系统配置
- 移除后端 AI 专用持久化字段、迁移表与测试通知文案，保留非 AI 告警主链路的数据结构与运行路径
- 新增单元回归测试与一条可直接执行的后端闭环脚本，证明移除 AI 后核心告警链路仍可实际跑通
- 认证后的 React 壳层已不再暴露 AI 页面入口，`AIAssistant` 页面文件、barrel 导出和 `/ai` 路由簇已被整体删除
- Dashboard、告警卡片和告警列表已去除 AI 操作与分析展示，只保留现有运维告警处理入口
- 前端共享 API/类型层已去除 AI 合同，登录页与浏览器标题完成非 AI 文案收口，并通过生产构建与风险修正验证
- README、代码审查入口与本地 `.env` 基线已统一为无 AI 的游戏运维告警系统表述，并清除启动配置中的 AI 专用键
- Phase 03 已补齐可复用的前后端无 AI 验证入口，并记录了实际通过的验证证据

---

## v1.1 Enterprise Access Control (Shipped: 2026-04-15)

**Phases completed:** 5 phases, 15 plans, 12 summary-reported tasks

**Key accomplishments:**

- 建立了统一的 `admin` / `operator` / `viewer` 角色真源、JWT principal 和 capability matrix 基线
- 收紧了用户管理边界，补齐了账号禁用、强制改密和旧会话失效控制
- 对配置写接口与告警动作完成了后端权限收口，并落地了持久化审计日志
- 前端完成了权限感知菜单、页面、按钮、只读提示和角色矩阵验证
- 收口了 `PROJECT` 真相文档、前端测试 warning 噪音以及 capability-only authz seam

---

## v1.2 Alert Pipeline Hardening (Shipped: 2026-04-21)

**Phases completed:** 4 phases, 8 plans, 8 summary-reported tasks

**Key accomplishments:**

- 收紧了 `/ws/alerts` 实时告警访问面，WebSocket 握手已要求 JWT 且受来源 allowlist 限制
- 前端质量基线已恢复为 green，`pnpm lint`、`pnpm test -- --run` 与 `pnpm build` 均已打通
- 仓库已新增 GitHub Actions 质量门禁，覆盖后端测试与前端 lint/test/build
- README 与前端包名等低风险入口已继续对齐到当前非 AI 告警系统基线
- webhook 异步通知 goroutine 已补 panic recover，失败日志具备告警/渠道上下文
- 通知链路可靠性路径已纳入直接 Go 测试和后端自动门禁

---

## Current Narrative

- 当前系统能力基线由已发版的 `v1.3 Notification Reliability and Observability` 定义，重点是通知可靠性、告警链路可观测性、维护者 runbook 和真相分层
- `AI Removal Complete` 只保留为 v1.0 已发版历史事实；后续维护与运行入口不应再把该标题当作当前系统定位
- 深层运行时历史命名例如 `go.mod` module path 与 JWT issuer 仍是 deferred runtime contracts，应在当前真源中标注为暂缓迁移，而不是在里程碑摘要里误读为推荐现状
