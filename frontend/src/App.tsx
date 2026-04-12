import React from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Layout, Menu, Button, Space, Dropdown, Avatar, Alert } from 'antd'
import {
  DashboardOutlined,
  AlertOutlined,
  DatabaseOutlined,
  SendOutlined,
  BranchesOutlined,
  AudioOutlined,
  CalendarOutlined,
  UserOutlined,
  TeamOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import {
  Dashboard,
  Alerts,
  DataSources,
  Channels,
  RouteRules,
  Silences,
  OnDutyPage,
  Login,
  Users,
  Profile,
} from './pages'
import { useUserStore } from './stores/userStore'
import type { User } from './types'

const { Header, Sider, Content } = Layout

type MenuItem = {
  key: string
  icon: React.ReactNode
  label: string
}

const baseMenuItems: MenuItem[] = [
  { key: '/', icon: <DashboardOutlined />, label: '告警大盘' },
  { key: '/alerts', icon: <AlertOutlined />, label: '告警管理' },
  { key: '/datasources', icon: <DatabaseOutlined />, label: '数据源' },
  { key: '/channels', icon: <SendOutlined />, label: '推送渠道' },
  { key: '/routes', icon: <BranchesOutlined />, label: '路由规则' },
  { key: '/silences', icon: <AudioOutlined />, label: '静默管理' },
  { key: '/onduty', icon: <CalendarOutlined />, label: '值班管理' },
]

const profileMenuItem: MenuItem = { key: '/profile', icon: <UserOutlined />, label: '个人资料' }
const usersMenuItem: MenuItem = { key: '/users', icon: <TeamOutlined />, label: '用户管理' }

function buildMenuItems(user: User | null): MenuItem[] {
  if (!user) {
    return []
  }

  if (user.force_password_reset) {
    return [profileMenuItem]
  }

  const items = [...baseMenuItems, profileMenuItem]
  if (user.role === 'admin') {
    items.push(usersMenuItem)
  }

  return items
}

function AppMenu() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useUserStore((state) => state.user)
  const menuItems = buildMenuItems(user)

  return (
    <Menu
      mode="inline"
      selectedKeys={[location.pathname]}
      items={menuItems}
      style={{ borderRight: 0 }}
      onClick={({ key }) => navigate(key)}
    />
  )
}

function AppHeader() {
  const location = useLocation()
  const navigate = useNavigate()
  const user = useUserStore((state) => state.user)
  const logout = useUserStore((state) => state.logout)
  const menuItems = buildMenuItems(user)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const handleProfile = () => {
    navigate('/profile')
  }

  const currentLabel = menuItems.find((item) => item.key === location.pathname)?.label || '告警系统'

  return (
    <Header
      style={{
        padding: '0 24px',
        background: '#fff',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        borderBottom: '1px solid #f0f0f0',
      }}
    >
      <span style={{ fontSize: 18, fontWeight: 500 }}>{currentLabel}</span>
      <Space>
        {user && (
          <Dropdown
            menu={{
              items: [
                { key: 'profile', icon: <UserOutlined />, label: '个人资料', onClick: handleProfile },
                { type: 'divider' as const },
                { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout },
              ],
            }}
            placement="bottomRight"
          >
            <Button type="text" style={{ height: 'auto', padding: '4px 8px' }}>
              <Space>
                <Avatar size="small" icon={<UserOutlined />} />
                <span>{user.name || user.username}</span>
              </Space>
            </Button>
          </Dropdown>
        )}
      </Space>
    </Header>
  )
}

function MainLayout({ children }: { children: React.ReactNode }) {
  const user = useUserStore((state) => state.user)

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={200} theme="light" style={{ borderRight: '1px solid #f0f0f0' }}>
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #f0f0f0',
            fontWeight: 600,
            fontSize: 16,
          }}
        >
          游戏运维告警系统
        </div>
        <AppMenu />
      </Sider>
      <Layout>
        <AppHeader />
        <Content style={{ padding: '24px', background: '#f0f2f5', minHeight: 'calc(100vh - 64px)' }}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            {user?.force_password_reset && (
              <Alert
                type="warning"
                showIcon
                message="必须先完成密码修改"
                description="当前账号处于强制改密状态，完成密码更新前只能访问个人资料页面。"
              />
            )}
            <div style={{ background: '#fff', borderRadius: 8, padding: 24, minHeight: '100%' }}>
              {children}
            </div>
          </Space>
        </Content>
      </Layout>
    </Layout>
  )
}

function RequireAuth({
  children,
  adminOnly = false,
}: {
  children: React.ReactNode
  adminOnly?: boolean
}) {
  const token = useUserStore((state) => state.token)
  const user = useUserStore((state) => state.user)
  const location = useLocation()

  if (!token) {
    return <Navigate to="/login" replace />
  }

  if (user?.force_password_reset && location.pathname !== '/profile') {
    return <Navigate to="/profile" replace />
  }

  if (adminOnly && user?.role !== 'admin') {
    return <Navigate to="/" replace />
  }

  return <MainLayout>{children}</MainLayout>
}

function LoginPage() {
  const navigate = useNavigate()
  const token = useUserStore((state) => state.token)
  const user = useUserStore((state) => state.user)
  const setUser = useUserStore((state) => state.setUser)
  const setToken = useUserStore((state) => state.setToken)

  if (token) {
    return <Navigate to={user?.force_password_reset ? '/profile' : '/'} replace />
  }

  const handleSuccess = (newToken: string, nextUser: User) => {
    setToken(newToken)
    setUser(nextUser)
    navigate(nextUser.force_password_reset ? '/profile' : '/')
  }

  return <Login onSuccess={handleSuccess} />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<RequireAuth><Dashboard /></RequireAuth>} />
        <Route path="/alerts" element={<RequireAuth><Alerts /></RequireAuth>} />
        <Route path="/datasources" element={<RequireAuth><DataSources /></RequireAuth>} />
        <Route path="/channels" element={<RequireAuth><Channels /></RequireAuth>} />
        <Route path="/routes" element={<RequireAuth><RouteRules /></RequireAuth>} />
        <Route path="/silences" element={<RequireAuth><Silences /></RequireAuth>} />
        <Route path="/onduty" element={<RequireAuth><OnDutyPage /></RequireAuth>} />
        <Route path="/users" element={<RequireAuth adminOnly><Users /></RequireAuth>} />
        <Route path="/profile" element={<RequireAuth><Profile /></RequireAuth>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
