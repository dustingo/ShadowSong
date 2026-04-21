# Milestones

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
