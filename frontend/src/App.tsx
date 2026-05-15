import React from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { ConfirmDialog } from 'primereact/confirmdialog'
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
  Deliveries,
  OpsHealth,
  ColorDemo,
} from './pages'
import { AppLayout } from './components/layout/AppLayout'
import {
  canUser,
  capabilityManageUsers,
  capabilityViewConfig,
} from './authz/capabilities'
import { useUserStore } from './stores/userStore'
import type { User } from './types'
import { PermissionNotice } from './components'

const routerFuture = {
  v7_startTransition: true,
  v7_relativeSplatPath: true,
} as const

function RequireAuth({
  children,
  requiredCapability,
}: {
  children: React.ReactNode
  requiredCapability?: typeof capabilityManageUsers | typeof capabilityViewConfig
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

  if (requiredCapability && !canUser(user, requiredCapability)) {
    return (
      <AppLayout>
        <PermissionNotice />
      </AppLayout>
    )
  }

  return <AppLayout>{children}</AppLayout>
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
    <BrowserRouter future={routerFuture}>
      <ConfirmDialog />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<RequireAuth><Dashboard /></RequireAuth>} />
        <Route path="/alerts" element={<RequireAuth><Alerts /></RequireAuth>} />
        <Route path="/deliveries" element={<RequireAuth requiredCapability={capabilityViewConfig}><Deliveries /></RequireAuth>} />
        <Route path="/ops-health" element={<RequireAuth requiredCapability={capabilityViewConfig}><OpsHealth /></RequireAuth>} />
        <Route path="/datasources" element={<RequireAuth requiredCapability={capabilityViewConfig}><DataSources /></RequireAuth>} />
        <Route path="/channels" element={<RequireAuth requiredCapability={capabilityViewConfig}><Channels /></RequireAuth>} />
        <Route path="/routes" element={<RequireAuth requiredCapability={capabilityViewConfig}><RouteRules /></RequireAuth>} />
        <Route path="/silences" element={<RequireAuth requiredCapability={capabilityViewConfig}><Silences /></RequireAuth>} />
        <Route path="/onduty" element={<RequireAuth requiredCapability={capabilityViewConfig}><OnDutyPage /></RequireAuth>} />
        <Route path="/users" element={<RequireAuth requiredCapability={capabilityManageUsers}><Users /></RequireAuth>} />
        <Route path="/profile" element={<RequireAuth><Profile /></RequireAuth>} />
        <Route path="/color-demo" element={<ColorDemo />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}