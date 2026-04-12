package models

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/stretchr/testify/assert"
)

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		wantErr  bool
		wantRole string
	}{
		{
			name: "defaults empty role to viewer",
			user: User{
				Username: "alice",
				Name:     "Alice",
			},
			wantRole: authz.RoleViewer,
		},
		{
			name: "accepts supported role",
			user: User{
				Username: "bob",
				Name:     "Bob",
				Role:     authz.RoleOperator,
			},
			wantRole: authz.RoleOperator,
		},
		{
			name: "rejects unsupported role",
			user: User{
				Username: "charlie",
				Name:     "Charlie",
				Role:     "owner",
			},
			wantErr:  true,
			wantRole: "owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantRole, tt.user.Role)
		})
	}
}

func TestUserHooks(t *testing.T) {
	t.Run("before create defaults empty role", func(t *testing.T) {
		user := &User{
			Username: "alice",
			Name:     "Alice",
		}

		err := user.BeforeCreate(nil)

		assert.NoError(t, err)
		assert.Equal(t, authz.RoleViewer, user.Role)
	})

	t.Run("before update rejects unsupported role", func(t *testing.T) {
		user := &User{
			Username: "alice",
			Name:     "Alice",
			Role:     "owner",
		}

		err := user.BeforeUpdate(nil)

		assert.EqualError(t, err, "invalid role")
	})
}

func TestUserAccountStateHelpers(t *testing.T) {
	now := time.Date(2026, 4, 12, 1, 2, 3, 0, time.UTC)
	before := now.Add(-time.Minute)
	after := now.Add(time.Minute)

	t.Run("defaults to active token-valid state", func(t *testing.T) {
		user := User{}

		assert.False(t, user.IsDisabled())
		assert.False(t, user.RequiresPasswordReset())
		assert.True(t, user.IsTokenValid(before))
	})

	t.Run("set disabled marks account disabled and invalidates older tokens", func(t *testing.T) {
		user := User{}

		user.SetDisabled(true, now)

		if assert.NotNil(t, user.DisabledAt) {
			assert.Equal(t, now, user.DisabledAt.UTC())
		}
		if assert.NotNil(t, user.TokenInvalidBefore) {
			assert.Equal(t, now, user.TokenInvalidBefore.UTC())
		}
		assert.True(t, user.IsDisabled())
		assert.False(t, user.IsTokenValid(before))
		assert.True(t, user.IsTokenValid(after))
	})

	t.Run("force password reset flips flag and invalidates older tokens", func(t *testing.T) {
		user := User{}

		user.SetForcePasswordReset(true, now)

		assert.True(t, user.RequiresPasswordReset())
		if assert.NotNil(t, user.TokenInvalidBefore) {
			assert.Equal(t, now, user.TokenInvalidBefore.UTC())
		}
		assert.False(t, user.IsTokenValid(before))
		assert.True(t, user.IsTokenValid(after))
	})

	t.Run("update password clears forced reset and rotates token cutoff", func(t *testing.T) {
		user := User{ForcePasswordReset: true}

		err := user.UpdatePassword("next-password", now)

		assert.NoError(t, err)
		assert.False(t, user.RequiresPasswordReset())
		assert.NotEmpty(t, user.PasswordHash)
		assert.True(t, user.CheckPassword("next-password"))
		if assert.NotNil(t, user.TokenInvalidBefore) {
			assert.Equal(t, now, user.TokenInvalidBefore.UTC())
		}
		assert.False(t, user.IsTokenValid(before))
	})
}
