import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { Deliveries } from './Deliveries'
import { useUserStore } from '../stores/userStore'
import type { Delivery, User } from '../types'

const deliveryApiState = vi.hoisted(() => ({
  list: vi.fn(),
  get: vi.fn(),
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
})
