import React, { useEffect } from 'react'
import { Button, Card, Form, Input, Space, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { authApi } from '../api/auth'
import { useUserStore } from '../stores/userStore'

type ProfileFormValues = {
  name: string
  email?: string
}

type PasswordFormValues = {
  password: string
}

export const Profile: React.FC = () => {
  const navigate = useNavigate()
  const user = useUserStore((state) => state.user)
  const setUser = useUserStore((state) => state.setUser)
  const logout = useUserStore((state) => state.logout)
  const [profileForm] = Form.useForm<ProfileFormValues>()
  const [passwordForm] = Form.useForm<PasswordFormValues>()

  useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        name: user.name,
        email: user.email,
      })
    }
  }, [profileForm, user])

  const handleProfileSubmit = async () => {
    try {
      const values = await profileForm.validateFields()
      const nextUser = await authApi.updateOwnProfile(values)
      setUser(nextUser)
      message.success('个人资料已更新')
    } catch (error) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return
      }
      message.error('更新个人资料失败')
    }
  }

  const handlePasswordSubmit = async () => {
    try {
      const values = await passwordForm.validateFields()
      await authApi.updateOwnPassword(values.password)
      logout()
      message.success('密码已更新，请重新登录')
      navigate('/login')
    } catch (error) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return
      }
      message.error('更新密码失败')
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Card title="个人资料">
        <Form form={profileForm} layout="vertical">
          <Form.Item label="用户名">
            <Input value={user?.username} disabled />
          </Form.Item>
          <Form.Item name="name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="姓名" />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="邮箱" />
          </Form.Item>
          <Button type="primary" onClick={handleProfileSubmit}>
            保存资料
          </Button>
        </Form>
      </Card>

      <Card title="修改密码">
        <Form form={passwordForm} layout="vertical">
          <Form.Item
            name="password"
            label="新密码"
            rules={[{ required: true, message: '请输入新密码' }]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Button type="primary" onClick={handlePasswordSubmit}>
            更新密码
          </Button>
        </Form>
      </Card>
    </Space>
  )
}
