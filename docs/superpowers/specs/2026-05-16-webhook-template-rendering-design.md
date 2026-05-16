# Channel Webhook 请求体模板渲染

## 问题

Channel webhook 的"请求体模板"字段当前是死文本，不做 Go 模板渲染。用户填写的 `{{.alert_name}}` 等占位符会原样发出，OutputTemplate 渲染的 `title`/`content` 也被丢弃。

## 方案

请求体模板支持 Go 模板渲染，数据上下文包含 OutputTemplate 渲染结果 + 告警原始字段。

## 改动

### 1. Sender 接口签名变更

```go
// 之前
type Sender interface {
    Send(title, content string) error
}

// 之后
type Sender interface {
    Send(title, content string, data map[string]interface{}) error
}
```

飞书/钉钉/企微的 Sender 实现忽略 `data` 参数。

### 2. WebhookSender.Send 渲染逻辑

JSON content_type:
- Template 不为空 → 用 `template.Renderer` 渲染，数据上下文 = `data`，渲染结果作为请求体
- Template 为空 → 保持现有行为，发 `{"title": ..., "content": ...}`

form-urlencoded content_type:
- Template 不为空 → 渲染模板，结果作为 form 请求体
- Template 为空 → 保持现有行为，用 `content` 作为请求体

### 3. 数据上下文

复用 `buildNotificationRenderContext` 的字段，额外加入 `title` 和 `content`：

```
title, content          — OutputTemplate 渲染结果
alert_id, alert_name, severity, severity_code, severity_raw, message, source, status, trigger_time, labels
event                   — 原始事件数据
alert                   — 嵌套告警对象 (id, name, severity, ...)
route_name              — 匹配的路由规则名
```

### 4. 渲染器

WebhookSender 内部持有 `*template.Renderer` 实例，在 `NewWebhookSender` 中创建。

### 5. 错误处理

- 模板解析失败 → 回退到默认行为（发 `{"title": ..., "content": ...}`），返回错误
- 模板执行失败 → 同上

### 6. 调用链变更

`sendChannelNotification` 构建 data 上下文并传给 sender：

```go
data := h.buildNotificationRenderContext(alert, matchedRouteRule)
data["title"] = title
data["content"] = content
sender(channel, title, content, data)
```

`SendToChannel` 签名同步更新，透传 `data`。

### 7. 测试

- Template 为空时保持现有行为
- Template 包含 `{{.title}}`/`{{.content}}` 时正确渲染
- Template 包含 `{{.alert_name}}`/`{{.severity}}`/`{{.event.xxx}}` 时正确渲染
- Template 包含 `{{.labels.xxx}}` 时正确渲染
- Template 语法错误时回退到默认行为
- form-urlencoded + Template 时正确渲染
- form-urlencoded + 无 Template 时保持现有行为
