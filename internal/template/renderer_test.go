package template

import (
	"sync"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/datatypes"
)

func TestRender_SimpleTemplate(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Hello, World!"

	result, err := renderer.Render(tmplStr, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}
}

func TestRender_VariableSubstitution(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Alert: {{.alert_name}} with severity {{.severity}}"
	data := map[string]interface{}{
		"alert_name": "TestAlert",
		"severity":   "P1",
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Alert: TestAlert with severity P1"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_ToJsonFunction(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Labels: {{toJson .labels}}"
	data := map[string]interface{}{
		"labels": map[string]interface{}{
			"env":     "production",
			"service": "api",
		},
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// html/template escapes JSON characters, so we need to unescape
	// The result will have HTML entities like &#34; instead of "
	if !contains(result, "env") || !contains(result, "production") {
		t.Errorf("toJson should contain label data, got '%s'", result)
	}
}

func TestRender_GetFunction(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Value: {{get .labels \"env\"}}"
	data := map[string]interface{}{
		"labels": map[string]interface{}{
			"env": "production",
		},
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Value: production"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_GetFunctionNilMap(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Value: {{get .labels \"env\"}}"
	data := map[string]interface{}{
		"labels": nil,
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Value: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_DefaultFunction(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Status: {{default .status \"unknown\"}}"
	data := map[string]interface{}{
		"status": nil,
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Status: unknown"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_DefaultFunctionWithValue(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Status: {{default .status \"unknown\"}}"
	data := map[string]interface{}{
		"status": "firing",
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Status: firing"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_DefaultFunctionWithEmptyString(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Status: {{default .status \"unknown\"}}"
	data := map[string]interface{}{
		"status": "",
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Status: unknown"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_LookupFunction(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Severity: {{lookup .event \"severity\" \"level\" \"priority\"}}"
	data := map[string]interface{}{
		"event": map[string]interface{}{
			"level": "high",
		},
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Severity: high"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_LookupFunctionNilMap(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Severity: {{lookup .event \"severity\" \"level\"}}"
	data := map[string]interface{}{
		"event": nil,
	}

	result, err := renderer.Render(tmplStr, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	expected := "Severity: "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRender_InvalidTemplate(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Alert: {{.alert_name" // Missing closing brace

	_, err := renderer.Render(tmplStr, map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for invalid template, got nil")
	}
}

func TestRenderAlert_BasicAlert(t *testing.T) {
	renderer := NewRenderer()
	triggerTime := time.Now()

	alert := &models.Alert{
		AlertID:     "alert-123",
		AlertName:   "HighMemoryUsage",
		Severity:    "P1",
		Message:     "Memory usage exceeded 90%",
		Source:      "prometheus",
		Status:      "firing",
		TriggerTime: triggerTime,
		Labels:      datatypes.JSON(`{"env":"production","service":"api"}`),
	}

	routeRule := &models.RouteRule{
		Name: "critical-alerts",
	}

	tmplStr := `{"title": "[{{.severity}}] {{.alert_name}}", "content": "{{.message}}\n\nLabels: {{toJson .labels}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, routeRule)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if title != "[P1] HighMemoryUsage" {
		t.Errorf("Expected title '[P1] HighMemoryUsage', got '%s'", title)
	}

	if content == "" {
		t.Error("Expected non-empty content")
	}

	// Verify content contains expected parts
	if !contains(content, "Memory usage exceeded 90%") {
		t.Errorf("Content should contain message, got '%s'", content)
	}
}

func TestRenderAlert_AllFields(t *testing.T) {
	renderer := NewRenderer()
	triggerTime := time.Now().Truncate(time.Second)

	alert := &models.Alert{
		AlertID:     "alert-456",
		AlertName:   "DiskSpaceLow",
		Severity:    "P2",
		Message:     "Disk space below 10%",
		Source:      "node-exporter",
		Status:      "firing",
		TriggerTime: triggerTime,
		Labels:      datatypes.JSON(`{"host":"server-01","mount":"/data"}`),
	}

	routeRule := &models.RouteRule{
		Name: "infrastructure-alerts",
	}

	// Template using all available fields
	tmplStr := `{"title": "[{{.severity}}] {{.alert_name}} from {{.source}}", "content": "Alert ID: {{.alert_id}}\nStatus: {{.status}}\nTime: {{.trigger_time}}\nRoute: {{.route_name}}\nLabels: {{toJson .labels}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, routeRule)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	// Verify title
	if title != "[P2] DiskSpaceLow from node-exporter" {
		t.Errorf("Expected title '[P2] DiskSpaceLow from node-exporter', got '%s'", title)
	}

	// Verify content contains all fields
	if !contains(content, "alert-456") {
		t.Error("Content should contain alert_id")
	}
	if !contains(content, "firing") {
		t.Error("Content should contain status")
	}
	// Check trigger_time is present (format may have timezone suffix)
	triggerTimeStr := triggerTime.Format(time.RFC3339)
	if !contains(content, triggerTimeStr[:19]) { // Check just the date-time part without timezone
		t.Errorf("Content should contain trigger_time, expected '%s' in '%s'", triggerTimeStr, content)
	}
	if !contains(content, "infrastructure-alerts") {
		t.Error("Content should contain route_name")
	}
}

func TestRenderAlert_NilRouteRule(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-789",
		AlertName:   "TestAlert",
		Severity:    "P3",
		Message:     "Test message",
		Source:      "test",
		Status:      "pending",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	tmplStr := `{"title": "{{.alert_name}}", "content": "{{.message}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if title != "TestAlert" {
		t.Errorf("Expected title 'TestAlert', got '%s'", title)
	}
	if content != "Test message" {
		t.Errorf("Expected content 'Test message', got '%s'", content)
	}
}

func TestRenderAlert_NonJSONTemplate(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-plain",
		AlertName:   "PlainAlert",
		Severity:    "P1",
		Message:     "Plain text message",
		Source:      "plain",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	// Non-JSON template (plain text)
	tmplStr := "Alert: {{.alert_name}} - {{.message}}"

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	// For non-JSON templates, title should be default
	if title != "告警通知" {
		t.Errorf("Expected default title '告警通知', got '%s'", title)
	}
	if content != "Alert: PlainAlert - Plain text message" {
		t.Errorf("Expected content 'Alert: PlainAlert - Plain text message', got '%s'", content)
	}
}

func TestRenderAlert_EventField(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-event",
		AlertName:   "EventAlert",
		Severity:    "P1",
		Message:     "Event-based alert",
		Source:      "custom",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
		Raw:         datatypes.JSON(`{"original_severity":"critical","custom_field":"value"}`),
	}

	tmplStr := `{"title": "{{.alert_name}}", "content": "Original: {{toJson .event}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if title != "EventAlert" {
		t.Errorf("Expected title 'EventAlert', got '%s'", title)
	}

	// Verify event field contains raw data
	if !contains(content, "original_severity") {
		t.Error("Content should contain event data")
	}
}

func TestRenderAlert_AlertNestedField(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-nested",
		AlertName:   "NestedAlert",
		Severity:    "P2",
		Message:     "Nested field test",
		Source:      "test",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{"team":"sre"}`),
	}

	// Template using nested .alert field
	tmplStr := `{"title": "{{.alert.name}}", "content": "ID: {{.alert.id}}\nSeverity: {{.alert.severity}}\nLabels: {{toJson .alert.labels}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if title != "NestedAlert" {
		t.Errorf("Expected title 'NestedAlert', got '%s'", title)
	}
	if !contains(content, "alert-nested") {
		t.Error("Content should contain alert.id")
	}
	if !contains(content, "P2") {
		t.Error("Content should contain alert.severity")
	}
	if !contains(content, "team") {
		t.Error("Content should contain alert.labels")
	}
}

func TestRenderer_TemplateCaching(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Cached template: {{.value}}"

	// First render - should parse and cache
	data1 := map[string]interface{}{"value": "first"}
	result1, err := renderer.Render(tmplStr, data1)
	if err != nil {
		t.Fatalf("First render failed: %v", err)
	}

	// Second render - should use cached template
	data2 := map[string]interface{}{"value": "second"}
	result2, err := renderer.Render(tmplStr, data2)
	if err != nil {
		t.Fatalf("Second render failed: %v", err)
	}

	if result1 != "Cached template: first" {
		t.Errorf("First render: expected 'Cached template: first', got '%s'", result1)
	}
	if result2 != "Cached template: second" {
		t.Errorf("Second render: expected 'Cached template: second', got '%s'", result2)
	}

	// Verify cache has entry
	count := 0
	renderer.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	if count == 0 {
		t.Error("Expected template to be cached")
	}
}

func TestRenderer_ConcurrentCaching(t *testing.T) {
	renderer := NewRenderer()
	tmplStr := "Concurrent: {{.value}}"

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Run 100 concurrent renders with the same template
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			data := map[string]interface{}{"value": i}
			_, err := renderer.Render(tmplStr, data)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent render failed: %v", err)
	default:
	}

	// Verify only one template is cached (same template string)
	count := 0
	renderer.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	if count != 1 {
		t.Errorf("Expected exactly 1 cached template, got %d", count)
	}
}

func TestRenderAlert_InvalidTemplate(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-invalid",
		AlertName:   "InvalidTemplate",
		Severity:    "P1",
		Message:     "Test",
		Source:      "test",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	tmplStr := "{{.alert_name" // Missing closing brace

	_, _, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err == nil {
		t.Error("Expected error for invalid template, got nil")
	}
}

func TestRenderAlert_EmptyTitleInJSON(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-empty-title",
		AlertName:   "EmptyTitle",
		Severity:    "P1",
		Message:     "Message only",
		Source:      "test",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	// JSON with empty title
	tmplStr := `{"title": "", "content": "{{.message}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	// Empty title should use default
	if title != "告警通知" {
		t.Errorf("Expected default title for empty title, got '%s'", title)
	}
	if content != "Message only" {
		t.Errorf("Expected content 'Message only', got '%s'", content)
	}
}

func TestRenderAlert_EmptyContentInJSON(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "alert-empty-content",
		AlertName:   "EmptyContent",
		Severity:    "P1",
		Message:     "Title only",
		Source:      "test",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	// JSON with empty content
	tmplStr := `{"title": "{{.alert_name}}", "content": ""}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if title != "EmptyContent" {
		t.Errorf("Expected title 'EmptyContent', got '%s'", title)
	}
	// Empty content should use the whole result string
	if content == "" {
		t.Error("Expected non-empty content as fallback")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRenderAlert_ResolvedStatusProducesResolvedPrefix(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "resolved-1",
		AlertName:   "ECS监控策略",
		Severity:    "CRITICAL",
		Message:     "实例状态改变",
		Source:      "aliyun",
		Status:      "resolved",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	tmplStr := `{"title": "{{if eq .status "resolved"}}[RESOLVED] {{end}}[{{.severity}}] {{.alert_name}}", "content": "{{.message}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if !contains(title, "[RESOLVED]") {
		t.Errorf("Resolved alert title should contain [RESOLVED], got '%s'", title)
	}
	if !contains(title, "[CRITICAL]") {
		t.Errorf("Resolved alert title should contain severity, got '%s'", title)
	}
	if !contains(content, "实例状态改变") {
		t.Errorf("Content should contain message, got '%s'", content)
	}
}

func TestRenderAlert_FiringStatusDoesNotProduceResolvedPrefix(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "firing-1",
		AlertName:   "ECS监控策略",
		Severity:    "CRITICAL",
		Message:     "实例状态改变",
		Source:      "aliyun",
		Status:      "firing",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	tmplStr := `{"title": "{{if eq .status "resolved"}}[RESOLVED] {{end}}[{{.severity}}] {{.alert_name}}", "content": "{{.message}}"}`

	title, content, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if contains(title, "[RESOLVED]") {
		t.Errorf("Firing alert title should NOT contain [RESOLVED], got '%s'", title)
	}
	if !contains(title, "[CRITICAL]") {
		t.Errorf("Firing alert title should contain severity, got '%s'", title)
	}
	if !contains(content, "实例状态改变") {
		t.Errorf("Content should contain message, got '%s'", content)
	}
}

func TestRenderAlert_TemplateWithoutResolvedCondition(t *testing.T) {
	renderer := NewRenderer()

	alert := &models.Alert{
		AlertID:     "resolved-2",
		AlertName:   "ECS监控策略",
		Severity:    "CRITICAL",
		Message:     "实例状态改变",
		Source:      "aliyun",
		Status:      "resolved",
		TriggerTime: time.Now(),
		Labels:      datatypes.JSON(`{}`),
	}

	tmplStr := `{"title": "[{{.severity}}] {{.alert_name}}", "content": "{{.message}}"}`

	title, _, err := renderer.RenderAlert(tmplStr, alert, nil)
	if err != nil {
		t.Fatalf("RenderAlert failed: %v", err)
	}

	if contains(title, "[RESOLVED]") {
		t.Errorf("Template without resolved condition should NOT produce [RESOLVED], got '%s'", title)
	}
}