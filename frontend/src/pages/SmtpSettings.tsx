import { useEffect, useState } from 'react'
import { Button } from 'primereact/button'
import { InputText } from 'primereact/inputtext'
import { InputNumber } from 'primereact/inputnumber'
import { InputSwitch } from 'primereact/inputswitch'
import { Card } from 'primereact/card'
import { Dialog } from 'primereact/dialog'
import { useConfigStore } from '../stores/configStore'
import { useToast } from '../hooks/useToast'
import { getApiErrorMessage } from '../utils/api'
import type { SmtpConfig } from '../types'

const defaultConfig: SmtpConfig = {
  host: '',
  port: 465,
  username: '',
  password: '',
  from_addr: '',
  from_name: '告警系统',
  tls: true,
  enabled: true,
  updated_at: '',
}

export default function SmtpSettings() {
  const { smtpConfig, fetchSmtpConfig, updateSmtpConfig, testSmtpConfig } = useConfigStore()
  const toast = useToast()
  const [form, setForm] = useState<SmtpConfig>(defaultConfig)
  const [testRecipients, setTestRecipients] = useState('')
  const [testDialogVisible, setTestDialogVisible] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    fetchSmtpConfig()
  }, [fetchSmtpConfig])

  useEffect(() => {
    if (smtpConfig) {
      setForm({ ...smtpConfig, password: '' })
    }
  }, [smtpConfig])

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload: Partial<SmtpConfig> = { ...form }
      if (!payload.password) {
        delete payload.password
      }
      await updateSmtpConfig(payload)
      toast.showSuccess('SMTP 配置已保存')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  const handleTest = async () => {
    const recipients = testRecipients.split(',').map(s => s.trim()).filter(Boolean)
    if (recipients.length === 0) return
    try {
      const msg = await testSmtpConfig(recipients)
      toast.showSuccess(msg || '测试邮件已发送')
      setTestDialogVisible(false)
      setTestRecipients('')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '发送测试邮件失败'))
    }
  }

  return (
    <div className="p-4">
      <div className="flex justify-content-between align-items-center mb-4">
        <h2 className="m-0">SMTP 邮件服务配置</h2>
        <div className="flex gap-2">
          <Button
            label="测试连接"
            icon="pi pi-send"
            severity="info"
            onClick={() => setTestDialogVisible(true)}
            disabled={!form.host}
          />
          <Button
            label="保存"
            icon="pi pi-check"
            onClick={handleSave}
            loading={saving}
          />
        </div>
      </div>

      <Card>
        <div className="grid">
          <div className="col-6">
            <div className="flex flex-column gap-2 mb-3">
              <label className="text-sm font-medium">SMTP 服务器 *</label>
              <InputText
                value={form.host}
                onChange={(e) => setForm({ ...form, host: e.target.value })}
                placeholder="smtp.example.com"
              />
            </div>
          </div>
          <div className="col-6">
            <div className="flex flex-column gap-2 mb-3">
              <label className="text-sm font-medium">端口 *</label>
              <InputNumber
                value={form.port}
                onValueChange={(e) => setForm({ ...form, port: e.value ?? 465 })}
                min={1}
                max={65535}
              />
            </div>
          </div>
          <div className="col-6">
            <div className="flex flex-column gap-2 mb-3">
              <label className="text-sm font-medium">用户名 *</label>
              <InputText
                value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })}
                placeholder="user@example.com"
              />
            </div>
          </div>
          <div className="col-6">
            <div className="flex flex-column gap-2 mb-3">
              <label className="text-sm font-medium">密码</label>
              <InputText
                type="password"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                placeholder={smtpConfig?.password === '****' ? '已配置，留空保持不变' : ''}
              />
            </div>
          </div>
          <div className="col-6">
            <div className="flex flex-column gap-2 mb-3">
              <label className="text-sm font-medium">发件人地址 *</label>
              <InputText
                value={form.from_addr}
                onChange={(e) => setForm({ ...form, from_addr: e.target.value })}
                placeholder="alert@example.com"
              />
            </div>
          </div>
          <div className="col-6">
            <div className="flex flex-column gap-2 mb-3">
              <label className="text-sm font-medium">发件人名称</label>
              <InputText
                value={form.from_name}
                onChange={(e) => setForm({ ...form, from_name: e.target.value })}
                placeholder="告警系统"
              />
            </div>
          </div>
          <div className="col-6">
            <div className="flex align-items-center gap-2 mt-1">
              <InputSwitch
                checked={form.tls}
                onChange={(e) => setForm({ ...form, tls: e.value })}
              />
              <label className="text-sm">启用 TLS/SSL</label>
            </div>
          </div>
          <div className="col-6">
            <div className="flex align-items-center gap-2 mt-1">
              <InputSwitch
                checked={form.enabled}
                onChange={(e) => setForm({ ...form, enabled: e.value })}
              />
              <label className="text-sm">启用邮件服务</label>
            </div>
          </div>
        </div>
      </Card>

      <Dialog
        header="发送测试邮件"
        visible={testDialogVisible}
        onHide={() => { setTestDialogVisible(false); setTestRecipients('') }}
        style={{ width: '400px' }}
        footer={
          <div>
            <Button label="取消" icon="pi pi-times" className="p-button-text" onClick={() => { setTestDialogVisible(false); setTestRecipients('') }} />
            <Button label="发送" icon="pi pi-send" onClick={handleTest} disabled={!testRecipients.trim()} />
          </div>
        }
      >
        <div className="flex flex-column gap-2">
          <label className="text-sm">收件人邮箱</label>
          <InputText
            placeholder="多个邮箱用逗号分隔，如: a@example.com, b@example.com"
            value={testRecipients}
            onChange={(e) => setTestRecipients(e.target.value)}
          />
        </div>
      </Dialog>
    </div>
  )
}
