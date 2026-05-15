---
name: user-management-navigation-and-audit-logs
description: Add sidebar navigation entry for user management (admin-only) and audit logs viewing with tab-based UI on Users page
---

# User Management Navigation & Audit Logs

## Problem

The project has a complete user management system (backend CRUD API, frontend Users.tsx page, JWT auth, RBAC), but:

1. **No navigation entry**: The sidebar has 9 menu items but no "用户管理" entry. Admins can only reach `/users` by typing the URL directly.
2. **No audit log viewing**: The backend writes `AuditLog` records for all admin operations (user create/disable/role_change/delete, config changes, alert ack/silence), but there is no API or UI to query and view these logs.

## Design

### 1. Sidebar Navigation Entry

Add a "用户管理" menu item to `AppSidebar.tsx`, visible only to admin-role users.

- **Menu item**: `{ key: '/users', icon: 'pi pi-users', label: '用户管理' }`
- **Conditional rendering**: Only show when `user.role === 'admin'`
- **Position**: Last item in the sidebar menu, after "值班管理"
- **Implementation**: `AppSidebar` needs access to the current user. Pass `user` from `AppLayout` → `AppSidebar` via props, or use `useUserStore` directly inside `AppSidebar`.

No changes to `AppHeader.tsx` user menu — sidebar entry is sufficient.

### 2. Users Page TabView Structure

Refactor `Users.tsx` to use PrimeReact `TabView` with two tabs:

**Tab 1: "用户列表"**
- Preserves all existing functionality (DataTable, create/edit/disable/force-password-reset/delete)
- No functional changes to the existing user management operations

**Tab 2: "审计日志"**
- DataTable showing audit log entries
- Columns: 时间、操作人(用户名+角色)、操作类型、目标类型、目标ID、结果、详情
- Result column uses PrimeReact `Tag`: `allowed` → success (green), `denied` → danger (red)
- Pagination: server-side, default 20 per page
- Default sort: created_at DESC (newest first)
- No client-side filtering (removed per user preference)

### 3. Backend Audit Log Query API

**Endpoint**: `GET /api/v1/users/audit-logs`

**Authorization**: `requireCapability(authz.CapabilityManageUsers)` — admin only

**Query parameters**:
- `page` — page number, default 1
- `page_size` — items per page, default 20, max 100
- `action` — filter by action type (optional, e.g. "user.create", "user.disable")
- `result` — filter by result (optional: "allowed" or "denied")
- `start_time` — filter start time (optional, RFC3339 format)
- `end_time` — filter end time (optional, RFC3339 format)

**Response**:
```json
{
  "items": [
    {
      "id": 1,
      "actor_user_id": 1,
      "actor_username": "admin",
      "actor_role": "admin",
      "action": "user.create",
      "target_type": "user",
      "target_id": "2",
      "result": "allowed",
      "detail": "role=operator",
      "created_at": "2026-05-15T10:30:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20
}
```

**Implementation**:
- Add `ListAuditLogs` handler method to `UserHandler` in `internal/handlers/user.go`
- Register route in `internal/router/router.go` under the `users` group with `RequireCapability(authz.CapabilityManageUsers)`
- Use GORM pagination: `db.Where(...).Order("created_at DESC").Count(&total).Offset(offset).Limit(limit).Find(&items)`

### 4. Frontend Audit Log API Client

Add audit log API methods to `frontend/src/api/auth.ts` (or a new `frontend/src/api/users.ts`):

```typescript
listAuditLogs: async (params: {
  page?: number
  page_size?: number
  action?: string
  result?: string
  start_time?: string
  end_time?: string
}): Promise<{ items: AuditLog[]; total: number; page: number; page_size: number }>
```

### 5. Frontend AuditLog Type

Add `AuditLog` type to `frontend/src/types/index.ts`:

```typescript
export interface AuditLog {
  id: number
  actor_user_id: number
  actor_username: string
  actor_role: string
  action: string
  target_type: string
  target_id: string
  result: string
  detail: string
  created_at: string
}
```

## Scope

- **In scope**: Sidebar nav entry, Users page TabView refactor, audit log query API, audit log UI tab
- **Out of scope**: Password complexity policy, email verification, SSO/LDAP, session management, user search/filtering on user list, separate audit log page

## Files to Modify

| File | Change |
|------|--------|
| `internal/handlers/user.go` | Add `ListAuditLogs` handler |
| `internal/router/router.go` | Register `GET /users/audit-logs` route |
| `frontend/src/components/layout/AppSidebar.tsx` | Add conditional "用户管理" menu item |
| `frontend/src/pages/Users.tsx` | Refactor to TabView with "用户列表" and "审计日志" tabs |
| `frontend/src/api/auth.ts` | Add `listAuditLogs` API method |
| `frontend/src/types/index.ts` | Add `AuditLog` type |

## Files to Create

None — all changes are modifications to existing files.