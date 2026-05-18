import React, { useCallback, useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Dialog } from 'primereact/dialog'
import { InputTextarea } from 'primereact/inputtextarea'
import { InputNumber } from 'primereact/inputnumber'
import { Tag } from 'primereact/tag'
import { ProgressSpinner } from 'primereact/progressspinner'
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
    groupedActiveAlerts,
    groupedActiveLoading,
    fetchGroupedActiveAlerts,
    ackAlert,
    quickSilence,
  } = useAlertStore()

  const [ackModalVisible, setAckModalVisible] = useState(false)
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null)
  const [ackComment, setAckComment] = useState('')

  const [silenceModalVisible, setSilenceModalVisible] = useState(false)
  const [silenceDuration, setSilenceDuration] = useState(3600)

  const [expandedRows, setExpandedRows] = useState<Record<string, boolean>>({})

  const handleRowToggle = (e: { data: GroupedActiveAlert[] }) => {
    const newExpanded: Record<string, boolean> = {}
    e.data.forEach((row) => {
      newExpanded[row.fingerprint] = true
    })
    // Find newly expanded rows and fetch their deliveries
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
    } catch {
      // silently fail — the section will simply not appear
    } finally {
      setDeliveriesLoading(prev => ({ ...prev, [alertId]: false }))
    }
  }, [alertDeliveries, deliveriesLoading])

  const canProcessAlerts = canCurrentUserProcessAlerts(user)

  useEffect(() => {
    fetchGroupedActiveAlerts()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

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

  const severityBodyTemplate = (row: GroupedActiveAlert) => (
    <SeverityBadge severity={row.latest_alert.severity} />
  )

  const alertNameBodyTemplate = (row: GroupedActiveAlert) => (
    <div className="flex align-items-center gap-2">
      <span>{row.latest_alert.alert_name}</span>
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
    </div>
  )

  const sourceBodyTemplate = (row: GroupedActiveAlert) => (
    <Tag
      value={row.latest_alert.source}
      style={{
        background: 'var(--surface-hover)',
        color: 'var(--text-secondary)',
        border: '1px solid var(--surface-border)',
      }}
    />
  )

  const timeBodyTemplate = (row: GroupedActiveAlert) => (
    dayjs(row.last_triggered_at).format('YYYY-MM-DD HH:mm:ss')
  )

  const actionBodyTemplate = (row: GroupedActiveAlert) => {
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

  const rowExpansionTemplate = (row: GroupedActiveAlert) => {
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

        {/* Delivery records section */}
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
              <Column
                header="尝试次数"
                style={{ width: '80px' }}
                field="attempt_count"
              />
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

  return (
    <div className="flex flex-column gap-4">
      {!canProcessAlerts && (
        <PermissionNotice
          title="当前角色可查看告警，但不能确认或静默"
          type="info"
        />
      )}

      <Card className="shadow-sm border-0">
        <DataTable
          value={groupedActiveAlerts}
          dataKey="fingerprint"
          loading={groupedActiveLoading}
          expandedRows={expandedRows}
          onRowToggle={handleRowToggle}
          rowExpansionTemplate={rowExpansionTemplate}
        >
          <Column expander style={{ width: '40px' }} />
          <Column field="latest_alert.severity" header="级别" body={severityBodyTemplate} style={{ width: '120px' }} />
          <Column header="告警名称" body={alertNameBodyTemplate} />
          <Column field="latest_alert.source" header="来源" body={sourceBodyTemplate} style={{ width: '100px' }} />
          <Column field="last_triggered_at" header="最后触发时间" body={timeBodyTemplate} style={{ width: '180px' }} />
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
