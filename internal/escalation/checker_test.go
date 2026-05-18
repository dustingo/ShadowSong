package escalation

import (
	"fmt"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEscalationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Alert{},
		&models.RouteRule{},
		&models.Channel{},
		&models.DataSource{},
		&models.NotificationDelivery{},
		&models.NotificationDeliveryAttempt{},
	))
	return db
}

type escalationSenderCall struct {
	channelID uint
	title     string
	content   string
}

func newTestChecker(db *gorm.DB) (*Checker, *[]escalationSenderCall) {
	var senderCalls []escalationSenderCall
	deliverySvc := delivery.NewService(db)
	checker := NewChecker(db, deliverySvc)
	checker.sendToChannel = func(channel *models.Channel, title, content string, data map[string]interface{}) error {
		senderCalls = append(senderCalls, escalationSenderCall{
			channelID: channel.ID,
			title:     title,
			content:   content,
		})
		return nil
	}
	checker.sleep = func(time.Duration) {}
	return checker, &senderCalls
}

func seedEscalationBaseData(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.Create(&models.DataSource{
		Name:           "prometheus",
		DisplayName:    "Prometheus",
		InputTemplate:  `{"alert_id":"{{.alert_id}}"}`,
		OutputTemplate: `{"title":"[{{.severity}}] {{.alert_name}}","content":"{{.message}}"}`,
		Enabled:        true,
		GroupByLabels:  datatypes.JSON("[]"),
	}).Error)

	require.NoError(t, db.Create(&models.Channel{
		ID:      100,
		Name:    "ops-webhook",
		Type:    "webhook",
		Config:  datatypes.JSON(`{"url":"https://example.com/hook"}`),
		Enabled: true,
	}).Error)
}

func seedEscalationRule(t *testing.T, db *gorm.DB, timeoutMinutes, maxRepeats int) *models.RouteRule {
	t.Helper()
	rule := &models.RouteRule{
		ID:                  200,
		Name:                "escalation-route",
		Priority:            1,
		Sources:             datatypes.JSON(`["prometheus"]`),
		Severities:          datatypes.JSON(`["P0","P1"]`),
		LabelMatchers:       datatypes.JSON(`[]`),
		ChannelIDs:          datatypes.JSON(`[100]`),
		TimeRanges:          datatypes.JSON(`[]`),
		Enabled:             true,
		EscalationEnabled:   true,
		EscalationTimeout:   timeoutMinutes,
		EscalationMaxRepeats: maxRepeats,
	}
	require.NoError(t, db.Create(rule).Error)
	return rule
}

func seedFiringAlert(t *testing.T, db *gorm.DB, alertID string, lastNotifiedAt time.Time, notifyCount int) *models.Alert {
	t.Helper()
	alert := &models.Alert{
		AlertID:       alertID,
		TraceID:       "trace-" + alertID,
		Source:        "prometheus",
		AlertName:     "CPUHigh",
		Severity:      "P0",
		Message:       "cpu critical",
		Status:        "firing",
		Fingerprint:   "fp-" + alertID,
		TriggerTime:   time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
		ReceivedAt:    time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
		LastNotifiedAt: &lastNotifiedAt,
		NotifyCount:   notifyCount,
		Labels:        datatypes.JSON(`{}`),
	}
	require.NoError(t, db.Create(alert).Error)
	return alert
}

func TestAlertWithinTimeoutIsNotEscalated(t *testing.T) {
	db := setupEscalationTestDB(t)
	seedEscalationBaseData(t, db)
	seedEscalationRule(t, db, 30, 3)

	// Alert was notified 10 minutes ago; timeout is 30 minutes
	lastNotified := time.Now().Add(-10 * time.Minute)
	seedFiringAlert(t, db, "alert-within-timeout", lastNotified, 1)

	checker, senderCalls := newTestChecker(db)
	checker.checkAndEscalate()

	assert.Empty(t, *senderCalls, "alert within timeout should not be escalated")
}

func TestAlertPastTimeoutWithRemainingRepeatsIsEscalated(t *testing.T) {
	db := setupEscalationTestDB(t)
	seedEscalationBaseData(t, db)
	seedEscalationRule(t, db, 30, 3)

	// Alert was notified 31 minutes ago; timeout is 30 minutes, notifyCount=1, maxRepeats=3
	lastNotified := time.Now().Add(-31 * time.Minute)
	alert := seedFiringAlert(t, db, "alert-past-timeout", lastNotified, 1)

	checker, senderCalls := newTestChecker(db)
	checker.checkAndEscalate()

	assert.NotEmpty(t, *senderCalls, "alert past timeout with remaining repeats should be escalated")

	// Verify alert was updated
	var updated models.Alert
	require.NoError(t, db.First(&updated, "alert_id = ?", alert.AlertID).Error)
	assert.Equal(t, 2, updated.NotifyCount, "notify_count should be incremented")
	assert.NotNil(t, updated.LastNotifiedAt)
}

func TestAlertAtMaxRepeatsIsNotEscalated(t *testing.T) {
	db := setupEscalationTestDB(t)
	seedEscalationBaseData(t, db)
	seedEscalationRule(t, db, 30, 3)

	// notifyCount=4 (initial + 3 repeats), maxRepeats=3 => no more escalation
	lastNotified := time.Now().Add(-31 * time.Minute)
	seedFiringAlert(t, db, "alert-max-repeats", lastNotified, 4)

	checker, senderCalls := newTestChecker(db)
	checker.checkAndEscalate()

	assert.Empty(t, *senderCalls, "alert at max repeats should not be escalated")
}

func TestAlertWithEscalationDisabledIsNotEscalated(t *testing.T) {
	db := setupEscalationTestDB(t)
	seedEscalationBaseData(t, db)

	// Create a rule with escalation_enabled = false
	rule := &models.RouteRule{
		ID:                  201,
		Name:                "no-escalation-route",
		Priority:            1,
		Sources:             datatypes.JSON(`["prometheus"]`),
		Severities:          datatypes.JSON(`["P0","P1"]`),
		LabelMatchers:       datatypes.JSON(`[]`),
		ChannelIDs:          datatypes.JSON(`[100]`),
		TimeRanges:          datatypes.JSON(`[]`),
		Enabled:             true,
		EscalationEnabled:   false,
		EscalationTimeout:   30,
		EscalationMaxRepeats: 3,
	}
	require.NoError(t, db.Create(rule).Error)

	lastNotified := time.Now().Add(-31 * time.Minute)
	seedFiringAlert(t, db, "alert-no-escalation", lastNotified, 1)

	checker, senderCalls := newTestChecker(db)
	checker.checkAndEscalate()

	assert.Empty(t, *senderCalls, "alert with escalation disabled on route rule should not be escalated")
}

func TestAckedAlertIsNotEscalated(t *testing.T) {
	db := setupEscalationTestDB(t)
	seedEscalationBaseData(t, db)
	seedEscalationRule(t, db, 30, 3)

	// Create an acked alert (acked_at is set)
	lastNotified := time.Now().Add(-31 * time.Minute)
	ackedAt := time.Now()
	alert := &models.Alert{
		AlertID:       "alert-acked",
		TraceID:       "trace-alert-acked",
		Source:        "prometheus",
		AlertName:     "CPUHigh",
		Severity:      "P0",
		Message:       "cpu critical",
		Status:        "firing",
		Fingerprint:   "fp-alert-acked",
		TriggerTime:   time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
		ReceivedAt:    time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
		LastNotifiedAt: &lastNotified,
		NotifyCount:   1,
		AckedAt:       &ackedAt,
		AckedBy:       "admin",
		Labels:        datatypes.JSON(`{}`),
	}
	require.NoError(t, db.Create(alert).Error)

	checker, senderCalls := newTestChecker(db)
	checker.checkAndEscalate()

	assert.Empty(t, *senderCalls, "acked alert should not be escalated")
}
