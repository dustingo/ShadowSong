package models

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DataSource represents a data source configuration
type DataSource struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	DisplayName string `gorm:"size:128;not null" json:"display_name"`
	APIKey      string `gorm:"size:128" json:"api_key"` // API Key for webhook authentication

	// 去重/聚合配置
	DeduplicateEnabled bool `gorm:"default:true" json:"deduplicate_enabled"` // 是否启用去重
	DeduplicateWindow  int  `gorm:"default:3600" json:"deduplicate_window"`  // 去重窗口时间（秒），默认1小时
	GroupEnabled       bool `gorm:"default:false" json:"group_enabled"`      // 是否启用分组聚合
	GroupWindow        int  `gorm:"default:300" json:"group_window"`         // 分组窗口时间（秒），默认5分钟

	InputTemplate  string         `gorm:"type:text;not null" json:"input_template"`
	OutputTemplate string         `gorm:"type:text;not null" json:"output_template"`
	GroupByLabels  datatypes.JSON `json:"group_by_labels"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	LastTriggerAt  *time.Time     `json:"last_trigger_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (d *DataSource) BeforeCreate(tx *gorm.DB) error {
	if d.GroupByLabels == nil {
		d.GroupByLabels = datatypes.JSON("[]")
	}
	return nil
}

func (d *DataSource) Validate() error {
	if d.Name == "" {
		return errors.New("name is required")
	}
	if d.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if d.InputTemplate == "" {
		return errors.New("input_template is required")
	}
	if d.OutputTemplate == "" {
		return errors.New("output_template is required")
	}
	return nil
}

// Channel represents a notification channel
type Channel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Type      string         `gorm:"size:32;not null" json:"type"` // feishu, dingtalk, wecom, webhook
	Config    datatypes.JSON `gorm:"type:jsonb" json:"config"`
	Enabled   bool           `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (c *Channel) BeforeCreate(tx *gorm.DB) error {
	if c.Config == nil {
		c.Config = datatypes.JSON("{}")
	}
	return nil
}

func (c *Channel) BeforeUpdate(tx *gorm.DB) error {
	// 确保 Config 不为空
	if c.Config == nil {
		c.Config = datatypes.JSON("{}")
	}
	return nil
}

func (c *Channel) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Type == "" {
		return errors.New("type is required")
	}
	validTypes := map[string]bool{
		"feishu":   true,
		"dingtalk": true,
		"wecom":    true,
		"webhook":  true,
	}
	if !validTypes[c.Type] {
		return errors.New("invalid channel type")
	}
	return nil
}

// RouteRule represents a routing rule
type RouteRule struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:64;not null" json:"name"`
	Priority      int            `gorm:"default:0" json:"priority"`
	Severities    datatypes.JSON `json:"severities"`     // []string
	Sources       datatypes.JSON `json:"sources"`        // []string
	LabelMatchers datatypes.JSON `json:"label_matchers"` // []LabelMatcher
	ChannelIDs    datatypes.JSON `json:"channel_ids"`    // []uint
	TimeRanges    datatypes.JSON `json:"time_ranges"`    // []TimeRange
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type LabelMatcher struct {
	Key     string `json:"key"`
	Pattern string `json:"pattern"`
}

type TimeRange struct {
	StartTime string `json:"start_time"` // HH:mm format
	EndTime   string `json:"end_time"`   // HH:mm format
}

func (r *RouteRule) BeforeCreate(tx *gorm.DB) error {
	if r.Severities == nil {
		r.Severities = datatypes.JSON("[]")
	}
	if r.Sources == nil {
		r.Sources = datatypes.JSON("[]")
	}
	if r.LabelMatchers == nil {
		r.LabelMatchers = datatypes.JSON("[]")
	}
	if r.ChannelIDs == nil {
		r.ChannelIDs = datatypes.JSON("[]")
	}
	if r.TimeRanges == nil {
		r.TimeRanges = datatypes.JSON("[]")
	}
	return nil
}

func (r *RouteRule) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// SilenceRule represents a silence rule
type SilenceRule struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:64;not null" json:"name"`
	Comment          string         `gorm:"size:256" json:"comment"`
	Source           string         `gorm:"size:64" json:"source"`
	AlertNamePattern string         `gorm:"size:128" json:"alert_name_pattern"`
	Severities       datatypes.JSON `json:"severities"`
	LabelMatchers    datatypes.JSON `json:"label_matchers"`
	StartsAt         time.Time      `gorm:"index" json:"starts_at"`
	EndsAt           time.Time      `gorm:"index" json:"ends_at"`
	CreatedBy        string         `gorm:"size:64" json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (s *SilenceRule) BeforeCreate(tx *gorm.DB) error {
	if s.Severities == nil {
		s.Severities = datatypes.JSON("[]")
	}
	if s.LabelMatchers == nil {
		s.LabelMatchers = datatypes.JSON("[]")
	}
	return nil
}

func (s *SilenceRule) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.StartsAt.IsZero() {
		return errors.New("starts_at is required")
	}
	if s.EndsAt.IsZero() {
		return errors.New("ends_at is required")
	}
	return nil
}

// OnDuty represents an on-duty schedule
type OnDuty struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:64;not null" json:"user_id"`
	UserName  string    `gorm:"size:64;not null" json:"user_name"`
	ChannelID uint      `gorm:"not null" json:"channel_id"`
	StartTime time.Time `gorm:"index" json:"start_time"`
	EndTime   time.Time `gorm:"index" json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (o *OnDuty) Validate() error {
	if o.UserName == "" {
		return errors.New("user_name is required")
	}
	if o.ChannelID == 0 {
		return errors.New("channel_id is required")
	}
	if o.StartTime.IsZero() {
		return errors.New("start_time is required")
	}
	if o.EndTime.IsZero() {
		return errors.New("end_time is required")
	}
	return nil
}
