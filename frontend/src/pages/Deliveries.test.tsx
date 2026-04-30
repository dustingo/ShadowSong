import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { Deliveries } from './Deliveries'
import { useUserStore } from '../stores/userStore'
import type { Delivery, User } from '../types'

const deliveryApiState = vi.hoisted(() => ({
  list: vi.fn(),
  get: vi.fn(),
  retry: vi.fn(),
  replay: vi.fn(),
}))

vi.mock('../api/client', () => ({
  deliveryApi: deliveryApiState,
  getApiErrorMessage: (error: unknown, fallback: string) =>
    error instanceof Error ? error.message : fallback,
}))

const baseUser: User = {
  id: 1,
  username: 'viewer',
  name: 'Viewer',
  role: 'viewer',
  created_at: '2026-04-30T00:00:00Z',
  updated_at: '2026-04-30T00:00:00Z',
  force_password_reset: false,
}

const operatorUser: User = {
  ...baseUser,
  id: 2,
  username: 'operator',
  name: 'Operator',
  role: 'operator',
}

const baseDelivery: Delivery = {
  id: 11,
  alert_id: 'alert-123',
  trace_id: 'trace-123',
  channel_id: 7,
  delivery_status: 'failed',
  delivery_mode: 'rendered',
  attempt_count: 3,
  final_failure_summary: {
    result: 'failed',
    retryable: true,
    error_message: 'timeout from upstream',
    attempt_count: 3,
    trigger_kind: 'pipeline',
  },
  alert_snapshot: {
    alert_id: 'alert-123',
    trace_id: 'trace-123',
    source: 'prometheus',
    alert_name: 'LatencyHigh',
    severity: 'P1',
    message: 'latency high',
    trigger_time: '2026-04-30T00:00:00Z',
    fingerprint: 'fp-1',
    status: 'firing',
    labels: { env: 'prod' },
  },
  channel_snapshot: {
    id: 7,
    name: '值班飞书',
    type: 'feishu',
    enabled: true,
  },
  route_snapshot: {
    id: 9,
    name: '核心路由',
    priority: 1,
    enabled: true,
    channel_ids: [7],
  },
  rendered_payload_snapshot: {
    title: 'LatencyHigh',
    content: 'timeout from upstream',
  },
  last_attempt_at: '2026-04-30T00:03:00Z',
  created_at: '2026-04-30T00:00:00Z',
  updated_at: '2026-04-30T00:03:00Z',
  attempts: [
    {
      id: 1,
      attempt_number: 1,
      result: 'failed',
      retryable: true,
      error_message: 'timeout from upstream',
      duration_ms: 1500,
      trigger_kind: 'pipeline',
      created_at: '2026-04-30T00:01:00Z',
    },
  ],
}

const renderDeliveries = async (initialEntry = '/deliveries?alert_id=alert-123') => {
  const view = render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/deliveries" element={<Deliveries />} />
      </Routes>
    </MemoryRouter>
  )

  await act(async () => {
    await Promise.resolve()
  })

  return view
}

describe('Deliveries page', () => {
  beforeEach(() => {
    useUserStore.setState({ user: baseUser, token: 'token' })
    deliveryApiState.list.mockReset()
    deliveryApiState.get.mockReset()
    deliveryApiState.retry.mockReset()
    deliveryApiState.replay.mockReset()
    deliveryApiState.list.mockResolvedValue({
      list: [baseDelivery],
      total: 1,
    })
    deliveryApiState.get.mockResolvedValue({
      ...baseDelivery,
      attempts: [
        ...baseDelivery.attempts,
        {
          id: 2,
          attempt_number: 2,
          result: 'failed',
          retryable: true,
          error_message: 'still failing',
          duration_ms: 1600,
          trigger_kind: 'pipeline',
          created_at: '2026-04-30T00:02:00Z',
        },
      ],
    })
    deliveryApiState.retry.mockResolvedValue({
      recovery_id: 91,
      action: 'retry',
      status: 'succeeded',
      original_delivery_id: 11,
      result_delivery_id: 44,
      error_message: '',
    })
    deliveryApiState.replay.mockResolvedValue({
      recovery_id: 92,
      action: 'replay',
      status: 'failed',
      original_delivery_id: 11,
      error_message: 'route changed',
    })
  })

  it('loads initial filters from alert_id query', async () => {
    await renderDeliveries()

    await waitFor(() => {
      expect(deliveryApiState.list).toHaveBeenCalledWith({
        alert_id: 'alert-123',
        limit: 20,
        offset: 0,
      })
    })

    expect(await screen.findByDisplayValue('alert-123')).toBeInTheDocument()
    expect(await screen.findByText('alert_id=alert-123')).toBeInTheDocument()
  })

  it('shows failure evidence for viewer without recovery actions', async () => {
    await renderDeliveries()

    fireEvent.click(await screen.findByRole('button', { name: '查看证据' }))

    expect(await screen.findByText('投递证据')).toBeInTheDocument()
    await waitFor(() => {
      expect(deliveryApiState.get).toHaveBeenCalledWith(11)
    })
    expect(await screen.findByText('attempts')).toBeInTheDocument()
    expect(await screen.findByText('rendered_payload_snapshot')).toBeInTheDocument()
    expect(await screen.findByText('still failing')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '重试' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '重放' })).not.toBeInTheDocument()
  })

  it('allows operator to open recovery modal and requires a reason', async () => {
    useUserStore.setState({ user: operatorUser, token: 'token' })

    await renderDeliveries()

    fireEvent.click(await screen.findByRole('button', { name: '重试' }))

    expect(await screen.findByText('重试失败投递')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '确认重试' }))

    expect(await screen.findByText('请填写恢复原因')).toBeInTheDocument()
    expect(deliveryApiState.retry).not.toHaveBeenCalled()
  })

  it(
    'refreshes list and detail after successful retry and shows recovery feedback',
    async () => {
    useUserStore.setState({ user: operatorUser, token: 'token' })

    await renderDeliveries()

    fireEvent.click(await screen.findByRole('button', { name: '查看证据' }))
    await waitFor(() => {
      expect(deliveryApiState.get).toHaveBeenCalledWith(11)
    })

    const initialListCalls = deliveryApiState.list.mock.calls.length
    const initialGetCalls = deliveryApiState.get.mock.calls.length

    fireEvent.click(await screen.findByRole('button', { name: '重试' }))
    fireEvent.change(screen.getByPlaceholderText('说明为什么需要执行这次恢复，原因会进入后端审计记录'), {
      target: { value: 'upstream fixed, retry now' },
    })
    fireEvent.click(screen.getByRole('button', { name: '确认重试' }))

    await waitFor(() => {
      expect(deliveryApiState.retry).toHaveBeenCalledWith(11, {
        reason: 'upstream fixed, retry now',
      })
    })

    await waitFor(() => {
      expect(deliveryApiState.list.mock.calls.length).toBeGreaterThan(initialListCalls)
      expect(deliveryApiState.get.mock.calls.length).toBeGreaterThan(initialGetCalls)
    })

    expect(await screen.findByText('恢复结果: retry / succeeded')).toBeInTheDocument()
    expect(await screen.findByText('recovery_id=91')).toBeInTheDocument()
    expect(await screen.findByText('resulting_delivery_id=44')).toBeInTheDocument()
    },
    10000
  )
})
