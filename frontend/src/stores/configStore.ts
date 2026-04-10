import { create } from 'zustand'
import type {
  DataSource,
  DataSourcePreviewRequest,
  DataSourcePreviewResponse,
  Channel,
  RouteRule,
  SilenceRule,
  OnDuty,
} from '../types'
import { dataSourceApi, channelApi, routeRuleApi, silenceRuleApi, onDutyApi } from '../api/client'

interface ConfigState {
  dataSources: DataSource[]
  dataSourcesLoading: boolean
  channels: Channel[]
  channelsLoading: boolean
  routeRules: RouteRule[]
  routeRulesLoading: boolean
  silenceRules: SilenceRule[]
  silenceRulesLoading: boolean
  onDutyList: OnDuty[]
  currentOnDuty: OnDuty[]
  onDutyLoading: boolean
  fetchDataSources: () => Promise<void>
  createDataSource: (data: Partial<DataSource>) => Promise<void>
  updateDataSource: (id: number, data: Partial<DataSource>) => Promise<void>
  deleteDataSource: (id: number) => Promise<void>
  toggleDataSource: (id: number, enabled: boolean) => Promise<void>
  previewDataSource: (data: DataSourcePreviewRequest) => Promise<DataSourcePreviewResponse>
  fetchChannels: () => Promise<void>
  createChannel: (data: Partial<Channel>) => Promise<void>
  updateChannel: (id: number, data: Partial<Channel>) => Promise<void>
  deleteChannel: (id: number) => Promise<void>
  toggleChannel: (id: number, enabled: boolean) => Promise<void>
  testChannel: (id: number) => Promise<void>
  fetchRouteRules: () => Promise<void>
  createRouteRule: (data: Partial<RouteRule>) => Promise<void>
  updateRouteRule: (id: number, data: Partial<RouteRule>) => Promise<void>
  deleteRouteRule: (id: number) => Promise<void>
  reorderRouteRules: (ids: number[]) => Promise<void>
  fetchSilenceRules: (params?: { status?: 'active' | 'expired' }) => Promise<void>
  createSilenceRule: (data: Partial<SilenceRule>) => Promise<void>
  updateSilenceRule: (id: number, data: Partial<SilenceRule>) => Promise<void>
  deleteSilenceRule: (id: number) => Promise<void>
  createSilenceFromAlert: (alertId: string, duration: number) => Promise<void>
  fetchOnDuty: () => Promise<void>
  createOnDuty: (data: Partial<OnDuty>) => Promise<void>
  updateOnDuty: (id: number, data: Partial<OnDuty>) => Promise<void>
  deleteOnDuty: (id: number) => Promise<void>
}

export const useConfigStore = create<ConfigState>((set, get) => ({
  dataSources: [],
  dataSourcesLoading: false,
  channels: [],
  channelsLoading: false,
  routeRules: [],
  routeRulesLoading: false,
  silenceRules: [],
  silenceRulesLoading: false,
  onDutyList: [],
  currentOnDuty: [],
  onDutyLoading: false,

  fetchDataSources: async () => {
    set({ dataSourcesLoading: true })
    try {
      const data = await dataSourceApi.list() as unknown as DataSource[]
      set({ dataSources: data, dataSourcesLoading: false })
    } catch (error) {
      set({ dataSourcesLoading: false })
      throw error
    }
  },

  createDataSource: async (data) => {
    await dataSourceApi.create(data)
    get().fetchDataSources()
  },

  updateDataSource: async (id, data) => {
    await dataSourceApi.update(id, data)
    get().fetchDataSources()
  },

  deleteDataSource: async (id) => {
    await dataSourceApi.delete(id)
    get().fetchDataSources()
  },

  toggleDataSource: async (id, enabled) => {
    await dataSourceApi.toggle(id, enabled)
    get().fetchDataSources()
  },

  previewDataSource: async (data) => dataSourceApi.preview(data) as unknown as DataSourcePreviewResponse,

  fetchChannels: async () => {
    set({ channelsLoading: true })
    try {
      const data = await channelApi.list() as unknown as Channel[]
      set({ channels: data, channelsLoading: false })
    } catch (error) {
      set({ channelsLoading: false })
      throw error
    }
  },

  createChannel: async (data) => {
    await channelApi.create(data)
    get().fetchChannels()
  },

  updateChannel: async (id, data) => {
    await channelApi.update(id, data)
    get().fetchChannels()
  },

  deleteChannel: async (id) => {
    await channelApi.delete(id)
    get().fetchChannels()
  },

  toggleChannel: async (id, enabled) => {
    await channelApi.toggle(id, enabled)
    get().fetchChannels()
  },

  testChannel: async (id) => {
    await channelApi.test(id)
  },

  fetchRouteRules: async () => {
    set({ routeRulesLoading: true })
    try {
      const data = await routeRuleApi.list() as unknown as RouteRule[]
      set({ routeRules: data.sort((a, b) => a.priority - b.priority), routeRulesLoading: false })
    } catch (error) {
      set({ routeRulesLoading: false })
      throw error
    }
  },

  createRouteRule: async (data) => {
    await routeRuleApi.create(data)
    get().fetchRouteRules()
  },

  updateRouteRule: async (id, data) => {
    await routeRuleApi.update(id, data)
    get().fetchRouteRules()
  },

  deleteRouteRule: async (id) => {
    await routeRuleApi.delete(id)
    get().fetchRouteRules()
  },

  reorderRouteRules: async (ids) => {
    await routeRuleApi.reorder(ids)
    get().fetchRouteRules()
  },

  fetchSilenceRules: async (params) => {
    set({ silenceRulesLoading: true })
    try {
      const data = await silenceRuleApi.list(params) as unknown as SilenceRule[]
      set({ silenceRules: data, silenceRulesLoading: false })
    } catch (error) {
      set({ silenceRulesLoading: false })
      throw error
    }
  },

  createSilenceRule: async (data) => {
    await silenceRuleApi.create(data)
    get().fetchSilenceRules({ status: 'active' })
  },

  updateSilenceRule: async (id, data) => {
    await silenceRuleApi.update(id, data)
    get().fetchSilenceRules({ status: 'active' })
  },

  deleteSilenceRule: async (id) => {
    await silenceRuleApi.delete(id)
    get().fetchSilenceRules({ status: 'active' })
  },

  createSilenceFromAlert: async (alertId, duration) => {
    await silenceRuleApi.createFromAlert(alertId, { duration })
    get().fetchSilenceRules({ status: 'active' })
  },

  fetchOnDuty: async () => {
    set({ onDutyLoading: true })
    try {
      const [list, current] = await Promise.all([
        onDutyApi.list() as unknown as Promise<OnDuty[]>,
        onDutyApi.current() as unknown as Promise<OnDuty[]>,
      ])
      set({ onDutyList: await list, currentOnDuty: await current, onDutyLoading: false })
    } catch (error) {
      set({ onDutyLoading: false })
      throw error
    }
  },

  createOnDuty: async (data) => {
    await onDutyApi.create(data)
    get().fetchOnDuty()
  },

  updateOnDuty: async (id, data) => {
    await onDutyApi.update(id, data)
    get().fetchOnDuty()
  },

  deleteOnDuty: async (id) => {
    await onDutyApi.delete(id)
    get().fetchOnDuty()
  },
}))
