package handlers

import (
	"encoding/json"
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

// GroupedActiveAlert represents a group of alerts with the same fingerprint.
type GroupedActiveAlert struct {
	Fingerprint      string       `json:"fingerprint"`
	LatestAlert      models.Alert `json:"latest_alert"`
	Count            int          `json:"count"`
	FirstTriggeredAt time.Time    `json:"first_triggered_at"`
	LastTriggeredAt  time.Time    `json:"last_triggered_at"`
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

func (h *AlertHandler) BatchAck(c *gin.Context) {
	var input struct {
		AlertIDs []string `json:"alert_ids" binding:"required"`
		Comment  string   `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := 0
	skipped := 0
	var errs []string
	username := middleware.GetUsername(c)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range input.AlertIDs {
			var alert models.Alert
			if err := tx.First(&alert, "alert_id = ?", id).Error; err != nil {
				skipped++
				errs = append(errs, fmt.Sprintf("%s: not found", id))
				continue
			}
			if err := alert.Ack(username, input.Comment); err != nil {
				skipped++
				errs = append(errs, fmt.Sprintf("%s: %s", id, err.Error()))
				_ = recordAudit(tx, c, "alert.batch_ack", "alert", alert.AlertID, auditResultDenied, err.Error())
				continue
			}
			if err := tx.Save(&alert).Error; err != nil {
				skipped++
				errs = append(errs, fmt.Sprintf("%s: db error", id))
				continue
			}
			_ = recordAudit(tx, c, "alert.batch_ack", "alert", alert.AlertID, auditResultAllowed, fmt.Sprintf("comment=%s", input.Comment))
			updated++
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": updated, "skipped": skipped, "errors": errs})
}

func (h *AlertHandler) BatchSilence(c *gin.Context) {
	var input struct {
		AlertIDs []string `json:"alert_ids" binding:"required"`
		Duration int      `json:"duration" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := 0
	skipped := 0
	var errs []string
	username := middleware.GetUsername(c)
	if username == "" {
		username = "system"
	}
	now := time.Now()
	alertNameSeverities := make(map[string]map[string]bool)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range input.AlertIDs {
			var alert models.Alert
			if err := tx.First(&alert, "alert_id = ?", id).Error; err != nil {
				skipped++
				errs = append(errs, fmt.Sprintf("%s: not found", id))
				continue
			}
			if alert.Status != "firing" && alert.Status != "acked" {
				skipped++
				errs = append(errs, fmt.Sprintf("%s: cannot silence %s alert", id, alert.Status))
				_ = recordAudit(tx, c, "alert.batch_silence", "alert", alert.AlertID, auditResultDenied, fmt.Sprintf("cannot silence %s alert", alert.Status))
				continue
			}
			alert.Status = "silenced"
			if err := tx.Save(&alert).Error; err != nil {
				skipped++
				errs = append(errs, fmt.Sprintf("%s: db error", id))
				continue
			}
			if alertNameSeverities[alert.AlertName] == nil {
				alertNameSeverities[alert.AlertName] = make(map[string]bool)
			}
			alertNameSeverities[alert.AlertName][alert.Severity] = true
			_ = recordAudit(tx, c, "alert.batch_silence", "alert", alert.AlertID, auditResultAllowed, fmt.Sprintf("duration_seconds=%d", input.Duration))
			updated++
		}

		for name, severities := range alertNameSeverities {
			var sevList []string
			for s := range severities {
				sevList = append(sevList, s)
			}
			sevJSON, _ := json.Marshal(sevList)
			silence := models.SilenceRule{
				Name:             "Batch Silence - " + name,
				AlertNamePattern: name,
				Severities:       sevJSON,
				StartsAt:         now,
				EndsAt:           now.Add(time.Duration(input.Duration) * time.Second),
				CreatedBy:        username,
			}
			if err := tx.Create(&silence).Error; err != nil {
				errs = append(errs, fmt.Sprintf("silence rule for %s: db error", name))
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": updated, "skipped": skipped, "errors": errs})
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
			Time:  t.Time.Local().Format("15:04"),
			Count: int64(t.Count),
		})
	}

	c.JSON(http.StatusOK, response)
}

// Get active alerts. Supports ?grouped=true to group by fingerprint.
func (h *AlertHandler) Active(c *gin.Context) {
	var alerts []models.Alert
	if err := h.db.Where("status = ?", "firing").
		Order("CASE severity WHEN 'P0' THEN 0 WHEN 'P1' THEN 1 WHEN 'P2' THEN 2 ELSE 3 END, trigger_time DESC").
		Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.Query("grouped") == "true" {
		grouped := groupAlertsByFingerprint(alerts)
		c.JSON(http.StatusOK, grouped)
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// groupAlertsByFingerprint groups alerts by fingerprint and returns GroupedActiveAlert slices.
// For each group, the alert with the most recent trigger_time is selected as the latest alert.
func groupAlertsByFingerprint(alerts []models.Alert) []GroupedActiveAlert {
	groups := make(map[string][]models.Alert)
	// Preserve insertion order for deterministic output
	var order []string

	for _, a := range alerts {
		fp := a.Fingerprint
		if _, exists := groups[fp]; !exists {
			order = append(order, fp)
		}
		groups[fp] = append(groups[fp], a)
	}

	result := make([]GroupedActiveAlert, 0, len(groups))
	for _, fp := range order {
		alertGroup := groups[fp]
		var latest models.Alert
		var firstTriggered, lastTriggered time.Time

		for i, a := range alertGroup {
			if i == 0 || a.TriggerTime.After(lastTriggered) {
				latest = a
				lastTriggered = a.TriggerTime
			}
			if i == 0 || a.TriggerTime.Before(firstTriggered) {
				firstTriggered = a.TriggerTime
			}
		}

		result = append(result, GroupedActiveAlert{
			Fingerprint:      fp,
			LatestAlert:      latest,
			Count:            len(alertGroup),
			FirstTriggeredAt: firstTriggered,
			LastTriggeredAt:  lastTriggered,
		})
	}

	return result
}

// AlertDeliveries returns the notification delivery records for a specific alert.
func (h *AlertHandler) AlertDeliveries(c *gin.Context) {
	alertID := c.Param("id")
	var deliveries []models.NotificationDelivery
	err := h.db.Where("alert_id = ?", alertID).
		Preload("Attempts", func(db *gorm.DB) *gorm.DB {
			return db.Order("attempt_number ASC")
		}).
		Order("created_at DESC").
		Find(&deliveries).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deliveries)
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
