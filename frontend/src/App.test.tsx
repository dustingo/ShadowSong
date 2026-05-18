import React from 'react'
import { act, render, screen } from '@testing-library/react'
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
  Deliveries: () => <div>Deliveries Page</div>,
  OpsHealth: () => <div>OpsHealth Page</div>,
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

const waitForAntd = async () => {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0))
  })
}

const renderAt = async (path: string) => {
  window.history.pushState({}, '', path)
  const view = render(<App />)
  await waitForAntd()
  return view
}

const collectCalls = (calls: unknown[][]) =>
  calls
    .map((args) => args.map((value) => String(value)).join(' '))
    .join('\n')

describe('App routing', () => {
  beforeEach(() => {
    localStorage.clear()
    setAuthState(null, '')
  })

  it('redirects force-password-reset users to profile', async () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    setAuthState({
      ...baseUser,
      force_password_reset: true,
    })

    await renderAt('/alerts')

    expect(await screen.findByText('Profile Page')).toBeInTheDocument()
    expect(collectCalls(warnSpy.mock.calls)).not.toMatch(/future flag|React Router will begin wrapping state updates/i)
    warnSpy.mockRestore()
  })

  it('shows forbidden notice when viewer opens user management', async () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

    setAuthState(baseUser)

    await renderAt('/users')

    expect(await screen.findByText('当前角色无权执行该操作')).toBeInTheDocument()
    expect(screen.queryByText('Users Page')).not.toBeInTheDocument()
    expect(collectCalls(warnSpy.mock.calls)).not.toMatch(/future flag|React Router will begin wrapping state updates/i)
    warnSpy.mockRestore()
  })
})
