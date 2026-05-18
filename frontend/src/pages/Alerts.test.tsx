import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Alerts } from './Alerts'
import { useUserStore } from '../stores/userStore'
import type { Alert, GroupedActiveAlert, User } from '../types'

const alertStoreState = vi.hoisted(() => ({
  groupedActiveAlerts: [] as GroupedActiveAlert[],
  groupedActiveLoading: false,
  fetchGroupedActiveAlerts: vi.fn(),
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
  PermissionNotice: ({ title }: { title: string }) => <div>{title}</div>,
  SeverityBadge: ({ severity }: { severity: string }) => <span>{severity}</span>,
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
  count: 3,
  first_triggered_at: '2026-04-12T00:00:00Z',
  last_triggered_at: '2026-04-12T01:00:00Z',
}

const renderAlerts = async () => {
  const view = render(
    <MemoryRouter>
      <Alerts />
    </MemoryRouter>
  )
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0))
  })
  return view
}

describe('Alerts page permissions', () => {
  beforeEach(() => {
    alertStoreState.fetchGroupedActiveAlerts.mockClear()
    alertStoreState.ackAlert.mockClear()
    alertStoreState.quickSilence.mockClear()
    alertStoreState.groupedActiveAlerts = [groupedAlert]
    useUserStore.setState({ user: null, token: null })
  })

  it('renders read-only state for viewer', async () => {
    useUserStore.setState({ user: baseUser, token: 'token' })

    await renderAlerts()

    expect(await screen.findByText('当前角色可查看告警，但不能确认或静默')).toBeInTheDocument()
    expect(await screen.findByText('只读')).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: '投递历史' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '确认' })).not.toBeInTheDocument()
  })

  it('renders alert action buttons for operator', async () => {
    useUserStore.setState({
      user: { ...baseUser, role: 'operator' },
      token: 'token',
    })

    await renderAlerts()

    expect(screen.queryByText('当前角色可查看告警，但不能确认或静默')).not.toBeInTheDocument()
    expect(await screen.findByRole('button', { name: '确认' })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: '静默' })).toBeInTheDocument()
  })

  it('displays occurrence count tag when count > 1', async () => {
    useUserStore.setState({ user: baseUser, token: 'token' })

    await renderAlerts()

    expect(await screen.findByText('共 3 次')).toBeInTheDocument()
  })

  it('displays notify count tag when notify_count > 0', async () => {
    useUserStore.setState({ user: baseUser, token: 'token' })

    // Use a recent last_notified_at so the escalation limit check (120 min) does not trigger
    const recentTime = new Date(Date.now() - 30 * 60 * 1000).toISOString()
    const notifiedAlert: Alert = {
      ...firingAlert,
      notify_count: 2,
      last_notified_at: recentTime,
    }
    alertStoreState.groupedActiveAlerts = [
      { ...groupedAlert, latest_alert: notifiedAlert },
    ]

    await renderAlerts()

    expect(await screen.findByText('已通知 2 次')).toBeInTheDocument()
  })

  it('does not display notify count tag when notify_count is 0', async () => {
    useUserStore.setState({ user: baseUser, token: 'token' })

    await renderAlerts()

    expect(screen.queryByText(/已通知 \d+ 次/)).not.toBeInTheDocument()
  })
})
