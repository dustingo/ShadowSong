import React, { useEffect, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Drawer,
  Form,
  Input,
  Modal,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import { CopyOutlined, DeleteOutlined, EditOutlined, EyeOutlined, KeyOutlined, PlusOutlined } from '@ant-design/icons'
import { CodeEditor } from '../components/CodeEditor'
import { dataSourceApi } from '../api/client'
import { useConfigStore } from '../stores/configStore'
import type { DataSource, DataSourcePreviewResponse } from '../types'

const { Paragraph, Text } = Typography
const { TextArea } = Input

const defaultPreviewPayload = JSON.stringify(
  {
    status: 'firing',
    labels: {
      alertname: 'ServerLatencyHigh',
      severity: 'warning',
      instance: 'game-01',
    },
    annotations: {
      summary: 'Latency above threshold',
      runbook: 'https://runbook.internal/game-latency',
    },
    summary: 'raw summary from webhook',
    value: 187,
    timestamp: '2026-04-10T07:00:00Z',
  },
  null,
  2
)

const generateApiKey = (): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = 'ds_'
  for (let i = 0; i < 32; i += 1) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

const maskApiKey = (key?: string): string => {
  if (!key) return '-'
  if (key.length <= 8) return '****'
  return `${key.substring(0, 8)}****${key.substring(key.length - 4)}`
}

const formatJson = (value: unknown): string => JSON.stringify(value, null, 2)

export const DataSources: React.FC = () => {
  const {
    dataSources,
    dataSourcesLoading,
    fetchDataSources,
    createDataSource,
    updateDataSource,
    deleteDataSource,
    toggleDataSource,
    previewDataSource,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingDataSource, setEditingDataSource] = useState<DataSource | null>(null)
  const [currentApiKey, setCurrentApiKey] = useState('')
  const [previewDrawerVisible, setPreviewDrawerVisible] = useState(false)
  const [previewPayload, setPreviewPayload] = useState(defaultPreviewPayload)
  const [previewResult, setPreviewResult] = useState<DataSourcePreviewResponse | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchDataSources()
  }, [fetchDataSources])

  useEffect(() => {
    if (!editingDataSource) {
      return
    }

    form.setFieldsValue({
      ...editingDataSource,
      group_by_labels: editingDataSource.group_by_labels?.join(', '),
      deduplicate_enabled: editingDataSource.deduplicate_enabled === true,
      deduplicate_window: Number(editingDataSource.deduplicate_window) || 3600,
      group_enabled: editingDataSource.group_enabled === true,
      group_window: Number(editingDataSource.group_window) || 300,
    })
  }, [editingDataSource, form])

  const closeEditor = () => {
    setModalVisible(false)
    setEditingDataSource(null)
    setCurrentApiKey('')
    setPreviewResult(null)
    setPreviewDrawerVisible(false)
  }

  const handleCreate = () => {
    setEditingDataSource(null)
    setCurrentApiKey('')
    setPreviewResult(null)
    setPreviewPayload(defaultPreviewPayload)
    form.resetFields()
    form.setFieldsValue({
      enabled: true,
      group_by_labels: '',
      deduplicate_enabled: true,
      deduplicate_window: 3600,
      group_enabled: false,
      group_window: 300,
    })
    setModalVisible(true)
  }

  const handleEdit = async (record: DataSource) => {
    const fullData = (await dataSourceApi.get(record.id)) as unknown as DataSource
    setEditingDataSource(fullData)
    setCurrentApiKey(fullData.api_key || '')
    setPreviewResult(null)
    setPreviewPayload(defaultPreviewPayload)
    setModalVisible(true)
  }

  const handleDelete = (record: DataSource) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除数据源 "${record.display_name}" 吗？`,
      onOk: async () => {
        try {
          await deleteDataSource(record.id)
          message.success('删除成功')
        } catch {
          message.error('删除失败')
        }
      },
    })
  }

  const handleToggle = async (record: DataSource) => {
    try {
      await toggleDataSource(record.id, !record.enabled)
      message.success(record.enabled ? '已禁用' : '已启用')
    } catch {
      message.error('操作失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const payload = {
        ...values,
        api_key: currentApiKey,
        group_by_labels: values.group_by_labels
          ? values.group_by_labels
              .split(',')
              .map((item: string) => item.trim())
              .filter(Boolean)
          : [],
      }

      if (editingDataSource) {
        await updateDataSource(editingDataSource.id, payload)
        message.success('更新成功')
      } else {
        await createDataSource(payload)
        message.success('创建成功')
      }

      closeEditor()
    } catch {
      // validation handled by form
    }
  }

  const handlePreview = async () => {
    try {
      const values = await form.validateFields(['name', 'input_template', 'output_template'])
      const samplePayload = JSON.parse(previewPayload)

      setPreviewLoading(true)
      const result = await previewDataSource({
        datasource_id: editingDataSource?.id,
        source_name: values.name || editingDataSource?.name || 'preview',
        input_template: values.input_template,
        output_template: values.output_template,
        sample_payload: samplePayload,
      })

      setPreviewResult(result)
      setPreviewDrawerVisible(true)
      message.success('模板预览已更新')
    } catch (error: any) {
      const errorMessage =
        error instanceof SyntaxError
          ? `JSON 格式错误: ${error.message}`
          : error?.response?.data?.error || error?.message || '模板预览失败'
      message.error(errorMessage)
    } finally {
      setPreviewLoading(false)
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <Tag>{name}</Tag>,
    },
    {
      title: '显示名称',
      dataIndex: 'display_name',
      key: 'display_name',
    },
    {
      title: 'Webhook 地址',
      dataIndex: 'name',
      key: 'webhook',
      render: (name: string) => <Text copyable={{ text: `/webhook/${name}` }} code>/webhook/{name}</Text>,
    },
    {
      title: 'API Key',
      dataIndex: 'api_key',
      key: 'api_key',
      render: (key: string) => (
        <Space>
          <Text code>{maskApiKey(key)}</Text>
          {key && (
            <Button type="text" size="small" icon={<CopyOutlined />} onClick={() => navigator.clipboard.writeText(key)} />
          )}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => <Tag color={enabled ? 'green' : 'default'}>{enabled ? '已启用' : '已禁用'}</Tag>,
    },
    {
      title: '最近触发',
      dataIndex: 'last_trigger_at',
      key: 'last_trigger_at',
      render: (value?: string) => (value ? new Date(value).toLocaleString() : '-'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: DataSource) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button type="link" size="small" onClick={() => handleToggle(record)}>
            {record.enabled ? '禁用' : '启用'}
          </Button>
          <Button type="link" danger size="small" icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="数据源管理"
        extra={
          <Space>
            <Text type="secondary" style={{ fontSize: 12 }}>
              接收 Webhook、标准化事件，再按 output template 渲染通知
            </Text>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建数据源
            </Button>
          </Space>
        }
      >
        <Table columns={columns} dataSource={dataSources} rowKey="id" loading={dataSourcesLoading} />
      </Card>

      <Modal
        title={editingDataSource ? '编辑数据源' : '新建数据源'}
        open={modalVisible}
        onCancel={closeEditor}
        width={960}
        destroyOnClose
        footer={[
          <Button key="cancel" onClick={closeEditor}>
            取消
          </Button>,
          <Button
            key="preview"
            icon={<EyeOutlined />}
            loading={previewLoading}
            onClick={handlePreview}
          >
            预览模板
          </Button>,
          <Button key="submit" type="primary" onClick={handleSubmit}>
            保存
          </Button>,
        ]}
      >
        <Form form={form} layout="vertical" key={editingDataSource?.id || 'new'}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="唯一标识，如 prometheus" disabled={!!editingDataSource} />
          </Form.Item>

          <Form.Item name="display_name" label="显示名称" rules={[{ required: true, message: '请输入显示名称' }]}>
            <Input placeholder="友好显示名称" />
          </Form.Item>

          <Form.Item
            label={
              <Space>
                <KeyOutlined />
                <span>API Key</span>
              </Space>
            }
            extra="用于 Webhook 安全校验。留空会拒绝该数据源的所有请求。"
          >
            <Space.Compact style={{ width: '100%' }}>
              <Input
                value={currentApiKey}
                onChange={(event) => setCurrentApiKey(event.target.value)}
                placeholder="留空则拒绝所有请求，或点击生成按钮"
                allowClear
              />
              <Button type="primary" icon={<KeyOutlined />} onClick={() => setCurrentApiKey(generateApiKey())}>
                生成
              </Button>
            </Space.Compact>
          </Form.Item>

          <Card size="small" style={{ marginBottom: 16, background: '#fafafa' }}>
            <Space direction="vertical" size={8} style={{ width: '100%' }}>
              <Text strong>模板字段契约</Text>
              <Text>{'旧模板继续使用顶层字段：`alert_name`、`severity`、`message`、`source`、`status`、`trigger_time`、`labels`。'}</Text>
              <Text>{'其中 `severity` / `severity_code` 是系统标准化后的等级：`critical -> P0`、`warning/error -> P1`、`info -> P2`、`debug -> P3`。'}</Text>
              <Text>{'如果你想判断原始 webhook 值，直接用 `severity_raw` 或 `event.severity`。原始 webhook JSON 也会完整暴露在 `event` 上。'}</Text>
              <Paragraph style={{ marginBottom: 0 }}>
                标准化等级示例：
                <br />
                <Text code>{`{"title":"[{{.severity_code}}] {{.alert_name}}","content":"{{default .event.annotations.runbook \"无 runbook\"}}"}`}</Text>
              </Paragraph>
              <Paragraph style={{ marginBottom: 0 }}>
                原始 severity 示例：
                <br />
                <Text code>{`{{ if eq .severity_raw "critical" }}严重告警{{ else if eq .severity_raw "warning" }}一般告警{{ else }}提示信息{{ end }}`}</Text>
              </Paragraph>
            </Space>
          </Card>

          <Form.Item name="group_by_labels" label="分组 Labels">
            <Input placeholder="用逗号分隔，如: instance, env" />
          </Form.Item>

          <Card size="small" style={{ marginBottom: 16 }}>
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <Text strong>去重 / 聚合配置</Text>
              <Form.Item name="deduplicate_enabled" label="启用去重" valuePropName="checked" style={{ marginBottom: 0 }}>
                <Switch />
              </Form.Item>
              <Form.Item name="deduplicate_window" label="去重窗口" style={{ marginBottom: 0 }}>
                <Input type="number" placeholder="3600" addonAfter="秒" style={{ width: 220 }} />
              </Form.Item>
              <Form.Item name="group_enabled" label="启用分组" valuePropName="checked" style={{ marginBottom: 0 }}>
                <Switch />
              </Form.Item>
              <Form.Item name="group_window" label="分组窗口" style={{ marginBottom: 0 }}>
                <Input type="number" placeholder="300" addonAfter="秒" style={{ width: 220 }} />
              </Form.Item>
            </Space>
          </Card>

          <Form.Item name="input_template" label="输入模板 (Go Template)" rules={[{ required: true, message: '请输入输入模板' }]}>
            <CodeEditor
              height={180}
              value={form.getFieldValue('input_template') || ''}
              onChange={(value) => form.setFieldsValue({ input_template: value })}
              language="go"
            />
          </Form.Item>

          <Form.Item name="output_template" label="输出模板 (Go Template)" rules={[{ required: true, message: '请输入输出模板' }]}>
            <CodeEditor
              height={180}
              value={form.getFieldValue('output_template') || ''}
              onChange={(value) => form.setFieldsValue({ output_template: value })}
              language="go"
            />
          </Form.Item>

          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            message="预览说明"
            description="点击“预览模板”会把当前编辑中的 input/output template 和样例 JSON 一起发到后端，用真实通知渲染契约返回规范化告警与最终 title/content。"
          />

          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title="模板预览"
        placement="right"
        width={720}
        open={previewDrawerVisible}
        onClose={() => setPreviewDrawerVisible(false)}
        extra={
          <Button type="primary" loading={previewLoading} onClick={handlePreview}>
            重新预览
          </Button>
        }
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Text strong>样例 Webhook JSON</Text>
            <TextArea rows={12} value={previewPayload} onChange={(event) => setPreviewPayload(event.target.value)} />
          </div>

          {previewResult && (
            <>
              <Card size="small" title="上下文键预览">
                <Space direction="vertical" size={8} style={{ width: '100%' }}>
                  <Text>Top-level: {previewResult.context_preview.top_level_keys.join(', ')}</Text>
                  <Text>event: {previewResult.context_preview.event_keys.join(', ') || '(empty)'}</Text>
                  <Text>alert: {previewResult.context_preview.alert_keys.join(', ')}</Text>
                  <Text>labels: {previewResult.context_preview.label_keys.join(', ') || '(empty)'}</Text>
                </Space>
              </Card>

              <Card size="small" title="渲染结果">
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  <div>
                    <Text strong>Title</Text>
                    <pre style={{ background: '#f5f5f5', padding: 12, whiteSpace: 'pre-wrap' }}>
                      {previewResult.rendered.title}
                    </pre>
                  </div>
                  <div>
                    <Text strong>Content</Text>
                    <pre style={{ background: '#f5f5f5', padding: 12, whiteSpace: 'pre-wrap' }}>
                      {previewResult.rendered.content}
                    </pre>
                  </div>
                </Space>
              </Card>

              <Card size="small" title="规范化告警">
                <pre style={{ background: '#f5f5f5', padding: 12, overflow: 'auto' }}>
                  {formatJson(previewResult.normalized_alert)}
                </pre>
              </Card>
            </>
          )}
        </Space>
      </Drawer>
    </div>
  )
}
