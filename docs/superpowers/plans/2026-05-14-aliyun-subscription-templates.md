# 阿里云监控订阅模板实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完善阿里云监控订阅通知的模板支持，确保时间戳正确处理、labels 正确存储、测试覆盖完整。

**Architecture:** 修改 `internal/handlers/webhook.go` 中的 `parseTime` 函数支持毫秒时间戳，添加集成测试验证阿里云模板渲染流程。

**Tech Stack:** Go, text/template, GORM, PostgreSQL

---

## 文件结构

**修改文件:**
- `internal/handlers/webhook.go` - 修复 parseTime 函数
- `internal/handlers/webhook_test.go` - 添加阿里云模板测试

**已完成的修改（需要提交）:**
- `internal/handlers/webhook.go` - 已将 html/template 改为 text/template，已修复 labels 提取逻辑

---

### Task 1: 修复 parseTime 函数支持毫秒时间戳

**Files:**
- Modify: `internal/handlers/webhook.go:450-480`
- Test: `internal/handlers/webhook_test.go`

**背景:** 阿里云推送的 `time` 字段是毫秒时间戳（如 `1715665800000`），但当前 `parseTime` 函数只处理字符串类型，导致时间解析失败。

- [ ] **Step 1: 编写失败的测试用例**

在 `internal/handlers/webhook_test.go` 中添加测试：

```go
func TestWebhookHandlerParseTimeHandlesMillisecondTimestamp(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	// 测试毫秒时间戳
	result := handler.parseTime(float64(1715665800000))
	assert.NotEqual(t, "", result)
	assert.Contains(t, result, "2024") // 1715665800000 对应 2024 年

	// 测试 int64 类型
	result2 := handler.parseTime(int64(1715665800000))
	assert.NotEqual(t, "", result2)
	assert.Contains(t, result2, "2024")

	// 测试字符串时间（现有功能）
	result3 := handler.parseTime("2024-05-14T10:30:00Z")
	assert.Equal(t, "2024-05-14T10:30:00Z", result3)

	// 测试 nil 返回当前时间
	result4 := handler.parseTime(nil)
	assert.NotEqual(t, "", result4)
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
cd d:/goproject/shadowsongAI && go test -v -run TestWebhookHandlerParseTimeHandlesMillisecondTimestamp ./internal/handlers/
```

预期: FAIL - parseTime 不处理数字类型

- [ ] **Step 3: 修复 parseTime 函数**

修改 `internal/handlers/webhook.go` 中的 `parseTime` 函数：

```go
// parseTime 解析时间字符串或时间戳
func (h *WebhookHandler) parseTime(v interface{}) string {
	if v == nil {
		return time.Now().Format(time.RFC3339)
	}

	// 处理数字类型时间戳（毫秒）
	switch t := v.(type) {
	case float64:
		return time.Unix(0, int64(t)*int64(time.Millisecond)).Format(time.RFC3339)
	case int64:
		return time.Unix(0, t*int64(time.Millisecond)).Format(time.RFC3339)
	case int:
		return time.Unix(0, int64(t)*int64(time.Millisecond)).Format(time.RFC3339)
	}

	s, ok := v.(string)
	if !ok {
		return time.Now().Format(time.RFC3339)
	}

	// 尝试解析多种时间格式
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.Format(time.RFC3339)
		}
	}

	// 返回原始值
	return s
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
cd d:/goproject/shadowsongAI && go test -v -run TestWebhookHandlerParseTimeHandlesMillisecondTimestamp ./internal/handlers/
```

预期: PASS

- [ ] **Step 5: 提交**

```bash
cd d:/goproject/shadowsongAI && git add internal/handlers/webhook.go internal/handlers/webhook_test.go && git commit -m "fix: parseTime supports millisecond timestamps for Aliyun webhook"
```

---

### Task 2: 添加阿里云模板集成测试

**Files:**
- Modify: `internal/handlers/webhook_test.go`

**背景:** 验证阿里云订阅通知的完整渲染流程，确保 input template 正确渲染、labels 正确提取、output template 正确生成通知内容。

- [ ] **Step 1: 编写阿里云模板集成测试**

在 `internal/handlers/webhook_test.go` 中添加：

```go
func TestWebhookHandlerAliyunSubscriptionTemplate(t *testing.T) {
	db := newWebhookTestDB(t)
	handler, _ := newWebhookTestHandler(db)

	// 创建阿里云数据源
	require.NoError(t, db.Create(&models.DataSource{
		Name:          "aliyun_event_push",
		DisplayName:   "阿里云事件订阅",
		APIKey:        "test-api-key",
		Enabled:       true,
		InputTemplate: `{
			"alert_id": "{{.alert.dedupId}}",
			"alert_name": "{{.alert.meta.sysEventMeta.eventNameZh}}",
			"severity": "{{toSeverity .alert.meta.sysEventMeta.level}}",
			"message": "{{.userInfo.aliyunId}} {{.alert.meta.sysEventMeta.eventTime}} {{.alert.meta.sysEventMeta.serviceTypeZh}} {{.alert.meta.sysEventMeta.eventNameZh}}",
			"source": "aliyun_event_push",
			"status": "{{toStatus .alert.alertStatus}}",
			"trigger_time": "{{toTime .time}}",
			"labels": {
				"instance": "{{.alert.meta.sysEventMeta.instanceName}}",
				"region": "{{.alert.meta.sysEventMeta.regionId}}",
				"service_type": "{{.alert.meta.sysEventMeta.serviceTypeZh}}",
				"event_type": "{{.alert.meta.sysEventMeta.eventType}}",
				"product": "{{.alert.meta.sysEventMeta.product}}",
				"resource_id": "{{.alert.meta.sysEventMeta.resourceId}}"
			}
		}`,
		OutputTemplate:    `{"title": "[{{.severity}}] {{.alert_name}}", "content": "实例: {{.labels.instance}}\n区域: {{.labels.region}}"}`,
		GroupByLabels:     datatypes.JSON(`["alert_name","severity","source"]`),
		DeduplicateEnabled: true,
		DeduplicateWindow:  3600,
	}).Error)

	// 模拟阿里云推送数据
	payload := map[string]interface{}{
		"severity": "CRITICAL",
		"userInfo": map[string]interface{}{
			"aliyunId": "user123@example.com",
		},
		"strategyName": "ECS监控策略",
		"alert": map[string]interface{}{
			"alertStatus": "TRIGGERED",
			"dedupId":     "dedup-test-123",
			"meta": map[string]interface{}{
				"sysEventMeta": map[string]interface{}{
					"eventNameZh":    "实例状态改变",
					"instanceName":   "ecs-production-01",
					"regionId":       "cn-hangzhou",
					"serviceTypeZh":  "云服务器ECS",
					"eventType":      "StatusChange",
					"product":        "ECS",
					"resourceId":     "i-bp1234567890",
					"level":          "CRITICAL",
					"eventTime":      "2026-05-14 10:30:00",
				},
			},
		},
		"time": float64(1715665800000),
	}

	// 发送请求
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/aliyun_event_push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "test-api-key")

	w := httptest.NewRecorder()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/webhook/:source_name", handler.HandleWebhook)
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "success", resp["status"])
	assert.Equal(t, 1, resp["processed"])

	// 验证告警数据
	alerts := resp["alerts"].([]interface{})
	alert := alerts[0].(map[string]interface{})
	assert.Equal(t, "dedup-test-123", alert["alert_id"])
	assert.Equal(t, "实例状态改变", alert["alert_name"])
	assert.Equal(t, "P0", alert["severity"])
	assert.Equal(t, "firing", alert["status"])
	assert.Equal(t, "aliyun_event_push", alert["source"])

	// 验证 labels 结构
	labels := alert["labels"].(map[string]interface{})
	assert.Equal(t, "ecs-production-01", labels["instance"])
	assert.Equal(t, "cn-hangzhou", labels["region"])
	assert.Equal(t, "云服务器ECS", labels["service_type"])
}
```

- [ ] **Step 2: 运行测试验证通过**

```bash
cd d:/goproject/shadowsongAI && go test -v -run TestWebhookHandlerAliyunSubscriptionTemplate ./internal/handlers/
```

预期: PASS

- [ ] **Step 3: 提交**

```bash
cd d:/goproject/shadowsongAI && git add internal/handlers/webhook_test.go && git commit -m "test: add integration test for Aliyun subscription template"
```

---

### Task 3: 提交现有代码变更

**Files:**
- Modify: `internal/handlers/webhook.go` (已修改)
- Add: `docs/superpowers/specs/2026-05-14-aliyun-subscription-templates-design.md`

**背景:** 将之前的修改（html/template → text/template, labels 提取逻辑）提交到版本控制。

- [ ] **Step 1: 验证所有测试通过**

```bash
cd d:/goproject/shadowsongAI && go test ./internal/handlers/ -v
```

预期: 所有测试 PASS

- [ ] **Step 2: 提交代码变更**

```bash
cd d:/goproject/shadowsongAI && git add internal/handlers/webhook.go docs/superpowers/specs/2026-05-14-aliyun-subscription-templates-design.md && git commit -m "feat: add Aliyun subscription template support

- Change html/template to text/template to avoid HTML escaping
- Extract labels from template rendering result
- Fallback to raw data if labels is empty
- Add design spec document"
```

---

### Task 4: 运行完整测试套件

**Files:**
- None (验证任务)

- [ ] **Step 1: 运行所有测试**

```bash
cd d:/goproject/shadowsongAI && go test ./... -v
```

预期: 所有测试 PASS

- [ ] **Step 2: 确认构建成功**

```bash
cd d:/goproject/shadowsongAI && go build ./cmd/server
```

预期: 构建成功，无错误

---

## 验收标准

1. `parseTime` 函数正确处理毫秒时间戳
2. 阿里云模板集成测试通过
3. 所有现有测试保持通过
4. 代码已提交到版本控制
