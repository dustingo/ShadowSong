# Testing Patterns

**Analysis Date:** 2026-04-09

## Test Framework

**Runner:**
- Backend tests use Go's built-in `testing` package plus `github.com/stretchr/testify/assert`, as shown in `internal/models/alert_test.go`.
- Config: no custom Go test config file detected; tests run through standard `go test` behavior from the module root defined by `go.mod`.
- Frontend: no test runner configuration detected. There is no `vitest.config.*`, `jest.config.*`, `playwright.config.*`, or `cypress.config.*` in the repository root or `frontend/`.

**Assertion Library:**
- `assert` from `github.com/stretchr/testify/assert` in `internal/models/alert_test.go`.
- Frontend assertion library: not detected.

**Run Commands:**
```bash
go test ./...          # Run all backend tests from repository root
make test              # Wrapper around go test -v ./...
go test -cover ./...   # Coverage command pattern supported by Go, not wired into Makefile
```

## Test File Organization

**Location:**
- Backend tests are colocated with production code. The only committed test file is `internal/models/alert_test.go`.
- Frontend has no colocated or separate test directories under `frontend/src/`.

**Naming:**
- Backend follows the standard Go `_test.go` suffix, for example `internal/models/alert_test.go`.
- Frontend naming pattern is not established because no `.test.ts`, `.test.tsx`, `.spec.ts`, or `.spec.tsx` files were found.

**Structure:**
```text
internal/models/alert.go
internal/models/alert_test.go
frontend/src/...       # no test files detected
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
- Prefer direct method calls on domain models instead of heavy test setup; the existing suite instantiates `Alert` structs inline and calls `Validate`, `IsValidSeverity`, and `IsValidStatus`.
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
- For new backend tests, prefer exercising pure model logic directly before introducing mocks.
- For handler, notifier, AI client, and database tests, local seams would need to be introduced first because current code stores concrete dependencies like `*gorm.DB`, `*ai.Client`, and HTTP clients in `internal/handlers/*.go` and `internal/notifier/notifier.go`.

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
- Inline fixtures live directly inside `internal/models/alert_test.go`.
- No shared fixture or factory package exists.

## Coverage

**Requirements:** None enforced.

**Current Coverage Shape:**
- `go test ./...` passes on 2026-04-09.
- Only `internal/models` is covered by committed tests; `cmd/server`, `internal/ai`, `internal/auth`, `internal/config`, `internal/database`, `internal/handlers`, `internal/middleware`, `internal/notifier`, and `internal/router` reported `[no test files]` when `go test ./...` ran.
- The frontend has no automated test coverage at all. `frontend/package.json` contains `dev`, `build`, `preview`, `lint`, and `format` scripts, but no `test` script.

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Test Types

**Unit Tests:**
- Present only for model validation helpers in `internal/models/alert_test.go`.
- Current unit testing style favors pure logic over integration with Gin, Gorm, Redis, or external APIs.

**Integration Tests:**
- Not detected.
- No database-backed handler tests, router tests, or end-to-end webhook ingestion tests are committed for `internal/handlers/webhook.go`, `internal/handlers/config.go`, or `internal/router/router.go`.

**E2E Tests:**
- Not used.
- No Playwright, Cypress, or browser automation configuration was found for `frontend/`.

## Common Patterns

**Async Testing:**
```go
// No async-specific pattern is established in committed tests.
// Existing tests are synchronous and deterministic.
```

**Error Testing:**
```go
err := tt.alert.Validate()
assert.Error(t, err)
```

## Current Gaps

**Backend Gaps:**
- `internal/handlers/alert.go`, `internal/handlers/config.go`, `internal/handlers/user.go`, and `internal/handlers/ai.go` have no request/response tests for status codes, auth failures, or malformed payloads.
- `internal/handlers/webhook.go` has no coverage for template parsing, deduplication, fallback alert creation, routing, or notification dispatch, despite being the most behavior-dense backend file.
- `internal/middleware/auth.go` has no tests for bearer token parsing, role enforcement, or context population.
- `internal/ai/client.go`, `internal/notifier/notifier.go`, and `internal/database/*.go` have no tests around external service failures or configuration edge cases.

**Frontend Gaps:**
- `frontend/src/pages/*.tsx` has no component, interaction, or route-guard tests.
- `frontend/src/stores/*.ts` has no tests for loading-state transitions, optimistic updates, or localStorage persistence.
- `frontend/src/api/client.ts` and `frontend/src/api/auth.ts` have no tests for interceptors, 401 handling, or token propagation.
- `frontend/src/App.tsx` has no route protection tests for `RequireAuth` or login redirect behavior.

## Recommended Organization For New Tests

**Backend:**
- Continue colocated Go tests beside the package under test, for example `internal/middleware/auth_test.go`, `internal/handlers/user_test.go`, and `internal/notifier/notifier_test.go`.
- Follow the existing table-driven style used in `internal/models/alert_test.go`.
- Start with model and middleware units, then add handler tests using `httptest` against `internal/router/router.go` or isolated handlers.

**Frontend:**
- Introduce a dedicated runner in `frontend/` before adding tests.
- Once a runner exists, colocate tests beside pages/stores/components using names like `Alerts.test.tsx`, `userStore.test.ts`, and `client.test.ts`.
- Prioritize `frontend/src/stores/alertStore.ts`, `frontend/src/App.tsx`, and form-heavy pages such as `frontend/src/pages/Login.tsx` and `frontend/src/pages/Alerts.tsx`.

---

*Testing analysis: 2026-04-09*
