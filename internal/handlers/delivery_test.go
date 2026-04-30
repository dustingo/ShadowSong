package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/authz"
	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeliveryHandlerListFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryTestDB(t)
	seeded := seedDeliveryRecords(t, db)
	handler := NewDeliveryHandler(delivery.NewService(db))

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"/deliveries?alert_id=%s&trace_id=%s&channel_id=%d&delivery_status=%s&created_from=%s&created_to=%s&limit=1&offset=0",
			seeded.matching.AlertID,
			seeded.matching.TraceID,
			seeded.matching.ChannelID,
			seeded.matching.DeliveryStatus,
			seeded.matching.CreatedAt.Add(-time.Minute).Format(time.RFC3339),
			seeded.matching.CreatedAt.Add(time.Minute).Format(time.RFC3339),
		),
		nil,
	)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		List  []deliveryListItemResponse `json:"list"`
		Total int64                      `json:"total"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.List, 1)
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, seeded.matching.ID, response.List[0].ID)
	assert.Equal(t, seeded.matching.AlertID, response.List[0].AlertID)
	assert.Equal(t, seeded.matching.TraceID, response.List[0].TraceID)
	assert.Equal(t, seeded.matching.ChannelID, response.List[0].ChannelID)
	assert.Equal(t, seeded.matching.DeliveryStatus, response.List[0].DeliveryStatus)
	assert.Len(t, response.List[0].Attempts, 2)
}

func TestDeliveryHandlerListRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryTestDB(t)
	handler := NewDeliveryHandler(delivery.NewService(db))

	testCases := []string{
		"/deliveries?channel_id=oops",
		"/deliveries?limit=0",
		"/deliveries?limit=201",
		"/deliveries?offset=-1",
		"/deliveries?created_from=not-a-time",
	}

	for _, target := range testCases {
		t.Run(target, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, target, nil)
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = req

			handler.List(context)

			assert.Equal(t, http.StatusBadRequest, recorder.Code)
			assert.Contains(t, recorder.Body.String(), `"error"`)
		})
	}
}

func TestDeliveryHandlerGetReturnsDetailShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryTestDB(t)
	seeded := seedDeliveryRecords(t, db)
	handler := NewDeliveryHandler(delivery.NewService(db))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/deliveries/%d", seeded.matching.ID), nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req
	context.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", seeded.matching.ID)}}

	handler.Get(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response deliveryDetailResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, seeded.matching.ID, response.ID)
	assert.Equal(t, seeded.matching.AlertID, response.AlertID)
	assert.Equal(t, seeded.matching.TraceID, response.TraceID)
	assert.Equal(t, seeded.matching.ChannelID, response.ChannelID)
	require.NotNil(t, response.RouteRuleID)
	assert.Equal(t, *seeded.matching.RouteRuleID, *response.RouteRuleID)
	assert.Equal(t, seeded.matching.DeliveryStatus, response.DeliveryStatus)
	assert.Equal(t, seeded.matching.DeliveryMode, response.DeliveryMode)
	assert.Equal(t, seeded.matching.AttemptCount, response.AttemptCount)
	assert.Len(t, response.Attempts, 2)
	assert.Equal(t, models.AttemptResultFailed, response.Attempts[0].Result)
	assert.Equal(t, models.AttemptResultSuccess, response.Attempts[1].Result)
	assert.Equal(t, "ledger alert", response.AlertSnapshot.AlertName)
	assert.Equal(t, "ops-webhook", response.ChannelSnapshot.Name)
	assert.Equal(t, "primary-route", response.RouteSnapshot.Name)
	assert.Equal(t, "title-1", response.RenderedPayloadSnapshot.Title)
	assert.NotNil(t, response.FinalFailureSummary)
	assert.Equal(t, "transient issue", response.FinalFailureSummary.ErrorMessage)
}

func TestDeliveryHandlerGetRejectsMissingRecordAndBadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryTestDB(t)
	handler := NewDeliveryHandler(delivery.NewService(db))

	t.Run("bad id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deliveries/nope", nil)
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = req
		context.Params = gin.Params{{Key: "id", Value: "nope"}}

		handler.Get(context)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), `"error"`)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deliveries/999", nil)
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = req
		context.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.Get(context)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "delivery not found")
	})
}

func TestRouterDeliveriesAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryAuthTestDB(t)
	seeded := seedDeliveryRecords(t, db)
	handler := NewDeliveryHandler(delivery.NewService(db))
	jwtAuth := auth.NewJWT(&config.SecurityConfig{
		JWTSecret:   "delivery-test-secret",
		TokenExpiry: time.Hour,
	})
	viewer := seedDeliveryUser(t, db, "viewer-user", authz.RoleViewer)
	operator := seedDeliveryUser(t, db, "operator-user", authz.RoleOperator)

	viewerToken, err := jwtAuth.GenerateToken(viewer.ID, viewer.Username, viewer.Role)
	require.NoError(t, err)
	operatorToken, err := jwtAuth.GenerateToken(operator.ID, operator.Username, operator.Role)
	require.NoError(t, err)

	router := gin.New()
	v1 := router.Group("/api/v1")
	deliveries := v1.Group("/deliveries")
	deliveries.Use(middleware.JWTAuth(jwtAuth, db), middleware.RequireCapability(authz.CapabilityViewConfig))
	{
		deliveries.GET("", handler.List)
		deliveries.GET("/:id", handler.Get)
	}
	deniedRouter := gin.New()
	deniedGroup := deniedRouter.Group("/api/v1/deliveries")
	deniedGroup.Use(middleware.JWTAuth(jwtAuth, db), middleware.RequireCapability(authz.CapabilityManageUsers))
	{
		deniedGroup.GET("", handler.List)
	}

	t.Run("unauthorized without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/deliveries", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("forbidden on stricter capability chain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/deliveries", nil)
		req.Header.Set("Authorization", "Bearer "+operatorToken)
		recorder := httptest.NewRecorder()

		deniedRouter.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("authorized list and detail", func(t *testing.T) {
		listReq := httptest.NewRequest(http.MethodGet, "/api/v1/deliveries", nil)
		listReq.Header.Set("Authorization", "Bearer "+viewerToken)
		listRecorder := httptest.NewRecorder()

		router.ServeHTTP(listRecorder, listReq)

		assert.Equal(t, http.StatusOK, listRecorder.Code)

		detailReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/deliveries/%d", seeded.matching.ID), nil)
		detailReq.Header.Set("Authorization", "Bearer "+viewerToken)
		detailRecorder := httptest.NewRecorder()

		router.ServeHTTP(detailRecorder, detailReq)

		assert.Equal(t, http.StatusOK, detailRecorder.Code)
	})
}

type seededDeliveries struct {
	matching models.NotificationDelivery
	other    models.NotificationDelivery
}

func newDeliveryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.NotificationDelivery{},
		&models.NotificationDeliveryAttempt{},
	))

	return db
}

func newDeliveryAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s-auth?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.NotificationDelivery{},
		&models.NotificationDeliveryAttempt{},
	))

	return db
}

func seedDeliveryRecords(t *testing.T, db *gorm.DB) seededDeliveries {
	t.Helper()

	routeRuleID := uint(7)
	now := time.Date(2026, 4, 30, 9, 0, 0, 0, time.UTC)
	httpStatus := 503

	matching := models.NotificationDelivery{
		AlertID:         "alert-match",
		TraceID:         "trace-match",
		ChannelID:       11,
		RouteRuleID:     &routeRuleID,
		DeliveryStatus:  models.DeliveryStatusDelivered,
		DeliveryMode:    models.DeliveryModeRendered,
		AttemptCount:    2,
		AlertSnapshot:   mustJSON(t, models.AlertSnapshot{AlertID: "alert-match", TraceID: "trace-match", Source: "prometheus", AlertName: "ledger alert", Severity: "P1", Message: "cpu high", Status: "firing"}),
		ChannelSnapshot: mustJSON(t, models.ChannelSnapshot{ID: 11, Name: "ops-webhook", Type: "webhook", Enabled: true}),
		RouteSnapshot:   mustJSON(t, models.RouteSnapshot{ID: routeRuleID, Name: "primary-route", Priority: 1, Enabled: true, ChannelIDs: []uint{11}}),
		RenderedPayloadSnapshot: mustJSON(t, models.RenderedPayloadSnapshot{
			Title:   "title-1",
			Content: "body-1",
		}),
		FinalFailureSummary: mustJSON(t, models.FinalFailureSummary{
			Result:       models.AttemptResultFailed,
			Retryable:    true,
			ErrorMessage: "transient issue",
			HTTPStatus:   &httpStatus,
			AttemptCount: 2,
			TriggerKind:  models.TriggerKindPipeline,
		}),
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(&matching).Error)
	require.NoError(t, db.Create(&models.NotificationDeliveryAttempt{
		DeliveryID:    matching.ID,
		AttemptNumber: 1,
		Result:        models.AttemptResultFailed,
		Retryable:     true,
		ErrorMessage:  "transient issue",
		HTTPStatus:    &httpStatus,
		DurationMS:    10,
		TriggerKind:   models.TriggerKindPipeline,
		CreatedAt:     now,
	}).Error)
	require.NoError(t, db.Create(&models.NotificationDeliveryAttempt{
		DeliveryID:    matching.ID,
		AttemptNumber: 2,
		Result:        models.AttemptResultSuccess,
		Retryable:     false,
		DurationMS:    20,
		TriggerKind:   models.TriggerKindPipeline,
		CreatedAt:     now.Add(time.Second),
	}).Error)

	other := models.NotificationDelivery{
		AlertID:                 "alert-other",
		TraceID:                 "trace-other",
		ChannelID:               22,
		DeliveryStatus:          models.DeliveryStatusFailed,
		DeliveryMode:            models.DeliveryModeDefault,
		AttemptCount:            1,
		AlertSnapshot:           mustJSON(t, models.AlertSnapshot{AlertID: "alert-other", TraceID: "trace-other", Source: "grafana", AlertName: "other alert", Severity: "P2", Message: "disk high", Status: "firing"}),
		ChannelSnapshot:         mustJSON(t, models.ChannelSnapshot{ID: 22, Name: "ops-feishu", Type: "feishu", Enabled: true}),
		RouteSnapshot:           datatypes.JSON([]byte("null")),
		RenderedPayloadSnapshot: mustJSON(t, models.RenderedPayloadSnapshot{Title: "title-2", Content: "body-2"}),
		FinalFailureSummary:     mustJSON(t, models.FinalFailureSummary{Result: models.AttemptResultFailed, Retryable: false, ErrorMessage: "hard fail", AttemptCount: 1, TriggerKind: models.TriggerKindPipeline}),
		CreatedAt:               now.Add(2 * time.Hour),
		UpdatedAt:               now.Add(2 * time.Hour),
	}
	require.NoError(t, db.Create(&other).Error)

	return seededDeliveries{
		matching: matching,
		other:    other,
	}
}

func mustJSON(t *testing.T, value interface{}) datatypes.JSON {
	t.Helper()

	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return datatypes.JSON(raw)
}

func seedDeliveryUser(t *testing.T, db *gorm.DB, username, role string) *models.User {
	t.Helper()

	user := &models.User{
		Username: username,
		Name:     username,
		Role:     role,
	}
	require.NoError(t, user.SetPassword("plain-text-password"))
	require.NoError(t, db.Create(user).Error)

	return user
}
