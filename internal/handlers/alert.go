package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/models"
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
	c.ShouldBindJSON(&input)

	var alert models.Alert
	if err := h.db.First(&alert, "alert_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if alert.Status != "firing" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alert cannot be acknowledged"})
		return
	}

	now := time.Now()
	alert.Status = "acked"
	alert.AckedAt = &now
	alert.AckComment = input.Comment

	if err := h.db.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	// Create a silence rule
	silence := models.SilenceRule{
		Name:            "Quick Silence - " + alert.AlertName,
		AlertNamePattern: alert.AlertName,
		Severities:       []byte(`["` + alert.Severity + `"]`),
		StartsAt:        time.Now(),
		EndsAt:          time.Now().Add(time.Duration(input.Duration) * time.Second),
		CreatedBy:       "system",
	}
	h.db.Create(&silence)

	c.JSON(http.StatusOK, gin.H{"message": "alert silenced", "silence_id": silence.ID})
}

// Get alert statistics
func (h *AlertHandler) Stats(c *gin.Context) {
	var stats struct {
		Total      int64                      `json:"total"`
		Firing     int64                      `json:"firing"`
		Acked      int64                      `json:"acked"`
		Silenced   int64                      `json:"silenced"`
		BySeverity map[string]int64            `json:"by_severity"`
		Trend      []struct {
			Time  string `json:"time"`
			Count int64  `json:"count"`
		} `json:"trend"`
	}
	stats.BySeverity = make(map[string]int64)

	h.db.Model(&models.Alert{}).Count(&stats.Total)
	h.db.Model(&models.Alert{}).Where("status = ?", "firing").Count(&stats.Firing)
	h.db.Model(&models.Alert{}).Where("status = ?", "acked").Count(&stats.Acked)
	h.db.Model(&models.Alert{}).Where("status = ?", "silenced").Count(&stats.Silenced)

	// By severity (firing only)
	for _, sev := range []string{"P0", "P1", "P2", "P3"} {
		var count int64
		h.db.Model(&models.Alert{}).Where("severity = ? AND status = ?", sev, "firing").Count(&count)
		stats.BySeverity[sev] = count
	}

	// Trend - last 24 hours
	for i := 23; i >= 0; i-- {
		hour := time.Now().Add(-time.Duration(i) * time.Hour)
		start := hour.Truncate(time.Hour)
		end := start.Add(time.Hour)
		var count int64
		h.db.Model(&models.Alert{}).Where("trigger_time >= ? AND trigger_time < ?", start, end).Count(&count)
		stats.Trend = append(stats.Trend, struct {
			Time  string `json:"time"`
			Count int64  `json:"count"`
		}{
			Time:  start.Format("15:04"),
			Count: count,
		})
	}

	c.JSON(http.StatusOK, stats)
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
