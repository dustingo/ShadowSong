package models

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Alert represents a unified alert structure
type Alert struct {
	AlertID     string         `gorm:"primaryKey;type:varchar(64)" json:"alert_id"`
	TraceID     string         `gorm:"index;type:varchar(64)" json:"trace_id"`
	Source      string         `gorm:"index" json:"source"`
	AlertName   string         `gorm:"index" json:"alert_name"`
	Severity    string         `gorm:"index" json:"severity"` // P0/P1/P2/P3
	Message     string         `json:"message"`
	Labels      datatypes.JSON `json:"labels"` // map[string]string
	Fingerprint string         `gorm:"index" json:"fingerprint"`
	TriggerTime time.Time      `gorm:"index" json:"trigger_time"`
	ReceivedAt  time.Time      `gorm:"index" json:"received_at"`
	Status      string         `gorm:"index" json:"status"` // pending/firing/acked/silenced/resolved/deduplicated
	Raw         datatypes.JSON `json:"raw"`

	// 去重/聚合信息
	DeduplicateUntil *time.Time `json:"deduplicate_until"` // 去重截止时间
	LastRepeatAt     *time.Time `json:"last_repeat_at"`    // 最后重复时间

	// 确认信息
	AckedBy    string     `json:"acked_by"`
	AckedAt    *time.Time `json:"acked_at"`
	AckComment string     `json:"ack_comment"`

	// 统计
	TriggerCount int `json:"trigger_count"` // 去重计数

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ValidSeverities defines the allowed severity levels
var ValidSeverities = []string{"P0", "P1", "P2", "P3"}

// ValidStatuses defines the allowed alert statuses
var ValidStatuses = []string{"pending", "firing", "acked", "silenced", "resolved", "deduplicated"}

// BeforeCreate is a GORM hook that runs before creating a record
func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	if a.ReceivedAt.IsZero() {
		a.ReceivedAt = time.Now()
	}
	if a.Status == "" {
		a.Status = "pending"
	}
	if a.TriggerCount == 0 {
		a.TriggerCount = 1
	}
	return a.Validate()
}

// BeforeUpdate is a GORM hook that runs before updating a record
func (a *Alert) BeforeUpdate(tx *gorm.DB) error {
	return a.Validate()
}

// Validate validates the alert fields
func (a *Alert) Validate() error {
	if a.AlertID == "" {
		return errors.New("alert_id is required")
	}
	if a.Source == "" {
		return errors.New("source is required")
	}
	if a.AlertName == "" {
		return errors.New("alert_name is required")
	}
	if a.Message == "" {
		return errors.New("message is required")
	}
	if !a.IsValidSeverity() {
		return errors.New("invalid severity level")
	}
	if !a.IsValidStatus() {
		return errors.New("invalid status")
	}
	return nil
}

// IsValidSeverity checks if the severity is valid
func (a *Alert) IsValidSeverity() bool {
	for _, valid := range ValidSeverities {
		if a.Severity == valid {
			return true
		}
	}
	return false
}

// IsValidStatus checks if the status is valid
func (a *Alert) IsValidStatus() bool {
	for _, valid := range ValidStatuses {
		if a.Status == valid {
			return true
		}
	}
	return false
}

// IsActive returns true if the alert is in an active state
func (a *Alert) IsActive() bool {
	return a.Status == "firing"
}

// CanBeAcked returns true if the alert can be acknowledged
func (a *Alert) CanBeAcked() bool {
	return a.Status == "firing" && a.AckedAt == nil
}

// Ack acknowledges the alert
func (a *Alert) Ack(userID, comment string) error {
	if !a.CanBeAcked() {
		return errors.New("alert cannot be acknowledged")
	}

	now := time.Now()
	a.Status = "acked"
	a.AckedBy = userID
	a.AckedAt = &now
	a.AckComment = comment

	return nil
}
