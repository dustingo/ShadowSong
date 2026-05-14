import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { InputTextarea } from 'primereact/inputtext'
import { InputNumber } from 'primereact/inputnumber'
import { Dropdown } from 'primereact/dropdown'
import { Calendar } from 'primereact/calendar'
import { Tag } from 'primereact/tag'
import { useNavigate } from 'react-router-dom'
import { useAlertStore } from '../stores/alertStore'
import { SeverityBadge } from '../components/SeverityBadge'
import { PermissionNotice, useToast } from '../components'
import { canProcessAlerts as canCurrentUserProcessAlerts } from '../authz/capabilities'
import { useUserStore } from '../stores/userStore'
import { getApiErrorMessage } from '../api/client'
import dayjs from 'dayjs'
import type { Alert } from '../types'

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

  const severityBodyTemplate = (row: Alert) => <SeverityBadge severity={row.severity} />

  const sourceBodyTemplate = (row: Alert) => <Tag value={row.source} />

  const statusBodyTemplate = (row: Alert) => {
    const statusMap: Record<string, { severity: 'danger' | 'success' | 'warning' | 'info' | 'secondary', text: string }> = {
      firing: { severity: 'danger', text: '告警中' },
      acked: { severity: 'success', text: '已确认' },
      silenced: { severity: 'warning', text: '已静默' },
      resolved: { severity: 'info', text: '已解决' },
      deduplicated: { severity: 'secondary', text: '已去重' },
    }
    const config = statusMap[row.status] || { severity: 'secondary', text: row.status }
    return <Tag value={config.text} severity={config.severity} />
  }

  const timeBodyTemplate = (row: Alert) => dayjs(row.trigger_time).format('YYYY-MM-DD HH:mm:ss')

  const countBodyTemplate = (row: Alert) =>
    row.trigger_count > 1 ? <Tag value={`x${row.trigger_count}`} severity="warning" /> : row.trigger_count

  const actionBodyTemplate = (row: Alert) => (
    <div className="flex gap-1">
      <Button label="投递历史" link size="small" onClick={() => handleOpenDeliveries(row)} />
      {row.status === 'firing' && (
        canProcessAlerts ? (
          <>
            <Button label="确认" link size="small" onClick={() => handleAck(row)} />
            <Button label="静默" link size="small" onClick={() => handleQuickSilence(row)} />
          </>
        ) : (
          <Tag value="只读" />
        )
      )}
    </div>
  )

  const rowExpansionTemplate = (row: Alert) => (
    <div className="p-3">
      <p className="m-0 mb-2"><strong>消息:</strong> {row.message}</p>
      <div className="flex gap-1 flex-wrap">
        <strong>Labels:</strong>
        {row.labels && Object.entries(row.labels).map(([k, v]) => (
          <Tag key={k} value={`${k}: ${String(v)}`} className="ml-1" />
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

      <Card className="shadow-sm border-0">
        <div className="flex flex-wrap gap-3 align-items-end">
          <div className="flex flex-column gap-2">
            <label className="text-sm">级别</label>
            <Dropdown
              placeholder="级别"
              showClear
              style={{ width: '120px' }}
              value={filters.severity}
              options={severityOptions}
              onChange={(e) => setFilters({ ...filters, severity: e.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">来源</label>
            <InputText
              placeholder="来源"
              style={{ width: '120px' }}
              value={filters.source || ''}
              onChange={(e) => setFilters({ ...filters, source: e.target.value })}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">状态</label>
            <Dropdown
              placeholder="状态"
              showClear
              style={{ width: '120px' }}
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
              style={{ width: '300px' }}
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
              style={{ width: '200px' }}
              value={filters.labelSelector || ''}
              onChange={(e) => setFilters({ ...filters, labelSelector: e.target.value })}
            />
          </div>
          <Button label="搜索" icon="pi pi-search" onClick={handleSearch} />
          <Button label="重置" icon="pi pi-refresh" outlined onClick={handleReset} />
        </div>
      </Card>

      <Card className="shadow-sm border-0">
        <DataTable
          value={alerts}
          dataKey="alert_id"
          loading={loading}
          lazy
          paginator
          first={(page - 1) * pageSize}
          rows={pageSize}
          totalRecords={total}
          onPage={handlePageChange}
          rowsPerPageOptions={[10, 20, 50]}
          expandedRows={null}
          rowExpansionTemplate={rowExpansionTemplate}
        >
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
              onValueChange={(e) => setSilenceDuration(e.value || 3600)}
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