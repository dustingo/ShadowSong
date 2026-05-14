import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Tag } from 'primereact/tag'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { Dropdown } from 'primereact/dropdown'
import { InputSwitch } from 'primereact/inputswitch'
import { InputNumber } from 'primereact/inputnumber'
import { Badge } from 'primereact/badge'
import { useToast } from '../components'
import { PermissionNotice } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { getApiErrorMessage } from '../api/client'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import type { RouteRule } from '../types'

interface RouteRuleFormData {
  name: string
  priority: number
  severities: string[]
  sources: string[]
  channel_ids: number[]
  enabled: boolean
}

const severityOptions = [
  { label: 'P0', value: 'P0' },
  { label: 'P1', value: 'P1' },
  { label: 'P2', value: 'P2' },
  { label: 'P3', value: 'P3' },
]

export const RouteRules: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
  const {
    routeRules,
    routeRulesLoading,
    dataSources,
    channels,
    fetchDataSources,
    fetchRouteRules,
    fetchChannels,
    createRouteRule,
    updateRouteRule,
    deleteRouteRule,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingRule, setEditingRule] = useState<RouteRule | null>(null)
  const [formData, setFormData] = useState<RouteRuleFormData>({
    name: '',
    priority: 1,
    severities: [],
    sources: [],
    channel_ids: [],
    enabled: true,
  })
  const [deleteDialogVisible, setDeleteDialogVisible] = useState(false)
  const [ruleToDelete, setRuleToDelete] = useState<RouteRule | null>(null)
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  const sourceOptions = dataSources.map((source) => ({
    label: source.display_name || source.name,
    value: source.name,
  }))

  const channelOptions = channels.map((c) => ({
    label: c.name,
    value: c.id,
  }))

  useEffect(() => {
    fetchRouteRules()
    fetchDataSources()
    fetchChannels()
  }, [fetchChannels, fetchDataSources, fetchRouteRules])

  const resetForm = () => {
    setFormData({
      name: '',
      priority: routeRules.length + 1,
      severities: [],
      sources: [],
      channel_ids: [],
      enabled: true,
    })
  }

  const handleCreate = () => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    setEditingRule(null)
    resetForm()
    setModalVisible(true)
  }

  const handleEdit = (record: RouteRule) => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    setEditingRule(record)
    setFormData({
      name: record.name,
      priority: record.priority,
      severities: record.severities || [],
      sources: record.sources || [],
      channel_ids: record.channel_ids || [],
      enabled: record.enabled,
    })
    setModalVisible(true)
  }

  const handleDeleteClick = (record: RouteRule) => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    setRuleToDelete(record)
    setDeleteDialogVisible(true)
  }

  const handleDeleteConfirm = async () => {
    if (!ruleToDelete) return
    try {
      await deleteRouteRule(ruleToDelete.id)
      toast.showSuccess('删除成功')
      setDeleteDialogVisible(false)
      setRuleToDelete(null)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '删除失败'))
    }
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    if (!formData.name.trim()) {
      toast.showError('请输入名称')
      return
    }
    if (!formData.channel_ids || formData.channel_ids.length === 0) {
      toast.showError('请选择目标渠道')
      return
    }
    try {
      if (editingRule) {
        await updateRouteRule(editingRule.id, formData)
        toast.showSuccess('更新成功')
      } else {
        await createRouteRule(formData)
        toast.showSuccess('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '操作失败'))
    }
  }

  const handleModalHide = () => {
    setModalVisible(false)
    setEditingRule(null)
    resetForm()
  }

  const priorityBodyTemplate = (rowData: RouteRule) => {
    return <Badge value={rowData.priority} severity="info" />
  }

  const severitiesBodyTemplate = (rowData: RouteRule) => {
    return (
      <div className="flex flex-wrap gap-1">
        {rowData.severities?.map((s) => (
          <Tag key={s} value={s} severity="info" />
        ))}
      </div>
    )
  }

  const sourcesBodyTemplate = (rowData: RouteRule) => {
    return (
      <div className="flex flex-wrap gap-1">
        {rowData.sources?.map((s) => (
          <Tag key={s} value={s} />
        ))}
      </div>
    )
  }

  const channelsBodyTemplate = (rowData: RouteRule) => {
    const channelNames = rowData.channel_ids
      ?.map((id) => channels.find((c) => c.id === id)?.name)
      .filter(Boolean)
    return (
      <div className="flex flex-wrap gap-1">
        {channelNames?.map((name) => (
          <Tag key={name} value={name} severity="success" />
        ))}
      </div>
    )
  }

  const statusBodyTemplate = (rowData: RouteRule) => {
    return (
      <Tag
        value={rowData.enabled ? '已启用' : '已禁用'}
        severity={rowData.enabled ? 'success' : 'secondary'}
      />
    )
  }

  const actionBodyTemplate = (rowData: RouteRule) => {
    if (canManageConfig) {
      return (
        <div className="flex gap-2">
          <Button
            icon="pi pi-pencil"
            label="编辑"
            link
            size="small"
            onClick={() => handleEdit(rowData)}
          />
          <Button
            icon="pi pi-trash"
            label="删除"
            link
            severity="danger"
            size="small"
            onClick={() => handleDeleteClick(rowData)}
          />
        </div>
      )
    }
    return <Tag value="只读" />
  }

  const dialogFooter = canManageConfig ? (
    <div className="flex justify-content-end gap-2">
      <Button label="取消" outlined onClick={handleModalHide} />
      <Button label={editingRule ? '更新' : '创建'} onClick={handleSubmit} />
    </div>
  ) : null

  const deleteDialogFooter = (
    <div className="flex justify-content-end gap-2">
      <Button label="取消" outlined onClick={() => setDeleteDialogVisible(false)} />
      <Button label="删除" severity="danger" onClick={handleDeleteConfirm} />
    </div>
  )

  return (
    <div>
      <Card
        title="路由规则管理"
        pt={{
          title: { className: 'text-xl font-semibold' },
        }}
      >
        <template title="header">
          {canManageConfig && (
            <div className="flex justify-content-end mb-3">
              <Button icon="pi pi-plus" label="新建规则" onClick={handleCreate} />
            </div>
          )}
        </template>
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="路由规则的新增、编辑、删除和排序都只对 `admin` 开放。"
            type="info"
          />
        )}
        <DataTable
          value={routeRules}
          dataKey="id"
          loading={routeRulesLoading}
          rowClassName={() => (canManageConfig ? '' : 'permission-readonly-row')}
          stripedRows
        >
          <Column field="priority" header="优先级" body={priorityBodyTemplate} style={{ width: '80px' }} />
          <Column field="name" header="名称" />
          <Column field="severities" header="级别" body={severitiesBodyTemplate} />
          <Column field="sources" header="来源" body={sourcesBodyTemplate} />
          <Column field="channel_ids" header="目标渠道" body={channelsBodyTemplate} />
          <Column field="enabled" header="状态" body={statusBodyTemplate} style={{ width: '100px' }} />
          <Column header="操作" body={actionBodyTemplate} style={{ width: '150px' }} />
        </DataTable>
      </Card>

      <Dialog
        header={editingRule ? '编辑规则' : '新建规则'}
        visible={modalVisible}
        onHide={handleModalHide}
        style={{ width: '700px' }}
        footer={dialogFooter}
      >
        <div className="flex flex-column gap-3 p-fluid">
          <div className="field">
            <label htmlFor="name" className="font-medium">名称</label>
            <InputText
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="规则名称"
              disabled={!canManageConfig}
            />
          </div>

          <div className="field">
            <label htmlFor="priority" className="font-medium">优先级</label>
            <InputNumber
              id="priority"
              value={formData.priority}
              onValueChange={(e) => setFormData({ ...formData, priority: e.value ?? 1 })}
              min={1}
              max={100}
              disabled={!canManageConfig}
            />
          </div>

          <div className="field">
            <label htmlFor="severities" className="font-medium">匹配级别</label>
            <Dropdown
              id="severities"
              value={formData.severities}
              options={severityOptions}
              onChange={(e) => setFormData({ ...formData, severities: e.value })}
              multiple
              placeholder="选择级别"
              disabled={!canManageConfig}
            />
          </div>

          <div className="field">
            <label htmlFor="sources" className="font-medium">匹配来源</label>
            <Dropdown
              id="sources"
              value={formData.sources}
              options={sourceOptions}
              onChange={(e) => setFormData({ ...formData, sources: e.value })}
              multiple
              placeholder="选择数据源"
              disabled={!canManageConfig}
            />
          </div>

          <div className="field">
            <label htmlFor="channel_ids" className="font-medium">目标渠道</label>
            <Dropdown
              id="channel_ids"
              value={formData.channel_ids}
              options={channelOptions}
              onChange={(e) => setFormData({ ...formData, channel_ids: e.value })}
              multiple
              placeholder="选择推送渠道"
              disabled={!canManageConfig}
            />
          </div>

          <div className="field flex align-items-center gap-2">
            <label htmlFor="enabled" className="font-medium mb-0">启用</label>
            <InputSwitch
              id="enabled"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.value })}
              disabled={!canManageConfig}
            />
          </div>
        </div>
      </Dialog>

      <Dialog
        header="确认删除"
        visible={deleteDialogVisible}
        onHide={() => setDeleteDialogVisible(false)}
        style={{ width: '400px' }}
        footer={deleteDialogFooter}
      >
        <p>确定要删除规则 "{ruleToDelete?.name}" 吗？</p>
      </Dialog>
    </div>
  )
}
