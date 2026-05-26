# 风险分析报告

> 基于 graphify 知识图谱分析，生成日期：2026-05-26
> 调查与修复日期：2026-05-26

## 调查结果摘要

| 风险 | 严重度 | 确认? | 修复? | 说明 |
|------|--------|-------|-------|------|
| #1 T-16-01 | P0 | **已确认** | **已修复** | async panic 时 currentAlert 可能为 nil，trace_id 丢失 |
| #1 T-16-10 | P0 | **未确认** | N/A | 实际使用 JSON 解析 labels，不存在空格截断问题 |
| #2 | P0 | **已确认** | **已修复** | readPump 退出不调 hub.Unregister，死连接残留在 map 中 |
| #3 | P1 | 部分确认 | 未修复 | 已有 public 端点 (health/webhook)，JWT 轮换不支持是事实但影响有限 |
| #5 | P1 | 部分确认 | N/A | deliveryStats 数据竞争不存在（已移除），架构耦合是事实但非紧急 |
| #8 | P1 | **已确认** | **已修复** | template.Execute 无 panic recovery，保存时不验证模板语法 |
| #4 | P2 | 低风险 | N/A | Contains 断言模式合理，`//go:linkname` 是唯一真正的 INFERRED 依赖 |
| #6 | P2 | 低风险 | N/A | 典型 SPA 模式，AppLayout 未导入页面级 store |
| #7 | P3 | 低风险 | N/A | 孤立节点多为小工具函数，风险很低 |

## 风险优先级矩阵

| 风险 | 严重度 | 可能性 | 修复成本 | 优先级 |
|------|--------|--------|----------|--------|
| #1 开放安全威胁 | **致命** | 中 | 低 | **P0** |
| #2 WebSocket泄漏 | 高 | 高（已发生过） | 中 | **P0** |
| #3 认证单点 | 高 | 低 | 中 | P1 |
| #5 Delivery瓶颈 | 中高 | 中 | 高 | P1 |
| #8 Renderer桥接 | 中高 | 低 | 中 | P1 |
| #4 INFERRED验证缺口 | 中 | 高 | 低 | P2 |
| #6 前端内聚力 | 中 | 中 | 中 | P2 |
| #7 孤立节点 | 低 | 中 | 低 | P3 |

---

## 风险 1：开放安全威胁（T-16-01 / T-16-10）— 严重

> **调查结果**: T-16-01 已确认并修复，T-16-10 未确认（实际使用 JSON 解析，不存在空格截断）

### 威胁链路

**T-16-01: `async_panic` 缺少 `trace_id`** ✅ 已修复

实际代码链路（比报告更准确）：

```
HandleWebhook()  [返回 HTTP 200]
  → asyncRunner()(func() { processAlertNotificationsAsync(newAlerts) })  [goroutine]
    → processAlertNotificationsAsync()  [defer recover()]
      → ❌ panic 在 beforeSend hook 之前发生时，currentAlert 为 nil
        → baseAlertLogFields(nil, nil) 返回空 map → logNotification 无 trace_id
          → 无法关联到原始 webhook 请求
```

根因：panic recovery 依赖 beforeSend hook 更新 currentAlert，但 panic 可能发生在 hook 之前（如路由规则加载阶段）。

**修复方案**: 从 alerts 切片收集 batchTraceIDs，panic recovery 中当 currentAlert 为 nil 时使用 batchTraceIDs 而非空 map。

- 图中追踪：`WebhookHandler`(44度) → `StartDelivery()` → `executeRecoveredDelivery()` → `RecordAttempt()`/`MarkFailed()`
- 爆炸半径：**所有入站 webhook 告警**。一旦 async goroutine panic，对应的告警会静默丢失——调用方（监控系统）认为已投递，但实际上告警消失了
- 代码验证：`internal/delivery/service.go` 中 `executeRecoveredDelivery` 是异步执行的核心方法，panic recovery 后无法将失败关联回原始请求

**T-16-10: space-delimited key=value 解析** ❌ 未确认

> **调查结果**: 实际代码使用 `json.Unmarshal` 解析 labels（JSON 格式），不使用空格分隔的 key=value。
> `webhook.go` 中 `handleAlertmanager()` 直接将 JSON body 中的 `alerts[].labels` 反序列化为 `map[string]string`。
> 不存在值被空格截断的风险。此威胁基于过时或不正确的假设。

- 图中追踪：`WebhookHandler` → `AlertModel` → `routing.Matcher` → `Notifier`
- 爆炸半径：**所有依赖 label 匹配的告警路由**。如果 key 中含有 `=` 或 value 含空格，整条路由链断裂
- 关键连接：`routing.Matcher`（度数=6）同时连接 `Config` 和 `Notifier`，错误的 label 会导致路由决策完全偏移

### 缓解建议

- T-16-01：在 `StartDelivery` 入口将 `trace_id` 注入 `context.Context`，async goroutine 从 context 读取
- T-16-10：改用 JSON 解析 labels，或至少用 `strings.SplitN(keyval, "=", 2)` 保护 value 中的 `=`

---

## 风险 2：WebSocket 泄漏 — 高 ✅ 已修复

> **调查确认**: readPump() 退出时只调用 c.Close()，不调用 hub.Unregister(c)，导致死连接残留在 Hub 的 clients map 中。
> Pong 超时（60s）会关闭物理连接但仍不清理 map 条目。前端 Dashboard 的 useEffect 在 logout 时确实关闭 WebSocket，但 server 端缺少主动清理。
>
> **修复方案**: 在 readPump() 的 defer 中增加 c.hub.Unregister(c) 调用，确保连接断开时客户端从 Hub map 中移除。

### 影响链路

```
前端 useUserStore (31度, 最高连接度前端节点)
  → App.tsx / AppLayout 组件
    → WebSocket 连接建立 (Hub.Register())
      → ❌ 用户登出/Token过期时 useUserStore 重置
        → ❌ WebSocket 连接未正确关闭
          → Hub 中残留 client channel
            → Hub.Broadcast() 继续向已断开的 client 发送
              → goroutine 泄漏 + 内存泄漏
                → 长时间运行后 OOM
```

### 图中证据

- `Hub`(度数=5) → `Client` → `Register`/`Unregister` → `Broadcast`
- `useUserStore`(度数=31) 连接到 `Login.tsx`、`AppLayout`、`AppHeader`、`Alerts.tsx`、`Dashboard.tsx` 等 5+ 个页面
- 图中存在 "WebSocket Leak Fix" 节点，说明**这个问题已经被发现过一次但可能修复不彻底**

### 级联影响

1. **内存泄漏**：Hub 的 `clients map` 持续增长，每个泄漏的 client 占用一个 goroutine + channel
2. **广播风暴**：`Broadcast()` 遍历所有 clients（包括死连接），发送超时阻塞，影响活跃 client 的推送延迟
3. **告警延迟**：Dashboard 页面的实时告警更新变慢，运维看到的告警状态是过时的

### 缓解建议

- Hub 中增加 `pong` 超时检测，自动清理无响应 client
- `useUserStore` 重置时显式调用 WebSocket close
- 添加 Hub 连接数 metrics，告警超阈值

---

## 风险 3：认证中间件单点故障 — 高

> **调查结果**: 部分确认。已有 public 端点（`/health`、`/ready`、`/webhook/*`），不存在"所有 API 不可用"的问题。
> 但 JWT 轮换确实不支持（单密钥），密钥轮换会导致所有已登录用户立即失效。
> 当前优先级低，建议后续添加 JWT 密钥轮换支持（新旧密钥并存期）。

### 影响链路

```
所有 API 请求
  → Router.SetupRoutes()
    → AuthMiddleware() [割点 - articulation point]
      → JWT 解析验证
        → RequireCapability() [二割点]
          → ❌ JWT 密钥过期/配置错误
            → 全部 API 不可用
              → 前端所有页面白屏
                → 运维无法通过 UI 操作
```

### 图中证据

- `AuthMiddleware`(度数=8) 是图的割点——移除后认证子图与其他节点完全断开
- 所有 handler 依赖链：`AuthMiddleware` → `RequireCapability` → `AlertHandler`/`DeliveryHandler`/`ConfigHandler`/`UserHandler`/`WebhookHandler`...
- `RequireCapability`(度数=6) 又是二割点，串联了 RBAC 能力检查

### 级联影响

1. **全量 API 不可用**：JWT 验证失败 = 所有端点 401，零降级能力
2. **无 graceful degradation**：没有 public endpoint 可以用来健康检查或紧急修复
3. **配置变更即爆炸**：修改 JWT secret 后所有已登录用户同时失效，产生登录风暴

### 缓解建议

- 分离 public/protected 路由，健康检查和 webhook 入站不受 JWT 保护
- JWT 密钥轮换支持 overlap period（新旧密钥同时有效）
- 添加 auth bypass 的 emergency mode（受限操作，仅用于恢复）

---

## 风险 4：INFERRED 边 / `contains()` 验证缺口 — 中高

> **调查结果**: 低风险。Contains 断言（140处）主要用于结构化日志输出检查，多数足够具体。
> 唯一真正的 INFERRED 依赖是 `delivery_test.go` 中的 `//go:linkname`（2处），访问 authz 包的未导出变量。

### 影响链路

```
contains() (度数=48, 最大 INFERRED 边节点)
  → 41条 INFERRED 连接 (未验证)
    → 测试代码依赖关系不确定
      → 重构时可能漏改测试
        → 测试通过但业务逻辑已损坏
          → 部署到生产后才发现 bug
```

### 图中证据

- `contains()`：48 条边，其中大量 INFERRED，连接到 `Setup()`、`recordAudit()` 等测试辅助函数
- `NewRenderer()`：INFERRED 边占比高，模板渲染器的实际调用关系未确认
- `recordAudit()`：INFERRED 边连接到多个 handler，审计日志的触发点不确定

### 级联影响

1. **重构盲区**：INFERRED 关系不是代码中显式声明的，重构工具（IDE rename、go mod tidy）不会更新它们
2. **测试假阳性**：`contains()` 的大量 INFERRED 测试断言可能在重构后仍然通过（断言太宽泛），掩盖真实 bug
3. **架构漂移**：随着代码演进，INFERRED 关系与实际代码逐渐脱节，图的参考价值下降

### 缓解建议

- 对 INFERRED 边做人工审计，将确认的改为 EXTRACTED，错误的删除
- 测试中用 `assert.Equal` 替代 `assert.Contains`，使断言精确化
- 代码评审时重点审查 `contains()` 相关的测试

---

## 风险 5：Delivery Service 架构瓶颈 — 中高

> **调查结果**: 部分确认。deliveryStats 数据竞争不存在（当前代码已移除内存中的 stats map，全部基于 DB）。
> 架构耦合是事实（Service 跨越多个关注点），但当前代码量可控，暂不需要拆分。

### 影响链路

```
NewService() (betweenness=0.191, 跨7个社区)
  → StartDelivery() / RecordAttempt() / MarkDelivered() / MarkFailed()
    → RetryDelivery() → executeRecoveredDelivery()
      → ReplayDelivery() → executeRecoveredDelivery()
        → EscalationChecker.sendEscalationNotification()
          → Notifier.SendToChannel()
            → 所有通知渠道 (飞书/钉钉/企微/Webhook)
```

反向依赖：

```
DeliveryHandler (API层)
  → Service.ListDeliveries() / GetDeliveryByID()
    → DeliveryRecoveryModel
      → NotificationDeliveryModel
        → AlertModel
          → DataSource
```

### 图中证据

- `NewService()` 跨越 7 个社区：Delivery、Escalation、Notifier、Recovery、Handler、Model、Stats
- betweenness 0.191 意味着约 19% 的最短路径经过它——是系统最大的交通枢纽
- 所有通知路径（飞书/钉钉/企微/Webhook）的最后一公里都经过 Service

### 级联影响

1. **修改 Service 接口的代价极高**：任何方法签名变更会同时影响 API 层、定时任务、升级检查器
2. **性能瓶颈**：所有通知投递串行经过 Service，高并发时成为吞吐量天花板
3. **故障放大**：Service 的一个 bug（如 retry 逻辑错误）会影响所有通知渠道，而不是只影响一个渠道
4. **测试困难**：7 个社区的依赖使得 mock Service 的 setup 非常复杂

### 缓解建议

- 拆分 Service 为 DeliveryOrchestrator + DeliveryRepository + DeliveryNotifier
- 引入事件总线（channel/pubsub），让 Escalation 和 Stats 订阅事件而非直接调用
- Service 层增加 circuit breaker，单个渠道失败不影响其他渠道

---

## 风险 6：前端 App Shell 内聚力低 — 中

### 影响链路

```
Community 0 (Frontend App Shell, cohesion=0.05)
  → App.tsx, AppLayout, AppHeader, AppSidebar
    → 大量跨社区边连接到:
      → Alerts 页面 (社区20)
      → Channels 页面
      → Dashboard 页面
      → 用户相关页面
    → ❌ 布局组件与页面逻辑耦合
      → 修改侧边栏影响页面渲染
        → 修改 Header 影响路由行为
```

### 图中证据

- 社区 0 内聚力 0.05（几乎为零），说明内部节点之间几乎没有紧密的内部连接
- 跨社区边数量远大于内部边——这是一个"交通枢纽"而非"功能模块"
- `useUserStore`(度数=31) 同时属于 App Shell 又被所有页面使用

### 级联影响

1. **布局变更的回归风险**：改侧边栏菜单可能意外影响页面状态
2. **Bundle 膨胀**：App Shell 引入了页面级依赖，首屏加载变慢
3. **团队协作冲突**：多人同时修改前端时容易产生 merge conflict

### 缓解建议

- 将 `useUserStore` 从 App Shell 中抽出，作为独立的 auth context
- AppLayout 只做布局骨架（header + sidebar + outlet），不导入任何页面级 store
- 用 React.lazy 做页面级代码分割

---

## 风险 7：孤立/弱连接节点 — 低中

### 图中证据

- 约 15 个度数 ≤1 的节点
- 代码孤立点：主要是小型工具函数或测试辅助
- 文档孤立点：部分设计规范与代码实现无交叉引用

### 具体风险节点

| 类型 | 风险 |
|------|------|
| 代码孤立项 | 可能是死代码，增加了维护负担和 bundle 大小 |
| 文档孤立项 | 设计规范可能与实际实现脱节，产生误导 |
| 测试孤立项 | 测试存在但未连接到被测代码，可能是无效测试 |

### 缓解建议

- 代码孤立项：检查是否被反射/dynamic import 使用，否则删除
- 文档孤立项：在代码中添加注释引用对应规范，建立双向链接
- 定期运行 `/graphify --update` 跟踪孤立节点的变化趋势

---

## 风险 8：Renderer 桥接告警模型与通知管道 — 中 ✅ 已修复

> **调查确认**: 两处关键发现：
> 1. `Renderer.Render()` 和 `WebhookHandler.renderNotificationPreview()` 中的 `template.Execute()` 无 panic recovery，
>    模板执行 panic 会导致 goroutine 崩溃 + 通知丢失
> 2. `DataSource.Validate()` 不验证模板语法，损坏模板在保存时不报错，运行时才导致通知失败
>
> **修复方案**:
> 1. 在 `Renderer.Render()` 和 `renderNotificationPreview()` 中添加 `defer recover()` 将 panic 转为 error
> 2. 在 `DataSource.Validate()` 中添加 `text/template` 和 `html/template` 语法验证
> 3. 在 `UpdateDataSource` handler 中增加 `Validate()` 调用（原 CreateDataSource 已有）

### 影响链路

```
AlertModel (告警数据)
  → routing.Matcher (路由决策)
    → Renderer (betweenness=0.223, 最高!)
      → ❌ 模板渲染失败
        → 通知内容为空或乱码
          → Notifier.SendToChannel() 发送空消息
            → 飞书/钉钉显示空告警
              → 运维看不到告警内容
                → 等同于告警丢失
```

反向链路：

```
Pipeline Stage: Render Notification
  → Renderer
    → AlertModel (labels作为模板变量)
      → ❌ label缺失/类型错误
        → 模板执行 panic
          → ❌ 如果没有 recover
            → 整个通知 goroutine 崩溃
```

### 图中证据

- `Renderer` betweenness = 0.223，是**全图最高**的桥接节点
- 桥接社区：Alert Domain Models ↔ Template Rendering ↔ Async/Send Pipeline ↔ Alert Pipeline Stages
- 图中 4 个社区通过 Renderer 互联——它是数据模型到通知系统的唯一通道

### 级联影响

1. **单点故障**：Renderer 是告警从"数据"到"人"的唯一桥梁，它坏了等于整个系统坏了
2. **模板注入风险**：用户定义的告警标签直接注入模板，恶意 label 可能导致模板执行异常
3. **格式耦合**：所有通知渠道共享同一个 Renderer，模板变更可能对某个渠道造成意外影响

### 缓解建议

- Renderer 中增加 `template.Execute` 的 panic recovery
- 为每个通知渠道使用独立模板，避免一个模板错误影响所有渠道
- 添加模板预览/验证 API，在保存时就校验模板语法
- Renderer 增加缓存失效策略，避免模板更新后仍使用旧缓存
