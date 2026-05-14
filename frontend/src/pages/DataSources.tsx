import React, { useEffect, useState } from 'react'
import { Button } from 'primereact/button'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Dialog } from 'primereact/dialog'
import { Sidebar } from 'primereact/sidebar'
import { InputText } from 'primereact/inputtext'
import { InputSwitch } from 'primereact/inputswitch'
import { Tag } from 'primereact/tag'
import { Message } from 'primereact/message'
import { Divider } from 'primereact/divider'
import { InputNumber } from 'primereact/inputnumber'
import { confirmDialog } from 'primereact/confirmdialog'
import { CopyOutlined, DeleteOutlined, EditOutlined, EyeOutlined, KeyOutlined, PlusOutlined } from '@ant-design/icons'
import { CodeEditor } from '../components/CodeEditor'
import { PermissionNotice } from '../components'
import { useToast } from '../components'
import { dataSourceApi, getApiErrorMessage } from '../api/client'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import type { DataSource, DataSourcePreviewResponse } from '../types'

const defaultPreviewPayload = JSON.stringify(
  {
    status: 'firing',
    labels: {
      alertname: 'ServerLatencyHigh',
      severity: 'warning',
      instance: 'game-01',
    },
    annotations: {
      summary: 'Latency above threshold',
      runbook: 'https://runbook.internal/game-latency',
    },
    summary: 'raw summary from webhook',
    value: 187,
    timestamp: '2026-04-10T07:00:00Z',
  },
  null,
  2
)

const generateApiKey = (): string => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = 'ds_'
  for (let i = 0; i < 32; i += 1) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

const maskApiKey = (key?: string): string => {
  if (!key) return '-'
  if (key.length <= 8) return '****'
  return `${key.substring(0, 8)}****${key.substring(key.length - 4)}`
}

const formatJson = (value: unknown): string => JSON.stringify(value, null, 2)

interface FormValues {
  name: string
  display_name: string
  enabled: boolean
  group_by_labels: string
  deduplicate_enabled: boolean
  deduplicate_window: number
  group_enabled: boolean
  group_window: number
  input_template: string
  output_template: string
}

const defaultFormValues: FormValues = {
  name: '',
  display_name: '',
  enabled: true,
  group_by_labels: '',
  deduplicate_enabled: true,
  deduplicate_window: 3600,
  group_enabled: false,
  group_window: 300,
  input_template: '',
  output_template: '',
}

export const DataSources: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
  const {
    dataSources,
    dataSourcesLoading,
    fetchDataSources,
    createDataSource,
    updateDataSource,
    deleteDataSource,
    toggleDataSource,
    previewDataSource,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingDataSource, setEditingDataSource] = useState<DataSource | null>(null)
  const [currentApiKey, setCurrentApiKey] = useState('')
  const [previewDrawerVisible, setPreviewDrawerVisible] = useState(false)
  const [previewPayload, setPreviewPayload] = useState(defaultPreviewPayload)
  const [previewResult, setPreviewResult] = useState<DataSourcePreviewResponse | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [formValues, setFormValues] = useState<FormValues>(defaultFormValues)
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  useEffect(() => {
    fetchDataSources()
  }, [fetchDataSources])

  useEffect(() => {
    if (!editingDataSource) {
      return
    }

    setFormValues({
      name: editingDataSource.name || '',
      display_name: editingDataSource.display_name || '',
      enabled: editingDataSource.enabled ?? true,
      group_by_labels: editingDataSource.group_by_labels?.join(', ') || '',
      deduplicate_enabled: editingDataSource.deduplicate_enabled === true,
      deduplicate_window: Number(editingDataSource.deduplicate_window) || 3600,
      group_enabled: editingDataSource.group_enabled === true,
      group_window: Number(editingDataSource.group_window) || 300,
      input_template: editingDataSource.input_template || '',
      output_template: editingDataSource.output_template || '',
    })
  }, [editingDataSource])

  const closeEditor = () => {
    setModalVisible(false)
    setEditingDataSource(null)
    setCurrentApiKey('')
    setPreviewResult(null)
    setPreviewDrawerVisible(false)
    setFormValues(defaultFormValues)
  }

  const handleCreate = () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    setEditingDataSource(null)
    setCurrentApiKey('')
    setPreviewResult(null)
    setPreviewPayload(defaultPreviewPayload)
    setFormValues(defaultFormValues)
    setModalVisible(true)
  }

  const handleEdit = async (record: DataSource) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    const fullData = (await dataSourceApi.get(record.id)) as unknown as DataSource
    setEditingDataSource(fullData)
    setCurrentApiKey(fullData.api_key || '')
    setPreviewResult(null)
    setPreviewPayload(defaultPreviewPayload)
    setModalVisible(true)
  }

  const handleDelete = (record: DataSource) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    confirmDialog({
      message: `确定要删除数据源 "${record.display_name}" 吗？`,
      header: '确认删除',
      icon: 'pi pi-exclamation-triangle',
      accept: async () => {
        try {
          await deleteDataSource(record.id)
          toast.showSuccess('删除成功')
        } catch (error) {
          toast.showError(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleToggle = async (record: DataSource) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    try {
      await toggleDataSource(record.id, !record.enabled)
      toast.showSuccess(record.enabled ? '已禁用' : '已启用')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '操作失败'))
    }
  }

  const validateForm = (): boolean => {
    if (!formValues.name.trim()) {
      toast.showError('请输入名称')
      return false
    }
    if (!formValues.display_name.trim()) {
      toast.showError('请输入显示名称')
      return false
    }
    if (!formValues.input_template.trim()) {
      toast.showError('请输入输入模板')
      return false
    }
    if (!formValues.output_template.trim()) {
      toast.showError('请输入输出模板')
      return false
    }
    return true
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    if (!validateForm()) {
      return
    }
    try {
      const payload = {
        ...formValues,
        api_key: currentApiKey,
        group_by_labels: formValues.group_by_labels
          ? formValues.group_by_labels
              .split(',')
              .map((item) => item.trim())
              .filter(Boolean)
          : [],
      }

      if (editingDataSource) {
        await updateDataSource(editingDataSource.id, payload)
        toast.showSuccess('更新成功')
      } else {
        await createDataSource(payload)
        toast.showSuccess('创建成功')
      }

      closeEditor()
    } catch {
      // validation handled by form
    }
  }

  const handlePreview = async () => {
    if (!formValues.name.trim() && !editingDataSource?.name) {
      toast.showError('请输入名称')
      return
    }
    if (!formValues.input_template.trim()) {
      toast.showError('请输入输入模板')
      return
    }
    if (!formValues.output_template.trim()) {
      toast.showError('请输入输出模板')
      return
    }
    try {
      const samplePayload = JSON.parse(previewPayload)

      setPreviewLoading(true)
      const result = await previewDataSource({
        datasource_id: editingDataSource?.id,
        source_name: formValues.name || editingDataSource?.name || 'preview',
        input_template: formValues.input_template,
        output_template: formValues.output_template,
        sample_payload: samplePayload,
      })

      setPreviewResult(result)
      setPreviewDrawerVisible(true)
      toast.showSuccess('模板预览已更新')
    } catch (error: unknown) {
      const errorMessage =
        error instanceof SyntaxError
          ? `JSON 格式错误: ${error.message}`
          : getApiErrorMessage(error, '模板预览失败')
      toast.showError(errorMessage)
    } finally {
      setPreviewLoading(false)
    }
  }

  const updateFormField = <K extends keyof FormValues>(field: K, value: FormValues[K]) => {
    setFormValues((prev) => ({ ...prev, [field]: value }))
  }

  const nameBodyTemplate = (rowData: DataSource) => <Tag value={rowData.name} />

  const webhookBodyTemplate = (rowData: DataSource) => (
    <div className="flex align-items-center gap-2">
      <code className="cursor-pointer" onClick={() => navigator.clipboard.writeText(`/webhook/${rowData.name}`)}>
        /webhook/{rowData.name}
      </code>
      <Button
        icon={<CopyOutlined />}
        className="p-button-text p-button-sm"
        onClick={() => navigator.clipboard.writeText(`/webhook/${rowData.name}`)}
      />
    </div>
  )

  const apiKeyBodyTemplate = (rowData: DataSource) => (
    <div className="flex align-items-center gap-2">
      <code>{maskApiKey(rowData.api_key)}</code>
      {rowData.api_key && (
        <Button
          icon={<CopyOutlined />}
          className="p-button-text p-button-sm"
          onClick={() => navigator.clipboard.writeText(rowData.api_key)}
        />
      )}
    </div>
  )

  const statusBodyTemplate = (rowData: DataSource) => (
    <Tag value={rowData.enabled ? '已启用' : '已禁用'} severity={rowData.enabled ? 'success' : 'secondary'} />
  )

  const lastTriggerBodyTemplate = (rowData: DataSource) =>
    rowData.last_trigger_at ? new Date(rowData.last_trigger_at).toLocaleString() : '-'

  const actionBodyTemplate = (rowData: DataSource) => {
    if (canManageConfig) {
      return (
        <div className="flex gap-2">
          <Button
            icon={<EditOutlined />}
            label="编辑"
            className="p-button-text p-button-sm"
            onClick={() => handleEdit(rowData)}
          />
          <Button
            label={rowData.enabled ? '禁用' : '启用'}
            className="p-button-text p-button-sm"
            onClick={() => handleToggle(rowData)}
          />
          <Button
            icon={<DeleteOutlined />}
            label="删除"
            className="p-button-text p-button-sm p-button-danger"
            onClick={() => handleDelete(rowData)}
          />
        </div>
      )
    }
    return <Tag value="只读" severity="secondary" />
  }

  const dialogFooter = (
    <div className="flex justify-content-end gap-2">
      <Button label="取消" icon="pi pi-times" className="p-button-text" onClick={closeEditor} />
      <Button label="预览模板" icon={<EyeOutlined />} loading={previewLoading} onClick={handlePreview} />
      {canManageConfig && <Button label="保存" icon="pi pi-check" onClick={handleSubmit} />}
    </div>
  )

  return (
    <div>
      <Card
        title={
          <div className="flex justify-content-between align-items-center">
            <span>数据源管理</span>
            <div className="flex align-items-center gap-3">
              <span className="text-sm text-color-secondary">
                接收任意 Webhook JSON，经 input template 映射后再渲染通知
              </span>
              {canManageConfig && (
                <Button icon={<PlusOutlined />} label="新建数据源" onClick={handleCreate} />
              )}
            </div>
          </div>
        }
      >
        {readOnly && (
          <div className="mb-3">
            <PermissionNotice
              title="当前角色可查看配置，但不能修改"
              description="数据源页面对 `operator` 和 `viewer` 保持只读。创建、编辑、删除和启停操作仅对 `admin` 开放。"
              type="info"
            />
          </div>
        )}
        <DataTable value={dataSources} dataKey="id" loading={dataSourcesLoading}>
          <Column field="name" header="名称" body={nameBodyTemplate} />
          <Column field="display_name" header="显示名称" />
          <Column header="Webhook 地址" body={webhookBodyTemplate} />
          <Column field="api_key" header="API Key" body={apiKeyBodyTemplate} />
          <Column header="状态" body={statusBodyTemplate} />
          <Column header="最近触发" body={lastTriggerBodyTemplate} />
          <Column header="操作" body={actionBodyTemplate} />
        </DataTable>
      </Card>

      <Dialog
        header={editingDataSource ? '编辑数据源' : '新建数据源'}
        visible={modalVisible}
        onHide={closeEditor}
        style={{ width: '960px' }}
        footer={dialogFooter}
      >
        <div className="flex flex-column gap-4">
          <div className="flex flex-column gap-2">
            <label htmlFor="name" className="font-semibold">
              名称 <span className="text-red-500">*</span>
            </label>
            <InputText
              id="name"
              value={formValues.name}
              onChange={(e) => updateFormField('name', e.target.value)}
              placeholder="唯一标识，如 prometheus"
              disabled={!!editingDataSource}
            />
          </div>

          <div className="flex flex-column gap-2">
            <label htmlFor="display_name" className="font-semibold">
              显示名称 <span className="text-red-500">*</span>
            </label>
            <InputText
              id="display_name"
              value={formValues.display_name}
              onChange={(e) => updateFormField('display_name', e.target.value)}
              placeholder="友好显示名称"
              disabled={!canManageConfig}
            />
          </div>

          <div className="flex flex-column gap-2">
            <label className="font-semibold flex align-items-center gap-2">
              <KeyOutlined />
              <span>API Key</span>
            </label>
            <small className="text-color-secondary">用于 Webhook 安全校验。留空会拒绝该数据源的所有请求。</small>
            <div className="flex gap-2">
              <InputText
                value={currentApiKey}
                onChange={(e) => setCurrentApiKey(e.target.value)}
                placeholder="留空则拒绝所有请求，或点击生成按钮"
                className="flex-1"
                disabled={!canManageConfig}
              />
              <Button
                icon={<KeyOutlined />}
                label="生成"
                onClick={() => setCurrentApiKey(generateApiKey())}
                disabled={!canManageConfig}
              />
            </div>
          </div>

          <div className="surface-100 p-3 border-round">
            <div className="flex flex-column gap-2">
              <span className="font-semibold">模板字段契约</span>
              <span className="text-sm">
                原始 webhook payload 可以是任意 JSON。`input_template` 的职责是把它映射成系统主链路使用的告警字段，而不是要求上游天然符合内部 `Alert` 结构。
              </span>
              <span className="text-sm">
                映射后的原始 payload 仍会完整保存在 `event` 中；需要直接读取平台原字段时，优先使用 `event.xxx`。
              </span>
              <span className="text-sm">
                旧模板继续使用顶层字段：`alert_name`、`severity`、`message`、`source`、`status`、`trigger_time`、`labels`。
              </span>
              <span className="text-sm">
                其中 `severity` / `severity_code` 是系统标准化后的等级：`critical -&gt; P0`、`warning/error -&gt; P1`、`info -&gt; P2`、`debug -&gt; P3`。
              </span>
              <span className="text-sm">
                如果你想判断原始 webhook 值，直接用 `severity_raw` 或 `event.severity`。原始 webhook JSON 也会完整暴露在 `event` 上。
              </span>
              <p className="text-sm mb-0">
                标准化等级示例：
                <br />
                <code>{`{"title":"[{{.severity_code}}] {{.alert_name}}","content":"{{default .event.annotations.runbook "无 runbook"}}"}`}</code>
              </p>
              <p className="text-sm mb-0">
                原始 severity 示例：
                <br />
                <code>{`{{ if eq .severity_raw "critical" }}严重告警{{ else if eq .severity_raw "warning" }}一般告警{{ else }}提示信息{{ end }}`}</code>
              </p>
            </div>
          </div>

          <div className="flex flex-column gap-2">
            <label htmlFor="group_by_labels" className="font-semibold">
              分组 Labels
            </label>
            <InputText
              id="group_by_labels"
              value={formValues.group_by_labels}
              onChange={(e) => updateFormField('group_by_labels', e.target.value)}
              placeholder="用逗号分隔，如: instance, env"
              disabled={!canManageConfig}
            />
          </div>

          <div className="surface-50 p-3 border-round">
            <div className="flex flex-column gap-4">
              <span className="font-semibold">去重 / 聚合配置</span>

              <div className="flex align-items-center justify-content-between">
                <label htmlFor="deduplicate_enabled" className="font-semibold">
                  启用去重
                </label>
                <InputSwitch
                  id="deduplicate_enabled"
                  checked={formValues.deduplicate_enabled}
                  onChange={(e) => updateFormField('deduplicate_enabled', e.value)}
                  disabled={!canManageConfig}
                />
              </div>

              <div className="flex align-items-center gap-2">
                <label htmlFor="deduplicate_window" className="font-semibold" style={{ width: '100px' }}>
                  去重窗口
                </label>
                <InputNumber
                  id="deduplicate_window"
                  value={formValues.deduplicate_window}
                  onValueChange={(e) => updateFormField('deduplicate_window', e.value ?? 3600)}
                  placeholder="3600"
                  disabled={!canManageConfig}
                  suffix=" 秒"
                  style={{ width: '200px' }}
                />
              </div>

              <Divider className="my-0" />

              <div className="flex align-items-center justify-content-between">
                <label htmlFor="group_enabled" className="font-semibold">
                  启用分组
                </label>
                <InputSwitch
                  id="group_enabled"
                  checked={formValues.group_enabled}
                  onChange={(e) => updateFormField('group_enabled', e.value)}
                  disabled={!canManageConfig}
                />
              </div>

              <div className="flex align-items-center gap-2">
                <label htmlFor="group_window" className="font-semibold" style={{ width: '100px' }}>
                  分组窗口
                </label>
                <InputNumber
                  id="group_window"
                  value={formValues.group_window}
                  onValueChange={(e) => updateFormField('group_window', e.value ?? 300)}
                  placeholder="300"
                  disabled={!canManageConfig}
                  suffix=" 秒"
                  style={{ width: '200px' }}
                />
              </div>
            </div>
          </div>

          <div className="flex flex-column gap-2">
            <label htmlFor="input_template" className="font-semibold">
              输入模板 (Go Template) <span className="text-red-500">*</span>
            </label>
            <small className="text-color-secondary">
              原始 webhook 可以是任意 JSON。这里负责把上游字段映射成系统告警字段；预览会显示映射后的规范化告警。
            </small>
            <CodeEditor
              height={180}
              value={formValues.input_template}
              onChange={(value) => {
                if (canManageConfig) {
                  updateFormField('input_template', value)
                }
              }}
              language="go"
            />
          </div>

          <div className="flex flex-column gap-2">
            <label htmlFor="output_template" className="font-semibold">
              输出模板 (Go Template) <span className="text-red-500">*</span>
            </label>
            <CodeEditor
              height={180}
              value={formValues.output_template}
              onChange={(value) => {
                if (canManageConfig) {
                  updateFormField('output_template', value)
                }
              }}
              language="go"
            />
          </div>

          <Message
            severity="info"
            text="点击「预览模板」会把当前编辑中的 input/output template 和任意样例 JSON 一起发到后端，返回映射后的规范化告警与最终 title/content。"
          />

          <div className="flex align-items-center justify-content-between">
            <label htmlFor="enabled" className="font-semibold">
              启用
            </label>
            <InputSwitch
              id="enabled"
              checked={formValues.enabled}
              onChange={(e) => updateFormField('enabled', e.value)}
              disabled={!canManageConfig}
            />
          </div>
        </div>
      </Dialog>

      <Sidebar
        header="模板预览"
        visible={previewDrawerVisible}
        onHide={() => setPreviewDrawerVisible(false)}
        position="right"
        style={{ width: '720px' }}
      >
        <div className="flex flex-column gap-4">
          <div className="flex justify-content-end">
            <Button label="重新预览" loading={previewLoading} onClick={handlePreview} />
          </div>

          <div className="flex flex-column gap-2">
            <label className="font-semibold">样例 Webhook JSON</label>
            <textarea
              rows={12}
              className="w-full p-3 border-1 border-round surface-border"
              value={previewPayload}
              onChange={(e) => setPreviewPayload(e.target.value)}
            />
          </div>

          {previewResult && (
            <>
              <Card title="上下文键预览" className="mb-0">
                <div className="flex flex-column gap-2">
                  <span>Top-level: {previewResult.context_preview.top_level_keys.join(', ')}</span>
                  <span>event: {previewResult.context_preview.event_keys.join(', ') || '(empty)'}</span>
                  <span>alert: {previewResult.context_preview.alert_keys.join(', ')}</span>
                  <span>labels: {previewResult.context_preview.label_keys.join(', ') || '(empty)'}</span>
                </div>
              </Card>

              <Card title="渲染结果" className="mb-0">
                <div className="flex flex-column gap-3">
                  <div>
                    <span className="font-semibold">Title</span>
                    <pre className="surface-100 p-3 white-space-pre-wrap border-round">{previewResult.rendered.title}</pre>
                  </div>
                  <div>
                    <span className="font-semibold">Content</span>
                    <pre className="surface-100 p-3 white-space-pre-wrap border-round">{previewResult.rendered.content}</pre>
                  </div>
                </div>
              </Card>

              <Card title="规范化告警" className="mb-0">
                <pre className="surface-100 p-3 overflow-auto border-round">{formatJson(previewResult.normalized_alert)}</pre>
              </Card>
            </>
          )}
        </div>
      </Sidebar>
    </div>
  )
}
