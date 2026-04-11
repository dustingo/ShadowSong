package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/gin-gonic/gin"
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
func JWTAuth(jwtAuth *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := jwtAuth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		principal, err := NewPrincipal(claims.UserID, claims.Username, claims.Role)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		SetPrincipal(c, principal)

		c.Next()
	}
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
