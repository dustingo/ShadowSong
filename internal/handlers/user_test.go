package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateUserRejectsInvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := &UserHandler{db: db}

	body := map[string]string{
		"username": "alice",
		"name":     "Alice",
		"role":     "owner",
		"password": "plain-text-password",
	}

	req := newJSONRequest(t, http.MethodPost, "/users", body)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.CreateUser(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"invalid role"}`, recorder.Body.String())

	var count int64
	assert.NoError(t, db.Model(&models.User{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestCreateUserDefaultsEmptyRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := &UserHandler{db: db}

	body := map[string]string{
		"username": "alice",
		"name":     "Alice",
		"password": "plain-text-password",
	}

	req := newJSONRequest(t, http.MethodPost, "/users", body)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.CreateUser(context)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var saved models.User
	assert.NoError(t, db.First(&saved).Error)
	assert.Equal(t, authz.RoleViewer, saved.Role)
}

func TestUpdateUserRejectsInvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := &UserHandler{db: db}

	user := models.User{
		Username: "alice",
		Name:     "Alice",
		Role:     authz.RoleViewer,
	}
	assert.NoError(t, user.SetPassword("password"))
	assert.NoError(t, db.Create(&user).Error)

	body := map[string]string{
		"role": "owner",
	}

	req := newJSONRequest(t, http.MethodPut, "/users/1", body)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	context.Set(middleware.UserIDKey, uint(999))
	context.Set(middleware.RoleKey, authz.RoleAdmin)

	handler.UpdateUser(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"invalid role"}`, recorder.Body.String())

	var saved models.User
	assert.NoError(t, db.First(&saved, user.ID).Error)
	assert.Equal(t, authz.RoleViewer, saved.Role)
}

func openUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate user model: %v", err)
	}

	return db
}

func newJSONRequest(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	return req
}
