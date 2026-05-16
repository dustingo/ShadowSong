package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/notifier"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ConfigHandler struct {
	db *gorm.DB
}

func NewConfigHandler(db *gorm.DB) *ConfigHandler {
	return &ConfigHandler{db: db}
}

// ============ DataSource ============

func (h *ConfigHandler) ListDataSources(c *gin.Context) {
	var datasources []models.DataSource
	if err := h.db.Order("id DESC").Find(&datasources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, datasources)
}

func (h *ConfigHandler) GetDataSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ds models.DataSource
	if err := h.db.First(&ds, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ds)
}

func (h *ConfigHandler) CreateDataSource(c *gin.Context) {
	var ds models.DataSource
	if err := c.ShouldBindJSON(&ds); err != nil {
		_ = recordAudit(h.db, c, "config.datasource.create", "datasource", "new", auditResultDenied, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ds.Validate(); err != nil {
		_ = recordAudit(h.db, c, "config.datasource.create", "datasource", "new", auditResultDenied, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int64
	h.db.Model(&models.DataSource{}).Where("name = ?", ds.Name).Count(&count)
	if count > 0 {
		_ = recordAudit(h.db, c, "config.datasource.create", "datasource", ds.Name, auditResultDenied, "name already exists")
		c.JSON(http.StatusBadRequest, gin.H{"error": "name already exists"})
		return
	}

	if err := h.db.Create(&ds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = recordAudit(h.db, c, "config.datasource.create", "datasource", strconv.FormatUint(uint64(ds.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", ds.Name))
	c.JSON(http.StatusOK, ds)
}

func (h *ConfigHandler) UpdateDataSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ds models.DataSource
	if err := h.db.First(&ds, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input models.DataSource
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" && input.Name != ds.Name {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be changed"})
		return
	}

	ds.DisplayName = input.DisplayName
	ds.APIKey = input.APIKey
	ds.InputTemplate = input.InputTemplate
	ds.OutputTemplate = input.OutputTemplate
	ds.GroupByLabels = input.GroupByLabels
	ds.Enabled = input.Enabled

	// 去重/聚合配置
	ds.DeduplicateEnabled = input.DeduplicateEnabled
	ds.DeduplicateWindow = input.DeduplicateWindow
	ds.GroupEnabled = input.GroupEnabled
	ds.GroupWindow = input.GroupWindow

	if err := h.db.Save(&ds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = recordAudit(h.db, c, "config.datasource.update", "datasource", strconv.FormatUint(uint64(ds.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", ds.Name))
	c.JSON(http.StatusOK, ds)
}

func (h *ConfigHandler) DeleteDataSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.db.Delete(&models.DataSource{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = recordAudit(h.db, c, "config.datasource.delete", "datasource", strconv.FormatUint(id, 10), auditResultAllowed, "deleted datasource")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) ToggleDataSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ds models.DataSource
	if err := h.db.First(&ds, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var input struct {
		Enabled bool `json:"enabled"`
	}
	c.ShouldBindJSON(&input)

	ds.Enabled = input.Enabled
	h.db.Save(&ds)
	_ = recordAudit(h.db, c, "config.datasource.toggle", "datasource", strconv.FormatUint(uint64(ds.ID), 10), auditResultAllowed, fmt.Sprintf("enabled=%t", ds.Enabled))
	c.JSON(http.StatusOK, ds)
}

func (h *ConfigHandler) PreviewDataSource(c *gin.Context) {
	var input struct {
		DataSourceID   uint            `json:"datasource_id"`
		SourceName     string          `json:"source_name"`
		InputTemplate  string          `json:"input_template"`
		OutputTemplate string          `json:"output_template"`
		SamplePayload  json.RawMessage `json:"sample_payload" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var ds models.DataSource
	if input.DataSourceID > 0 {
		if err := h.db.First(&ds, input.DataSourceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if strings.TrimSpace(input.InputTemplate) == "" {
		input.InputTemplate = ds.InputTemplate
	}
	if strings.TrimSpace(input.OutputTemplate) == "" {
		input.OutputTemplate = ds.OutputTemplate
	}
	if strings.TrimSpace(input.SourceName) == "" {
		if ds.Name != "" {
			input.SourceName = ds.Name
		} else {
			input.SourceName = "preview"
		}
	}

	if strings.TrimSpace(input.InputTemplate) == "" || strings.TrimSpace(input.OutputTemplate) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input_template and output_template are required for preview"})
		return
	}

	var rawData interface{}
	if err := json.Unmarshal(input.SamplePayload, &rawData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sample_payload: " + err.Error()})
		return
	}

	webhookHandler := NewWebhookHandler(h.db, nil)
	alerts := webhookHandler.normalizeData(rawData)
	if len(alerts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sample_payload must be a JSON object or array containing at least one object"})
		return
	}

	rawAlertBody := marshalRawAlertData(alerts[0], input.SamplePayload)
	alert, err := webhookHandler.renderAlert(alerts[0], input.InputTemplate, input.SourceName, rawAlertBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("input template render error: %v", err)})
		return
	}

	title, content, renderContext, err := webhookHandler.renderNotificationPreview(&alert, input.OutputTemplate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("output template render error: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"normalized_alert": alert,
		"rendered": gin.H{
			"title":   title,
			"content": content,
		},
		"context_preview": gin.H{
			"top_level_keys": sortedKeys(renderContext),
			"event_keys":     sortedKeysFromValue(renderContext["event"]),
			"label_keys":     sortedKeysFromValue(renderContext["labels"]),
			"alert_keys":     sortedKeysFromValue(renderContext["alert"]),
		},
	})
}

// ============ Channel ============

func (h *ConfigHandler) ListChannels(c *gin.Context) {
	var channels []models.Channel
	if err := h.db.Order("id DESC").Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range channels {
		channels[i].Config = maskChannelConfig(channels[i].Type, channels[i].Config)
	}
	c.JSON(http.StatusOK, channels)
}

func (h *ConfigHandler) GetChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ch models.Channel
	if err := h.db.First(&ch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// GetChannel 返回原始配置，便于编辑
	c.JSON(http.StatusOK, ch)
}

func (h *ConfigHandler) CreateChannel(c *gin.Context) {
	var ch models.Channel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ch.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&ch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = recordAudit(h.db, c, "config.channel.create", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", ch.Name))
	ch.Config = maskChannelConfig(ch.Type, ch.Config)
	c.JSON(http.StatusOK, ch)
}

func (h *ConfigHandler) UpdateChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ch models.Channel
	if err := h.db.First(&ch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var input struct {
		Name    string          `json:"name"`
		Type    string          `json:"type"`
		Config  json.RawMessage `json:"config"`
		Enabled bool            `json:"enabled"`
	}
	c.ShouldBindJSON(&input)

	if input.Name != "" {
		ch.Name = input.Name
	}
	if input.Type != "" {
		ch.Type = input.Type
	}
	// 只有当 config 不为空时才更新
	if len(input.Config) > 0 {
		ch.Config = datatypes.JSON(input.Config)
	}
	ch.Enabled = input.Enabled

	h.db.Save(&ch)
	_ = recordAudit(h.db, c, "config.channel.update", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", ch.Name))
	// UpdateChannel 返回原始配置，便于确认编辑结果
	c.JSON(http.StatusOK, ch)
}

func (h *ConfigHandler) DeleteChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	h.db.Delete(&models.Channel{}, id)
	_ = recordAudit(h.db, c, "config.channel.delete", "channel", strconv.FormatUint(id, 10), auditResultAllowed, "deleted channel")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) ToggleChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ch models.Channel
	h.db.First(&ch, id)

	var input struct {
		Enabled bool `json:"enabled"`
	}
	c.ShouldBindJSON(&input)

	ch.Enabled = input.Enabled
	h.db.Save(&ch)
	_ = recordAudit(h.db, c, "config.channel.toggle", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultAllowed, fmt.Sprintf("enabled=%t", ch.Enabled))
	c.JSON(http.StatusOK, ch)
}

func (h *ConfigHandler) TestChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ch models.Channel
	if err := h.db.First(&ch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	if !ch.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel is disabled"})
		return
	}

	testTitle := "测试通知"
	testContent := "这是一条来自游戏运维告警系统的测试消息。"

	if err := notifier.SendToChannel(&ch, testTitle, testContent); err != nil {
		_ = recordAudit(h.db, c, "config.channel.test", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultDenied, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "config.channel.test", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultAllowed, "test notification sent")
	c.JSON(http.StatusOK, gin.H{"message": "test notification sent successfully"})
}

// ============ RouteRule ============

func (h *ConfigHandler) ListRouteRules(c *gin.Context) {
	var rules []models.RouteRule
	if err := h.db.Order("priority ASC").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *ConfigHandler) GetRouteRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rule models.RouteRule
	if err := h.db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *ConfigHandler) CreateRouteRule(c *gin.Context) {
	var rule models.RouteRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := rule.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.db.Create(&rule)
	_ = recordAudit(h.db, c, "config.route.create", "route_rule", strconv.FormatUint(uint64(rule.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", rule.Name))
	c.JSON(http.StatusOK, rule)
}

func (h *ConfigHandler) UpdateRouteRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rule models.RouteRule
	if err := h.db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var input models.RouteRule
	c.ShouldBindJSON(&input)

	rule.Name = input.Name
	rule.Priority = input.Priority
	rule.Severities = input.Severities
	rule.Sources = input.Sources
	rule.LabelMatchers = input.LabelMatchers
	rule.ChannelIDs = input.ChannelIDs
	rule.TimeRanges = input.TimeRanges
	rule.Enabled = input.Enabled

	h.db.Save(&rule)
	_ = recordAudit(h.db, c, "config.route.update", "route_rule", strconv.FormatUint(uint64(rule.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", rule.Name))
	c.JSON(http.StatusOK, rule)
}

func (h *ConfigHandler) DeleteRouteRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	h.db.Delete(&models.RouteRule{}, id)
	_ = recordAudit(h.db, c, "config.route.delete", "route_rule", strconv.FormatUint(id, 10), auditResultAllowed, "deleted route rule")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) ReorderRouteRules(c *gin.Context) {
	var input struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, id := range input.IDs {
		h.db.Model(&models.RouteRule{}).Where("id = ?", id).Update("priority", i+1)
	}
	_ = recordAudit(h.db, c, "config.route.reorder", "route_rule", "bulk", auditResultAllowed, fmt.Sprintf("count=%d", len(input.IDs)))
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// ============ SilenceRule ============

func (h *ConfigHandler) ListSilenceRules(c *gin.Context) {
	var rules []models.SilenceRule
	query := h.db.Model(&models.SilenceRule{})

	status := c.Query("status")
	if status == "active" {
		query = query.Where("ends_at > ?", time.Now())
	} else if status == "expired" {
		query = query.Where("ends_at <= ?", time.Now())
	}

	if err := query.Order("starts_at DESC").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *ConfigHandler) GetSilenceRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rule models.SilenceRule
	if err := h.db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *ConfigHandler) CreateSilenceRule(c *gin.Context) {
	var rule models.SilenceRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := rule.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.db.Create(&rule)
	_ = recordAudit(h.db, c, "config.silence.create", "silence_rule", strconv.FormatUint(uint64(rule.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", rule.Name))
	c.JSON(http.StatusOK, rule)
}

func (h *ConfigHandler) UpdateSilenceRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rule models.SilenceRule
	if err := h.db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var input models.SilenceRule
	c.ShouldBindJSON(&input)

	rule.Name = input.Name
	rule.Comment = input.Comment
	rule.Source = input.Source
	rule.AlertNamePattern = input.AlertNamePattern
	rule.Severities = input.Severities
	rule.LabelMatchers = input.LabelMatchers
	rule.StartsAt = input.StartsAt
	rule.EndsAt = input.EndsAt

	h.db.Save(&rule)
	_ = recordAudit(h.db, c, "config.silence.update", "silence_rule", strconv.FormatUint(uint64(rule.ID), 10), auditResultAllowed, fmt.Sprintf("name=%s", rule.Name))
	c.JSON(http.StatusOK, rule)
}

func (h *ConfigHandler) DeleteSilenceRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	h.db.Delete(&models.SilenceRule{}, id)
	_ = recordAudit(h.db, c, "config.silence.delete", "silence_rule", strconv.FormatUint(id, 10), auditResultAllowed, "deleted silence rule")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) CreateSilenceFromAlert(c *gin.Context) {
	alertID := c.Param("alertId")
	var alert models.Alert
	if err := h.db.First(&alert, "alert_id = ?", alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	var input struct {
		Duration int `json:"duration"`
	}
	c.ShouldBindJSON(&input)

	rule := models.SilenceRule{
		Name:             "Quick Silence - " + alert.AlertName,
		AlertNamePattern: alert.AlertName,
		Severities:       []byte(`["` + alert.Severity + `"]`),
		StartsAt:         time.Now(),
		EndsAt:           time.Now().Add(time.Duration(input.Duration) * time.Second),
		CreatedBy:        "system",
	}

	h.db.Create(&rule)
	_ = recordAudit(h.db, c, "config.silence.create_from_alert", "alert", alert.AlertID, auditResultAllowed, fmt.Sprintf("silence_id=%d", rule.ID))
	c.JSON(http.StatusOK, rule)
}

// ============ OnDuty ============

func (h *ConfigHandler) ListOnDuty(c *gin.Context) {
	var duties []models.OnDuty
	if err := h.db.Order("start_time DESC").Find(&duties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, duties)
}

func (h *ConfigHandler) GetOnDuty(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var duty models.OnDuty
	if err := h.db.First(&duty, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, duty)
}

func (h *ConfigHandler) CreateOnDuty(c *gin.Context) {
	var duty models.OnDuty
	if err := c.ShouldBindJSON(&duty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := duty.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.db.Create(&duty)
	_ = recordAudit(h.db, c, "config.onduty.create", "onduty", strconv.FormatUint(uint64(duty.ID), 10), auditResultAllowed, fmt.Sprintf("user_name=%s", duty.UserName))
	c.JSON(http.StatusOK, duty)
}

func (h *ConfigHandler) UpdateOnDuty(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var duty models.OnDuty
	if err := h.db.First(&duty, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var input models.OnDuty
	c.ShouldBindJSON(&input)

	duty.UserID = input.UserID
	duty.UserName = input.UserName
	duty.ChannelID = input.ChannelID
	duty.StartTime = input.StartTime
	duty.EndTime = input.EndTime

	h.db.Save(&duty)
	_ = recordAudit(h.db, c, "config.onduty.update", "onduty", strconv.FormatUint(uint64(duty.ID), 10), auditResultAllowed, fmt.Sprintf("user_name=%s", duty.UserName))
	c.JSON(http.StatusOK, duty)
}

func (h *ConfigHandler) DeleteOnDuty(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	h.db.Delete(&models.OnDuty{}, id)
	_ = recordAudit(h.db, c, "config.onduty.delete", "onduty", strconv.FormatUint(id, 10), auditResultAllowed, "deleted on-duty schedule")
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) CurrentOnDuty(c *gin.Context) {
	var duties []models.OnDuty
	now := time.Now()
	if err := h.db.Where("start_time <= ? AND end_time >= ?", now, now).Find(&duties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, duties)
}

// Helper functions

func maskChannelConfig(chType string, config []byte) []byte {
	if config == nil {
		return []byte(`{}`)
	}
	configStr := string(config)
	sensitiveKeys := []string{"webhook_url", "secret", "sign_key", "password", "header_value"}
	for _, key := range sensitiveKeys {
		if strings.Contains(configStr, key) {
			return []byte(`{"masked": true}`)
		}
	}
	return config
}

func sortedKeys(input map[string]interface{}) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysFromValue(input interface{}) []string {
	m, ok := input.(map[string]interface{})
	if !ok {
		return []string{}
	}
	return sortedKeys(m)
}
