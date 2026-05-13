package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

type ChannelHealthHandler struct {
	db *gorm.DB
}

func NewChannelHealthHandler(db *gorm.DB) *ChannelHealthHandler {
	return &ChannelHealthHandler{db: db}
}

type ChannelHealthResponse struct {
	ChannelID       uint              `json:"channel_id"`
	ChannelName     string            `json:"channel_name"`
	Period          string            `json:"period"`
	TotalDeliveries int64             `json:"total_deliveries"`
	Successful      int64             `json:"successful"`
	Failed          int64             `json:"failed"`
	SuccessRate     float64           `json:"success_rate"`
	LastFailure     *LastFailureInfo  `json:"last_failure,omitempty"`
}

type LastFailureInfo struct {
	DeliveryID   uint      `json:"delivery_id"`
	ErrorMessage string    `json:"error_message"`
	FailedAt     time.Time `json:"failed_at"`
}

func (h *ChannelHealthHandler) GetChannelHealth(c *gin.Context) {
	channelID := c.Param("id")
	period := c.DefaultQuery("period", "24h")

	duration, err := time.ParseDuration(period)
	if err != nil {
		duration = 24 * time.Hour
	}
	since := time.Now().Add(-duration)

	// Get channel info
	var channel models.Channel
	if err := h.db.First(&channel, channelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	// Aggregate deliveries
	var total, successful, failed int64

	h.db.Model(&models.NotificationDelivery{}).
		Where("channel_id = ? AND created_at >= ?", channel.ID, since).
		Count(&total)

	h.db.Model(&models.NotificationDelivery{}).
		Where("channel_id = ? AND created_at >= ? AND delivery_status = ?", channel.ID, since, "delivered").
		Count(&successful)

	h.db.Model(&models.NotificationDelivery{}).
		Where("channel_id = ? AND created_at >= ? AND delivery_status = ?", channel.ID, since, "failed").
		Count(&failed)

	// Calculate success rate
	var successRate float64
	if total > 0 {
		successRate = float64(successful) / float64(total)
	}

	// Get last failure within the period
	var lastFailure *LastFailureInfo
	var lastFailedDelivery models.NotificationDelivery
	if err := h.db.Where("channel_id = ? AND delivery_status = ? AND created_at >= ?", channel.ID, "failed", since).
		Order("created_at DESC").
		First(&lastFailedDelivery).Error; err == nil {

		errorMsg := ""
		if lastFailedDelivery.FinalFailureSummary != nil {
			var summary models.FinalFailureSummary
			if err := json.Unmarshal(lastFailedDelivery.FinalFailureSummary, &summary); err == nil {
				errorMsg = summary.ErrorMessage
			}
		}

		lastFailure = &LastFailureInfo{
			DeliveryID:   lastFailedDelivery.ID,
			ErrorMessage: errorMsg,
			FailedAt:     lastFailedDelivery.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, ChannelHealthResponse{
		ChannelID:       channel.ID,
		ChannelName:     channel.Name,
		Period:          period,
		TotalDeliveries: total,
		Successful:      successful,
		Failed:          failed,
		SuccessRate:     successRate,
		LastFailure:     lastFailure,
	})
}