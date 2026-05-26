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

// Renderer handles template rendering with caching support.
type Renderer struct {
	cache sync.Map // template string -> *template.Template
}

// NewRenderer creates a new Renderer instance.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render renders a template string with the given data.
// Templates are cached for performance.
// If template.Execute panics (e.g. nil field accessed as method), the panic is
// recovered and returned as an error, preventing goroutine crash and alert loss.
func (r *Renderer) Render(tmplStr string, data map[string]interface{}) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("template execute panic: %v", r)
		}
	}()

	// Check cache first
 cached, ok := r.cache.Load(tmplStr)
	if ok {
		tmpl := cached.(*template.Template)
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("template execute error: %w", err)
		}
		return buf.String(), nil
	}

	// Parse and cache template
	tmpl, err := template.New("template").Funcs(r.templateFuncMap()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	// Store in cache
	r.cache.Store(tmplStr, tmpl)

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}

// RenderAlert renders an alert notification template.
// Returns title and content strings.
func (r *Renderer) RenderAlert(tmplStr string, alert *models.Alert, routeRule *models.RouteRule) (string, string, error) {
	data := r.buildNotificationRenderContext(alert, routeRule)

	resultStr, err := r.Render(tmplStr, data)
	if err != nil {
		return "", "", err
	}

	// Try to parse as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		// Not JSON, use default title and full result as content
		return "告警通知", resultStr, nil
	}

	// Extract title and content from JSON
	title := strings.TrimSpace(fmt.Sprintf("%v", result["title"]))
	content := strings.TrimSpace(fmt.Sprintf("%v", result["content"]))

	// Apply defaults
	if title == "" {
		title = "告警通知"
	}
	if content == "" {
		content = resultStr
	}

	return title, content, nil
}

// buildNotificationRenderContext builds the render context for an alert.
func (r *Renderer) buildNotificationRenderContext(alert *models.Alert, routeRule *models.RouteRule) map[string]interface{} {
	event := utils.DecodeJSONMap(alert.Raw)

	// Extract severity_raw from event
	severityRaw := ""
	if raw := lookupString(event, "severity", "level", "priority"); raw != "" {
		severityRaw = raw
	}
	if severityRaw == "" {
		if labels, ok := event["labels"].(map[string]interface{}); ok {
			severityRaw = lookupString(labels, "severity", "level", "priority")
		}
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
		"route_name":    "",
	}

	if routeRule != nil {
		data["route_name"] = routeRule.Name
	}

	// Add event field (raw event data)
	data["event"] = event

	// Add nested alert field for convenience
	data["alert"] = map[string]interface{}{
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
	}

	return data
}

// templateFuncMap returns the template function map.
func (r *Renderer) templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"toJson": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"get": func(m interface{}, key string) interface{} {
			if m == nil {
				return nil
			}
			if mm, ok := m.(map[string]interface{}); ok {
				return mm[key]
			}
			return nil
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
		"lookup": func(m interface{}, keys ...string) interface{} {
			if m == nil {
				return nil
			}
			if mm, ok := m.(map[string]interface{}); ok {
				for _, key := range keys {
					if val, ok := mm[key]; ok && val != nil {
						return val
					}
				}
			}
			return nil
		},
	}
}

// lookupString looks up a string value from a map using multiple possible keys.
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