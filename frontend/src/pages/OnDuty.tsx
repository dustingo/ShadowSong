import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Tag } from 'primereact/tag'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { Dropdown } from 'primereact/dropdown'
import { Calendar } from 'primereact/calendar'
import { confirmDialog } from 'primereact/confirmdialog'
import { PermissionNotice } from '../components'
import { useToast } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { getApiErrorMessage } from '../api/client'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import type { OnDuty as OnDutyType } from '../types'
import dayjs from 'dayjs'

export const OnDutyPage: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
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
  const [formData, setFormData] = useState({
    user_name: '',
    channel_id: null as number | null,
    timeRange: null as [Date, Date] | null,
  })
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  useEffect(() => {
    fetchOnDuty()
    fetchChannels()
  }, [fetchChannels, fetchOnDuty])

  const resetForm = () => {
    setFormData({
      user_name: '',
      channel_id: null,
      timeRange: null,
    })
    setFormErrors({})
  }

  const handleCreate = () => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    setEditingOnDuty(null)
    resetForm()
    setModalVisible(true)
  }

  const handleEdit = (record: OnDutyType) => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    setEditingOnDuty(record)
    setFormData({
      user_name: record.user_name,
      channel_id: record.channel_id,
      timeRange: [new Date(record.start_time), new Date(record.end_time)],
    })
    setFormErrors({})
    setModalVisible(true)
  }

  const handleDelete = (record: OnDutyType) => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    confirmDialog({
      message: `确定要删除这条值班记录吗？`,
      header: '确认删除',
      icon: 'pi pi-exclamation-triangle',
      accept: async () => {
        try {
          await deleteOnDuty(record.id)
          toast.showSuccess('删除成功')
        } catch (error) {
          toast.showError(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {}
    if (!formData.user_name.trim()) {
      errors.user_name = '请输入值班人员'
    }
    if (!formData.channel_id) {
      errors.channel_id = '请选择通知渠道'
    }
    if (!formData.timeRange || !formData.timeRange[0] || !formData.timeRange[1]) {
      errors.timeRange = '请选择值班时间'
    }
    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      toast.showWarning('当前角色无权执行该操作')
      return
    }
    if (!validateForm()) {
      return
    }
    try {
      const data = {
        user_name: formData.user_name,
        channel_id: formData.channel_id!,
        start_time: formData.timeRange![0].toISOString(),
        end_time: formData.timeRange![1].toISOString(),
      }

      if (editingOnDuty) {
        await updateOnDuty(editingOnDuty.id, data)
        toast.showSuccess('更新成功')
      } else {
        await createOnDuty(data)
        toast.showSuccess('创建成功')
      }
      setModalVisible(false)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '操作失败'))
    }
  }

  const hideModal = () => {
    setModalVisible(false)
    setEditingOnDuty(null)
    resetForm()
  }

  const channelOptions = channels
    .filter((c) => c.enabled)
    .map((c) => ({ label: c.name, value: c.id }))

  const actionBodyTemplate = (record: OnDutyType) => {
    if (canManageConfig) {
      return (
        <div className="flex gap-2">
          <Button
            label="编辑"
            icon="pi pi-pencil"
            link
            size="small"
            onClick={() => handleEdit(record)}
          />
          <Button
            label="删除"
            icon="pi pi-trash"
            link
            severity="danger"
            size="small"
            onClick={() => handleDelete(record)}
          />
        </div>
      )
    }
    return <Tag value="只读" />
  }

  const channelBodyTemplate = (rowData: OnDutyType) => {
    const channel = channels.find((c) => c.id === rowData.channel_id)
    return channel ? <Tag value={channel.name} severity="info" /> : '-'
  }

  const timeBodyTemplate = (time: string) => {
    return dayjs(time).format('YYYY-MM-DD HH:mm')
  }

  const dialogFooter = canManageConfig ? (
    <div>
      <Button label="取消" icon="pi pi-times" outlined onClick={hideModal} />
      <Button label="确定" icon="pi pi-check" onClick={handleSubmit} />
    </div>
  ) : null

  return (
    <div>
      {/* Current on duty */}
      {currentOnDuty.length > 0 && (
        <Card title="当前值班" className="mb-4">
          <div className="flex flex-wrap gap-2">
            {currentOnDuty.map((duty) => {
              const channel = channels.find((c) => c.id === duty.channel_id)
              return (
                <Tag
                  key={duty.id}
                  value={`${duty.user_name} - ${channel?.name || '未知渠道'}`}
                  severity="success"
                  className="px-4 py-2 text-base"
                />
              )
            })}
          </div>
        </Card>
      )}

      <Card
        title="值班管理"
        pt={{
          header: {
            className: 'flex align-items-center justify-content-between',
          },
        }}
      >
        <div className="flex align-items-center justify-content-between mb-3">
          <span className="text-xl font-semibold">值班管理</span>
          {canManageConfig && (
            <Button label="新建值班" icon="pi pi-plus" onClick={handleCreate} />
          )}
        </div>
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="值班配置对非 `admin` 角色保持只读，编辑和删除入口不会开放。"
            type="info"
          />
        )}
        <DataTable
          value={onDutyList}
          dataKey="id"
          loading={onDutyLoading}
          tableStyle={{ minWidth: '50rem' }}
        >
          <Column field="user_name" header="值班人员" />
          <Column field="channel_id" header="渠道" body={channelBodyTemplate} />
          <Column field="start_time" header="开始时间" body={(rowData) => timeBodyTemplate(rowData.start_time)} />
          <Column field="end_time" header="结束时间" body={(rowData) => timeBodyTemplate(rowData.end_time)} />
          <Column header="操作" body={actionBodyTemplate} />
        </DataTable>
      </Card>

      <Dialog
        header={editingOnDuty ? '编辑值班' : '新建值班'}
        visible={modalVisible}
        onHide={hideModal}
        footer={dialogFooter}
        style={{ width: '450px' }}
        modal
      >
        <div className="flex flex-column gap-4 mt-3">
          <div className="flex flex-column gap-2">
            <label htmlFor="user_name" className="font-semibold">
              值班人员 <span className="text-red-500">*</span>
            </label>
            <InputText
              id="user_name"
              value={formData.user_name}
              onChange={(e) => setFormData({ ...formData, user_name: e.target.value })}
              placeholder="姓名"
              disabled={!canManageConfig}
              className={formErrors.user_name ? 'p-invalid' : ''}
            />
            {formErrors.user_name && <small className="p-error">{formErrors.user_name}</small>}
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="channel_id" className="font-semibold">
              通知渠道 <span className="text-red-500">*</span>
            </label>
            <Dropdown
              id="channel_id"
              value={formData.channel_id}
              options={channelOptions}
              onChange={(e) => setFormData({ ...formData, channel_id: e.value })}
              placeholder="选择渠道"
              disabled={!canManageConfig}
              className={formErrors.channel_id ? 'p-invalid' : ''}
            />
            {formErrors.channel_id && <small className="p-error">{formErrors.channel_id}</small>}
          </div>
          <div className="flex flex-column gap-2">
            <label htmlFor="timeRange" className="font-semibold">
              值班时间 <span className="text-red-500">*</span>
            </label>
            <Calendar
              id="timeRange"
              value={formData.timeRange}
              onChange={(e) => setFormData({ ...formData, timeRange: e.value as [Date, Date] })}
              selectionMode="range"
              showTime
              showIcon
              dateFormat="yy-mm-dd"
              placeholder="选择时间范围"
              disabled={!canManageConfig}
              className={formErrors.timeRange ? 'p-invalid' : ''}
              style={{ width: '100%' }}
            />
            {formErrors.timeRange && <small className="p-error">{formErrors.timeRange}</small>}
          </div>
        </div>
      </Dialog>
    </div>
  )
}
