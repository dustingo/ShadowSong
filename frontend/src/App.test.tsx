import React from 'react'
import { render, screen } from '@testing-library/react'
import App from './App'
import { useUserStore } from './stores/userStore'
import type { User } from './types'

vi.mock('./pages', () => ({
  Dashboard: () => <div>Dashboard Page</div>,
  Alerts: () => <div>Alerts Page</div>,
  DataSources: () => <div>DataSources Page</div>,
  Channels: () => <div>Channels Page</div>,
  RouteRules: () => <div>RouteRules Page</div>,
  Silences: () => <div>Silences Page</div>,
  OnDutyPage: () => <div>OnDuty Page</div>,
  Login: ({ onSuccess }: { onSuccess?: (token: string, user: User) => void }) => (
    <button
      type="button"
      onClick={() =>
        onSuccess?.('token', {
          id: 1,
          username: 'tester',
          name: 'Tester',
          role: 'admin',
          created_at: '',
          updated_at: '',
          force_password_reset: false,
        })
      }
    >
      Login Page
    </button>
  ),
  Users: () => <div>Users Page</div>,
  Profile: () => <div>Profile Page</div>,
}))

const baseUser: User = {
  id: 1,
  username: 'tester',
  name: 'Tester',
  role: 'viewer',
  created_at: '2026-04-12T00:00:00Z',
  updated_at: '2026-04-12T00:00:00Z',
  force_password_reset: false,
}

const setAuthState = (user: User | null, token = 'token') => {
  useUserStore.setState({
    user,
    token: user ? token : null,
  })
}

const renderAt = (path: string) => {
  window.history.pushState({}, '', path)
  return render(<App />)
}

describe('App routing', () => {
  beforeEach(() => {
    localStorage.clear()
    setAuthState(null, '')
  })

  it('redirects force-password-reset users to profile', () => {
    setAuthState({
      ...baseUser,
      force_password_reset: true,
    })

    renderAt('/alerts')

    expect(screen.getByText('Profile Page')).toBeInTheDocument()
    expect(screen.getByText('必须先完成密码修改')).toBeInTheDocument()
  })

  it('shows forbidden notice when viewer opens user management', () => {
    setAuthState(baseUser)

    renderAt('/users')

    expect(screen.getByText('当前角色无权执行该操作')).toBeInTheDocument()
    expect(screen.queryByText('Users Page')).not.toBeInTheDocument()
  })
})
