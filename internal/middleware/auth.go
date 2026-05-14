package middleware

import (
	"errors"
	"net/http"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	AuthorizationHeader = "Authorization"
	UserIDKey           = "user_id"
	UsernameKey         = "username"
	RoleKey             = "role"
)

type Principal struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// JWTAuth creates a middleware that validates JWT tokens
func JWTAuth(jwtAuth *auth.JWT, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := utils.BearerTokenFromHeader(c.GetHeader(AuthorizationHeader))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		user, principal, err := AuthenticateToken(jwtAuth, db, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if user.RequiresPasswordReset() && !isAllowedForcedResetPath(c) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "password reset required"})
			c.Abort()
			return
		}

		SetPrincipal(c, principal)

		c.Next()
	}
}

func AuthenticateToken(jwtAuth *auth.JWT, db *gorm.DB, tokenString string) (*models.User, Principal, error) {
	claims, err := jwtAuth.ValidateToken(tokenString)
	if err != nil {
		return nil, Principal{}, errors.New("invalid or expired token")
	}

	var user models.User
	if err := db.First(&user, claims.UserID).Error; err != nil {
		return nil, Principal{}, errors.New("invalid or expired token")
	}

	if user.IsDisabled() {
		return nil, Principal{}, errors.New("account disabled")
	}

	if !authz.IsSupportedRole(user.Role) {
		return nil, Principal{}, errors.New("invalid or expired token")
	}

	if claims.IssuedAt == nil || !user.IsTokenValid(claims.IssuedAt.Time) {
		return nil, Principal{}, errors.New("session expired")
	}

	principal, err := NewPrincipal(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, Principal{}, errors.New("invalid or expired token")
	}

	return &user, principal, nil
}

func isAllowedForcedResetPath(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	if path == "/api/v1/users/me" && method == http.MethodGet {
		return true
	}
	if path == "/api/v1/users/me/profile" && method == http.MethodPatch {
		return true
	}
	if path == "/api/v1/users/me/password" && method == http.MethodPut {
		return true
	}

	return false
}

func NewPrincipal(userID uint, username, role string) (Principal, error) {
	if !authz.IsSupportedRole(role) {
		return Principal{}, errors.New("invalid role")
	}

	return Principal{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, nil
}

func SetPrincipal(c *gin.Context, principal Principal) {
	c.Set(UserIDKey, principal.UserID)
	c.Set(UsernameKey, principal.Username)
	c.Set(RoleKey, principal.Role)
}

func GetPrincipal(c *gin.Context) (Principal, bool) {
	userIDValue, ok := c.Get(UserIDKey)
	if !ok {
		return Principal{}, false
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		return Principal{}, false
	}

	usernameValue, ok := c.Get(UsernameKey)
	if !ok {
		return Principal{}, false
	}
	username, ok := usernameValue.(string)
	if !ok {
		return Principal{}, false
	}

	roleValue, ok := c.Get(RoleKey)
	if !ok {
		return Principal{}, false
	}
	role, ok := roleValue.(string)
	if !ok {
		return Principal{}, false
	}

	principal, err := NewPrincipal(userID, username, role)
	if err != nil {
		return Principal{}, false
	}

	return principal, true
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) uint {
	if id, exists := c.Get(UserIDKey); exists {
		if userID, ok := id.(uint); ok {
			return userID
		}
	}
	return 0
}

// GetUsername retrieves the username from the context
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(UsernameKey); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetUserRole retrieves the user role from the context
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get(RoleKey); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}
