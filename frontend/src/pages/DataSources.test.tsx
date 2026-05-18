import React from 'react'
import { act, render, screen } from '@testing-library/react'
import { DataSources } from './DataSources'
import { useUserStore } from '../stores/userStore'
import type { DataSource, User } from '../types'

vi.mock('../components/CodeEditor', () => ({
  CodeEditor: ({ value }: { value?: string }) => <textarea readOnly value={value ?? ''} />,
}))

vi.mock('../components', () => ({
  useToast: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarn: vi.fn(),
    showInfo: vi.fn(),
  }),
  PermissionNotice: ({ title }: { title: string }) => <div>{title}</div>,
}))

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client')
  return {
    ...actual,
    dataSourceApi: {
      get: vi.fn(),
    },
  }
})

const configStoreState = vi.hoisted(() => ({
  dataSources: [] as DataSource[],
  dataSourcesLoading: false,
  fetchDataSources: vi.fn(),
  createDataSource: vi.fn(),
  updateDataSource: vi.fn(),
  deleteDataSource: vi.fn(),
  toggleDataSource: vi.fn(),
  previewDataSource: vi.fn(),
}))

vi.mock('../stores/configStore', () => ({
  useConfigStore: () => configStoreState,
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

const sampleDataSource: DataSource = {
  id: 1,
  name: 'prometheus',
  display_name: 'Prometheus',
  api_key: 'secret-key',
  input_template: '{{ . }}',
  output_template: '{{ .message }}',
  group_by_labels: [],
  enabled: true,
  created_at: '2026-04-12T00:00:00Z',
  updated_at: '2026-04-12T00:00:00Z',
}

const renderDataSources = async () => {
  const view = render(<DataSources />)
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0))
  })
  return view
}

describe('DataSources page permissions', () => {
  beforeEach(() => {
    configStoreState.fetchDataSources.mockClear()
    configStoreState.dataSources = [sampleDataSource]
    useUserStore.setState({ user: null, token: null })
  })

  it('shows read-only config view for viewer', async () => {
    useUserStore.setState({ user: baseUser, token: 'token' })

    await renderDataSources()

    expect(await screen.findByText('当前角色可查看配置，但不能修改')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '新建数据源' })).not.toBeInTheDocument()
    expect(await screen.findByText('只读')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '编辑' })).not.toBeInTheDocument()
  })

  it('shows write controls for admin', async () => {
    useUserStore.setState({
      user: { ...baseUser, role: 'admin' },
      token: 'token',
    })

    await renderDataSources()

    expect(await screen.findByRole('button', { name: /新建数据源/ })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /编辑/ })).toBeInTheDocument()
    expect(screen.queryByText('当前角色可查看配置，但不能修改')).not.toBeInTheDocument()
  })
})
