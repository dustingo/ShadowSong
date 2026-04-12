import React from 'react'
import { render, screen } from '@testing-library/react'
import { Alerts } from './Alerts'
import { useUserStore } from '../stores/userStore'
import type { Alert, User } from '../types'

const alertStoreState = vi.hoisted(() => ({
  alerts: [] as Alert[],
  total: 0,
  page: 1,
  pageSize: 20,
  loading: false,
  filters: {},
  fetchAlerts: vi.fn(),
  setFilters: vi.fn(),
  ackAlert: vi.fn(),
  quickSilence: vi.fn(),
}))

vi.mock('../stores/alertStore', () => ({
  useAlertStore: () => alertStoreState,
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

const firingAlert: Alert = {
  alert_id: 'a-1',
  source: 'prometheus',
  alert_name: 'LatencyHigh',
  severity: 'P1',
  message: 'latency high',
  labels: { env: 'prod' },
  fingerprint: 'fp',
  trigger_time: '2026-04-12T00:00:00Z',
  received_at: '2026-04-12T00:00:00Z',
  status: 'firing',
  raw: {},
  trigger_count: 1,
  created_at: '2026-04-12T00:00:00Z',
  updated_at: '2026-04-12T00:00:00Z',
}

describe('Alerts page permissions', () => {
  beforeEach(() => {
    alertStoreState.fetchAlerts.mockClear()
    alertStoreState.setFilters.mockClear()
    alertStoreState.ackAlert.mockClear()
    alertStoreState.quickSilence.mockClear()
    alertStoreState.alerts = [firingAlert]
    alertStoreState.total = 1
    useUserStore.setState({ user: null, token: null })
  })

  it('renders read-only state for viewer', () => {
    useUserStore.setState({ user: baseUser, token: 'token' })

    render(<Alerts />)

    expect(screen.getByText('当前角色可查看告警，但不能确认或静默')).toBeInTheDocument()
    expect(screen.getByText('只读')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '确认' })).not.toBeInTheDocument()
  })

  it('renders alert action buttons for operator', () => {
    useUserStore.setState({
      user: { ...baseUser, role: 'operator' },
      token: 'token',
    })

    render(<Alerts />)

    expect(screen.queryByText('当前角色可查看告警，但不能确认或静默')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: '确认' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '静默' })).toBeInTheDocument()
  })
})
