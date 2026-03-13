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
} from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useConfigStore } from '../stores/configStore'
import type { OnDuty as OnDutyType } from '../types'
import dayjs from 'dayjs'

const { Option } = Select
const { RangePicker } = DatePicker

export const OnDutyPage: React.FC = () => {
  const {
    onDutyList,
    currentOnDuty,
    onDutyLoading,
    channels,
    fetchOnDuty,
    fetchChannels,
    createOnDuty,
    updateOnDuty,
    deleteOnDuty,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingOnDuty, setEditingOnDuty] = useState<OnDutyType | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchOnDuty()
    fetchChannels()
  }, [])

  const handleCreate = () => {
    setEditingOnDuty(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: OnDutyType) => {
    setEditingOnDuty(record)
    form.setFieldsValue({
      ...record,
      start_time: dayjs(record.start_time),
      end_time: dayjs(record.end_time),
    })
    setModalVisible(true)
  }

  const handleDelete = (record: OnDutyType) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除这条值班记录吗？`,
      onOk: async () => {
        try {
          await deleteOnDuty(record.id)
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
      const data = {
        ...values,
        start_time: values.timeRange[0].toISOString(),
        end_time: values.timeRange[1].toISOString(),
      }
      delete data.timeRange

      if (editingOnDuty) {
        await updateOnDuty(editingOnDuty.id, data)
        message.success('更新成功')
      } else {
        await createOnDuty(data)
        message.success('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      // Validation error
    }
  }

  const columns = [
    {
      title: '值班人员',
      dataIndex: 'user_name',
      key: 'user_name',
    },
    {
      title: '渠道',
      dataIndex: 'channel_id',
      key: 'channel',
      render: (channelId: number) => {
        const channel = channels.find((c) => c.id === channelId)
        return channel ? <Tag color="blue">{channel.name}</Tag> : '-'
      },
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: OnDutyType) => (
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
      {/* Current on duty */}
      {currentOnDuty.length > 0 && (
        <Card title="当前值班" style={{ marginBottom: 16 }}>
          <Space>
            {currentOnDuty.map((duty) => {
              const channel = channels.find((c) => c.id === duty.channel_id)
              return (
                <Tag key={duty.id} color="green" style={{ padding: '8px 16px', fontSize: 14 }}>
                  {duty.user_name} - {channel?.name || '未知渠道'}
                </Tag>
              )
            })}
          </Space>
        </Card>
      )}

      <Card
        title="值班管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新建值班
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={onDutyList}
          rowKey="id"
          loading={onDutyLoading}
        />
      </Card>

      <Modal
        title={editingOnDuty ? '编辑值班' : '新建值班'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="user_name" label="值班人员" rules={[{ required: true, message: '请输入值班人员' }]}>
            <Input placeholder="姓名" />
          </Form.Item>
          <Form.Item name="channel_id" label="通知渠道" rules={[{ required: true, message: '请选择通知渠道' }]}>
            <Select placeholder="选择渠道">
              {channels.filter((c) => c.enabled).map((c) => (
                <Option key={c.id} value={c.id}>{c.name}</Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="timeRange" label="值班时间" rules={[{ required: true, message: '请选择值班时间' }]}>
            <RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
