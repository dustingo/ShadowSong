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
  DatePicker,
  message,
  Tabs,
  Typography,
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, StopOutlined } from '@ant-design/icons'
import { PermissionNotice } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { getApiErrorMessage } from '../api/client'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import type { SilenceRule } from '../types'
import dayjs from 'dayjs'

const { Option } = Select
const { RangePicker } = DatePicker
const { Text } = Typography

export const Silences: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const {
    silenceRules,
    silenceRulesLoading,
    fetchSilenceRules,
    createSilenceRule,
    updateSilenceRule,
    deleteSilenceRule,
  } = useConfigStore()

  const [activeTab, setActiveTab] = useState('active')
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<SilenceRule | null>(null)
  const [form] = Form.useForm()
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  useEffect(() => {
    fetchSilenceRules({ status: activeTab as 'active' | 'expired' })
  }, [activeTab, fetchSilenceRules])

  const handleTabChange = (key: string) => {
    setActiveTab(key)
  }

  const handleCreate = () => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    setEditingRule(null)
    form.resetFields()
    form.setFieldsValue({
      enabled: true,
      severities: [],
    })
    setModalVisible(true)
  }

  const handleEdit = (record: SilenceRule) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    setEditingRule(record)
    form.setFieldsValue({
      ...record,
      starts_at: dayjs(record.starts_at),
      ends_at: dayjs(record.ends_at),
    })
    setModalVisible(true)
  }

  const handleDelete = (record: SilenceRule) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除静默规则 "${record.name}" 吗？`,
      onOk: async () => {
        try {
          await deleteSilenceRule(record.id)
          message.success('删除成功')
        } catch (error) {
          message.error(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleCancel = (record: SilenceRule) => {
    if (!canManageConfig) {
      message.warning('当前角色无权执行该操作')
      return
    }
    Modal.confirm({
      title: '确认取消',
      content: `确定要提前取消静默规则 "${record.name}" 吗？`,
      okText: '确认取消',
      onOk: async () => {
        try {
          await updateSilenceRule(record.id, {
            ends_at: new Date().toISOString(),
          })
          message.success('已取消')
          fetchSilenceRules({ status: activeTab as 'active' | 'expired' })
        } catch (error) {
          message.error(getApiErrorMessage(error, '取消失败'))
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
      const data = {
        ...values,
        starts_at: values.timeRange[0].toISOString(),
        ends_at: values.timeRange[1].toISOString(),
      }
      delete data.timeRange

      if (editingRule) {
        await updateSilenceRule(editingRule.id, data)
        message.success('更新成功')
      } else {
        await createSilenceRule(data)
        message.success('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      // Validation error
    }
  }

  const getTimeRemaining = (endTime: string) => {
    const now = dayjs()
    const end = dayjs(endTime)
    const diff = end.diff(now, 'second')

    if (diff <= 0) return '已过期'

    const hours = Math.floor(diff / 3600)
    const minutes = Math.floor((diff % 3600) / 60)

    if (hours > 24) {
      return `${Math.floor(hours / 24)}天 ${hours % 24}小时`
    }
    return `${hours}小时 ${minutes}分钟`
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
      render: (source: string) => source ? <Tag>{source}</Tag> : '-',
    },
    {
      title: '告警名称匹配',
      dataIndex: 'alert_name_pattern',
      key: 'alert_name_pattern',
      render: (pattern: string) => pattern || '-',
    },
    {
      title: '级别',
      dataIndex: 'severities',
      key: 'severities',
      render: (severities: string[]) => (
        <Space>
          {severities?.map((s) => (
            <Tag key={s}>{s}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'starts_at',
      key: 'starts_at',
      render: (time: string) => dayjs(time).format('MM-DD HH:mm'),
    },
    {
      title: '结束时间',
      dataIndex: 'ends_at',
      key: 'ends_at',
      render: (time: string) => dayjs(time).format('MM-DD HH:mm'),
    },
    {
      title: '剩余时间',
      key: 'remaining',
      render: (_: unknown, record: SilenceRule) => {
        if (activeTab === 'expired') return '-'
        return <Text type="warning">{getTimeRemaining(record.ends_at)}</Text>
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: SilenceRule) => (
        canManageConfig ? (
          <Space>
            <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
              编辑
            </Button>
            {activeTab === 'active' && (
              <Button type="link" size="small" icon={<StopOutlined />} onClick={() => handleCancel(record)}>
                取消
              </Button>
            )}
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
        title="静默规则管理"
        extra={
          canManageConfig ? (
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建静默规则
            </Button>
          ) : null
        }
      >
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="静默规则对非 `admin` 角色保持只读。取消、编辑和删除操作不会显示。"
            type="info"
          />
        )}
        <Tabs activeKey={activeTab} onChange={handleTabChange}>
          <Tabs.TabPane tab="活跃" key="active" />
          <Tabs.TabPane tab="历史" key="expired" />
        </Tabs>

        <Table
          columns={columns}
          dataSource={silenceRules}
          rowKey="id"
          loading={silenceRulesLoading}
        />
      </Card>

      <Modal
        title={editingRule ? '编辑静默规则' : '新建静默规则'}
        open={modalVisible}
        onOk={canManageConfig ? handleSubmit : undefined}
        onCancel={() => {
          setModalVisible(false)
          setEditingRule(null)
          form.resetFields()
        }}
        width={600}
        okButtonProps={{ style: { display: canManageConfig ? 'inline-block' : 'none' } }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="规则名称" disabled={!canManageConfig} />
          </Form.Item>
          <Form.Item name="comment" label="备注">
            <Input.TextArea rows={2} placeholder="添加备注" disabled={!canManageConfig} />
          </Form.Item>
          <Form.Item name="source" label="匹配来源">
            <Input placeholder="留空匹配所有来源" disabled={!canManageConfig} />
          </Form.Item>
          <Form.Item name="alert_name_pattern" label="告警名称正则">
            <Input placeholder="正则表达式，如: ^disk.*" disabled={!canManageConfig} />
          </Form.Item>
          <Form.Item name="severities" label="匹配级别">
            <Select mode="multiple" placeholder="留空匹配所有级别" disabled={!canManageConfig}>
              <Option value="P0">P0</Option>
              <Option value="P1">P1</Option>
              <Option value="P2">P2</Option>
              <Option value="P3">P3</Option>
            </Select>
          </Form.Item>
          <Form.Item name="timeRange" label="时间范围" rules={[{ required: true, message: '请选择时间范围' }]}>
            <RangePicker showTime style={{ width: '100%' }} disabled={!canManageConfig} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
