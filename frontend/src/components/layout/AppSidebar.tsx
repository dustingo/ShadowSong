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
        background: 'var(--surface-card)',
        borderRight: '1px solid var(--surface-border)',
        overflow: 'hidden',
        height: '100%',
      }}
    >
      {/* Logo area */}
      <div
        className="flex align-items-center justify-content-center"
        style={{
          height: '60px',
          borderBottom: '1px solid var(--surface-border)',
        }}
      >
        {collapsed ? (
          <i className="pi pi-bolt text-2xl" style={{ color: 'var(--primary-color)' }} />
        ) : (
          <span className="text-lg font-semibold" style={{ color: 'var(--text-primary)' }}>
            <i className="pi pi-bolt mr-2" style={{ color: 'var(--primary-color)' }} />
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
              `}
              style={{
                padding: '12px 16px',
                marginLeft: collapsed ? '0' : '8px',
                marginRight: collapsed ? '0' : '8px',
                borderRadius: collapsed ? '0' : '8px',
                justifyContent: collapsed ? 'center' : 'flex-start',
                background: isActive ? 'var(--primary-light-color)' : 'transparent',
                borderRight: isActive ? '3px solid var(--primary-color)' : 'none',
              }}
              onClick={() => navigate(item.key)}
              onMouseEnter={(e) => {
                if (!isActive) {
                  e.currentTarget.style.background = 'var(--surface-hover)'
                }
              }}
              onMouseLeave={(e) => {
                if (!isActive) {
                  e.currentTarget.style.background = 'transparent'
                }
              }}
            >
              <i
                className={`${item.icon} text-lg`}
                style={{ minWidth: '24px', color: isActive ? 'var(--primary-color)' : 'var(--text-secondary)' }}
              />
              {!collapsed && (
                <span
                  className="ml-3 text-sm"
                  style={{ color: isActive ? 'var(--primary-color)' : 'var(--text-secondary)' }}
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