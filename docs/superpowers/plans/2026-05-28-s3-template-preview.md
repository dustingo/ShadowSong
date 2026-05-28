# S3: Notification Template Preview

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let operators preview the final rendered notification content for a channel without actually sending it.

**Architecture:** A new `POST /api/v1/channels/:id/preview` endpoint loads the channel, a sample or specified alert, renders through the output template, and returns the result. Frontend adds a "预览" button that opens a dialog showing the rendered title and content.

**Tech Stack:** Go, GORM, Gin, PrimeReact, TypeScript

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `internal/handlers/config.go` | Add PreviewChannel handler |
| Modify | `internal/handlers/config_test.go` | Test preview endpoint |
| Modify | `internal/router/router.go` | Register preview route |
| Modify | `frontend/src/pages/Channels.tsx` | Add preview button and dialog |
| Modify | `frontend/src/api/client.ts` | Add preview API call |

---

## Task 1: Add PreviewChannel backend endpoint

**Files:**
- Modify: `internal/handlers/config.go`
- Modify: `internal/router/router.go`

- [ ] **Step 1: Add PreviewChannel handler**

In `internal/handlers/config.go`, add:

```go
// PreviewChannel renders the notification content for a channel without sending it.
func (h *ConfigHandler) PreviewChannel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	var channel models.Channel
	if err := h.db.First(&channel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the most recent firing alert as sample, or create a synthetic one
	var sampleAlert models.Alert
	err = h.db.Where("status = ?", "firing").Order("trigger_time DESC").First(&sampleAlert).Error
	if err != nil {
		// No firing alerts — use synthetic
		sampleAlert = models.Alert{
			AlertID:     "preview-sample",
			Source:      "preview",
			AlertName:   "SampleAlert",
			Severity:    "P1",
			Message:     "This is a sample alert for template preview",
			Status:      "firing",
			TriggerTime: time.Now(),
			Labels:      []byte(`{"host":"game-server-01","zone":"east-1"}`),
		}
	}

	// Build render context
	renderer := template.NewRenderer()
	data := map[string]interface{}{
		"alert_id":   sampleAlert.AlertID,
		"alert_name": sampleAlert.AlertName,
		"severity":   sampleAlert.Severity,
		"message":    sampleAlert.Message,
		"source":     sampleAlert.Source,
		"status":     sampleAlert.Status,
		"labels":     utils.DecodeJSONMap(sampleAlert.Labels),
	}

	// Try to find a matching route rule for output template
	var routeRule *models.RouteRule
	var routes []models.RouteRule
	h.db.Where("enabled = ?", true).Order("priority ASC").Find(&routes)
	matcher := routing.NewMatcher(&gormAdapter{db: h.db})
	targets, _ := matcher.FindMatchedTargets(&sampleAlert, routes)
	if len(targets) > 0 {
		routeRule = &targets[0].RouteRule
	}

	// Render using output template from a datasource matching the alert source
	var ds models.DataSource
	var outputTmpl string
	if err := h.db.Where("name = ?", sampleAlert.Source).First(&ds).Error; err == nil {
		outputTmpl = ds.OutputTemplate
	}

	title := fmt.Sprintf("[%s] %s", sampleAlert.Severity, sampleAlert.AlertName)
	content := sampleAlert.Message

	if outputTmpl != "" && routeRule != nil {
		renderedTitle, renderedContent, renderErr := renderer.RenderAlert(outputTmpl, &sampleAlert, routeRule)
		if renderErr == nil {
			title = renderedTitle
			content = renderedContent
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"title":        title,
		"content":      content,
		"channel_type": channel.Type,
		"channel_name": channel.Name,
		"alert_source": sampleAlert.Source,
	})
}
```

Add required imports to `config.go` if not already present:

```go
"github.com/game-ops/ai-alert-system/internal/template"
"github.com/game-ops/ai-alert-system/internal/routing"
"github.com/game-ops/ai-alert-system/internal/utils"
```

- [ ] **Step 2: Add gormAdapter for routing matcher**

If `gormAdapter` is not already defined in `config.go`, add:

```go
type gormAdapter struct {
	db *gorm.DB
}

func (a *gormAdapter) Find(dst interface{}, conds ...interface{}) error {
	return a.db.Find(dst, conds...).Error
}
```

- [ ] **Step 3: Register route**

In `internal/router/router.go`, in the channels group, add:

```go
channels.POST("/:id/preview", middleware.RequireCapability(authz.CapabilityViewConfig), configHandler.PreviewChannel)
```

- [ ] **Step 4: Verify compilation**

Run: `cd D:\goproject\shadowsongAI && go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/config.go internal/router/router.go
git commit -m "feat(channels): add notification template preview endpoint"
```

---

## Task 2: Add preview to frontend

**Files:**
- Modify: `frontend/src/api/client.ts`
- Modify: `frontend/src/pages/Channels.tsx`

- [ ] **Step 1: Add preview API method**

In `frontend/src/api/client.ts`, in the `channelApi` object, add:

```ts
preview: (id: number) =>
  client.post(`/channels/${id}/preview`).then((res) => res.data),
```

- [ ] **Step 2: Add preview button and dialog to Channels page**

In `frontend/src/pages/Channels.tsx`, add state:

```tsx
const [previewVisible, setPreviewVisible] = useState(false)
const [previewData, setPreviewData] = useState<{ title: string; content: string } | null>(null)
```

Add handler:

```tsx
const handlePreview = async (record: Channel) => {
  try {
    const data = await channelApi.preview(record.id)
    setPreviewData({ title: data.title, content: data.content })
    setPreviewVisible(true)
  } catch {
    toast.showError('预览失败')
  }
}
```

In the actions column, add a preview button:

```tsx
<Button
  icon="pi pi-eye"
  className="p-button-text p-button-info"
  tooltip="预览通知"
  onClick={() => handlePreview(record)}
/>
```

Add the preview dialog:

```tsx
<Dialog
  header="通知预览"
  visible={previewVisible}
  style={{ width: '50vw' }}
  onHide={() => setPreviewVisible(false)}
>
  {previewData && (
    <div>
      <h4>标题</h4>
      <pre style={{ background: '#f5f5f5', padding: '1rem', borderRadius: '4px' }}>{previewData.title}</pre>
      <h4>内容</h4>
      <pre style={{ background: '#f5f5f5', padding: '1rem', borderRadius: '4px', whiteSpace: 'pre-wrap' }}>{previewData.content}</pre>
    </div>
  )}
</Dialog>
```

- [ ] **Step 3: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/client.ts frontend/src/pages/Channels.tsx
git commit -m "feat(ui): add notification template preview to Channels page"
```

---

## Task 3: Final integration

- [ ] **Step 1: Run all backend tests**

Run: `cd D:\goproject\shadowsongAI && go test ./... -v 2>&1 | Select-Object -Last 30`
Expected: all PASS

- [ ] **Step 2: Run frontend type check**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit if any fixups**

```bash
git add -A
git commit -m "chore: s3 final integration fixes"
```
