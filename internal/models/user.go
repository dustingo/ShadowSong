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
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"size:256;not null" json:"-"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	Email        string    `gorm:"size:128" json:"email"`
	Role         string    `gorm:"size:32;not null;default:'viewer'" json:"role"` // admin, operator, viewer
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
