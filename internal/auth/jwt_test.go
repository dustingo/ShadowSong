package auth

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTGenerateTokenPreservesClaimKeys(t *testing.T) {
	jwtAuth := NewJWT(&config.SecurityConfig{
		JWTSecret:   "test-secret",
		TokenExpiry: time.Hour,
	})

	token, err := jwtAuth.GenerateToken(42, "alice", authz.RoleOperator)
	require.NoError(t, err)

	claims, err := jwtAuth.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, authz.RoleOperator, claims.Role)
}
