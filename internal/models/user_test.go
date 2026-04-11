package models

import (
	"testing"

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
