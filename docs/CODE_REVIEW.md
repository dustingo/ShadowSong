# Go 代码审查报告

**项目**: AI Alert System (shadowsongAI)
**审查日期**: 2026-03-12
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
| 1 | `config.go` | 77 | 安全风险 | JWT Secret 使用默认值 "default-secret-change-in-production" |
| 2 | `postgres.go` | 95 | 安全风险 | 硬编码默认管理员密码 "admin123" |
| 3 | `router.go` | 18 | 安全风险 | CORS 允许所有来源 (`*`) |

### 🟠 HIGH (应该修复)

| 序号 | 文件 | 行号 | 问题类型 | 描述 |
|------|------|------|----------|------|
| 1 | `handlers/user.go` | 90 | 逻辑错误 | Token 验证逻辑有误，`parts` 被赋值为字符串长度而非分割结果 |
| 2 | `webhook.go` | 184 | 潜在问题 | Goroutine 没有错误传播机制 |
| 3 | `alert.go` | 150 | 错误处理 | 创建 SilenceRule 错误被忽略 |
| 4 | `alert.go` | 27-40 | 输入验证 | Query 参数未做验证直接用于 SQL 查询 |
| 5 | `handlers/ai.go` | 74 | 错误处理 | 查询 Alert 错误被忽略 |
| 6 | `webhook.go` | 523 | 错误处理 | json.Unmarshal 错误被忽略 |

### 🟡 MEDIUM (建议改进)

| 序号 | 文件 | 行号 | 问题类型 | 描述 |
|------|------|------|----------|------|
| 1 | 多个文件 | - | 代码注释 | 缺少导出函数的 Godoc 注释 |
| 2 | `postgres.go` | 40-41 | 性能 | 未配置连接池超时和连接生命周期 |
| 3 | `redis.go` | 20 | 最佳实践 | 使用 context.Background() 而非带超时的 context |
| 4 | `handlers/alert.go` | 44-45 | 最佳实践 | 分页参数未限制最大值 |
| 5 | `webhook.go` | 444-458 | 代码简化 | 自定义正则匹配函数可以移除（Go 有标准库） |
| 6 | `handlers/user.go` | 66, 121, 134 | 最佳实践 | 重复的密码哈希清除逻辑可以提取为方法 |
| 7 | `middleware/auth.go` | 29 | 最佳实践 | 使用 strings.Cut 更简洁 |

---

## 详细问题说明

### 1. CRITICAL - JWT Secret 默认值

**文件**: `internal/config/config.go:77`

```go
JWTSecret: getEnv("JWT_SECRET", "default-secret-change-in-production"),
```

**问题**: 生产环境使用默认 secret 会导致 JWT 安全风险。

**建议**:
```go
func Load() *Config {
    jwtSecret := getEnv("JWT_SECRET", "")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    // ...
}
```

---

### 2. CRITICAL - 硬编码默认密码

**文件**: `internal/database/postgres.go:95`

```go
log.Println("Default admin user created: admin / admin123")
```

**问题**: 硬编码密码在代码库中，且日志会输出密码。

**建议**: 使用环境变量或配置生成随机密码。

---

### 3. CRITICAL - CORS 允许所有来源

**文件**: `internal/router/router.go:18`

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```

**问题**: 生产环境不应允许所有来源。

**建议**:
```go
allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
// 验证 origin 是否在允许列表中
```

---

### 4. HIGH - Token 验证逻辑错误

**文件**: `internal/handlers/user.go:90-91`

```go
parts := len(authHeader)
if parts < 8 || authHeader[:7] != "Bearer " {
```

**问题**: `parts` 被赋值为字符串长度（int），而非分割结果。应该是 `strings.Split`。

```go
parts := strings.SplitN(authHeader, " ", 2)
if len(parts) != 2 || parts[0] != "Bearer" {
```

---

### 5. HIGH - Goroutine 错误处理

**文件**: `internal/handlers/webhook.go:184`

```go
go h.processAlertNotifications(newAlerts)
```

**问题**: 异步处理无法返回错误，且 panic 会导致程序崩溃。

**建议**: 使用 errgroup 或通道传递错误。

---

### 6. HIGH - 忽略的错误

**文件**: `internal/handlers/alert.go:150`

```go
h.db.Create(&silence)  // 错误被忽略
```

**建议**:
```go
if err := h.db.Create(&silence).Error; err != nil {
    log.Printf("Warning: failed to create silence rule: %v", err)
}
```

---

### 7. MEDIUM - Query 参数未验证

**文件**: `internal/handlers/alert.go:27-40`

```go
if severity := c.Query("severity"); severity != "" {
    query = query.Where("severity = ?", severity)  // 直接使用未验证
}
```

**建议**: 添加输入验证，确保 severity 在允许的值范围内。

---

## 代码亮点

1. **良好的项目结构**: 清晰的 package 划分 (handlers, models, middleware, auth)
2. **GORM Hooks**: 使用 BeforeCreate/BeforeUpdate 进行数据验证
3. **JWT 实现**: 使用 golang-jwt 库，标准的 claims 结构
4. **密码安全**: 使用 bcrypt 加密
5. **WebHook 处理**: 完善的模板渲染和去重逻辑

---

## 改进建议优先级

### 立即修复 (CRITICAL)
- [ ] JWT Secret 改为必填环境变量
- [ ] 移除硬编码密码
- [ ] 配置化 CORS

### 尽快修复 (HIGH)
- [ ] 修复 user.go:90 逻辑错误
- [ ] 添加错误处理
- [ ] 输入参数验证

### 后续优化 (MEDIUM)
- [ ] 添加导出函数 Godoc
- [ ] 优化分页参数
- [ ] 统一错误处理模式

---

## 总结

| 级别 | 数量 |
|------|------|
| CRITICAL | 3 |
| HIGH | 6 |
| MEDIUM | 7 |

**建议**: 优先修复 CRITICAL 和 HIGH 级别问题，确保系统安全性和稳定性后再进行代码优化。
