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
  Badge,
  InputNumber,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, HolderOutlined } from '@ant-design/icons'
import { useConfigStore } from '../stores/configStore'
import type { RouteRule } from '../types'

const { Option } = Select

export const RouteRules: React.FC = () => {
  const {
    routeRules,
    routeRulesLoading,
    channels,
    fetchRouteRules,
    fetchChannels,
    createRouteRule,
    updateRouteRule,
    deleteRouteRule,
    reorderRouteRules,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<RouteRule | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchRouteRules()
    fetchChannels()
  }, [])

  const handleCreate = () => {
    setEditingRule(null)
    form.resetFields()
    form.setFieldsValue({
      enabled: true,
      priority: routeRules.length + 1,
      severities: [],
      sources: [],
      channel_ids: [],
    })
    setModalVisible(true)
  }

  const handleEdit = (record: RouteRule) => {
    setEditingRule(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = (record: RouteRule) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除规则 "${record.name}" 吗？`,
      onOk: async () => {
        try {
          await deleteRouteRule(record.id)
          message.success('删除成功')
        } catch (error) {
          message.error('删除失败')
        }
      },
    })
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingRule) {
        await updateRouteRule(editingRule.id, values)
        message.success('更新成功')
      } else {
        await createRouteRule(values)
        message.success('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      // Validation error
    }
  }

  const handleDragEnd = async (oldIndex: number, newIndex: number) => {
    if (oldIndex === newIndex) return

    const newRules = [...routeRules]
    const [removed] = newRules.splice(oldIndex, 1)
    newRules.splice(newIndex, 0, removed)

    // Update priorities
    const ids = newRules.map((r) => r.id)
    try {
      await reorderRouteRules(ids)
      message.success('排序已更新')
    } catch (error) {
      message.error('排序失败')
    }
  }

  const columns = [
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      render: (priority: number) => <Badge count={priority} showZero style={{ backgroundColor: '#1890ff' }} />,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '级别',
      dataIndex: 'severities',
      key: 'severities',
      render: (severities: string[]) => (
        <Space>
          {severities?.map((s) => (
            <Tag key={s} color="blue">{s}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '来源',
      dataIndex: 'sources',
      key: 'sources',
      render: (sources: string[]) => (
        <Space>
          {sources?.map((s) => (
            <Tag key={s}>{s}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '目标渠道',
      dataIndex: 'channel_ids',
      key: 'channels',
      render: (channelIds: number[]) => {
        const channelNames = channelIds
          ?.map((id) => channels.find((c) => c.id === id)?.name)
          .filter(Boolean)
        return (
          <Space>
            {channelNames?.map((name) => (
              <Tag key={name} color="green">{name}</Tag>
            ))}
          </Space>
        )
      },
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
      render: (_: any, record: RouteRule) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
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
        title="路由规则管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新建规则
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={routeRules}
          rowKey="id"
          loading={routeRulesLoading}
        />
      </Card>

      <Modal
        title={editingRule ? '编辑规则' : '新建规则'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={700}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="规则名称" />
          </Form.Item>
          <Form.Item name="priority" label="优先级" rules={[{ required: true, message: '请输入优先级' }]}>
            <InputNumber min={1} max={100} />
          </Form.Item>

          <Form.Item name="severities" label="匹配级别">
            <Select mode="multiple" placeholder="选择级别">
              <Option value="P0">P0</Option>
              <Option value="P1">P1</Option>
              <Option value="P2">P2</Option>
              <Option value="P3">P3</Option>
            </Select>
          </Form.Item>

          <Form.Item name="sources" label="匹配来源">
            <Select mode="multiple" placeholder="选择数据源">
              {channels.map((c) => (
                <Option key={c.id} value={c.name}>{c.name}</Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="channel_ids" label="目标渠道" rules={[{ required: true, message: '请选择目标渠道' }]}>
            <Select mode="multiple" placeholder="选择推送渠道">
              {channels.map((c) => (
                <Option key={c.id} value={c.id}>{c.name}</Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
