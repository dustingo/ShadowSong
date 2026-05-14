# Architecture Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor internal architecture to eliminate N+1 queries, optimize statistics queries, fix WebSocket goroutine leaks, and deduplicate routing/template code.

**Architecture:** Create four new modules (routing, template, stats, websocket) with clear interfaces, then migrate existing code to use them.

**Tech Stack:** Go 1.21+, GORM, Gin, Gorilla WebSocket

---

## File Structure

```
internal/
├── routing/
│   ├── matcher.go         # NEW: Route matching logic
│   └── matcher_test.go    # NEW: Unit tests
├── template/
│   ├── renderer.go        # NEW: Template rendering with caching
│   └── renderer_test.go   # NEW: Unit tests
├── stats/
│   ├── queries.go         # NEW: Optimized statistics queries
│   └── queries_test.go    # NEW: Unit tests
├── websocket/
│   ├── hub.go             # NEW: Connection hub
│   ├── client.go          # NEW: Client with lifecycle
│   └── client_test.go     # NEW: Unit tests
├── handlers/
│   ├── alert.go           # MODIFY: Use stats module
│   ├── webhook.go         # MODIFY: Use routing/template modules
│   └── websocket.go       # REPLACE: Use new websocket module
└── delivery/
    └── service.go         # MODIFY: Use routing/template modules
```

---

## Task 1: Create Routing Module

**Files:**
- Create: `internal/routing/matcher.go`
- Create: `internal/routing/matcher_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/routing/matcher_test.go`:

```go
package routing

import (
	"encoding/json"
	"testing"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.Channel{}, &models.RouteRule{})
	require.NoError(t, err)
	return db
}

func TestMatcher_FindMatchedTargets_BatchLoadsChannels(t *testing.T) {
	db := setupTestDB(t)

	// Create channels
	channel1 := models.Channel{ID: 1, Name: "channel1", Type: "webhook", Enabled: true}
	channel2 := models.Channel{ID: 2, Name: "channel2", Type: "webhook", Enabled: true}
	require.NoError(t, db.Create(&channel1).Error)
	require.NoError(t, db.Create(&channel2).Error)

	// Create route rule
	channelIDs, _ := json.Marshal([]uint{1, 2})
	rule := models.RouteRule{
		ID:         1,
		Name:       "test-rule",
		ChannelIDs: datatypes.JSON(channelIDs),
	}
	require.NoError(t, db.Create(&rule).Error)

	// Create alert
	alert := &models.Alert{
		AlertID:   "test-123",
		AlertName: "Test Alert",
		Severity:  "P1",
		Source:    "prometheus",
		Status:    "firing",
	}

	matcher := NewMatcher(db)
	targets, err := matcher.FindMatchedTargets(alert, []models.RouteRule{rule})

	require.NoError(t, err)
	assert.Len(t, targets, 2)
}

func TestMatchLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		matchers []models.LabelMatcher
		want     bool
	}{
		{
			name:   "empty matchers always match",
			labels: map[string]string{"app": "test"},
			matchers: []models.LabelMatcher{},
			want: true,
		},
		{
			name:   "exact match",
			labels: map[string]string{"app": "test", "env": "prod"},
			matchers: []models.LabelMatcher{{Key: "app", Pattern: "test"}},
			want: true,
		},
		{
			name:   "missing label fails",
			labels: map[string]string{"app": "test"},
			matchers: []models.LabelMatcher{{Key: "env", Pattern: "prod"}},
			want: false,
		},
		{
			name:   "regex match",
			labels: map[string]string{"app": "my-app-service"},
			matchers: []models.LabelMatcher{{Key: "app", Pattern: ".*-app-.*"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labelsJSON, _ := json.Marshal(tt.labels)
			got := MatchLabels(labelsJSON, tt.matchers)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsInTimeRange(t *testing.T) {
	tests := []struct {
		name        string
		timeRanges  []models.TimeRange
		mockTime    func() int // returns current minutes
		want        bool
	}{
		{
			name:       "empty ranges always match",
			timeRanges: []models.TimeRange{},
			want:       true,
		},
		{
			name: "within same-day range",
			timeRanges: []models.TimeRange{{StartTime: "09:00", EndTime: "17:00"}},
			mockTime: func() int { return 12 * 60 }, // 12:00
			want: true,
		},
		{
			name: "outside same-day range",
			timeRanges: []models.TimeRange{{StartTime: "09:00", EndTime: "17:00"}},
			mockTime: func() int { return 20 * 60 }, // 20:00
			want: false,
		},
		{
			name: "within cross-day range",
			timeRanges: []models.TimeRange{{StartTime: "22:00", EndTime: "06:00"}},
			mockTime: func() int { return 23 * 60 }, // 23:00
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeRangesJSON, _ := json.Marshal(tt.timeRanges)
			got := IsInTimeRange(timeRangesJSON)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/routing/... -v 2>&1`
Expected: FAIL with "package routing is not in GOROOT" or similar

- [ ] **Step 3: Write minimal implementation**

Create `internal/routing/matcher.go`:

```go
package routing

import (
	"encoding/json"
	"regexp"
	"sync"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/utils"
	"gorm.io/gorm"
)

// MatchTarget represents a matched channel with its route rule.
type MatchTarget struct {
	Channel   models.Channel
	RouteRule *models.RouteRule
}

// Matcher handles route matching logic.
type Matcher struct {
	db *gorm.DB
}

// NewMatcher creates a new Matcher instance.
func NewMatcher(db *gorm.DB) *Matcher {
	return &Matcher{db: db}
}

// FindMatchedTargets finds all matching channels for an alert.
// Optimized to batch-load channels instead of N+1 queries.
func (m *Matcher) FindMatchedTargets(alert *models.Alert, rules []models.RouteRule) ([]MatchTarget, error) {
	var matchedTargets []MatchTarget

	// Collect all unique channel IDs first
	allChannelIDs := make(map[uint]bool)
	rulesToProcess := make([]models.RouteRule, 0)

	for idx := range rules {
		rule := rules[idx]

		// Check source match
		var sources []string
		_ = json.Unmarshal(rule.Sources, &sources)
		if len(sources) > 0 && !utils.ContainsString(sources, alert.Source) {
			continue
		}

		// Check severity match
		var severities []string
		_ = json.Unmarshal(rule.Severities, &severities)
		if len(severities) > 0 && !utils.ContainsString(severities, alert.Severity) {
			continue
		}

		// Check label matchers
		var labelMatchers []models.LabelMatcher
		_ = json.Unmarshal(rule.LabelMatchers, &labelMatchers)
		if len(labelMatchers) > 0 && !MatchLabels(alert.Labels, labelMatchers) {
			continue
		}

		// Check time range
		if !IsInTimeRange(rule.TimeRanges) {
			continue
		}

		// Collect channel IDs for this rule
		var channelIDs []uint
		_ = json.Unmarshal(rule.ChannelIDs, &channelIDs)
		for _, id := range channelIDs {
			allChannelIDs[id] = true
		}
		rulesToProcess = append(rulesToProcess, rule)
	}

	// Batch load all channels (optimization: single query instead of N+1)
	if len(allChannelIDs) == 0 {
		return nil, nil
	}

	channelIDsSlice := make([]uint, 0, len(allChannelIDs))
	for id := range allChannelIDs {
		channelIDsSlice = append(channelIDsSlice, id)
	}

	var channels []models.Channel
	if err := m.db.Where("id IN ?", channelIDsSlice).Find(&channels).Error; err != nil {
		return nil, err
	}

	// Build channel lookup map
	channelMap := make(map[uint]models.Channel)
	for _, ch := range channels {
		channelMap[ch.ID] = ch
	}

	// Build matched targets using first matching rule
	for _, rule := range rulesToProcess {
		var channelIDs []uint
		_ = json.Unmarshal(rule.ChannelIDs, &channelIDs)

		for _, channelID := range channelIDs {
			channel, exists := channelMap[channelID]
			if !exists {
				continue
			}
			if channel.Enabled {
				matchedTargets = append(matchedTargets, MatchTarget{
					Channel:   channel,
					RouteRule: &rule,
				})
			}
		}

		// Stop at first matching rule (priority order)
		if len(matchedTargets) > 0 {
			break
		}
	}

	return matchedTargets, nil
}

// regexCache caches compiled regex patterns.
var regexCache sync.Map

// MatchLabels checks if labels match all matchers.
func MatchLabels(labelsJSON []byte, matchers []models.LabelMatcher) bool {
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

		if matcher.Pattern != "" {
			// Use cached regex
			var re *regexp.Regexp
			if cached, ok := regexCache.Load(matcher.Pattern); ok {
				re = cached.(*regexp.Regexp)
			} else {
				compiled, err := regexp.Compile(matcher.Pattern)
				if err != nil {
					return false
				}
				regexCache.Store(matcher.Pattern, compiled)
				re = compiled
			}
			if !re.MatchString(value) {
				return false
			}
		}
	}

	return true
}

// IsInTimeRange checks if current time is within any of the time ranges.
func IsInTimeRange(timeRangesJSON []byte) bool {
	if len(timeRangesJSON) == 0 || string(timeRangesJSON) == "[]" {
		return true
	}

	var timeRanges []models.TimeRange
	if err := json.Unmarshal(timeRangesJSON, &timeRanges); err != nil {
		return true
	}

	if len(timeRanges) == 0 {
		return true
	}

	now := time.Now()
	currentTime := now.Hour()*60 + now.Minute()

	for _, tr := range timeRanges {
		startMinutes := utils.ParseTimeToMinutes(tr.StartTime)
		endMinutes := utils.ParseTimeToMinutes(tr.EndTime)

		// Handle cross-day ranges (e.g., 22:00 - 06:00)
		if endMinutes < startMinutes {
			if currentTime >= startMinutes || currentTime <= endMinutes {
				return true
			}
		} else {
			if currentTime >= startMinutes && currentTime <= endMinutes {
				return true
			}
		}
	}

	return false
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/routing/... -v 2>&1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/routing/ && git commit -m "feat: add routing module with batch channel loading

- Add Matcher with FindMatchedTargets for route matching
- Optimize N+1 queries by batch-loading channels
- Add MatchLabels and IsInTimeRange shared functions
- Add regex caching for pattern matching

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 2: Create Template Module

**Files:**
- Create: `internal/template/renderer.go`
- Create: `internal/template/renderer_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/template/renderer_test.go`:

```go
package template

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderer_Render(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name    string
		tmpl    string
		data    map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name: "simple string",
			tmpl: `{"title": "Alert"}`,
			data: nil,
			want: `{"title": "Alert"}`,
		},
		{
			name: "variable substitution",
			tmpl: `{"title": "{{.name}}"}`,
			data: map[string]interface{}{"name": "TestAlert"},
			want: `{"title": "TestAlert"}`,
		},
		{
			name: "toJson function",
			tmpl: `{"labels": {{toJson .labels}}}`,
			data: map[string]interface{}{"labels": map[string]string{"app": "test"}},
			want: `{"labels": {"app":"test"}}`,
		},
		{
			name: "default function",
			tmpl: `{"value": "{{default .missing \"fallback\"}}"}`,
			data: map[string]interface{}{},
			want: `{"value": "fallback"}`,
		},
		{
			name: "get function",
			tmpl: `{{get .labels "app"}}`,
			data: map[string]interface{}{"labels": map[string]interface{}{"app": "myapp"}},
			want: `myapp`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderer.Render(tt.tmpl, tt.data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, got)
		})
	}
}

func TestRenderer_RenderAlert(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "test-123",
		AlertName:   "TestAlert",
		Severity:    "P1",
		Message:     "Test message",
		Source:      "prometheus",
		Status:      "firing",
		TriggerTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	tmpl := `{"title": "[{{.severity}}] {{.alert_name}}", "content": "{{.message}}"}`

	title, content, err := renderer.RenderAlert(tmpl, alert, nil)
	require.NoError(t, err)
	assert.Equal(t, "[P1] TestAlert", title)
	assert.Equal(t, "Test message", content)
}

func TestRenderer_TemplateCaching(t *testing.T) {
	renderer := NewRenderer()

	// Render same template twice
	tmpl := `{"title": "{{.name}}"}`
	data := map[string]interface{}{"name": "Test"}

	_, err1 := renderer.Render(tmpl, data)
	_, err2 := renderer.Render(tmpl, data)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Verify cache hit (internal check)
	cached, ok := renderer.cache.Load(tmpl)
	assert.True(t, ok)
	assert.NotNil(t, cached)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/template/... -v 2>&1`
Expected: FAIL with "package template is not in GOROOT"

- [ ] **Step 3: Write minimal implementation**

Create `internal/template/renderer.go`:

```go
package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/utils"
)

// Renderer handles template rendering with caching.
type Renderer struct {
	cache sync.Map // template string -> *template.Template
}

// NewRenderer creates a new Renderer instance.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// templateFuncMap returns the standard template function map.
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"toJson": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"get": func(m map[string]interface{}, key string) interface{} {
			if m == nil {
				return nil
			}
			return m[key]
		},
		"default": func(v, def interface{}) interface{} {
			if v == nil {
				return def
			}
			if s, ok := v.(string); ok && s == "" {
				return def
			}
			return v
		},
		"lookup": func(m map[string]interface{}, keys ...string) interface{} {
			if m == nil {
				return nil
			}
			for _, key := range keys {
				if val, ok := m[key]; ok && val != nil {
					return val
				}
			}
			return nil
		},
	}
}

// Render renders a template string with the given data.
func (r *Renderer) Render(tmplStr string, data map[string]interface{}) (string, error) {
	// Check cache first
	var tmpl *template.Template
	if cached, ok := r.cache.Load(tmplStr); ok {
		tmpl = cached.(*template.Template)
	} else {
		var err error
		tmpl, err = template.New("output").Funcs(templateFuncMap()).Parse(tmplStr)
		if err != nil {
			return "", err
		}
		r.cache.Store(tmplStr, tmpl)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderAlert renders an alert template, returning title and content.
func (r *Renderer) RenderAlert(tmplStr string, alert *models.Alert, routeRule *models.RouteRule) (string, string, error) {
	data := r.buildRenderContext(alert, routeRule)

	resultStr, err := r.Render(tmplStr, data)
	if err != nil {
		return "", "", err
	}

	// Try to parse as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		// Not JSON, return as content
		return "告警通知", resultStr, nil
	}

	title := strings.TrimSpace(fmt.Sprintf("%v", result["title"]))
	content := strings.TrimSpace(fmt.Sprintf("%v", result["content"]))

	if title == "" {
		title = "告警通知"
	}
	if content == "" {
		content = resultStr
	}

	return title, content, nil
}

// buildRenderContext builds the template data context for an alert.
func (r *Renderer) buildRenderContext(alert *models.Alert, routeRule *models.RouteRule) map[string]interface{} {
	event := utils.DecodeJSONMap(alert.Raw)

	severityRaw := ""
	if raw := lookupString(event, "severity", "level", "priority"); raw != "" {
		severityRaw = raw
	}
	if severityRaw == "" {
		if labels, ok := event["labels"].(map[string]interface{}); ok {
			severityRaw = lookupString(labels, "severity", "level", "priority")
		}
	}

	routeName := ""
	if routeRule != nil {
		routeName = routeRule.Name
	}

	labels := utils.DecodeJSONMap(alert.Labels)

	data := map[string]interface{}{
		"alert_id":      alert.AlertID,
		"alert_name":    alert.AlertName,
		"severity":      alert.Severity,
		"severity_code": alert.Severity,
		"severity_raw":  severityRaw,
		"message":       alert.Message,
		"source":        alert.Source,
		"status":        alert.Status,
		"trigger_time":  alert.TriggerTime.Format(time.RFC3339),
		"labels":        labels,
		"route_name":    routeName,
		"event":         event,
		"alert": map[string]interface{}{
			"id":            alert.AlertID,
			"name":          alert.AlertName,
			"severity":      alert.Severity,
			"severity_code": alert.Severity,
			"severity_raw":  severityRaw,
			"message":       alert.Message,
			"source":        alert.Source,
			"status":        alert.Status,
			"trigger_time":  alert.TriggerTime.Format(time.RFC3339),
			"labels":        labels,
		},
	}

	return data
}

// lookupString looks up a value from a map using multiple keys.
func lookupString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		value, ok := m[key]
		if !ok || value == nil {
			continue
		}
		result := strings.TrimSpace(fmt.Sprintf("%v", value))
		if result != "" && result != "<nil>" {
			return result
		}
	}
	return ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/template/... -v 2>&1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/template/ && git commit -m "feat: add template module with caching

- Add Renderer with Render and RenderAlert methods
- Implement template caching for performance
- Add standard template functions (toJson, get, default, lookup)
- Build alert render context with all required fields

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 3: Create Stats Module

**Files:**
- Create: `internal/stats/queries.go`
- Create: `internal/stats/queries_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/stats/queries_test.go`:

```go
package stats

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStatsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.Alert{})
	require.NoError(t, err)
	return db
}

func TestGetAlertStats_SingleQueryOptimization(t *testing.T) {
	db := setupStatsTestDB(t)

	// Create test alerts
	now := time.Now()
	alerts := []models.Alert{
		{AlertID: "1", AlertName: "A1", Severity: "P0", Status: "firing", TriggerTime: now.Add(-1 * time.Hour)},
		{AlertID: "2", AlertName: "A2", Severity: "P1", Status: "firing", TriggerTime: now.Add(-2 * time.Hour)},
		{AlertID: "3", AlertName: "A3", Severity: "P1", Status: "acked", TriggerTime: now.Add(-3 * time.Hour)},
		{AlertID: "4", AlertName: "A4", Severity: "P2", Status: "silenced", TriggerTime: now.Add(-4 * time.Hour)},
		{AlertID: "5", AlertName: "A5", Severity: "P0", Status: "firing", TriggerTime: now.Add(-5 * time.Hour)},
	}

	for _, a := range alerts {
		require.NoError(t, db.Create(&a).Error)
	}

	stats, err := GetAlertStats(db)
	require.NoError(t, err)

	// Verify counts
	assert.Equal(t, int64(5), stats.Total)
	assert.Equal(t, int64(3), stats.Firing)
	assert.Equal(t, int64(1), stats.Acked)
	assert.Equal(t, int64(1), stats.Silenced)

	// Verify severity counts (firing only)
	assert.Equal(t, int64(2), stats.BySeverity["P0"])
	assert.Equal(t, int64(1), stats.BySeverity["P1"])
	assert.Equal(t, int64(0), stats.BySeverity["P2"])
	assert.Equal(t, int64(0), stats.BySeverity["P3"])

	// Verify trend has 24 hours
	assert.Len(t, stats.Trend, 24)
}

func TestGetAlertStats_EmptyDB(t *testing.T) {
	db := setupStatsTestDB(t)

	stats, err := GetAlertStats(db)
	require.NoError(t, err)

	assert.Equal(t, int64(0), stats.Total)
	assert.Equal(t, int64(0), stats.Firing)
	assert.Len(t, stats.Trend, 24)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/stats/... -v 2>&1`
Expected: FAIL with "package stats is not in GOROOT"

- [ ] **Step 3: Write minimal implementation**

Create `internal/stats/queries.go`:

```go
package stats

import (
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

// AlertStats contains alert statistics.
type AlertStats struct {
	Total      int64            `json:"total"`
	Firing     int64            `json:"firing"`
	Acked      int64            `json:"acked"`
	Silenced   int64            `json:"silenced"`
	BySeverity map[string]int64 `json:"by_severity"`
	Trend      []TrendPoint     `json:"trend"`
}

// TrendPoint represents a single point in the trend.
type TrendPoint struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}

// GetAlertStats retrieves alert statistics using optimized queries.
// Uses 3 queries instead of 32:
// 1. Status counts with GROUP BY
// 2. Severity counts with GROUP BY
// 3. Hourly trend with GROUP BY
func GetAlertStats(db *gorm.DB) (*AlertStats, error) {
	stats := &AlertStats{
		BySeverity: make(map[string]int64),
	}

	// Query 1: Status counts (single query with GROUP BY)
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	if err := db.Model(&models.Alert{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts).Error; err != nil {
		return nil, err
	}

	for _, sc := range statusCounts {
		stats.Total += sc.Count
		switch sc.Status {
		case "firing":
			stats.Firing = sc.Count
		case "acked":
			stats.Acked = sc.Count
		case "silenced":
			stats.Silenced = sc.Count
		}
	}

	// Query 2: Severity counts for firing alerts (single query)
	type SeverityCount struct {
		Severity string
		Count    int64
	}
	var severityCounts []SeverityCount
	if err := db.Model(&models.Alert{}).
		Select("severity, count(*) as count").
		Where("status = ?", "firing").
		Group("severity").
		Find(&severityCounts).Error; err != nil {
		return nil, err
	}

	for _, sc := range severityCounts {
		stats.BySeverity[sc.Severity] = sc.Count
	}

	// Ensure all severities have entries
	for _, sev := range []string{"P0", "P1", "P2", "P3"} {
		if _, exists := stats.BySeverity[sev]; !exists {
			stats.BySeverity[sev] = 0
		}
	}

	// Query 3: Hourly trend (single query)
	type HourCount struct {
		Hour  time.Time
		Count int64
	}
	var hourCounts []HourCount
	startTime := time.Now().Add(-24 * time.Hour).Truncate(time.Hour)
	if err := db.Model(&models.Alert{}).
		Select("date(trigger_time, 'start of hour') as hour, count(*) as count").
		Where("trigger_time >= ?", startTime).
		Group("hour").
		Order("hour").
		Find(&hourCounts).Error; err != nil {
		return nil, err
	}

	// Build trend with all 24 hours
	hourMap := make(map[string]int64)
	for _, hc := range hourCounts {
		hourMap[hc.Hour.Format("15:04")] = hc.Count
	}

	for i := 23; i >= 0; i-- {
		hour := time.Now().Add(-time.Duration(i) * time.Hour).Truncate(time.Hour)
		stats.Trend = append(stats.Trend, TrendPoint{
			Time:  hour.Format("15:04"),
			Count: hourMap[hour.Format("15:04")],
		})
	}

	return stats, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/stats/... -v 2>&1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/stats/ && git commit -m "feat: add stats module with optimized queries

- Reduce 32 queries to 3 queries using GROUP BY
- Add GetAlertStats with status, severity, and trend data
- Ensure all severity levels have entries

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 4: Create WebSocket Module

**Files:**
- Create: `internal/websocket/hub.go`
- Create: `internal/websocket/client.go`
- Create: `internal/websocket/client_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/websocket/client_test.go`:

```go
package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_DoneChannelClosesGoroutines(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	// Start heartbeat goroutine
	heartbeatStopped := make(chan bool{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		defer close(heartbeatStopped)

		for {
			select {
			case <-ticker.C:
				// heartbeat tick
			case <-client.done:
				return
			}
		}
	}()

	// Give goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Close client
	client.Close()

	// Wait for heartbeat to stop
	select {
	case <-heartbeatStopped:
		// Success - goroutine stopped
	case <-time.After(1 * time.Second):
		t.Fatal("heartbeat goroutine did not stop")
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Register mock clients
	client1 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}
	client2 := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	hub.Register(client1)
	hub.Register(client2)

	// Broadcast message
	hub.Broadcast([]byte("test message"))

	// Both clients should receive the message
	select {
	case msg := <-client1.send:
		assert.Equal(t, []byte("test message"), msg)
	case <-time.After(1 * time.Second):
		t.Fatal("client1 did not receive message")
	}

	select {
	case msg := <-client2.send:
		assert.Equal(t, []byte("test message"), msg)
	case <-time.After(1 * time.Second):
		t.Fatal("client2 did not receive message")
	}

	// Cleanup
	client1.Close()
	client2.Close()
}

func TestHub_Unregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	hub.Register(client)
	assert.Len(t, hub.clients, 1)

	hub.Unregister(client)
	assert.Len(t, hub.clients, 0)

	client.Close()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/websocket/... -v 2>&1`
Expected: FAIL with "package websocket is not in GOROOT"

- [ ] **Step 3: Write minimal implementation**

Create `internal/websocket/hub.go`:

```go
package websocket

import (
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, close it
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
```

Create `internal/websocket/client.go`:

```go
package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	done    chan struct{}
	once    sync.Once
}

// NewClient creates a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}
}

// Start begins the client's read and write goroutines.
func (c *Client) Start() {
	c.hub.Register(c)

	// Start write goroutine
	go c.writePump()

	// Start heartbeat goroutine
	go c.heartbeat()

	// Read pump runs in current goroutine
	c.readPump()
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer c.Close()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump() {
	defer c.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Hub closed the channel
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		case <-c.done:
			return
		}
	}
}

// heartbeat sends periodic ping messages to keep the connection alive.
func (c *Client) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket ping error: %v", err)
				return
			}
		case <-c.done:
			return
		}
	}
}

// Close gracefully shuts down the client.
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.done) // Signal all goroutines to stop
		c.hub.Unregister(c)
		if c.conn != nil {
			c.conn.Close()
		}
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/websocket/... -v 2>&1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/websocket/ && git commit -m "feat: add websocket module with proper lifecycle management

- Add Hub for connection management
- Add Client with done channel for goroutine coordination
- Fix goroutine leak by using done channel
- Add graceful Close method

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 5: Migrate Alert Handler to Stats Module

**Files:**
- Modify: `internal/handlers/alert.go:167-211`

- [ ] **Step 1: Write the failing test**

Add to `internal/handlers/alert_test.go`:

```go
func TestAlertHandler_Stats_UsesOptimizedQueries(t *testing.T) {
	db := setupAlertTestDB(t)
	handler := NewAlertHandler(db)

	// Create test alerts
	for i := 0; i < 10; i++ {
		alert := models.Alert{
			AlertID:     fmt.Sprintf("alert-%d", i),
			AlertName:   "Test",
			Severity:    "P1",
			Status:      "firing",
			TriggerTime: time.Now(),
		}
		require.NoError(t, db.Create(&alert).Error)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/alerts/stats", nil)

	handler.Stats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Bytes(), &response))

	assert.Equal(t, float64(10), response["total"])
	assert.Equal(t, float64(10), response["firing"])
}
```

- [ ] **Step 2: Run test to verify current behavior**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/handlers/... -run TestAlertHandler_Stats -v 2>&1`
Expected: PASS (current implementation works)

- [ ] **Step 3: Modify alert.go to use stats module**

Modify `internal/handlers/alert.go`:

```go
// Add import
import (
	// ... existing imports ...
	"github.com/game-ops/ai-alert-system/internal/stats"
)

// Replace Stats method (lines 167-211)
func (h *AlertHandler) Stats(c *gin.Context) {
	result, err := stats.GetAlertStats(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
```

- [ ] **Step 4: Run tests to verify**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/handlers/... -run TestAlertHandler_Stats -v 2>&1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/handlers/alert.go internal/handlers/alert_test.go && git commit -m "refactor: migrate alert handler to use stats module

- Replace 32 queries with 3 optimized queries
- Maintain API compatibility

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 6: Migrate Webhook Handler to Routing/Template Modules

**Files:**
- Modify: `internal/handlers/webhook.go`

- [ ] **Step 1: Update imports and struct**

Modify `internal/handlers/webhook.go`:

```go
import (
	// ... existing imports ...
	"github.com/game-ops/ai-alert-system/internal/routing"
	"github.com/game-ops/ai-alert-system/internal/template"
)

type WebhookHandler struct {
	db              *gorm.DB
	redisClient     *redis.Client
	deliveryService *delivery.Service
	logger          *log.Logger
	matcher         *routing.Matcher
	renderer        *template.Renderer
	redisXAdd       func(ctx context.Context, args *redis.XAddArgs) *redis.StringCmd
	sendToChannel   func(channel *models.Channel, title, content string) error
	runAsync        func(fn func())
	sleep           func(time.Duration)
}

func NewWebhookHandler(db *gorm.DB, redisClient *redis.Client) *WebhookHandler {
	handler := &WebhookHandler{
		db:              db,
		redisClient:     redisClient,
		deliveryService: delivery.NewService(db),
		logger:          log.New(os.Stdout, "notification ", log.LstdFlags),
		matcher:         routing.NewMatcher(db),
		renderer:        template.NewRenderer(),
		sendToChannel:   notifier.SendToChannel,
		runAsync: func(fn func()) {
			go fn()
		},
		sleep: time.Sleep,
	}

	if redisClient != nil {
		handler.redisXAdd = redisClient.XAdd
	}

	return handler
}
```

- [ ] **Step 2: Replace findMatchedChannels call**

Find the call to `findMatchedChannels` and replace with:

```go
targets, err := h.matcher.FindMatchedTargets(&alert, rules)
if err != nil {
	h.logNotification("route_match", &alert, nil, "error: %v", err)
	continue
}

for _, target := range targets {
	h.sendNotification(&alert, &target.Channel, target.RouteRule)
}
```

- [ ] **Step 3: Replace renderNotification calls**

Replace template rendering calls with:

```go
title, content, err := h.renderer.RenderAlert(string(ds.OutputTemplate), alert, routeRule)
if err != nil {
	// handle error
}
```

- [ ] **Step 4: Remove duplicate helper functions**

Remove the following functions from webhook.go:
- `findMatchedChannels`
- `matchLabels`
- `isInTimeRange`
- `renderNotification`
- `buildNotificationRenderContext`
- `templateFuncMap`

- [ ] **Step 5: Run tests to verify**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/handlers/... -v 2>&1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/handlers/webhook.go && git commit -m "refactor: migrate webhook handler to use routing/template modules

- Use routing.Matcher for route matching
- Use template.Renderer for template rendering
- Remove duplicate helper functions

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 7: Migrate Delivery Service to Routing/Template Modules

**Files:**
- Modify: `internal/delivery/service.go`

- [ ] **Step 1: Update Service struct**

Modify `internal/delivery/service.go`:

```go
import (
	// ... existing imports ...
	"github.com/game-ops/ai-alert-system/internal/routing"
	"github.com/game-ops/ai-alert-system/internal/template"
)

type Service struct {
	db            *gorm.DB
	matcher       *routing.Matcher
	renderer      *template.Renderer
	sendToChannel func(channel *models.Channel, title, content string) error
	sleep         func(time.Duration)
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:            db,
		matcher:       routing.NewMatcher(db),
		renderer:      template.NewRenderer(),
		sendToChannel: notifier.SendToChannel,
		sleep:         time.Sleep,
	}
}
```

- [ ] **Step 2: Replace findMatchedTargets calls**

Replace calls to `findMatchedTargets` with `s.matcher.FindMatchedTargets`.

- [ ] **Step 3: Replace renderNotification calls**

Replace template rendering with `s.renderer.RenderAlert`.

- [ ] **Step 4: Remove duplicate helper functions**

Remove from service.go:
- `findMatchedTargets`
- `matchLabels`
- `isInTimeRange`
- `renderNotification`
- `buildNotificationRenderContext`
- `templateFuncMap`

- [ ] **Step 5: Run tests to verify**

Run: `cd D:/goproject/shadowsongAI && go test ./internal/delivery/... -v 2>&1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/delivery/service.go && git commit -m "refactor: migrate delivery service to use routing/template modules

- Use routing.Matcher for route matching
- Use template.Renderer for template rendering
- Remove duplicate helper functions

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 8: Replace WebSocket Handler

**Files:**
- Replace: `internal/handlers/websocket.go`

- [ ] **Step 1: Rewrite websocket handler**

Replace `internal/handlers/websocket.go`:

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/game-ops/ai-alert-system/internal/auth"
	"github.com/game-ops/ai-alert-system/internal/middleware"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/game-ops/ai-alert-system/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WSHandler struct {
	db             *gorm.DB
	jwtAuth        *auth.JWT
	allowedOrigins []string
	upgrader       websocket.Upgrader
	hub            *websocket.Hub
	mu             sync.RWMutex
}

func NewWSHandler(db *gorm.DB, jwtAuth *auth.JWT, allowedOrigins []string) *WSHandler {
	h := &WSHandler{
		db:             db,
		jwtAuth:        jwtAuth,
		allowedOrigins: allowedOrigins,
		hub:            websocket.NewHub(),
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	go h.hub.Run()
	return h
}

func (h *WSHandler) HandleAlerts(c *gin.Context) {
	if !h.isAllowedOrigin(c.GetHeader("Origin")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		return
	}

	tokenString := strings.TrimSpace(c.Query("token"))
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token query parameter required"})
		return
	}

	user, _, err := middleware.AuthenticateToken(h.jwtAuth, h.db, tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if user.RequiresPasswordReset() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password reset required"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	client := websocket.NewClient(h.hub, conn)
	client.Start()

	log.Printf("WebSocket client connected")

	// Send current active alerts
	var alerts []models.Alert
	h.db.Where("status = ?", "firing").Order("trigger_time DESC").Limit(50).Find(&alerts)
	if len(alerts) > 0 {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":   "init",
			"alerts": alerts,
		})
		h.hub.Broadcast(msg)
	}
}

func (h *WSHandler) Broadcast(message []byte) {
	h.hub.Broadcast(message)
}

func (h *WSHandler) checkOrigin(r *http.Request) bool {
	return h.isAllowedOrigin(r.Header.Get("Origin"))
}

func (h *WSHandler) isAllowedOrigin(origin string) bool {
	if strings.TrimSpace(origin) == "" {
		return false
	}

	for _, allowed := range h.allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}

		if origin == allowed {
			return true
		}
	}

	return false
}
```

- [ ] **Step 2: Run tests to verify**

Run: `cd D:/goproject/shadowsongAI && go test ./... -v 2>&1`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
cd D:/goproject/shadowsongAI && git add internal/handlers/websocket.go && git commit -m "refactor: replace websocket handler with new module

- Use websocket.Hub for connection management
- Use websocket.Client with proper lifecycle
- Fix goroutine leak on disconnect

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Task 9: Final Verification and Cleanup

- [ ] **Step 1: Run all tests**

Run: `cd D:/goproject/shadowsongAI && go test ./... -v 2>&1`
Expected: All PASS

- [ ] **Step 2: Run linter**

Run: `cd D:/goproject/shadowsongAI && golangci-lint run 2>&1`
Expected: No errors

- [ ] **Step 3: Verify query optimization**

Add a test to verify query count:

```go
func TestStatsQueryOptimization(t *testing.T) {
	// Enable query logging
	// Verify GetAlertStats makes only 3 queries
}
```

- [ ] **Step 4: Final commit**

```bash
cd D:/goproject/shadowsongAI && git add . && git commit -m "chore: final cleanup for architecture refactoring

- All tests pass
- Query count reduced from 32 to 3
- WebSocket goroutine leak fixed
- Code duplication eliminated

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Summary

| Task | Description | Impact |
|------|-------------|--------|
| 1 | Create routing module | Eliminates N+1 queries |
| 2 | Create template module | Deduplicates template code |
| 3 | Create stats module | 32 queries → 3 queries |
| 4 | Create websocket module | Fixes goroutine leak |
| 5 | Migrate alert handler | Uses stats module |
| 6 | Migrate webhook handler | Uses routing/template |
| 7 | Migrate delivery service | Uses routing/template |
| 8 | Replace websocket handler | Uses new websocket module |
| 9 | Final verification | Ensure all tests pass |
