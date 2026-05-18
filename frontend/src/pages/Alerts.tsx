import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Dialog } from 'primereact/dialog'
import { InputTextarea } from 'primereact/inputtextarea'
import { InputNumber } from 'primereact/inputnumber'
import { Tag } from 'primereact/tag'
import { useNavigate } from 'react-router-dom'
import { useAlertStore } from '../stores/alertStore'
import { SeverityBadge } from '../components/SeverityBadge'
import { PermissionNotice, useToast } from '../components'
import { canProcessAlerts as canCurrentUserProcessAlerts } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import { getApiErrorMessage } from '../api/client'
import dayjs from 'dayjs'
import type { Alert, GroupedActiveAlert } from '../types'

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

  const rowExpansionTemplate = (row: GroupedActiveAlert) => {
    const alert = row.latest_alert
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
          onRowToggle={(e) => setExpandedRows(e.data as Record<string, boolean>)}
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
