# Testing Patterns

**Analysis Date:** 2026-03-13

## Test Framework

### Backend (Go)

**Test Runner:**
- Framework: Go's built-in `testing` package
- Assertion Library: `github.com/stretchr/testify` v1.11.1
- Config: Standard Go test configuration

**Run Commands:**
```bash
go test ./...                 # Run all tests
go test -v ./internal/models  # Run with verbose output
go test -cover ./...          # Run with coverage
```

### Frontend (React + TypeScript)

**Status: NO TEST FRAMEWORK INSTALLED**

The frontend currently has no testing infrastructure:
- No Jest, Vitest, or other test runners
- No React Testing Library
- No test scripts in `package.json`

**To add testing, install:**
```bash
pnpm add -D vitest @testing-library/react @testing-library/jest-dom jsdom
# or
npm install --save-dev vitest @testing-library/react @testing-library/jest-dom jsdom
```

---

## Test File Organization

### Backend

**Location:**
- Co-located with source files in same package
- Test file naming: `*_test.go`

**Structure:**
```
internal/
├── models/
│   ├── alert.go
│   └── alert_test.go    # Tests in same package
├── handlers/
│   ├── alert.go
│   └── alert_test.go    # Optional integration tests
```

### Frontend

**Recommended Structure (not currently implemented):**
```
src/
├── components/
│   ├── AlertCard.tsx
│   └── AlertCard.test.tsx   # or AlertCard.spec.tsx
├── pages/
│   ├── Dashboard.tsx
│   └── Dashboard.test.tsx
└── __tests__/               # Optional centralized tests
```

---

## Backend Test Patterns

### Test File Structure

```go
package models

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "gorm.io/datatypes"
)

func TestAlert_Validation(t *testing.T) {
    tests := []struct {
        name    string
        alert   Alert
        wantErr bool
    }{
        // test cases
    }

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

### Test Organization

**Pattern: Table-Driven Tests**
- Use `tests := []struct{...}{}` for test cases
- Use `t.Run()` for subtests
- Clear test case names in Chinese or English

### Example: Model Validation Tests

From `internal/models/alert_test.go`:

```go
func TestAlert_IsValidSeverity(t *testing.T) {
    tests := []struct {
        severity string
        want     bool
    }{
        {"P0", true},
        {"P1", true},
        {"P2", true},
        {"P3", true},
        {"INVALID", false},
        {"", false},
    }

    for _, tt := range tests {
        t.Run(tt.severity, func(t *testing.T) {
            alert := Alert{Severity: tt.severity}
            assert.Equal(t, tt.want, alert.IsValidSeverity())
        })
    }
}
```

### Assertion Patterns

**Using testify/assert:**
```go
import "github.com/stretchr/testify/assert"

// Common assertions
assert.NoError(t, err)
assert.Error(t, err)
assert.Equal(t, expected, actual)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.True(t, condition)
assert.False(t, condition)
assert.Contains(t, string, substring)
```

### Mocking

**Database Mocking:**
- GORM's `gorm.DB` can be mocked using custom GORM callbacks
- For unit tests, consider using `github.com/DATA-DOG/go-sqlmock`

**Example Pattern:**
```go
// Create a mock DB for testing handlers
func setupTestDB(t *testing.T) *gorm.DB {
    // Use in-memory SQLite or mock
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    assert.NoError(t, err)
    return db
}
```

---

## Frontend Test Patterns (Recommended)

### Test Framework Setup

**Recommended: Vitest + React Testing Library**

Install:
```bash
pnpm add -D vitest @testing-library/react @testing-library/jest-dom jsdom @types/jest
```

Configure `vite.config.ts`:
```typescript
/// <reference types="vitest" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
})
```

Create `src/test/setup.ts`:
```typescript
import '@testing-library/jest-dom'
```

### Test Structure

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'

describe('AlertCard', () => {
  it('renders alert information correctly', () => {
    const alert = {
      alert_id: 'test-1',
      alert_name: 'HighMemory',
      severity: 'P0' as const,
      message: 'Memory usage is high',
      status: 'firing' as const,
    }

    render(<AlertCard alert={alert} />)
    expect(screen.getByText('HighMemory')).toBeInTheDocument()
    expect(screen.getByText('P0')).toBeInTheDocument()
  })

  it('calls onAck when acknowledge button is clicked', async () => {
    const onAck = vi.fn()
    const alert = { /* ... */ }

    render(<AlertCard alert={alert} onAck={onAck} />)
    fireEvent.click(screen.getByText('Acknowledge'))

    await waitFor(() => {
      expect(onAck).toHaveBeenCalled()
    })
  })
})
```

### Mocking Patterns

**API Mocking:**
```typescript
import { vi } from 'vitest'

// Mock API module
vi.mock('./api/client', () => ({
  alertApi: {
    list: vi.fn().mockResolvedValue({ list: [], total: 0 }),
    ack: vi.fn().mockResolvedValue({}),
  },
}))
```

**Store Mocking:**
```typescript
import { useUserStore } from './stores/userStore'

// Reset store before each test
beforeEach(() => {
  useUserStore.setState({ user: null, token: null })
})
```

**Component Mocking:**
```typescript
// Mock child components
vi.mock('./components/SeverityBadge', () => ({
  default: ({ severity }: { severity: string }) => (
    <span data-testid="severity">{severity}</span>
  ),
}))
```

### Testing Hooks

```typescript
import { renderHook, act } from '@testing-library/react'
import { useUserStore } from './stores/userStore'

describe('useUserStore', () => {
  it('sets user correctly', () => {
    const { result } = renderHook(() => useUserStore())

    act(() => {
      result.current.setUser({ id: 1, username: 'test', name: 'Test', role: 'admin' })
    })

    expect(result.current.user?.username).toBe('test')
  })

  it('clears state on logout', () => {
    // Setup state first
    // ...
    act(() => {
      result.current.logout()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.token).toBeNull()
  })
})
```

---

## Coverage

### Backend

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Coverage Target:** Not currently enforced - recommend 70%+ for new code

### Frontend

**View Coverage (with Vitest):**
```bash
vitest --coverage
```

**Recommended Coverage:**
- Components: 70%+
- Utils/Helpers: 90%+
- Critical paths: 100%

---

## Test Types

### Unit Tests

**Backend:**
- Model validation (`alert_test.go`)
- Utility functions
- Business logic

**Frontend (recommended):**
- Component rendering
- Hook behavior
- Utility functions

### Integration Tests

**Backend:**
- Handler tests with real database
- API endpoint tests
- Database transactions

**Frontend (recommended):**
- Page-level component tests
- Store integration tests
- API client integration

### E2E Tests

**Not implemented.** Recommended to add:
- Playwright or Cypress for end-to-end testing
- Critical user flows (login, view alerts, acknowledge)

---

## Common Patterns

### Async Testing (Backend)

```go
func TestAsyncOperation(t *testing.T) {
    done := make(chan bool)

    go func() {
        // async operation
        done <- true
    }()

    select {
    case <-done:
        // success
    case <-time.After(time.Second):
        t.Fatal("timeout")
    }
}
```

### Async Testing (Frontend - Vitest)

```typescript
import { act } from '@testing-library/react'

it('fetches alerts async', async () => {
  // For API calls mocked with vi.fn()
  const promise = Promise.resolve({ list: alerts, total: 1 })

  render(<AlertsPage />)

  await act(async () => {
    await promise
  })

  expect(screen.getByText('Alert 1')).toBeInTheDocument()
})
```

### Error Testing

**Backend:**
```go
func TestAlert_Validate_MissingFields(t *testing.T) {
    alert := Alert{}
    err := alert.Validate()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "alert_id is required")
}
```

**Frontend:**
```typescript
it('displays error message on API failure', async () => {
  apiClient.get.mockRejectedValue(new Error('Network error'))

  render(<Dashboard />)

  await waitFor(() => {
    expect(screen.getByText('Failed to load')).toBeInTheDocument()
  })
})
```

---

## Missing Testing Infrastructure

### Frontend Gaps

1. **No test framework** - Vitest recommended
2. **No test utilities** - Need @testing-library/react
3. **No test scripts** - Add to package.json
4. **No E2E tests** - Recommend Playwright
5. **No CI integration** - Add test step to pipeline

### Recommended package.json Scripts

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage",
    "test:run": "vitest run"
  }
}
```

---

*Testing analysis: 2026-03-13*
