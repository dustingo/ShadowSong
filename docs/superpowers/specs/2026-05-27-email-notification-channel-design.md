# Email 通知渠道设计

## 概述

为告警系统新增 `email` 通知渠道类型，支持通过 SMTP 协议发送 HTML 格式告警邮件。SMTP 服务器配置为全局系统配置，收件人在路由规则中动态指定。

## 决策记录

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 发送方式 | SMTP 直连 | 最通用，任何企业邮箱都支持 |
| 收件人配置位置 | 路由规则中动态指定 | 不同路由发给不同人，灵活 |
| 邮件格式 | HTML | 样式可控，展示丰富 |
| SMTP 配置位置 | 全局系统配置 | 所有邮件渠道共享，避免重复配置 |
| 测试发送收件人 | 手动填写 | 安全，不会误发 |

## 数据模型

### SmtpConfig（全局 SMTP 配置）

单行配置表，全局只有一条记录。

```go
type SmtpConfig struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Host      string    `gorm:"size:128;not null" json:"host"`
    Port      int       `gorm:"not null;default:465" json:"port"`
    Username  string    `gorm:"size:128;not null" json:"username"`
    Password  string    `gorm:"size:256" json:"password"`
    FromAddr  string    `gorm:"size:128;not null" json:"from_addr"`
    FromName  string    `gorm:"size:64" json:"from_name"`
    TLS       bool      `gorm:"default:true" json:"tls"`
    Enabled   bool      `gorm:"default:true" json:"enabled"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Channel（新增 email 类型）

Channel.Type 新增枚举值 `email`。email 渠道的 Config JSONB 结构：

```json
{
  "from_name": "告警系统"
}
```

仅一个可选字段，覆盖全局 FromName。核心 SMTP 配置在 SmtpConfig 中。

### RouteRule（新增 recipients 字段）

```go
Recipients datatypes.JSON `gorm:"type:jsonb" json:"recipients"`
// 存储: ["a@b.com", "c@d.com"] 或 null
```

当路由规则关联的渠道包含 email 类型时，recipients 必填。

## API

### SMTP 配置

| 方法 | 路径 | Handler | 权限 | 说明 |
|------|------|---------|------|------|
| GET | `/api/v1/smtp-config` | GetSmtpConfig | viewConfig | 返回配置，password 字段 mask |
| PUT | `/api/v1/smtp-config` | UpdateSmtpConfig | manageConfig | 全量替换，password 为 `{"masked": true}` 时保留原值 |
| POST | `/api/v1/smtp-config/test` | TestSmtpConfig | manageConfig | 测试发送，请求体 `{"recipients": ["test@example.com"]}` |

### 渠道测试（已有接口变更）

`POST /api/v1/channels/:id/test` 增加可选请求体字段：

```json
{
  "recipients": ["test@example.com"]
}
```

当渠道类型为 email 时，recipients 必填。其他渠道类型忽略该字段。

### 路由规则（已有接口变更）

Create/Update RouteRule 的请求体新增 `recipients` 字段：

```json
{
  "name": "...",
  "channel_ids": [1, 2],
  "recipients": ["ops@example.com", "dev@example.com"],
  ...
}
```

## EmailSender 实现

### 发送流程

1. 路由匹配后，`findMatchedChannels()` 返回 `(Channel, RouteRule)` 对
2. 将 `routeRule.Recipients` 注入 `data["recipients"]`
3. `SendToChannel()` 分派到 `EmailSender.Send()`
4. `EmailSender` 从数据库读取全局 SmtpConfig（每次发送时读取，不做内存缓存。告警发送频率不高，DB 读取开销可接受，且保证配置变更即时生效）
5. 检查 SmtpConfig 存在且 Enabled，否则返回错误
6. 检查 recipients 非空，否则返回错误
7. 构造 MIME 邮件，TLS 连接 SMTP 服务器发送

### 邮件构造

- From: `FromName <FromAddr>`（渠道 config 的 from_name 优先，否则用全局 SmtpConfig.FromName）
- To: recipients 列表
- Subject: 渲染后的 title，含非 ASCII 字符时使用 RFC 2047 编码（`=?UTF-8?B?...?=`）
- Content-Type: `text/html; charset=UTF-8`
- Body: 渲染后的 content（HTML，由 output_template 生成）
- 超时: 10 秒，与现有渠道一致

### 错误分类

- 连接超时、认证失败、TLS 握手失败 → retryable，走 3 次重试
- 收件人地址无效（SMTP 550 拒收）→ non-retryable
- SMTP 未配置或未启用 → non-retryable，返回 "SMTP 服务未配置"
- 收件人为空 → non-retryable，返回 "邮件收件人为空"

复用现有 `IsRetryableSendError()` 分类逻辑。

## 前端

### SMTP 设置页面

在设置页面新增"邮件服务"Tab：
- 表单字段：Host、Port（默认465）、Username、Password、发件人地址、发件人显示名、TLS 开关、启用开关
- Password 字段用 password 输入框，编辑时显示占位符
- "测试连接"按钮：点击后弹出对话框输入收件人地址，发送测试邮件

### 渠道管理页面（email 类型）

- 渠道类型下拉新增 `email`（"邮件"）
- 选中 email 后，config 表单只显示"发件人显示名"一个可选字段
- 测试渠道按钮：点击后弹出对话框，要求输入测试收件人地址

### 路由规则页面

- 编辑路由规则时，当关联的渠道列表中包含 email 类型，自动显示"收件人"输入区
- 使用 TagInput 组件，支持输入多个邮箱地址
- 校验：关联渠道含 email 时 recipients 必填，每个地址做基础格式校验（包含 `@`）

## 边界情况

- **SmtpConfig 并发安全**：全局单行配置，更新用 `WHERE id = 1` 全量替换，无并发问题
- **Password mask**：GET 接口返回 `{"masked": true}`，PUT 接口若 password 为 `{"masked": true}` 则保留原值
- **创建 email 渠道时 SMTP 未配置**：前端显示警告提示，但不阻止创建（渠道可先创建，后续配置 SMTP）
- **邮件内容编码**：Subject 使用 RFC 2047 编码，HTML Body 使用 UTF-8 charset
