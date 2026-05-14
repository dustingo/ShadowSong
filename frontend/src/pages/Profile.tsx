import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Password } from 'primereact/password'
import { Button } from 'primereact/button'
import { useNavigate } from 'react-router-dom'
import { authApi } from '../api/auth'
import { useUserStore } from '../stores/userStore'
import { useToast } from '../components'

export const Profile: React.FC = () => {
  const navigate = useNavigate()
  const toast = useToast()
  const user = useUserStore((state) => state.user)
  const setUser = useUserStore((state) => state.setUser)
  const logout = useUserStore((state) => state.logout)

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  useEffect(() => {
    if (user) {
      setName(user.name)
      setEmail(user.email || '')
    }
  }, [user])

  const handleProfileSubmit = async () => {
    if (!name.trim()) {
      toast.showError('请输入姓名')
      return
    }
    try {
      const nextUser = await authApi.updateOwnProfile({ name, email: email || undefined })
      setUser(nextUser)
      toast.showSuccess('个人资料已更新')
    } catch {
      toast.showError('更新个人资料失败')
    }
  }

  const handlePasswordSubmit = async () => {
    if (!password.trim()) {
      toast.showError('请输入新密码')
      return
    }
    try {
      await authApi.updateOwnPassword(password)
      logout()
      toast.showSuccess('密码已更新，请重新登录')
      navigate('/login')
    } catch {
      toast.showError('更新密码失败')
    }
  }

  return (
    <div className="flex flex-column gap-4">
      <Card title="个人资料">
        <div className="flex flex-column gap-3">
          <div className="flex flex-column gap-2">
            <label className="font-medium">用户名</label>
            <InputText value={user?.username || ''} disabled className="p-inputtext-sm" />
          </div>
          <div className="flex flex-column gap-2">
            <label className="font-medium">姓名 *</label>
            <InputText
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="姓名"
              className="p-inputtext-sm"
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="font-medium">邮箱</label>
            <InputText
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="邮箱"
              className="p-inputtext-sm"
            />
          </div>
          <Button label="保存资料" onClick={handleProfileSubmit} className="w-fit" />
        </div>
      </Card>

      <Card title="修改密码">
        <div className="flex flex-column gap-3">
          <div className="flex flex-column gap-2">
            <label className="font-medium">新密码 *</label>
            <Password
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入新密码"
              feedback={false}
              toggleMask
              className="p-inputtext-sm"
            />
          </div>
          <Button label="更新密码" onClick={handlePasswordSubmit} className="w-fit" />
        </div>
      </Card>
    </div>
  )
}
