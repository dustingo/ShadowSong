import React, { useEffect, useState } from 'react'
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  message,
  Divider,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined } from '@ant-design/icons'
import { PermissionNotice } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { useConfigStore } from '../stores/configStore'
import { channelApi, getApiErrorMessage } from '../api/client'
import { useUserStore } from '../stores/userStore'
import type { Channel } from '../types'

const { Option } = Select

const formatChannelConfigForForm = (channel: Channel) => {
  const config = channel.config ?? {}

  if (channel.type === 'feishu') {
    return {
      ...channel,
      config: {
        webhook_url: config.webhook_url ?? '',
        secret: config.secret ?? '',
      },
    }
  }

  if (channel.type === 'dingtalk') {
    return {
      ...channel,
      config: {
        webhook_url: config.webhook_url ?? '',
        secret: config.secret ?? '',
      },
    }
  }

  if (channel.type === 'wecom') {
    return {
      ...channel,
      config: {
        webhook_url: config.webhook_url ?? '',
      },
    }
  }

  if (channel.type === 'webhook') {
    return {
      ...channel,
      config: {
        url: config.url ?? '',
        method: config.method ?? 'POST',
        headers:
          typeof config.headers === 'string'
            ? config.headers
            : JSON.stringify(config.headers ?? {}, null, 2),
        template: config.template ?? '',
      },
    }
  }

  return channel
}

const buildChannelPayload = (values: Partial<Channel>) => {
  const config = values.config ?? {}

  if (values.type === 'webhook') {
    let headers: Record<string, string> = {}

    if (typeof config.headers === 'string' && config.headers.trim()) {
      headers = JSON.parse(config.headers) as Record<string, string>
    }

    return {
      ...values,
      config: {
        url: config.url ?? '',
        method: config.method ?? 'POST',
        headers,
        template: config.template ?? '',
      },
    }
  }

  if (values.type === 'feishu' || values.type === 'dingtalk') {
    return {
      ...values,
      config: {
        webhook_url: config.webhook_url ?? '',
        secret: config.secret ?? '',
      },
    }
  }

  if (values.type === 'wecom') {
    return {
      ...values,
      config: {
        webhook_url: config.webhook_url ?? '',
      },
    }
  }

  return values
}

export const Channels: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const {
    channels,
    channelsLoading,
    fetchChannels,
    createChannel,
    updateChannel,
    deleteChannel,
    toggleChannel,
    testChannel,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null)
  const [form] = Form.useForm()
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  useEffect(() => {
    fetchChannels()
  }, [])

  const channelTypeOptions = [
    { value: 'feishu', label: '飞书', icon: '📱' },
    { value: 'dingtalk', label: '钉钉', icon: '💬' },
    { value: 'wecom', label: '企业微信', icon: '💼' },
    { value: 'webhook', label: '自定义 Webhook', icon: '🔗' },
  ]

  const handleCreate = () => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    setEditingChannel(null)
    form.resetFields()
    form.setFieldsValue({ enabled: true, type: 'feishu' })
    setModalVisible(true)
  }

  const handleEdit = async (record: Channel) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    try {
      const fullChannel = await channelApi.get(record.id) as unknown as Channel
      const formValues = formatChannelConfigForForm(fullChannel)
      setEditingChannel(fullChannel)
      form.setFieldsValue(formValues)
      setModalVisible(true)
    } catch (error) {
      message.error(getApiErrorMessage(error, '加载渠道配置失败'))
    }
  }

  const handleDelete = (record: Channel) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除渠道 "${record.name}" 吗？`,
      onOk: async () => {
        try {
          await deleteChannel(record.id)
          message.success('删除成功')
        } catch (error) {
          message.error(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleToggle = async (record: Channel) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    try {
      await toggleChannel(record.id, !record.enabled)
      message.success(record.enabled ? '已禁用' : '已启用')
    } catch (error) {
      message.error(getApiErrorMessage(error, '操作失败'))
    }
  }

  const handleTest = async (record: Channel) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    try {
      await testChannel(record.id)
      message.success('测试消息已发送')
    } catch (error) {
      message.error(getApiErrorMessage(error, '发送失败'))
    }
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    try {
      const values = await form.validateFields()
      const payload = buildChannelPayload(values)
      const channelId = values.id || editingChannel?.id
      if (channelId) {
        await updateChannel(channelId, payload)
        message.success('更新成功')
      } else {
        await createChannel(payload)
        message.success('创建成功')
      }
      setModalVisible(false)
      setEditingChannel(null)
      form.resetFields()
    } catch (error) {
      if (error instanceof SyntaxError) {
        message.error('请求头必须是合法 JSON')
      }
    }
  }

  const getTypeLabel = (type: string) => {
    const opt = channelTypeOptions.find((o) => o.value === type)
    return opt ? `${opt.icon} ${opt.label}` : type
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag>{getTypeLabel(type)}</Tag>,
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
      title: '操作',
      key: 'action',
      render: (_: any, record: Channel) => (
        canManageConfig ? (
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
              icon={<SendOutlined />}
              onClick={() => handleTest(record)}
            >
              测试
            </Button>
            <Button type="link" size="small" onClick={() => handleToggle(record)}>
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
        ) : (
          <Tag>只读</Tag>
        )
      ),
    },
  ]

  return (
    <div>
      <Card
        title="推送渠道管理"
        extra={
          canManageConfig ? (
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建渠道
            </Button>
          ) : null
        }
      >
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="渠道配置对非 `admin` 角色保持只读，测试、启停和删除操作不会开放。"
            type="info"
          />
        )}
        <Table
          columns={columns}
          dataSource={channels}
          rowKey="id"
          loading={channelsLoading}
        />
      </Card>

      <Modal
        title={editingChannel ? '编辑渠道' : '新建渠道'}
        open={modalVisible}
        onOk={canManageConfig ? handleSubmit : undefined}
        onCancel={() => {
          setModalVisible(false)
          setEditingChannel(null)
          form.resetFields()
        }}
        width={600}
        okButtonProps={{ style: { display: canManageConfig ? 'inline-block' : 'none' } }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="渠道名称" disabled={!canManageConfig} />
          </Form.Item>
          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select placeholder="选择渠道类型" disabled={!canManageConfig}>
              {channelTypeOptions.map((opt) => (
                <Option key={opt.value} value={opt.value}>
                  {opt.icon} {opt.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Divider>配置信息</Divider>

          <Form.Item
            noStyle
            shouldUpdate={(prev, curr) => prev.type !== curr.type}
          >
            {() => {
              const type = form.getFieldValue('type')
              if (type === 'feishu') {
                return (
                  <>
                    <Form.Item
                      name={['config', 'webhook_url']}
                      label="Webhook URL"
                      rules={[{ required: true, message: '请输入 Webhook URL' }]}
                    >
                      <Input
                        placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
                        disabled={!canManageConfig}
                      />
                    </Form.Item>
                    <Form.Item name={['config', 'secret']} label="签名密钥（可选）">
                      <Input placeholder="签名密钥" disabled={!canManageConfig} />
                    </Form.Item>
                  </>
                )
              }
              if (type === 'dingtalk') {
                return (
                  <>
                    <Form.Item
                      name={['config', 'webhook_url']}
                      label="Webhook URL"
                      rules={[{ required: true, message: '请输入 Webhook URL' }]}
                    >
                      <Input
                        placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
                        disabled={!canManageConfig}
                      />
                    </Form.Item>
                    <Form.Item name={['config', 'secret']} label="加签密钥（可选）">
                      <Input placeholder="密钥" disabled={!canManageConfig} />
                    </Form.Item>
                  </>
                )
              }
              if (type === 'wecom') {
                return (
                  <>
                    <Form.Item
                      name={['config', 'webhook_url']}
                      label="Webhook URL"
                      rules={[{ required: true, message: '请输入 Webhook URL' }]}
                    >
                      <Input
                        placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
                        disabled={!canManageConfig}
                      />
                    </Form.Item>
                  </>
                )
              }
              if (type === 'webhook') {
                return (
                  <>
                    <Form.Item
                      name={['config', 'url']}
                      label="请求 URL"
                      rules={[{ required: true, message: '请输入 URL' }]}
                    >
                      <Input placeholder="https://example.com/webhook" disabled={!canManageConfig} />
                    </Form.Item>
                    <Form.Item name={['config', 'method']} label="请求方法">
                      <Select disabled={!canManageConfig}>
                        <Option value="POST">POST</Option>
                        <Option value="PUT">PUT</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item name={['config', 'headers']} label="请求头（JSON）">
                      <Input.TextArea
                        rows={2}
                        placeholder='{"Content-Type": "application/json"}'
                        disabled={!canManageConfig}
                      />
                    </Form.Item>
                    <Form.Item name={['config', 'template']} label="请求体模板">
                      <Input.TextArea
                        rows={3}
                        placeholder='{"text": "{{.message}}"}'
                        disabled={!canManageConfig}
                      />
                    </Form.Item>
                  </>
                )
              }
              return null
            }}
          </Form.Item>

          <Divider />

          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch disabled={!canManageConfig} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
