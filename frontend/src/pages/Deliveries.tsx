import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Dialog } from 'primereact/dialog'
import { Sidebar } from 'primereact/sidebar'
import { InputText } from 'primereact/inputtext'
import { InputTextarea } from 'primereact/inputtextarea'
import { Dropdown } from 'primereact/dropdown'
import { Calendar } from 'primereact/calendar'
import { Tag } from 'primereact/tag'
import { Message } from 'primereact/message'
import { Divider } from 'primereact/divider'
import { useSearchParams } from 'react-router-dom'
import dayjs from 'dayjs'
import { deliveryApi, getApiErrorMessage } from '../api/client'
import { canRecoverDeliveries } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import { useToast } from '../components'
import type { Delivery, DeliveryFilters, DeliveryRecoveryResult } from '../types'

type DeliveryFilterForm = {
  alert_id: string
  trace_id: string
  channel_id: string
  delivery_status: string
  created_range: [Date | null, Date | null] | null
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
  alert_id: filters.alert_id || '',
  trace_id: filters.trace_id || '',
  channel_id: typeof filters.channel_id === 'number' ? String(filters.channel_id) : '',
  delivery_status: filters.delivery_status || '',
  created_range:
    filters.created_from || filters.created_to
      ? [
          filters.created_from ? new Date(filters.created_from) : null,
          filters.created_to ? new Date(filters.created_to) : null,
        ]
      : null,
})

const buildFiltersFromForm = (values: DeliveryFilterForm, base?: DeliveryFilters): DeliveryFilters => ({
  alert_id: values.alert_id?.trim() || undefined,
  trace_id: values.trace_id?.trim() || undefined,
  channel_id: values.channel_id ? Number(values.channel_id) : undefined,
  delivery_status: values.delivery_status || undefined,
  created_from: values.created_range?.[0] ? values.created_range[0].toISOString() : undefined,
  created_to: values.created_range?.[1] ? values.created_range[1].toISOString() : undefined,
  limit: base?.limit ?? defaultLimit,
  offset: 0,
})

const statusSeverityMap: Record<string, 'success' | 'danger' | 'warning' | undefined> = {
  delivered: 'success',
  failed: 'danger',
  pending: 'warning',
}

export const Deliveries: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
  const [searchParams, setSearchParams] = useSearchParams()
  const [formValues, setFormValues] = useState<DeliveryFilterForm>(buildFormValues({}))
  const [recoveryReason, setRecoveryReason] = useState('')
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
      toast.showError(getApiErrorMessage(error, '加载投递历史失败'))
    } finally {
      setLoading(false)
    }
  }, [toast])

  useEffect(() => {
    setFormValues(buildFormValues(filters))
    void fetchDeliveries(filters)
  }, [fetchDeliveries, filters])

  const handleSearch = () => {
    const nextFilters = buildFiltersFromForm(formValues, filters)
    setSearchParams(buildSearchParams(nextFilters))
  }

  const handleReset = () => {
    setFormValues(buildFormValues({}))
    setSearchParams(buildSearchParams({ limit: defaultLimit, offset: 0 }))
  }

  const handlePageChange = (event: { first: number; rows: number }) => {
    setSearchParams(
      buildSearchParams({
        ...filters,
        limit: event.rows,
        offset: event.first,
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
      toast.showError(getApiErrorMessage(error, '加载投递详情失败'))
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
    setRecoveryReason('')
  }

  const openRecoveryModal = (delivery: Delivery, action: RecoveryAction) => {
    setRecoveryTarget(delivery)
    setRecoveryAction(action)
    setRecoveryReason('')
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

    if (!recoveryReason.trim()) {
      toast.showError('请填写恢复原因')
      return
    }

    const deliveryID = recoveryTarget.id

    setRecoveryLoadingById((current) => ({ ...current, [deliveryID]: recoveryAction }))
    try {
      const response =
        recoveryAction === 'retry'
          ? await deliveryApi.retry(deliveryID, { reason: recoveryReason.trim() })
          : await deliveryApi.replay(deliveryID, { reason: recoveryReason.trim() })

      setRecoveryFeedback({
        ...response,
        original_delivery_status: recoveryTarget.delivery_status,
      })
      await refreshCurrentView(deliveryID)

      if (response.status === 'succeeded') {
        toast.showSuccess(`${recoveryAction} 已提交并执行成功`)
      } else {
        toast.showWarn(`${recoveryAction} 已记录，结果为 ${response.status}`)
      }

      closeRecoveryModal()
    } catch (error) {
      toast.showError(getApiErrorMessage(error, `${recoveryAction} 执行失败`))
    } finally {
      setRecoveryLoadingById((current) => {
        const nextState = { ...current }
        delete nextState[deliveryID]
        return nextState
      })
    }
  }

  const deliveryStatusOptions = [
    { label: '成功', value: 'delivered' },
    { label: '失败', value: 'failed' },
    { label: '处理中', value: 'pending' },
  ]

  // Column body templates
  const alertIdBodyTemplate = (row: Delivery) => (
    <code className="text-sm">{row.alert_id}</code>
  )

  const channelBodyTemplate = (row: Delivery) => (
    <div className="flex flex-column gap-0">
      <span>{row.channel_snapshot.name}</span>
      <span className="text-color-secondary text-sm">#{row.channel_id}</span>
    </div>
  )

  const statusBodyTemplate = (row: Delivery) => (
    <Tag
      value={row.delivery_status}
      severity={statusSeverityMap[row.delivery_status] ?? undefined}
    />
  )

  const failureBodyTemplate = (row: Delivery) =>
    row.final_failure_summary ? (
      <span className="text-red-500">{row.final_failure_summary.error_message}</span>
    ) : (
      <span className="text-color-secondary">-</span>
    )

  const timeBodyTemplate = (row: Delivery) =>
    dayjs(row.created_at).format('YYYY-MM-DD HH:mm:ss')

  const actionBodyTemplate = (row: Delivery) => (
    <div className="flex gap-1 flex-wrap">
      <Button
        label="查看证据"
        link
        size="small"
        onClick={() => void handleViewDetail(row)}
      />
      {canRecover && row.delivery_status === 'failed' ? (
        <>
          <Button
            label="重试"
            link
            size="small"
            loading={recoveryLoadingById[row.id] === 'retry'}
            disabled={Boolean(recoveryLoadingById[row.id])}
            onClick={() => openRecoveryModal(row, 'retry')}
          />
          <Button
            label="重放"
            link
            size="small"
            loading={recoveryLoadingById[row.id] === 'replay'}
            disabled={Boolean(recoveryLoadingById[row.id])}
            onClick={() => openRecoveryModal(row, 'replay')}
          />
        </>
      ) : null}
    </div>
  )

  const recoveryDialogFooter = (
    <div>
      <Button label="取消" outlined onClick={closeRecoveryModal} />
      <Button
        label={recoveryAction === 'retry' ? '确认重试' : '确认重放'}
        loading={Boolean(recoveryTarget && recoveryLoadingById[recoveryTarget.id])}
        onClick={() => void handleRecoverySubmit()}
      />
    </div>
  )

  return (
    <div className="flex flex-column gap-3">
      {recoveryFeedback ? (
        <Message
          severity={recoveryFeedback.status === 'succeeded' ? 'success' : 'warn'}
          text={`恢复结果: ${recoveryFeedback.action} / ${recoveryFeedback.status}`}
        />
      ) : null}

      {recoveryFeedback && (
        <Card className="shadow-sm border-0">
          <div className="flex flex-column gap-1">
            <span>recovery_id={recoveryFeedback.recovery_id}</span>
            <span>original_delivery_id={recoveryFeedback.original_delivery_id}</span>
            <span>
              resulting_delivery_id=
              {recoveryFeedback.result_delivery_id ? recoveryFeedback.result_delivery_id : '无'}
            </span>
            <span>error_message={recoveryFeedback.error_message || '无'}</span>
          </div>
          <div className="flex justify-content-end mt-2">
            <Button label="关闭" size="small" text onClick={() => setRecoveryFeedback(null)} />
          </div>
        </Card>
      )}

      <Card className="shadow-sm border-0">
        <div className="flex flex-wrap gap-3 align-items-end">
          <div className="flex flex-column gap-2">
            <label className="text-sm">告警 ID</label>
            <InputText
              placeholder="例如 alert-123"
              style={{ width: '220px' }}
              value={formValues.alert_id}
              onChange={(e) => setFormValues({ ...formValues, alert_id: e.target.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">Trace ID</label>
            <InputText
              placeholder="例如 trace-123"
              style={{ width: '220px' }}
              value={formValues.trace_id}
              onChange={(e) => setFormValues({ ...formValues, trace_id: e.target.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">渠道 ID</label>
            <InputText
              inputMode="numeric"
              placeholder="例如 3"
              style={{ width: '140px' }}
              value={formValues.channel_id}
              onChange={(e) => setFormValues({ ...formValues, channel_id: e.target.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">结果</label>
            <Dropdown
              showClear
              placeholder="全部结果"
              style={{ width: '160px' }}
              value={formValues.delivery_status || null}
              options={deliveryStatusOptions}
              onChange={(e) => setFormValues({ ...formValues, delivery_status: e.value || '' })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">创建时间</label>
            <Calendar
              selectionMode="range"
              showTime
              style={{ width: '300px' }}
              value={formValues.created_range}
              onChange={(e) => {
                const dates = e.value as [Date, Date] | null
                setFormValues({ ...formValues, created_range: dates })
              }}
            />
          </div>
          <div className="flex gap-2">
            <Button label="搜索" icon="pi pi-search" onClick={handleSearch} />
            <Button label="重置" icon="pi pi-refresh" outlined onClick={handleReset} />
          </div>
        </div>
      </Card>

      <Card
        className="shadow-sm border-0"
        title="通知投递历史"
        header={
          filters.alert_id ? (
            <div className="flex justify-content-end p-3">
              <Tag value={`alert_id=${filters.alert_id}`} />
            </div>
          ) : undefined
        }
      >
        <DataTable
          value={deliveries}
          dataKey="id"
          loading={loading}
          lazy
          paginator
          first={(currentPage - 1) * pageSize}
          rows={pageSize}
          totalRecords={total}
          onPage={handlePageChange}
          rowsPerPageOptions={[10, 20, 50]}
        >
          <Column field="alert_id" header="告警 ID" body={alertIdBodyTemplate} />
          <Column field="channel_name" header="渠道" body={channelBodyTemplate} />
          <Column field="delivery_status" header="结果" body={statusBodyTemplate} style={{ width: '100px' }} />
          <Column field="attempt_count" header="尝试次数" style={{ width: '100px' }} />
          <Column header="最后失败摘要" body={failureBodyTemplate} />
          <Column field="created_at" header="创建时间" body={timeBodyTemplate} style={{ width: '180px' }} />
          <Column body={actionBodyTemplate} header="操作" style={{ width: '200px' }} />
        </DataTable>
      </Card>

      <Sidebar
        header="投递证据"
        visible={drawerOpen}
        onHide={() => {
          setDrawerOpen(false)
          setSelectedDelivery(null)
        }}
        position="right"
        style={{ width: '720px' }}
      >
        {detailLoading || !selectedDelivery ? (
          <span>正在加载详情...</span>
        ) : (
          <div className="flex flex-column gap-4">
            <div className="flex flex-wrap gap-2">
              {evidenceTags.map((item) => (
                <Tag key={item} value={item} />
              ))}
            </div>

            <div>
              <h4 className="m-0 mb-2">基础信息</h4>
              <div className="grid">
                <div className="col-6 flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">投递 ID</span>
                  <span>{selectedDelivery.id}</span>
                </div>
                <div className="col-6 flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">投递模式</span>
                  <span>{selectedDelivery.delivery_mode}</span>
                </div>
                <div className="col-6 flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">渠道类型</span>
                  <span>{selectedDelivery.channel_snapshot.type}</span>
                </div>
                <div className="col-6 flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">最后成功时间</span>
                  <span>
                    {selectedDelivery.last_success_at
                      ? dayjs(selectedDelivery.last_success_at).format('YYYY-MM-DD HH:mm:ss')
                      : '-'}
                  </span>
                </div>
              </div>
            </div>

            <Divider />

            <div>
              <h4 className="m-0 mb-2">最终失败摘要</h4>
              {selectedDelivery.final_failure_summary ? (
                <div className="flex flex-column gap-1">
                  <span className="text-red-500">{selectedDelivery.final_failure_summary.error_message}</span>
                  <span className="text-color-secondary text-sm">
                    result={selectedDelivery.final_failure_summary.result} retryable=
                    {String(selectedDelivery.final_failure_summary.retryable)} attempts=
                    {selectedDelivery.final_failure_summary.attempt_count}
                  </span>
                </div>
              ) : (
                <span>无</span>
              )}
            </div>

            <Divider />

            <div>
              <h4 className="m-0 mb-2">冻结快照</h4>
              <div className="flex flex-column gap-2">
                <div className="flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">rendered_payload_snapshot</span>
                  <pre className="surface-100 p-2 border-round text-sm overflow-auto m-0" style={{ whiteSpace: 'pre-wrap' }}>
                    {selectedDelivery.rendered_payload_snapshot.title}
                    {'\n'}
                    {selectedDelivery.rendered_payload_snapshot.content}
                  </pre>
                </div>
                <div className="flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">channel_snapshot</span>
                  <span>{selectedDelivery.channel_snapshot.name}</span>
                </div>
                <div className="flex flex-column gap-1">
                  <span className="text-color-secondary text-sm">route_snapshot</span>
                  <span>{selectedDelivery.route_snapshot ? selectedDelivery.route_snapshot.name : '未命中路由'}</span>
                </div>
              </div>
            </div>

            <Divider />

            <Card className="shadow-sm border-0" title="attempts">
              <DataTable
                value={selectedDelivery.attempts}
                dataKey="id"
              >
                <Column field="attempt_number" header="第几次" style={{ width: '100px' }} />
                <Column field="result" header="结果" style={{ width: '120px' }} />
                <Column field="trigger_kind" header="触发来源" style={{ width: '120px' }} />
                <Column field="error_message" header="错误" body={(row) => row.error_message || '-'} />
              </DataTable>
            </Card>
          </div>
        )}
      </Sidebar>

      <Dialog
        header={recoveryAction === 'retry' ? '重试失败投递' : '重放失败投递'}
        visible={recoveryModalOpen}
        onHide={closeRecoveryModal}
        footer={recoveryDialogFooter}
        style={{ width: '500px' }}
      >
        <div className="flex flex-column gap-3">
          <div className="flex flex-column gap-2">
            <label className="text-sm">恢复原因</label>
            <InputTextarea
              rows={4}
              maxLength={200}
              placeholder="说明为什么需要执行这次恢复，原因会进入后端审计记录"
              value={recoveryReason}
              onChange={(e) => setRecoveryReason(e.target.value)}
            />
          </div>
        </div>
      </Dialog>
    </div>
  )
}
