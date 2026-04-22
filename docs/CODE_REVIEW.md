# Go 代码审查报告

> 历史审查快照（Historical Snapshot）
>
> 时间范围：本报告反映的是 2026-03-16 的一次静态审查结果，不代表 2026-04-22 之后的当前运行指导。
> 当前维护者真源请优先查看：
> - `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`
> - `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`
> - `.planning/phases/16-standardize-alert-path-logging/16-VERIFICATION.md`
> - Phase 17 maintainer runbook / alert-path operations 文档（后续计划产出）
>
> 尤其是本报告中关于全仓日志迁移、旧高风险问题清单的内容，只能作为历史背景，不应替代当前 phase verification 与运行说明。

**项目**: 游戏运维告警系统 (shadowsongAI)
**审查日期**: 2026-03-16
**审查范围**: 全部Go源文件

---

## 静态分析结果

| 检查项 | 结果 |
|--------|------|
| go vet | ✅ 通过 |
| go build | ✅ 通过 |
| go test | ✅ 通过 |

---

## 问题汇总表

### 🔴 CRITICAL (必须修复)

| 序号 | 文件 | 行号 | 问题类型 | 描述 |
|------|------|------|----------|------|
| 1 | `config.go` | 51-55 | 安全风险 | JWT Secret 已在上一版本修复为必填，但需确认无默认值 |
| 2 | `postgres.go` | 82 | 安全风险 | 使用密码生成器生成随机密码，但仍需确认安全性 |
| 3 | `router.go` | 19 | 安全风险 | CORS 只允许 `http://127.0.0.1`，生产环境需可配置 |

### 🟠 HIGH (应该修复)

| 序号 | 文件 | 行号 | 问题类型 | 描述 |
|------|------|------|----------|------|
| 1 | `user.go` | 186 | 权限漏洞 | 非管理员可修改其他用户角色 |
| 2 | `webhook.go` | 184 | Goroutine | 异步处理无错误传播机制，panic会导致崩溃 |
| 3 | `alert.go` | 153 | 错误处理 | 创建 SilenceRule 错误被忽略 |
| 4 | `alert.go` | 52 | 错误处理 | 查询错误后未 return，继续执行 |
| 5 | `alert.go` | 27-40 | 输入验证 | Query参数未做验证直接用于SQL查询 |
| 6 | `webhook.go` | 523, 767 | 错误处理 | json.Unmarshal 错误被忽略 |
| 7 | `config.go` | 52 | 用户体验 | JWT缺失时直接os.Exit，缺少友好提示 |

### 🟡 MEDIUM (建议改进)

| 序号 | 文件 | 行号 | 问题类型 | 描述 |
|------|------|------|----------|------|
| 1 | 多个文件 | - | 代码注释 | 缺少导出函数的 Godoc 注释 |
| 2 | `postgres.go` | 40-42 | 性能 | 连接池参数可优化 |
| 3 | `redis.go` | 20 | 最佳实践 | 使用带超时的 context |
| 4 | `alert.go` | 44-45 | 最佳实践 | 分页参数未限制最大值 |
| 5 | `webhook.go` | 444-458 | 代码简化 | 自定义正则匹配可使用标准库 |
| 6 | `user.go` | 66, 121, 134 | 代码复用 | 重复的密码哈希清除逻辑可提取 |
| 7 | `middleware/auth.go` | 29 | 代码优化 | 使用 strings.Cut 更简洁 |
| 8 | 多处 | - | 日志规范 | 使用 fmt.Println/log.Printf，建议统一使用 slog |
| 9 | `websocket.go` | 102-109 | 资源泄漏 | goroutine未正确退出机制 |
| 10 | `webhook.go` | 1003-1013 | 未使用 | CleanAlerts 函数未被调用 |

---

## 新增问题详细说明

### 1. CRITICAL - CORS 配置不灵活

**文件**: `internal/router/router.go:19`

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1")
```

**问题**: 硬编码只允许本机访问，生产环境需要可配置的允许源列表。

**建议**:
```go
// config.go 添加
type ServerConfig struct {
    Port            string
    Mode            string
    AllowedOrigins []string
}

// router.go
allowedOrigins := cfg.Server.AllowedOrigins
if len(allowedOrigins) == 0 {
    allowedOrigins = []string{"http://127.0.0.1"}
}
origin := c.Request.Header.Get("Origin")
// 验证 origin 是否在允许列表中
```

---

### 2. HIGH - 用户权限漏洞

**文件**: `internal/handlers/user.go:186`

```go
if currentUserID != uint(id) && currentUserRole != "admin" {
    c.JSON(http.StatusForbidden, gin.H{"error": "cannot update other users"})
    return
}
// ...
if input.Role != "" {
    user.Role = input.Role  // 任何登录用户都可以修改自己的角色！
}
```

**问题**: 代码逻辑允许任何用户修改自己的角色为 admin。

**建议**:
```go
if input.Role != "" {
    if currentUserRole != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "only admin can change role"})
        return
    }
    user.Role = input.Role
}
```

---

### 3. HIGH - 错误后未 return

**文件**: `internal/handlers/alert.go:52`

```go
if err := query.Find(&alerts).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    // 缺少 return，会继续执行并返回空列表
}
```

**问题**: 错误处理后未 return，导致后续代码继续执行。

**建议**: 添加 `return`

---

### 4. HIGH - Goroutine 缺少 recover

**文件**: `internal/handlers/webhook.go:184`

```go
go h.processAlertNotifications(newAlerts)
```

**问题**: 异步处理 panic 会导致程序崩溃。

**建议**:
```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("panic in processAlertNotifications: %v", r)
        }
    }()
    h.processAlertNotifications(newAlerts)
}()
```

---

### 5. MEDIUM - 日志不统一（历史建议）

多个文件使用 `fmt.Println` 和 `log.Printf`，建议统一使用 Go 1.21+ 的 `slog` 包。

说明：这是 2026-03-16 的历史建议。当前仓库在 Phase 16 已明确采用告警主链路 canonical `key=value` logging contract，而不是把本阶段扩展为 repo-wide `slog` 迁移；请以 `16-VERIFICATION.md` 为当前真源。

**示例**:
```go
// 替换前
log.Printf("Starting server on port %s", port)

// 替换后
slog.Info("starting server", "port", port)
```

---

### 6. MEDIUM - WebSocket goroutine 泄漏

**文件**: `internal/handlers/websocket.go:102-109`

```go
go func() {
    for {
        time.Sleep(30 * time.Second)
        err := conn.WriteMessage(websocket.PingMessage, []byte{})
        if err != nil {
            break  // 只退出 for 循环，未关闭 goroutine
        }
    }
}()
```

**问题**: 当连接断开时，goroutine 不会自动退出。

**建议**: 添加 context 或 signal 机制管理 goroutine 生命周期。

---

## 代码亮点

1. **良好的项目结构**: 清晰的 package 划分 (handlers, models, middleware, auth)
2. **GORM Hooks**: 使用 BeforeCreate/BeforeUpdate 进行数据验证
3. **JWT 实现**: 使用 golang-jwt 库，标准的 claims 结构
4. **密码安全**: 使用 bcrypt 加密
5. **Webhook 处理**: 完善的模板渲染和去重逻辑
6. **告警指纹**: SHA256 指纹生成实现合理
7. **通知渠道**: 多种通知渠道支持 (飞书、钉钉、企业微信、Webhook)
8. **运维告警闭环**: Webhook 接入、路由分发、通知发送与告警处理链路完整

---

## 改进建议优先级

### 立即修复 (CRITICAL)
- [x] JWT Secret 必填 - 已修复
- [ ] 配置化 CORS 允许源
- [ ] 确认密码生成安全性

### 尽快修复 (HIGH)
- [ ] 修复 user.go 权限漏洞
- [ ] 添加错误后 return
- [ ] 为 goroutine 添加 panic recovery
- [ ] 输入参数验证

### 后续优化 (MEDIUM)
- [ ] 统一使用 slog 日志
- [ ] 添加导出函数 Godoc
- [ ] 优化分页参数限制
- [ ] 清理未使用代码

---

## 测试覆盖评估

- `internal/models/alert_test.go` - 存在基础测试
- **建议**: 增加更多单元测试，特别是 handlers 层

---

## 总结

> 历史定位说明：本总结反映审查当时的风险排序。当前是否仍然适用，必须结合 Phase 14-16 verification 证据与 Phase 17 维护者文档一起判断，不能单独当作运行变更清单。

| 级别 | 数量 |
|------|------|
| CRITICAL | 3 (1已修复) |
| HIGH | 7 |
| MEDIUM | 10 |

**整体评分**: 7.5/10

**优点**:
- 代码结构清晰，模块化良好
- 功能完整，告警处理流程完善
- 安全基础良好 (JWT, bcrypt)

**需要改进**:
- 安全配置需可配置化
- 错误处理需更完善
- 日志需统一
- 需增加测试覆盖

**建议**: 优先修复 HIGH 级别问题，确保系统安全性和稳定性。
