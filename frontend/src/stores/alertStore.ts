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
      .then((res) => {
        const data = (res as unknown as { data: GroupedActiveAlert[] }).data ?? res as unknown as GroupedActiveAlert[]
        set({ groupedActiveAlerts: data, groupedActiveLoading: false })
      })
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
    get().fetchStats()
  },

  quickSilence: async (id, duration) => {
    await alertApi.quickSilence(id, { duration })
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.alert_id === id ? { ...a, status: 'silenced' } : a
      ),
      activeAlerts: state.activeAlerts.filter((a) => a.alert_id !== id),
    }))
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
