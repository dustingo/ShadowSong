package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/notifier"
	"github.com/game-ops/ai-alert-system/internal/routing"
	"github.com/game-ops/ai-alert-system/internal/template"
	"github.com/game-ops/ai-alert-system/internal/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Service struct {
	db            *gorm.DB
	matcher       *routing.Matcher
	renderer      *template.Renderer
	sendToChannel func(channel *models.Channel, title, content string, data map[string]interface{}) error
	sleep         func(time.Duration)
}

// gormAdapter wraps *gorm.DB to implement routing.dbLoader interface
type gormAdapter struct {
	db *gorm.DB
}

func (a *gormAdapter) Find(dst interface{}, conds ...interface{}) error {
	return a.db.Find(dst, conds...).Error
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

type RetryDeliveryInput struct {
	OriginalDeliveryID uint
	Reason             string
	ActorUserID        uint
	ActorUsername      string
	ActorRole          string
}

type ReplayDeliveryInput struct {
	OriginalDeliveryID uint
	Reason             string
	ActorUserID        uint
	ActorUsername      string
	ActorRole          string
}

type RecoveryResult struct {
	Action         string
	Recovery       *models.NotificationDeliveryRecovery
	ResultDelivery *models.NotificationDelivery
}

type recoveredExecutionInput struct {
	alert        *models.Alert
	channel      *models.Channel
	routeRule    *models.RouteRule
	title        string
	content      string
	deliveryMode string
	triggerKind  string
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:            db,
		matcher:       routing.NewMatcher(&gormAdapter{db: db}),
		renderer:      template.NewRenderer(),
		sendToChannel: notifier.SendToChannel,
		sleep:         time.Sleep,
	}
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
		Labels:      utils.DecodeJSONMap(input.Alert.Labels),
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

func (s *Service) ListDeliveries(ctx context.Context, input ListDeliveriesInput) ([]models.NotificationDelivery, int64, error) {
	if s.db == nil {
		return nil, 0, errors.New("delivery service requires db")
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
	if input.Offset < 0 {
		input.Offset = 0
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count deliveries: %w", err)
	}

	var deliveries []models.NotificationDelivery
	if err := query.Limit(input.Limit).Offset(input.Offset).Find(&deliveries).Error; err != nil {
		return nil, 0, fmt.Errorf("list deliveries: %w", err)
	}

	return deliveries, total, nil
}

func (s *Service) RetryDelivery(ctx context.Context, input RetryDeliveryInput) (*RecoveryResult, error) {
	if s.db == nil {
		return nil, errors.New("delivery service requires db")
	}
	if err := validateRecoveryActor(input.OriginalDeliveryID, input.Reason, input.ActorUserID, input.ActorUsername, input.ActorRole); err != nil {
		return nil, err
	}

	var (
		result      *RecoveryResult
		recoveryErr error
	)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		original, err := loadDeliveryForRecovery(ctx, tx, input.OriginalDeliveryID)
		if err != nil {
			return err
		}

		recovery, err := s.startRecoveryRecord(tx, original.ID, models.TriggerKindRetry, input.Reason, input.ActorUserID, input.ActorUsername, input.ActorRole)
		result = &RecoveryResult{Action: models.TriggerKindRetry, Recovery: recovery}
		if err != nil {
			return err
		}
		if recovery.Status == models.RecoveryStatusRejected {
			recoveryErr = errors.New(recovery.ErrorMessage)
			return nil
		}

		if err := ensureDeliveryRecoverable(original); err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusRejected, err.Error())
		}

		channel, err := s.loadRetryChannel(ctx, tx, original.ChannelID)
		if err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, err.Error())
		}

		alert, routeRule, payload, err := decodeRetryExecution(original)
		if err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, err.Error())
		}

		if err := tx.Save(recoveryInProgress(recovery)).Error; err != nil {
			return fmt.Errorf("mark recovery in progress: %w", err)
		}

		resultDelivery, execErr := s.executeRecoveredDelivery(ctx, tx, recoveredExecutionInput{
			alert:        alert,
			channel:      channel,
			routeRule:    routeRule,
			title:        payload.Title,
			content:      payload.Content,
			deliveryMode: original.DeliveryMode,
			triggerKind:  models.TriggerKindRetry,
		})
		result.ResultDelivery = resultDelivery
		if execErr != nil {
			recoveryErr = execErr
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, execErr.Error())
		}

		if err := s.completeRecovery(tx, recovery, resultDelivery.ID, models.RecoveryStatusSucceeded, ""); err != nil {
			return err
		}
		result.Recovery = recovery
		return nil
	})

	if err != nil {
		return result, err
	}
	return result, recoveryErr
}

func (s *Service) ReplayDelivery(ctx context.Context, input ReplayDeliveryInput) (*RecoveryResult, error) {
	if s.db == nil {
		return nil, errors.New("delivery service requires db")
	}
	if err := validateRecoveryActor(input.OriginalDeliveryID, input.Reason, input.ActorUserID, input.ActorUsername, input.ActorRole); err != nil {
		return nil, err
	}

	var (
		result      *RecoveryResult
		recoveryErr error
	)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		original, err := loadDeliveryForRecovery(ctx, tx, input.OriginalDeliveryID)
		if err != nil {
			return err
		}

		recovery, err := s.startRecoveryRecord(tx, original.ID, models.TriggerKindReplay, input.Reason, input.ActorUserID, input.ActorUsername, input.ActorRole)
		result = &RecoveryResult{Action: models.TriggerKindReplay, Recovery: recovery}
		if err != nil {
			return err
		}
		if recovery.Status == models.RecoveryStatusRejected {
			recoveryErr = errors.New(recovery.ErrorMessage)
			return nil
		}

		if err := ensureDeliveryRecoverable(original); err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusRejected, err.Error())
		}

		alert, err := s.loadReplayAlert(ctx, tx, original.AlertID)
		if err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, err.Error())
		}

		ds, err := s.loadReplayDataSource(ctx, tx, alert.Source)
		if err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, err.Error())
		}

		channel, routeRule, err := s.findReplayTarget(ctx, tx, alert, original.ChannelID)
		if err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, err.Error())
		}

		title, content, err := s.renderer.RenderAlert(string(ds.OutputTemplate), alert, routeRule)
		if err != nil {
			recoveryErr = err
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, err.Error())
		}

		if err := tx.Save(recoveryInProgress(recovery)).Error; err != nil {
			return fmt.Errorf("mark recovery in progress: %w", err)
		}

		resultDelivery, execErr := s.executeRecoveredDelivery(ctx, tx, recoveredExecutionInput{
			alert:        alert,
			channel:      channel,
			routeRule:    routeRule,
			title:        title,
			content:      content,
			deliveryMode: models.DeliveryModeRendered,
			triggerKind:  models.TriggerKindReplay,
		})
		result.ResultDelivery = resultDelivery
		if execErr != nil {
			recoveryErr = execErr
			return s.completeRecovery(tx, recovery, 0, models.RecoveryStatusFailed, execErr.Error())
		}

		if err := s.completeRecovery(tx, recovery, resultDelivery.ID, models.RecoveryStatusSucceeded, ""); err != nil {
			return err
		}
		result.Recovery = recovery
		return nil
	})

	if err != nil {
		return result, err
	}
	return result, recoveryErr
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
	case models.TriggerKindPipeline, models.TriggerKindRetry, models.TriggerKindReplay, models.TriggerKindEscalation:
		return triggerKind, true
	default:
		return "", false
	}
}

func validateRecoveryActor(originalDeliveryID uint, reason string, actorUserID uint, actorUsername, actorRole string) error {
	if originalDeliveryID == 0 {
		return errors.New("original delivery id is required")
	}
	if strings.TrimSpace(reason) == "" {
		return errors.New("reason is required")
	}
	if actorUserID == 0 || strings.TrimSpace(actorUsername) == "" || strings.TrimSpace(actorRole) == "" {
		return errors.New("recovery actor is required")
	}
	if len(strings.TrimSpace(reason)) > 512 {
		return errors.New("reason exceeds 512 characters")
	}
	return nil
}

func loadDeliveryForRecovery(ctx context.Context, tx *gorm.DB, deliveryID uint) (*models.NotificationDelivery, error) {
	var delivery models.NotificationDelivery
	if err := tx.WithContext(ctx).
		Preload("Attempts", func(db *gorm.DB) *gorm.DB { return db.Order("attempt_number ASC") }).
		First(&delivery, deliveryID).Error; err != nil {
		return nil, fmt.Errorf("load original delivery: %w", err)
	}
	return &delivery, nil
}

func ensureDeliveryRecoverable(original *models.NotificationDelivery) error {
	if original.DeliveryStatus != models.DeliveryStatusFailed {
		return errors.New("only failed deliveries can be recovered")
	}
	return nil
}

func (s *Service) startRecoveryRecord(tx *gorm.DB, originalDeliveryID uint, action, reason string, actorUserID uint, actorUsername, actorRole string) (*models.NotificationDeliveryRecovery, error) {
	recovery := &models.NotificationDeliveryRecovery{
		OriginalDeliveryID: originalDeliveryID,
		Action:             action,
		Reason:             strings.TrimSpace(reason),
		ActorUserID:        actorUserID,
		ActorUsername:      actorUsername,
		ActorRole:          actorRole,
		Status:             models.RecoveryStatusPending,
		RequestedAt:        time.Now(),
	}

	var activeCount int64
	if err := tx.Model(&models.NotificationDeliveryRecovery{}).
		Where("original_delivery_id = ? AND status IN ?", originalDeliveryID, []string{models.RecoveryStatusPending, models.RecoveryStatusInProgress}).
		Count(&activeCount).Error; err != nil {
		return nil, fmt.Errorf("check active recoveries: %w", err)
	}
	if activeCount > 0 {
		recovery.Status = models.RecoveryStatusRejected
		recovery.ErrorMessage = "recovery already in progress for this delivery"
		now := time.Now()
		recovery.CompletedAt = &now
		if err := tx.Create(recovery).Error; err != nil {
			return nil, fmt.Errorf("create rejected recovery: %w", err)
		}
		return recovery, nil
	}

	if err := tx.Create(recovery).Error; err != nil {
		return nil, fmt.Errorf("create recovery: %w", err)
	}
	return recovery, nil
}

func recoveryInProgress(recovery *models.NotificationDeliveryRecovery) *models.NotificationDeliveryRecovery {
	recovery.Status = models.RecoveryStatusInProgress
	recovery.ErrorMessage = ""
	recovery.CompletedAt = nil
	return recovery
}

func (s *Service) completeRecovery(tx *gorm.DB, recovery *models.NotificationDeliveryRecovery, resultDeliveryID uint, status, errorMessage string) error {
	recovery.Status = status
	recovery.ErrorMessage = errorMessage
	now := time.Now()
	recovery.CompletedAt = &now
	if resultDeliveryID != 0 {
		recovery.ResultDeliveryID = &resultDeliveryID
	}
	if err := tx.Save(recovery).Error; err != nil {
		return fmt.Errorf("update recovery status: %w", err)
	}
	return nil
}

func (s *Service) loadRetryChannel(ctx context.Context, tx *gorm.DB, channelID uint) (*models.Channel, error) {
	var channel models.Channel
	if err := tx.WithContext(ctx).First(&channel, channelID).Error; err != nil {
		return nil, fmt.Errorf("load retry channel: %w", err)
	}
	if !channel.Enabled {
		return nil, errors.New("retry channel is disabled")
	}
	return &channel, nil
}

func decodeRetryExecution(original *models.NotificationDelivery) (*models.Alert, *models.RouteRule, models.RenderedPayloadSnapshot, error) {
	var alertSnapshot models.AlertSnapshot
	if err := json.Unmarshal(original.AlertSnapshot, &alertSnapshot); err != nil {
		return nil, nil, models.RenderedPayloadSnapshot{}, fmt.Errorf("decode alert snapshot: %w", err)
	}
	var payload models.RenderedPayloadSnapshot
	if err := json.Unmarshal(original.RenderedPayloadSnapshot, &payload); err != nil {
		return nil, nil, models.RenderedPayloadSnapshot{}, fmt.Errorf("decode rendered payload snapshot: %w", err)
	}

	alert := &models.Alert{
		AlertID:     alertSnapshot.AlertID,
		TraceID:     alertSnapshot.TraceID,
		Source:      alertSnapshot.Source,
		AlertName:   alertSnapshot.AlertName,
		Severity:    alertSnapshot.Severity,
		Message:     alertSnapshot.Message,
		Status:      alertSnapshot.Status,
		Fingerprint: alertSnapshot.Fingerprint,
		Labels:      mustMarshalJSON(alertSnapshot.Labels),
	}
	if alertSnapshot.TriggerTime != "" {
		if parsed, err := time.Parse(time.RFC3339, alertSnapshot.TriggerTime); err == nil {
			alert.TriggerTime = parsed
			alert.ReceivedAt = parsed
		}
	}

	var routeRule *models.RouteRule
	if len(original.RouteSnapshot) > 0 && !isNullJSONBytes(original.RouteSnapshot) {
		var routeSnapshot models.RouteSnapshot
		if err := json.Unmarshal(original.RouteSnapshot, &routeSnapshot); err != nil {
			return nil, nil, models.RenderedPayloadSnapshot{}, fmt.Errorf("decode route snapshot: %w", err)
		}
		routeRule = &models.RouteRule{
			ID:         routeSnapshot.ID,
			Name:       routeSnapshot.Name,
			Priority:   routeSnapshot.Priority,
			Enabled:    routeSnapshot.Enabled,
			ChannelIDs: mustMarshalJSON(routeSnapshot.ChannelIDs),
		}
	}

	return alert, routeRule, payload, nil
}

func (s *Service) loadReplayAlert(ctx context.Context, tx *gorm.DB, alertID string) (*models.Alert, error) {
	var alert models.Alert
	if err := tx.WithContext(ctx).First(&alert, "alert_id = ?", alertID).Error; err != nil {
		return nil, fmt.Errorf("load replay alert: %w", err)
	}
	return &alert, nil
}

func (s *Service) loadReplayDataSource(ctx context.Context, tx *gorm.DB, source string) (*models.DataSource, error) {
	var ds models.DataSource
	if err := tx.WithContext(ctx).First(&ds, "name = ?", source).Error; err != nil {
		return nil, fmt.Errorf("load replay datasource: %w", err)
	}
	if !ds.Enabled {
		return nil, errors.New("replay datasource is disabled")
	}
	return &ds, nil
}

func (s *Service) findReplayTarget(ctx context.Context, tx *gorm.DB, alert *models.Alert, preferredChannelID uint) (*models.Channel, *models.RouteRule, error) {
	var rules []models.RouteRule
	if err := tx.WithContext(ctx).Where("enabled = ?", true).Order("priority ASC").Find(&rules).Error; err != nil {
		return nil, nil, fmt.Errorf("load route rules: %w", err)
	}
	// Create a matcher that uses the transaction
	txMatcher := routing.NewMatcher(&gormAdapter{db: tx})
	targets, err := txMatcher.FindMatchedTargets(alert, rules)
	if err != nil {
		return nil, nil, err
	}
	if len(targets) == 0 {
		return nil, nil, errors.New("no live route matched for replay")
	}
	for _, target := range targets {
		if target.Channel.ID == preferredChannelID {
			return &target.Channel, &target.RouteRule, nil
		}
	}
	return &targets[0].Channel, &targets[0].RouteRule, nil
}

func (s *Service) executeRecoveredDelivery(ctx context.Context, tx *gorm.DB, input recoveredExecutionInput) (*models.NotificationDelivery, error) {
	deliveryRecord, err := s.startDeliveryTx(ctx, tx, StartDeliveryInput{
		Alert:         input.alert,
		Channel:       input.channel,
		RouteRule:     input.routeRule,
		DeliveryMode:  input.deliveryMode,
		TriggerKind:   input.triggerKind,
		RenderedTitle: input.title,
		RenderedBody:  input.content,
	})
	if err != nil {
		return nil, err
	}

	for attempt := 1; attempt <= 3; attempt++ {
		startedAt := time.Now()
		sendErr := s.sender()(input.channel, input.title, input.content, nil)
		_, recordErr := s.recordAttemptTx(ctx, tx, deliveryRecord.ID, RecordAttemptInput{
			AttemptNumber: attempt,
			Result:        notificationAttemptResult(sendErr),
			Retryable:     sendErr != nil && notifier.IsRetryableSendError(sendErr),
			ErrorMessage:  notificationErrorMessage(sendErr),
			DurationMS:    time.Since(startedAt).Milliseconds(),
			TriggerKind:   input.triggerKind,
		})
		if recordErr != nil {
			return nil, recordErr
		}

		if sendErr == nil {
			if err := s.markDeliveredTx(ctx, tx, deliveryRecord.ID, MarkDeliveredInput{
				AttemptCount: attempt,
				DeliveredAt:  time.Now(),
			}); err != nil {
				return nil, err
			}
			return deliveryRecord, nil
		}

		if !notifier.IsRetryableSendError(sendErr) {
			if err := s.markFailedTx(ctx, tx, deliveryRecord.ID, MarkFailedInput{
				AttemptCount: attempt,
				FailedAt:     time.Now(),
				Result:       models.AttemptResultFailed,
				Retryable:    false,
				ErrorMessage: notificationErrorMessage(sendErr),
				TriggerKind:  input.triggerKind,
			}); err != nil {
				return nil, err
			}
			return nil, sendErr
		}

		if attempt == 3 {
			if err := s.markFailedTx(ctx, tx, deliveryRecord.ID, MarkFailedInput{
				AttemptCount: attempt,
				FailedAt:     time.Now(),
				Result:       models.AttemptResultFailed,
				Retryable:    true,
				ErrorMessage: fmt.Sprintf("retry budget exhausted: %s", notificationErrorMessage(sendErr)),
				TriggerKind:  input.triggerKind,
			}); err != nil {
				return nil, err
			}
			return nil, sendErr
		}

		s.sleeper()(50 * time.Millisecond)
	}

	return nil, errors.New("recovery execution exhausted without terminal state")
}

func (s *Service) startDeliveryTx(ctx context.Context, tx *gorm.DB, input StartDeliveryInput) (*models.NotificationDelivery, error) {
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
		Labels:      utils.DecodeJSONMap(input.Alert.Labels),
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
	if err := tx.WithContext(ctx).Create(delivery).Error; err != nil {
		return nil, fmt.Errorf("create delivery: %w", err)
	}
	return delivery, nil
}

func (s *Service) recordAttemptTx(ctx context.Context, tx *gorm.DB, deliveryID uint, input RecordAttemptInput) (*models.NotificationDeliveryAttempt, error) {
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
	if err := tx.WithContext(ctx).Create(attempt).Error; err != nil {
		return nil, fmt.Errorf("create delivery attempt: %w", err)
	}
	return attempt, nil
}

func (s *Service) markDeliveredTx(ctx context.Context, tx *gorm.DB, deliveryID uint, input MarkDeliveredInput) error {
	delivery, err := s.getDeliveryByIDTx(ctx, tx, deliveryID)
	if err != nil {
		return fmt.Errorf("load delivery before mark delivered: %w", err)
	}
	deliveredAt := input.DeliveredAt
	if deliveredAt.IsZero() {
		deliveredAt = time.Now()
	}
	delivery.DeliveryStatus = models.DeliveryStatusDelivered
	delivery.AttemptCount = input.AttemptCount
	delivery.LastAttemptAt = &deliveredAt
	delivery.LastSuccessAt = &deliveredAt
	delivery.FinalFailureSummary = datatypes.JSON([]byte("null"))
	if err := tx.WithContext(ctx).Save(delivery).Error; err != nil {
		return fmt.Errorf("mark delivery delivered: %w", err)
	}
	return nil
}

func (s *Service) markFailedTx(ctx context.Context, tx *gorm.DB, deliveryID uint, input MarkFailedInput) error {
	delivery, err := s.getDeliveryByIDTx(ctx, tx, deliveryID)
	if err != nil {
		return fmt.Errorf("load delivery before mark failed: %w", err)
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
	delivery.DeliveryStatus = models.DeliveryStatusFailed
	delivery.AttemptCount = input.AttemptCount
	delivery.LastAttemptAt = &failedAt
	delivery.FinalFailureSummary = finalFailureSummary
	if err := tx.WithContext(ctx).Save(delivery).Error; err != nil {
		return fmt.Errorf("mark delivery failed: %w", err)
	}
	return nil
}

func (s *Service) getDeliveryByIDTx(ctx context.Context, tx *gorm.DB, deliveryID uint) (*models.NotificationDelivery, error) {
	var delivery models.NotificationDelivery
	if err := tx.WithContext(ctx).
		Preload("Attempts", func(db *gorm.DB) *gorm.DB { return db.Order("attempt_number ASC") }).
		First(&delivery, deliveryID).Error; err != nil {
		return nil, fmt.Errorf("get delivery by id: %w", err)
	}
	return &delivery, nil
}

func (s *Service) sender() func(channel *models.Channel, title, content string, data map[string]interface{}) error {
	if s.sendToChannel != nil {
		return s.sendToChannel
	}
	return notifier.SendToChannel
}

func (s *Service) sleeper() func(time.Duration) {
	if s.sleep != nil {
		return s.sleep
	}
	return time.Sleep
}

func notificationAttemptResult(err error) string {
	if err == nil {
		return models.AttemptResultSuccess
	}
	return models.AttemptResultFailed
}

func notificationErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func mustMarshalJSON(value interface{}) datatypes.JSON {
	if value == nil {
		return datatypes.JSON(`{}`)
	}
	encoded, err := json.Marshal(value)
	if err != nil || len(encoded) == 0 {
		return datatypes.JSON(`{}`)
	}
	return datatypes.JSON(encoded)
}

func isNullJSONBytes(raw []byte) bool {
	return string(raw) == "null"
}

func marshalJSON(v interface{}) (datatypes.JSON, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(encoded), nil
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
