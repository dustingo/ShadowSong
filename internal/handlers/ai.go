package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/ai"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

type AIHandler struct {
	db        *gorm.DB
	client    *ai.Client
}

func NewAIHandler(db *gorm.DB, cfg *config.Config) *AIHandler {
	return &AIHandler{
		db:     db,
		client: ai.NewClient(&cfg.AI),
	}
}

// System prompt for the AI assistant
const systemPrompt = `你是一个游戏运维 AI 助手，专门帮助用户处理告警问题。你可以：
1. 分析告警原因和影响
2. 提供处置建议
3. 解释告警的含义
4. 帮助创建静默规则
5. 回答关于系统的问题

请用中文回答问题，保持简洁和专业。`

// AI Chat - 智能问答
func (h *AIHandler) Chat(c *gin.Context) {
	var input struct {
		Message string `json:"message" binding:"required"`
		Context struct {
			AlertID string `json:"alert_id"`
		} `json:"context"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reply string
	var err error

	// Try to use real AI API
	if h.client != nil && h.client.Config.APIKey != "" {
		reply, err = h.client.Chat(systemPrompt, input.Message)
		if err != nil {
			// Fallback to mock response
			reply = "抱歉，AI 服务暂时不可用: " + err.Error()
		}
	} else {
		// Mock response when no API key
		reply = "我已收到您的问题: " + input.Message + "\n\n要使用 AI 功能，请配置 OPENAI_API_KEY 环境变量。"
	}

	// Log the interaction
	log := models.AILog{
		Input:    input.Message,
		Output:   reply,
		Accurate: nil,
	}
	if input.Context.AlertID != "" {
		log.AlertID = input.Context.AlertID
		var alert models.Alert
		h.db.First(&alert, "alert_id = ?", input.Context.AlertID)
		log.AlertName = alert.AlertName
	}
	h.db.Create(&log)

	c.JSON(http.StatusOK, gin.H{
		"reply":      reply,
		"suggestions": []string{"查看告警详情", "创建静默规则", "联系值班人员"},
	})
}

// Get suggestions for an alert - 处置建议
func (h *AIHandler) Suggestions(c *gin.Context) {
	alertID := c.Param("alertId")
	var alert models.Alert
	if err := h.db.First(&alert, "alert_id = ?", alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	var suggestions []string

	// Try to use real AI API
	if h.client != nil && h.client.Config.APIKey != "" {
		labelsJSON := string(alert.Labels)
		summary, rootCause, sugs, aiErr := h.client.AnalyzeAlert(alert.AlertName, alert.Message, labelsJSON)
		if aiErr == nil {
			suggestions = sugs
			// Also update alert with AI analysis
			alert.AISummary = summary
			alert.AIRootCause = rootCause
			h.db.Save(&alert)
		}
	}

	if suggestions == nil {
		// Default suggestions
		suggestions = []string{
			"检查 " + alert.AlertName + " 指标",
			"查看相关日志",
			"联系值班人员",
		}
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// List AI logs - AI日志
func (h *AIHandler) ListLogs(c *gin.Context) {
	var logs []models.AILog
	query := h.db.Model(&models.AILog{})

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var total int64
	query.Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  logs,
		"total": total,
	})
}

// Mark AI log accuracy - 标记准确性
func (h *AIHandler) MarkAccuracy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var log models.AILog
	if err := h.db.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}

	var input struct {
		Accurate bool `json:"accurate"`
	}
	c.ShouldBindJSON(&input)

	log.Accurate = &input.Accurate
	h.db.Save(&log)

	c.JSON(http.StatusOK, log)
}

// Delete AI log - 删除AI日志
func (h *AIHandler) DeleteLog(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.db.Delete(&models.AILog{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Get silence recommendations - 静默推荐
func (h *AIHandler) ListRecommendations(c *gin.Context) {
	var recs []models.SilenceRecommendation
	if err := h.db.Where("status = ?", "pending").Order("created_at DESC").Find(&recs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recs)
}

// Adopt a silence recommendation
func (h *AIHandler) AdoptRecommendation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rec models.SilenceRecommendation
	if err := h.db.First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
		return
	}

	// Create silence rule
	silence := models.SilenceRule{
		Name:              "AI Recommended - " + rec.AlertName,
		AlertNamePattern: rec.AlertName,
		Severities:        []byte(`["P1", "P2", "P3"]`),
		StartsAt:          rec.CreatedAt,
		EndsAt:            rec.CreatedAt.Add(time.Duration(rec.SuggestedDuration) * time.Second),
		CreatedBy:        "ai",
		Comment:          "Adopted from AI recommendation: " + rec.Reason,
	}
	h.db.Create(&silence)

	// Update recommendation status
	rec.Status = "adopted"
	h.db.Save(&rec)

	c.JSON(http.StatusOK, gin.H{"message": "adopted", "silence_id": silence.ID})
}

// Ignore a silence recommendation
func (h *AIHandler) IgnoreRecommendation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rec models.SilenceRecommendation
	if err := h.db.First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
		return
	}

	rec.Status = "ignored"
	h.db.Save(&rec)

	c.JSON(http.StatusOK, gin.H{"message": "ignored"})
}

// Generate silence recommendations (could be called by cron)
func (h *AIHandler) GenerateRecommendations(c *gin.Context) {
	// TODO: Analyze recent alerts and generate recommendations
	// This would typically run as a background job
	c.JSON(http.StatusOK, gin.H{"message": "recommendation generation triggered"})
}
