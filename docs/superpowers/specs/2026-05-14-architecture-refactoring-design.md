# Architecture Refactoring Design Document

**Date:** 2026-05-14
**Status:** Approved
**Scope:** Efficiency optimization and code deduplication

---

## Problem Statement

Code review identified multiple issues affecting performance and maintainability:

1. **N+1 Query Pattern**: `findMatchedChannels` loads channels one-by-one in a loop
2. **Sequential DB Queries**: `Stats()` executes 32 separate queries
3. **WebSocket Goroutine Leak**: Heartbeat goroutine has no exit mechanism
4. **Code Duplication**: Route matching and template rendering logic duplicated across files

---

## Architecture Overview

### New Module Structure

```
internal/
├── routing/
│   ├── matcher.go      # Route matching logic (unified)
│   └── matcher_test.go
├── template/
│   ├── renderer.go     # Template rendering with caching
│   └── renderer_test.go
├── stats/
│   ├── queries.go      # Optimized statistics queries
│   └── queries_test.go
└── websocket/
    ├── hub.go          # Connection hub
    ├── client.go       # Individual client with lifecycle
    └── client_test.go
```

---

## Component Details

### 1. Routing Module (`internal/routing`)

**Responsibility**: Unified alert-to-channel routing logic

**Interface**:
```go
package routing

type Matcher struct {
    db *gorm.DB
}

type MatchTarget struct {
    Channel   models.Channel
    RouteRule *models.RouteRule
}

func NewMatcher(db *gorm.DB) *Matcher
func (m *Matcher) FindMatchedTargets(alert *models.Alert, rules []models.RouteRule) ([]MatchTarget, error)
```

**Optimization**: Batch load all channels with single `WHERE id IN (?)` query.

**Shared Functions** (moved from webhook.go and service.go):
- `MatchLabels(labelsJSON []byte, matchers []models.LabelMatcher) bool`
- `IsInTimeRange(timeRangesJSON []byte) bool`

---

### 2. Template Module (`internal/template`)

**Responsibility**: Template rendering with caching

**Interface**:
```go
package template

type Renderer struct {
    cache sync.Map  // template string -> *template.Template
}

func NewRenderer() *Renderer
func (r *Renderer) Render(tmplStr string, data map[string]interface{}) (string, error)
func (r *Renderer) RenderAlert(tmplStr string, alert *models.Alert) (title, content string, error)
```

**Template Functions** (shared):
- `toJson`, `get`, `default`, `lookup`

---

### 3. Stats Module (`internal/stats`)

**Responsibility**: Optimized statistics queries

**Interface**:
```go
package stats

type AlertStats struct {
    Total      int64            `json:"total"`
    Firing     int64            `json:"firing"`
    Acked      int64            `json:"acked"`
    Silenced   int64            `json:"silenced"`
    BySeverity map[string]int64 `json:"by_severity"`
    Trend      []TrendPoint     `json:"trend"`
}

type TrendPoint struct {
    Time  string `json:"time"`
    Count int64  `json:"count"`
}

func GetAlertStats(db *gorm.DB) (*AlertStats, error)
```

**Optimization**:
- Single query with `GROUP BY status, severity` for counts
- Single query with `date_trunc('hour', trigger_time)` for trend

---

### 4. WebSocket Module (`internal/websocket`)

**Responsibility**: Robust WebSocket connection management

**Interface**:
```go
package websocket

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Client struct {
    hub     *Hub
    conn    *websocket.Conn
    send    chan []byte
    done    chan struct{}  // Lifecycle control
}

func NewHub() *Hub
func (h *Hub) Run()
func (h *Hub) Broadcast(message []byte)
func (c *Client) Close()  // Properly closes all goroutines
```

**Fix**: Use `done` channel to coordinate goroutine shutdown.

---

## Migration Strategy

### Phase 1: Create New Modules (Non-Breaking)
1. Create `internal/routing` with shared matching logic
2. Create `internal/template` with shared rendering logic
3. Create `internal/stats` with optimized queries
4. Create `internal/websocket` with proper lifecycle

### Phase 2: Migrate Consumers
1. Update `webhook.go` to use `routing.Matcher`
2. Update `delivery/service.go` to use `routing.Matcher`
3. Update `alert.go` to use `stats.GetAlertStats`
4. Replace existing WebSocket handler with new implementation

### Phase 3: Cleanup
1. Remove duplicate functions from `webhook.go`
2. Remove duplicate functions from `delivery/service.go`
3. Remove old WebSocket implementation

---

## API Compatibility

All changes are internal refactoring. API endpoints remain unchanged:
- `GET /api/v1/alerts/stats` - same response format
- `POST /api/v1/webhook/:source_name` - same behavior
- `GET /api/v1/ws/alerts` - same message format

---

## Testing Requirements

1. **Unit Tests**: Each new module must have >80% coverage
2. **Integration Tests**: Verify API responses unchanged
3. **Performance Tests**: Verify query count reduction

---

## Success Criteria

1. N+1 queries eliminated (verify with query logging)
2. Stats endpoint uses ≤3 queries (down from 32)
3. WebSocket clients properly disconnect without goroutine leaks
4. All existing tests pass
5. No duplicate route matching code
