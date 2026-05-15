package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/stats"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AlertHandler struct {
	db *gorm.DB
}

func NewAlertHandler(db *gorm.DB) *AlertHandler {
	return &AlertHandler{db: db}
}

// List alerts with filters
func (h *AlertHandler) List(c *gin.Context) {
	var alerts []models.Alert
	query := h.db.Model(&models.Alert{})

	// Apply filters
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime := c.Query("start_time"); startTime != "" {
		query = query.Where("trigger_time >= ?", startTime)
	}
	if endTime := c.Query("end_time"); endTime != "" {
		query = query.Where("trigger_time <= ?", endTime)
	}
	if labelSelector := c.Query("label_selector"); labelSelector != "" {
		query = applyLabelSelector(query, labelSelector)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	query.Count(&total)

	query = query.Order("trigger_time DESC").Offset(offset).Limit(pageSize)
	if err := query.Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  alerts,
		"total": total,
	})
}

// Get alert by ID
func (h *AlertHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var alert models.Alert
	if err := h.db.First(&alert, "alert_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alert)
}

// Acknowledge an alert
func (h *AlertHandler) Ack(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var alert models.Alert
	if err := h.db.First(&alert, "alert_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	username := middleware.GetUsername(c)
	if err := alert.Ack(username, input.Comment); err != nil {
		_ = recordAudit(h.db, c, "alert.ack", "alert", alert.AlertID, auditResultDenied, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert cannot be acknowledged"})
		return
	}

	if err := h.db.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "alert.ack", "alert", alert.AlertID, auditResultAllowed, fmt.Sprintf("comment=%s", input.Comment))
	c.JSON(http.StatusOK, alert)
}

// Quick silence an alert
func (h *AlertHandler) QuickSilence(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Duration int `json:"duration" binding:"required"` // in seconds
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var alert models.Alert
	if err := h.db.First(&alert, "alert_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	alert.Status = "silenced"
	if err := h.db.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	createdBy := middleware.GetUsername(c)
	if createdBy == "" {
		createdBy = "system"
	}

	// Create a silence rule
	silence := models.SilenceRule{
		Name:             "Quick Silence - " + alert.AlertName,
		AlertNamePattern: alert.AlertName,
		Severities:       []byte(`["` + alert.Severity + `"]`),
		StartsAt:         time.Now(),
		EndsAt:           time.Now().Add(time.Duration(input.Duration) * time.Second),
		CreatedBy:        createdBy,
	}
	if err := h.db.Create(&silence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "alert.quick_silence", "alert", alert.AlertID, auditResultAllowed, fmt.Sprintf("silence_id=%d duration_seconds=%d", silence.ID, input.Duration))
	c.JSON(http.StatusOK, gin.H{"message": "alert silenced", "silence_id": silence.ID})
}

// Get alert statistics
func (h *AlertHandler) Stats(c *gin.Context) {
	alertStats, err := stats.GetAlertStats(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to API response format with string time for trend
	response := struct {
		Total      int64            `json:"total"`
		Firing     int64            `json:"firing"`
		Acked      int64            `json:"acked"`
		Silenced   int64            `json:"silenced"`
		BySeverity map[string]int64 `json:"by_severity"`
		Trend      []struct {
			Time  string `json:"time"`
			Count int64  `json:"count"`
		} `json:"trend"`
	}{
		Total:      int64(alertStats.Total),
		Firing:     int64(alertStats.Firing),
		Acked:      int64(alertStats.Acked),
		Silenced:   int64(alertStats.Silenced),
		BySeverity: make(map[string]int64),
	}

	// Convert by_severity
	for k, v := range alertStats.BySeverity {
		response.BySeverity[k] = int64(v)
	}

	// Convert trend with formatted time string
	for _, t := range alertStats.Trend {
		response.Trend = append(response.Trend, struct {
			Time  string `json:"time"`
			Count int64  `json:"count"`
		}{
			Time:  t.Time.Format("15:04"),
			Count: int64(t.Count),
		})
	}

	c.JSON(http.StatusOK, response)
}

// Get active alerts
func (h *AlertHandler) Active(c *gin.Context) {
	var alerts []models.Alert
	if err := h.db.Where("status = ?", "firing").
		Order("CASE severity WHEN 'P0' THEN 0 WHEN 'P1' THEN 1 WHEN 'P2' THEN 2 ELSE 3 END, trigger_time DESC").
		Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

// applyLabelSelector parses a label selector string and applies JSON-based filtering.
// Format: "key1=value1,key2=~regex2" (~ prefix means regex match on value).
func applyLabelSelector(query *gorm.DB, selector string) *gorm.DB {
	pairs := strings.Split(selector, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eqIdx := strings.Index(pair, "=")
		if eqIdx < 1 {
			continue
		}
		key := strings.TrimSpace(pair[:eqIdx])
		value := strings.TrimSpace(pair[eqIdx+1:])
		if key == "" {
			continue
		}
		if strings.HasPrefix(value, "~") {
			pattern := value[1:]
			if _, err := regexp.Compile(pattern); err != nil {
				continue
			}
			query = query.Where("labels->>? ~ ?", key, pattern)
		} else {
			query = query.Where("labels->>? = ?", key, value)
		}
	}
	return query
}
