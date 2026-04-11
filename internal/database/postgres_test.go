package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateDefaultAdminUserUsesCanonicalRoleConstant(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))

	createDefaultAdminUser(db)

	var user models.User
	require.NoError(t, db.First(&user).Error)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, authz.RoleAdmin, user.Role)

	source, err := os.ReadFile("postgres.go")
	require.NoError(t, err)
	assert.Contains(t, string(source), "Role:     authz.RoleAdmin")
}
