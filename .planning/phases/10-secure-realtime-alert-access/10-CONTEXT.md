# Phase 10: Secure Realtime Alert Access - Context

**Gathered:** 2026-04-20
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只解决实时告警 WebSocket 访问面的安全边界问题：让 `/ws/alerts` 与现有 JWT / principal / capability 基线对齐，并对来源进行显式限制。范围不包含通知链路可靠性、CI、前端 lint 清理或新的实时产品功能设计。

</domain>

<decisions>
## Implementation Decisions

### Authentication Boundary
- **D-01:** `/ws/alerts` 复用现有 JWT 鉴权基线，要求客户端在握手请求中提供与 REST 接口一致的 Bearer token，而不是为实时链路设计新的认证协议。
- **D-02:** WebSocket 连接建立前必须校验用户仍然有效，包括账号禁用、强制改密和 token 失效规则；这些规则与 `middleware.JWTAuth` 保持一致，避免实时链路绕开现有账户控制。

### Authorization Scope
- **D-03:** Phase 10 先把实时告警流至少收口到“必须是已认证用户”，不在本 phase 引入新的细粒度 capability 分支或单独的 WebSocket 权限矩阵。
- **D-04:** 当前实时告警能力的授权目标是与现有“查看告警”产品能力保持一致，因此实现应优先复用已有 principal / role 语义，而不是重新发明独立角色判断。

### Origin Policy
- **D-05:** WebSocket `CheckOrigin` 不能继续全放行，必须改为显式 allowlist 策略，并由服务端配置驱动。
- **D-06:** 本地开发需要继续支持 `http://127.0.0.1:*` 和 `http://localhost:*` 一类前端来源；生产来源由配置补充，而不是硬编码在 handler 内。

### Compatibility And Rollout
- **D-07:** 不改变前端 `/ws` 代理入口和现有实时消费模型，优先保持现有连接路径可用，只在需要传 token 的位置做最小兼容改动。
- **D-08:** Phase 10 必须补验证，覆盖未授权拒绝、非法来源拒绝和合法客户端成功连接三类结果，确保安全收口不会把现有主流程打断。

### the agent's Discretion
- 握手 token 的具体承载位置可由研究和计划阶段决定，只要满足“沿用现有 JWT 契约、前端改动最小、Gin/Gorilla WebSocket 可稳定实现”这三个约束即可。
- 配置字段的具体命名和解析方式可由后续阶段决定，但必须落在现有 `internal/config/config.go` 体系内。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — 当前里程碑目标、系统约束和 brownfield 前提
- `.planning/REQUIREMENTS.md` — Phase 10 对应的 `RTAL-01` 到 `RTAL-03`
- `.planning/ROADMAP.md` — Phase 10 目标、依赖和 success criteria
- `.planning/STATE.md` — 当前里程碑位置与 phase 编号连续性

### Architecture And Conventions
- `.planning/codebase/ARCHITECTURE.md` — `/ws/alerts` 所在层次、数据流和当前未集成事实
- `.planning/codebase/CONVENTIONS.md` — Go handler / middleware / config 代码风格与错误处理约定
- `.planning/codebase/STACK.md` — Gin、Gorilla WebSocket、Vite 代理等当前技术栈事实

### Live Code To Reuse
- `internal/router/router.go` — 当前 `/ws/alerts` 路由挂载方式和其他受保护接口的鉴权模式
- `internal/handlers/websocket.go` — 当前 upgrader、连接生命周期和客户端注册逻辑
- `internal/middleware/auth.go` — 现有 JWT / principal / disabled / force-reset / token-invalid-before 校验逻辑
- `frontend/src/stores/alertStore.ts` — 前端实时状态承载点 `wsConnected`
- `frontend/vite.config.ts` — `/ws` 前端开发代理和现有连接路径约束

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `middleware.JWTAuth`：已经封装了 Bearer token 解析、用户状态校验、principal 注入，是实时链路应复用的安全真源。
- `auth.JWT` 与 `models.User`：现有 token 校验、账号禁用、强制改密和旧会话失效控制都已存在，不需要为 WebSocket 重写账户规则。
- `useAlertStore.setWsConnected`：前端已经有实时连接状态位，兼容握手增强时无需新增 store 结构。

### Established Patterns
- 受保护 REST 路由统一在 `router.go` 中挂 `middleware.JWTAuth(jwtAuth, db)`，说明 Phase 10 应优先保持“router 负责接线、handler 负责业务”的模式。
- 配置目前统一由 `internal/config/config.go` 从环境变量读取，allow origins 也应沿用这一模式，不应在 handler 中散落常量。
- 错误返回在后端以即时 JSON 或连接拒绝为主，没有统一错误翻译层，所以 WebSocket 拒绝路径需要保持简单直白。

### Integration Points
- `/ws/alerts` 当前直接从 `router.Setup(...)` 调到 `wsHandler.HandleAlerts(c)`，这里是接入鉴权和 origin 校验的主入口。
- `websocket.Upgrader.CheckOrigin` 当前为包级变量，后续需要让它能拿到配置或由 handler 封装出带配置的 upgrader。
- 前端开发环境走 `frontend/vite.config.ts` 的 `/ws` 代理，所以 origin 策略需要兼容 Vite 开发服务器来源。

</code_context>

<specifics>
## Specific Ideas

- 这期目标是“最小改动收口风险面”，不是重做实时架构。
- 当前架构分析已经确认 `BroadcastAlert` 还没有接入 webhook 主链路，因此本 phase 不应被扩展为“补齐完整实时推送功能”。
- 如果实现细节需要在 `Authorization` header 不稳定与 query/cookie 之间取舍，优先选兼容现有浏览器与代理链路、同时不削弱服务端校验的一条。

</specifics>

<deferred>
## Deferred Ideas

- WebSocket 消息广播正式接入 webhook 入库链路
- 实时告警流的细粒度 capability 或按角色裁剪消息内容
- 更完整的实时事件总线或跨节点广播架构

None of the above belongs in Phase 10 scope.

</deferred>

---

*Phase: 10-secure-realtime-alert-access*
*Context gathered: 2026-04-20*
