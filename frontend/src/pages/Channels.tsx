import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Button } from 'primereact/button'
import { Dialog } from 'primereact/dialog'
import { InputText } from 'primereact/inputtext'
import { InputTextarea } from 'primereact/inputtextarea'
import { Dropdown } from 'primereact/dropdown'
import { InputSwitch } from 'primereact/inputswitch'
import { Tag } from 'primereact/tag'
import { Divider } from 'primereact/divider'
import { confirmDialog } from 'primereact/confirmdialog'
import { PermissionNotice, useToast } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { useConfigStore } from '../stores/configStore'
import { channelApi, getApiErrorMessage } from '../api/client'
import { useUserStore } from '../stores/userStore'
import type { Channel } from '../types'

interface ChannelFormValues {
  id?: number
  name: string
  type: string
  enabled: boolean
  config: {
    webhook_url?: string
    secret?: string
    url?: string
    method?: string
    content_type?: string
    headers?: string
    template?: string
    auth_type?: string
    auth_config?: {
      username?: string
      password?: string
      header_name?: string
      header_value?: string
    }
    from_name?: string
  }
}

const initialFormValues: ChannelFormValues = {
  name: '',
  type: 'feishu',
  enabled: true,
  config: {},
}

const formatChannelConfigForForm = (channel: Channel): ChannelFormValues => {
  const config = channel.config ?? {}

  if (channel.type === 'feishu') {
    return {
      ...channel,
      config: {
        webhook_url: String(config.webhook_url ?? ''),
        secret: String(config.secret ?? ''),
      },
    }
  }

  if (channel.type === 'dingtalk') {
    return {
      ...channel,
      config: {
        webhook_url: String(config.webhook_url ?? ''),
        secret: String(config.secret ?? ''),
      },
    }
  }

  if (channel.type === 'wecom') {
    return {
      ...channel,
      config: {
        webhook_url: String(config.webhook_url ?? ''),
      },
    }
  }

  if (channel.type === 'webhook') {
    const authConfig = config.auth_config ?? {}
    return {
      ...channel,
      config: {
        url: String(config.url ?? ''),
        method: String(config.method ?? 'POST'),
        content_type: String(config.content_type ?? 'application/json'),
        headers:
          typeof config.headers === 'string'
            ? config.headers
            : JSON.stringify(config.headers ?? {}, null, 2),
        template: String(config.template ?? ''),
        auth_type: String(config.auth_type ?? 'none'),
        auth_config: {
          username: String(authConfig.username ?? ''),
          password: String(authConfig.password ?? ''),
          header_name: String(authConfig.header_name ?? ''),
          header_value: String(authConfig.header_value ?? ''),
        },
      },
    }
  }

  return { ...channel, config: {} }
}

const buildChannelPayload = (values: ChannelFormValues) => {
  const config = values.config ?? {}

  if (values.type === 'webhook') {
    let headers: Record<string, string> = {}

    if (typeof config.headers === 'string' && config.headers.trim()) {
      headers = JSON.parse(config.headers) as Record<string, string>
    }

    const authConfig = config.auth_config ?? {}
    const authPayload: Record<string, string> = {}
    if (values.config.auth_type === 'basic') {
      authPayload.username = String(authConfig.username ?? '')
      authPayload.password = String(authConfig.password ?? '')
    } else if (values.config.auth_type === 'custom') {
      authPayload.header_name = String(authConfig.header_name ?? '')
      authPayload.header_value = String(authConfig.header_value ?? '')
    }

    return {
      ...values,
      config: {
        url: config.url ?? '',
        method: config.method ?? 'POST',
        content_type: config.content_type ?? 'application/json',
        headers,
        template: config.template ?? '',
        auth_type: config.auth_type ?? 'none',
        auth_config: authPayload,
      },
    }
  }

  if (values.type === 'feishu' || values.type === 'dingtalk') {
    return {
      ...values,
      config: {
        webhook_url: String(config.webhook_url ?? ''),
        secret: String(config.secret ?? ''),
      },
    }
  }

  if (values.type === 'wecom') {
    return {
      ...values,
      config: {
        webhook_url: String(config.webhook_url ?? ''),
      },
    }
  }

  return values
}

export const Channels: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
  const {
    channels,
    channelsLoading,
    fetchChannels,
    createChannel,
    updateChannel,
    deleteChannel,
    toggleChannel,
    testChannel,
  } = useConfigStore()

  const [modalVisible, setModalVisible] = useState(false)
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null)
  const [formValues, setFormValues] = useState<ChannelFormValues>(initialFormValues)
  const [testDialogVisible, setTestDialogVisible] = useState(false)
  const [testDialogChannel, setTestDialogChannel] = useState<Channel | null>(null)
  const [testRecipients, setTestRecipients] = useState('')
  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  useEffect(() => {
    fetchChannels()
  }, [fetchChannels])

  const channelTypeOptions = [
    { value: 'feishu', label: '飞书', icon: 'pi pi-mobile' },
    { value: 'dingtalk', label: '钉钉', icon: 'pi pi-comment' },
    { value: 'wecom', label: '企业微信', icon: 'pi pi-briefcase' },
    { value: 'webhook', label: '自定义 Webhook', icon: 'pi pi-link' },
    { value: 'email', label: '邮件', icon: 'pi pi-envelope' },
  ]

  const methodOptions = [
    { label: 'POST', value: 'POST' },
    { label: 'PUT', value: 'PUT' },
  ]

  const contentTypeOptions = [
    { label: 'JSON (application/json)', value: 'application/json' },
    { label: 'Form (application/x-www-form-urlencoded)', value: 'application/x-www-form-urlencoded' },
  ]

  const authTypeOptions = [
    { label: '无认证', value: 'none' },
    { label: 'Basic Auth', value: 'basic' },
    { label: '自定义 Header', value: 'custom' },
  ]

  const handleCreate = () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    setEditingChannel(null)
    setFormValues({ ...initialFormValues })
    setModalVisible(true)
  }

  const handleEdit = async (record: Channel) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    try {
      const fullChannel = await channelApi.get(record.id) as unknown as Channel
      const formVals = formatChannelConfigForForm(fullChannel)
      setEditingChannel(fullChannel)
      setFormValues(formVals)
      setModalVisible(true)
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '加载渠道配置失败'))
    }
  }

  const handleDelete = (record: Channel) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    confirmDialog({
      message: `确定要删除渠道 "${record.name}" 吗？`,
      header: '确认删除',
      icon: 'pi pi-exclamation-triangle',
      acceptLabel: '确认',
      rejectLabel: '取消',
      accept: async () => {
        try {
          await deleteChannel(record.id)
          toast.showSuccess('删除成功')
        } catch (error) {
          toast.showError(getApiErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleToggle = async (record: Channel) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    try {
      await toggleChannel(record.id, !record.enabled)
      toast.showSuccess(record.enabled ? '已禁用' : '已启用')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '操作失败'))
    }
  }

  const handleTest = async (record: Channel) => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    if (record.type === 'email') {
      setTestDialogChannel(record)
      setTestDialogVisible(true)
      return
    }
    try {
      await testChannel(record.id)
      toast.showSuccess('测试消息已发送')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '发送失败'))
    }
  }

  const handleTestWithEmail = async () => {
    if (!testDialogChannel) return
    try {
      const recipients = testRecipients.split(',').map(s => s.trim()).filter(Boolean)
      await testChannel(testDialogChannel.id, recipients)
      toast.showSuccess('测试邮件已发送')
      setTestDialogVisible(false)
      setTestRecipients('')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '发送失败'))
    }
  }

  const handleSubmit = async () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    if (!formValues.name.trim()) {
      toast.showError('请输入名称')
      return
    }
    try {
      const payload = buildChannelPayload(formValues)
      const channelId = formValues.id || editingChannel?.id
      if (channelId) {
        await updateChannel(channelId, payload as Partial<Channel>)
        toast.showSuccess('更新成功')
      } else {
        await createChannel(payload as Partial<Channel>)
        toast.showSuccess('创建成功')
      }
      setModalVisible(false)
      setEditingChannel(null)
      setFormValues(initialFormValues)
    } catch (error) {
      if (error instanceof SyntaxError) {
        toast.showError('请求头必须是合法 JSON')
      } else {
        toast.showError(getApiErrorMessage(error, '保存失败'))
      }
    }
  }

  const handleModalHide = () => {
    setModalVisible(false)
    setEditingChannel(null)
    setFormValues(initialFormValues)
  }

  const getTypeLabel = (type: string) => {
    const opt = channelTypeOptions.find((o) => o.value === type)
    return opt ? `${opt.label}` : type
  }

  const nameBodyTemplate = (row: Channel) => row.name

  const typeBodyTemplate = (row: Channel) => {
    const opt = channelTypeOptions.find((o) => o.value === row.type)
    return <Tag value={opt?.label || row.type} icon={opt?.icon} />
  }

  const statusBodyTemplate = (row: Channel) => {
    if (row.enabled) {
      return <Tag value="已启用" severity="success" />
    }
    return <Tag value="已禁用" severity="secondary" />
  }

  const actionBodyTemplate = (row: Channel) => {
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
      <div className="flex gap-1">
        <Button
          label="编辑"
          link
          size="small"
          icon="pi pi-pencil"
          style={{ color: 'var(--primary-color)' }}
          onClick={() => handleEdit(row)}
        />
        <Button
          label="测试"
          link
          size="small"
          icon="pi pi-send"
          style={{ color: 'var(--primary-color)' }}
          onClick={() => handleTest(row)}
        />
        <Button
          label={row.enabled ? '禁用' : '启用'}
          link
          size="small"
          style={{ color: 'var(--text-secondary)' }}
          onClick={() => handleToggle(row)}
        />
        <Button
          label="删除"
          outlined
          size="small"
          icon="pi pi-trash"
          style={{
            color: 'var(--danger-color)',
            borderColor: 'var(--danger-color)',
          }}
          onClick={() => handleDelete(row)}
        />
      </div>
    )
  }

  const dialogFooter = (
    <div>
      <Button label="取消" outlined onClick={handleModalHide} />
      {canManageConfig && <Button label="保存" onClick={handleSubmit} />}
    </div>
  )

  const renderConfigFields = () => {
    const { type } = formValues

    if (type === 'feishu') {
      return (
        <>
          <div className="flex flex-column gap-2">
            <label className="text-sm">Webhook URL</label>
            <InputText
              placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
              value={formValues.config.webhook_url || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, webhook_url: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">签名密钥（可选）</label>
            <InputText
              placeholder="签名密钥"
              value={formValues.config.secret || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, secret: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
        </>
      )
    }

    if (type === 'dingtalk') {
      return (
        <>
          <div className="flex flex-column gap-2">
            <label className="text-sm">Webhook URL</label>
            <InputText
              placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
              value={formValues.config.webhook_url || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, webhook_url: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">加签密钥（可选）</label>
            <InputText
              placeholder="密钥"
              value={formValues.config.secret || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, secret: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
        </>
      )
    }

    if (type === 'wecom') {
      return (
        <div className="flex flex-column gap-2">
          <label className="text-sm">Webhook URL</label>
          <InputText
            placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
            value={formValues.config.webhook_url || ''}
            onChange={(e) =>
              setFormValues({
                ...formValues,
                config: { ...formValues.config, webhook_url: e.target.value },
              })
            }
            disabled={!canManageConfig}
          />
        </div>
      )
    }

    if (type === 'webhook') {
      return (
        <>
          <div className="flex flex-column gap-2">
            <label className="text-sm">请求 URL</label>
            <InputText
              placeholder="https://example.com/webhook"
              value={formValues.config.url || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, url: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">请求方法</label>
            <Dropdown
              value={formValues.config.method || 'POST'}
              options={methodOptions}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, method: e.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">请求体格式</label>
            <Dropdown
              value={formValues.config.content_type || 'application/json'}
              options={contentTypeOptions}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, content_type: e.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">请求头（JSON）</label>
            <InputTextarea
              rows={2}
              placeholder='{"Content-Type": "application/json"}'
              value={formValues.config.headers || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, headers: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">请求体模板</label>
            <InputTextarea
              rows={3}
              placeholder={
                formValues.config.content_type === 'application/x-www-form-urlencoded'
                  ? 'teams=ops&title={{.alert_name}}&app_content={{.message}}'
                  : '{"text": "{{.message}}"}'
              }
              value={formValues.config.template || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: { ...formValues.config, template: e.target.value },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <Divider align="center">
            <span className="text-sm">认证配置</span>
          </Divider>
          <div className="flex flex-column gap-2">
            <label className="text-sm">认证方式</label>
            <Dropdown
              value={formValues.config.auth_type || 'none'}
              options={authTypeOptions}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: {
                    ...formValues.config,
                    auth_type: e.value,
                    auth_config: {},
                  },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          {formValues.config.auth_type === 'basic' && (
            <>
              <div className="flex flex-column gap-2">
                <label className="text-sm">用户名</label>
                <InputText
                  placeholder="app_key"
                  value={formValues.config.auth_config?.username || ''}
                  onChange={(e) =>
                    setFormValues({
                      ...formValues,
                      config: {
                        ...formValues.config,
                        auth_config: { ...formValues.config.auth_config, username: e.target.value },
                      },
                    })
                  }
                  disabled={!canManageConfig}
                />
              </div>
              <div className="flex flex-column gap-2">
                <label className="text-sm">密码</label>
                <InputText
                  placeholder="appsecret"
                  value={formValues.config.auth_config?.password || ''}
                  onChange={(e) =>
                    setFormValues({
                      ...formValues,
                      config: {
                        ...formValues.config,
                        auth_config: { ...formValues.config.auth_config, password: e.target.value },
                      },
                    })
                  }
                  disabled={!canManageConfig}
                />
              </div>
            </>
          )}
          {formValues.config.auth_type === 'custom' && (
            <>
              <div className="flex flex-column gap-2">
                <label className="text-sm">Header 名称</label>
                <InputText
                  placeholder="Authorization"
                  value={formValues.config.auth_config?.header_name || ''}
                  onChange={(e) =>
                    setFormValues({
                      ...formValues,
                      config: {
                        ...formValues.config,
                        auth_config: { ...formValues.config.auth_config, header_name: e.target.value },
                      },
                    })
                  }
                  disabled={!canManageConfig}
                />
              </div>
              <div className="flex flex-column gap-2">
                <label className="text-sm">Header 值</label>
                <InputText
                  placeholder="Bearer xxx"
                  value={formValues.config.auth_config?.header_value || ''}
                  onChange={(e) =>
                    setFormValues({
                      ...formValues,
                      config: {
                        ...formValues.config,
                        auth_config: { ...formValues.config.auth_config, header_value: e.target.value },
                      },
                    })
                  }
                  disabled={!canManageConfig}
                />
              </div>
            </>
          )}
        </>
      )
    }

    if (type === 'email') {
      return (
        <div className="flex flex-column gap-2">
          <label className="text-sm">发件人名称（可选）</label>
          <InputText
            placeholder="告警系统"
            value={formValues.config.from_name || ''}
            onChange={(e) =>
              setFormValues({
                ...formValues,
                config: { ...formValues.config, from_name: e.target.value },
              })
            }
            disabled={!canManageConfig}
          />
          <small className="text-color-secondary">留空则使用 SMTP 全局配置中的发件人名称</small>
        </div>
      )
    }

    return null
  }

  const cardHeader = (
    <div className="flex align-items-center justify-content-between">
      <div>
        <span className="text-xl font-bold">推送渠道管理</span>
        <span className="text-color-secondary text-sm ml-2">管理告警通知推送渠道</span>
      </div>
      {canManageConfig && (
        <Button label="新建渠道" icon="pi pi-plus" onClick={handleCreate} />
      )}
    </div>
  )

  return (
    <div>
      <Card
        className="shadow-sm border-0"
        header={cardHeader}
      >
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="渠道配置对非 `admin` 角色保持只读，测试、启停和删除操作不会开放。"
            type="info"
          />
        )}
        <DataTable
          value={channels}
          dataKey="id"
          loading={channelsLoading}
        >
          <Column field="name" header="名称" body={nameBodyTemplate} />
          <Column field="type" header="类型" body={typeBodyTemplate} />
          <Column field="enabled" header="状态" body={statusBodyTemplate} />
          <Column body={actionBodyTemplate} header="操作" style={{ width: '280px' }} />
        </DataTable>
      </Card>

      <Dialog
        header={editingChannel ? '编辑渠道' : '新建渠道'}
        visible={modalVisible}
        onHide={handleModalHide}
        footer={dialogFooter}
        style={{ width: '600px' }}
      >
        <div className="flex flex-column gap-3">
          <div className="flex flex-column gap-2">
            <label className="text-sm">名称</label>
            <InputText
              placeholder="渠道名称"
              value={formValues.name}
              onChange={(e) => setFormValues({ ...formValues, name: e.target.value })}
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">类型</label>
            <Dropdown
              placeholder="选择渠道类型"
              value={formValues.type}
              options={channelTypeOptions}
              optionLabel="label"
              optionValue="value"
              onChange={(e) =>
                setFormValues({ ...formValues, type: e.value, config: {} })
              }
              disabled={!canManageConfig}
            />
          </div>

          <Divider align="center">
            <span className="text-sm">配置信息</span>
          </Divider>

          {renderConfigFields()}

          <Divider />

          <div className="flex align-items-center gap-2">
            <label className="text-sm">启用</label>
            <InputSwitch
              checked={formValues.enabled}
              onChange={(e) => setFormValues({ ...formValues, enabled: e.value })}
              disabled={!canManageConfig}
            />
          </div>
        </div>
      </Dialog>

      <Dialog
        header="发送测试邮件"
        visible={testDialogVisible}
        onHide={() => { setTestDialogVisible(false); setTestRecipients('') }}
        style={{ width: '400px' }}
        footer={
          <div>
            <Button label="取消" icon="pi pi-times" className="p-button-text" onClick={() => { setTestDialogVisible(false); setTestRecipients('') }} />
            <Button label="发送" icon="pi pi-send" onClick={handleTestWithEmail} disabled={!testRecipients.trim()} />
          </div>
        }
      >
        <div className="flex flex-column gap-3">
          <div className="flex flex-column gap-2">
            <label className="text-sm">收件人邮箱</label>
            <InputText
              placeholder="多个邮箱用逗号分隔，如: a@example.com, b@example.com"
              value={testRecipients}
              onChange={(e) => setTestRecipients(e.target.value)}
            />
          </div>
        </div>
      </Dialog>
    </div>
  )
}
