# Testing Patterns

**Analysis Date:** 2026-04-10

## Test Framework

**Runner:**
- Backend tests use Go's built-in `testing` package plus `github.com/stretchr/testify/assert`, as shown in `internal/models/alert_test.go`, `internal/config/config_test.go`, `internal/router/router_test.go`, and `internal/handlers/webhook_test.go`.
- Config: no custom Go test config file detected; tests run through standard `go test` behavior from the module root defined by `go.mod`.
- Frontend: no test runner configuration detected. There is no `vitest.config.*`, `jest.config.*`, `playwright.config.*`, or `cypress.config.*` in the repository root or `frontend/`.

**Assertion Library:**
- `assert` from `github.com/stretchr/testify/assert` in committed Go tests.
- Frontend assertion library: not detected.

**Run Commands:**
```bash
go test ./...          # Run all backend tests from repository root
make test              # Wrapper around go test -v ./...
go test -cover ./...   # Coverage command pattern supported by Go, not wired into Makefile
pwsh -ExecutionPolicy Bypass -File scripts/verify_backend_no_ai.ps1
cd frontend && pnpm build
```

## Test File Organization

**Location:**
- Backend tests are colocated with production code. Current committed test files include `internal/models/alert_test.go`, `internal/config/config_test.go`, `internal/router/router_test.go`, and `internal/handlers/webhook_test.go`.
- Frontend has no colocated or separate test directories under `frontend/src/`.

**Naming:**
- Backend follows the standard Go `_test.go` suffix.
- Frontend naming pattern is not established because no `.test.ts`, `.test.tsx`, `.spec.ts`, or `.spec.tsx` files were found.

**Structure:**
```text
internal/models/alert.go
internal/models/alert_test.go
internal/config/config.go
internal/config/config_test.go
internal/router/router.go
internal/router/router_test.go
internal/handlers/webhook.go
internal/handlers/webhook_test.go
frontend/src/...       # no automated test files detected
```

## Test Structure

**Suite Organization:**
```go
func TestAlert_Validation(t *testing.T) {
	tests := []struct {
		name    string
		alert   Alert
		wantErr bool
	}{ ... }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.alert.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

**Patterns:**
- Use table-driven tests with a `tests := []struct{...}` slice and `t.Run(...)`, as in `internal/models/alert_test.go`.
- Prefer direct method calls on domain models or focused helper behavior instead of heavy test setup.
- Assertions are straightforward equality and error checks through `assert.Equal`, `assert.Error`, and `assert.NoError`.

## Mocking

**Framework:** No mocking framework detected.

**Patterns:**
```go
alert := Alert{Severity: tt.severity}
assert.Equal(t, tt.want, alert.IsValidSeverity())
```

**What to Mock:**
- No established mocking pattern exists in the codebase.
- For new backend tests, prefer exercising pure model logic and router behavior directly before introducing mocks.
- For handler, notifier, and database tests, local seams would need to be introduced first because current code stores concrete dependencies like `*gorm.DB` and HTTP senders.

**What NOT to Mock:**
- Domain validation methods in `internal/models/alert.go` and `internal/models/models.go` should be tested directly, matching the existing pattern in `internal/models/alert_test.go`.

## Fixtures and Factories

**Test Data:**
```go
alert: Alert{
	AlertID:     "test-alert-1",
	Source:      "prometheus",
	AlertName:   "HighMemory",
	Severity:    "P0",
	Message:     "Memory usage is high",
	Labels:      datatypes.JSON(`{"host": "server-01"}`),
	Fingerprint: "abc123",
	TriggerTime: time.Now(),
	ReceivedAt:  time.Now(),
	Status:      "pending",
}
```

**Location:**
- Inline fixtures live directly inside the Go test files.
- No shared fixture or factory package exists.

## Coverage

**Requirements:** None enforced.

**Current Coverage Shape:**
- Backend now includes committed regression coverage for config loading, router registration, webhook severity normalization, and alert model validation.
- Phase 1 verification added `scripts/verify_backend_no_ai.ps1`, which proves the retained backend flow without AI runtime: startup, login, webhook ingestion, notification dispatch, alert listing, stats, ack, quick silence, and `/api/v1/ai/chat=404`.
- Phase 2 verification used `cd frontend && pnpm build` as the frontend regression gate after removing AI pages, routes, contracts, and UI surfaces.
- The frontend still has no automated test coverage beyond build-time type and bundler checks. `frontend/package.json` contains `dev`, `build`, `preview`, `lint`, and `format` scripts, but no `test` script.

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Test Types

**Unit Tests:**
- Present for model validation helpers in `internal/models/alert_test.go`.
- Present for config loading in `internal/config/config_test.go`.
- Present for route registration in `internal/router/router_test.go`.
- Present for webhook severity normalization in `internal/handlers/webhook_test.go`.

**Integration Tests:**
- Scripted backend verification exists through `scripts/verify_backend_no_ai.ps1`.
- No database-backed Go integration suite is committed beyond that script-driven path.

**E2E Tests:**
- Not used.
- No Playwright, Cypress, or browser automation configuration was found for `frontend/`.

## Common Patterns

**Async Testing:**
```go
// No async-specific Go test pattern is established in committed tests.
// The current suite is mostly synchronous and deterministic.
```

**Error Testing:**
```go
err := tt.alert.Validate()
assert.Error(t, err)
```

## Current Gaps

**Backend Gaps:**
- `internal/handlers/alert.go`, `internal/handlers/config.go`, and `internal/handlers/user.go` still have no request or response tests for status codes, auth failures, or malformed payloads.
- `internal/handlers/webhook.go` still lacks coverage for template parsing, deduplication, fallback alert creation, routing, or notification dispatch beyond severity normalization.
- `internal/middleware/auth.go` has no tests for bearer token parsing, role enforcement, or context population.
- `internal/notifier/notifier.go` and `internal/database/*.go` have no tests around external service failures or configuration edge cases.

**Frontend Gaps:**
- `frontend/src/pages/*.tsx` has no component, interaction, or route-guard tests.
- `frontend/src/stores/*.ts` has no tests for loading-state transitions, optimistic updates, or `localStorage` persistence.
- `frontend/src/api/client.ts` and `frontend/src/api/auth.ts` have no tests for interceptors, 401 handling, or token propagation.
- `frontend/src/App.tsx` has no route protection tests for `RequireAuth` or login redirect behavior.

## Recommended Organization For New Tests

**Backend:**
- Continue colocated Go tests beside the package under test, for example `internal/middleware/auth_test.go`, `internal/handlers/user_test.go`, and `internal/notifier/notifier_test.go`.
- Follow the existing table-driven style used in `internal/models/alert_test.go`.
- Start with middleware and handler units, then add broader request tests using `httptest` against `internal/router/router.go` or isolated handlers.

**Frontend:**
- Introduce a dedicated runner in `frontend/` before adding tests.
- Once a runner exists, colocate tests beside pages, stores, and components using names like `Alerts.test.tsx`, `userStore.test.ts`, and `client.test.ts`.
- Prioritize `frontend/src/stores/alertStore.ts`, `frontend/src/App.tsx`, and form-heavy pages such as `frontend/src/pages/Login.tsx` and `frontend/src/pages/Alerts.tsx`.

---

*Testing analysis: 2026-04-10*
