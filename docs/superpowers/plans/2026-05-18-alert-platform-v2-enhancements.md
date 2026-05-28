# Alert Platform v2 Enhancements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the remaining gaps in the Alert Platform v2: backend bug where escalation fields are lost on route rule update, ack dialog missing grouped alert context, AlertCard missing escalation limit tag, and row expander missing structured sections.

**Architecture:** Most of the v2 spec is already implemented (alert aggregation, escalation checker, delivery observability, OnDuty removal). This plan fixes 4 specific gaps found during spec compliance review.

**Tech Stack:** Go (Gin + GORM), React (PrimeReact + Zustand), TypeScript

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/handlers/config.go:418-425` | Modify | Copy escalation fields in UpdateRouteRule |
| `frontend/src/pages/Alerts.tsx:78-85,380-401` | Modify | Track grouped context in ack/silence, update dialog text |
| `frontend/src/components/AlertCard.tsx:57-62` | Modify | Add "已达通知上限" tag |
| `frontend/src/pages/Alerts.tsx:238-349` | Modify | Add section headers and trigger_time/fingerprint to expander |

---

### Task 1: Fix UpdateRouteRule to persist escalation fields

**Files:**
- Modify: `internal/handlers/config.go:418-425`
- Test: `internal/handlers/config_test.go` (new)

The `UpdateRouteRule` handler copies Name, Priority, Severities, Sources, LabelMatchers, ChannelIDs, TimeRanges, and Enabled from the input — but NOT EscalationEnabled, EscalationTimeout, or EscalationMaxRepeats. This means editing a route rule via the frontend silently discards its escalation configuration.

- [ ] **Step 1: Write the failing test**

Create `internal/handlers/config_test.go`:

```go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupConfigTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.RouteRule{}))
	return db
}

func TestUpdateRouteRule_PreservesEscalationFields(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewConfigHandler(db)

	// Create a route rule with escalation enabled
	rule := models.RouteRule{
		Name:                 "test-route",
		Priority:             1,
		Severities:           []byte(`["P0"]`),
		Sources:              []byte(`[]`),
		LabelMatchers:        []byte(`[]`),
		ChannelIDs:           []byte(`[]`),
		TimeRanges:           []byte(`[]`),
		Enabled:              true,
		EscalationEnabled:    true,
		EscalationTimeout:    15,
		EscalationMaxRepeats: 5,
	}
	require.NoError(t, db.Create(&rule).Error)

	// Update via handler (simulating frontend edit that sends all fields)
	body, _ := json.Marshal(map[string]interface{}{
		"name":                  "test-route-updated",
		"priority":              1,
		"severities":            []string{"P0"},
		"sources":               []string{},
		"label_matchers":        []interface{}{},
		"channel_ids":           []int{},
		"time_ranges":           []interface{}{},
		"enabled":               true,
		"escalation_enabled":    true,
		"escalation_timeout":    15,
		"escalation_max_repeats": 5,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateRouteRule(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.RouteRule
	require.NoError(t, db.First(&updated, 1).Error)
	assert.Equal(t, "test-route-updated", updated.Name)
	assert.True(t, updated.EscalationEnabled, "escalation_enabled should be preserved")
	assert.Equal(t, 15, updated.EscalationTimeout, "escalation_timeout should be preserved")
	assert.Equal(t, 5, updated.EscalationMaxRepeats, "escalation_max_repeats should be preserved")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/handlers/ -run TestUpdateRouteRule_PreservesEscalationFields -v -count=1`
Expected: FAIL — `escalation_enabled should be preserved` (the field is false instead of true because UpdateRouteRule doesn't copy it)

- [ ] **Step 3: Fix UpdateRouteRule to copy escalation fields**

In `internal/handlers/config.go`, after line 425 (`rule.Enabled = input.Enabled`), add:

```go
rule.EscalationEnabled = input.EscalationEnabled
rule.EscalationTimeout = input.EscalationTimeout
rule.EscalationMaxRepeats = input.EscalationMaxRepeats
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/handlers/ -run TestUpdateRouteRule_PreservesEscalationFields -v -count=1`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `go test ./... -count=1`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/handlers/config.go internal/handlers/config_test.go
git commit -m "fix: persist escalation fields in UpdateRouteRule handler"
```

---

### Task 2: Show grouped alert context in ack/silence dialogs

**Files:**
- Modify: `frontend/src/pages/Alerts.tsx:33-34,78-85,380-401`

When acking a grouped alert, the spec requires showing "确认此告警（含 N 次触发）" to make it clear what's being acked. Currently the dialog just shows the alert name without any grouped context.

- [ ] **Step 1: Write the failing test**

In `frontend/src/pages/Alerts.test.tsx`, add:

```tsx
it('shows occurrence count in ack dialog for grouped alert', async () => {
  useUserStore.setState({
    user: { ...baseUser, role: 'operator' },
    token: 'token',
  })

  await renderAlerts()

  const ackButton = await screen.findByRole('button', { name: '确认' })
  fireEvent.click(ackButton)

  // Dialog should show grouped context
  expect(await screen.findByText(/含 3 次触发/)).toBeInTheDocument()
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/pages/Alerts.test.tsx --reporter=verbose`
Expected: FAIL — text matching /含 3 次触发/ not found

- [ ] **Step 3: Track grouped context when opening ack/silence dialogs**

In `frontend/src/pages/Alerts.tsx`, add a `selectedGroupedAlert` state to track the grouped context:

```tsx
const [selectedGroupedAlert, setSelectedGroupedAlert] = useState<GroupedActiveAlert | null>(null)
```

Update `handleAck` to accept a `GroupedActiveAlert` and store it:

```tsx
const handleAck = (row: GroupedActiveAlert) => {
  if (!canProcessAlerts) {
    toast.showWarn('当前角色无权执行该操作')
    return
  }
  setSelectedAlert(row.latest_alert)
  setSelectedGroupedAlert(row)
  setAckModalVisible(true)
}
```

Update `handleQuickSilence` similarly:

```tsx
const handleQuickSilence = (row: GroupedActiveAlert) => {
  if (!canProcessAlerts) {
    toast.showWarn('当前角色无权执行该操作')
    return
  }
  setSelectedAlert(row.latest_alert)
  setSelectedGroupedAlert(row)
  setSilenceModalVisible(true)
}
```

Update `actionBodyTemplate` to pass the full row:

```tsx
const actionBodyTemplate = (row: GroupedActiveAlert) => {
  const alert = row.latest_alert
  return (
    <div className="flex gap-1">
      <Button label="投递历史" link size="small" style={{ color: 'var(--primary-color)' }} onClick={() => handleOpenDeliveries(alert)} />
      {alert.status === 'firing' && (
        canProcessAlerts ? (
          <>
            <Button label="确认" link size="small" style={{ color: 'var(--primary-color)' }} onClick={() => handleAck(row)} />
            <Button label="静默" link size="small" style={{ color: 'var(--warning-color)' }} onClick={() => handleQuickSilence(row)} />
          </>
        ) : (
          <Tag value="只读" style={{ background: 'var(--surface-hover)', color: 'var(--text-secondary)' }} />
        )
      )}
    </div>
  )
}
```

Update the ack dialog to show grouped context:

```tsx
<Dialog
  header="确认告警"
  visible={ackModalVisible}
  onHide={() => setAckModalVisible(false)}
  footer={
    <div>
      <Button label="取消" outlined onClick={() => setAckModalVisible(false)} />
      <Button label="确认" onClick={handleAckConfirm} />
    </div>
  }
>
  <div className="flex flex-column gap-3">
    <p>
      确认告警: <strong>{selectedAlert?.alert_name}</strong>
      {selectedGroupedAlert && selectedGroupedAlert.count > 1 && (
        <Tag
          value={`含 ${selectedGroupedAlert.count} 次触发`}
          style={{
            marginLeft: '8px',
            background: 'var(--warning-light-color)',
            color: 'var(--warning-color)',
            border: '1px solid var(--warning-color)',
          }}
        />
      )}
    </p>
    <InputTextarea
      rows={3}
      placeholder="添加备注（可选）"
      value={ackComment}
      onChange={(e) => setAckComment(e.target.value)}
    />
  </div>
</Dialog>
```

Update the silence dialog similarly:

```tsx
<p>
  静默告警: <strong>{selectedAlert?.alert_name}</strong>
  {selectedGroupedAlert && selectedGroupedAlert.count > 1 && (
    <Tag
      value={`含 ${selectedGroupedAlert.count} 次触发`}
      style={{
        marginLeft: '8px',
        background: 'var(--warning-light-color)',
        color: 'var(--warning-color)',
        border: '1px solid var(--warning-color)',
      }}
    />
  )}
</p>
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/pages/Alerts.test.tsx --reporter=verbose`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/Alerts.tsx frontend/src/pages/Alerts.test.tsx
git commit -m "feat: show grouped alert occurrence count in ack/silence dialogs"
```

---

### Task 3: Add "已达通知上限" tag to AlertCard

**Files:**
- Modify: `frontend/src/components/AlertCard.tsx:8-13,57-62`
- Test: `frontend/src/pages/Dashboard.test.tsx`

The `AlertCard` component (used on the Dashboard) shows "已通知 N 次" but doesn't show "已达通知上限" when the alert has reached its escalation limit. The `Alerts.tsx` page has this logic, but `AlertCard.tsx` doesn't.

- [ ] **Step 1: Write the failing test**

In `frontend/src/pages/Dashboard.test.tsx`, add a test with a notified alert that has reached its limit:

```tsx
it('shows escalation limit tag on AlertCard', async () => {
  const notifiedAlert: Alert = {
    ...firingAlert,
    notify_count: 4,
    last_notified_at: new Date(Date.now() - 180 * 60 * 1000).toISOString(), // 3 hours ago
  }
  alertStoreState.groupedActiveAlerts = [{
    ...groupedAlert,
    latest_alert: notifiedAlert,
  }]

  await renderDashboard()

  expect(await screen.findByText('已达通知上限')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/pages/Dashboard.test.tsx --reporter=verbose`
Expected: FAIL — text "已达通知上限" not found

- [ ] **Step 3: Add escalation limit detection to AlertCard**

In `frontend/src/components/AlertCard.tsx`, update the props interface to accept optional escalation info, and add the "已达通知上限" tag:

```tsx
interface AlertCardProps {
  alert: Alert
  onAck?: (alert: Alert) => void
  onQuickSilence?: (alert: Alert) => void
  showActions?: boolean
  escalationMaxRepeats?: number
}
```

Update the component to destructure the new prop and add the tag:

```tsx
export const AlertCard: React.FC<AlertCardProps> = ({
  alert,
  onAck,
  onQuickSilence,
  showActions = true,
  escalationMaxRepeats,
}) => {
```

After the existing `notify_count` tag (around line 62), add:

```tsx
{alert.notify_count > 0 && escalationMaxRepeats !== undefined && alert.notify_count >= escalationMaxRepeats + 1 && (
  <Tag
    value="已达通知上限"
    severity="warning"
  />
)}
```

- [ ] **Step 4: Pass escalationMaxRepeats from Dashboard.tsx**

In `frontend/src/pages/Dashboard.tsx`, when rendering `AlertCard`, pass the escalation info. Since the Dashboard doesn't have direct access to route rules, use a heuristic similar to `Alerts.tsx`: if `last_notified_at` is more than 2 hours ago and the alert is still firing, it likely reached the limit.

Alternatively, pass a simple computed flag. The simplest approach matching the existing `Alerts.tsx` pattern:

In `Dashboard.tsx`, update the AlertCard rendering:

```tsx
{sortedGroupedAlerts.map((grouped) => {
  const alert = grouped.latest_alert
  const reachedEscalationLimit = alert.status === 'firing'
    && alert.notify_count > 0
    && alert.last_notified_at
    && dayjs().diff(dayjs(alert.last_notified_at), 'minute') > 120

  return (
    <div key={grouped.fingerprint} className="mb-3">
      <AlertCard
        alert={alert}
        showActions={true}
        onAck={handleAck}
        onQuickSilence={handleQuickSilence}
        escalationMaxRepeats={reachedEscalationLimit ? alert.notify_count - 1 : undefined}
      />
      {grouped.count > 1 && (
        <div className="mt-1 ml-4">
          <Tag
            value={`共 ${grouped.count} 次`}
            style={{
              background: 'var(--warning-light-color)',
              color: 'var(--warning-color)',
              border: '1px solid var(--warning-color)',
            }}
          />
        </div>
      )}
    </div>
  )
})}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/pages/Dashboard.test.tsx --reporter=verbose`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/AlertCard.tsx frontend/src/pages/Dashboard.tsx frontend/src/pages/Dashboard.test.tsx
git commit -m "feat: show escalation limit tag on AlertCard in dashboard"
```

---

### Task 4: Add structured sections to row expander

**Files:**
- Modify: `frontend/src/pages/Alerts.tsx:238-349`

The spec requires the row expander to show three structured sections: (1) 告警信息 with message, labels, trigger_time, fingerprint; (2) 投递记录; (3) 操作. Currently it shows message+labels without a section header, then delivery records, but no trigger_time, fingerprint, or action buttons.

- [ ] **Step 1: Update rowExpansionTemplate with structured sections**

In `frontend/src/pages/Alerts.tsx`, replace the `rowExpansionTemplate` function body:

```tsx
const rowExpansionTemplate = (row: GroupedActiveAlert) => {
  const alert = row.latest_alert
  const deliveries = alertDeliveries[alert.alert_id]
  const loading = deliveriesLoading[alert.alert_id]

  return (
    <div className="p-3" style={{ background: 'var(--surface-hover)', borderRadius: '8px' }}>
      {/* Section 1: 告警信息 */}
      <div className="mb-3">
        <strong style={{ color: 'var(--text-primary)', fontSize: '0.875rem' }}>告警信息</strong>
        <div className="mt-2">
          <p className="m-0 mb-2" style={{ color: 'var(--text-primary)' }}>
            <strong>消息:</strong> {alert.message}
          </p>
          <div className="flex gap-1 flex-wrap align-items-center mb-2">
            <strong style={{ color: 'var(--text-primary)' }}>Labels:</strong>
            {alert.labels && Object.entries(alert.labels).map(([k, v]) => (
              <Tag
                key={k}
                value={`${k}: ${String(v)}`}
                style={{
                  background: 'var(--surface-card)',
                  color: 'var(--text-secondary)',
                  marginLeft: '4px',
                }}
              />
            ))}
          </div>
          <div className="flex gap-4 text-sm" style={{ color: 'var(--text-disabled)' }}>
            <span>触发时间: {dayjs(alert.trigger_time).format('YYYY-MM-DD HH:mm:ss')}</span>
            <span>Fingerprint: <code style={{ fontSize: '0.8rem' }}>{alert.fingerprint}</code></span>
          </div>
        </div>
      </div>

      {/* Section 2: 投递记录 */}
      <div className="mb-3">
        <strong style={{ color: 'var(--text-primary)', fontSize: '0.875rem' }}>投递记录</strong>
        {loading ? (
          <div className="flex justify-content-center p-3">
            <ProgressSpinner style={{ width: '24px', height: '24px' }} />
          </div>
        ) : deliveries && deliveries.length > 0 ? (
          <DataTable
            value={deliveries}
            dataKey="id"
            size="small"
            stripedRows
            className="mt-2"
            style={{ fontSize: '0.875rem' }}
          >
            <Column
              header="渠道"
              style={{ width: '140px' }}
              body={(d: Delivery) => (
                <span>{d.channel_snapshot?.name || `#${d.channel_id}`}</span>
              )}
            />
            <Column
              header="状态"
              style={{ width: '90px' }}
              body={(d: Delivery) => {
                const cfg = deliveryStatusConfig[d.delivery_status] || {
                  label: d.delivery_status,
                  bgColor: 'var(--surface-hover)',
                  color: 'var(--text-secondary)',
                }
                return (
                  <Tag
                    value={cfg.label}
                    style={{
                      background: cfg.bgColor,
                      color: cfg.color,
                      border: `1px solid ${cfg.color}`,
                    }}
                  />
                )
              }}
            />
            <Column
              header="尝试次数"
              style={{ width: '80px' }}
              field="attempt_count"
            />
            <Column
              header="通知时间"
              style={{ width: '160px' }}
              body={(d: Delivery) => (
                <span>{dayjs(d.created_at).format('YYYY-MM-DD HH:mm:ss')}</span>
              )}
            />
            <Column
              header="触发类型"
              style={{ width: '100px' }}
              body={(d: Delivery) => {
                const kind = d.attempts?.length > 0
                  ? d.attempts[d.attempts.length - 1].trigger_kind
                  : (d.final_failure_summary?.trigger_kind || '')
                const label = triggerKindLabels[kind] || kind
                return <span>{label}</span>
              }}
            />
            <Column
              header="错误信息"
              body={(d: Delivery) =>
                d.delivery_status === 'failed' && d.final_failure_summary ? (
                  <span style={{ color: 'var(--danger-color)' }}>
                    {d.final_failure_summary.error_message}
                  </span>
                ) : (
                  <span style={{ color: 'var(--text-disabled)' }}>-</span>
                )
              }
            />
          </DataTable>
        ) : deliveries && deliveries.length === 0 ? (
          <div className="mt-2 text-sm" style={{ color: 'var(--text-disabled)' }}>
            暂无投递记录
          </div>
        ) : null}
      </div>

      {/* Section 3: 操作 */}
      {canProcessAlerts && alert.status === 'firing' && (
        <div>
          <strong style={{ color: 'var(--text-primary)', fontSize: '0.875rem' }}>操作</strong>
          <div className="flex gap-2 mt-2">
            <Button
              label="确认"
              icon="pi pi-check"
              size="small"
              style={{ background: 'var(--primary-color)', border: 'none' }}
              onClick={() => handleAck(row)}
            />
            <Button
              label="静默"
              icon="pi pi-volume-off"
              size="small"
              style={{
                background: 'var(--warning-light-color)',
                color: 'var(--warning-color)',
                border: '1px solid var(--warning-color)',
              }}
              onClick={() => handleQuickSilence(row)}
            />
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Run frontend type check**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS (no type errors)

- [ ] **Step 3: Run existing tests**

Run: `cd frontend && npx vitest run --reporter=verbose`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/Alerts.tsx
git commit -m "feat: add structured sections to alert row expander with trigger time and fingerprint"
```
