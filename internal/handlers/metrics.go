package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

type MetricsHandler struct {
	db *gorm.DB
}

func NewMetricsHandler(db *gorm.DB) *MetricsHandler {
	return &MetricsHandler{db: db}
}

type MetricsResponse struct {
	Period                           string `json:"period"`
	WebhookIngestTotal               int64  `json:"webhook_ingest_total"`
	NotificationSendSuccessTotal     int64  `json:"notification_send_success_total"`
	NotificationSendFailureTotal     int64  `json:"notification_send_failure_total"`
	NotificationRetryTotal           int64  `json:"notification_retry_total"`
	NotificationTerminalFailureTotal int64  `json:"notification_terminal_failure_total"`
}

func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	period := c.DefaultQuery("period", "24h")

	duration, err := time.ParseDuration(period)
	if err != nil {
		duration = 24 * time.Hour
	}
	since := time.Now().Add(-duration)

	var totalDeliveries int64
	var successCount int64
	var failureCount int64
	var retryCount int64
	var terminalFailureCount int64

	h.db.Model(&models.NotificationDelivery{}).
		Where("created_at >= ?", since).
		Count(&totalDeliveries)

	h.db.Model(&models.NotificationDelivery{}).
		Where("created_at >= ? AND delivery_status = ?", since, "delivered").
		Count(&successCount)

	h.db.Model(&models.NotificationDelivery{}).
		Where("created_at >= ? AND delivery_status = ?", since, "failed").
		Count(&failureCount)

	h.db.Model(&models.NotificationDelivery{}).
		Where("created_at >= ? AND attempt_count > ?", since, 1).
		Count(&retryCount)

	// Terminal failures: failed with attempt_count >= 3
	h.db.Model(&models.NotificationDelivery{}).
		Where("created_at >= ? AND delivery_status = ? AND attempt_count >= ?", since, "failed", 3).
		Count(&terminalFailureCount)

	c.JSON(http.StatusOK, MetricsResponse{
		Period:                           period,
		WebhookIngestTotal:               totalDeliveries,
		NotificationSendSuccessTotal:     successCount,
		NotificationSendFailureTotal:     failureCount,
		NotificationRetryTotal:           retryCount,
		NotificationTerminalFailureTotal: terminalFailureCount,
	})
}
