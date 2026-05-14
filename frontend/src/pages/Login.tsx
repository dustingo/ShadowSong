import React, { useState } from 'react'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Password } from 'primereact/password'
import { Button } from 'primereact/button'
import { Message } from 'primereact/message'
import axios from 'axios'
import { getApiErrorMessage } from '../api/client'
import type { User } from '../types'

interface LoginProps {
  onSuccess?: (token: string, user: User) => void
}

interface LoginResponse {
  token: string
  user: User
}

export const Login: React.FC<LoginProps> = ({ onSuccess }) => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username || !password) {
      setError('请输入用户名和密码')
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await axios.post<LoginResponse>('/api/v1/auth/login', {
        username,
        password,
      })

      const { token, user } = res.data

      if (onSuccess) {
        onSuccess(token, user)
      } else {
        localStorage.setItem('token', token)
        localStorage.setItem('user', JSON.stringify(user))
        window.location.href = '/'
      }
    } catch (err) {
      setError(getApiErrorMessage(err, '登录失败'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="flex align-items-center justify-content-center"
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '20px',
      }}
    >
      <Card
        style={{
          width: '400px',
          boxShadow: '0 8px 32px rgba(0,0,0,0.2)',
          borderRadius: '12px',
        }}
      >
        <div className="text-center mb-4">
          <h2 className="text-xl font-semibold text-slate-700 mb-2">
            游戏运维告警系统
          </h2>
          <p className="text-slate-500 text-sm">请登录您的账户</p>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="flex flex-column gap-3">
            {error && (
              <Message severity="error" text={error} />
            )}

            <div className="flex flex-column gap-2">
              <label htmlFor="username" className="text-sm text-slate-600">
                用户名
              </label>
              <InputText
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
                className="w-full"
              />
            </div>

            <div className="flex flex-column gap-2">
              <label htmlFor="password" className="text-sm text-slate-600">
                密码
              </label>
              <Password
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                feedback={false}
                toggleMask
                className="w-full"
                inputClassName="w-full"
              />
            </div>

            <Button
              type="submit"
              label="登 录"
              loading={loading}
              className="w-full mt-2"
              style={{ height: '44px' }}
            />
          </div>
        </form>

        <p className="text-center text-slate-400 text-xs mt-4">
          如需创建或重置账户，请联系管理员
        </p>
      </Card>
    </div>
  )
}