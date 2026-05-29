import { create } from 'zustand'
import type { Alert, GroupedActiveAlert } from '../types'
import { alertApi } from '../api/client'

interface AlertFilters {
  severity?: string
  source?: string
  status?: string
  startTime?: string
  endTime?: string
  labelSelector?: string
}

interface AlertStats {
  total: number
  firing: number
  acked: number
  silenced: number
  by_severity: Record<string, number>
  trend: Array<{ time: string; count: number }>
}

interface AlertState {
  alerts: Alert[]
  activeAlerts: Alert[]
  groupedActiveAlerts: GroupedActiveAlert[]
  groupedActiveLoading: boolean
  stats: AlertStats | null
  filters: AlertFilters
  loading: boolean
  total: number
  page: number
  pageSize: number
  wsConnected: boolean
  fetchAlerts: (page?: number, pageSize?: number) => Promise<void>
  fetchActiveAlerts: () => Promise<void>
  fetchGroupedActiveAlerts: () => void
  fetchStats: () => Promise<void>
  setFilters: (filters: AlertFilters) => void
  ackAlert: (id: string, comment: string) => Promise<void>
  quickSilence: (id: string, duration: number) => Promise<void>
  batchAck: (ids: string[], comment: string) => Promise<{ updated: number; skipped: number; errors: string[] }>
  batchSilence: (ids: string[], duration: number) => Promise<{ updated: number; skipped: number; errors: string[] }>
  addAlert: (alert: Alert) => void
  updateAlert: (alert: Alert) => void
  setWsConnected: (connected: boolean) => void
}

export const useAlertStore = create<AlertState>((set, get) => ({
  alerts: [],
  activeAlerts: [],
  groupedActiveAlerts: [],
  groupedActiveLoading: false,
  stats: null,
  filters: {},
  loading: false,
  total: 0,
  page: 1,
  pageSize: 20,
  wsConnected: false,

  fetchAlerts: async (page = 1, pageSize = 20) => {
    set({ loading: true })
    try {
      const filters = get().filters
      const res = await alertApi.list({
        page,
        page_size: pageSize,
        severity: filters.severity,
        source: filters.source,
        status: filters.status,
        start_time: filters.startTime,
        end_time: filters.endTime,
        label_selector: filters.labelSelector,
      }) as unknown as { list: Alert[]; total: number }
      set({ alerts: res.list, total: res.total, page, pageSize, loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },

  fetchActiveAlerts: async () => {
    try {
      const res = await alertApi.active() as unknown as Alert[]
      set({ activeAlerts: res })
    } catch (error) {
      console.error('Failed to fetch active alerts:', error)
    }
  },

  fetchGroupedActiveAlerts: () => {
    set({ groupedActiveLoading: true })
    alertApi.activeGrouped()
      .then((res) => set({ groupedActiveAlerts: res as unknown as GroupedActiveAlert[], groupedActiveLoading: false }))
      .catch(() => set({ groupedActiveLoading: false }))
  },

  fetchStats: async () => {
    try {
      const res = await alertApi.stats() as unknown as AlertStats
      set({ stats: res })
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    }
  },

  setFilters: (filters) => {
    set({ filters })
    get().fetchAlerts(1)
  },

  ackAlert: async (id, comment) => {
    await alertApi.ack(id, { comment })
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.alert_id === id
          ? { ...a, status: 'acked', acked_at: new Date().toISOString(), ack_comment: comment }
          : a
      ),
      activeAlerts: state.activeAlerts.filter((a) => a.alert_id !== id),
    }))
    get().fetchGroupedActiveAlerts()
    get().fetchStats()
  },

  batchAck: async (ids, comment) => {
    const result = await alertApi.batchAck({ alert_ids: ids, comment })
    const updatedCount = (result as any).updated ?? ids.length
    const skippedCount = (result as any).skipped ?? 0
    const updatedIds = ids.slice(0, updatedCount)
    set((state) => ({
      alerts: state.alerts.map((a) =>
        updatedIds.includes(a.alert_id)
          ? { ...a, status: 'acked', acked_at: new Date().toISOString(), ack_comment: comment }
          : a
      ),
      activeAlerts: skippedCount > 0
        ? state.activeAlerts.filter((a) => updatedIds.includes(a.alert_id) === false || !ids.includes(a.alert_id))
        : state.activeAlerts.filter((a) => !ids.includes(a.alert_id)),
    }))
    get().fetchGroupedActiveAlerts()
    get().fetchStats()
    return result
  },

  batchSilence: async (ids, duration) => {
    const result = await alertApi.batchSilence({ alert_ids: ids, duration })
    const updatedCount = (result as any).updated ?? ids.length
    const skippedCount = (result as any).skipped ?? 0
    const updatedIds = ids.slice(0, updatedCount)
    set((state) => ({
      alerts: state.alerts.map((a) =>
        updatedIds.includes(a.alert_id) ? { ...a, status: 'silenced' } : a
      ),
      activeAlerts: skippedCount > 0
        ? state.activeAlerts.filter((a) => updatedIds.includes(a.alert_id) === false || !ids.includes(a.alert_id))
        : state.activeAlerts.filter((a) => !ids.includes(a.alert_id)),
    }))
    get().fetchGroupedActiveAlerts()
    get().fetchStats()
    return result
  },

  quickSilence: async (id, duration) => {
    await alertApi.quickSilence(id, { duration })
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.alert_id === id ? { ...a, status: 'silenced' } : a
      ),
      activeAlerts: state.activeAlerts.filter((a) => a.alert_id !== id),
    }))
    get().fetchGroupedActiveAlerts()
    get().fetchStats()
  },

  addAlert: (alert) => {
    set((state) => ({
      activeAlerts: [alert, ...state.activeAlerts],
    }))
    get().fetchStats()
  },

  updateAlert: (alert) => {
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.alert_id === alert.alert_id ? alert : a
      ),
      activeAlerts: state.activeAlerts.map((a) =>
        a.alert_id === alert.alert_id ? alert : a
      ),
    }))
  },

  setWsConnected: (connected) => set({ wsConnected: connected }),
}))
