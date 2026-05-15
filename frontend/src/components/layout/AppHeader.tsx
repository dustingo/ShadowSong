import React, { useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from 'primereact/button'
import { Avatar } from 'primereact/avatar'
import { InputText } from 'primereact/inputtext'
import { Menu } from 'primereact/menu'
import { useUserStore } from '../../stores/userStore'

interface AppHeaderProps {
  title: string
  onToggleSidebar: () => void
  sidebarCollapsed: boolean
}

export const AppHeader: React.FC<AppHeaderProps> = ({
  title,
  onToggleSidebar,
  sidebarCollapsed,
}) => {
  const navigate = useNavigate()
  const user = useUserStore((state) => state.user)
  const logout = useUserStore((state) => state.logout)

  const menuRef = useRef<Menu>(null)

  const userMenuItems = [
    {
      label: '个人资料',
      icon: 'pi pi-user',
      command: () => navigate('/profile'),
    },
    { separator: true },
    {
      label: '退出登录',
      icon: 'pi pi-sign-out',
      command: () => {
        logout()
        navigate('/login')
      },
    },
  ]

  return (
    <header
      className="flex align-items-center justify-content-between px-4"
      style={{
        height: '60px',
        background: 'var(--surface-card)',
        borderBottom: '1px solid var(--surface-border)',
      }}
    >
      <div className="flex align-items-center gap-3">
        <Button
          icon={sidebarCollapsed ? 'pi pi-bars' : 'pi pi-times'}
          text
          rounded
          severity="secondary"
          onClick={onToggleSidebar}
          style={{ color: 'var(--text-secondary)' }}
        />
        <span className="text-lg font-medium" style={{ color: 'var(--text-primary)' }}>{title}</span>
      </div>

      <div className="flex align-items-center gap-3">
        <span className="p-input-icon-left hidden md:block">
          <i className="pi pi-search" style={{ color: 'var(--text-disabled)' }} />
          <InputText
            placeholder="搜索..."
            className="p-inputtext-sm"
            style={{ width: '220px' }}
          />
        </span>

        <div className="flex align-items-center gap-2 cursor-pointer">
          <Avatar
            label={user?.name?.charAt(0) || user?.username?.charAt(0) || 'U'}
            shape="circle"
            style={{ backgroundColor: 'var(--success-color)', color: 'var(--text-inverse)' }}
            size="normal"
            onClick={(e) => menuRef.current?.toggle(e)}
          />
          <span className="text-sm hidden md:block" style={{ color: 'var(--text-secondary)' }}>
            {user?.name || user?.username}
          </span>
          <Menu model={userMenuItems} popup ref={menuRef} />
        </div>
      </div>
    </header>
  )
}
