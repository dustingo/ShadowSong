# S2: Alert List Filters + Batch Operations + Data Retention

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Improve daily operations UX with quick alert filters, batch ack/silence, and automated data retention cleanup.

**Architecture:** Alert list filters are frontend-only (TabView + existing status query param). Batch operations add two new POST endpoints in `alert.go` that process arrays in a single transaction. Data retention is a background goroutine using raw SQL DELETE with configurable TTL.

**Tech Stack:** Go, GORM, Gin, PrimeReact, TypeScript, Zustand

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `frontend/src/pages/Alerts.tsx` | Add status tabs + multi-select + batch action bar |
| Modify | `frontend/src/stores/alertStore.ts` | Add batch ack/silence actions |
| Modify | `frontend/src/api/client.ts` | Add batch API calls |
| Modify | `frontend/src/types/index.ts` | Add batch request/response types |
| Modify | `internal/handlers/alert.go` | Add BatchAck and BatchSilence handlers |
| Modify | `internal/handlers/alert_test.go` | Tests for batch endpoints |
| Modify | `internal/router/router.go` | Register batch routes |
| Create | `internal/retention/cleanup.go` | Background data retention cleanup |
| Create | `internal/retention/cleanup_test.go` | Retention unit tests |
| Modify | `internal/config/config.go` | Add RetentionDays config |
| Modify | `cmd/server/main.go` | Wire retention goroutine |

---

## Task 1: Add status filter tabs to Alerts page

**Files:**
- Modify: `frontend/src/pages/Alerts.tsx`

- [ ] **Step 1: Add TabView import**

Add to imports at top of `Alerts.tsx`:

```tsx
import { TabView, TabPanel } from 'primereact/tabview'
```

- [ ] **Step 2: Add status tab state and handler**

Add state for the active tab index:

```tsx
const [activeTabIndex, setActiveTabIndex] = useState(0)
```

Add a mapping from tab index to status filter:

```tsx
const statusTabs = [
  { label: '全部', status: undefined },
  { label: '活跃', status: 'firing' },
  { label: '已确认', status: 'acked' },
  { label: '已静默', status: 'silenced' },
  { label: '已恢复', status: 'resolved' },
]
```

Add handler:

```tsx
const handleTabChange = (e: { index: number }) => {
  setActiveTabIndex(e.index)
  const tab = statusTabs[e.index]
  setFilters({ ...filters, status: tab.status })
}
```

- [ ] **Step 3: Add TabView above the DataTable**

Insert before the DataTable:

```tsx
<TabView activeIndex={activeTabIndex} onTabChange={handleTabChange}>
  {statusTabs.map((tab) => (
    <TabPanel key={tab.label} header={tab.label} />
  ))}
</TabView>
```

- [ ] **Step 4: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/Alerts.tsx
git commit -m "feat(ui): add status filter tabs to Alerts page"
```

---

## Task 2: Add batch API types and client methods

**Files:**
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/api/client.ts`

- [ ] **Step 1: Add batch types**

In `frontend/src/types/index.ts`, add at the end:

```ts
export interface BatchAckRequest {
  alert_ids: string[]
  comment: string
}

export interface BatchSilenceRequest {
  alert_ids: string[]
  duration: number
}

export interface BatchResult {
  updated: number
  skipped: number
  errors: string[]
}
```

- [ ] **Step 2: Add batch API methods**

In `frontend/src/api/client.ts`, in the `alertApi` object, add:

```ts
batchAck: (data: BatchAckRequest) =>
  client.post('/alerts/batch-ack', data).then((res) => res.data),

batchSilence: (data: BatchSilenceRequest) =>
  client.post('/alerts/batch-silence', data).then((res) => res.data),
```

- [ ] **Step 3: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/api/client.ts
git commit -m "feat(api): add batch ack/silence types and client methods"
```

---

## Task 3: Add batch store actions

**Files:**
- Modify: `frontend/src/stores/alertStore.ts`

- [ ] **Step 1: Add batch actions to store**

In `frontend/src/stores/alertStore.ts`, add to the `AlertState` interface:

```ts
batchAck: (ids: string[], comment: string) => Promise<BatchResult>
batchSilence: (ids: string[], duration: number) => Promise<BatchResult>
```

Add the implementations:

```ts
batchAck: async (ids, comment) => {
  const result = await alertApi.batchAck({ alert_ids: ids, comment })
  // Update local state
  set((state) => ({
    alerts: state.alerts.map((a) =>
      ids.includes(a.alert_id)
        ? { ...a, status: 'acked', acked_at: new Date().toISOString(), ack_comment: comment }
        : a
    ),
    activeAlerts: state.activeAlerts.filter((a) => !ids.includes(a.alert_id)),
  }))
  get().fetchGroupedActiveAlerts()
  get().fetchStats()
  return result
},

batchSilence: async (ids, duration) => {
  const result = await alertApi.batchSilence({ alert_ids: ids, duration })
  set((state) => ({
    alerts: state.alerts.map((a) =>
      ids.includes(a.alert_id) ? { ...a, status: 'silenced' } : a
    ),
    activeAlerts: state.activeAlerts.filter((a) => !ids.includes(a.alert_id)),
  }))
  get().fetchGroupedActiveAlerts()
  get().fetchStats()
  return result
},
```

- [ ] **Step 2: Add import for BatchResult type**

Add to imports:

```ts
import type { Alert, GroupedActiveAlert, BatchResult } from '../types'
```

- [ ] **Step 3: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/stores/alertStore.ts
git commit -m "feat(store): add batch ack/silence actions"
```

---

## Task 4: Add multi-select and batch action bar to Alerts page

**Files:**
- Modify: `frontend/src/pages/Alerts.tsx`

- [ ] **Step 1: Add selection state**

```tsx
const [selectedAlerts, setSelectedAlerts] = useState<Alert[]>([])
```

- [ ] **Step 2: Add batch handler functions**

```tsx
const { batchAck, batchSilence } = useAlertStore()

const handleBatchAck = async () => {
  try {
    const ids = selectedAlerts.map((a) => a.alert_id)
    await batchAck(ids, 'batch ack')
    toast.showSuccess(`已确认 ${ids.length} 条告警`)
    setSelectedAlerts([])
  } catch {
    toast.showError('批量确认失败')
  }
}

const handleBatchSilence = async () => {
  try {
    const ids = selectedAlerts.map((a) => a.alert_id)
    await batchSilence(ids, 3600)
    toast.showSuccess(`已静默 ${ids.length} 条告警`)
    setSelectedAlerts([])
  } catch {
    toast.showError('批量静默失败')
  }
}
```

- [ ] **Step 3: Add selection to DataTable and action bar**

In the DataTable, add selection props:

```tsx
<DataTable
  value={alerts}
  selection={selectedAlerts}
  onSelectionChange={(e) => setSelectedAlerts(e.value)}
  dataKey="alert_id"
  // ... existing props
>
  <Column selectionMode="multiple" headerStyle={{ width: '3rem' }} />
  {/* ... existing columns */}
</DataTable>
```

Add action bar above the table (show when selection is non-empty):

```tsx
{selectedAlerts.length > 0 && (
  <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', alignItems: 'center' }}>
    <span>已选 {selectedAlerts.length} 条</span>
    <Button label="批量确认" icon="pi pi-check" severity="success" onClick={handleBatchAck} />
    <Button label="批量静默" icon="pi pi-volume-off" severity="warning" onClick={handleBatchSilence} />
    <Button label="取消选择" icon="pi pi-times" severity="secondary" onClick={() => setSelectedAlerts([])} />
  </div>
)}
```

- [ ] **Step 4: Verify frontend compiles**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/Alerts.tsx
git commit -m "feat(ui): add multi-select and batch action bar to Alerts"
```

---

## Task 5: Add batch backend endpoints

**Files:**
- Modify: `internal/handlers/alert.go` — add BatchAck and BatchSilence
- Modify: `internal/handlers/alert_test.go` — test batch endpoints
- Modify: `internal/router/router.go` — register routes

- [ ] **Step 1: Write the failing tests**

In `internal/handlers/alert_test.go`, add:

```go
func TestBatchAck(t *testing.T) {
	db := newAlertTestDB(t)
	handler := NewAlertHandler(db)

	// Create two firing alerts
	for _, id := range []string{"batch-1", "batch-2"} {
		require.NoError(t, db.Create(&models.Alert{
			AlertID:     id,
			Source:      "test",
			AlertName:   "TestAlert",
			Severity:    "P1",
			Message:     "test",
			Fingerprint: "fp-" + id,
			Status:      "firing",
			TriggerTime: time.Now(),
			ReceivedAt:  time.Now(),
		}).Error)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"alert_ids":["batch-1","batch-2"],"comment":"batch test"}`
	c.Request = httptest.NewRequest("POST", "/alerts/batch-ack", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.BatchAck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var alert1 models.Alert
	db.First(&alert1, "alert_id = ?", "batch-1")
	assert.Equal(t, "acked", alert1.Status)
	assert.Equal(t, "batch test", alert1.AckComment)
}

func TestBatchSilence(t *testing.T) {
	db := newAlertTestDB(t)
	handler := NewAlertHandler(db)

	require.NoError(t, db.Create(&models.Alert{
		AlertID:     "batch-silence-1",
		Source:      "test",
		AlertName:   "DiskFull",
		Severity:    "P0",
		Message:     "disk full",
		Fingerprint: "fp-batch-silence",
		Status:      "firing",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
	}).Error)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"alert_ids":["batch-silence-1"],"duration":3600}`
	c.Request = httptest.NewRequest("POST", "/alerts/batch-silence", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.BatchSilence(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var alert models.Alert
	db.First(&alert, "alert_id = ?", "batch-silence-1")
	assert.Equal(t, "silenced", alert.Status)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -run TestBatch -v`
Expected: FAIL — methods not defined

- [ ] **Step 3: Implement BatchAck**

In `internal/handlers/alert.go`:

```go
func (h *AlertHandler) BatchAck(c *gin.Context) {
	var input struct {
		AlertIDs []string `json:"alert_ids" binding:"required"`
		Comment  string   `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := 0
	skipped := 0
	var errs []string
	username := middleware.GetUsername(c)

	for _, id := range input.AlertIDs {
		var alert models.Alert
		if err := h.db.First(&alert, "alert_id = ?", id).Error; err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("%s: not found", id))
			continue
		}
		if err := alert.Ack(username, input.Comment); err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("%s: %s", id, err.Error()))
			continue
		}
		if err := h.db.Save(&alert).Error; err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("%s: db error", id))
			continue
		}
		updated++
	}

	c.JSON(http.StatusOK, gin.H{
		"updated": updated,
		"skipped": skipped,
		"errors":  errs,
	})
}
```

- [ ] **Step 4: Implement BatchSilence**

```go
func (h *AlertHandler) BatchSilence(c *gin.Context) {
	var input struct {
		AlertIDs []string `json:"alert_ids" binding:"required"`
		Duration int      `json:"duration" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := 0
	skipped := 0
	var errs []string
	username := middleware.GetUsername(c)
	if username == "" {
		username = "system"
	}
	now := time.Now()

	// Collect unique alert names for silence rules
	alertNames := make(map[string]bool)

	for _, id := range input.AlertIDs {
		var alert models.Alert
		if err := h.db.First(&alert, "alert_id = ?", id).Error; err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("%s: not found", id))
			continue
		}

		alert.Status = "silenced"
		if err := h.db.Save(&alert).Error; err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("%s: db error", id))
			continue
		}

		alertNames[alert.AlertName] = true
		updated++
	}

	// Create silence rules per unique alert name
	for name := range alertNames {
		silence := models.SilenceRule{
			Name:             "Batch Silence - " + name,
			AlertNamePattern: name,
			Severities:       []byte("[]"),
			StartsAt:         now,
			EndsAt:           now.Add(time.Duration(input.Duration) * time.Second),
			CreatedBy:        username,
		}
		h.db.Create(&silence)
	}

	c.JSON(http.StatusOK, gin.H{
		"updated": updated,
		"skipped": skipped,
		"errors":  errs,
	})
}
```

- [ ] **Step 5: Register routes**

In `internal/router/router.go`, add inside the alerts group:

```go
alerts.POST("/batch-ack", middleware.RequireCapability(authz.CapabilityProcessAlerts), alertHandler.BatchAck)
alerts.POST("/batch-silence", middleware.RequireCapability(authz.CapabilityProcessAlerts), alertHandler.BatchSilence)
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/handlers/ -run TestBatch -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/handlers/alert.go internal/handlers/alert_test.go internal/router/router.go
git commit -m "feat(alerts): add batch ack and batch silence endpoints"
```

---

## Task 6: Data retention cleanup

**Files:**
- Create: `internal/retention/cleanup.go`
- Create: `internal/retention/cleanup_test.go`
- Modify: `internal/config/config.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Add RetentionDays to config**

In `internal/config/config.go`, add to `ServerConfig`:

```go
RetentionDays int
```

In the `Load()` function, add to `Server`:

```go
RetentionDays: getEnvAsInt("ALERT_RETENTION_DAYS", 30),
```

- [ ] **Step 2: Write the failing test**

Create `internal/retention/cleanup_test.go`:

```go
package retention

import (
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRetentionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}, &models.NotificationDelivery{}, &models.NotificationDeliveryAttempt{}))
	return db
}

func TestCleanupDeletesOldRecords(t *testing.T) {
	db := newRetentionTestDB(t)

	// Create an old alert (40 days ago)
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	require.NoError(t, db.Create(&models.Alert{
		AlertID:     "old-alert",
		Source:      "test",
		AlertName:   "Old",
		Severity:    "P1",
		Message:     "old",
		Fingerprint: "fp-old",
		Status:      "resolved",
		TriggerTime: oldTime,
		ReceivedAt:  oldTime,
	}).Error)

	// Create a recent alert
	require.NoError(t, db.Create(&models.Alert{
		AlertID:     "new-alert",
		Source:      "test",
		AlertName:   "New",
		Severity:    "P1",
		Message:     "new",
		Fingerprint: "fp-new",
		Status:      "firing",
		TriggerTime: time.Now(),
		ReceivedAt:  time.Now(),
	}).Error)

	result := Cleanup(db, 30)

	assert.Equal(t, int64(1), result.AlertsDeleted)

	// Verify recent alert still exists
	var count int64
	db.Model(&models.Alert{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestCleanupDisabledWhenZero(t *testing.T) {
	db := newRetentionTestDB(t)
	result := Cleanup(db, 0)
	assert.Equal(t, int64(0), result.AlertsDeleted)
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/retention/ -v`
Expected: FAIL — package not found

- [ ] **Step 4: Implement cleanup**

Create `internal/retention/cleanup.go`:

```go
package retention

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// CleanupResult reports how many records were deleted.
type CleanupResult struct {
	AlertsDeleted      int64
	DeliveriesDeleted  int64
	AttemptsDeleted    int64
}

// Cleanup deletes records older than retentionDays. Returns immediately if retentionDays <= 0.
func Cleanup(db *gorm.DB, retentionDays int) CleanupResult {
	if retentionDays <= 0 {
		return CleanupResult{}
	}

	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	var result CleanupResult

	// Delete old delivery attempts first (child table)
	res := db.Exec("DELETE FROM notification_delivery_attempts WHERE created_at < ?", cutoff)
	result.AttemptsDeleted = res.RowsAffected

	// Delete old deliveries
	res = db.Exec("DELETE FROM notification_deliveries WHERE created_at < ?", cutoff)
	result.DeliveriesDeleted = res.RowsAffected

	// Delete old recovery records
	db.Exec("DELETE FROM notification_delivery_recoveries WHERE created_at < ?", cutoff)

	// Delete old alerts
	res = db.Exec("DELETE FROM alerts WHERE created_at < ?", cutoff)
	result.AlertsDeleted = res.RowsAffected

	log.Printf("retention: cleaned alerts=%d deliveries=%d attempts=%d (retention_days=%d)",
		result.AlertsDeleted, result.DeliveriesDeleted, result.AttemptsDeleted, retentionDays)

	return result
}

// Run starts the retention cleanup loop. It blocks until stop is closed.
func Run(db *gorm.DB, retentionDays int, interval time.Duration, stop <-chan struct{}) {
	// Run once at startup
	Cleanup(db, retentionDays)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			Cleanup(db, retentionDays)
		case <-stop:
			return
		}
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd D:\goproject\shadowsongAI && go test ./internal/retention/ -v`
Expected: PASS

- [ ] **Step 6: Wire into main.go**

In `cmd/server/main.go`, add import:

```go
"github.com/game-ops/ai-alert-system/internal/retention"
```

After database initialization and before router setup, add:

```go
if cfg.Server.RetentionDays > 0 {
	retentionStop := make(chan struct{})
	go retention.Run(db, cfg.Server.RetentionDays, 1*time.Hour, retentionStop)
	log.Printf("Data retention enabled: %d days", cfg.Server.RetentionDays)
}
```

- [ ] **Step 7: Verify full build**

Run: `cd D:\goproject\shadowsongAI && go build ./...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add internal/retention/cleanup.go internal/retention/cleanup_test.go internal/config/config.go cmd/server/main.go
git commit -m "feat(retention): add background data retention cleanup"
```

---

## Task 7: Final integration

- [ ] **Step 1: Run all backend tests**

Run: `cd D:\goproject\shadowsongAI && go test ./... -v 2>&1 | Select-Object -Last 30`
Expected: all PASS

- [ ] **Step 2: Run frontend type check**

Run: `cd D:\goproject\shadowsongAI\frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit if any fixups**

```bash
git add -A
git commit -m "chore: s2 final integration fixes"
```
