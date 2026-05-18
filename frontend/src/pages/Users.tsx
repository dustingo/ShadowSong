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
import { Paginator, PaginatorPageChangeEvent } from 'primereact/paginator'
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
  const [activeTabIndex, setActiveTabIndex] = useState(0)

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
    if (activeTabIndex === 1) {
      fetchAuditLogs(auditPage)
    }
  }, [fetchAuditLogs, auditPage, activeTabIndex])

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

  const handleAuditPageChange = (event: PaginatorPageChangeEvent) => {
    setAuditPage(event.page)
  }

  return (
    <Card
      title="用户管理"
      pt={{
        title: { className: 'text-xl font-semibold' },
      }}
    >
      <TabView activeIndex={activeTabIndex} onTabChange={(e) => setActiveTabIndex(e.index)}>
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
