import axios from 'axios'
import type {
  Alert,
  GroupedActiveAlert,
  DataSource,
  DataSourcePreviewRequest,
  DataSourcePreviewResponse,
  Channel,
  Delivery,
  DeliveryFilters,
  DeliveryListResponse,
  DeliveryRecoveryRequest,
  DeliveryRecoveryResult,
  RouteRule,
  SmtpConfig,
  SilenceRule,
} from '../types'

export const getApiErrorMessage = (error: unknown, fallback: string): string => {
  if (
    error &&
    typeof error === 'object' &&
    'response' in error &&
    error.response &&
    typeof error.response === 'object' &&
    'data' in error.response &&
    error.response.data &&
    typeof error.response.data === 'object' &&
    'error' in error.response.data &&
    typeof error.response.data.error === 'string'
  ) {
    return error.response.data.error
  }

  if (error instanceof Error && error.message) {
    return error.message
  }

  return fallback
}

const unwrapData = <T>(promise: Promise<unknown>) => promise as Promise<T>

const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// ============ Alert API ============

export const alertApi = {
  list: (params?: {
    page?: number
    page_size?: number
    severity?: string
    source?: string
    status?: string
    start_time?: string
    end_time?: string
    label_selector?: string
  }) => apiClient.get<{ list: Alert[]; total: number }>('/alerts', { params }),

  get: (id: string) => apiClient.get<Alert>(`/alerts/${id}`),

  ack: (id: string, data: { comment: string }) =>
    apiClient.post<Alert>(`/alerts/${id}/ack`, data),

  quickSilence: (id: string, data: { duration: number }) =>
    apiClient.post(`/alerts/${id}/quick-silence`, data),

  stats: () => apiClient.get<{
    total: number
    firing: number
    acked: number
    silenced: number
    by_severity: Record<string, number>
    trend: Array<{ time: string; count: number }>
  }>('/alerts/stats'),

  active: () => apiClient.get<Alert[]>('/alerts/active'),

  activeGrouped: () => apiClient.get<GroupedActiveAlert[]>('/alerts/active?grouped=true'),

  deliveries: (id: string) => apiClient.get(`/alerts/${id}/deliveries`).then(r => r.data as Delivery[]),

  batchAck: (data: { alert_ids: string[]; comment: string }) =>
    client.post('/alerts/batch-ack', data).then((res) => res.data),

  batchSilence: (data: { alert_ids: string[]; duration: number }) =>
    client.post('/alerts/batch-silence', data).then((res) => res.data),
}

// ============ DataSource API ============

export const dataSourceApi = {
  list: () => apiClient.get<DataSource[]>('/datasources'),

  get: (id: number) => apiClient.get<DataSource>(`/datasources/${id}`),

  create: (data: Partial<DataSource>) =>
    apiClient.post<DataSource>('/datasources', data),

  update: (id: number, data: Partial<DataSource>) =>
    apiClient.put<DataSource>(`/datasources/${id}`, data),

  delete: (id: number) => apiClient.delete(`/datasources/${id}`),

  toggle: (id: number, enabled: boolean) =>
    apiClient.patch(`/datasources/${id}/toggle`, { enabled }),

  preview: (data: DataSourcePreviewRequest) =>
    apiClient.post<DataSourcePreviewResponse>('/datasources/preview', data),
}

// ============ Channel API ============

export const channelApi = {
  list: () => apiClient.get<Channel[]>('/channels'),

  get: (id: number) => apiClient.get<Channel>(`/channels/${id}`),

  create: (data: Partial<Channel>) =>
    apiClient.post<Channel>('/channels', data),

  update: (id: number, data: Partial<Channel>) =>
    apiClient.put<Channel>(`/channels/${id}`, data),

  delete: (id: number) => apiClient.delete(`/channels/${id}`),

  toggle: (id: number, enabled: boolean) =>
    apiClient.patch(`/channels/${id}/toggle`, { enabled }),

  test: (id: number, recipients?: string[]) =>
    apiClient.post(`/channels/${id}/test`, recipients?.length ? { recipients } : {}),
}

// ============ SMTP Config API ============

export const smtpConfigApi = {
  get: () => apiClient.get<SmtpConfig>('/smtp-config'),
  update: (data: Partial<SmtpConfig>) => apiClient.put<SmtpConfig>('/smtp-config', data),
  test: (recipients: string[]) => apiClient.post('/smtp-config/test', { recipients }),
}

// ============ RouteRule API ============

export const routeRuleApi = {
  list: () => apiClient.get<RouteRule[]>('/routes'),

  get: (id: number) => apiClient.get<RouteRule>(`/routes/${id}`),

  create: (data: Partial<RouteRule>) =>
    apiClient.post<RouteRule>('/routes', data),

  update: (id: number, data: Partial<RouteRule>) =>
    apiClient.put<RouteRule>(`/routes/${id}`, data),

  delete: (id: number) => apiClient.delete(`/routes/${id}`),

  reorder: (ids: number[]) => apiClient.post('/routes/reorder', { ids }),
}

// ============ SilenceRule API ============

export const silenceRuleApi = {
  list: (params?: { status?: 'active' | 'expired' }) =>
    apiClient.get<SilenceRule[]>('/silences', { params }),

  get: (id: number) => apiClient.get<SilenceRule>(`/silences/${id}`),

  create: (data: Partial<SilenceRule>) =>
    apiClient.post<SilenceRule>('/silences', data),

  update: (id: number, data: Partial<SilenceRule>) =>
    apiClient.put<SilenceRule>(`/silences/${id}`, data),

  delete: (id: number) => apiClient.delete(`/silences/${id}`),

  createFromAlert: (alertId: string, data: { duration: number }) =>
    apiClient.post<SilenceRule>(`/silences/from-alert/${alertId}`, data),
}

// ============ Delivery API ============

export const deliveryApi = {
  list: (params?: DeliveryFilters) =>
    unwrapData<DeliveryListResponse>(apiClient.get<DeliveryListResponse>('/deliveries', { params })),

  get: (id: number) => unwrapData<Delivery>(apiClient.get<Delivery>(`/deliveries/${id}`)),

  retry: (id: number, data: DeliveryRecoveryRequest) =>
    unwrapData<DeliveryRecoveryResult>(
      apiClient.post<DeliveryRecoveryResult>(`/deliveries/${id}/retry`, data)
    ),

  replay: (id: number, data: DeliveryRecoveryRequest) =>
    unwrapData<DeliveryRecoveryResult>(
      apiClient.post<DeliveryRecoveryResult>(`/deliveries/${id}/replay`, data)
    ),
}

// ============ Metrics API (OPER-03) ============

export interface MetricsResponse {
  period: string
  webhook_ingest_total: number
  notification_send_success_total: number
  notification_send_failure_total: number
  notification_retry_total: number
  notification_terminal_failure_total: number
}

export const metricsApi = {
  get: (period?: string) => apiClient.get<MetricsResponse>('/metrics', { params: { period } }),
}

// ============ Channel Health API (OPER-02) ============

export interface ChannelHealthResponse {
  channel_id: number
  channel_name: string
  period: string
  total_deliveries: number
  successful: number
  failed: number
  success_rate: number
  last_failure?: {
    delivery_id: number
    error_message: string
    failed_at: string
  }
}

export const channelHealthApi = {
  get: (channelId: number, period?: string) =>
    apiClient.get<ChannelHealthResponse>(`/channels/${channelId}/health`, { params: { period } }),
}

export default apiClient
