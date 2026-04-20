import React, { useState } from 'react'
import { Form, Input, Button, Card, message, Typography, Space } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import axios from 'axios'
import { getApiErrorMessage } from '../api/client'
import type { User } from '../types'

const { Title, Text } = Typography

interface LoginForm {
  username: string
  password: string
  remember?: boolean
}

interface LoginProps {
  onSuccess?: (token: string, user: User) => void
}

interface LoginResponse {
  token: string
  user: User
}

export const Login: React.FC<LoginProps> = ({ onSuccess }) => {
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true)
    try {
      const res = await axios.post<LoginResponse>('/api/v1/auth/login', {
        username: values.username,
        password: values.password,
      })

      const { token, user } = res.data

      if (onSuccess) {
        onSuccess(token, user)
      } else {
        localStorage.setItem('token', token)
        localStorage.setItem('user', JSON.stringify(user))
        message.success('登录成功')
        window.location.href = '/'
      }
    } catch (error: unknown) {
      message.error(getApiErrorMessage(error, '登录失败'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: '20px',
    }}>
      <Card style={{ width: 400, boxShadow: '0 8px 32px rgba(0,0,0,0.2)', borderRadius: 12 }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title level={3} style={{ marginBottom: 8, color: '#333' }}>
            游戏运维告警系统
          </Title>
          <Text type="secondary">请登录您的账户</Text>
        </div>

        <Form
          name="login"
          onFinish={handleSubmit}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#bbb' }} />}
              placeholder="用户名"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#bbb' }} />}
              placeholder="密码"
            />
          </Form.Item>

          <Form.Item name="remember" valuePropName="checked">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ color: '#666' }}>记住登录状态</span>
            </div>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{ height: 44, fontSize: 16 }}
            >
              登 录
            </Button>
          </Form.Item>
        </Form>

        <div style={{ marginTop: 24, textAlign: 'center', color: '#999', fontSize: 12 }}>
          <Space direction="vertical" size={0}>
            <span>如需创建或重置账户，请联系管理员</span>
          </Space>
        </div>
      </Card>
    </div>
  )
}
