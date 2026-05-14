package routing

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/utils"
)

// MatchTarget represents a matched channel with its route rule
type MatchTarget struct {
	Channel   models.Channel
	RouteRule models.RouteRule
}

// Matcher handles alert routing logic
type Matcher struct {
	db         dbLoader
	regexCache map[string]*regexp.Regexp
	now        func() time.Time
}

// dbLoader abstracts database operations for testing
type dbLoader interface {
	Find(dst interface{}, conds ...interface{}) error
}

// NewMatcher creates a new Matcher instance
func NewMatcher(db dbLoader) *Matcher {
	return &Matcher{
		db:         db,
		regexCache: make(map[string]*regexp.Regexp),
		now:        time.Now,
	}
}

// FindMatchedTargets finds matching channels for an alert based on route rules.
// Uses batch loading to avoid N+1 queries by loading all channels in a single query.
func (m *Matcher) FindMatchedTargets(alert *models.Alert, rules []models.RouteRule) ([]MatchTarget, error) {
	for idx := range rules {
		rule := rules[idx]
		if !rule.Enabled {
			continue
		}

		// Check source filter
		var sources []string
		_ = json.Unmarshal(rule.Sources, &sources)
		if len(sources) > 0 && !utils.ContainsString(sources, alert.Source) {
			continue
		}

		// Check severity filter
		var severities []string
		_ = json.Unmarshal(rule.Severities, &severities)
		if len(severities) > 0 && !utils.ContainsString(severities, alert.Severity) {
			continue
		}

		// Check label matchers
		var labelMatchers []models.LabelMatcher
		_ = json.Unmarshal(rule.LabelMatchers, &labelMatchers)
		if len(labelMatchers) > 0 && !m.MatchLabels(alert.Labels, labelMatchers) {
			continue
		}

		// Check time range
		if !m.IsInTimeRange(rule.TimeRanges) {
			continue
		}

		// Rule matches - batch load channels
		var channelIDs []uint
		_ = json.Unmarshal(rule.ChannelIDs, &channelIDs)
		if len(channelIDs) == 0 {
			continue
		}

		channels, err := m.batchLoadChannels(channelIDs)
		if err != nil {
			return nil, err
		}

		var targets []MatchTarget
		for _, channel := range channels {
			if channel.Enabled {
				targets = append(targets, MatchTarget{
					Channel:   channel,
					RouteRule: rule,
				})
			}
		}

		if len(targets) > 0 {
			return targets, nil
		}
	}

	return nil, nil
}

// batchLoadChannels loads multiple channels in a single query using IN clause
func (m *Matcher) batchLoadChannels(channelIDs []uint) ([]models.Channel, error) {
	if len(channelIDs) == 0 {
		return nil, nil
	}

	var channels []models.Channel
	if err := m.db.Find(&channels, "id IN ?", channelIDs); err != nil {
		return nil, err
	}

	return channels, nil
}

// MatchLabels checks if labels match all matchers.
// Returns true if all matchers match, false otherwise.
func (m *Matcher) MatchLabels(labelsJSON []byte, matchers []models.LabelMatcher) bool {
	if len(matchers) == 0 {
		return true
	}

	var labels map[string]string
	if err := json.Unmarshal(labelsJSON, &labels); err != nil {
		return false
	}

	for _, matcher := range matchers {
		value, exists := labels[matcher.Key]
		if !exists {
			return false
		}

		// Empty pattern matches any value
		if matcher.Pattern == "" {
			continue
		}

		// Check if pattern is a regex pattern
		if strings.HasPrefix(matcher.Pattern, "^") || strings.HasSuffix(matcher.Pattern, "$") {
			re, err := m.getCompiledRegex(matcher.Pattern)
			if err != nil {
				return false
			}
			if !re.MatchString(value) {
				return false
			}
		} else {
			// Simple substring match for non-regex patterns
			if !strings.Contains(value, matcher.Pattern) {
				return false
			}
		}
	}

	return true
}

// IsInTimeRange checks if current time is within any of the time ranges.
// Returns true if no time ranges are defined or if current time falls within any range.
func (m *Matcher) IsInTimeRange(timeRangesJSON []byte) bool {
	if len(timeRangesJSON) == 0 || string(timeRangesJSON) == "[]" {
		return true
	}

	var timeRanges []models.TimeRange
	if err := json.Unmarshal(timeRangesJSON, &timeRanges); err != nil || len(timeRanges) == 0 {
		return true
	}

	now := m.now()
	currentMinutes := now.Hour()*60 + now.Minute()

	for _, tr := range timeRanges {
		startMinutes := utils.ParseTimeToMinutes(tr.StartTime)
		endMinutes := utils.ParseTimeToMinutes(tr.EndTime)

		// Cross-day time range (e.g., 22:00 - 06:00)
		if endMinutes < startMinutes {
			if currentMinutes >= startMinutes || currentMinutes <= endMinutes {
				return true
			}
		} else {
			// Same-day time range
			if currentMinutes >= startMinutes && currentMinutes <= endMinutes {
				return true
			}
		}
	}

	return false
}

// getCompiledRegex returns a cached compiled regex pattern.
// Caching improves performance by avoiding repeated regex compilation.
func (m *Matcher) getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	if re, ok := m.regexCache[pattern]; ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	m.regexCache[pattern] = re
	return re, nil
}
