package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxDeliveryListLimit = 200

type DeliveryHandler struct {
	service *delivery.Service
}

type deliveryListItemResponse struct {
	ID                      uint                           `json:"id"`
	AlertID                 string                         `json:"alert_id"`
	TraceID                 string                         `json:"trace_id"`
	ChannelID               uint                           `json:"channel_id"`
	RouteRuleID             *uint                          `json:"route_rule_id,omitempty"`
	DeliveryStatus          string                         `json:"delivery_status"`
	DeliveryMode            string                         `json:"delivery_mode"`
	AttemptCount            int                            `json:"attempt_count"`
	FinalFailureSummary     *models.FinalFailureSummary    `json:"final_failure_summary,omitempty"`
	AlertSnapshot           models.AlertSnapshot           `json:"alert_snapshot"`
	ChannelSnapshot         models.ChannelSnapshot         `json:"channel_snapshot"`
	RouteSnapshot           *models.RouteSnapshot          `json:"route_snapshot,omitempty"`
	RenderedPayloadSnapshot models.RenderedPayloadSnapshot `json:"rendered_payload_snapshot"`
	LastAttemptAt           *time.Time                     `json:"last_attempt_at,omitempty"`
	LastSuccessAt           *time.Time                     `json:"last_success_at,omitempty"`
	CreatedAt               time.Time                      `json:"created_at"`
	UpdatedAt               time.Time                      `json:"updated_at"`
	Attempts                []deliveryAttemptResponse      `json:"attempts"`
}

type deliveryDetailResponse struct {
	ID                      uint                           `json:"id"`
	AlertID                 string                         `json:"alert_id"`
	TraceID                 string                         `json:"trace_id"`
	ChannelID               uint                           `json:"channel_id"`
	RouteRuleID             *uint                          `json:"route_rule_id,omitempty"`
	DeliveryStatus          string                         `json:"delivery_status"`
	DeliveryMode            string                         `json:"delivery_mode"`
	AttemptCount            int                            `json:"attempt_count"`
	FinalFailureSummary     *models.FinalFailureSummary    `json:"final_failure_summary,omitempty"`
	AlertSnapshot           models.AlertSnapshot           `json:"alert_snapshot"`
	ChannelSnapshot         models.ChannelSnapshot         `json:"channel_snapshot"`
	RouteSnapshot           *models.RouteSnapshot          `json:"route_snapshot,omitempty"`
	RenderedPayloadSnapshot models.RenderedPayloadSnapshot `json:"rendered_payload_snapshot"`
	LastAttemptAt           *time.Time                     `json:"last_attempt_at,omitempty"`
	LastSuccessAt           *time.Time                     `json:"last_success_at,omitempty"`
	CreatedAt               time.Time                      `json:"created_at"`
	UpdatedAt               time.Time                      `json:"updated_at"`
	Attempts                []deliveryAttemptResponse      `json:"attempts"`
}

type deliveryAttemptResponse struct {
	ID            uint      `json:"id"`
	AttemptNumber int       `json:"attempt_number"`
	Result        string    `json:"result"`
	Retryable     bool      `json:"retryable"`
	ErrorMessage  string    `json:"error_message"`
	HTTPStatus    *int      `json:"http_status,omitempty"`
	DurationMS    int64     `json:"duration_ms"`
	TriggerKind   string    `json:"trigger_kind"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewDeliveryHandler(service *delivery.Service) *DeliveryHandler {
	return &DeliveryHandler{service: service}
}

func (h *DeliveryHandler) List(c *gin.Context) {
	input, err := buildListDeliveriesInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deliveries, total, err := h.service.ListDeliveries(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items, err := mapDeliveries(deliveries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  items,
		"total": total,
	})
}

func (h *DeliveryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delivery id"})
		return
	}

	record, err := h.service.GetDeliveryByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "delivery not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := mapDeliveryDetail(*record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func buildListDeliveriesInput(c *gin.Context) (delivery.ListDeliveriesInput, error) {
	var input delivery.ListDeliveriesInput
	input.AlertID = c.Query("alert_id")
	input.TraceID = c.Query("trace_id")
	input.DeliveryStatus = c.Query("delivery_status")

	if channelID := c.Query("channel_id"); channelID != "" {
		parsed, err := strconv.ParseUint(channelID, 10, 64)
		if err != nil {
			return input, errors.New("invalid channel_id")
		}
		input.ChannelID = uint(parsed)
	}

	if createdFrom := c.Query("created_from"); createdFrom != "" {
		parsed, err := time.Parse(time.RFC3339, createdFrom)
		if err != nil {
			return input, errors.New("invalid created_from")
		}
		input.CreatedAfter = &parsed
	}

	if createdTo := c.Query("created_to"); createdTo != "" {
		parsed, err := time.Parse(time.RFC3339, createdTo)
		if err != nil {
			return input, errors.New("invalid created_to")
		}
		input.CreatedBefore = &parsed
	}

	limit := 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			return input, errors.New("invalid limit")
		}
		if parsed <= 0 || parsed > maxDeliveryListLimit {
			return input, errors.New("limit must be between 1 and 200")
		}
		limit = parsed
	}
	input.Limit = limit

	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil {
			return input, errors.New("invalid offset")
		}
		if parsed < 0 {
			return input, errors.New("offset must be greater than or equal to 0")
		}
		input.Offset = parsed
	}

	return input, nil
}

func mapDeliveries(deliveries []models.NotificationDelivery) ([]deliveryListItemResponse, error) {
	items := make([]deliveryListItemResponse, 0, len(deliveries))
	for _, record := range deliveries {
		item, err := mapDeliveryListItem(record)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func mapDeliveryListItem(record models.NotificationDelivery) (deliveryListItemResponse, error) {
	common, err := mapDeliveryCommon(record)
	if err != nil {
		return deliveryListItemResponse{}, err
	}

	return deliveryListItemResponse(common), nil
}

func mapDeliveryDetail(record models.NotificationDelivery) (deliveryDetailResponse, error) {
	common, err := mapDeliveryCommon(record)
	if err != nil {
		return deliveryDetailResponse{}, err
	}

	return deliveryDetailResponse(common), nil
}

func mapDeliveryCommon(record models.NotificationDelivery) (deliveryDetailResponse, error) {
	var alertSnapshot models.AlertSnapshot
	if err := unmarshalOptionalJSON(record.AlertSnapshot, &alertSnapshot); err != nil {
		return deliveryDetailResponse{}, err
	}

	var channelSnapshot models.ChannelSnapshot
	if err := unmarshalOptionalJSON(record.ChannelSnapshot, &channelSnapshot); err != nil {
		return deliveryDetailResponse{}, err
	}

	var routeSnapshot *models.RouteSnapshot
	if !isNullJSON(record.RouteSnapshot) {
		var decoded models.RouteSnapshot
		if err := unmarshalOptionalJSON(record.RouteSnapshot, &decoded); err != nil {
			return deliveryDetailResponse{}, err
		}
		routeSnapshot = &decoded
	}

	var renderedPayload models.RenderedPayloadSnapshot
	if err := unmarshalOptionalJSON(record.RenderedPayloadSnapshot, &renderedPayload); err != nil {
		return deliveryDetailResponse{}, err
	}

	var finalFailureSummary *models.FinalFailureSummary
	if !isNullJSON(record.FinalFailureSummary) {
		var decoded models.FinalFailureSummary
		if err := unmarshalOptionalJSON(record.FinalFailureSummary, &decoded); err != nil {
			return deliveryDetailResponse{}, err
		}
		finalFailureSummary = &decoded
	}

	attempts := make([]deliveryAttemptResponse, 0, len(record.Attempts))
	for _, attempt := range record.Attempts {
		attempts = append(attempts, deliveryAttemptResponse{
			ID:            attempt.ID,
			AttemptNumber: attempt.AttemptNumber,
			Result:        attempt.Result,
			Retryable:     attempt.Retryable,
			ErrorMessage:  attempt.ErrorMessage,
			HTTPStatus:    attempt.HTTPStatus,
			DurationMS:    attempt.DurationMS,
			TriggerKind:   attempt.TriggerKind,
			CreatedAt:     attempt.CreatedAt,
		})
	}

	return deliveryDetailResponse{
		ID:                      record.ID,
		AlertID:                 record.AlertID,
		TraceID:                 record.TraceID,
		ChannelID:               record.ChannelID,
		RouteRuleID:             record.RouteRuleID,
		DeliveryStatus:          record.DeliveryStatus,
		DeliveryMode:            record.DeliveryMode,
		AttemptCount:            record.AttemptCount,
		FinalFailureSummary:     finalFailureSummary,
		AlertSnapshot:           alertSnapshot,
		ChannelSnapshot:         channelSnapshot,
		RouteSnapshot:           routeSnapshot,
		RenderedPayloadSnapshot: renderedPayload,
		LastAttemptAt:           record.LastAttemptAt,
		LastSuccessAt:           record.LastSuccessAt,
		CreatedAt:               record.CreatedAt,
		UpdatedAt:               record.UpdatedAt,
		Attempts:                attempts,
	}, nil
}

func unmarshalOptionalJSON(raw []byte, target interface{}) error {
	if len(raw) == 0 || isNullJSON(raw) {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func isNullJSON(raw []byte) bool {
	return string(raw) == "null"
}
