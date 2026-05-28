package escalation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/game-ops/ai-alert-system/internal/delivery"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/notifier"
	"github.com/game-ops/ai-alert-system/internal/routing"
	"github.com/game-ops/ai-alert-system/internal/template"
	"gorm.io/gorm"
)

// Checker periodically scans for firing alerts that have exceeded their
// escalation timeout and re-sends notifications up to EscalationMaxRepeats.
type Checker struct {
	db            *gorm.DB
	deliverySvc   *delivery.Service
	matcher       *routing.Matcher
	renderer      *template.Renderer
	sendToChannel func(channel *models.Channel, title, content string, data map[string]interface{}) error
	throttle       *notifier.ChannelThrottle
	sleep         func(time.Duration)
}

// NewChecker creates a new escalation Checker.
func NewChecker(db *gorm.DB, deliverySvc *delivery.Service) *Checker {
	return &Checker{
		db:            db,
		deliverySvc:   deliverySvc,
		matcher:       routing.NewMatcher(&gormAdapter{db: db}),
		renderer:      template.NewRenderer(),
		sendToChannel: notifier.SendToChannel,
		throttle:       notifier.NewChannelThrottle(),
		sleep:         time.Sleep,
	}
}

// Run starts the periodic escalation check loop. It blocks until stop is closed.
func (c *Checker) Run(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.checkAndEscalate()
		case <-stop:
			return
		}
	}
}

func (c *Checker) checkAndEscalate() {
	// 1. Load all enabled RouteRules with escalation_enabled = true
	var escalationRules []models.RouteRule
	if err := c.db.Where("enabled = ? AND escalation_enabled = ?", true, true).Find(&escalationRules).Error; err != nil {
		log.Printf("escalation: failed to load rules: %v", err)
		return
	}
	if len(escalationRules) == 0 {
		return
	}

	// 2. Load all rules for matching
	var allRules []models.RouteRule
	if err := c.db.Where("enabled = ?", true).Order("priority ASC").Find(&allRules).Error; err != nil {
		log.Printf("escalation: failed to load all rules: %v", err)
		return
	}

	// 3. Find firing alerts that might need escalation
	var alerts []models.Alert
	if err := c.db.Where("status = ? AND acked_at IS NULL AND last_notified_at IS NOT NULL", "firing").Find(&alerts).Error; err != nil {
		log.Printf("escalation: failed to load alerts: %v", err)
		return
	}

	now := time.Now()
	for i := range alerts {
		alert := &alerts[i]

		// Re-run route matching
		targets, err := c.matcher.FindMatchedTargets(alert, allRules)
		if err != nil || len(targets) == 0 {
			continue
		}

		// Find the matching escalation rule
		var matchedRule *models.RouteRule
		for idx := range escalationRules {
			if escalationRules[idx].ID == targets[0].RouteRule.ID {
				matchedRule = &escalationRules[idx]
				break
			}
		}
		if matchedRule == nil {
			continue
		}

		// Check timeout
		timeout := time.Duration(matchedRule.EscalationTimeout) * time.Minute
		if alert.LastNotifiedAt.Add(timeout).After(now) {
			continue
		}

		// Check max repeats (+1 because NotifyCount includes the initial notification)
		if alert.NotifyCount >= matchedRule.EscalationMaxRepeats+1 {
			continue
		}

		// Escalate: send notification to all matched channels
		for _, target := range targets {
			c.sendEscalationNotification(alert, &target.Channel, &target.RouteRule)
		}

		// Update alert
		alert.LastNotifiedAt = &now
		alert.NotifyCount++
		c.db.Save(alert)
	}
}

func (c *Checker) sendEscalationNotification(alert *models.Alert, channel *models.Channel, routeRule *models.RouteRule) {
	// Load datasource for output template
	var ds models.DataSource
	if err := c.db.Where("name = ?", alert.Source).First(&ds).Error; err != nil {
		log.Printf("escalation: failed to load datasource for alert %s: %v", alert.AlertID, err)
		return
	}

	title, content, err := c.renderer.RenderAlert(string(ds.OutputTemplate), alert, routeRule)
	if err != nil {
		title = fmt.Sprintf("[%s] %s", alert.Severity, alert.AlertName)
		content = alert.Message
	}

	// Create delivery record via delivery service
	ctx := context.Background()
	deliveryRecord, err := c.deliverySvc.StartDelivery(ctx, delivery.StartDeliveryInput{
		Alert:         alert,
		Channel:       channel,
		RouteRule:     routeRule,
		DeliveryMode:  models.DeliveryModeRendered,
		TriggerKind:   models.TriggerKindEscalation,
		RenderedTitle: title,
		RenderedBody:  content,
	})
	if err != nil {
		log.Printf("escalation: failed to start delivery: %v", err)
		return
	}

	data := map[string]interface{}{
		"title":   title,
		"content": content,
	}

	// Send with retry (same pattern as webhook.go sendChannelNotification)
	for attempt := 1; attempt <= 3; attempt++ {
		if !c.throttle.Allow(channel.ID, channel.RateLimit) {
			log.Printf("escalation: throttled channel %d (%s)", channel.ID, channel.Name)
			return
		}
		sendErr := c.sendToChannel(channel, title, content, data)
		result := models.AttemptResultSuccess
		errMsg := ""
		if sendErr != nil {
			result = models.AttemptResultFailed
			errMsg = sendErr.Error()
		}
		retryable := sendErr != nil && notifier.IsRetryableSendError(sendErr)

		_, _ = c.deliverySvc.RecordAttempt(ctx, deliveryRecord.ID, delivery.RecordAttemptInput{
			AttemptNumber: attempt,
			Result:        result,
			Retryable:     retryable,
			ErrorMessage:  errMsg,
			TriggerKind:   models.TriggerKindEscalation,
		})

		if sendErr == nil {
			_ = c.deliverySvc.MarkDelivered(ctx, deliveryRecord.ID, delivery.MarkDeliveredInput{
				AttemptCount: attempt,
				DeliveredAt:  time.Now(),
			})
			return
		}

		if !notifier.IsRetryableSendError(sendErr) || attempt == 3 {
			_ = c.deliverySvc.MarkFailed(ctx, deliveryRecord.ID, delivery.MarkFailedInput{
				AttemptCount: attempt,
				FailedAt:     time.Now(),
				Result:       models.AttemptResultFailed,
				TriggerKind:  models.TriggerKindEscalation,
			})
			return
		}
		c.sleep(50 * time.Millisecond)
	}
}

// gormAdapter wraps *gorm.DB to implement routing.dbLoader interface
type gormAdapter struct {
	db *gorm.DB
}

func (a *gormAdapter) Find(dst interface{}, conds ...interface{}) error {
	return a.db.Find(dst, conds...).Error
}
