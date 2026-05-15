# User Management Navigation & Audit Logs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add sidebar navigation entry for user management (admin-only) and audit logs viewing with tab-based UI on Users page.

**Architecture:** Backend adds a paginated audit log query endpoint under the existing users route group. Frontend refactors Users.tsx into a TabView with "用户列表" and "审计日志" tabs, and adds a conditional sidebar menu item visible only to admins.

**Tech Stack:** Go (Gin + GORM), React (PrimeReact TabView + DataTable), TypeScript, Zustand

---

### Task 1: Backend — Add ListAuditLogs Handler

**Files:**
- Modify: `internal/handlers/user.go:510-511` (append new handler after `recordAudit`)
- Test: `internal/handlers/user_test.go:504` (append new test)

- [ ] **Step 1: Write the failing test for ListAuditLogs**

Append to `internal/handlers/user_test.go`:

```go
func TestListAuditLogsReturnsPaginatedResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	// Seed some audit logs
	now := time.Now()
	for i := 0; i < 5; i++ {
		require.NoError(t, db.Create(&models.AuditLog{
			ActorUserID:   1,
			ActorUsername: "admin",
			ActorRole:     authz.RoleAdmin,
			Action:        "user.create",
			TargetType:    "user",
			TargetID:      fmt.Sprintf("%d", i+2),
			Result:        auditResultAllowed,
			Detail:        fmt.Sprintf("role=viewer"),
		}).Error)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/audit-logs?page=1&page_size=3", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.ListAuditLogs(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float64(5), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(3), response["page_size"])

	items := response["items"].([]any)
	assert.Len(t, items, 3)
}

func TestListAuditLogsFiltersByAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.create", TargetType: "user", TargetID: "2", Result: auditResultAllowed,
	}).Error)
	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.disable", TargetType: "user", TargetID: "2", Result: auditResultAllowed,
	}).Error)

	req := httptest.NewRequest(http.MethodGet, "/users/audit-logs?action=user.create", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.ListAuditLogs(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float64(1), response["total"])
}

func TestListAuditLogsFiltersByResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := openUserTestDB(t)
	handler := NewUserHandler(db, newUserTestJWT())

	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.create", TargetType: "user", TargetID: "2", Result: auditResultAllowed,
	}).Error)
	require.NoError(t, db.Create(&models.AuditLog{
		ActorUserID: 1, ActorUsername: "admin", ActorRole: authz.RoleAdmin,
		Action: "user.disable", TargetType: "user", TargetID: "2", Result: auditResultDenied,
	}).Error)

	req := httptest.NewRequest(http.MethodGet, "/users/audit-logs?result=denied", nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = req

	handler.ListAuditLogs(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float64(1), response["total"])
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd d:/goproject/shadowsongAI && go test ./internal/handlers/ -run "TestListAuditLogs" -v`
Expected: FAIL — `UserHandler.ListAuditLogs` is undefined

- [ ] **Step 3: Write the ListAuditLogs handler**

Append to `internal/handlers/user.go` after the `recordAudit` function:

```go
type auditLogResponse struct {
	ID            uint      `json:"id"`
	ActorUserID   uint      `json:"actor_user_id"`
	ActorUsername string    `json:"actor_username"`
	ActorRole     string    `json:"actor_role"`
	Action        string    `json:"action"`
	TargetType    string    `json:"target_type"`
	TargetID      string    `json:"target_id"`
	Result        string    `json:"result"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"created_at"`
}

type listAuditLogsResponse struct {
	Items    []auditLogResponse `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// ListAuditLogs returns paginated audit log entries (admin only)
func (h *UserHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.AuditLog{})

	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}
	if result := c.Query("result"); result != "" {
		query = query.Where("result = ?", result)
	}
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var logs []models.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]auditLogResponse, len(logs))
	for i, log := range logs {
		items[i] = auditLogResponse{
			ID:            log.ID,
			ActorUserID:   log.ActorUserID,
			ActorUsername: log.ActorUsername,
			ActorRole:     log.ActorRole,
			Action:        log.Action,
			TargetType:    log.TargetType,
			TargetID:      log.TargetID,
			Result:        log.Result,
			Detail:        log.Detail,
			CreatedAt:     log.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, listAuditLogsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd d:/goproject/shadowsongAI && go test ./internal/handlers/ -run "TestListAuditLogs" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/user.go internal/handlers/user_test.go
git commit -m "feat: add ListAuditLogs handler with pagination and filtering"
```

---

### Task 2: Backend — Register Audit Logs Route

**Files:**
- Modify: `internal/router/router.go:84` (add route after existing users routes)
- Test: `internal/router/router_test.go` (verify route exists)

- [ ] **Step 1: Add the route**

In `internal/router/router.go`, after the existing `users.DELETE("/:id", ...)` line (line 84), add:

```go
				users.GET("/audit-logs", middleware.RequireCapability(authz.CapabilityManageUsers), userHandler.ListAuditLogs)
```

- [ ] **Step 2: Run existing router tests to verify nothing broke**

Run: `cd d:/goproject/shadowsongAI && go test ./internal/router/ -v -count=1`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/router/router.go
git commit -m "feat: register GET /users/audit-logs route with manage_users capability"
```

---

### Task 3: Frontend — Add AuditLog Type and API Method

**Files:**
- Modify: `frontend/src/types/index.ts:248` (append AuditLog interface)
- Modify: `frontend/src/api/auth.ts:111` (append listAuditLogs method)

- [ ] **Step 1: Add AuditLog type**

Append to `frontend/src/types/index.ts` after the `DeliveryRecoveryResult` interface:

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

export interface AuditLogListResponse {
  items: AuditLog[]
  total: number
  page: number
  page_size: number
}
```

- [ ] **Step 2: Add listAuditLogs API method**

Append to `frontend/src/api/auth.ts` inside the `authApi` object, before the closing `}`:

```typescript
  listAuditLogs: async (params: {
    page?: number
    page_size?: number
    action?: string
    result?: string
    start_time?: string
    end_time?: string
  }): Promise<{ items: AuditLog[]; total: number; page: number; page_size: number }> => {
    const res = await authClient.get<{ items: AuditLog[]; total: number; page: number; page_size: number }>('/users/audit-logs', { params })
    return res.data
  },
```

Also add the import at the top of `auth.ts`:

```typescript
import type { User, AuditLog } from '../types'
```

And update the existing import line from `import type { User } from '../types'` to include `AuditLog`.

- [ ] **Step 3: Verify TypeScript compiles**

Run: `cd d:/goproject/shadowsongAI/frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/index.ts frontend/src/api/auth.ts
git commit -m "feat: add AuditLog type and listAuditLogs API method"
```

---

### Task 4: Frontend — Add Sidebar Navigation Entry

**Files:**
- Modify: `frontend/src/components/layout/AppSidebar.tsx`

- [ ] **Step 1: Add conditional user management menu item**

In `AppSidebar.tsx`:

1. Add import for `useUserStore`:
```typescript
import { useUserStore } from '../../stores/userStore'
import { isAdmin } from '../../authz/capabilities'
```

2. Inside the `AppSidebar` component, before the `return`, add:
```typescript
  const user = useUserStore((state) => state.user)
```

3. In the JSX, change the menu rendering. Replace the static `menuItems.map(...)` with a filtered approach. After the existing `menuItems` array definition (line 19), add a new item and filter logic:

Replace the `menuItems` constant with:
```typescript
const baseMenuItems: MenuItem[] = [
  { key: '/', icon: 'pi pi-home', label: '告警大盘' },
  { key: '/alerts', icon: 'pi pi-bell', label: '告警管理' },
  { key: '/deliveries', icon: 'pi pi-send', label: '通知投递' },
  { key: '/ops-health', icon: 'pi pi-chart-bar', label: '运维健康' },
  { key: '/datasources', icon: 'pi pi-database', label: '数据源' },
  { key: '/channels', icon: 'pi pi-telegram', label: '推送渠道' },
  { key: '/routes', icon: 'pi pi-sitemap', label: '路由规则' },
  { key: '/silences', icon: 'pi pi-volume-off', label: '静默管理' },
  { key: '/onduty', icon: 'pi pi-calendar', label: '值班管理' },
]

const userManagementItem: MenuItem = { key: '/users', icon: 'pi pi-users', label: '用户管理' }
```

4. Inside the component, compute the visible menu items:
```typescript
  const menuItems = isAdmin(user)
    ? [...baseMenuItems, userManagementItem]
    : baseMenuItems
```

5. The existing `menuItems.map(...)` in the JSX continues to work unchanged since the variable name is the same.

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd d:/goproject/shadowsongAI/frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/layout/AppSidebar.tsx
git commit -m "feat: add admin-only user management sidebar menu item"
```

---

### Task 5: Frontend — Refactor Users Page with TabView and Audit Logs Tab

**Files:**
- Modify: `frontend/src/pages/Users.tsx`

- [ ] **Step 1: Refactor Users.tsx to TabView with audit logs tab**

Replace the entire content of `Users.tsx` with:

```tsx
import React, { useCallback, useEffect, useState } from 'react'
import { Button } from 'primereact/button'
import { Card } from 'primereact/card'
import { Dialog } from 'primereact/dialog'
import { confirmDialog } from 'primereact/confirmdialog'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { InputText } from 'primereact/inputtext'
import { Password } from 'primereact/password'
import { Dropdown } from 'primereact/dropdown'
import { InputSwitch } from 'primereact/inputswitch'
import { Tag } from 'primereact/tag'
import { TabView, TabPanel } from 'primereact/tabview'
import { Paginator } from 'primereact/paginator'
import { authApi, getApiErrorMessage } from '../api/auth'
import { PermissionNotice, useToast } from '../components'
import { canUser, capabilityManageUsers } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import type { User, AuditLog } from '../types'

type UserFormValues = {
  username: string
  password: string
  name: string
  email: string
  role: User['role']
  force_password_reset: boolean
}

const roleOptions: Array<{ value: User['role']; label: string }> = [
  { value: 'admin', label: '管理员' },
  { value: 'operator', label: '值班操作员' },
  { value: 'viewer', label: '只读用户' },
]

const getRoleSeverity = (role: User['role']): 'danger' | 'info' | 'secondary' => {
  switch (role) {
    case 'admin':
      return 'danger'
    case 'operator':
      return 'info'
    default:
      return 'secondary'
  }
}

const actionLabelMap: Record<string, string> = {
  'user.create': '创建用户',
  'user.update_profile': '更新资料',
  'user.role_change': '角色变更',
  'user.disable': '禁用用户',
  'user.enable': '启用用户',
  'user.force_password_reset': '强制改密',
  'user.clear_force_password_reset': '取消强制改密',
  'user.password_change': '修改密码',
  'user.delete': '删除用户',
  'alert.ack': '确认告警',
  'alert.quick_silence': '快速静默',
  'config.datasource.create': '创建数据源',
  'config.datasource.update': '更新数据源',
  'config.datasource.delete': '删除数据源',
  'config.datasource.toggle': '切换数据源',
  'config.channel.create': '创建渠道',
  'config.channel.update': '更新渠道',
  'config.channel.delete': '删除渠道',
  'config.channel.toggle': '切换渠道',
  'config.channel.test': '测试渠道',
  'config.route.create': '创建路由',
  'config.route.update': '更新路由',
  'config.route.delete': '删除路由',
  'config.route.reorder': '重排路由',
  'config.silence.create': '创建静默',
  'config.silence.update': '更新静默',
  'config.silence.delete': '删除静默',
  'config.silence.create_from_alert': '从告警创建静默',
  'config.onduty.create': '创建值班',
  'config.onduty.update': '更新值班',
  'config.onduty.delete': '删除值班',
  'delivery.retry': '重试投递',
  'delivery.replay': '重放投递',
}

const targetTypeLabelMap: Record<string, string> = {
  user: '用户',
  alert: '告警',
  datasource: '数据源',
  channel: '渠道',
  route_rule: '路由规则',
  silence_rule: '静默规则',
  onduty: '值班',
  delivery_recovery: '投递恢复',
}

export const Users: React.FC = () => {
  const currentUser = useUserStore((state) => state.user)
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [formValues, setFormValues] = useState<UserFormValues>({
    username: '',
    password: '',
    name: '',
    email: '',
    role: 'viewer',
    force_password_reset: true,
  })
  const [formErrors, setFormErrors] = useState<Partial<Record<keyof UserFormValues, string>>>({})
  const canManageUsers = canUser(currentUser, capabilityManageUsers)
  const toast = useToast()

  // Audit logs state
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [auditLoading, setAuditLoading] = useState(false)
  const [auditTotal, setAuditTotal] = useState(0)
  const [auditPage, setAuditPage] = useState(0)
  const [auditPageSize] = useState(20)

  const fetchUsers = useCallback(async () => {
    if (!canManageUsers) {
      return
    }
    setLoading(true)
    try {
      const nextUsers = await authApi.listUsers()
      setUsers(nextUsers)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '加载用户列表失败'))
    } finally {
      setLoading(false)
    }
  }, [canManageUsers, toast])

  const fetchAuditLogs = useCallback(async (page: number) => {
    if (!canManageUsers) {
      return
    }
    setAuditLoading(true)
    try {
      const res = await authApi.listAuditLogs({ page: page + 1, page_size: auditPageSize })
      setAuditLogs(res.items)
      setAuditTotal(res.total)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '加载审计日志失败'))
    } finally {
      setAuditLoading(false)
    }
  }, [canManageUsers, auditPageSize, toast])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  useEffect(() => {
    fetchAuditLogs(auditPage)
  }, [fetchAuditLogs, auditPage])

  if (!canManageUsers) {
    return (
      <PermissionNotice
        title="当前角色无权访问用户管理"
        description="用户管理仅对具备用户管理能力的角色开放。"
        type="error"
      />
    )
  }

  const handleCreate = () => {
    setEditingUser(null)
    setFormValues({
      username: '',
      password: '',
      name: '',
      email: '',
      role: 'viewer',
      force_password_reset: true,
    })
    setFormErrors({})
    setModalVisible(true)
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    setFormValues({
      username: user.username,
      password: '',
      name: user.name,
      email: user.email || '',
      role: user.role,
      force_password_reset: user.force_password_reset ?? false,
    })
    setFormErrors({})
    setModalVisible(true)
  }

  const validateForm = (): boolean => {
    const errors: Partial<Record<keyof UserFormValues, string>> = {}

    if (!editingUser) {
      if (!formValues.username.trim()) {
        errors.username = '请输入用户名'
      }
      if (!formValues.password.trim()) {
        errors.password = '请输入初始密码'
      }
    }

    if (!formValues.name.trim()) {
      errors.name = '请输入姓名'
    }

    if (!formValues.role) {
      errors.role = '请选择角色'
    }

    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async () => {
    if (!validateForm()) {
      return
    }

    try {
      if (editingUser) {
        await authApi.updateUser(editingUser.id, {
          name: formValues.name,
          email: formValues.email,
          role: formValues.role,
          force_password_reset: formValues.force_password_reset,
        })
        toast.showSuccess('用户已更新')
      } else {
        await authApi.createUser({
          username: formValues.username,
          password: formValues.password,
          name: formValues.name,
          email: formValues.email,
          role: formValues.role,
          force_password_reset: formValues.force_password_reset,
        })
        toast.showSuccess('用户已创建')
      }
      setModalVisible(false)
      setEditingUser(null)
      fetchUsers()
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '保存用户失败'))
    }
  }

  const handleToggleDisabled = async (user: User) => {
    try {
      await authApi.updateUser(user.id, { disabled: !user.disabled_at })
      toast.showSuccess(user.disabled_at ? '账号已启用' : '账号已禁用')
      fetchUsers()
    } catch (error: unknown) {
      toast.showError(getApiErrorMessage(error, '更新账号状态失败'))
    }
  }

  const handleForcePasswordReset = async (user: User) => {
    try {
      await authApi.updateUser(user.id, { force_password_reset: true })
      toast.showSuccess('已标记为强制改密')
      fetchUsers()
    } catch (error: unknown) {
      toast.showError(getApiErrorMessage(error, '设置强制改密失败'))
    }
  }

  const handleDelete = async (user: User) => {
    try {
      await authApi.deleteUser(user.id)
      toast.showSuccess('用户已删除')
      fetchUsers()
    } catch (error: unknown) {
      toast.showError(getApiErrorMessage(error, '删除用户失败'))
    }
  }

  const confirmDelete = (user: User) => {
    confirmDialog({
      message: `确定删除用户 ${user.username} 吗？`,
      header: '确认删除',
      icon: 'pi pi-exclamation-triangle',
      accept: () => handleDelete(user),
    })
  }

  const hideModal = () => {
    setModalVisible(false)
    setEditingUser(null)
    setFormErrors({})
  }

  const updateFormField = <K extends keyof UserFormValues>(field: K, value: UserFormValues[K]) => {
    setFormValues((prev) => ({ ...prev, [field]: value }))
    if (formErrors[field]) {
      setFormErrors((prev) => ({ ...prev, [field]: undefined }))
    }
  }

  const roleBodyTemplate = (user: User) => {
    const label = roleOptions.find((item) => item.value === user.role)?.label || user.role
    return <Tag value={label} severity={getRoleSeverity(user.role)} />
  }

  const statusBodyTemplate = (user: User) => {
    return (
      <div className="flex gap-2">
        <Tag value={user.disabled_at ? '已禁用' : '正常'} severity={user.disabled_at ? 'secondary' : 'success'} />
        {user.force_password_reset && <Tag value="强制改密" severity="warning" />}
      </div>
    )
  }

  const emailBodyTemplate = (user: User) => {
    return user.email || '-'
  }

  const actionBodyTemplate = (user: User) => {
    return (
      <div className="flex flex-wrap gap-1">
        <Button
          type="button"
          icon="pi pi-pencil"
          label="编辑"
          link
          size="small"
          style={{ color: 'var(--primary-color)' }}
          onClick={() => handleEdit(user)}
        />
        <Button
          type="button"
          icon="pi pi-lock"
          label="强制改密"
          link
          size="small"
          style={{ color: 'var(--warning-color)' }}
          onClick={() => handleForcePasswordReset(user)}
        />
        <Button
          type="button"
          label={user.disabled_at ? '启用' : '禁用'}
          link
          size="small"
          style={{ color: 'var(--text-secondary)' }}
          onClick={() => handleToggleDisabled(user)}
        />
        <Button
          type="button"
          icon="pi pi-trash"
          label="删除"
          outlined
          size="small"
          style={{
            color: 'var(--danger-color)',
            borderColor: 'var(--danger-color)',
          }}
          onClick={() => confirmDelete(user)}
        />
      </div>
    )
  }

  const dialogFooter = (
    <div className="flex justify-content-end gap-2">
      <Button type="button" label="取消" outlined onClick={hideModal} />
      <Button type="button" label="保存" onClick={handleSubmit} />
    </div>
  )

  // Audit log column templates
  const auditTimeTemplate = (log: AuditLog) => {
    return new Date(log.created_at).toLocaleString('zh-CN')
  }

  const auditActorTemplate = (log: AuditLog) => {
    const roleLabel = roleOptions.find((item) => item.value === log.actor_role)?.label || log.actor_role
    return (
      <div className="flex align-items-center gap-2">
        <span>{log.actor_username}</span>
        <Tag value={roleLabel} severity={getRoleSeverity(log.actor_role as User['role'])} />
      </div>
    )
  }

  const auditActionTemplate = (log: AuditLog) => {
    return actionLabelMap[log.action] || log.action
  }

  const auditTargetTypeTemplate = (log: AuditLog) => {
    return targetTypeLabelMap[log.target_type] || log.target_type
  }

  const auditResultTemplate = (log: AuditLog) => {
    return (
      <Tag
        value={log.result === 'allowed' ? '成功' : '拒绝'}
        severity={log.result === 'allowed' ? 'success' : 'danger'}
      />
    )
  }

  const handleAuditPageChange = (event: { page: number }) => {
    setAuditPage(event.page)
  }

  return (
    <Card
      title="用户管理"
      pt={{
        title: { className: 'text-xl font-semibold' },
      }}
    >
      <TabView>
        <TabPanel header="用户列表">
          <div className="flex justify-content-end mb-3">
            <Button type="button" icon="pi pi-plus" label="新建用户" onClick={handleCreate} />
          </div>

          <DataTable value={users} dataKey="id" loading={loading} stripedRows>
            <Column field="username" header="用户名" />
            <Column field="name" header="姓名" />
            <Column header="角色" body={roleBodyTemplate} />
            <Column header="状态" body={statusBodyTemplate} />
            <Column field="email" header="邮箱" body={emailBodyTemplate} />
            <Column header="操作" body={actionBodyTemplate} style={{ minWidth: '280px' }} />
          </DataTable>
        </TabPanel>

        <TabPanel header="审计日志">
          <DataTable value={auditLogs} dataKey="id" loading={auditLoading} stripedRows>
            <Column header="时间" body={auditTimeTemplate} style={{ minWidth: '160px' }} />
            <Column header="操作人" body={auditActorTemplate} />
            <Column header="操作类型" body={auditActionTemplate} />
            <Column header="目标类型" body={auditTargetTypeTemplate} />
            <Column field="target_id" header="目标ID" />
            <Column header="结果" body={auditResultTemplate} />
            <Column field="detail" header="详情" style={{ minWidth: '200px' }} />
          </DataTable>
          <Paginator
            first={auditPage * auditPageSize}
            rows={auditPageSize}
            totalRecords={auditTotal}
            onPageChange={handleAuditPageChange}
          />
        </TabPanel>
      </TabView>

      <Dialog
        header={editingUser ? `编辑用户: ${editingUser.username}` : '新建用户'}
        visible={modalVisible}
        onHide={hideModal}
        footer={dialogFooter}
        style={{ width: '450px' }}
        modal
      >
        <div className="flex flex-column gap-3">
          {!editingUser && (
            <>
              <div className="flex flex-column gap-2">
                <label htmlFor="username" className="font-medium">
                  用户名 <span className="text-red-500">*</span>
                </label>
                <InputText
                  id="username"
                  value={formValues.username}
                  onChange={(e) => updateFormField('username', e.target.value)}
                  placeholder="用户名"
                  className={formErrors.username ? 'p-invalid' : ''}
                />
                {formErrors.username && <small className="p-error">{formErrors.username}</small>}
              </div>
              <div className="flex flex-column gap-2">
                <label htmlFor="password" className="font-medium">
                  初始密码 <span className="text-red-500">*</span>
                </label>
                <Password
                  id="password"
                  value={formValues.password}
                  onChange={(e) => updateFormField('password', e.target.value)}
                  placeholder="初始密码"
                  feedback={false}
                  toggleMask
                  className={formErrors.password ? 'p-invalid' : ''}
                  inputClassName={formErrors.password ? 'p-invalid' : ''}
                />
                {formErrors.password && <small className="p-error">{formErrors.password}</small>}
              </div>
            </>
          )}
          <div className="flex flex-column gap-2">
            <label htmlFor="name" className="font-medium">
              姓名 <span className="text-red-500">*</span>
            </label>
            <InputText
              id="name"
              value={formValues.name}
              onChange={(e) => updateFormField('name', e.target.value)}
              placeholder="姓名"
              className={formErrors.name ? 'p-invalid' : ''}
            />
            {formErrors.name && <small className="p-error">{formErrors.name}</small>}
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="email" className="font-medium">
              邮箱
            </label>
            <InputText
              id="email"
              value={formValues.email}
              onChange={(e) => updateFormField('email', e.target.value)}
              placeholder="邮箱"
            />
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="role" className="font-medium">
              角色 <span className="text-red-500">*</span>
            </label>
            <Dropdown
              id="role"
              value={formValues.role}
              onChange={(e) => updateFormField('role', e.value)}
              options={roleOptions}
              optionLabel="label"
              optionValue="value"
              placeholder="选择角色"
              className={formErrors.role ? 'p-invalid' : ''}
            />
            {formErrors.role && <small className="p-error">{formErrors.role}</small>}
          </div>
          <div className="flex align-items-center gap-2">
            <InputSwitch
              id="force_password_reset"
              checked={formValues.force_password_reset}
              onChange={(e) => updateFormField('force_password_reset', e.value)}
            />
            <label htmlFor="force_password_reset" className="font-medium">
              首次登录/下次登录需要改密
            </label>
          </div>
        </div>
      </Dialog>
    </Card>
  )
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `cd d:/goproject/shadowsongAI/frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/Users.tsx
git commit -m "feat: refactor Users page with TabView and audit logs tab"
```

---

### Task 6: Run Full Test Suite and Verify

**Files:** None (verification only)

- [ ] **Step 1: Run backend tests**

Run: `cd d:/goproject/shadowsongAI && go test ./... -count=1`
Expected: All tests PASS

- [ ] **Step 2: Run frontend type check**

Run: `cd d:/goproject/shadowsongAI/frontend && npx tsc --noEmit`
Expected: No errors

- [ ] **Step 3: Run frontend tests**

Run: `cd d:/goproject/shadowsongAI/frontend && npx vitest run`
Expected: All tests PASS (existing tests should still pass; no new test file added for Users.tsx since it was missing before)

- [ ] **Step 4: Final commit if any fixes needed**

If any test failures required fixes, commit them:

```bash
git add -A
git commit -m "fix: resolve test failures from user management changes"
```
