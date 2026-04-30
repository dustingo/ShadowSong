package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	_ "unsafe"

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

//go:linkname testRoleCapabilities github.com/game-ops/ai-alert-system/internal/authz.roleCapabilities
var testRoleCapabilities map[string]map[authz.Capability]struct{}

//go:linkname testSupportedRoleSet github.com/game-ops/ai-alert-system/internal/authz.supportedRoleSet
var testSupportedRoleSet map[string]struct{}

func TestDeliveryHandlerListFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryTestDB(t)
	seeded := seedDeliveryRecords(t, db)
	handler := NewDeliveryHandler(db, delivery.NewService(db))

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
	handler := NewDeliveryHandler(db, delivery.NewService(db))

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
	handler := NewDeliveryHandler(db, delivery.NewService(db))

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
	handler := NewDeliveryHandler(db, delivery.NewService(db))

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
	handler := NewDeliveryHandler(db, delivery.NewService(db))
	jwtAuth := auth.NewJWT(&config.SecurityConfig{
		JWTSecret:   "delivery-test-secret",
		TokenExpiry: time.Hour,
	})
	viewer := seedDeliveryUser(t, db, "viewer-user", authz.RoleViewer)
	forbiddenRole := registerDeliveryDeniedRole(t)
	forbiddenUser := seedDeliveryUser(t, db, "forbidden-user", forbiddenRole)

	viewerToken, err := jwtAuth.GenerateToken(viewer.ID, viewer.Username, viewer.Role)
	require.NoError(t, err)
	forbiddenToken, err := jwtAuth.GenerateToken(forbiddenUser.ID, forbiddenUser.Username, forbiddenUser.Role)
	require.NoError(t, err)

	router := gin.New()
	v1 := router.Group("/api/v1")
	deliveryRoutes := v1.Group("/deliveries")
	deliveryRoutes.Use(middleware.JWTAuth(jwtAuth, db))
	{
		deliveryRoutes.GET("", middleware.RequireCapability(authz.CapabilityViewConfig), handler.List)
		deliveryRoutes.GET("/:id", middleware.RequireCapability(authz.CapabilityViewConfig), handler.Get)
		deliveryRoutes.POST("/:id/retry", middleware.RequireCapability(authz.CapabilityProcessAlerts), handler.Retry)
	}

	t.Run("unauthorized without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/deliveries", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("forbidden on real delivery route without view capability", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/deliveries", nil)
		req.Header.Set("Authorization", "Bearer "+forbiddenToken)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

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

func TestDeliveryHandlerRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDeliveryRecoveryAuthTestDB(t)
	handler := NewDeliveryHandler(db, delivery.NewService(db))
	jwtAuth := auth.NewJWT(&config.SecurityConfig{
		JWTSecret:   "delivery-recovery-test-secret",
		TokenExpiry: time.Hour,
	})

	viewer := seedDeliveryUser(t, db, "viewer-recovery", authz.RoleViewer)
	operator := seedDeliveryUser(t, db, "operator-recovery", authz.RoleOperator)
	viewerToken, err := jwtAuth.GenerateToken(viewer.ID, viewer.Username, viewer.Role)
	require.NoError(t, err)
	operatorToken, err := jwtAuth.GenerateToken(operator.ID, operator.Username, operator.Role)
	require.NoError(t, err)

	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(successServer.Close)

	recoverableID := seedRecoveryRetryFixture(t, db, "alert-retry-handler", 51, true, models.DeliveryStatusFailed, successServer.URL)
	unrecoverableID := seedRecoveryRetryFixture(t, db, "alert-reject-handler", 52, true, models.DeliveryStatusDelivered, successServer.URL)
	replayMissingAlertID := seedRecoveryRetryFixture(t, db, "alert-missing-handler", 53, true, models.DeliveryStatusFailed, successServer.URL)

	router := gin.New()
	v1 := router.Group("/api/v1")
	deliveries := v1.Group("/deliveries")
	deliveries.Use(middleware.JWTAuth(jwtAuth, db))
	{
		deliveries.GET("", middleware.RequireCapability(authz.CapabilityViewConfig), handler.List)
		deliveries.GET("/:id", middleware.RequireCapability(authz.CapabilityViewConfig), handler.Get)
		deliveries.POST("/:id/retry", middleware.RequireCapability(authz.CapabilityProcessAlerts), handler.Retry)
		deliveries.POST("/:id/replay", middleware.RequireCapability(authz.CapabilityProcessAlerts), handler.Replay)
	}

	t.Run("401 without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deliveries/%d/retry", recoverableID), bytes.NewBufferString(`{"reason":"retry now"}`))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("viewer can get but cannot retry", func(t *testing.T) {
		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/deliveries", nil)
		getReq.Header.Set("Authorization", "Bearer "+viewerToken)
		getRecorder := httptest.NewRecorder()
		router.ServeHTTP(getRecorder, getReq)
		assert.Equal(t, http.StatusOK, getRecorder.Code)

		postReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deliveries/%d/retry", recoverableID), bytes.NewBufferString(`{"reason":"retry now"}`))
		postReq.Header.Set("Authorization", "Bearer "+viewerToken)
		postReq.Header.Set("Content-Type", "application/json")
		postRecorder := httptest.NewRecorder()
		router.ServeHTTP(postRecorder, postReq)
		assert.Equal(t, http.StatusForbidden, postRecorder.Code)
	})

	t.Run("operator retry succeeds and writes audit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deliveries/%d/retry", recoverableID), bytes.NewBufferString(`{"reason":"manual retry after endpoint recovery"}`))
		req.Header.Set("Authorization", "Bearer "+operatorToken)
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		var response struct {
			RecoveryID         uint   `json:"recovery_id"`
			Action             string `json:"action"`
			Status             string `json:"status"`
			OriginalDeliveryID uint   `json:"original_delivery_id"`
			ResultDeliveryID   *uint  `json:"result_delivery_id"`
			ErrorMessage       string `json:"error_message"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		assert.NotZero(t, response.RecoveryID)
		assert.Equal(t, models.TriggerKindRetry, response.Action)
		assert.Equal(t, models.RecoveryStatusSucceeded, response.Status)
		require.NotNil(t, response.ResultDeliveryID)

		audit := latestAuditLog(t, db)
		assert.Equal(t, "delivery.retry", audit.Action)
		assert.Equal(t, "delivery_recovery", audit.TargetType)
		assert.Contains(t, audit.Detail, fmt.Sprintf("original_delivery_id=%d", recoverableID))
		assert.Contains(t, audit.Detail, fmt.Sprintf("recovery_id=%d", response.RecoveryID))
		assert.Contains(t, audit.Detail, fmt.Sprintf("result_delivery_id=%d", *response.ResultDeliveryID))
	})

	t.Run("empty reason and bad id return 400", func(t *testing.T) {
		emptyReasonReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deliveries/%d/replay", recoverableID), bytes.NewBufferString(`{"reason":" "}`))
		emptyReasonReq.Header.Set("Authorization", "Bearer "+operatorToken)
		emptyReasonReq.Header.Set("Content-Type", "application/json")
		emptyReasonRecorder := httptest.NewRecorder()
		router.ServeHTTP(emptyReasonRecorder, emptyReasonReq)
		assert.Equal(t, http.StatusBadRequest, emptyReasonRecorder.Code)

		badIDReq := httptest.NewRequest(http.MethodPost, "/api/v1/deliveries/not-a-number/replay", bytes.NewBufferString(`{"reason":"retry"}`))
		badIDReq.Header.Set("Authorization", "Bearer "+operatorToken)
		badIDReq.Header.Set("Content-Type", "application/json")
		badIDRecorder := httptest.NewRecorder()
		router.ServeHTTP(badIDRecorder, badIDReq)
		assert.Equal(t, http.StatusBadRequest, badIDRecorder.Code)
	})

	t.Run("controlled failures stay auditable", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deliveries/%d/replay", replayMissingAlertID), bytes.NewBufferString(`{"reason":"rebuild using current route"}`))
		req.Header.Set("Authorization", "Bearer "+operatorToken)
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), `"error"`)
		audit := latestAuditLog(t, db)
		assert.Equal(t, "delivery.replay", audit.Action)
		assert.Contains(t, audit.Detail, fmt.Sprintf("original_delivery_id=%d", replayMissingAlertID))
		assert.Contains(t, audit.Detail, "action=replay")
	})

	t.Run("non recoverable delivery returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/deliveries/%d/retry", unrecoverableID), bytes.NewBufferString(`{"reason":"should fail"}`))
		req.Header.Set("Authorization", "Bearer "+operatorToken)
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "only failed deliveries can be recovered")
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

func newDeliveryRecoveryAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s-recovery-auth?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.Alert{},
		&models.DataSource{},
		&models.Channel{},
		&models.RouteRule{},
		&models.AuditLog{},
		&models.NotificationDelivery{},
		&models.NotificationDeliveryAttempt{},
		&models.NotificationDeliveryRecovery{},
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

func registerDeliveryDeniedRole(t *testing.T) string {
	t.Helper()

	const role = "delivery_denied_test"
	previousCapabilities, hadCapabilities := testRoleCapabilities[role]
	_, wasSupported := testSupportedRoleSet[role]

	testRoleCapabilities[role] = map[authz.Capability]struct{}{}
	testSupportedRoleSet[role] = struct{}{}

	t.Cleanup(func() {
		if hadCapabilities {
			testRoleCapabilities[role] = previousCapabilities
		} else {
			delete(testRoleCapabilities, role)
		}

		if !wasSupported {
			delete(testSupportedRoleSet, role)
		}
	})

	return role
}

func seedRecoveryRetryFixture(t *testing.T, db *gorm.DB, alertID string, channelID uint, channelEnabled bool, status string, targetURL string) uint {
	t.Helper()

	require.NoError(t, db.Create(&models.Channel{
		ID:      channelID,
		Name:    fmt.Sprintf("channel-%d", channelID),
		Type:    "webhook",
		Config:  datatypes.JSON(fmt.Sprintf(`{"url":%q}`, targetURL)),
		Enabled: channelEnabled,
	}).Error)

	record := models.NotificationDelivery{
		AlertID:        alertID,
		TraceID:        fmt.Sprintf("trace-%s", alertID),
		ChannelID:      channelID,
		DeliveryStatus: status,
		DeliveryMode:   models.DeliveryModeRendered,
		AttemptCount:   1,
		AlertSnapshot: mustJSON(t, models.AlertSnapshot{
			AlertID:   alertID,
			TraceID:   fmt.Sprintf("trace-%s", alertID),
			Source:    "prometheus",
			AlertName: "CPUHigh",
			Severity:  "P0",
			Message:   "cpu critical",
			Status:    "firing",
		}),
		ChannelSnapshot: mustJSON(t, models.ChannelSnapshot{
			ID:      channelID,
			Name:    fmt.Sprintf("channel-%d", channelID),
			Type:    "webhook",
			Enabled: channelEnabled,
		}),
		RouteSnapshot: datatypes.JSON([]byte("null")),
		RenderedPayloadSnapshot: mustJSON(t, models.RenderedPayloadSnapshot{
			Title:   "frozen-title",
			Content: "frozen-body",
		}),
		FinalFailureSummary: mustJSON(t, models.FinalFailureSummary{
			Result:       models.AttemptResultFailed,
			Retryable:    true,
			ErrorMessage: "boom",
			AttemptCount: 1,
			TriggerKind:  models.TriggerKindPipeline,
		}),
	}
	require.NoError(t, db.Create(&record).Error)
	require.NoError(t, db.Create(&models.NotificationDeliveryAttempt{
		DeliveryID:    record.ID,
		AttemptNumber: 1,
		Result:        models.AttemptResultFailed,
		Retryable:     true,
		ErrorMessage:  "boom",
		TriggerKind:   models.TriggerKindPipeline,
	}).Error)
	return record.ID
}

func latestAuditLog(t *testing.T, db *gorm.DB) models.AuditLog {
	t.Helper()

	var audit models.AuditLog
	require.NoError(t, db.Order("id DESC").First(&audit).Error)
	return audit
}
