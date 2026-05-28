import React, { useCallback, useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { InputTextarea } from 'primereact/inputtextarea'
import { InputNumber } from 'primereact/inputnumber'
import { Dropdown } from 'primereact/dropdown'
import { Calendar } from 'primereact/calendar'
import { Tag } from 'primereact/tag'
import { ProgressSpinner } from 'primereact/progressspinner'
import { TabView, TabPanel } from 'primereact/tabview'
import { useNavigate } from 'react-router-dom'
import { useAlertStore } from '../stores/alertStore'
import { SeverityBadge } from '../components/SeverityBadge'
import { PermissionNotice, useToast } from '../components'
import { canProcessAlerts as canCurrentUserProcessAlerts } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import { alertApi, getApiErrorMessage } from '../api/client'
import dayjs from 'dayjs'
import type { Alert, Delivery, GroupedActiveAlert } from '../types'

export const Alerts: React.FC = () => {
  const navigate = useNavigate()
  const toast = useToast()
  const user = useUserStore((state) => state.user)
  const {
    alerts,
    total,
    page,
    pageSize,
    loading,
    filters,
    groupedActiveAlerts,
    groupedActiveLoading,
    fetchAlerts,
    setFilters,
    fetchGroupedActiveAlerts,
    ackAlert,
    quickSilence,
    batchAck,
    batchSilence,
  } = useAlertStore()

  const [ackModalVisible, setAckModalVisible] = useState(false)
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null)
  const [ackComment, setAckComment] = useState('')

  const [silenceModalVisible, setSilenceModalVisible] = useState(false)
  const [silenceDuration, setSilenceDuration] = useState(3600)

  const [expandedRows, setExpandedRows] = useState<Record<string, boolean>>({})
  const [selectedAlerts, setSelectedAlerts] = useState<Alert[]>([])

  const [activeTabIndex, setActiveTabIndex] = useState(0)

  const statusTabs = [
    { label: '全部', status: undefined },
    { label: '活跃', status: 'firing' },
    { label: '已确认', status: 'acked' },
    { label: '已静默', status: 'silenced' },
    { label: '已恢复', status: 'resolved' },
  ]

  const handleTabChange = (e: { index: number }) => {
    setActiveTabIndex(e.index)
    const tab = statusTabs[e.index]
    setFilters({ ...filters, status: tab.status })
  }

  const handleBatchAck = async () => {
    try {
      const ids = selectedAlerts.map((a) => a.alert_id)
      const result = await batchAck(ids, '批量确认')
      toast.showSuccess(已确认  条告警)
      setSelectedAlerts([])
    } catch (error) {
      toast.showError('批量确认失败')
    }
  }

  const handleBatchSilence = async () => {
    try {
      const ids = selectedAlerts.map((a) => a.alert_id)
      const result = await batchSilence(ids, 3600)
      toast.showSuccess(已静默  条告警)
      setSelectedAlerts([])
    } catch (error) {
      toast.showError('批量静默失败')
    }
  }
  }

  const handleRowToggle = (e: { data: GroupedActiveAlert[] }) => {
    const newExpanded: Record<string, boolean> = {}
    e.data.forEach((row) => {
      newExpanded[row.fingerprint] = true
    })
    e.data.forEach((row) => {
      if (!expandedRows[row.fingerprint]) {
        fetchDeliveries(row.latest_alert.alert_id)
      }
    })
    setExpandedRows(newExpanded)
  }

  const [alertDeliveries, setAlertDeliveries] = useState<Record<string, Delivery[]>>({})
  const [deliveriesLoading, setDeliveriesLoading] = useState<Record<string, boolean>>({})

  const fetchDeliveries = useCallback(async (alertId: string) => {
    if (alertDeliveries[alertId] || deliveriesLoading[alertId]) return
    setDeliveriesLoading(prev => ({ ...prev, [alertId]: true }))
    try {
      const data = await alertApi.deliveries(alertId)
      setAlertDeliveries(prev => ({ ...prev, [alertId]: data }))
    } catch (error) {
      // silently fail
    } finally {
      setDeliveriesLoading(prev => ({ ...prev, [alertId]: false }))
    }
  }, [alertDeliveries, deliveriesLoading])

  const canProcessAlerts = canCurrentUserProcessAlerts(user)

  useEffect(() => {
    fetchGroupedActiveAlerts()
    fetchAlerts()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSearch = () => {
    fetchAlerts(1)
  }

  const handleReset = () => {
    setFilters({})
    fetchAlerts(1)
  }

  const handlePageChange = (e: { first: number; rows: number }) => {
    fetchAlerts(Math.floor(e.first / e.rows) + 1, e.rows)
  }

  const handleAck = (alert: Alert) => {
    if (!canProcessAlerts) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    setSelectedAlert(alert)
    setAckModalVisible(true)
  }

  const handleAckConfirm = async () => {
    if (!selectedAlert) return
    try {
      await ackAlert(selectedAlert.alert_id, ackComment)
      toast.showSuccess('告警已确认')
      setAckModalVisible(false)
      setAckComment('')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '确认失败'))
    }
  }

  const handleQuickSilence = (alert: Alert) => {
    if (!canProcessAlerts) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    setSelectedAlert(alert)
    setSilenceModalVisible(true)
  }

  const handleSilenceConfirm = async () => {
    if (!selectedAlert) return
    try {
      await quickSilence(selectedAlert.alert_id, silenceDuration)
      toast.showSuccess('告警已静默')
      setSilenceModalVisible(false)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '静默失败'))
    }
  }

  const handleOpenDeliveries = (alert: Alert) => {
    navigate(`/deliveries?alert_id=${encodeURIComponent(alert.alert_id)}`)
  }

  // === Grouped active alerts templates ===

  const groupedSeverityBodyTemplate = (row: GroupedActiveAlert) => (
    <SeverityBadge severity={row.latest_alert.severity} />
  )

  const groupedAlertNameBodyTemplate = (row: GroupedActiveAlert) => {
    const alert = row.latest_alert
    const reachedEscalationLimit = alert.status === 'firing'
      && alert.notify_count > 0
      && alert.last_notified_at
      && (() => {
        const minutesSinceLastNotify = dayjs().diff(dayjs(alert.last_notified_at), 'minute')
        return minutesSinceLastNotify > 120
      })()

    return (
      <div className="flex align-items-center gap-2">
        <span>{alert.alert_name}</span>
        {row.count > 1 && (
          <Tag
            value={`共 ${row.count} 次`}
            style={{
              background: 'var(--warning-light-color)',
              color: 'var(--warning-color)',
              border: '1px solid var(--warning-color)',
            }}
          />
        )}
        {reachedEscalationLimit ? (
          <Tag value="已达通知上限" severity="warning" />
        ) : alert.notify_count > 0 ? (
          <Tag value={`已通知 ${alert.notify_count} 次`} severity="info" />
        ) : null}
      </div>
    )
  }

  const groupedSourceBodyTemplate = (row: GroupedActiveAlert) => (
    <Tag
      value={row.latest_alert.source}
      style={{
        background: 'var(--surface-hover)',
        color: 'var(--text-secondary)',
        border: '1px solid var(--surface-border)',
      }}
    />
  )

  const groupedTimeBodyTemplate = (row: GroupedActiveAlert) => (
    dayjs(row.last_triggered_at).format('YYYY-MM-DD HH:mm:ss')
  )

  const groupedActionBodyTemplate = (row: GroupedActiveAlert) => {
    const alert = row.latest_alert
    return (
      <div className="flex gap-1">
        <Button
          label="投递历史"
          link
          size="small"
          style={{ color: 'var(--primary-color)' }}
          onClick={() => handleOpenDeliveries(alert)}
        />
        {alert.status === 'firing' && (
          canProcessAlerts ? (
            <>
              <Button
                label="确认"
                link
                size="small"
                style={{ color: 'var(--primary-color)' }}
                onClick={() => handleAck(alert)}
              />
              <Button
                label="静默"
                link
                size="small"
                style={{ color: 'var(--warning-color)' }}
                onClick={() => handleQuickSilence(alert)}
              />
            </>
          ) : (
            <Tag
              value="只读"
              style={{
                background: 'var(--surface-hover)',
                color: 'var(--text-secondary)',
              }}
            />
          )
        )}
      </div>
    )
  }

  const triggerKindLabels: Record<string, string> = {
    pipeline: '首次通知',
    retry: '重试',
    replay: '重放',
    escalation: '升级通知',
  }

  const deliveryStatusConfig: Record<string, { label: string; bgColor: string; color: string }> = {
    delivered: { label: '成功', bgColor: 'var(--success-light-color)', color: 'var(--success-color)' },
    failed: { label: '失败', bgColor: 'var(--danger-light-color)', color: 'var(--danger-color)' },
    pending: { label: '处理中', bgColor: 'var(--warning-light-color)', color: 'var(--warning-color)' },
  }

  const groupedRowExpansionTemplate = (row: GroupedActiveAlert) => {
    const alert = row.latest_alert
    const deliveries = alertDeliveries[alert.alert_id]
    const loading = deliveriesLoading[alert.alert_id]

    return (
      <div className="p-3" style={{ background: 'var(--surface-hover)', borderRadius: '8px' }}>
        <p className="m-0 mb-2" style={{ color: 'var(--text-primary)' }}>
          <strong>消息:</strong> {alert.message}
        </p>
        <div className="flex gap-1 flex-wrap align-items-center">
          <strong style={{ color: 'var(--text-primary)' }}>Labels:</strong>
          {alert.labels && Object.entries(alert.labels).map(([k, v]) => (
            <Tag
              key={k}
              value={`${k}: ${String(v)}`}
              style={{
                background: 'var(--surface-card)',
                color: 'var(--text-secondary)',
                marginLeft: '4px',
              }}
            />
          ))}
        </div>

        <div className="mt-3">
          <strong style={{ color: 'var(--text-primary)' }}>投递记录</strong>
          {loading ? (
            <div className="flex justify-content-center p-3">
              <ProgressSpinner style={{ width: '24px', height: '24px' }} />
            </div>
          ) : deliveries && deliveries.length > 0 ? (
            <DataTable
              value={deliveries}
              dataKey="id"
              size="small"
              stripedRows
              className="mt-2"
              style={{ fontSize: '0.875rem' }}
            >
              <Column
                header="渠道"
                style={{ width: '140px' }}
                body={(d: Delivery) => (
                  <span>{d.channel_snapshot?.name || `#${d.channel_id}`}</span>
                )}
              />
              <Column
                header="状态"
                style={{ width: '90px' }}
                body={(d: Delivery) => {
                  const cfg = deliveryStatusConfig[d.delivery_status] || {
                    label: d.delivery_status,
                    bgColor: 'var(--surface-hover)',
                    color: 'var(--text-secondary)',
                  }
                  return (
                    <Tag
                      value={cfg.label}
                      style={{
                        background: cfg.bgColor,
                        color: cfg.color,
                        border: `1px solid ${cfg.color}`,
                      }}
                    />
                  )
                }}
              />
              <Column header="尝试次数" style={{ width: '80px' }} field="attempt_count" />
              <Column
                header="通知时间"
                style={{ width: '160px' }}
                body={(d: Delivery) => (
                  <span>{dayjs(d.created_at).format('YYYY-MM-DD HH:mm:ss')}</span>
                )}
              />
              <Column
                header="触发类型"
                style={{ width: '100px' }}
                body={(d: Delivery) => {
                  const kind = d.attempts?.length > 0
                    ? d.attempts[d.attempts.length - 1].trigger_kind
                    : (d.final_failure_summary?.trigger_kind || '')
                  const label = triggerKindLabels[kind] || kind
                  return <span>{label}</span>
                }}
              />
              <Column
                header="错误信息"
                body={(d: Delivery) =>
                  d.delivery_status === 'failed' && d.final_failure_summary ? (
                    <span style={{ color: 'var(--danger-color)' }}>
                      {d.final_failure_summary.error_message}
                    </span>
                  ) : (
                    <span style={{ color: 'var(--text-disabled)' }}>-</span>
                  )
                }
              />
            </DataTable>
          ) : deliveries && deliveries.length === 0 ? (
            <div className="mt-2 text-sm" style={{ color: 'var(--text-disabled)' }}>
              暂无投递记录
            </div>
          ) : null}
        </div>
      </div>
    )
  }

  // === All alerts table templates ===

  const severityBodyTemplate = (row: Alert) => <SeverityBadge severity={row.severity} />

  const sourceBodyTemplate = (row: Alert) => (
    <Tag
      value={row.source}
      style={{
        background: 'var(--surface-hover)',
        color: 'var(--text-secondary)',
        border: '1px solid var(--surface-border)',
      }}
    />
  )

  const statusBodyTemplate = (row: Alert) => {
    const statusConfig: Record<string, { bgColor: string; color: string; text: string }> = {
      firing: { bgColor: 'var(--danger-light-color)', color: 'var(--danger-color)', text: '告警中' },
      acked: { bgColor: 'var(--success-light-color)', color: 'var(--success-color)', text: '已确认' },
      silenced: { bgColor: 'var(--warning-light-color)', color: 'var(--warning-color)', text: '已静默' },
      resolved: { bgColor: 'var(--info-light-color)', color: 'var(--info-color)', text: '已解决' },
      deduplicated: { bgColor: 'var(--surface-hover)', color: 'var(--text-secondary)', text: '已去重' },
    }
    const config = statusConfig[row.status] || { bgColor: 'var(--surface-hover)', color: 'var(--text-secondary)', text: row.status }
    return (
      <Tag
        value={config.text}
        style={{
          background: config.bgColor,
          color: config.color,
          border: `1px solid ${config.color}`,
        }}
      />
    )
  }

  const timeBodyTemplate = (row: Alert) => dayjs(row.trigger_time).format('YYYY-MM-DD HH:mm:ss')

  const countBodyTemplate = (row: Alert) =>
    row.trigger_count > 1 ? (
      <Tag
        value={`x${row.trigger_count}`}
        style={{
          background: 'var(--warning-light-color)',
          color: 'var(--warning-color)',
          border: '1px solid var(--warning-color)',
        }}
      />
    ) : (
      <span style={{ color: 'var(--text-secondary)' }}>{row.trigger_count}</span>
    )

  const actionBodyTemplate = (row: Alert) => (
    <div className="flex gap-1">
      <Button
        label="投递历史"
        link
        size="small"
        style={{ color: 'var(--primary-color)' }}
        onClick={() => handleOpenDeliveries(row)}
      />
      {row.status === 'firing' && (
        canProcessAlerts ? (
          <>
            <Button
              label="确认"
              link
              size="small"
              style={{ color: 'var(--primary-color)' }}
              onClick={() => handleAck(row)}
            />
            <Button
              label="静默"
              link
              size="small"
              style={{ color: 'var(--warning-color)' }}
              onClick={() => handleQuickSilence(row)}
            />
          </>
        ) : (
          <Tag
            value="只读"
            style={{
              background: 'var(--surface-hover)',
              color: 'var(--text-secondary)',
            }}
          />
        )
      )}
    </div>
  )

  const rowExpansionTemplate = (row: Alert) => (
    <div className="p-3" style={{ background: 'var(--surface-hover)', borderRadius: '8px' }}>
      <p className="m-0 mb-2" style={{ color: 'var(--text-primary)' }}>
        <strong>消息:</strong> {row.message}
      </p>
      <div className="flex gap-1 flex-wrap align-items-center">
        <strong style={{ color: 'var(--text-primary)' }}>Labels:</strong>
        {row.labels && Object.entries(row.labels).map(([k, v]) => (
          <Tag
            key={k}
            value={`${k}: ${String(v)}`}
            style={{
              background: 'var(--surface-card)',
              color: 'var(--text-secondary)',
              marginLeft: '4px',
            }}
          />
        ))}
      </div>
    </div>
  )

  const severityOptions = [
    { label: 'P0', value: 'P0' },
    { label: 'P1', value: 'P1' },
    { label: 'P2', value: 'P2' },
    { label: 'P3', value: 'P3' },
  ]

  const statusOptions = [
    { label: '告警中', value: 'firing' },
    { label: '已确认', value: 'acked' },
    { label: '已静默', value: 'silenced' },
    { label: '已解决', value: 'resolved' },
  ]

  return (
    <div className="flex flex-column gap-4">
      {!canProcessAlerts && (
        <PermissionNotice
          title="当前角色可查看告警，但不能确认或静默"
          type="info"
        />
      )}

      {/* Active alerts grouped by fingerprint */}
      {groupedActiveAlerts.length > 0 && (
        <Card className="shadow-sm border-0">
          <div className="mb-2 font-semibold" style={{ color: 'var(--text-primary)' }}>
            活跃告警 ({groupedActiveAlerts.length})
          </div>
          <DataTable
            value={groupedActiveAlerts}
            dataKey="fingerprint"
            loading={groupedActiveLoading}
            expandedRows={expandedRows}
            onRowToggle={handleRowToggle}
            rowExpansionTemplate={groupedRowExpansionTemplate}
          >
          <Column expander style={{ width: '40px' }} />
            <Column field="latest_alert.severity" header="级别" body={groupedSeverityBodyTemplate} style={{ width: '120px' }} />
            <Column header="告警名称" body={groupedAlertNameBodyTemplate} />
            <Column field="latest_alert.source" header="来源" body={groupedSourceBodyTemplate} style={{ width: '100px' }} />
            <Column field="last_triggered_at" header="最后触发时间" body={groupedTimeBodyTemplate} style={{ width: '180px' }} />
            <Column body={groupedActionBodyTemplate} header="操作" style={{ width: '200px' }} />
          </DataTable>
        </Card>
      )}

      {/* Status Tabs */}
      <TabView activeIndex={activeTabIndex} onTabChange={handleTabChange}>
        {statusTabs.map((tab) => (
          <TabPanel key={tab.label} header={tab.label} />
        ))}
      </TabView>

      {/* Filters */}
      <Card className="shadow-sm">
        <div className="flex flex-wrap gap-3 align-items-end">
          <div className="flex flex-column gap-2">
            <label className="text-sm">级别</label>
            <Dropdown
              placeholder="级别"
              showClear
              className="w-10rem"
              value={filters.severity}
              options={severityOptions}
              onChange={(e) => setFilters({ ...filters, severity: e.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">来源</label>
            <InputText
              placeholder="来源"
              className="w-10rem"
              value={filters.source || ''}
              onChange={(e) => setFilters({ ...filters, source: e.target.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">状态</label>
            <Dropdown
              placeholder="状态"
              showClear
              className="w-10rem"
              value={filters.status}
              options={statusOptions}
              onChange={(e) => setFilters({ ...filters, status: e.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">时间范围</label>
            <Calendar
              selectionMode="range"
              showTime
              className="w-18rem"
              value={
                filters.startTime && filters.endTime
                  ? [new Date(filters.startTime), new Date(filters.endTime)]
                  : null
              }
              onChange={(e) => {
                const dates = e.value as [Date, Date] | null
                if (dates && dates[0] && dates[1]) {
                  setFilters({
                    ...filters,
                    startTime: dates[0].toISOString(),
                    endTime: dates[1].toISOString(),
                  })
                } else {
                  setFilters({ ...filters, startTime: undefined, endTime: undefined })
                }
              }}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">Labels</label>
            <InputText
              placeholder="Labels (如: env=prod)"
              className="w-14rem"
              value={filters.labelSelector || ''}
              onChange={(e) => setFilters({ ...filters, labelSelector: e.target.value })}
            />
          </div>
          <div className="flex gap-2">
            <Button label="搜索" icon="pi pi-search" onClick={handleSearch} />
            <Button label="重置" icon="pi pi-refresh" outlined onClick={handleReset} />
          </div>
        </div>
      </Card>

      {selectedAlerts.length > 0 && (
        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', alignItems: 'center' }}>
          <span>已选 {selectedAlerts.length} 条</span>
          <Button label="批量确认" icon="pi pi-check" severity="success" onClick={handleBatchAck} />
          <Button label="批量静默" icon="pi pi-volume-off" severity="warning" onClick={handleBatchSilence} />
          <Button label="取消选择" icon="pi pi-times" severity="secondary" onClick={() => setSelectedAlerts([])} />
        </div>
      )}

      {/* All alerts table */}
      <Card className="shadow-sm border-0">
        <DataTable
          value={alerts}
          selection={selectedAlerts}
          onSelectionChange={(e) => setSelectedAlerts(e.value as Alert[])}
          dataKey="alert_id"
          loading={loading}
          lazy
          paginator
          first={(page - 1) * pageSize}
          rows={pageSize}
          totalRecords={total}
          onPage={handlePageChange}
          rowsPerPageOptions={[10, 20, 50]}
          expandedRows={expandedRows}
          onRowToggle={(e) => setExpandedRows(e.data as Record<string, boolean>)}
          rowExpansionTemplate={rowExpansionTemplate}
        >
          <Column selectionMode="multiple" headerStyle={{ width: '3rem' }} />
          <Column expander style={{ width: '40px' }} />
          <Column field="severity" header="级别" body={severityBodyTemplate} style={{ width: '120px' }} />
          <Column field="alert_name" header="告警名称" />
          <Column field="source" header="来源" body={sourceBodyTemplate} style={{ width: '100px' }} />
          <Column field="status" header="状态" body={statusBodyTemplate} style={{ width: '100px' }} />
          <Column field="trigger_time" header="触发时间" body={timeBodyTemplate} style={{ width: '180px' }} />
          <Column field="trigger_count" header="触发次数" body={countBodyTemplate} style={{ width: '100px' }} />
          <Column body={actionBodyTemplate} header="操作" style={{ width: '200px' }} />
        </DataTable>
      </Card>

      {/* Ack Dialog */}
      <Dialog
        header="确认告警"
        visible={ackModalVisible}
        onHide={() => setAckModalVisible(false)}
        footer={
          <div>
            <Button label="取消" outlined onClick={() => setAckModalVisible(false)} />
            <Button label="确认" onClick={handleAckConfirm} />
          </div>
        }
      >
        <div className="flex flex-column gap-3">
          <p>确认告警: <strong>{selectedAlert?.alert_name}</strong></p>
          <InputTextarea
            rows={3}
            placeholder="添加备注（可选）"
            value={ackComment}
            onChange={(e) => setAckComment(e.target.value)}
          />
        </div>
      </Dialog>

      {/* Quick Silence Dialog */}
      <Dialog
        header="快速静默"
        visible={silenceModalVisible}
        onHide={() => setSilenceModalVisible(false)}
        footer={
          <div>
            <Button label="取消" outlined onClick={() => setSilenceModalVisible(false)} />
            <Button label="确认" onClick={handleSilenceConfirm} />
          </div>
        }
      >
        <div className="flex flex-column gap-3">
          <p>静默告警: <strong>{selectedAlert?.alert_name}</strong></p>
          <div className="flex align-items-center gap-2">
            <label>静默时长:</label>
            <InputNumber
              min={60}
              max={86400}
              value={silenceDuration}
              onChange={(e: { value: number | null }) => setSilenceDuration(e.value || 3600)}
              suffix=" 秒"
            />
          </div>
          <div className="flex gap-2">
            <Button label="1小时" outlined onClick={() => setSilenceDuration(3600)} />
            <Button label="4小时" outlined onClick={() => setSilenceDuration(14400)} />
            <Button label="今天" outlined onClick={() => setSilenceDuration(86400)} />
          </div>
        </div>
      </Dialog>
    </div>
  )
}
