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
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { PermissionNotice } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { getApiErrorMessage } from '../api/client'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import type { RouteRule } from '../types'

const { Option } = Select

export const RouteRules: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const {
    routeRules,
    routeRulesLoading,
    dataSources,
    channels,
    fetchDataSources,
    fetchRouteRules,
    fetchChannels,
    createRouteRule,
    updateRouteRule,
    deleteRouteRule,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<RouteRule | null>(null)
  const [form] = Form.useForm()
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  useEffect(() => {
    fetchRouteRules()
    fetchDataSources()
    fetchChannels()
  }, [fetchChannels, fetchDataSources, fetchRouteRules])

  const handleCreate = () => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
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
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    setEditingRule(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = (record: RouteRule) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除规则 "${record.name}" 吗？`,
      onOk: async () => {
        try {
          await deleteRouteRule(record.id)
          message.success('删除成功')
        } catch (error) {
          message.error(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
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
      render: (_: unknown, record: RouteRule) => (
        canManageConfig ? (
          <Space>
            <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
              编辑
            </Button>
            <Button type="link" danger size="small" icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
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
        title="路由规则管理"
        extra={
          canManageConfig ? (
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建规则
            </Button>
          ) : null
        }
      >
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="路由规则的新增、编辑、删除和排序都只对 `admin` 开放。"
            type="info"
          />
        )}
        <Table
          columns={columns}
          dataSource={routeRules}
          rowKey="id"
          loading={routeRulesLoading}
          rowClassName={() => (canManageConfig ? '' : 'permission-readonly-row')}
        />
      </Card>

      <Modal
        title={editingRule ? '编辑规则' : '新建规则'}
        open={modalVisible}
        onOk={canManageConfig ? handleSubmit : undefined}
        onCancel={() => {
          setModalVisible(false)
          setEditingRule(null)
          form.resetFields()
        }}
        width={700}
        okButtonProps={{ style: { display: canManageConfig ? 'inline-block' : 'none' } }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="规则名称" disabled={!canManageConfig} />
          </Form.Item>
          <Form.Item name="priority" label="优先级" rules={[{ required: true, message: '请输入优先级' }]}>
            <InputNumber min={1} max={100} disabled={!canManageConfig} />
          </Form.Item>

          <Form.Item name="severities" label="匹配级别">
            <Select mode="multiple" placeholder="选择级别" disabled={!canManageConfig}>
              <Option value="P0">P0</Option>
              <Option value="P1">P1</Option>
              <Option value="P2">P2</Option>
              <Option value="P3">P3</Option>
            </Select>
          </Form.Item>

          <Form.Item name="sources" label="匹配来源">
            <Select mode="multiple" placeholder="选择数据源" disabled={!canManageConfig}>
              {dataSources.map((source) => (
                <Option key={source.id} value={source.name}>
                  {source.display_name || source.name}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="channel_ids" label="目标渠道" rules={[{ required: true, message: '请选择目标渠道' }]}>
            <Select mode="multiple" placeholder="选择推送渠道" disabled={!canManageConfig}>
              {channels.map((c) => (
                <Option key={c.id} value={c.id}>{c.name}</Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch disabled={!canManageConfig} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
