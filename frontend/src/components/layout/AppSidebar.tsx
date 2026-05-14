import React from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

interface MenuItem {
  key: string
  icon: string
  label: string
}

const menuItems: MenuItem[] = [
  { key: '/', icon: 'pi pi-home', label: '告警大盘' },
  { key: '/alerts', icon: 'pi pi-bell', label: '告警管理' },
  { key: '/deliveries', icon: 'pi pi-send', label: '通知投递' },
  { key: '/ops-health', icon: 'pi pi-chart-bar', label: '运维健康' },
  { key: '/datasources', icon: 'pi pi-database', label: '数据源' },
  { key: '/channels', icon: 'pi pi-telegram', label: '推送渠道' },
  { key: '/routes', icon: 'pi pi-sitemap', label: '路由规则' },
  { key: '/silences', icon: 'pi pi-volume-off', label: '静默管理' },
  { key: '/onduty', icon: 'pi pi-calendar', label: '值班管理' },
]

interface AppSidebarProps {
  collapsed: boolean
}

export const AppSidebar: React.FC<AppSidebarProps> = ({ collapsed }) => {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <div
      className="flex flex-column transition-all transition-duration-300"
      style={{
        width: collapsed ? '60px' : '220px',
        background: '#ffffff',
        borderRight: '1px solid #e2e8f0',
        overflow: 'hidden',
        height: '100%',
      }}
    >
      {/* Logo area */}
      <div
        className="flex align-items-center justify-content-center"
        style={{
          height: '60px',
          borderBottom: '1px solid #e2e8f0',
        }}
      >
        {collapsed ? (
          <i className="pi pi-bolt text-2xl" style={{ color: '#10b981' }} />
        ) : (
          <span className="text-lg font-semibold text-slate-700">
            <i className="pi pi-bolt mr-2" style={{ color: '#10b981' }} />
            告警系统
          </span>
        )}
      </div>

      {/* Menu */}
      <div className="flex-1 py-3 overflow-auto">
        {menuItems.map((item) => {
          const isActive = location.pathname === item.key
          return (
            <div
              key={item.key}
              className={`
                flex align-items-center cursor-pointer transition-all transition-duration-200
                ${isActive ? 'bg-emerald-50' : 'hover:bg-slate-50'}
              `}
              style={{
                padding: '12px 16px',
                marginLeft: collapsed ? '0' : '8px',
                marginRight: collapsed ? '0' : '8px',
                borderRadius: collapsed ? '0' : '8px',
                justifyContent: collapsed ? 'center' : 'flex-start',
                borderRight: isActive ? '3px solid #10b981' : 'none',
              }}
              onClick={() => navigate(item.key)}
            >
              <i
                className={`${item.icon} text-lg`}
                style={{ minWidth: '24px', color: isActive ? '#10b981' : '#64748b' }}
              />
              {!collapsed && (
                <span
                  className="ml-3 text-sm"
                  style={{ color: isActive ? '#10b981' : '#475569' }}
                >
                  {item.label}
                </span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}