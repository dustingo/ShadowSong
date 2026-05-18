import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Dashboard } from './Dashboard'
import { useUserStore } from '../stores/userStore'
import type { Alert, GroupedActiveAlert, User } from '../types'

const alertStoreState = vi.hoisted(() => ({
  groupedActiveAlerts: [] as GroupedActiveAlert[],
  groupedActiveLoading: false,
  stats: { firing: 1, acked: 0, silenced: 0, by_severity: { P0: 0, P1: 1 }, trend: [] },
  wsConnected: true,
  fetchGroupedActiveAlerts: vi.fn(),
  fetchStats: vi.fn(),
  setWsConnected: vi.fn(),
  ackAlert: vi.fn(),
  quickSilence: vi.fn(),
}))

vi.mock('../stores/alertStore', () => ({
  useAlertStore: () => alertStoreState,
}))

vi.mock('../components', () => ({
  useToast: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarn: vi.fn(),
    showInfo: vi.fn(),
  }),
  StatisticCard: ({ label, value }: { label: string; value: number }) => (
    <div data-testid="stat-card">
      {label}: {value}
    </div>
  ),
  PermissionNotice: ({ title }: { title: string }) => <div>{title}</div>,
}))

vi.mock('../components/AlertCard', () => ({
  AlertCard: ({ alert, showActions }: { alert: Alert; showActions: boolean }) => (
    <div data-testid="alert-card">
      {alert.alert_name} - {showActions ? 'actions' : 'no-actions'}
    </div>
  ),
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
  trigger_count: 3,
  last_notified_at: null,
  notify_count: 0,
  created_at: '2026-04-12T00:00:00Z',
  updated_at: '2026-04-12T00:00:00Z',
}

const groupedAlert: GroupedActiveAlert = {
  fingerprint: 'fp',
  latest_alert: firingAlert,
  count: 5,
  first_triggered_at: '2026-04-12T00:00:00Z',
  last_triggered_at: '2026-04-12T01:00:00Z',
}

const renderDashboard = async () => {
  const view = render(
    <MemoryRouter>
      <Dashboard />
    </MemoryRouter>
  )
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0))
  })
  return view
}

describe('Dashboard', () => {
  beforeEach(() => {
    alertStoreState.fetchGroupedActiveAlerts.mockClear()
    alertStoreState.fetchStats.mockClear()
    alertStoreState.groupedActiveAlerts = [groupedAlert]
    useUserStore.setState({ user: baseUser, token: 'token' })
  })

  it('renders stats cards', async () => {
    await renderDashboard()

    expect(await screen.findByText('活跃告警: 1')).toBeInTheDocument()
    expect(await screen.findByText('P0 告警: 0')).toBeInTheDocument()
    expect(await screen.findByText('待确认告警: 0')).toBeInTheDocument()
  })

  it('computes pending ack count from notified but un-acked firing alerts', async () => {
    const notifiedAlert: Alert = {
      ...firingAlert,
      alert_id: 'a-2',
      alert_name: 'NotifiedAlert',
      notify_count: 2,
      last_notified_at: '2026-04-12T00:30:00Z',
    }
    const ackedAlert: Alert = {
      ...firingAlert,
      alert_id: 'a-3',
      alert_name: 'AckedAlert',
      status: 'acked',
      acked_at: '2026-04-12T01:00:00Z',
      notify_count: 1,
      last_notified_at: '2026-04-12T00:30:00Z',
    }
    const notNotifiedAlert: Alert = {
      ...firingAlert,
      alert_id: 'a-4',
      alert_name: 'NotNotifiedAlert',
      notify_count: 0,
      last_notified_at: null,
    }

    alertStoreState.groupedActiveAlerts = [
      { fingerprint: 'fp-2', latest_alert: notifiedAlert, count: 1, first_triggered_at: '2026-04-12T00:00:00Z', last_triggered_at: '2026-04-12T01:00:00Z' },
      { fingerprint: 'fp-3', latest_alert: ackedAlert, count: 1, first_triggered_at: '2026-04-12T00:00:00Z', last_triggered_at: '2026-04-12T01:00:00Z' },
      { fingerprint: 'fp-4', latest_alert: notNotifiedAlert, count: 1, first_triggered_at: '2026-04-12T00:00:00Z', last_triggered_at: '2026-04-12T01:00:00Z' },
    ]

    await renderDashboard()

    // Only notifiedAlert qualifies: firing + un-acked + notify_count > 0
    expect(await screen.findByText('待确认告警: 1')).toBeInTheDocument()
  })

  it('renders grouped active alerts with occurrence count', async () => {
    await renderDashboard()

    expect(await screen.findByText('LatencyHigh - actions')).toBeInTheDocument()
    expect(await screen.findByText('共 5 次')).toBeInTheDocument()
  })

  it('shows empty state when no active alerts', async () => {
    alertStoreState.groupedActiveAlerts = []

    await renderDashboard()

    expect(await screen.findByText('暂无活跃告警，系统运行正常')).toBeInTheDocument()
  })
})
