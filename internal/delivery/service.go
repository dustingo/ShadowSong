package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type StartDeliveryInput struct {
	Alert         *models.Alert
	Channel       *models.Channel
	RouteRule     *models.RouteRule
	DeliveryMode  string
	TriggerKind   string
	RenderedTitle string
	RenderedBody  string
}

type RecordAttemptInput struct {
	AttemptNumber int
	Result        string
	Retryable     bool
	ErrorMessage  string
	HTTPStatus    *int
	DurationMS    int64
	TriggerKind   string
}

type MarkDeliveredInput struct {
	AttemptCount int
	DeliveredAt  time.Time
}

type MarkFailedInput struct {
	AttemptCount int
	FailedAt     time.Time
	Result       string
	Retryable    bool
	ErrorMessage string
	HTTPStatus   *int
	TriggerKind  string
}

type ListDeliveriesInput struct {
	AlertID        string
	TraceID        string
	ChannelID      uint
	DeliveryStatus string
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	Limit          int
	Offset         int
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) StartDelivery(ctx context.Context, input StartDeliveryInput) (*models.NotificationDelivery, error) {
	if s.db == nil {
		return nil, errors.New("delivery service requires db")
	}
	if err := validateStartDeliveryInput(input); err != nil {
		return nil, err
	}

	alertSnapshot, err := marshalJSON(models.AlertSnapshot{
		AlertID:     input.Alert.AlertID,
		TraceID:     input.Alert.TraceID,
		Source:      input.Alert.Source,
		AlertName:   input.Alert.AlertName,
		Severity:    input.Alert.Severity,
		Message:     input.Alert.Message,
		Status:      input.Alert.Status,
		Fingerprint: input.Alert.Fingerprint,
		TriggerTime: input.Alert.TriggerTime.Format(time.RFC3339),
		Labels:      decodeJSONMap(input.Alert.Labels),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal alert snapshot: %w", err)
	}

	channelSnapshot, err := marshalJSON(models.ChannelSnapshot{
		ID:      input.Channel.ID,
		Name:    input.Channel.Name,
		Type:    input.Channel.Type,
		Enabled: input.Channel.Enabled,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal channel snapshot: %w", err)
	}

	routeSnapshot := datatypes.JSON([]byte("null"))
	if input.RouteRule != nil {
		channelIDs, err := decodeChannelIDs(input.RouteRule.ChannelIDs)
		if err != nil {
			return nil, fmt.Errorf("decode route snapshot channel ids: %w", err)
		}
		routeSnapshot, err = marshalJSON(models.RouteSnapshot{
			ID:         input.RouteRule.ID,
			Name:       input.RouteRule.Name,
			Priority:   input.RouteRule.Priority,
			Enabled:    input.RouteRule.Enabled,
			ChannelIDs: channelIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal route snapshot: %w", err)
		}
	}

	renderedPayloadSnapshot, err := marshalJSON(models.RenderedPayloadSnapshot{
		Title:   input.RenderedTitle,
		Content: input.RenderedBody,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal rendered payload snapshot: %w", err)
	}

	delivery := &models.NotificationDelivery{
		AlertID:                 input.Alert.AlertID,
		TraceID:                 input.Alert.TraceID,
		ChannelID:               input.Channel.ID,
		DeliveryMode:            input.DeliveryMode,
		AlertSnapshot:           alertSnapshot,
		ChannelSnapshot:         channelSnapshot,
		RouteSnapshot:           routeSnapshot,
		RenderedPayloadSnapshot: renderedPayloadSnapshot,
	}
	if input.RouteRule != nil {
		delivery.RouteRuleID = &input.RouteRule.ID
	}

	if err := s.db.WithContext(ctx).Create(delivery).Error; err != nil {
		return nil, fmt.Errorf("create delivery: %w", err)
	}

	return delivery, nil
}

func (s *Service) RecordAttempt(ctx context.Context, deliveryID uint, input RecordAttemptInput) (*models.NotificationDeliveryAttempt, error) {
	if s.db == nil {
		return nil, errors.New("delivery service requires db")
	}
	if deliveryID == 0 {
		return nil, errors.New("delivery id is required")
	}
	if _, ok := validTriggerKind(input.TriggerKind); !ok {
		return nil, errors.New("invalid trigger_kind")
	}

	attempt := &models.NotificationDeliveryAttempt{
		DeliveryID:    deliveryID,
		AttemptNumber: input.AttemptNumber,
		Result:        input.Result,
		Retryable:     input.Retryable,
		ErrorMessage:  input.ErrorMessage,
		HTTPStatus:    input.HTTPStatus,
		DurationMS:    input.DurationMS,
		TriggerKind:   input.TriggerKind,
	}

	if err := s.db.WithContext(ctx).Create(attempt).Error; err != nil {
		return nil, fmt.Errorf("create delivery attempt: %w", err)
	}

	return attempt, nil
}

func (s *Service) MarkDelivered(ctx context.Context, deliveryID uint, input MarkDeliveredInput) error {
	if s.db == nil {
		return errors.New("delivery service requires db")
	}
	if deliveryID == 0 {
		return errors.New("delivery id is required")
	}
	if input.AttemptCount <= 0 {
		return errors.New("attempt_count must be greater than 0")
	}
	deliveredAt := input.DeliveredAt
	if deliveredAt.IsZero() {
		deliveredAt = time.Now()
	}

	delivery, err := s.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("load delivery before mark delivered: %w", err)
	}

	delivery.DeliveryStatus = models.DeliveryStatusDelivered
	delivery.AttemptCount = input.AttemptCount
	delivery.LastAttemptAt = &deliveredAt
	delivery.LastSuccessAt = &deliveredAt
	delivery.FinalFailureSummary = datatypes.JSON([]byte("null"))

	if err := s.db.WithContext(ctx).Save(delivery).Error; err != nil {
		return fmt.Errorf("mark delivery delivered: %w", err)
	}

	return nil
}

func (s *Service) MarkFailed(ctx context.Context, deliveryID uint, input MarkFailedInput) error {
	if s.db == nil {
		return errors.New("delivery service requires db")
	}
	if deliveryID == 0 {
		return errors.New("delivery id is required")
	}
	if input.AttemptCount <= 0 {
		return errors.New("attempt_count must be greater than 0")
	}
	if _, ok := validTriggerKind(input.TriggerKind); !ok {
		return errors.New("invalid trigger_kind")
	}

	failedAt := input.FailedAt
	if failedAt.IsZero() {
		failedAt = time.Now()
	}

	finalFailureSummary, err := marshalJSON(models.FinalFailureSummary{
		Result:       input.Result,
		Retryable:    input.Retryable,
		ErrorMessage: input.ErrorMessage,
		HTTPStatus:   input.HTTPStatus,
		AttemptCount: input.AttemptCount,
		TriggerKind:  input.TriggerKind,
	})
	if err != nil {
		return fmt.Errorf("marshal final failure summary: %w", err)
	}

	delivery, err := s.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("load delivery before mark failed: %w", err)
	}

	delivery.DeliveryStatus = models.DeliveryStatusFailed
	delivery.AttemptCount = input.AttemptCount
	delivery.LastAttemptAt = &failedAt
	delivery.FinalFailureSummary = finalFailureSummary

	if err := s.db.WithContext(ctx).Save(delivery).Error; err != nil {
		return fmt.Errorf("mark delivery failed: %w", err)
	}

	return nil
}

func (s *Service) GetDeliveryByID(ctx context.Context, deliveryID uint) (*models.NotificationDelivery, error) {
	if s.db == nil {
		return nil, errors.New("delivery service requires db")
	}
	if deliveryID == 0 {
		return nil, errors.New("delivery id is required")
	}

	var delivery models.NotificationDelivery
	if err := s.db.WithContext(ctx).
		Preload("Attempts", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("attempt_number ASC")
		}).
		First(&delivery, deliveryID).Error; err != nil {
		return nil, fmt.Errorf("get delivery by id: %w", err)
	}

	return &delivery, nil
}

func (s *Service) ListDeliveries(ctx context.Context, input ListDeliveriesInput) ([]models.NotificationDelivery, error) {
	if s.db == nil {
		return nil, errors.New("delivery service requires db")
	}

	query := s.db.WithContext(ctx).
		Model(&models.NotificationDelivery{}).
		Preload("Attempts", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("attempt_number ASC")
		}).
		Order("created_at DESC")

	if input.AlertID != "" {
		query = query.Where("alert_id = ?", input.AlertID)
	}
	if input.TraceID != "" {
		query = query.Where("trace_id = ?", input.TraceID)
	}
	if input.ChannelID != 0 {
		query = query.Where("channel_id = ?", input.ChannelID)
	}
	if input.DeliveryStatus != "" {
		query = query.Where("delivery_status = ?", input.DeliveryStatus)
	}
	if input.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *input.CreatedAfter)
	}
	if input.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *input.CreatedBefore)
	}
	if input.Limit <= 0 {
		input.Limit = 50
	}

	var deliveries []models.NotificationDelivery
	if err := query.Limit(input.Limit).Offset(input.Offset).Find(&deliveries).Error; err != nil {
		return nil, fmt.Errorf("list deliveries: %w", err)
	}

	return deliveries, nil
}

func validateStartDeliveryInput(input StartDeliveryInput) error {
	if input.Alert == nil {
		return errors.New("alert is required")
	}
	if input.Channel == nil {
		return errors.New("channel is required")
	}
	if input.DeliveryMode == "" {
		return errors.New("delivery_mode is required")
	}
	if _, ok := validTriggerKind(input.TriggerKind); !ok {
		return errors.New("invalid trigger_kind")
	}
	if input.RenderedTitle == "" && input.RenderedBody == "" {
		return errors.New("rendered payload is required")
	}
	return nil
}

func validTriggerKind(triggerKind string) (string, bool) {
	switch triggerKind {
	case models.TriggerKindPipeline, models.TriggerKindRetry, models.TriggerKindReplay:
		return triggerKind, true
	default:
		return "", false
	}
}

func marshalJSON(v interface{}) (datatypes.JSON, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(encoded), nil
}

func decodeJSONMap(raw []byte) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil || decoded == nil {
		return map[string]interface{}{}
	}

	return decoded
}

func decodeChannelIDs(raw []byte) ([]uint, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var channelIDs []uint
	if err := json.Unmarshal(raw, &channelIDs); err != nil {
		return nil, err
	}

	return channelIDs, nil
}
