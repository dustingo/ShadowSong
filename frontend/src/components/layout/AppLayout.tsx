import React, { useState } from 'react'
import { useLocation } from 'react-router-dom'
import { AppSidebar } from './AppSidebar'
import { AppHeader } from './AppHeader'
import { ToastProvider } from '../common/ToastProvider'

interface AppLayoutProps {
  children: React.ReactNode
}

// 页面标题映射
const pageTitles: Record<string, string> = {
  '/': '告警大盘',
  '/alerts': '告警管理',
  '/deliveries': '通知投递',
  '/ops-health': '运维健康',
  '/datasources': '数据源',
  '/channels': '推送渠道',
  '/routes': '路由规则',
  '/silences': '静默管理',
  '/onduty': '值班管理',
  '/users': '用户管理',
  '/profile': '个人资料',
}

export const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const [pinned, setPinned] = useState(false)
  const [hovered, setHovered] = useState(false)
  const location = useLocation()
  const title = pageTitles[location.pathname] || '告警系统'

  const collapsed = !(pinned || hovered)

  return (
    <ToastProvider>
      <div className="flex h-screen" style={{ background: 'var(--surface-ground)' }}>
        <AppSidebar
          collapsed={collapsed}
          pinned={pinned}
          onPinChange={setPinned}
          onMouseEnter={() => setHovered(true)}
          onMouseLeave={() => setHovered(false)}
        />
        <div className="flex flex-column flex-1 overflow-hidden">
          <AppHeader
            title={title}
            pinned={pinned}
            onPinChange={setPinned}
          />
          <main className="flex-1 overflow-auto p-4">
            {children}
          </main>
        </div>
      </div>
    </ToastProvider>
  )
}