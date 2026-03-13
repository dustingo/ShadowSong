# Coding Conventions

**Analysis Date:** 2026-03-13

## Project Overview

This project consists of two parts:
- **Frontend:** React 18 + TypeScript + Vite (located in `frontend/`)
- **Backend:** Go + Gin + GORM (located in root `internal/`)

---

## Frontend Conventions (TypeScript + React)

### Naming Patterns

**Files:**
- Components: PascalCase with `.tsx` extension (`AlertCard.tsx`, `SeverityBadge.tsx`)
- Utilities/Helpers: camelCase with `.ts` extension (`formatter.ts`)
- Types: `index.ts` in `types/` directory
- Stores: camelCase (`userStore.ts`, `alertStore.ts`)
- API clients: camelCase (`client.ts`, `auth.ts`)

**Directories:**
- Plural, lowercase: `components/`, `pages/`, `stores/`, `hooks/`, `utils/`, `api/`, `types/`

**Functions:**
- camelCase for all functions
- Event handlers: `handleXxx` (e.g., `handleLogout`, `handleSubmit`)
- API methods: action-based (e.g., `list`, `get`, `create`, `update`, `delete`)

**Variables:**
- camelCase
- Interfaces: PascalCase (`interface UserState`)
- Types: PascalCase

### Code Style

**Formatting:**
- Tool: Prettier 3.1.1
- Config (`.prettierrc`):
  ```json
  {
    "semi": false,
    "singleQuote": true,
    "tabWidth": 2,
    "trailingComma": "es5",
    "printWidth": 100,
    "arrowParens": "always"
  }
  ```

**Linting:**
- Tool: ESLint 8.56.0
- Config (`.eslintrc.cjs`):
  - Extends: `eslint:recommended`, `plugin:@typescript-eslint/recommended`, `plugin:react-hooks/recommended`
  - Parser: `@typescript-eslint/parser`
  - Key rules:
    - `react-refresh/only-export-components`: warn
    - `@typescript-eslint/no-explicit-any`: warn
    - `@typescript-eslint/no-unused-vars`: warn (argsIgnorePattern: `^_`)

**TypeScript:**
- Strict mode enabled (`tsconfig.json`)
- Target: ES2020
- Module: ESNext
- JSX: react-jsx
- No unused locals/parameters checks disabled

### Import Organization

**Order:**
1. React imports (`react`)
2. React Router imports (`react-router-dom`)
3. Third-party UI components (`antd`, `@ant-design/icons`)
4. Third-party libraries (`axios`, `dayjs`, `echarts`)
5. Local imports (paths, stores, api, types)

**Example:**
```typescript
import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Layout, Menu, Button } from 'antd'
import { DashboardOutlined, AlertOutlined } from '@ant-design/icons'
import axios from 'axios'
import { useUserStore } from './stores/userStore'
import { alertApi } from './api/client'
import { Alert } from './types'
import { Dashboard, Alerts } from './pages'
```

**Path Aliases:**
- `@/*` maps to `src/*` (configured in `tsconfig.json`)

### Component Patterns

**Function Components:**
```typescript
// Standard component structure
function ComponentName({ prop1, prop2 }: { prop1: string; prop2: number }) {
  // hooks
  const value = useSomeHook()

  // handlers
  const handleClick = () => { ... }

  return <JSX />
}
```

**Layout Components:**
- Located in `pages/` as separate route components
- Use Chinese labels for menu items (e.g., `告警大盘`, `告警管理`)

### Error Handling

**API Errors:**
- Handled via Axios interceptors in `src/api/client.ts`
- 401 responses trigger logout and redirect to `/login`
- Errors returned as rejected promises for callers to handle

**Component Error Boundaries:**
- Not implemented - should be added

### State Management

**Zustand Stores:**
- Use `create<UserState>()` pattern
- Persist to localStorage where needed
- Pattern in `src/stores/userStore.ts`:
  ```typescript
  export const useUserStore = create<UserState>((set) => ({
    user: initialState.user,
    token: initialState.token,
    setUser: (user) => { ... },
    setToken: (token) => { ... },
    logout: () => { ... },
  }))
  ```

---

## Backend Conventions (Go + Gin)

### Naming Patterns

**Files:**
- lowercase with underscores: `alert.go`, `alert_test.go`, `jwt.go`
- Package-scoped test files: `*_test.go`

**Functions:**
- Exported: PascalCase (`NewAlertHandler`, `List`)
- Unexported: camelCase (`handleRequest`)
- Constructor functions: `NewXxx()` pattern

**Variables:**
- camelCase
- Constants: PascalCase if exported, camelCase if unexported

### Package Structure

```
internal/
├── handlers/     # HTTP handlers
├── models/       # Data models and business logic
├── database/     # Database connections (postgres.go, redis.go)
├── auth/         # Authentication (jwt.go)
├── middleware/   # Gin middleware
├── config/       # Configuration
├── notifier/     # Notification logic
├── router/       # Route definitions
└── ai/           # AI client integrations
```

### Handler Pattern

```go
type AlertHandler struct {
    db *gorm.DB
}

func NewAlertHandler(db *gorm.DB) *AlertHandler {
    return &AlertHandler{db: db}
}

func (h *AlertHandler) List(c *gin.Context) {
    // 1. Parse query params
    // 2. Build query
    // 3. Execute query
    // 4. Return JSON response
}
```

### Error Handling

**HTTP Errors:**
- Return JSON with error message and appropriate HTTP status code
- Pattern:
  ```go
  c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
  c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
  c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
  ```

**Validation:**
- Model-level validation methods (`Validate()`, `IsValidSeverity()`)
- Input binding with `ShouldBindJSON()`

### Database

**ORM:**
- GORM for database operations
- Models defined in `internal/models/`
- Hooks: `BeforeCreate`, `BeforeUpdate`

---

## Shared Conventions

### API Design (Frontend-Backend Contract)

**RESTful Endpoints:**
- List: `GET /resource`
- Get: `GET /resource/:id`
- Create: `POST /resource`
- Update: `PUT /resource/:id`
- Delete: `DELETE /resource/:id`
- Custom actions: `POST /resource/:id/action`

**Response Format:**
```typescript
// List response
{ list: T[], total: number }

// Single resource
T

// Errors
{ error: string }
```

**Authentication:**
- Bearer token in Authorization header
- Token stored in localStorage on frontend

### Git/Version Control

- Commit messages: descriptive, imperative mood
- Branch naming: feature/description, bugfix/description

---

*Convention analysis: 2026-03-13*
