import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Tag } from 'primereact/tag'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { InputTextarea } from 'primereact/inputtextarea'
import { MultiSelect } from 'primereact/multiselect'
import { Calendar } from 'primereact/calendar'
import { TabView, TabPanel } from 'primereact/tabview'
import { confirmDialog } from 'primereact/confirmdialog'
import { PermissionNotice, useToast } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { getApiErrorMessage } from '../api/client'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import type { SilenceRule } from '../types'
import dayjs from 'dayjs'

const severityOptions = [
  { label: 'P0', value: 'P0' },
  { label: 'P1', value: 'P1' },
  { label: 'P2', value: 'P2' },
  { label: 'P3', value: 'P3' },
]

export const Silences: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
  const {
    silenceRules,
    silenceRulesLoading,
    fetchSilenceRules,
    createSilenceRule,
    updateSilenceRule,
    deleteSilenceRule,
  } = useConfigStore()

  const [activeTab, setActiveTab] = useState(0)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<SilenceRule | null>(null)
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  // Form state
  const [formName, setFormName] = useState('')
  const [formComment, setFormComment] = useState('')
  const [formSource, setFormSource] = useState('')
  const [formAlertNamePattern, setFormAlertNamePattern] = useState('')
  const [formSeverities, setFormSeverities] = useState<string[]>([])
  const [formTimeRange, setFormTimeRange] = useState<[Date, Date] | null>(null)
  const [formEnabled, setFormEnabled] = useState(true)

  useEffect(() => {
    fetchSilenceRules({ status: activeTab === 0 ? 'active' : 'expired' })
  }, [activeTab, fetchSilenceRules])

  const resetForm = () => {
    setFormName('')
    setFormComment('')
    setFormSource('')
    setFormAlertNamePattern('')
    setFormSeverities([])
    setFormTimeRange(null)
    setFormEnabled(true)
  }

  const handleCreate = () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    setEditingRule(null)
    resetForm()
    setFormEnabled(true)
    setFormSeverities([])
    setModalVisible(true)
  }

  const handleEdit = (record: SilenceRule) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    setEditingRule(record)
    setFormName(record.name)
    setFormComment(record.comment || '')
    setFormSource(record.source || '')
    setFormAlertNamePattern(record.alert_name_pattern || '')
    setFormSeverities(record.severities || [])
    setFormTimeRange([new Date(record.starts_at), new Date(record.ends_at)])
    setFormEnabled(record.enabled ?? true)
    setModalVisible(true)
  }

  const handleDelete = (record: SilenceRule) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    confirmDialog({
      message: `确定要删除静默规则 "${record.name}" 吗？`,
      header: '确认删除',
      icon: 'pi pi-exclamation-triangle',
      accept: async () => {
        try {
          await deleteSilenceRule(record.id)
          toast.showSuccess('删除成功')
        } catch (error) {
          toast.showError(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleCancel = (record: SilenceRule) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    confirmDialog({
      message: `确定要提前取消静默规则 "${record.name}" 吗？`,
      header: '确认取消',
      icon: 'pi pi-exclamation-triangle',
      acceptLabel: '确认取消',
      accept: async () => {
        try {
          await updateSilenceRule(record.id, {
            ends_at: new Date().toISOString(),
          })
          toast.showSuccess('已取消')
          fetchSilenceRules({ status: activeTab === 0 ? 'active' : 'expired' })
        } catch (error) {
          toast.showError(getApiErrorMessage(error, '取消失败'))
        }
      },
    })
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    if (!formName) {
      toast.showError('请输入名称')
      return
    }
    if (!formTimeRange || !formTimeRange[0] || !formTimeRange[1]) {
      toast.showError('请选择时间范围')
      return
    }

    try {
      const data = {
        name: formName,
        comment: formComment,
        source: formSource,
        alert_name_pattern: formAlertNamePattern,
        severities: formSeverities,
        enabled: formEnabled,
        starts_at: formTimeRange[0].toISOString(),
        ends_at: formTimeRange[1].toISOString(),
      }

      if (editingRule) {
        await updateSilenceRule(editingRule.id, data)
        toast.showSuccess('更新成功')
      } else {
        await createSilenceRule(data)
        toast.showSuccess('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '操作失败'))
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

  const nameBodyTemplate = (rowData: SilenceRule) => rowData.name

  const sourceBodyTemplate = (rowData: SilenceRule) =>
    rowData.source ? <Tag value={rowData.source} /> : '-'

  const alertNamePatternBodyTemplate = (rowData: SilenceRule) =>
    rowData.alert_name_pattern || '-'

  const severitiesBodyTemplate = (rowData: SilenceRule) => (
    <div className="flex flex-wrap gap-1">
      {rowData.severities?.map((s) => (
        <Tag key={s} value={s} />
      ))}
    </div>
  )

  const startTimeBodyTemplate = (rowData: SilenceRule) =>
    dayjs(rowData.starts_at).format('MM-DD HH:mm')

  const endTimeBodyTemplate = (rowData: SilenceRule) =>
    dayjs(rowData.ends_at).format('MM-DD HH:mm')

  const remainingBodyTemplate = (rowData: SilenceRule) => {
    if (activeTab === 1) return '-'
    const remaining = getTimeRemaining(rowData.ends_at)
    return <span className="text-orange-500">{remaining}</span>
  }

  const actionBodyTemplate = (rowData: SilenceRule) => {
    if (!canManageConfig) {
      return (
        <Tag
          value="只读"
          style={{
            background: 'var(--surface-hover)',
            color: 'var(--text-secondary)',
          }}
        />
      )
    }
    return (
      <div className="flex gap-2">
        <Button
          icon="pi pi-pencil"
          label="编辑"
          link
          size="small"
          style={{ color: 'var(--primary-color)' }}
          onClick={() => handleEdit(rowData)}
        />
        {activeTab === 0 && (
          <Button
            icon="pi pi-stop"
            label="取消"
            link
            size="small"
            style={{ color: 'var(--text-secondary)' }}
            onClick={() => handleCancel(rowData)}
          />
        )}
        <Button
          icon="pi pi-trash"
          label="删除"
          outlined
          size="small"
          style={{
            color: 'var(--danger-color)',
            borderColor: 'var(--danger-color)',
          }}
          onClick={() => handleDelete(rowData)}
        />
      </div>
    )
  }

  const cardHeader = (
    <div className="flex align-items-center justify-content-between">
      <span className="text-xl font-bold">静默规则管理</span>
      {canManageConfig && (
        <Button icon="pi pi-plus" label="新建静默规则" onClick={handleCreate} />
      )}
    </div>
  )

  const dialogFooter = canManageConfig ? (
    <div>
      <Button label="取消" icon="pi pi-times" outlined onClick={() => setModalVisible(false)} />
      <Button label="保存" icon="pi pi-check" onClick={handleSubmit} />
    </div>
  ) : null

  return (
    <div>
      <Card header={cardHeader}>
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="静默规则对非 `admin` 角色保持只读。取消、编辑和删除操作不会显示。"
            type="info"
          />
        )}
        <TabView activeIndex={activeTab} onTabChange={(e) => setActiveTab(e.index)}>
          <TabPanel header="活跃" />
          <TabPanel header="历史" />
        </TabView>

        <DataTable
          value={silenceRules}
          dataKey="id"
          loading={silenceRulesLoading}
          emptyMessage="暂无数据"
        >
          <Column field="name" header="名称" body={nameBodyTemplate} />
          <Column field="source" header="来源" body={sourceBodyTemplate} />
          <Column field="alert_name_pattern" header="告警名称匹配" body={alertNamePatternBodyTemplate} />
          <Column field="severities" header="级别" body={severitiesBodyTemplate} />
          <Column field="starts_at" header="开始时间" body={startTimeBodyTemplate} />
          <Column field="ends_at" header="结束时间" body={endTimeBodyTemplate} />
          <Column header="剩余时间" body={remainingBodyTemplate} />
          <Column header="操作" body={actionBodyTemplate} />
        </DataTable>
      </Card>

      <Dialog
        header={editingRule ? '编辑静默规则' : '新建静默规则'}
        visible={modalVisible}
        onHide={() => {
          setModalVisible(false)
          setEditingRule(null)
          resetForm()
        }}
        footer={dialogFooter}
        style={{ width: '600px' }}
      >
        <div className="flex flex-column gap-3">
          <div className="flex flex-column gap-2">
            <label htmlFor="name" className="font-semibold">名称 <span className="text-red-500">*</span></label>
            <InputText
              id="name"
              value={formName}
              onChange={(e) => setFormName(e.target.value)}
              placeholder="规则名称"
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="comment" className="font-semibold">备注</label>
            <InputTextarea
              id="comment"
              value={formComment}
              onChange={(e) => setFormComment(e.target.value)}
              rows={2}
              placeholder="添加备注"
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="source" className="font-semibold">匹配来源</label>
            <InputText
              id="source"
              value={formSource}
              onChange={(e) => setFormSource(e.target.value)}
              placeholder="留空匹配所有来源"
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="alertNamePattern" className="font-semibold">告警名称正则</label>
            <InputText
              id="alertNamePattern"
              value={formAlertNamePattern}
              onChange={(e) => setFormAlertNamePattern(e.target.value)}
              placeholder="正则表达式，如: ^disk.*"
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="severities" className="font-semibold">匹配级别</label>
            <MultiSelect
              id="severities"
              value={formSeverities}
              options={severityOptions}
              onChange={(e) => setFormSeverities(e.value)}
              placeholder="留空匹配所有级别"
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="timeRange" className="font-semibold">时间范围 <span className="text-red-500">*</span></label>
            <Calendar
              id="timeRange"
              value={formTimeRange}
              onChange={(e) => setFormTimeRange(e.value as [Date, Date])}
              selectionMode="range"
              showTime
              showIcon
              style={{ width: '100%' }}
              disabled={!canManageConfig}
              placeholder="选择时间范围"
            />
          </div>
        </div>
      </Dialog>
    </div>
  )
}
