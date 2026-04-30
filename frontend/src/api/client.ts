import axios from 'axios'
import type {
  Alert,
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
  SilenceRule,
  OnDuty,
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

  test: (id: number) => apiClient.post(`/channels/${id}/test`),
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

// ============ OnDuty API ============

export const onDutyApi = {
  list: () => apiClient.get<OnDuty[]>('/onduty'),

  get: (id: number) => apiClient.get<OnDuty>(`/onduty/${id}`),

  create: (data: Partial<OnDuty>) =>
    apiClient.post<OnDuty>('/onduty', data),

  update: (id: number, data: Partial<OnDuty>) =>
    apiClient.put<OnDuty>(`/onduty/${id}`, data),

  delete: (id: number) => apiClient.delete(`/onduty/${id}`),

  current: () => apiClient.get<OnDuty[]>('/onduty/current'),
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

export default apiClient
