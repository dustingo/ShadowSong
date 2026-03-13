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
import { useConfigStore } from '../stores/configStore'
import { channelApi } from '../api/client'
import type { Channel } from '../types'

const { Option } = Select

export const Channels: React.FC = () => {
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
    setEditingChannel(null)
    form.resetFields()
    form.setFieldsValue({ enabled: true, type: 'feishu' })
    setModalVisible(true)
  }

  const handleEdit = async (record: Channel) => {
    // 获取完整配置（包含原始 webhook_url）
    const fullChannel = await channelApi.get(record.id).then(res => res.data)
    setEditingChannel(fullChannel)
    form.setFieldsValue(fullChannel)
    setModalVisible(true)
  }

  const handleDelete = (record: Channel) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除渠道 "${record.name}" 吗？`,
      onOk: async () => {
        try {
          await deleteChannel(record.id)
          message.success('删除成功')
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  const handleToggle = async (record: Channel) => {
    try {
      await toggleChannel(record.id, !record.enabled)
      message.success(record.enabled ? '已禁用' : '已启用')
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleTest = async (record: Channel) => {
    try {
      await testChannel(record.id)
      message.success('测试消息已发送')
    } catch (error) {
      message.error('发送失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const channelId = values.id || editingChannel?.id
      if (channelId) {
        await updateChannel(channelId, values)
        message.success('更新成功')
      } else {
        await createChannel(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      setEditingChannel(null)
    } catch (error) {
      // Validation error
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
        title="推送渠道管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新建渠道
          </Button>
        }
      >
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
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false)
          setEditingChannel(null)
          form.resetFields()
        }}
        width={600}
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
            <Input placeholder="渠道名称" />
          </Form.Item>
          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true, message: '请选择类型' }]}
          >
            <Select placeholder="选择渠道类型">
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
                      <Input placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx" />
                    </Form.Item>
                    <Form.Item name={['config', 'sign_key']} label="签名密钥（可选）">
                      <Input placeholder="签名密钥" />
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
                      <Input placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" />
                    </Form.Item>
                    <Form.Item name={['config', 'secret']} label="加签密钥（可选）">
                      <Input placeholder="密钥" />
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
                      <Input placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" />
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
                      <Input placeholder="https://example.com/webhook" />
                    </Form.Item>
                    <Form.Item name={['config', 'method']} label="请求方法">
                      <Select>
                        <Option value="POST">POST</Option>
                        <Option value="PUT">PUT</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item name={['config', 'headers']} label="请求头（JSON）">
                      <Input.TextArea rows={2} placeholder='{"Content-Type": "application/json"}' />
                    </Form.Item>
                    <Form.Item name={['config', 'body_template']} label="请求体模板">
                      <Input.TextArea rows={3} placeholder='{"text": "{{.message}}"}' />
                    </Form.Item>
                  </>
                )
              }
              return null
            }}
          </Form.Item>

          <Divider />

          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
