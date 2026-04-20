import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { Dashboard } from './Dashboard'
import { useUserStore } from '../stores/userStore'
import type { Alert, User } from '../types'

vi.mock('echarts-for-react', () => ({
  default: () => <div data-testid="trend-chart" />,
}))

vi.mock('../components/AlertCard', () => ({
  AlertCard: ({ alert }: { alert: Alert }) => <div>{alert.alert_name}</div>,
}))

const alertStoreState = vi.hoisted(() => ({
  activeAlerts: [] as Alert[],
  stats: {
    total: 1,
    firing: 1,
    acked: 0,
    silenced: 0,
    by_severity: { P0: 0, P1: 1, P2: 0, P3: 0 },
    trend: [],
  },
  loading: false,
  wsConnected: false,
  fetchActiveAlerts: vi.fn(),
  fetchStats: vi.fn(),
  setWsConnected: vi.fn(),
  ackAlert: vi.fn(),
  quickSilence: vi.fn(),
  addAlert: vi.fn(),
  updateAlert: vi.fn(),
}))

vi.mock('../stores/alertStore', () => ({
  useAlertStore: () => alertStoreState,
}))

const wsInstances: MockWebSocket[] = []

class MockWebSocket {
  static OPEN = 1
  static CLOSED = 3

  readyState = MockWebSocket.OPEN
  onopen: (() => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  constructor(public url: string) {
    wsInstances.push(this)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
  }
}

const baseUser: User = {
  id: 1,
  username: 'tester',
  name: 'Tester',
  role: 'operator',
  created_at: '2026-04-20T00:00:00Z',
  updated_at: '2026-04-20T00:00:00Z',
  force_password_reset: false,
}

describe('Dashboard websocket auth', () => {
  const originalWebSocket = globalThis.WebSocket

  beforeEach(() => {
    wsInstances.length = 0
    alertStoreState.fetchActiveAlerts.mockClear()
    alertStoreState.fetchStats.mockClear()
    alertStoreState.setWsConnected.mockClear()
    useUserStore.setState({ user: null, token: null })
    vi.spyOn(window, 'setInterval').mockReturnValue(1 as unknown as ReturnType<typeof setInterval>)
    vi.spyOn(window, 'clearInterval').mockImplementation(() => {})
    globalThis.WebSocket = MockWebSocket as unknown as typeof WebSocket
  })

  afterEach(() => {
    vi.restoreAllMocks()
    globalThis.WebSocket = originalWebSocket
  })

  it('connects to /ws/alerts with the persisted token and marks websocket connected', async () => {
    useUserStore.setState({ user: baseUser, token: 'signed-token' })

    render(<Dashboard />)

    await act(async () => {
      await Promise.resolve()
    })

    expect(wsInstances).toHaveLength(1)
    expect(wsInstances[0].url).toContain('/ws/alerts?token=signed-token')

    act(() => {
      wsInstances[0].onopen?.()
    })

    expect(alertStoreState.setWsConnected).toHaveBeenCalledWith(true)
    expect(alertStoreState.fetchActiveAlerts).toHaveBeenCalled()
    expect(alertStoreState.fetchStats).toHaveBeenCalled()
  })

  it('skips websocket creation when no token exists and leaves polling fallback intact', async () => {
    render(<Dashboard />)

    await act(async () => {
      await Promise.resolve()
    })

    expect(wsInstances).toHaveLength(0)
    expect(alertStoreState.setWsConnected).toHaveBeenCalledWith(false)
    expect(alertStoreState.fetchActiveAlerts).toHaveBeenCalled()
    expect(alertStoreState.fetchStats).toHaveBeenCalled()
    expect(screen.getByText('实时连接已断开')).toBeInTheDocument()
  })
})
