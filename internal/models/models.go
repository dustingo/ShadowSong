package models

import (
	"encoding/json"
	"errors"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
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
	if _, err := texttemplate.New("input_template").Parse(d.InputTemplate); err != nil {
		return fmt.Errorf("input_template syntax error: %s", err.Error())
	}
	if d.OutputTemplate == "" {
		return errors.New("output_template is required")
	}
	if _, err := htmltemplate.New("output_template").Parse(d.OutputTemplate); err != nil {
		return fmt.Errorf("output_template syntax error: %s", err.Error())
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
	RateLimit int            `gorm:"default:0" json:"rate_limit"` // max notifications per minute, 0 = unlimited
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
		"email":    true,
	}
	if !validTypes[c.Type] {
		return errors.New("invalid channel type")
	}
	if c.Type == "webhook" {
		var cfg struct {
			Method      string          `json:"method"`
			ContentType string          `json:"content_type"`
			AuthType    string          `json:"auth_type"`
			AuthConfig  json.RawMessage `json:"auth_config"`
		}
		if err := json.Unmarshal(c.Config, &cfg); err == nil {
			validMethods := map[string]bool{"": true, "POST": true, "PUT": true}
			if !validMethods[cfg.Method] {
				return errors.New("invalid webhook method, must be POST or PUT")
			}
			validContentTypes := map[string]bool{"": true, "application/json": true, "application/x-www-form-urlencoded": true}
			if !validContentTypes[cfg.ContentType] {
				return errors.New("invalid webhook content_type")
			}
			validAuthTypes := map[string]bool{"": true, "none": true, "basic": true, "custom": true}
			if !validAuthTypes[cfg.AuthType] {
				return errors.New("invalid webhook auth_type, must be none, basic, or custom")
			}
			if cfg.AuthType == "basic" {
				var bc struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}
				if err := json.Unmarshal(cfg.AuthConfig, &bc); err != nil || bc.Username == "" {
					return errors.New("basic auth requires username in auth_config")
				}
			}
			if cfg.AuthType == "custom" {
				var cc struct {
					HeaderName string `json:"header_name"`
				}
				if err := json.Unmarshal(cfg.AuthConfig, &cc); err != nil || cc.HeaderName == "" {
					return errors.New("custom auth requires header_name in auth_config")
				}
			}
		}
	}
	if c.Type == "email" {
		// email channel config is minimal; from_name is optional
		// SMTP connection and recipients are validated at send time
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
	Recipients    datatypes.JSON `json:"recipients"`     // []string — email addresses for email channels
	TimeRanges    datatypes.JSON `json:"time_ranges"`    // []TimeRange
	Enabled             bool           `gorm:"default:true" json:"enabled"`
	EscalationEnabled   bool           `gorm:"default:false" json:"escalation_enabled"`
	EscalationTimeout   int            `gorm:"default:30" json:"escalation_timeout"`
	EscalationMaxRepeats int           `gorm:"default:3" json:"escalation_max_repeats"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
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
	if r.Recipients == nil {
		r.Recipients = datatypes.JSON("[]")
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

// AuditLog represents a persistent backend-authored audit event.
type AuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ActorUserID   uint      `gorm:"index" json:"actor_user_id"`
	ActorUsername string    `gorm:"size:64;index" json:"actor_username"`
	ActorRole     string    `gorm:"size:32;index" json:"actor_role"`
	Action        string    `gorm:"size:128;index;not null" json:"action"`
	TargetType    string    `gorm:"size:64;index;not null" json:"target_type"`
	TargetID      string    `gorm:"size:128;index;not null" json:"target_id"`
	Result        string    `gorm:"size:32;index;not null" json:"result"`
	Detail        string    `gorm:"type:text" json:"detail"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// SmtpConfig represents global SMTP server configuration
type SmtpConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Host      string    `gorm:"size:128;not null" json:"host"`
	Port      int       `gorm:"not null;default:465" json:"port"`
	Username  string    `gorm:"size:128;not null" json:"username"`
	Password  string    `gorm:"size:256" json:"password"`
	FromAddr  string    `gorm:"size:128;not null" json:"from_addr"`
	FromName  string    `gorm:"size:64" json:"from_name"`
	TLS       bool      `gorm:"default:true" json:"tls"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *SmtpConfig) Validate() error {
	if s.Host == "" {
		return errors.New("host is required")
	}
	if s.Port == 0 {
		return errors.New("port is required")
	}
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.FromAddr == "" {
		return errors.New("from_addr is required")
	}
	return nil
}

