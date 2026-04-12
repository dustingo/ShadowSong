import React, { useEffect, useState } from 'react'
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from 'antd'
import { EditOutlined, LockOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { authApi } from '../api/auth'
import type { User } from '../types'

type UserFormValues = {
  username?: string
  password?: string
  name: string
  email?: string
  role: User['role']
  force_password_reset?: boolean
}

const roleOptions: Array<{ value: User['role']; label: string }> = [
  { value: 'admin', label: '管理员' },
  { value: 'operator', label: '值班操作员' },
  { value: 'viewer', label: '只读用户' },
]

export const Users: React.FC = () => {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [form] = Form.useForm<UserFormValues>()

  const fetchUsers = async () => {
    setLoading(true)
    try {
      const nextUsers = await authApi.listUsers()
      setUsers(nextUsers)
    } catch (error) {
      message.error('加载用户列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchUsers()
  }, [])

  const handleCreate = () => {
    setEditingUser(null)
    form.resetFields()
    form.setFieldsValue({
      role: 'viewer',
      force_password_reset: true,
    })
    setModalVisible(true)
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    form.setFieldsValue({
      name: user.name,
      email: user.email,
      role: user.role,
      force_password_reset: user.force_password_reset,
    })
    setModalVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingUser) {
        await authApi.updateUser(editingUser.id, {
          name: values.name,
          email: values.email,
          role: values.role,
          force_password_reset: values.force_password_reset,
        })
        message.success('用户已更新')
      } else {
        await authApi.createUser({
          username: values.username || '',
          password: values.password || '',
          name: values.name,
          email: values.email,
          role: values.role,
          force_password_reset: values.force_password_reset,
        })
        message.success('用户已创建')
      }
      setModalVisible(false)
      setEditingUser(null)
      form.resetFields()
      fetchUsers()
    } catch (error) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return
      }
      message.error('保存用户失败')
    }
  }

  const handleToggleDisabled = async (user: User) => {
    try {
      await authApi.updateUser(user.id, { disabled: !user.disabled_at })
      message.success(user.disabled_at ? '账号已启用' : '账号已禁用')
      fetchUsers()
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || '更新账号状态失败'
      message.error(errorMsg)
    }
  }

  const handleForcePasswordReset = async (user: User) => {
    try {
      await authApi.updateUser(user.id, { force_password_reset: true })
      message.success('已标记为强制改密')
      fetchUsers()
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || '设置强制改密失败'
      message.error(errorMsg)
    }
  }

  const handleDelete = async (user: User) => {
    try {
      await authApi.deleteUser(user.id)
      message.success('用户已删除')
      fetchUsers()
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || '删除用户失败'
      message.error(errorMsg)
    }
  }

  const columns = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
    },
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: User['role']) => {
        const label = roleOptions.find((item) => item.value === role)?.label || role
        return <Tag color={role === 'admin' ? 'red' : role === 'operator' ? 'blue' : 'default'}>{label}</Tag>
      },
    },
    {
      title: '状态',
      key: 'status',
      render: (_: unknown, user: User) => (
        <Space>
          <Tag color={user.disabled_at ? 'default' : 'green'}>{user.disabled_at ? '已禁用' : '正常'}</Tag>
          {user.force_password_reset && <Tag color="orange">强制改密</Tag>}
        </Space>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      render: (email?: string) => email || '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, user: User) => (
        <Space wrap>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(user)}>
            编辑
          </Button>
          <Button type="link" size="small" icon={<LockOutlined />} onClick={() => handleForcePasswordReset(user)}>
            强制改密
          </Button>
          <Button type="link" size="small" onClick={() => handleToggleDisabled(user)}>
            {user.disabled_at ? '启用' : '禁用'}
          </Button>
          <Popconfirm title={`确定删除用户 ${user.username} 吗？`} onConfirm={() => handleDelete(user)}>
            <Button type="link" danger size="small" icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="用户管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建用户
        </Button>
      }
    >
      <Table columns={columns} dataSource={users} rowKey="id" loading={loading} />

      <Modal
        title={editingUser ? `编辑用户: ${editingUser.username}` : '新建用户'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false)
          setEditingUser(null)
          form.resetFields()
        }}
      >
        <Form form={form} layout="vertical">
          {!editingUser && (
            <>
              <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
                <Input placeholder="用户名" />
              </Form.Item>
              <Form.Item name="password" label="初始密码" rules={[{ required: true, message: '请输入初始密码' }]}>
                <Input.Password placeholder="初始密码" />
              </Form.Item>
            </>
          )}
          <Form.Item name="name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="姓名" />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="邮箱" />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="force_password_reset" label="首次登录/下次登录需要改密" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
