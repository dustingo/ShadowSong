import React, { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Alert as ResultAlert,
  Button,
  Card,
  DatePicker,
  Descriptions,
  Drawer,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import { useSearchParams } from 'react-router-dom'
import dayjs, { type Dayjs } from 'dayjs'
import { deliveryApi, getApiErrorMessage } from '../api/client'
import { canRecoverDeliveries } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import type { Delivery, DeliveryFilters, DeliveryRecoveryResult } from '../types'

const { RangePicker } = DatePicker
const { Text, Paragraph } = Typography

type DeliveryFilterForm = {
  alert_id?: string
  trace_id?: string
  channel_id?: string
  delivery_status?: string
  created_range?: [Dayjs | null, Dayjs | null]
}

type RecoveryAction = 'retry' | 'replay'

type RecoveryFormValues = {
  reason: string
}

type RecoveryFeedback = DeliveryRecoveryResult & {
  original_delivery_status?: string
}

const defaultLimit = 20

const parsePositiveNumber = (value: string | null, fallback?: number): number | undefined => {
  if (!value) {
    return fallback
  }

  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < 0) {
    return fallback
  }

  return parsed
}

const parseFilters = (searchParams: URLSearchParams): DeliveryFilters => {
  const filters: DeliveryFilters = {}

  const alertId = searchParams.get('alert_id')?.trim()
  if (alertId) {
    filters.alert_id = alertId
  }

  const traceId = searchParams.get('trace_id')?.trim()
  if (traceId) {
    filters.trace_id = traceId
  }

  const deliveryStatus = searchParams.get('delivery_status')?.trim()
  if (deliveryStatus) {
    filters.delivery_status = deliveryStatus
  }

  const createdFrom = searchParams.get('created_from')?.trim()
  if (createdFrom && dayjs(createdFrom).isValid()) {
    filters.created_from = dayjs(createdFrom).toISOString()
  }

  const createdTo = searchParams.get('created_to')?.trim()
  if (createdTo && dayjs(createdTo).isValid()) {
    filters.created_to = dayjs(createdTo).toISOString()
  }

  const channelId = parsePositiveNumber(searchParams.get('channel_id'))
  if (channelId && channelId > 0) {
    filters.channel_id = channelId
  }

  const limit = parsePositiveNumber(searchParams.get('limit'), defaultLimit)
  if (limit && limit > 0) {
    filters.limit = limit
  }

  const offset = parsePositiveNumber(searchParams.get('offset'), 0)
  if (typeof offset === 'number' && offset >= 0) {
    filters.offset = offset
  }

  return filters
}

const buildSearchParams = (filters: DeliveryFilters): URLSearchParams => {
  const searchParams = new URLSearchParams()

  if (filters.alert_id) {
    searchParams.set('alert_id', filters.alert_id)
  }
  if (filters.trace_id) {
    searchParams.set('trace_id', filters.trace_id)
  }
  if (typeof filters.channel_id === 'number' && filters.channel_id > 0) {
    searchParams.set('channel_id', String(filters.channel_id))
  }
  if (filters.delivery_status) {
    searchParams.set('delivery_status', filters.delivery_status)
  }
  if (filters.created_from) {
    searchParams.set('created_from', filters.created_from)
  }
  if (filters.created_to) {
    searchParams.set('created_to', filters.created_to)
  }
  if (typeof filters.limit === 'number' && filters.limit > 0) {
    searchParams.set('limit', String(filters.limit))
  }
  if (typeof filters.offset === 'number' && filters.offset > 0) {
    searchParams.set('offset', String(filters.offset))
  }

  return searchParams
}

const buildFormValues = (filters: DeliveryFilters): DeliveryFilterForm => ({
  alert_id: filters.alert_id,
  trace_id: filters.trace_id,
  channel_id: typeof filters.channel_id === 'number' ? String(filters.channel_id) : undefined,
  delivery_status: filters.delivery_status,
  created_range:
    filters.created_from || filters.created_to
      ? [
          filters.created_from ? dayjs(filters.created_from) : null,
          filters.created_to ? dayjs(filters.created_to) : null,
        ]
      : undefined,
})

const buildFiltersFromForm = (values: DeliveryFilterForm, base?: DeliveryFilters): DeliveryFilters => ({
  alert_id: values.alert_id?.trim() || undefined,
  trace_id: values.trace_id?.trim() || undefined,
  channel_id: values.channel_id ? Number(values.channel_id) : undefined,
  delivery_status: values.delivery_status || undefined,
  created_from: values.created_range?.[0]?.toISOString(),
  created_to: values.created_range?.[1]?.toISOString(),
  limit: base?.limit ?? defaultLimit,
  offset: 0,
})

const statusColorMap: Record<string, string> = {
  delivered: 'green',
  failed: 'red',
  pending: 'gold',
}

export const Deliveries: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const [searchParams, setSearchParams] = useSearchParams()
  const [filterForm] = Form.useForm<DeliveryFilterForm>()
  const [recoveryForm] = Form.useForm<RecoveryFormValues>()
  const [deliveries, setDeliveries] = useState<Delivery[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [selectedDelivery, setSelectedDelivery] = useState<Delivery | null>(null)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [recoveryModalOpen, setRecoveryModalOpen] = useState(false)
  const [recoveryTarget, setRecoveryTarget] = useState<Delivery | null>(null)
  const [recoveryAction, setRecoveryAction] = useState<RecoveryAction>('retry')
  const [recoveryLoadingById, setRecoveryLoadingById] = useState<Record<number, RecoveryAction>>({})
  const [recoveryFeedback, setRecoveryFeedback] = useState<RecoveryFeedback | null>(null)

  const filters = useMemo(() => parseFilters(searchParams), [searchParams])
  const canRecover = canRecoverDeliveries(user)
  const pageSize = filters.limit ?? defaultLimit
  const currentPage = Math.floor((filters.offset ?? 0) / pageSize) + 1

  const fetchDeliveries = useCallback(async (nextFilters: DeliveryFilters) => {
    setLoading(true)
    try {
      const response = await deliveryApi.list(nextFilters)
      setDeliveries(response.list)
      setTotal(response.total)
    } catch (error) {
      message.error(getApiErrorMessage(error, '加载投递历史失败'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    filterForm.setFieldsValue(buildFormValues(filters))
    void fetchDeliveries(filters)
  }, [fetchDeliveries, filterForm, filters])

  const handleSearch = async () => {
    const values = await filterForm.validateFields()
    const nextFilters = buildFiltersFromForm(values, filters)
    setSearchParams(buildSearchParams(nextFilters))
  }

  const handleReset = () => {
    filterForm.resetFields()
    setSearchParams(buildSearchParams({ limit: defaultLimit, offset: 0 }))
  }

  const handleTableChange = (page: number, nextPageSize: number) => {
    setSearchParams(
      buildSearchParams({
        ...filters,
        limit: nextPageSize,
        offset: (page - 1) * nextPageSize,
      })
    )
  }

  const handleViewDetail = async (delivery: Delivery) => {
    setDetailLoading(true)
    setDrawerOpen(true)
    try {
      const detail = await deliveryApi.get(delivery.id)
      setSelectedDelivery(detail)
    } catch (error) {
      setDrawerOpen(false)
      message.error(getApiErrorMessage(error, '加载投递详情失败'))
    } finally {
      setDetailLoading(false)
    }
  }

  const evidenceTags = useMemo(() => {
    if (!selectedDelivery) {
      return []
    }

    return [
      `alert_id=${selectedDelivery.alert_id}`,
      `trace_id=${selectedDelivery.trace_id}`,
      `channel=${selectedDelivery.channel_snapshot.name}`,
      `status=${selectedDelivery.delivery_status}`,
    ]
  }, [selectedDelivery])

  const closeRecoveryModal = () => {
    setRecoveryModalOpen(false)
    setRecoveryTarget(null)
    recoveryForm.resetFields()
  }

  const openRecoveryModal = (delivery: Delivery, action: RecoveryAction) => {
    setRecoveryTarget(delivery)
    setRecoveryAction(action)
    recoveryForm.resetFields()
    setRecoveryModalOpen(true)
  }

  const refreshCurrentView = async (deliveryID: number) => {
    await fetchDeliveries(filters)
    if (drawerOpen && selectedDelivery?.id === deliveryID) {
      const detail = await deliveryApi.get(deliveryID)
      setSelectedDelivery(detail)
    }
  }

  const handleRecoverySubmit = async () => {
    if (!recoveryTarget) {
      return
    }

    let values: RecoveryFormValues
    try {
      values = await recoveryForm.validateFields()
    } catch {
      return
    }

    const deliveryID = recoveryTarget.id

    setRecoveryLoadingById((current) => ({ ...current, [deliveryID]: recoveryAction }))
    try {
      const response =
        recoveryAction === 'retry'
          ? await deliveryApi.retry(deliveryID, { reason: values.reason.trim() })
          : await deliveryApi.replay(deliveryID, { reason: values.reason.trim() })

      setRecoveryFeedback({
        ...response,
        original_delivery_status: recoveryTarget.delivery_status,
      })
      await refreshCurrentView(deliveryID)

      if (response.status === 'succeeded') {
        message.success(`${recoveryAction} 已提交并执行成功`)
      } else {
        message.warning(`${recoveryAction} 已记录，结果为 ${response.status}`)
      }

      closeRecoveryModal()
    } catch (error) {
      message.error(getApiErrorMessage(error, `${recoveryAction} 执行失败`))
    } finally {
      setRecoveryLoadingById((current) => {
        const nextState = { ...current }
        delete nextState[deliveryID]
        return nextState
      })
    }
  }

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      {recoveryFeedback ? (
        <ResultAlert
          type={recoveryFeedback.status === 'succeeded' ? 'success' : 'warning'}
          showIcon
          message={`恢复结果: ${recoveryFeedback.action} / ${recoveryFeedback.status}`}
          description={
            <Space direction="vertical" size={0}>
              <Text>recovery_id={recoveryFeedback.recovery_id}</Text>
              <Text>original_delivery_id={recoveryFeedback.original_delivery_id}</Text>
              <Text>
                resulting_delivery_id=
                {recoveryFeedback.result_delivery_id ? recoveryFeedback.result_delivery_id : '无'}
              </Text>
              <Text>
                error_message={recoveryFeedback.error_message || '无'}
              </Text>
            </Space>
          }
          closable
          onClose={() => setRecoveryFeedback(null)}
        />
      ) : null}

      <Card>
        <Form form={filterForm} layout="vertical">
          <Space wrap size="middle" align="start">
            <Form.Item name="alert_id" label="告警 ID" style={{ marginBottom: 0 }}>
              <Input placeholder="例如 alert-123" style={{ width: 220 }} />
            </Form.Item>
            <Form.Item name="trace_id" label="Trace ID" style={{ marginBottom: 0 }}>
              <Input placeholder="例如 trace-123" style={{ width: 220 }} />
            </Form.Item>
            <Form.Item name="channel_id" label="渠道 ID" style={{ marginBottom: 0 }}>
              <Input inputMode="numeric" placeholder="例如 3" style={{ width: 140 }} />
            </Form.Item>
            <Form.Item name="delivery_status" label="结果" style={{ marginBottom: 0 }}>
              <Select
                allowClear
                placeholder="全部结果"
                style={{ width: 160 }}
                options={[
                  { label: '成功', value: 'delivered' },
                  { label: '失败', value: 'failed' },
                  { label: '处理中', value: 'pending' },
                ]}
              />
            </Form.Item>
            <Form.Item name="created_range" label="创建时间" style={{ marginBottom: 0 }}>
              <RangePicker showTime />
            </Form.Item>
            <Space style={{ paddingTop: 30 }}>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => void handleSearch()}>
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={handleReset}>
                重置
              </Button>
            </Space>
          </Space>
        </Form>
      </Card>

      <Card title="通知投递历史" extra={filters.alert_id ? <Tag color="blue">alert_id={filters.alert_id}</Tag> : null}>
        <Table
          rowKey="id"
          dataSource={deliveries}
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (count) => `共 ${count} 条`,
            onChange: handleTableChange,
          }}
          columns={[
            {
              title: '告警 ID',
              dataIndex: 'alert_id',
              key: 'alert_id',
              render: (value: string) => <Text code>{value}</Text>,
            },
            {
              title: '渠道',
              dataIndex: ['channel_snapshot', 'name'],
              key: 'channel_name',
              render: (_: unknown, record: Delivery) => (
                <Space direction="vertical" size={0}>
                  <Text>{record.channel_snapshot.name}</Text>
                  <Text type="secondary">#{record.channel_id}</Text>
                </Space>
              ),
            },
            {
              title: '结果',
              dataIndex: 'delivery_status',
              key: 'delivery_status',
              render: (status: string) => (
                <Tag color={statusColorMap[status] ?? 'default'}>{status}</Tag>
              ),
            },
            {
              title: '尝试次数',
              dataIndex: 'attempt_count',
              key: 'attempt_count',
            },
            {
              title: '最后失败摘要',
              key: 'final_failure_summary',
              render: (_: unknown, record: Delivery) =>
                record.final_failure_summary ? (
                  <Text type="danger">{record.final_failure_summary.error_message}</Text>
                ) : (
                  <Text type="secondary">-</Text>
                ),
            },
            {
              title: '创建时间',
              dataIndex: 'created_at',
              key: 'created_at',
              render: (value: string) => dayjs(value).format('YYYY-MM-DD HH:mm:ss'),
            },
            {
              title: '操作',
              key: 'action',
              render: (_: unknown, record: Delivery) => (
                <Space size="small" wrap>
                  <Button type="link" size="small" onClick={() => void handleViewDetail(record)}>
                    查看证据
                  </Button>
                  {canRecover && record.delivery_status === 'failed' ? (
                    <>
                      <Button
                        type="link"
                        size="small"
                        loading={recoveryLoadingById[record.id] === 'retry'}
                        disabled={Boolean(recoveryLoadingById[record.id])}
                        onClick={() => openRecoveryModal(record, 'retry')}
                      >
                        重试
                      </Button>
                      <Button
                        type="link"
                        size="small"
                        loading={recoveryLoadingById[record.id] === 'replay'}
                        disabled={Boolean(recoveryLoadingById[record.id])}
                        onClick={() => openRecoveryModal(record, 'replay')}
                      >
                        重放
                      </Button>
                    </>
                  ) : null}
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Drawer
        title="投递证据"
        width={720}
        open={drawerOpen}
        onClose={() => {
          setDrawerOpen(false)
          setSelectedDelivery(null)
        }}
        destroyOnClose
      >
        {detailLoading || !selectedDelivery ? (
          <Text>正在加载详情...</Text>
        ) : (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Space wrap>
              {evidenceTags.map((item) => (
                <Tag key={item}>{item}</Tag>
              ))}
            </Space>

            <Descriptions title="基础信息" column={2} bordered size="small">
              <Descriptions.Item label="投递 ID">{selectedDelivery.id}</Descriptions.Item>
              <Descriptions.Item label="投递模式">{selectedDelivery.delivery_mode}</Descriptions.Item>
              <Descriptions.Item label="渠道类型">
                {selectedDelivery.channel_snapshot.type}
              </Descriptions.Item>
              <Descriptions.Item label="最后成功时间">
                {selectedDelivery.last_success_at
                  ? dayjs(selectedDelivery.last_success_at).format('YYYY-MM-DD HH:mm:ss')
                  : '-'}
              </Descriptions.Item>
            </Descriptions>

            <Descriptions title="最终失败摘要" column={1} bordered size="small">
              <Descriptions.Item label="摘要">
                {selectedDelivery.final_failure_summary ? (
                  <Space direction="vertical" size={0}>
                    <Text type="danger">{selectedDelivery.final_failure_summary.error_message}</Text>
                    <Text type="secondary">
                      result={selectedDelivery.final_failure_summary.result} retryable=
                      {String(selectedDelivery.final_failure_summary.retryable)} attempts=
                      {selectedDelivery.final_failure_summary.attempt_count}
                    </Text>
                  </Space>
                ) : (
                  '无'
                )}
              </Descriptions.Item>
            </Descriptions>

            <Descriptions title="冻结快照" column={1} bordered size="small">
              <Descriptions.Item label="rendered_payload_snapshot">
                <Paragraph style={{ whiteSpace: 'pre-wrap', marginBottom: 0 }}>
                  {selectedDelivery.rendered_payload_snapshot.title}
                  {'\n'}
                  {selectedDelivery.rendered_payload_snapshot.content}
                </Paragraph>
              </Descriptions.Item>
              <Descriptions.Item label="channel_snapshot">
                <Text>{selectedDelivery.channel_snapshot.name}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="route_snapshot">
                {selectedDelivery.route_snapshot ? selectedDelivery.route_snapshot.name : '未命中路由'}
              </Descriptions.Item>
            </Descriptions>

            <Card title="attempts" size="small">
              <Table
                rowKey="id"
                pagination={false}
                dataSource={selectedDelivery.attempts}
                columns={[
                  {
                    title: '第几次',
                    dataIndex: 'attempt_number',
                    key: 'attempt_number',
                    width: 100,
                  },
                  {
                    title: '结果',
                    dataIndex: 'result',
                    key: 'result',
                    width: 120,
                  },
                  {
                    title: '触发来源',
                    dataIndex: 'trigger_kind',
                    key: 'trigger_kind',
                    width: 120,
                  },
                  {
                    title: '错误',
                    dataIndex: 'error_message',
                    key: 'error_message',
                    render: (value: string) => value || '-',
                  },
                ]}
              />
            </Card>
          </Space>
        )}
      </Drawer>

      <Modal
        title={recoveryAction === 'retry' ? '重试失败投递' : '重放失败投递'}
        open={recoveryModalOpen}
        okText={recoveryAction === 'retry' ? '确认重试' : '确认重放'}
        cancelText="取消"
        confirmLoading={Boolean(recoveryTarget && recoveryLoadingById[recoveryTarget.id])}
        onOk={() => void handleRecoverySubmit()}
        onCancel={closeRecoveryModal}
        destroyOnHidden
      >
        <Form form={recoveryForm} layout="vertical">
          <Form.Item
            label="恢复原因"
            name="reason"
            rules={[
              { required: true, message: '请填写恢复原因' },
              { whitespace: true, message: '请填写恢复原因' },
            ]}
          >
            <Input.TextArea
              rows={4}
              maxLength={200}
              placeholder="说明为什么需要执行这次恢复，原因会进入后端审计记录"
            />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  )
}
