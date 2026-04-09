package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/notifier"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Template field names
const (
	InputTemplate  = "input_template"
	OutputTemplate = "output_template"
)

type WebhookHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewWebhookHandler(db *gorm.DB, redisClient *redis.Client) *WebhookHandler {
	return &WebhookHandler{
		db:          db,
		redisClient: redisClient,
	}
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	sourceName := c.Param("source_name")

	// 1. 验证数据源是否存在且启用
	var ds models.DataSource
	if err := h.db.First(&ds, "name = ?", sourceName).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ds.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "data source is disabled"})
		return
	}

	// 2. 验证 API Key（支持 Header: X-API-Key 或 Bearer Token）
	// 如果数据源未配置 API Key，直接拒绝所有请求
	if ds.APIKey == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "data source requires API key, please configure in settings"})
		return
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		// 尝试从 Authorization header 获取
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 验证 API Key
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
		return
	}
	if apiKey != ds.APIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
		return
	}

	// 3. 读取原始数据
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// 3. 解析输入数据（支持数组和单对象）
	var rawData interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json format"})
		return
	}

	// 4. 处理数据（可能是数组或单个对象）
	var results []models.Alert
	var newAlerts []models.Alert // 新创建的告警（需要发送通知）
	var errors []string

	alerts := h.normalizeData(rawData)

	// 获取去重窗口配置
	dedupEnabled := ds.DeduplicateEnabled
	dedupWindow := time.Duration(ds.DeduplicateWindow) * time.Second
	if dedupWindow == 0 {
		dedupWindow = 1 * time.Hour // 默认1小时
	}

	for _, alertData := range alerts {
		// 5. 使用 input_template 渲染
		alert, err := h.renderAlert(alertData, ds.InputTemplate, sourceName, body)
		if err != nil {
			// 模板渲染失败，生成降级告警
			alert = h.createFallbackAlert(sourceName, body, err)
			errors = append(errors, fmt.Sprintf("render failed: %v", err))
		}

		// 6. 生成指纹
		alert.Fingerprint = h.generateFingerprint(alert, ds.GroupByLabels)

		// 7. 去重逻辑
		if dedupEnabled {
			// 查找是否存在相同指纹且在去重窗口内的告警
			var existing models.Alert
			query := h.db.Where("fingerprint = ? AND status IN ?", alert.Fingerprint, []string{"firing", "pending"})

			// 如果配置了去重时间窗口，只查找窗口内的
			if dedupWindow > 0 {
				windowStart := time.Now().Add(-dedupWindow)
				query = query.Where("trigger_time >= ?", windowStart)
			}

			err = query.First(&existing).Error
			if err == nil {
				// 已存在，更新 trigger count
				now := time.Now()
				existing.TriggerCount++
				existing.TriggerTime = now
				existing.LastRepeatAt = &now

				// 延长去重窗口
				if dedupWindow > 0 {
					dedupUntil := now.Add(dedupWindow)
					existing.DeduplicateUntil = &dedupUntil
				}

				h.db.Save(&existing)
				results = append(results, existing)
				// 去重告警不发送通知（可选：可以配置是否发送）
				continue
			}
		}

		// 8. 保存新告警
		now := time.Now()
		alert.TriggerTime = now
		alert.ReceivedAt = now
		if dedupEnabled && dedupWindow > 0 {
			dedupUntil := now.Add(dedupWindow)
			alert.DeduplicateUntil = &dedupUntil
		}

		if err := h.db.Create(&alert).Error; err != nil {
			errors = append(errors, fmt.Sprintf("failed to save alert: %v", err))
			continue
		}

		results = append(results, alert)
		newAlerts = append(newAlerts, alert)
	}

	// 9. 写入 Redis Stream（只写入新告警）
	if len(newAlerts) > 0 {
		h.publishToRedis(newAlerts)
	}

	// 10. 执行路由规则和推送通知（只发送新告警）
	if len(newAlerts) > 0 {
		go h.processAlertNotifications(newAlerts)
	}

	// 10. 返回结果
	if len(errors) > 0 && len(results) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"errors":  errors,
			"message": "failed to process alerts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"received":  len(alerts),
		"processed": len(results),
		"errors":    errors,
		"alerts":    results,
	})
}

// normalizeData 确保输入是数组格式
func (h *WebhookHandler) normalizeData(data interface{}) []map[string]interface{} {
	var alerts []map[string]interface{}

	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				alerts = append(alerts, m)
			}
		}
	case map[string]interface{}:
		alerts = append(alerts, v)
	}

	return alerts
}

// renderAlert 使用 input_template 渲染告警数据
func (h *WebhookHandler) renderAlert(data map[string]interface{}, tmplStr string, source string, rawBody []byte) (models.Alert, error) {
	// 创建模板函数
	funcMap := template.FuncMap{
		"toJson": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"get": func(m map[string]interface{}, key string) interface{} {
			if m == nil {
				return nil
			}
			val, ok := m[key]
			if !ok {
				return nil
			}
			return val
		},
		"default": func(v, def interface{}) interface{} {
			if v == nil {
				return def
			}
			// 如果是空字符串，也返回默认值
			if s, ok := v.(string); ok && s == "" {
				return def
			}
			return v
		},
		"toSeverity": func(v interface{}) string {
			return h.mapSeverity(v)
		},
		"toStatus": func(v interface{}) string {
			return h.mapStatus(v)
		},
		"toTime": func(v interface{}) string {
			return h.parseTime(v)
		},
		"lookup": func(m map[string]interface{}, keys ...string) interface{} {
			if m == nil {
				return nil
			}
			for _, key := range keys {
				if val, ok := m[key]; ok && val != nil {
					return val
				}
			}
			return nil
		},
	}

	tmpl, err := template.New("input").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return models.Alert{}, fmt.Errorf("template parse error: %v", err)
	}

	// 执行模板
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return models.Alert{}, fmt.Errorf("template execute error: %v", err)
	}

	// 解析渲染结果为 Alert
	var alert models.Alert
	resultStr := buf.String()

	// 尝试解析为 JSON
	if err := json.Unmarshal([]byte(resultStr), &alert); err != nil {
		// 如果不是 JSON，尝试手动解析
		alert = h.parseNonJsonAlert(resultStr, source)
	}

	// 设置默认值
	if alert.AlertID == "" {
		alert.AlertID = fmt.Sprintf("%s-%d", source, time.Now().UnixNano())
	}
	if alert.Source == "" {
		alert.Source = source
	}
	if alert.Status == "" {
		alert.Status = "pending"
	} else {
		// 标准化 status
		alert.Status = h.mapStatus(alert.Status)
	}
	if alert.Severity == "" {
		alert.Severity = "P2"
	} else {
		// 标准化 severity
		alert.Severity = h.mapSeverity(alert.Severity)
	}
	if alert.TriggerTime.IsZero() {
		alert.TriggerTime = time.Now()
	}
	if alert.ReceivedAt.IsZero() {
		alert.ReceivedAt = time.Now()
	}

	// 保存原始数据
	alert.Raw = rawBody

	// 设置 Labels
	labelsBytes, _ := json.Marshal(data)
	alert.Labels = labelsBytes

	return alert, nil
}

// mapSeverity 将各种 severity 格式映射为 P0-P3
func (h *WebhookHandler) mapSeverity(v interface{}) string {
	if v == nil {
		return "P2"
	}

	s := strings.ToLower(fmt.Sprintf("%v", v))

	// 已经是 P0-P3
	if len(s) == 2 && s[0] == 'p' && s[1] >= '0' && s[1] <= '3' {
		return strings.ToUpper(s)
	}

	// Prometheus/AlertManager 格式
	severityMap := map[string]string{
		// critical -> P0
		"critical":  "P0",
		"crit":      "P0",
		"emergency": "P0",
		"emerg":     "P0",
		"alert":     "P0",
		"fatal":     "P0",
		"high":      "P0",
		// warning -> P1
		"warning": "P1",
		"warn":    "P1",
		"error":   "P1",
		"err":     "P1",
		"severe":  "P1",
		// info -> P2
		"info":   "P2",
		"notice": "P2",
		"normal": "P2",
		"ok":     "P2",
		// debug -> P3
		"debug": "P3",
		"trace": "P3",
		"low":   "P3",
	}

	// 尝试精确匹配
	if p, ok := severityMap[s]; ok {
		return p
	}

	// 尝试前缀匹配
	for key, val := range severityMap {
		if strings.Contains(s, key) {
			return val
		}
	}

	// 默认返回 P2
	return "P2"
}

// mapStatus 将各种 status 格式映射为标准状态
func (h *WebhookHandler) mapStatus(v interface{}) string {
	if v == nil {
		return "pending"
	}

	s := strings.ToLower(fmt.Sprintf("%v", v))

	if s == "firing" || s == "active" || s == "triggered" {
		return "firing"
	}
	if s == "resolved" || s == "inactive" || s == "ok" || s == "success" {
		return "resolved"
	}
	if s == "acknowledged" || s == "acked" || s == "ack" {
		return "acked"
	}
	if s == "silenced" || s == "suppressed" {
		return "silenced"
	}

	return "pending"
}

// parseTime 解析时间字符串
func (h *WebhookHandler) parseTime(v interface{}) string {
	if v == nil {
		return time.Now().Format(time.RFC3339)
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

// regexpMatch 简单的正则匹配
func regexpMatch(pattern, s string) (bool, error) {
	// 简化版本：只支持基本的字符串包含和前缀/后缀匹配
	if strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$") {
		// 精确匹配
		pattern = strings.Trim(pattern, "^$")
		return pattern == s, nil
	}
	if strings.HasPrefix(pattern, "^") {
		return strings.HasPrefix(s, strings.TrimPrefix(pattern, "^")), nil
	}
	if strings.HasSuffix(pattern, "$") {
		return strings.HasSuffix(s, strings.TrimSuffix(pattern, "$")), nil
	}
	return strings.Contains(s, pattern), nil
}

// parseNonJsonAlert 解析非 JSON 格式的告警
func (h *WebhookHandler) parseNonJsonAlert(str, source string) models.Alert {
	alert := models.Alert{
		AlertID:     fmt.Sprintf("%s-%d", source, time.Now().UnixNano()),
		Source:      source,
		Status:      "pending",
		Message:     str,
		Severity:    "P2",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
	}

	// 尝试从文本中提取信息
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if strings.Contains(line, "severity:") {
			alert.Severity = extractValue(line, "severity:")
		} else if strings.Contains(line, "alertname:") {
			alert.AlertName = extractValue(line, "alertname:")
		}
	}

	return alert
}

func extractValue(line, key string) string {
	line = strings.ToLower(line)
	key = strings.ToLower(key)
	idx := strings.Index(line, key)
	if idx == -1 {
		return ""
	}
	parts := strings.SplitN(line[idx:], ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// createFallbackAlert 创建降级告警
func (h *WebhookHandler) createFallbackAlert(source string, rawBody []byte, err error) models.Alert {
	msg := string(rawBody)
	if len(msg) > 500 {
		msg = msg[:500] + "..."
	}

	return models.Alert{
		AlertID:     fmt.Sprintf("%s-fallback-%d", source, time.Now().UnixNano()),
		Source:      source,
		AlertName:   "TemplateRenderFailed",
		Severity:    "P1",
		Message:     fmt.Sprintf("Template rendering failed: %v. Raw data: %s", err, msg),
		Status:      "pending",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
		Raw:         rawBody,
		Labels:      []byte("{}"),
	}
}

// generateFingerprint 生成告警指纹
func (h *WebhookHandler) generateFingerprint(alert models.Alert, groupByLabels []byte) string {
	var labels []string
	json.Unmarshal(groupByLabels, &labels)

	if len(labels) == 0 {
		// 默认使用 alert_name + severity + source
		labels = []string{"alert_name", "severity", "source"}
	}

	// 从 Labels 中提取需要分组的值
	var labelData map[string]string
	json.Unmarshal(alert.Labels, &labelData)

	var fingerprintParts []string
	for _, label := range labels {
		if val, ok := labelData[label]; ok {
			fingerprintParts = append(fingerprintParts, fmt.Sprintf("%s=%s", label, val))
		} else {
			// 尝试从告警字段获取
			switch label {
			case "alert_name":
				fingerprintParts = append(fingerprintParts, fmt.Sprintf("alert_name=%s", alert.AlertName))
			case "severity":
				fingerprintParts = append(fingerprintParts, fmt.Sprintf("severity=%s", alert.Severity))
			case "source":
				fingerprintParts = append(fingerprintParts, fmt.Sprintf("source=%s", alert.Source))
			}
		}
	}

	fingerprint := strings.Join(fingerprintParts, ",")
	if fingerprint == "" {
		fingerprint = alert.AlertID
	}

	// 计算 SHA256
	hash := sha256.Sum256([]byte(fingerprint))
	return hex.EncodeToString(hash[:])
}

// publishToRedis 写入 Redis Stream
func (h *WebhookHandler) publishToRedis(alerts []models.Alert) {
	if h.redisClient == nil {
		return
	}

	ctx := context.Background()
	for _, alert := range alerts {
		data := map[string]interface{}{
			"alert_id":     alert.AlertID,
			"source":       alert.Source,
			"alert_name":   alert.AlertName,
			"severity":     alert.Severity,
			"message":      alert.Message,
			"fingerprint":  alert.Fingerprint,
			"status":       alert.Status,
			"trigger_time": alert.TriggerTime.Unix(),
		}

		h.redisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: "alerts:pending",
			Values: data,
		})
	}
}

// ============ 辅助方法 ============

// ValidateInputTemplate 验证模板是否有效
func (h *WebhookHandler) ValidateInputTemplate(tmplStr string, sampleData string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(sampleData), &data); err != nil {
		return "", fmt.Errorf("invalid sample data: %v", err)
	}

	_, err := template.New("test").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %v", err)
	}

	return "template is valid", nil
}

// TestInputTemplate 测试输入模板
func (h *WebhookHandler) TestInputTemplate(c *gin.Context) {
	var input struct {
		Template   string `json:"template" binding:"required"`
		SampleData string `json:"sample_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(input.SampleData), &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sample_data: " + err.Error()})
		return
	}

	// 创建模板
	funcMap := template.FuncMap{
		"toJson": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"get": func(m map[string]interface{}, key string) interface{} {
			return m[key]
		},
		"default": func(v, def interface{}) interface{} {
			if v == nil {
				return def
			}
			return v
		},
	}

	tmpl, err := template.New("input").Funcs(funcMap).Parse(input.Template)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("template parse error: %v", err),
		})
		return
	}

	// 执行模板
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("template execute error: %v", err),
		})
		return
	}

	// 尝试解析为 Alert 验证字段
	var alert models.Alert
	resultStr := buf.String()
	if err := json.Unmarshal([]byte(resultStr), &alert); err != nil {
		// 不是有效 JSON，但模板渲染成功
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"result":  resultStr,
			"warning": "result is not valid JSON, cannot validate fields",
		})
		return
	}

	// 验证必填字段
	var validationErrors []string
	if alert.AlertName == "" {
		validationErrors = append(validationErrors, "alert_name is required")
	}
	if alert.Severity == "" {
		validationErrors = append(validationErrors, "severity is required")
	} else if !alert.IsValidSeverity() {
		validationErrors = append(validationErrors, fmt.Sprintf("invalid severity: %s (must be P0-P3)", alert.Severity))
	}
	if alert.Message == "" {
		validationErrors = append(validationErrors, "message is required")
	}

	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      true,
			"result":       resultStr,
			"valid_alert":  false,
			"valid_errors": validationErrors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"result":      resultStr,
		"valid_alert": true,
		"alert":       alert,
	})
}

// ============ 路由和推送逻辑 ============

// processAlertNotifications 处理告警的路由和推送
func (h *WebhookHandler) processAlertNotifications(alerts []models.Alert) {
	// 1. 获取所有启用的路由规则（按优先级排序）
	var rules []models.RouteRule
	h.db.Where("enabled = ?", true).Order("priority ASC").Find(&rules)

	if len(rules) == 0 {
		fmt.Println("No route rules found, skipping notification")
		return
	}

	// 2. 遍历每个告警
	for _, alert := range alerts {
		// 3. 检查告警是否匹配任何路由规则
		matchedChannels := h.findMatchedChannels(alert, rules)

		if len(matchedChannels) == 0 {
			fmt.Printf("Alert %s no matched route rules\n", alert.AlertID)
			continue
		}

		// 4. 使用 output_template 生成通知内容
		for _, channel := range matchedChannels {
			h.sendNotification(&alert, &channel)
		}
	}
}

// findMatchedChannels 查找匹配的渠道
func (h *WebhookHandler) findMatchedChannels(alert models.Alert, rules []models.RouteRule) []models.Channel {
	var matchedChannels []models.Channel

	for _, rule := range rules {
		// 检查 source 是否匹配
		var sources []string
		json.Unmarshal(rule.Sources, &sources)
		if len(sources) > 0 && !contains(sources, alert.Source) {
			continue
		}

		// 检查 severity 是否匹配
		var severities []string
		json.Unmarshal(rule.Severities, &severities)
		if len(severities) > 0 && !contains(severities, alert.Severity) {
			continue
		}

		// 检查 label_matchers 是否匹配
		var labelMatchers []models.LabelMatcher
		json.Unmarshal(rule.LabelMatchers, &labelMatchers)
		if len(labelMatchers) > 0 && !h.matchLabels(alert.Labels, labelMatchers) {
			continue
		}

		// 检查时间范围
		if !h.isInTimeRange(rule.TimeRanges) {
			continue
		}

		// 获取规则对应的渠道
		var channelIDs []uint
		json.Unmarshal(rule.ChannelIDs, &channelIDs)

		for _, channelID := range channelIDs {
			var channel models.Channel
			if err := h.db.First(&channel, channelID).Error; err != nil {
				continue
			}
			if channel.Enabled {
				matchedChannels = append(matchedChannels, channel)
			}
		}

		// 找到第一个匹配的规则就停止（按优先级）
		if len(matchedChannels) > 0 {
			break
		}
	}

	return matchedChannels
}

// matchLabels 检查标签是否匹配
func (h *WebhookHandler) matchLabels(labelsJSON []byte, matchers []models.LabelMatcher) bool {
	if len(matchers) == 0 {
		return true
	}

	var labels map[string]string
	if err := json.Unmarshal(labelsJSON, &labels); err != nil {
		return false
	}

	for _, matcher := range matchers {
		value, exists := labels[matcher.Key]
		if !exists {
			return false
		}
		// 使用 pattern 作为正则匹配
		if matcher.Pattern != "" {
			matched, _ := regexpMatch(matcher.Pattern, value)
			if !matched {
				return false
			}
		} else {
			// 如果没有 pattern，则精确匹配
			// 这里假设 pattern 就是值（简化处理）
		}
	}

	return true
}

// isInTimeRange 检查是否在时间范围内
func (h *WebhookHandler) isInTimeRange(timeRangesJSON []byte) bool {
	if len(timeRangesJSON) == 0 || string(timeRangesJSON) == "[]" {
		return true
	}

	var timeRanges []models.TimeRange
	if err := json.Unmarshal(timeRangesJSON, &timeRanges); err != nil {
		return true
	}

	if len(timeRanges) == 0 {
		return true
	}

	now := time.Now()
	currentTime := now.Hour()*60 + now.Minute()

	for _, tr := range timeRanges {
		startMinutes := parseTimeToMinutes(tr.StartTime)
		endMinutes := parseTimeToMinutes(tr.EndTime)

		// 处理跨天的情况
		if endMinutes < startMinutes {
			// 跨天：例如 22:00 - 06:00
			if currentTime >= startMinutes || currentTime <= endMinutes {
				return true
			}
		} else {
			// 同一天
			if currentTime >= startMinutes && currentTime <= endMinutes {
				return true
			}
		}
	}

	return false
}

func parseTimeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}
	hour := 0
	minute := 0
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)
	return hour*60 + minute
}

// sendNotification 发送通知
func (h *WebhookHandler) sendNotification(alert *models.Alert, channel *models.Channel) {
	// 获取 output_template
	var ds models.DataSource
	if err := h.db.First(&ds, "name = ?", alert.Source).Error; err != nil {
		fmt.Printf("DataSource not found: %s\n", alert.Source)
		// 使用默认模板
		h.sendDefaultNotification(alert, channel)
		return
	}

	// 使用 output_template 渲染通知内容
	title, content, err := h.renderNotification(alert, string(ds.OutputTemplate))
	if err != nil {
		fmt.Printf("Failed to render notification: %v\n", err)
		h.sendDefaultNotification(alert, channel)
		return
	}

	// 发送通知
	if err := notifier.SendToChannel(channel, title, content); err != nil {
		fmt.Printf("Failed to send notification to channel %d: %v\n", channel.ID, err)
	} else {
		fmt.Printf("Notification sent to channel %d: %s\n", channel.ID, channel.Name)
	}
}

// sendDefaultNotification 发送默认格式的通知
func (h *WebhookHandler) sendDefaultNotification(alert *models.Alert, channel *models.Channel) {
	title := fmt.Sprintf("[%s] %s", alert.Severity, alert.AlertName)
	content := alert.Message
	if err := notifier.SendToChannel(channel, title, content); err != nil {
		fmt.Printf("Failed to send default notification: %v\n", err)
	}
}

// renderNotification 使用 output_template 渲染通知
func (h *WebhookHandler) renderNotification(alert *models.Alert, tmplStr string) (string, string, error) {
	// 准备数据
	data := map[string]interface{}{
		"alert_id":     alert.AlertID,
		"alert_name":   alert.AlertName,
		"severity":     alert.Severity,
		"message":      alert.Message,
		"source":       alert.Source,
		"status":       alert.Status,
		"trigger_time": alert.TriggerTime.Format(time.RFC3339),
	}

	// 解析 labels
	var labels map[string]string
	json.Unmarshal(alert.Labels, &labels)
	data["labels"] = labels

	// 创建模板函数
	funcMap := template.FuncMap{
		"toJson": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", "", err
	}

	resultStr := buf.String()

	// 尝试解析为 JSON
	var result map[string]string
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		// 不是 JSON，直接返回
		return "告警通知", resultStr, nil
	}

	title := result["title"]
	content := result["content"]
	if title == "" {
		title = "告警通知"
	}
	if content == "" {
		content = resultStr
	}

	return title, content, nil
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============ 预定义模板 ============

// GetDefaultTemplates 返回默认模板
func GetDefaultTemplates() map[string]map[string]string {
	return map[string]map[string]string{
		"prometheus": {
			InputTemplate:  `{"alert_id": "{{.fingerprint}}", "alert_name": "{{.labels.alertname}}", "severity": "{{.labels.severity}}", "message": "{{.annotations.summary}}", "source": "prometheus", "status": "{{.status}}", "trigger_time": "{{.startsAt}}"}`,
			OutputTemplate: `{"title": "[{{.severity}}] {{.alert_name}}", "content": "{{.message}}\n\nLabels: {{toJson .labels}}"}`,
		},
		"alertmanager": {
			InputTemplate:  `{"alert_id": "{{.fingerprint}}", "alert_name": "{{.labels.alertname}}", "severity": "{{.labels.severity}}", "message": "{{.annotations.description}}", "source": "alertmanager", "status": "{{.status}}", "trigger_time": "{{.startsAt}}"}`,
			OutputTemplate: `{"title": "[{{.severity}}] {{.alert_name}}", "content": "{{.message}}\n\nLabels: {{toJson .labels}}"}`,
		},
		"custom": {
			InputTemplate:  `{"alert_id": "{{.alert_id}}", "alert_name": "{{.alert_name}}", "severity": "{{.severity}}", "message": "{{.message}}", "source": "custom", "status": "firing", "trigger_time": "{{.timestamp}}"}`,
			OutputTemplate: `{"title": "[{{.severity}}] {{.alert_name}}", "content": "{{.message}}"}`,
		},
	}
}

// CleanAlerts 清理告警（处理已恢复的告警）
func (h *WebhookHandler) CleanAlerts() {
	// 查找已解决的告警并更新状态
	var resolvedAlerts []models.Alert
	h.db.Where("status = ? AND updated_at < ?", "firing", time.Now().Add(-1*time.Hour)).Find(&resolvedAlerts)

	for _, alert := range resolvedAlerts {
		alert.Status = "resolved"
		h.db.Save(&alert)
	}
}
