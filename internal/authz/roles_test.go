package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoles(t *testing.T) {
	t.Run("supported roles match canonical contract", func(t *testing.T) {
		assert.Equal(t, []string{RoleAdmin, RoleOperator, RoleViewer}, SupportedRoles())
	})

	t.Run("default role preserves valid values and fills empty input", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
			want  string
		}{
			{name: "empty defaults to viewer", input: "", want: RoleViewer},
			{name: "admin stays admin", input: RoleAdmin, want: RoleAdmin},
			{name: "operator stays operator", input: RoleOperator, want: RoleOperator},
			{name: "viewer stays viewer", input: RoleViewer, want: RoleViewer},
			{name: "unsupported role is not renamed", input: "owner", want: "owner"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, DefaultRole(tt.input))
			})
		}
	})

	t.Run("supported role helper exposes stable reusable path", func(t *testing.T) {
		assert.True(t, IsSupportedRole(RoleAdmin))
		assert.True(t, IsSupportedRole(RoleOperator))
		assert.True(t, IsSupportedRole(RoleViewer))
		assert.False(t, IsSupportedRole("owner"))
	})
}
