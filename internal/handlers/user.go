package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db      *gorm.DB
	jwtAuth *auth.JWT
}

func NewUserHandler(db *gorm.DB, jwtAuth *auth.JWT) *UserHandler {
	return &UserHandler{db: db, jwtAuth: jwtAuth}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token    string       `json:"token"`
	User     *models.User `json:"user"`
	ExpireAt int64        `json:"expire_at"`
}

type createUserRequest struct {
	Username           string `json:"username" binding:"required"`
	Name               string `json:"name" binding:"required"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	Password           string `json:"password" binding:"required"`
	ForcePasswordReset bool   `json:"force_password_reset"`
}

type adminUpdateUserRequest struct {
	Name               *string `json:"name"`
	Email              *string `json:"email"`
	Role               *string `json:"role"`
	Disabled           *bool   `json:"disabled"`
	ForcePasswordReset *bool   `json:"force_password_reset"`
}

type updateOwnProfileRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type updateOwnPasswordRequest struct {
	Password string `json:"password"`
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if user.IsDisabled() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account disabled"})
		return
	}

	if !authz.IsSupportedRole(user.Role) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	token, err := h.jwtAuth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	sanitizeUser(&user)
	c.JSON(http.StatusOK, LoginResponse{
		Token:    token,
		User:     &user,
		ExpireAt: 0,
	})
}

// Logout handles user logout (client-side token removal)
func (h *UserHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// RefreshToken handles token refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	tokenString, err := bearerTokenFromHeader(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.jwtAuth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	user, err := h.loadUserByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.IsDisabled() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account disabled"})
		return
	}

	if !authz.IsSupportedRole(user.Role) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	if claims.IssuedAt == nil || !user.IsTokenValid(claims.IssuedAt.Time) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
		return
	}

	newToken, err := h.jwtAuth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// GetCurrentUser returns the current logged-in user
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}

	sanitizeUser(user)
	c.JSON(http.StatusOK, user)
}

// ListUsers returns all users (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Order("id ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range users {
		sanitizeUser(&users[i])
	}

	c.JSON(http.StatusOK, users)
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user := models.User{
		Username:           req.Username,
		Name:               req.Name,
		Email:              req.Email,
		Role:               authz.DefaultRole(req.Role),
		ForcePasswordReset: req.ForcePasswordReset,
	}

	if !authz.IsSupportedRole(user.Role) {
		_ = recordAudit(h.db, c, "user.create", "user", req.Username, auditResultDenied, "invalid role")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if user.ForcePasswordReset {
		user.SetForcePasswordReset(true, time.Now())
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "user.create", "user", strconv.FormatUint(uint64(user.ID), 10), auditResultAllowed, fmt.Sprintf("role=%s", user.Role))
	sanitizeUser(&user)
	c.JSON(http.StatusCreated, user)
}

// AdminUpdateUser updates another user's profile, role, and account-control state.
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	targetID, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.loadUserByID(targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req adminUpdateUserRequest
	if err := bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	currentUserID := middleware.GetUserID(c)
	if currentUserID == user.ID {
		if req.Disabled != nil && *req.Disabled {
			_ = recordAudit(h.db, c, "user.disable", "user", strconv.FormatUint(uint64(user.ID), 10), auditResultDenied, "cannot disable yourself")
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot disable yourself"})
			return
		}
		if req.Role != nil && *req.Role != user.Role {
			_ = recordAudit(h.db, c, "user.role_change", "user", strconv.FormatUint(uint64(user.ID), 10), auditResultDenied, "cannot change your own role")
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own role"})
			return
		}
	}

	now := time.Now()
	securityChanged := false

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		role := authz.DefaultRole(*req.Role)
		if !authz.IsSupportedRole(role) {
			_ = recordAudit(h.db, c, "user.role_change", "user", strconv.FormatUint(uint64(user.ID), 10), auditResultDenied, "invalid role")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
			return
		}
		if user.Role != role {
			user.Role = role
			securityChanged = true
		}
	}
	if req.Disabled != nil && user.IsDisabled() != *req.Disabled {
		user.SetDisabled(*req.Disabled, now)
		securityChanged = true
	}
	if req.ForcePasswordReset != nil && user.ForcePasswordReset != *req.ForcePasswordReset {
		user.SetForcePasswordReset(*req.ForcePasswordReset, now)
		securityChanged = true
	}
	if securityChanged && !user.IsDisabled() && user.TokenInvalidBefore == nil {
		user.InvalidateTokens(now)
	}
	if securityChanged && req.Disabled == nil && req.ForcePasswordReset == nil && req.Role != nil {
		user.InvalidateTokens(now)
	}

	if err := h.db.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recordAdminUpdateAudit(h.db, c, user, req)
	sanitizeUser(user)
	c.JSON(http.StatusOK, user)
}

// UpdateOwnProfile updates the current user's safe profile fields.
func (h *UserHandler) UpdateOwnProfile(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}

	var req updateOwnProfileRequest
	if err := bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}

	if err := h.db.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sanitizeUser(user)
	c.JSON(http.StatusOK, user)
}

// UpdateOwnPassword updates the current user's password and clears forced-reset state.
func (h *UserHandler) UpdateOwnPassword(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}

	var req updateOwnPasswordRequest
	if err := bindStrictJSON(c, &req); err != nil || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := user.UpdatePassword(req.Password, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.db.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "user.password_change", "user", strconv.FormatUint(uint64(user.ID), 10), auditResultAllowed, "self-service password update")
	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

// DeleteUser deletes a user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := parseUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		_ = recordAudit(h.db, c, "user.delete", "user", strconv.FormatUint(uint64(id), 10), auditResultDenied, "cannot delete yourself")
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	if err := h.db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "user.delete", "user", strconv.FormatUint(uint64(id), 10), auditResultAllowed, "admin deleted user")
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *UserHandler) currentUser(c *gin.Context) (*models.User, bool) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, false
	}

	user, err := h.loadUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return nil, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}

	return user, true
}

func (h *UserHandler) loadUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func parseUserID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func bearerTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid token format")
	}

	return parts[1], nil
}

func bindStrictJSON(c *gin.Context, target any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty body")
	}

	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	if decoder.More() {
		return errors.New("invalid request")
	}

	return nil
}

func sanitizeUser(user *models.User) {
	user.PasswordHash = ""
}

const (
	auditResultAllowed = "allowed"
	auditResultDenied  = "denied"
)

func recordAdminUpdateAudit(db *gorm.DB, c *gin.Context, user *models.User, req adminUpdateUserRequest) {
	targetID := strconv.FormatUint(uint64(user.ID), 10)

	if req.Role != nil {
		_ = recordAudit(db, c, "user.role_change", "user", targetID, auditResultAllowed, fmt.Sprintf("role=%s", user.Role))
	}
	if req.Disabled != nil {
		action := "user.enable"
		if *req.Disabled {
			action = "user.disable"
		}
		_ = recordAudit(db, c, action, "user", targetID, auditResultAllowed, fmt.Sprintf("disabled=%t", *req.Disabled))
	}
	if req.ForcePasswordReset != nil {
		action := "user.clear_force_password_reset"
		if *req.ForcePasswordReset {
			action = "user.force_password_reset"
		}
		_ = recordAudit(db, c, action, "user", targetID, auditResultAllowed, fmt.Sprintf("force_password_reset=%t", *req.ForcePasswordReset))
	}
	if req.Name != nil || req.Email != nil {
		_ = recordAudit(db, c, "user.update_profile", "user", targetID, auditResultAllowed, "admin updated profile fields")
	}
}

func recordAudit(db *gorm.DB, c *gin.Context, action, targetType, targetID, result, detail string) error {
	entry := models.AuditLog{
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Result:     result,
		Detail:     detail,
	}

	if principal, ok := middleware.GetPrincipal(c); ok {
		entry.ActorUserID = principal.UserID
		entry.ActorUsername = principal.Username
		entry.ActorRole = principal.Role
	}

	return db.Create(&entry).Error
}
