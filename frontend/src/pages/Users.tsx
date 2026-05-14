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
import { authApi, getApiErrorMessage } from '../api/auth'
import { PermissionNotice, useToast } from '../components'
import { canUser, capabilityManageUsers } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import type { User } from '../types'

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

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

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
      force_password_reset: user.force_password_reset,
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
        {user.force_password_reset && <Tag value="强制改密" severity="warn" />}
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
          onClick={() => handleEdit(user)}
        />
        <Button
          type="button"
          icon="pi pi-lock"
          label="强制改密"
          link
          size="small"
          onClick={() => handleForcePasswordReset(user)}
        />
        <Button
          type="button"
          label={user.disabled_at ? '启用' : '禁用'}
          link
          size="small"
          onClick={() => handleToggleDisabled(user)}
        />
        <Button
          type="button"
          icon="pi pi-trash"
          label="删除"
          link
          size="small"
          severity="danger"
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

  return (
    <Card
      title="用户管理"
      pt={{
        title: { className: 'text-xl font-semibold' },
      }}
    >
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
