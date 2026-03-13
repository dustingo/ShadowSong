# Codebase Concerns

**Analysis Date:** 2026-03-13

## Tech Debt

**AI Recommendation Generation:**
- Issue: `GenerateRecommendations` handler is a stub that only returns a success message
- Files: `internal/handlers/ai.go:232`
- Impact: AI-based silence recommendations cannot be automatically generated
- Fix approach: Implement the recommendation generation logic using AI to analyze recent alerts

**Ignored Error Handling in Notifier:**
- Issue: JSON marshal errors are silently ignored in DingTalk and WeCom senders
- Files: `internal/notifier/notifier.go:152`, `internal/notifier/notifier.go:202`
- Impact: Notification content may be empty or malformed without any indication
- Fix approach: Handle marshal errors explicitly and return meaningful errors

**Missing Database Indexes:**
- Issue: No explicit indexes on frequently queried fields (status, severity, trigger_time, fingerprint)
- Files: `internal/models/models.go`
- Impact: Alert queries may become slow as data grows
- Fix approach: Add GORM indexes on Alert table for status, severity, trigger_time, fingerprint

**Incomplete AI Alert Analysis:**
- Issue: `AnalyzeAlert` function returns empty strings for root_cause and suggestions
- Files: `internal/ai/client.go:165`
- Impact: AI suggestions feature does not populate structured data
- Fix approach: Parse the JSON response from AI to extract structured analysis

## Known Bugs

**WebSocket Reconnection Race Condition:**
- Issue: Multiple WebSocket connection attempts can occur during rapid reconnect cycles
- Files: `frontend/src/pages/Dashboard.tsx:80-130`
- Trigger: When network briefly disconnects, multiple timers may fire
- Workaround: Current code uses `isConnecting` flag but race condition still possible

**Template Validation Returns 200 on Invalid JSON:**
- Issue: `TestInputTemplate` returns HTTP 200 even when validation errors exist
- Files: `internal/handlers/webhook.go:686-694`
- Trigger: When alert_name, severity, or message fields are missing
- Workaround: None - this is unexpected behavior

**Alert Stats Runs N+1 Queries:**
- Issue: Stats endpoint runs separate Count queries for each severity level
- Files: `internal/handlers/alert.go:173-199`
- Impact: Performance degrades with additional severity levels
- Fix approach: Use SQL GROUP BY aggregation

## Security Considerations

**Hardcoded CORS Origin:**
- Risk: CORS allows only `http://127.0.0.1` - not usable in production
- Files: `internal/router/router.go:19`
- Current mitigation: None - will break in production
- Recommendations: Make CORS origin configurable via environment variable

**API Key Stored in Plaintext:**
- Risk: API keys stored in database are compared as plaintext strings
- Files: `internal/handlers/webhook.go:81`
- Current mitigation: API key validation exists
- Recommendations: Consider hashing API keys or using a secrets manager

**No Rate Limiting:**
- Risk: Webhook endpoints have no rate limiting
- Files: `internal/router/router.go:161-165`
- Current mitigation: None
- Recommendations: Add rate limiting middleware for webhook endpoints

**No Input Sanitization on User Comments:**
- Risk: User-provided comments in alert acknowledgment are stored directly
- Files: `internal/handlers/alert.go:107`
- Current mitigation: None
- Recommendations: Sanitize user input to prevent XSS if displayed in notifications

## Performance Bottlenecks

**Dashboard Polling Every 10 Seconds:**
- Problem: Frontend polls for alerts and stats every 10 seconds regardless of WebSocket connection
- Files: `frontend/src/pages/Dashboard.tsx:135-138`
- Cause: Redundant polling when WebSocket is connected
- Improvement path: Disable polling when WebSocket is connected, or increase poll interval

**Notification Processing Runs in Background Without Queue:**
- Problem: `processAlertNotifications` runs in goroutine without proper job queue
- Files: `internal/handlers/webhook.go:184`
- Cause: If notification sending fails, no retry mechanism
- Improvement path: Use a job queue (Redis queue or dedicated worker) for reliable delivery

**Large Alert List Without Pagination on Active Endpoint:**
- Problem: `Active` endpoint returns all firing alerts without limit
- Files: `internal/handlers/alert.go:205-214`
- Cause: No pagination on active alerts endpoint
- Improvement path: Add optional limit parameter

## Fragile Areas

**Template Rendering is All-or-Nothing:**
- Why fragile: Template errors create fallback alerts that may not contain meaningful data
- Files: `internal/handlers/webhook.go:116-121`
- Safe modification: Test templates thoroughly before deployment
- Test coverage: Unit tests for template rendering needed

**Webhook Handler Has Many Responsibilities:**
- Why fragile: Single handler handles parsing, validation, deduplication, routing, and notification
- Files: `internal/handlers/webhook.go:41-204`
- Safe modification: Extract smaller functions for each responsibility
- Test coverage: Integration tests for webhook flow needed

**Manual JWT Secret Validation:**
- Why fragile: Application exits immediately if JWT_SECRET is missing with no graceful handling
- Files: `internal/config/config.go:51-55`
- Safe modification: Use proper error handling with logging
- Test coverage: Config validation tests needed

## Scaling Limits

**PostgreSQL Connection Pool:**
- Current capacity: Default GORM connection pool (uses database defaults)
- Limit: Not configured - may exhaust connections under load
- Scaling path: Configure `SetMaxOpenConns`, `SetMaxIdleConns` in database setup

**Redis as Message Bus:**
- Current capacity: Single Redis connection
- Limit: Single point of failure, limited throughput
- Scaling path: Use Redis Cluster for high availability

**In-Memory Alert Sorting:**
- Current capacity: All active alerts loaded into memory on frontend
- Limit: Will cause memory issues with thousands of active alerts
- Scaling path: Implement virtual scrolling or server-side pagination

## Dependencies at Risk

**golang-jwt/jwt v5:**
- Risk: Older JWT library - ensure latest patches applied
- Impact: Token validation depends on this
- Migration plan: Consider migrate to `github.com/golang-jwt/jwt/v5` for better support

**echarts and echarts-for-react:**
- Risk: Large bundle size (echarts is ~600KB minified)
- Impact: Frontend load time affected
- Migration plan: Consider lighter alternatives or tree-shaking

**No Dedicated Testing Framework:**
- Risk: Only basic test file exists (`internal/models/alert_test.go`)
- Impact: No regression detection
- Migration plan: Add comprehensive unit and integration tests

## Missing Critical Features

**No Alert Resolution Tracking:**
- Problem: System receives alerts but has no mechanism to track when alert conditions clear
- Blocks: Cannot auto-resolve alerts or show resolved state

**No User Audit Log:**
- Problem: No tracking of who acknowledged alerts or made changes
- Blocks: Compliance requirements, incident post-mortems

**No Backup/Restore:**
- Problem: No database backup mechanism
- Blocks: Disaster recovery

## Test Coverage Gaps

**Backend Handlers:**
- What's not tested: Handler HTTP endpoints (no HTTP-level tests)
- Files: `internal/handlers/*.go`
- Risk: Breaking changes go undetected
- Priority: High

**Frontend API Client:**
- What's not tested: API client error handling, response parsing
- Files: `frontend/src/api/client.ts`
- Risk: API changes break frontend silently
- Priority: Medium

**Webhook Processing:**
- What's not tested: End-to-end webhook flow with templates
- Files: `internal/handlers/webhook.go`
- Risk: Template errors in production
- Priority: High

**Authentication Flow:**
- What's not tested: Login, token refresh, token expiry handling
- Files: `internal/auth/`, `frontend/src/api/auth.ts`
- Risk: Auth edge cases undetected
- Priority: Medium

---

*Concerns audit: 2026-03-13*
