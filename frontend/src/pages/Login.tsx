import React, { useState } from 'react'
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
    <div className="login-wrapper">
      {/* Left side */}
      <div className="login-left">
        <div className="login-left-content">
          <div className="login-app-logo">
            <i className="pi pi-bell" />
          </div>
          <h1 className="login-app-name">ShadowSong</h1>
          <p className="login-app-desc">游戏运维告警系统</p>
          <div className="login-decoration">
            <svg viewBox="0 0 500 400" fill="none" xmlns="http://www.w3.org/2000/svg">
              <ellipse cx="250" cy="200" rx="120" ry="100" fill="rgba(255,255,255,0.6)" />
              <ellipse cx="150" cy="120" rx="60" ry="50" fill="rgba(255,255,255,0.4)" />
              <ellipse cx="350" cy="280" rx="80" ry="70" fill="rgba(255,255,255,0.3)" />
              <circle cx="100" cy="300" r="30" fill="rgba(255,255,255,0.25)" />
              <circle cx="400" cy="100" r="25" fill="rgba(255,255,255,0.25)" />
            </svg>
          </div>
        </div>
      </div>

      {/* Right side */}
      <div className="login-right">
        <div className="login-form-container">
          <div className="login-form-header">
            <div className="login-form-logo">
              <i className="pi pi-shield" />
            </div>
            <h2 className="login-form-title">LOGIN</h2>
          </div>

          <form onSubmit={handleSubmit} className="login-form">
            {error && (
              <div className="login-error">
                <Message severity="error" text={error} />
              </div>
            )}

            <div className="login-field">
              <label htmlFor="username" className="login-label">
                用户名
              </label>
              <InputText
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
                className="login-input"
              />
            </div>

            <div className="login-field">
              <label htmlFor="password" className="login-label">
                密码
              </label>
              <Password
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                feedback={false}
                toggleMask
                className="login-password"
                inputClassName="login-input"
              />
            </div>

            <Button
              type="submit"
              label="LOGIN"
              loading={loading}
              className="login-button"
            />
          </form>
        </div>
      </div>

      <style>{`
        .login-wrapper {
          display: flex;
          min-height: 100vh;
          width: 100%;
        }

        /* Left side - 使用主题配色 */
        .login-left {
          flex: 1;
          background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 50%, #BFDBFE 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 2rem;
          position: relative;
          overflow: hidden;
        }

        .login-left::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%233B82F6' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
        }

        .login-left-content {
          text-align: center;
          color: #1E40AF;
          z-index: 1;
          max-width: 400px;
        }

        .login-app-logo {
          margin-bottom: 1.5rem;
        }

        .login-app-logo i {
          font-size: 3rem;
          color: #3B82F6;
        }

        .login-app-name {
          font-size: 2.5rem;
          font-weight: 600;
          margin: 0 0 0.75rem 0;
          letter-spacing: 2px;
          color: #1E40AF;
        }

        .login-app-desc {
          font-size: 1rem;
          margin: 0 0 3rem 0;
          color: #64748B;
        }

        .login-decoration {
          width: 100%;
        }

        .login-decoration svg {
          width: 100%;
          height: auto;
        }

        /* Right side */
        .login-right {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--surface-card);
          padding: 2rem;
        }

        .login-form-container {
          width: 100%;
          max-width: 320px;
        }

        .login-form-header {
          text-align: center;
          margin-bottom: 2rem;
        }

        .login-form-logo {
          margin-bottom: 1rem;
        }

        .login-form-logo i {
          font-size: 2.5rem;
          color: #3B82F6;
        }

        .login-form-title {
          font-size: 1.25rem;
          font-weight: 600;
          color: var(--text-primary);
          margin: 0;
          letter-spacing: 1px;
        }

        .login-form {
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
        }

        .login-error {
          margin-bottom: 0.5rem;
        }

        .login-field {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
        }

        .login-label {
          font-size: 0.875rem;
          font-weight: 500;
          color: var(--text-secondary);
        }

        .login-input {
          width: 100% !important;
          height: 40px !important;
        }

        .login-password {
          width: 100% !important;
        }

        .login-password .p-icon-field {
          width: 100%;
        }

        .login-password .p-password-input {
          width: 100% !important;
          height: 40px !important;
          padding-right: 2.5rem !important;
        }

        .login-button {
          width: 100%;
          height: 40px;
          background: #3B82F6;
          border: none;
          font-weight: 500;
          letter-spacing: 1px;
        }

        .login-button:hover {
          background: #2563EB;
        }

        /* Responsive */
        @media (max-width: 968px) {
          .login-left {
            display: none;
          }

          .login-right {
            flex: none;
            width: 100%;
          }
        }

        @media (max-width: 480px) {
          .login-form-container {
            padding: 1rem;
          }
        }
      `}</style>
    </div>
  )
}