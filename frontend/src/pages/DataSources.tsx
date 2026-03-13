import React, { useEffect, useState, useRef } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Switch,
  message,
  Typography,
  Drawer,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, CopyOutlined } from '@ant-design/icons'
import { useConfigStore } from '../stores/configStore'
import { dataSourceApi } from '../api/client'
import { CodeEditor } from '../components/CodeEditor'
import type { DataSource } from '../types'

const { Text } = Typography
const { TextArea } = Input

// 生成随机 API Key
const generateApiKey = (): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  const prefix = 'ds_'
  let result = prefix
  for (let i = 0; i < 32; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

// 脱敏显示 API Key
const maskApiKey = (key?: string): string => {
  if (!key) return '-'
  if (key.length <= 8) return '****'
  return key.substring(0, 8) + '****' + key.substring(key.length - 4)
}

export const DataSources: React.FC = () => {
  const {
    dataSources,
    dataSourcesLoading,
    fetchDataSources,
    createDataSource,
    updateDataSource,
    deleteDataSource,
    toggleDataSource,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingDataSource, setEditingDataSource] = useState<DataSource | null>(null)
  const [form] = Form.useForm()
  const [currentApiKey, setCurrentApiKey] = useState('')

  const [testDrawerVisible, setTestDrawerVisible] = useState(false)
  const [testPayload, setTestPayload] = useState('')
  const [testResult, setTestResult] = useState<any>(null)
  const [testLoading, setTestLoading] = useState(false)
  const apiKeyRef = useRef<any>(null)

  useEffect(() => {
    fetchDataSources()
  }, [])

  // 当表单 key 变化时（编辑模式），设置初始值
  useEffect(() => {
    if (editingDataSource && form) {
      // 确保布尔值类型正确
      form.setFieldsValue({
        ...editingDataSource,
        group_by_labels: editingDataSource.group_by_labels?.join(', '),
        deduplicate_enabled: editingDataSource.deduplicate_enabled === true,
        deduplicate_window: Number(editingDataSource.deduplicate_window) || 3600,
        group_enabled: editingDataSource.group_enabled === true,
        group_window: Number(editingDataSource.group_window) || 300,
      })
    }
  }, [editingDataSource?.id, form])

  const handleCreate = () => {
    setEditingDataSource(null)
    setCurrentApiKey('')
    form.resetFields()
    // 确保所有值都是正确的类型
    const defaultValues = {
      enabled: true,
      group_by_labels: [],
      deduplicate_enabled: true,
      deduplicate_window: 3600,
      group_enabled: false,
      group_window: 300,
    }
    form.setFieldsValue(defaultValues)
    setModalVisible(true)
  }

  const handleEdit = async (record: DataSource) => {
    // 获取完整数据（包含 api_key）
    const fullData = await dataSourceApi.get(record.id) as unknown as DataSource
    setEditingDataSource(fullData)
    setCurrentApiKey(fullData.api_key || '')
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
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  const handleToggle = async (record: DataSource) => {
    try {
      await toggleDataSource(record.id, !record.enabled)
      message.success(record.enabled ? '已禁用' : '已启用')
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleGenerateApiKey = () => {
    const newKey = generateApiKey()
    setCurrentApiKey(newKey)
    message.success('API Key 已生成')
  }

  const handleCopyApiKey = (key: string) => {
    navigator.clipboard.writeText(key)
    message.success('API Key 已复制到剪贴板')
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data = {
        ...values,
        // 使用当前输入的 API Key
        api_key: currentApiKey,
        group_by_labels: values.group_by_labels
          ? values.group_by_labels.split(',').map((s: string) => s.trim()).filter(Boolean)
          : [],
      }

      if (editingDataSource) {
        await updateDataSource(editingDataSource.id, data)
        message.success('更新成功')
      } else {
        await createDataSource(data)
        message.success('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      // Validation error
    }
  }

  const handleTest = () => {
    setTestDrawerVisible(true)
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
      render: (name: string) => (
        <Text copyable code>/webhook/{name}</Text>
      ),
    },
    {
      title: 'API Key',
      dataIndex: 'api_key',
      key: 'api_key',
      render: (key: string) => (
        <Space>
          <Text code>{maskApiKey(key)}</Text>
          {key && (
            <Button
              type="text"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => handleCopyApiKey(key)}
            />
          )}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'default'}>
          {enabled ? '已启用' : '已禁用'}
        </Tag>
      ),
    },
    {
      title: '去重',
      key: 'deduplicate',
      render: (_: any, record: DataSource) => (
        <Space direction="vertical" size={0}>
          <Tag color={record.deduplicate_enabled ? 'blue' : 'default'}>
            {record.deduplicate_enabled ? '已启用' : '已禁用'}
          </Tag>
          {record.deduplicate_enabled && record.deduplicate_window && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              {record.deduplicate_window}秒
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: '最近触发',
      dataIndex: 'last_trigger_at',
      key: 'last_trigger_at',
      render: (time: string) => time ? new Date(time).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: DataSource) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => handleToggle(record)}
          >
            {record.enabled ? '禁用' : '启用'}
          </Button>
          <Button
            type="link"
            danger
            size="small"
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
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
              用于接收外部告警系统的 Webhook
            </Text>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建数据源
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={dataSources}
          rowKey="id"
          loading={dataSourcesLoading}
        />
      </Card>

      {/* Edit/Create Modal */}
      <Modal
        title={editingDataSource ? '编辑数据源' : '新建数据源'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false)
          setCurrentApiKey('')
          setEditingDataSource(null)
        }}
        width={800}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          key={editingDataSource?.id || 'new'}
        >
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="唯一标识，如 prometheus" disabled={!!editingDataSource} />
          </Form.Item>
          <Form.Item
            name="display_name"
            label="显示名称"
            rules={[{ required: true, message: '请输入显示名称' }]}
          >
            <Input placeholder="友好显示名称" />
          </Form.Item>

          {/* API Key 配置 */}
          <Form.Item
            label={
              <Space>
                <KeyOutlined />
                <span>API Key</span>
              </Space>
            }
            extra="用于 Webhook 安全性验证，生成后可在外接系统配置使用。留空则该数据源拒绝所有请求"
          >
            <Space.Compact style={{ width: '100%' }}>
              <Input
                value={currentApiKey}
                onChange={(e) => setCurrentApiKey(e.target.value)}
                placeholder="留空则该数据源拒绝所有请求，或点击生成按钮"
                allowClear
                style={{ flex: 1 }}
              />
              <Button
                type="primary"
                icon={<KeyOutlined />}
                onClick={handleGenerateApiKey}
              >
                生成
              </Button>
            </Space.Compact>
          </Form.Item>

          <Form.Item
            name="group_by_labels"
            label="分组 Labels"
          >
            <Input placeholder="用逗号分隔，如: instance, env" />
          </Form.Item>

          {/* 去重配置 */}
          <div style={{ marginBottom: 16, padding: '12px 16px', background: '#f6f8fa', borderRadius: 6 }}>
            <Text strong style={{ display: 'block', marginBottom: 12 }}>⚡ 去重/聚合配置</Text>
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Form.Item name="deduplicate_enabled" valuePropName="checked" style={{ marginBottom: 0 }}>
                <Switch />
              </Form.Item>
              <Form.Item name="deduplicate_window" label="去重窗口" style={{ marginBottom: 0 }}>
                <Input type="number" placeholder="3600" addonAfter="秒" style={{ width: 200 }} />
              </Form.Item>
              <Text type="secondary" style={{ fontSize: 12 }}>
                相同指纹的告警在去重窗口内只会触发一次通知。默认1小时(3600秒)
              </Text>

              <div style={{ marginTop: 16, paddingTop: 16, borderTop: '1px dashed #d9d9d9' }}>
                <Form.Item name="group_enabled" valuePropName="checked" style={{ marginBottom: 0 }}>
                  <Switch />
                </Form.Item>
                <Form.Item name="group_window" label="分组窗口" style={{ marginBottom: 0 }}>
                  <Input type="number" placeholder="300" addonAfter="秒" style={{ width: 200 }} />
                </Form.Item>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  在分组窗口内的告警会聚合为一条通知。默认5分钟(300秒)
                </Text>
              </div>
            </Space>
          </div>

          <Form.Item
            name="input_template"
            label="输入模板 (Go Template)"
            rules={[{ required: true, message: '请输入输入模板' }]}
          >
            <CodeEditor
              height={150}
              value={form.getFieldValue('input_template') || ''}
              onChange={(v) => form.setFieldsValue({ input_template: v })}
              language="go"
            />
          </Form.Item>
          <Form.Item
            name="output_template"
            label="输出模板 (Go Template)"
            rules={[{ required: true, message: '请输入输出模板' }]}
          >
            <CodeEditor
              height={150}
              value={form.getFieldValue('output_template') || ''}
              onChange={(v) => form.setFieldsValue({ output_template: v })}
              language="go"
            />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* Test Drawer */}
      <Drawer
        title="模板测试"
        placement="right"
        width={600}
        open={testDrawerVisible}
        onClose={() => setTestDrawerVisible(false)}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Text strong>输入 JSON 样本:</Text>
            <TextArea
              rows={8}
              placeholder='{"status": "firing", "labels": {...}}'
              value={testPayload}
              onChange={(e) => setTestPayload(e.target.value)}
            />
          </div>
          <Button type="primary" onClick={handleTest} loading={testLoading}>
            测试输入模板
          </Button>
          {testResult && (
            <div>
              <Text strong>测试结果:</Text>
              <pre style={{ background: '#f5f5f5', padding: 12, overflow: 'auto' }}>
                {JSON.stringify(testResult, null, 2)}
              </pre>
            </div>
          )}
        </Space>
      </Drawer>
    </div>
  )
}
