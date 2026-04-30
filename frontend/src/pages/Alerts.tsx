import React, { useEffect, useState } from 'react'
import {
  Card,
  Table,
  Input,
  Select,
  DatePicker,
  Button,
  Space,
  Modal,
  InputNumber,
  message,
  Tag,
  Typography,
} from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAlertStore } from '../stores/alertStore'
import { SeverityBadge } from '../components/SeverityBadge'
import { PermissionNotice } from '../components'
import { canProcessAlerts as canCurrentUserProcessAlerts } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import { getApiErrorMessage } from '../api/client'
import dayjs from 'dayjs'
import type { Alert } from '../types'

const { RangePicker } = DatePicker
const { Option } = Select
const { TextArea } = Input
const { Text } = Typography

export const Alerts: React.FC = () => {
  const navigate = useNavigate()
  const user = useUserStore((state) => state.user)
  const {
    alerts,
    total,
    page,
    pageSize,
    loading,
    filters,
    fetchAlerts,
    setFilters,
    ackAlert,
    quickSilence,
  } = useAlertStore()

  const [ackModalVisible, setAckModalVisible] = useState(false)
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null)
  const [ackComment, setAckComment] = useState('')

  const [silenceModalVisible, setSilenceModalVisible] = useState(false)
  const [silenceDuration, setSilenceDuration] = useState(3600)
  const canProcessAlerts = canCurrentUserProcessAlerts(user)

  useEffect(() => {
    fetchAlerts()
  }, [fetchAlerts])

  const handleSearch = () => {
    fetchAlerts(1)
  }

  const handleReset = () => {
    setFilters({})
    fetchAlerts(1)
  }

  const handlePageChange = (page: number, pageSize: number) => {
    fetchAlerts(page, pageSize)
  }

  const handleAck = (alert: Alert) => {
    if (!canProcessAlerts) {
      message.warning('当前角色无权执行该操作')
      return
    }
    setSelectedAlert(alert)
    setAckModalVisible(true)
  }

  const handleAckConfirm = async () => {
    if (!selectedAlert) return
    try {
      await ackAlert(selectedAlert.alert_id, ackComment)
      message.success('告警已确认')
      setAckModalVisible(false)
      setAckComment('')
    } catch (error) {
      message.error(getApiErrorMessage(error, '确认失败'))
    }
  }

  const handleQuickSilence = (alert: Alert) => {
    if (!canProcessAlerts) {
      message.warning('当前角色无权执行该操作')
      return
    }
    setSelectedAlert(alert)
    setSilenceModalVisible(true)
  }

  const handleSilenceConfirm = async () => {
    if (!selectedAlert) return
    try {
      await quickSilence(selectedAlert.alert_id, silenceDuration)
      message.success('告警已静默')
      setSilenceModalVisible(false)
    } catch (error) {
      message.error(getApiErrorMessage(error, '静默失败'))
    }
  }

  const handleOpenDeliveries = (alert: Alert) => {
    navigate(`/deliveries?alert_id=${encodeURIComponent(alert.alert_id)}`)
  }

  const columns = [
    {
      title: '级别',
      dataIndex: 'severity',
      key: 'severity',
      width: 100,
      render: (severity: string) => <SeverityBadge severity={severity} />,
    },
    {
      title: '告警名称',
      dataIndex: 'alert_name',
      key: 'alert_name',
      ellipsis: true,
    },
    {
      title: '来源',
      dataIndex: 'source',
      key: 'source',
      width: 120,
      render: (source: string) => <Tag>{source}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap: Record<string, { color: string; text: string }> = {
          firing: { color: 'red', text: '告警中' },
          acked: { color: 'green', text: '已确认' },
          silenced: { color: 'orange', text: '已静默' },
          resolved: { color: 'default', text: '已解决' },
          deduplicated: { color: 'default', text: '已去重' },
        }
        const config = statusMap[status] || { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '触发时间',
      dataIndex: 'trigger_time',
      key: 'trigger_time',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '触发次数',
      dataIndex: 'trigger_count',
      key: 'trigger_count',
      width: 100,
      render: (count: number) => count > 1 ? <Tag color="orange">x{count}</Tag> : count,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: Alert) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleOpenDeliveries(record)}>
            投递历史
          </Button>
          {record.status === 'firing' && (
            canProcessAlerts ? (
              <>
                <Button type="link" size="small" onClick={() => handleAck(record)}>
                  确认
                </Button>
                <Button type="link" size="small" onClick={() => handleQuickSilence(record)}>
                  静默
                </Button>
              </>
            ) : (
              <Tag>只读</Tag>
            )
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      {!canProcessAlerts && (
        <PermissionNotice
          title="当前角色可查看告警，但不能确认或静默"
          description="`viewer` 角色只能查看告警详情和统计结果，处理动作仅对具备告警处理能力的角色开放。"
          type="info"
        />
      )}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap style={{ width: '100%' }} size="middle">
          <Select
            placeholder="级别"
            allowClear
            style={{ width: 120 }}
            value={filters.severity}
            onChange={(value) => setFilters({ ...filters, severity: value })}
          >
            <Option value="P0">P0</Option>
            <Option value="P1">P1</Option>
            <Option value="P2">P2</Option>
            <Option value="P3">P3</Option>
          </Select>
          <Input
            placeholder="来源"
            style={{ width: 120 }}
            value={filters.source}
            onChange={(e) => setFilters({ ...filters, source: e.target.value })}
          />
          <Select
            placeholder="状态"
            allowClear
            style={{ width: 120 }}
            value={filters.status}
            onChange={(value) => setFilters({ ...filters, status: value })}
          >
            <Option value="firing">告警中</Option>
            <Option value="acked">已确认</Option>
            <Option value="silenced">已静默</Option>
            <Option value="resolved">已解决</Option>
          </Select>
          <RangePicker
            showTime
            onChange={(dates) => {
              if (dates) {
                setFilters({
                  ...filters,
                  startTime: dates[0]?.toISOString(),
                  endTime: dates[1]?.toISOString(),
                })
              } else {
                setFilters({ ...filters, startTime: undefined, endTime: undefined })
              }
            }}
          />
          <Input
            placeholder="Labels (如: env=prod)"
            style={{ width: 200 }}
            value={filters.labelSelector}
            onChange={(e) => setFilters({ ...filters, labelSelector: e.target.value })}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
            搜索
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
        </Space>
      </Card>

      <Card>
        <Table
          columns={columns}
          dataSource={alerts}
          rowKey="alert_id"
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: handlePageChange,
          }}
          expandable={{
            expandedRowRender: (record) => (
              <div style={{ padding: '12px 0' }}>
                <Text strong>消息: </Text>
                <Text>{record.message}</Text>
                <br /><br />
                <Text strong>Labels: </Text>
                {record.labels && Object.entries(record.labels).map(([k, v]) => (
                  <Tag key={k} style={{ marginLeft: 4 }}>{k}: {String(v)}</Tag>
                ))}
              </div>
            ),
          }}
        />
      </Card>

      {/* Ack Modal */}
      <Modal
        title="确认告警"
        open={ackModalVisible}
        onOk={handleAckConfirm}
        onCancel={() => setAckModalVisible(false)}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text>确认告警: <Text strong>{selectedAlert?.alert_name}</Text></Text>
          <TextArea
            rows={3}
            placeholder="添加备注（可选）"
            value={ackComment}
            onChange={(e) => setAckComment(e.target.value)}
          />
        </Space>
      </Modal>

      {/* Quick Silence Modal */}
      <Modal
        title="快速静默"
        open={silenceModalVisible}
        onOk={handleSilenceConfirm}
        onCancel={() => setSilenceModalVisible(false)}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text>静默告警: <Text strong>{selectedAlert?.alert_name}</Text></Text>
          <div>
            <Text>静默时长: </Text>
            <InputNumber
              min={60}
              max={86400}
              value={silenceDuration}
              onChange={(v) => setSilenceDuration(v || 3600)}
              addonAfter="秒"
            />
          </div>
          <Space>
            <Button onClick={() => setSilenceDuration(3600)}>1小时</Button>
            <Button onClick={() => setSilenceDuration(14400)}>4小时</Button>
            <Button onClick={() => setSilenceDuration(86400)}>今天</Button>
          </Space>
        </Space>
      </Modal>
    </div>
  )
}
