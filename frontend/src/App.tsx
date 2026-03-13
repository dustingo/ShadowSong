import { BrowserRouter, Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { Layout, Menu, Button, Space, Dropdown, Avatar } from 'antd'
import {
  DashboardOutlined,
  AlertOutlined,
  DatabaseOutlined,
  SendOutlined,
  BranchesOutlined,
  AudioOutlined,
  CalendarOutlined,
  RobotOutlined,
  UserOutlined,
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
  AIAssistant,
  Login,
} from './pages'
import { useUserStore } from './stores/userStore'
import { authApi } from './api/auth'
import React from 'react'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '告警大盘' },
  { key: '/alerts', icon: <AlertOutlined />, label: '告警管理' },
  { key: '/datasources', icon: <DatabaseOutlined />, label: '数据源' },
  { key: '/channels', icon: <SendOutlined />, label: '推送渠道' },
  { key: '/routes', icon: <BranchesOutlined />, label: '路由规则' },
  { key: '/silences', icon: <AudioOutlined />, label: '静默管理' },
  { key: '/onduty', icon: <CalendarOutlined />, label: '值班管理' },
  { key: '/ai', icon: <RobotOutlined />, label: 'AI 助手' },
]

// 菜单组件
function AppMenu() {
  const navigate = useNavigate()
  const location = useLocation()

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

// 头部标题
function AppHeader() {
  const location = useLocation()
  const user = useUserStore((state) => state.user)
  const logout = useUserStore((state) => state.logout)
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <Header style={{
      padding: '0 24px',
      background: '#fff',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      borderBottom: '1px solid #f0f0f0',
    }}>
      <span style={{ fontSize: 18, fontWeight: 500 }}>
        {menuItems.find(item => item.key === location.pathname)?.label || '告警系统'}
      </span>
      <Space>
        {user && (
          <Dropdown
            menu={{
              items: [
                { key: 'profile', icon: <UserOutlined />, label: user.name || user.username, disabled: true },
                { type: 'divider' as const },
                { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout },
              ]
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

// 主布局
function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={200} theme="light" style={{ borderRight: '1px solid #f0f0f0' }}>
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderBottom: '1px solid #f0f0f0',
          fontWeight: 600,
          fontSize: 16,
        }}>
          游戏运维 AI 告警系统
        </div>
        <AppMenu />
      </Sider>
      <Layout>
        <AppHeader />
        <Content style={{ padding: '24px', background: '#f0f2f5', minHeight: 'calc(100vh - 64px)' }}>
          <div style={{ background: '#fff', borderRadius: 8, padding: 24, minHeight: '100%' }}>
            {children}
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}

// 路由守卫
function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = useUserStore((state) => state.token)

  if (!token) {
    return <Navigate to="/login" replace />
  }

  return <MainLayout>{children}</MainLayout>
}

// 登录页
function LoginPage() {
  const navigate = useNavigate()
  const token = useUserStore((state) => state.token)
  const setUser = useUserStore((state) => state.setUser)
  const setToken = useUserStore((state) => state.setToken)

  if (token) {
    return <Navigate to="/" replace />
  }

  const handleSuccess = (newToken: string, user: any) => {
    setToken(newToken)
    setUser(user)
    navigate('/')
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
        <Route path="/ai" element={<RequireAuth><AIAssistant /></RequireAuth>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
