package models

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	DeliveryStatusPending   = "pending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"

	DeliveryModeRendered = "rendered"
	DeliveryModeDefault  = "default"

	AttemptResultSuccess = "success"
	AttemptResultFailed  = "failed"

	TriggerKindPipeline    = "pipeline"
	TriggerKindRetry       = "retry"
	TriggerKindReplay      = "replay"
	TriggerKindEscalation  = "escalation"
)

var validDeliveryStatuses = map[string]struct{}{
	DeliveryStatusPending:   {},
	DeliveryStatusDelivered: {},
	DeliveryStatusFailed:    {},
}

var validDeliveryModes = map[string]struct{}{
	DeliveryModeRendered: {},
	DeliveryModeDefault:  {},
}

var validAttemptResults = map[string]struct{}{
	AttemptResultSuccess: {},
	AttemptResultFailed:  {},
}

var validTriggerKinds = map[string]struct{}{
	TriggerKindPipeline:   {},
	TriggerKindRetry:      {},
	TriggerKindReplay:     {},
	TriggerKindEscalation: {},
}

// AlertSnapshot freezes the alert context used by a historical delivery.
type AlertSnapshot struct {
	AlertID     string                 `json:"alert_id"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Source      string                 `json:"source"`
	AlertName   string                 `json:"alert_name"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Status      string                 `json:"status"`
	Fingerprint string                 `json:"fingerprint,omitempty"`
	TriggerTime string                 `json:"trigger_time,omitempty"`
	Labels      map[string]interface{} `json:"labels,omitempty"`
}

// ChannelSnapshot freezes the non-secret channel identity needed for audit/retry.
type ChannelSnapshot struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// RouteSnapshot freezes the route identity without copying full runtime config.
type RouteSnapshot struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Priority   int    `json:"priority"`
	Enabled    bool   `json:"enabled"`
	ChannelIDs []uint `json:"channel_ids,omitempty"`
}

// RenderedPayloadSnapshot freezes the actual outbound message body.
type RenderedPayloadSnapshot struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// FinalFailureSummary keeps the terminal failure explanation for operators.
type FinalFailureSummary struct {
	Result       string `json:"result"`
	Retryable    bool   `json:"retryable"`
	ErrorMessage string `json:"error_message"`
	HTTPStatus   *int   `json:"http_status,omitempty"`
	AttemptCount int    `json:"attempt_count"`
	TriggerKind  string `json:"trigger_kind"`
}

// NotificationDelivery is the source-of-truth record for one alert x channel delivery.
type NotificationDelivery struct {
	ID                      uint                          `gorm:"primaryKey" json:"id"`
	AlertID                 string                        `gorm:"type:varchar(64);not null;index" json:"alert_id"`
	TraceID                 string                        `gorm:"type:varchar(64);not null;index" json:"trace_id"`
	ChannelID               uint                          `gorm:"not null;index" json:"channel_id"`
	RouteRuleID             *uint                         `gorm:"index" json:"route_rule_id,omitempty"`
	DeliveryStatus          string                        `gorm:"type:varchar(32);not null;default:'pending';index" json:"delivery_status"`
	DeliveryMode            string                        `gorm:"type:varchar(32);not null" json:"delivery_mode"`
	AttemptCount            int                           `gorm:"not null;default:0" json:"attempt_count"`
	FinalFailureSummary     datatypes.JSON                `gorm:"type:jsonb" json:"final_failure_summary"`
	AlertSnapshot           datatypes.JSON                `gorm:"type:jsonb;not null" json:"alert_snapshot"`
	ChannelSnapshot         datatypes.JSON                `gorm:"type:jsonb;not null" json:"channel_snapshot"`
	RouteSnapshot           datatypes.JSON                `gorm:"type:jsonb" json:"route_snapshot"`
	RenderedPayloadSnapshot datatypes.JSON                `gorm:"type:jsonb;not null" json:"rendered_payload_snapshot"`
	LastAttemptAt           *time.Time                    `json:"last_attempt_at,omitempty"`
	LastSuccessAt           *time.Time                    `json:"last_success_at,omitempty"`
	CreatedAt               time.Time                     `gorm:"index" json:"created_at"`
	UpdatedAt               time.Time                     `json:"updated_at"`
	Attempts                []NotificationDeliveryAttempt `gorm:"foreignKey:DeliveryID;references:ID" json:"attempts,omitempty"`
}

func (NotificationDelivery) TableName() string {
	return "notification_deliveries"
}

func (d *NotificationDelivery) BeforeCreate(tx *gorm.DB) error {
	if d.DeliveryStatus == "" {
		d.DeliveryStatus = DeliveryStatusPending
	}
	return d.Validate()
}

func (d *NotificationDelivery) BeforeUpdate(tx *gorm.DB) error {
	return d.Validate()
}

func (d *NotificationDelivery) Validate() error {
	if d.AlertID == "" {
		return errors.New("alert_id is required")
	}
	if d.TraceID == "" {
		return errors.New("trace_id is required")
	}
	if d.ChannelID == 0 {
		return errors.New("channel_id is required")
	}
	if d.DeliveryMode == "" {
		return errors.New("delivery_mode is required")
	}
	if _, ok := validDeliveryModes[d.DeliveryMode]; !ok {
		return errors.New("invalid delivery_mode")
	}
	if _, ok := validDeliveryStatuses[d.DeliveryStatus]; !ok {
		return errors.New("invalid delivery_status")
	}
	if len(d.AlertSnapshot) == 0 {
		return errors.New("alert_snapshot is required")
	}
	if len(d.ChannelSnapshot) == 0 {
		return errors.New("channel_snapshot is required")
	}
	if len(d.RenderedPayloadSnapshot) == 0 {
		return errors.New("rendered_payload_snapshot is required")
	}
	return nil
}

// NotificationDeliveryAttempt keeps immutable attempt history for a delivery.
type NotificationDeliveryAttempt struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DeliveryID    uint      `gorm:"not null;uniqueIndex:idx_delivery_attempt_number" json:"delivery_id"`
	AttemptNumber int       `gorm:"not null;uniqueIndex:idx_delivery_attempt_number" json:"attempt_number"`
	Result        string    `gorm:"type:varchar(32);not null" json:"result"`
	Retryable     bool      `gorm:"not null;default:false" json:"retryable"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	HTTPStatus    *int      `json:"http_status,omitempty"`
	DurationMS    int64     `gorm:"not null;default:0" json:"duration_ms"`
	TriggerKind   string    `gorm:"type:varchar(32);not null" json:"trigger_kind"`
	CreatedAt     time.Time `json:"created_at"`
}

func (NotificationDeliveryAttempt) TableName() string {
	return "notification_delivery_attempts"
}

func (a *NotificationDeliveryAttempt) BeforeCreate(tx *gorm.DB) error {
	return a.Validate()
}

func (a *NotificationDeliveryAttempt) BeforeUpdate(tx *gorm.DB) error {
	return errors.New("notification delivery attempts are append-only")
}

func (a *NotificationDeliveryAttempt) Validate() error {
	if a.DeliveryID == 0 {
		return errors.New("delivery_id is required")
	}
	if a.AttemptNumber <= 0 {
		return errors.New("attempt_number must be greater than 0")
	}
	if _, ok := validAttemptResults[a.Result]; !ok {
		return errors.New("invalid result")
	}
	if _, ok := validTriggerKinds[a.TriggerKind]; !ok {
		return errors.New("invalid trigger_kind")
	}
	return nil
}
