package models

import (
	"errors"
	"time"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Username           string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash       string     `gorm:"size:256;not null" json:"-"`
	Name               string     `gorm:"size:128;not null" json:"name"`
	Email              string     `gorm:"size:128" json:"email"`
	Role               string     `gorm:"size:32;not null;default:'viewer'" json:"role"` // admin, operator, viewer
	DisabledAt         *time.Time `json:"disabled_at,omitempty"`
	ForcePasswordReset bool       `gorm:"not null;default:false" json:"force_password_reset"`
	TokenInvalidBefore *time.Time `json:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	return u.normalizeAndValidate()
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	return u.normalizeAndValidate()
}

func (u *User) Validate() error {
	u.Role = authz.DefaultRole(u.Role)

	if u.Username == "" {
		return errors.New("username is required")
	}
	if u.Name == "" {
		return errors.New("name is required")
	}
	if !authz.IsSupportedRole(u.Role) {
		return errors.New("invalid role")
	}
	return nil
}

func (u *User) normalizeAndValidate() error {
	u.Role = authz.DefaultRole(u.Role)
	return u.Validate()
}

// SetPassword encrypts and sets the password
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) IsDisabled() bool {
	return u.DisabledAt != nil
}

func (u *User) RequiresPasswordReset() bool {
	return u.ForcePasswordReset
}

func (u *User) IsTokenValid(issuedAt time.Time) bool {
	if u.TokenInvalidBefore == nil {
		return true
	}

	return !issuedAt.Before(*u.TokenInvalidBefore)
}

func (u *User) InvalidateTokens(at time.Time) {
	ts := at.UTC()
	u.TokenInvalidBefore = &ts
}

func (u *User) SetDisabled(disabled bool, at time.Time) {
	if disabled {
		ts := at.UTC()
		u.DisabledAt = &ts
		u.InvalidateTokens(ts)
		return
	}

	u.DisabledAt = nil
	u.InvalidateTokens(at)
}

func (u *User) SetForcePasswordReset(required bool, at time.Time) {
	u.ForcePasswordReset = required
	if required {
		u.InvalidateTokens(at)
	}
}

func (u *User) UpdatePassword(password string, at time.Time) error {
	if err := u.SetPassword(password); err != nil {
		return err
	}

	u.ForcePasswordReset = false
	u.InvalidateTokens(at)
	return nil
}
