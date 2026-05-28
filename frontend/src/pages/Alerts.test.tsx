import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { Alerts } from './Alerts'
import { useAlertStore } from '../stores/alertStore'
import { useUserStore } from '../stores/userStore'

// Mock react-router-dom
const mockNavigate = vi.fn()
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}))

// Mock useToast
const mockShowSuccess = vi.fn()
const mockShowError = vi.fn()
const mockShowWarn = vi.fn()
vi.mock('../components', () => ({
  useToast: () => ({
    showSuccess: mockShowSuccess,
    showError: mockShowError,
    showWarn: mockShowWarn,
  }),
  PermissionNotice: ({ title }: { title: string }) => <div>{title}</div>,
}))

// Mock authz
vi.mock('../authz/capabilities', () => ({
  canProcessAlerts: () => true,
}))

// Mock dayjs
vi.mock('dayjs', () => ({
  default: (date: string) => ({
    format: () => '2024-01-01 00:00:00',
    diff: () => 0,
  }),
}))

const mockFetchAlerts = vi.fn()
const mockSetFilters = vi.fn()
const mockFetchGroupedActiveAlerts = vi.fn()
const mockAckAlert = vi.fn()
const mockQuickSilence = vi.fn()
const mockFetchStats = vi.fn()

const mockAlert = {
  alert_id: 'alert-1',
  alert_name: 'Test Alert',
  severity: 'P1',
  source: 'prometheus',
  status: 'firing',
  message: 'Test message',
  trigger_time: '2024-01-01T00:00:00Z',
  trigger_count: 1,
  labels: { env: 'prod' },
  notify_count: 0,
  last_notified_at: null,
  fingerprint: 'fp-1',
  received_at: '2024-01-01T00:00:00Z',
  raw: {},
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockGroupedAlert = {
  fingerprint: 'fp-1',
  count: 2,
  first_triggered_at: '2024-01-01T00:00:00Z',
  last_triggered_at: '2024-01-01T00:00:00Z',
  latest_alert: mockAlert,
}

const defaultStoreState = {
  alerts: [mockAlert],
  total: 1,
  page: 1,
  pageSize: 10,
  loading: false,
  filters: {},
  groupedActiveAlerts: [mockGroupedAlert],
  groupedActiveLoading: false,
  fetchAlerts: mockFetchAlerts,
  setFilters: mockSetFilters,
  fetchGroupedActiveAlerts: mockFetchGroupedActiveAlerts,
  ackAlert: mockAckAlert,
  quickSilence: mockQuickSilence,
  fetchStats: mockFetchStats,
}

vi.mock('../stores/alertStore', () => ({
  useAlertStore: vi.fn(() => defaultStoreState),
}))

vi.mock('../stores/userStore', () => ({
  useUserStore: vi.fn(() => ({
    user: { role: 'admin' },
  })),
}))

vi.mock('../api/client', () => ({
  alertApi: {
    deliveries: vi.fn().mockResolvedValue([]),
  },
  getApiErrorMessage: (_e: unknown, fallback: string) => fallback,
}))

describe('Alerts page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders grouped active alerts section when there are active alerts', () => {
    render(<Alerts />)
    expect(screen.getByText(/活跃告警/)).toBeInTheDocument()
  })

  it('renders all alerts table with alert data', () => {
    render(<Alerts />)
    expect(screen.getByText('Test Alert')).toBeInTheDocument()
  })

  it('fetches both grouped active alerts and all alerts on mount', () => {
    render(<Alerts />)
    expect(mockFetchGroupedActiveAlerts).toHaveBeenCalled()
    expect(mockFetchAlerts).toHaveBeenCalled()
  })

  it('does not show grouped active alerts section when no active alerts', () => {
    vi.mocked(useAlertStore).mockReturnValue({
      ...defaultStoreState,
      groupedActiveAlerts: [],
    } as ReturnType<typeof useAlertStore>)

    render(<Alerts />)
    expect(screen.queryByText(/活跃告警/)).not.toBeInTheDocument()
  })

  it('renders filter section with severity and status labels', () => {
    render(<Alerts />)
    expect(screen.getByText('级别')).toBeInTheDocument()
    expect(screen.getAllByText('状态').length).toBeGreaterThanOrEqual(1)
  })

  it('renders search and reset buttons', () => {
    render(<Alerts />)
    expect(screen.getByText('搜索')).toBeInTheDocument()
    expect(screen.getByText('重置')).toBeInTheDocument()
  })
})