package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	RecoveryStatusPending    = "pending"
	RecoveryStatusInProgress = "in_progress"
	RecoveryStatusSucceeded  = "succeeded"
	RecoveryStatusFailed     = "failed"
	RecoveryStatusRejected   = "rejected"
)

var validRecoveryStatuses = map[string]struct{}{
	RecoveryStatusPending:    {},
	RecoveryStatusInProgress: {},
	RecoveryStatusSucceeded:  {},
	RecoveryStatusFailed:     {},
	RecoveryStatusRejected:   {},
}

// NotificationDeliveryRecovery records one manual retry or replay request.
type NotificationDeliveryRecovery struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	OriginalDeliveryID uint       `gorm:"not null;index" json:"original_delivery_id"`
	Action             string     `gorm:"type:varchar(32);not null;index" json:"action"`
	Reason             string     `gorm:"type:text;not null" json:"reason"`
	ActorUserID        uint       `gorm:"not null;index" json:"actor_user_id"`
	ActorUsername      string     `gorm:"size:64;not null;index" json:"actor_username"`
	ActorRole          string     `gorm:"size:32;not null;index" json:"actor_role"`
	Status             string     `gorm:"type:varchar(32);not null;index" json:"status"`
	ResultDeliveryID   *uint      `gorm:"index" json:"result_delivery_id,omitempty"`
	ErrorMessage       string     `gorm:"type:text" json:"error_message"`
	RequestedAt        time.Time  `gorm:"not null;index" json:"requested_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (NotificationDeliveryRecovery) TableName() string {
	return "notification_delivery_recoveries"
}

func (r *NotificationDeliveryRecovery) BeforeCreate(tx *gorm.DB) error {
	if r.Status == "" {
		r.Status = RecoveryStatusPending
	}
	if r.RequestedAt.IsZero() {
		r.RequestedAt = time.Now()
	}
	return r.Validate()
}

func (r *NotificationDeliveryRecovery) BeforeUpdate(tx *gorm.DB) error {
	return r.Validate()
}

func (r *NotificationDeliveryRecovery) Validate() error {
	if r.OriginalDeliveryID == 0 {
		return errors.New("original_delivery_id is required")
	}
	if _, ok := validTriggerKinds[r.Action]; !ok || r.Action == TriggerKindPipeline {
		return errors.New("invalid recovery action")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return errors.New("reason is required")
	}
	if len(strings.TrimSpace(r.Reason)) > 512 {
		return errors.New("reason exceeds 512 characters")
	}
	if r.ActorUserID == 0 {
		return errors.New("actor_user_id is required")
	}
	if strings.TrimSpace(r.ActorUsername) == "" {
		return errors.New("actor_username is required")
	}
	if strings.TrimSpace(r.ActorRole) == "" {
		return errors.New("actor_role is required")
	}
	if _, ok := validRecoveryStatuses[r.Status]; !ok {
		return errors.New("invalid recovery status")
	}
	return nil
}
