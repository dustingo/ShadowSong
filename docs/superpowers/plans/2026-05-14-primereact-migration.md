# PrimeReact 迁移实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将游戏运维告警系统的前端从 Ant Design 迁移到 PrimeReact，采用 Freya 模板的浅色主题风格。

**Architecture:** 使用 PrimeReact 组件库替换所有 Ant Design 组件，创建自定义的可折叠侧边栏布局，通过 Toast 替代 message 全局提示，保持现有的 Zustand 状态管理和 React Router 路由。

**Tech Stack:** React 18, TypeScript, PrimeReact 10, PrimeIcons, PrimeFlex, ECharts, Monaco Editor, Zustand, React Router

---

## 文件结构

### 新建文件
- `frontend/src/components/layout/AppLayout.tsx` - 主布局组件（Sidebar + Header）
- `frontend/src/components/layout/AppSidebar.tsx` - 可折叠侧边栏
- `frontend/src/components/layout/AppHeader.tsx` - 顶部导航栏
- `frontend/src/components/common/StatisticCard.tsx` - 统计卡片组件
- `frontend/src/components/common/ToastProvider.tsx` - Toast 全局提供者
- `frontend/src/primereact-locale.ts` - 中文 locale 配置

### 修改文件
- `frontend/src/main.tsx` - 添加 PrimeReact Provider 和样式
- `frontend/src/App.tsx` - 使用新布局组件
- `frontend/src/pages/Login.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Dashboard.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Alerts.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Deliveries.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/OpsHealth.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/DataSources.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Channels.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/RouteRules.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Silences.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/OnDuty.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Users.tsx` - 迁移到 PrimeReact
- `frontend/src/pages/Profile.tsx` - 迁移到 PrimeReact
- `frontend/src/components/AlertCard.tsx` - 迁移到 PrimeReact
- `frontend/src/components/SeverityBadge.tsx` - 迁移到 PrimeReact
- `frontend/src/components/PermissionNotice.tsx` - 迁移到 PrimeReact
- `frontend/src/components/index.ts` - 更新导出

### 删除文件
- `frontend/src/FreyaDemo.tsx` - Demo 文件（迁移完成后删除）
- `frontend/src/main-demo.tsx` - Demo 入口（迁移完成后删除）
- `frontend/index-demo.html` - Demo HTML（迁移完成后删除）
- `frontend/vite.demo.config.ts` - Demo 配置（迁移完成后删除）

---

## 阶段 1: 基础设施

### Task 1: 配置 PrimeReact 全局样式和中文 locale

**Files:**
- Create: `frontend/src/primereact-locale.ts`
- Modify: `frontend/src/main.tsx`

- [ ] **Step 1: 创建中文 locale 配置文件**

创建 `frontend/src/primereact-locale.ts`:

```typescript
import { addLocale, locale } from 'primereact/api'

addLocale('zh-CN', {
  startsWith: '以...开始',
  contains: '包含',
  notContains: '不包含',
  endsWith: '以...结束',
  equals: '等于',
  notEquals: '不等于',
  noFilter: '无筛选',
  filter: '筛选',
  lt: '小于',
  lte: '小于或等于',
  gt: '大于',
  gte: '大于或等于',
  dateIs: '日期是',
  dateIsNot: '日期不是',
  dateBefore: '日期早于',
  dateAfter: '日期晚于',
  clear: '清空',
  apply: '应用',
  matchAll: '匹配全部',
  matchAny: '匹配任意',
  addRule: '添加规则',
  removeRule: '移除规则',
  accept: '确定',
  reject: '取消',
  choose: '选择',
  upload: '上传',
  cancel: '取消',
  completed: '已完成',
  pending: '待处理',
  dayNames: ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六'],
  dayNamesShort: ['日', '一', '二', '三', '四', '五', '六'],
  dayNamesMin: ['日', '一', '二', '三', '四', '五', '六'],
  monthNames: ['一月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月'],
  monthNamesShort: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月'],
  chooseYear: '选择年份',
  chooseMonth: '选择月份',
  chooseDate: '选择日期',
  prevDecade: '上一年代',
  nextDecade: '下一代',
  prevYear: '上一年',
  nextYear: '下一年',
  prevMonth: '上一月',
  nextMonth: '下一月',
  prevHour: '上一小时',
  nextHour: '下一小时',
  prevMinute: '上一分钟',
  nextMinute: '下一分钟',
  prevSecond: '上一秒',
  nextSecond: '下一秒',
  am: '上午',
  pm: '下午',
  today: '今天',
  now: '现在',
  weekHeader: '周',
  firstDayOfWeek: 1,
  dateFormat: 'yy-mm-dd',
  weak: '弱',
  medium: '中',
  strong: '强',
  passwordPrompt: '请输入密码',
  emptyMessage: '无可用选项',
  emptyFilterMessage: '未找到结果',
  emptySearchMessage: '未找到结果',
  selectionMessage: '已选择 {0} 项',
  emptySelectionMessage: '无选中项',
  aria: {
    trueLabel: '是',
    falseLabel: '否',
    nullLabel: '未选择',
    star: '1星',
    stars: '{star}星',
    selectAll: '全选',
    unselectAll: '取消全选',
    close: '关闭',
    previous: '上一页',
    next: '下一页',
    navigation: '导航',
    scrollTop: '滚动到顶部',
    moveTop: '移动到顶部',
    moveUp: '上移',
    moveDown: '下移',
    moveBottom: '移动到底部',
    moveToTarget: '移动到目标',
    moveToSource: '移动到源',
    moveAllToTarget: '全部移动到目标',
    moveAllToSource: '全部移动到源',
    pageLabel: '第{page}页',
    firstPageLabel: '第一页',
    lastPageLabel: '最后一页',
    nextPageLabel: '下一页',
    prevPageLabel: '上一页',
    rowsPerPageLabel: '每页行数',
    jumpToPageDropdownLabel: '跳转到页面',
    jumpToPageInputLabel: '跳转到页面',
    selectRow: '选择行',
    unselectRow: '取消选择行',
    expandRow: '展开行',
    collapseRow: '折叠行',
    showFilterMenu: '显示筛选菜单',
    hideFilterMenu: '隐藏筛选菜单',
    filterOperator: '筛选操作符',
    filterConstraint: '筛选条件',
    editRow: '编辑行',
    saveEdit: '保存编辑',
    cancelEdit: '取消编辑',
    listView: '列表视图',
    gridView: '网格视图',
    slide: '幻灯片',
    slideNumber: '{slideNumber}',
    zoomImage: '放大图片',
    zoomIn: '放大',
    zoomOut: '缩小',
    rotateRight: '向右旋转',
    rotateLeft: '向左旋转',
  },
})

export const initLocale = () => {
  locale('zh-CN')
}
```

- [ ] **Step 2: 修改 main.tsx 添加 PrimeReact 配置**

修改 `frontend/src/main.tsx`:

```typescript
import React from 'react'
import ReactDOM from 'react-dom/client'
import { PrimeReactProvider } from 'primereact/api'
import App from './App'
import { initLocale } from './primereact-locale'
import './index.css'

// PrimeReact 样式
import 'primereact/resources/themes/lara-light-indigo/theme.css'
import 'primereact/resources/primereact.min.css'
import 'primeicons/primeicons.css'
import 'primeflex/primeflex.css'

initLocale()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <PrimeReactProvider>
      <App />
    </PrimeReactProvider>
  </React.StrictMode>
)
```

- [ ] **Step 3: 验证应用能正常启动**

Run: `cd frontend && pnpm dev`
Expected: 应用在 http://localhost:5173 启动，控制台无报错

- [ ] **Step 4: 提交**

```bash
git add frontend/src/primereact-locale.ts frontend/src/main.tsx
git commit -m "feat: add PrimeReact global config and Chinese locale"
```

---

### Task 2: 创建 Toast 全局提供者

**Files:**
- Create: `frontend/src/components/common/ToastProvider.tsx`

- [ ] **Step 1: 创建 ToastProvider 组件**

创建 `frontend/src/components/common/ToastProvider.tsx`:

```typescript
import React, { createContext, useContext, useRef, ReactNode } from 'react'
import { Toast } from 'primereact/toast'
import type { ToastMessage } from 'primereact/toast'

interface ToastContextType {
  show: (message: ToastMessage) => void
  showSuccess: (summary: string, detail?: string) => void
  showError: (summary: string, detail?: string) => void
  showInfo: (summary: string, detail?: string) => void
  showWarn: (summary: string, detail?: string) => void
  clear: () => void
}

const ToastContext = createContext<ToastContextType | null>(null)

export const useToast = (): ToastContextType => {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context
}

interface ToastProviderProps {
  children: ReactNode
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const toastRef = useRef<Toast>(null)

  const show = (message: ToastMessage) => {
    toastRef.current?.show(message)
  }

  const showSuccess = (summary: string, detail?: string) => {
    show({ severity: 'success', summary, detail, life: 3000 })
  }

  const showError = (summary: string, detail?: string) => {
    show({ severity: 'error', summary, detail, life: 5000 })
  }

  const showInfo = (summary: string, detail?: string) => {
    show({ severity: 'info', summary, detail, life: 3000 })
  }

  const showWarn = (summary: string, detail?: string) => {
    show({ severity: 'warn', summary, detail, life: 4000 })
  }

  const clear = () => {
    toastRef.current?.clear()
  }

  return (
    <ToastContext.Provider value={{ show, showSuccess, showError, showInfo, showWarn, clear }}>
      <Toast ref={toastRef} position="top-right" />
      {children}
    </ToastContext.Provider>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/common/ToastProvider.tsx
git commit -m "feat: add ToastProvider for global toast notifications"
```

---

### Task 3: 创建统计卡片组件

**Files:**
- Create: `frontend/src/components/common/StatisticCard.tsx`

- [ ] **Step 1: 创建 StatisticCard 组件**

创建 `frontend/src/components/common/StatisticCard.tsx`:

```typescript
import React from 'react'
import { Card } from 'primereact/card'

interface StatisticCardProps {
  title: string
  value: number | string
  icon: string
  color: string
  suffix?: string
}

export const StatisticCard: React.FC<StatisticCardProps> = ({
  title,
  value,
  icon,
  color,
  suffix,
}) => {
  return (
    <Card className="shadow-sm border-0">
      <div className="flex align-items-center justify-content-between">
        <div>
          <div className="text-sm text-slate-500 mb-1">{title}</div>
          <div className="text-3xl font-bold" style={{ color }}>
            {value}
            {suffix && <span className="text-lg ml-1">{suffix}</span>}
          </div>
        </div>
        <div
          className="flex align-items-center justify-content-center"
          style={{
            width: '48px',
            height: '48px',
            borderRadius: '12px',
            background: `${color}15`,
          }}
        >
          <i className={`${icon} text-xl`} style={{ color }} />
        </div>
      </div>
    </Card>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/common/StatisticCard.tsx
git commit -m "feat: add StatisticCard component"
```

---

### Task 4: 创建可折叠侧边栏组件

**Files:**
- Create: `frontend/src/components/layout/AppSidebar.tsx`

- [ ] **Step 1: 创建 AppSidebar 组件**

创建 `frontend/src/components/layout/AppSidebar.tsx`:

```typescript
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
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/layout/AppSidebar.tsx
git commit -m "feat: add AppSidebar component with collapsible menu"
```

---

### Task 5: 创建顶部导航栏组件

**Files:**
- Create: `frontend/src/components/layout/AppHeader.tsx`

- [ ] **Step 1: 创建 AppHeader 组件**

创建 `frontend/src/components/layout/AppHeader.tsx`:

```typescript
import React from 'react'
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

  const menuRef = React.useRef<Menu>(null)

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
        background: 'white',
        borderBottom: '1px solid #e2e8f0',
      }}
    >
      <div className="flex align-items-center gap-3">
        <Button
          icon={`pi ${sidebarCollapsed ? 'pi pi-bars' : 'pi pi-times'}`}
          text
          rounded
          onClick={onToggleSidebar}
        />
        <span className="text-lg font-medium text-slate-700">{title}</span>
      </div>

      <div className="flex align-items-center gap-3">
        <span className="p-input-icon-left hidden md:block">
          <i className="pi pi-search" style={{ color: '#94a3b8' }} />
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
            style={{ backgroundColor: '#10b981', color: 'white' }}
            size="small"
            onClick={(e) => menuRef.current?.toggle(e)}
          />
          <span className="text-sm text-slate-600 hidden md:block">
            {user?.name || user?.username}
          </span>
          <Menu model={userMenuItems} popup ref={menuRef} />
        </div>
      </div>
    </header>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/layout/AppHeader.tsx
git commit -m "feat: add AppHeader component with user menu"
```

---

### Task 6: 创建主布局组件

**Files:**
- Create: `frontend/src/components/layout/AppLayout.tsx`

- [ ] **Step 1: 创建 AppLayout 组件**

创建 `frontend/src/components/layout/AppLayout.tsx`:

```typescript
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
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const location = useLocation()
  const title = pageTitles[location.pathname] || '告警系统'

  return (
    <ToastProvider>
      <div className="flex h-screen" style={{ background: '#f8fafc' }}>
        <AppSidebar collapsed={sidebarCollapsed} />
        <div className="flex flex-column flex-1 overflow-hidden">
          <AppHeader
            title={title}
            onToggleSidebar={() => setSidebarCollapsed(!sidebarCollapsed)}
            sidebarCollapsed={sidebarCollapsed}
          />
          <main className="flex-1 overflow-auto p-4">
            {children}
          </main>
        </div>
      </div>
    </ToastProvider>
  )
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/layout/AppLayout.tsx
git commit -m "feat: add AppLayout component with sidebar and header"
```

---

### Task 7: 更新组件导出

**Files:**
- Modify: `frontend/src/components/index.ts`

- [ ] **Step 1: 更新组件导出文件**

修改 `frontend/src/components/index.ts`:

```typescript
export { AlertCard } from './AlertCard'
export { SeverityBadge } from './SeverityBadge'
export { PermissionNotice } from './PermissionNotice'
export { CodeEditor } from './CodeEditor'
export { StatisticCard } from './common/StatisticCard'
export { ToastProvider, useToast } from './common/ToastProvider'
export { AppLayout } from './layout/AppLayout'
export { AppSidebar } from './layout/AppSidebar'
export { AppHeader } from './layout/AppHeader'
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/components/index.ts
git commit -m "feat: update component exports"
```

---

## 阶段 2: 核心页面迁移

### Task 8: 迁移 Login 页面

**Files:**
- Modify: `frontend/src/pages/Login.tsx`

- [ ] **Step 1: 重写 Login 页面使用 PrimeReact**

修改 `frontend/src/pages/Login.tsx`:

```typescript
import React, { useState } from 'react'
import { Card } from 'primereact/card'
import { InputText } from 'primereact/inputtext'
import { Password } from 'primereact/password'
import { Button } from 'primereact/button'
import { Message } from 'primereact/message'
import axios from 'axios'
import { getApiErrorMessage } from '../api/client'
import type { User } from '../types'

interface LoginProps {
  onSuccess?: (token: string, user: User) => void
}

interface LoginResponse {
  token: string
  user: User
}

export const Login: React.FC<LoginProps> = ({ onSuccess }) => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username || !password) {
      setError('请输入用户名和密码')
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await axios.post<LoginResponse>('/api/v1/auth/login', {
        username,
        password,
      })

      const { token, user } = res.data

      if (onSuccess) {
        onSuccess(token, user)
      } else {
        localStorage.setItem('token', token)
        localStorage.setItem('user', JSON.stringify(user))
        window.location.href = '/'
      }
    } catch (err) {
      setError(getApiErrorMessage(err, '登录失败'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="flex align-items-center justify-content-center"
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '20px',
      }}
    >
      <Card
        style={{
          width: '400px',
          boxShadow: '0 8px 32px rgba(0,0,0,0.2)',
          borderRadius: '12px',
        }}
      >
        <div className="text-center mb-4">
          <h2 className="text-xl font-semibold text-slate-700 mb-2">
            游戏运维告警系统
          </h2>
          <p className="text-slate-500 text-sm">请登录您的账户</p>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="flex flex-column gap-3">
            {error && (
              <Message severity="error" text={error} />
            )}

            <div className="flex flex-column gap-2">
              <label htmlFor="username" className="text-sm text-slate-600">
                用户名
              </label>
              <InputText
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="请输入用户名"
                className="w-full"
              />
            </div>

            <div className="flex flex-column gap-2">
              <label htmlFor="password" className="text-sm text-slate-600">
                密码
              </label>
              <Password
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                feedback={false}
                toggleMask
                className="w-full"
                inputClassName="w-full"
              />
            </div>

            <Button
              type="submit"
              label="登 录"
              loading={loading}
              className="w-full mt-2"
              style={{ height: '44px' }}
            />
          </div>
        </form>

        <p className="text-center text-slate-400 text-xs mt-4">
          如需创建或重置账户，请联系管理员
        </p>
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: 验证登录页面**

Run: `cd frontend && pnpm dev`
Expected: 访问 /login 能看到新的登录页面，样式正确

- [ ] **Step 3: 提交**

```bash
git add frontend/src/pages/Login.tsx
git commit -m "feat: migrate Login page to PrimeReact"
```

---

### Task 9: 迁移 App.tsx 主布局

**Files:**
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: 重写 App.tsx 使用新布局组件**

修改 `frontend/src/App.tsx`:

```typescript
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
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
```

- [ ] **Step 2: 验证主布局**

Run: `cd frontend && pnpm dev`
Expected: 登录后能看到新的侧边栏和顶部导航

- [ ] **Step 3: 提交**

```bash
git add frontend/src/App.tsx
git commit -m "feat: migrate App.tsx to use new PrimeReact layout"
```

---

### Task 10: 迁移 Dashboard 页面

**Files:**
- Modify: `frontend/src/pages/Dashboard.tsx`

- [ ] **Step 1: 重写 Dashboard 页面使用 PrimeReact**

修改 `frontend/src/pages/Dashboard.tsx`:

```typescript
import React, { useEffect, useRef } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Tag } from 'primereact/tag'
import { Button } from 'primereact/button'
import { Message } from 'primereact/message'
import { ProgressSpinner } from 'primereact/progressspinner'
import { Chart } from 'primereact/chart'
import { useAlertStore } from '../stores/alertStore'
import { useUserStore } from '../stores/userStore'
import { AlertCard } from '../components/AlertCard'
import { StatisticCard, useToast } from '../components'
import type { Alert as AlertItem } from '../types'

export const Dashboard: React.FC = () => {
  const token = useUserStore((state) => state.token)
  const toast = useToast()
  const {
    activeAlerts,
    stats,
    loading,
    wsConnected,
    fetchActiveAlerts,
    fetchStats,
    setWsConnected,
    ackAlert,
    quickSilence,
  } = useAlertStore()

  const handleAck = async (alert: AlertItem) => {
    try {
      await ackAlert(alert.alert_id, '')
      toast.showSuccess('已确认')
    } catch {
      toast.showError('确认失败')
    }
  }

  const handleQuickSilence = async (alert: AlertItem) => {
    try {
      await quickSilence(alert.alert_id, 3600)
      toast.showSuccess('已静默')
    } catch {
      toast.showError('静默失败')
    }
  }

  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    fetchActiveAlerts()
    fetchStats()

    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let isConnecting = false

    const connectWS = () => {
      if (isConnecting || wsRef.current?.readyState === WebSocket.OPEN) {
        return
      }
      if (!token) {
        setWsConnected(false)
        return
      }

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws/alerts?token=${encodeURIComponent(token)}`

      try {
        isConnecting = true
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          isConnecting = false
          setWsConnected(true)
        }

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            if (data.type === 'new_alert') {
              useAlertStore.getState().addAlert(data.alert)
            } else if (data.type === 'update_alert') {
              useAlertStore.getState().updateAlert(data.alert)
            }
          } catch {
            // Ignore parse errors
          }
        }

        ws.onclose = () => {
          isConnecting = false
          setWsConnected(false)
          reconnectTimer = setTimeout(connectWS, 3000)
        }

        ws.onerror = () => {
          isConnecting = false
        }
      } catch {
        isConnecting = false
        reconnectTimer = setTimeout(connectWS, 3000)
      }
    }

    connectWS()

    const interval = setInterval(() => {
      fetchActiveAlerts()
      fetchStats()
    }, 10000)

    return () => {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
      if (wsRef.current) {
        wsRef.current.onclose = null
        wsRef.current.close()
      }
      clearInterval(interval)
    }
  }, [fetchActiveAlerts, fetchStats, setWsConnected, token])

  const chartData = {
    labels: (stats?.trend || []).map((t) => t.time),
    datasets: [
      {
        label: '告警数量',
        data: (stats?.trend || []).map((t) => t.count),
        fill: true,
        backgroundColor: 'rgba(16, 185, 129, 0.2)',
        borderColor: '#10b981',
        tension: 0.4,
      },
    ],
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: { color: '#f1f5f9' },
      },
      x: {
        grid: { display: false },
      },
    },
  }

  const sortedAlerts = [...activeAlerts].sort((a, b) => {
    if (a.severity === 'P0' && b.severity !== 'P0') return -1
    if (b.severity === 'P0' && a.severity !== 'P0') return 1
    return new Date(b.trigger_time).getTime() - new Date(a.trigger_time).getTime()
  })

  const statsCards = [
    { label: '活跃告警', value: stats?.firing || 0, color: '#ef4444', icon: 'pi pi-bell' },
    { label: 'P0 告警', value: stats?.by_severity?.P0 || 0, color: '#ef4444', icon: 'pi pi-exclamation-triangle' },
    { label: '已确认', value: stats?.acked || 0, color: '#22c55e', icon: 'pi pi-check-circle' },
    { label: '已静默', value: stats?.silenced || 0, color: '#f59e0b', icon: 'pi pi-volume-off' },
  ]

  return (
    <div className="flex flex-column gap-4">
      {!wsConnected && (
        <Message
          severity="warn"
          text="实时连接已断开，正在尝试重新连接..."
        />
      )}

      {/* Stats cards */}
      <div className="grid grid-cols-12 gap-4">
        {statsCards.map((stat, index) => (
          <div key={index} className="col-span-12 md:col-span-3">
            <StatisticCard {...stat} />
          </div>
        ))}
      </div>

      {/* Chart */}
      <Card title="24 小时告警趋势" className="shadow-sm border-0">
        <div style={{ height: '250px' }}>
          <Chart type="line" data={chartData} options={chartOptions} />
        </div>
      </Card>

      {/* Active alerts */}
      <Card
        title={`活跃告警 (${activeAlerts.length})`}
        className="shadow-sm border-0"
        extra={<Button label="查看全部" link onClick={() => window.location.href = '/alerts'} />}
      >
        {loading ? (
          <div className="flex justify-content-center p-4">
            <ProgressSpinner />
          </div>
        ) : sortedAlerts.length === 0 ? (
          <Message severity="success" text="暂无活跃告警，系统运行正常" />
        ) : (
          <div style={{ maxHeight: '600px', overflowY: 'auto' }}>
            {sortedAlerts.map((alert) => (
              <AlertCard
                key={alert.alert_id}
                alert={alert}
                showActions={true}
                onAck={handleAck}
                onQuickSilence={handleQuickSilence}
              />
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: 验证 Dashboard 页面**

Run: `cd frontend && pnpm dev`
Expected: Dashboard 页面显示统计卡片、图表和告警列表

- [ ] **Step 3: 提交**

```bash
git add frontend/src/pages/Dashboard.tsx
git commit -m "feat: migrate Dashboard page to PrimeReact"
```

---

### Task 11: 迁移公共组件 (AlertCard, SeverityBadge, PermissionNotice)

**Files:**
- Modify: `frontend/src/components/AlertCard.tsx`
- Modify: `frontend/src/components/SeverityBadge.tsx`
- Modify: `frontend/src/components/PermissionNotice.tsx`

- [ ] **Step 1: 重写 SeverityBadge 组件**

修改 `frontend/src/components/SeverityBadge.tsx`:

```typescript
import React from 'react'
import { Tag } from 'primereact/tag'

interface SeverityBadgeProps {
  severity: string
}

const severityConfig: Record<string, { severity: 'danger' | 'warning' | 'success' | 'info' | 'secondary' | 'contrast', label: string }> = {
  P0: { severity: 'danger', label: '🔴 P0 紧急' },
  P1: { severity: 'warning', label: '🟠 P1 严重' },
  P2: { severity: 'info', label: '🟡 P2 警告' },
  P3: { severity: 'success', label: '🟢 P3 提示' },
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  const config = severityConfig[severity] || { severity: 'secondary', label: severity }
  return <Tag value={config.label} severity={config.severity} />
}
```

- [ ] **Step 2: 重写 PermissionNotice 组件**

修改 `frontend/src/components/PermissionNotice.tsx`:

```typescript
import React from 'react'
import { Message } from 'primereact/message'

type PermissionNoticeProps = {
  title?: string
  description?: string
  type?: 'info' | 'warn' | 'error' | 'success'
}

export const PermissionNotice: React.FC<PermissionNoticeProps> = ({
  title = '当前角色无权执行该操作',
  description = '你可以继续查看当前页面内容，但无法执行受限操作。',
  type = 'warn',
}) => {
  const severityMap: Record<string, 'info' | 'warn' | 'error' | 'success'> = {
    info: 'info',
    warning: 'warn',
    error: 'error',
    success: 'success',
  }

  return (
    <Message
      severity={severityMap[type] || 'warn'}
      text={title}
      style={{ marginBottom: '16px' }}
    />
  )
}
```

- [ ] **Step 3: 重写 AlertCard 组件**

修改 `frontend/src/components/AlertCard.tsx`:

```typescript
import React from 'react'
import { Tag } from 'primereact/tag'
import { Button } from 'primereact/button'
import { Tooltip } from 'primereact/tooltip'
import { SeverityBadge } from './SeverityBadge'
import dayjs from 'dayjs'
import type { Alert } from '../types'

interface AlertCardProps {
  alert: Alert
  onAck?: (alert: Alert) => void
  onQuickSilence?: (alert: Alert) => void
  showActions?: boolean
}

export const AlertCard: React.FC<AlertCardProps> = ({
  alert,
  onAck,
  onQuickSilence,
  showActions = true,
}) => {
  const isP0 = alert.severity === 'P0'
  const isActive = alert.status === 'firing'

  return (
    <div
      className="mb-3 p-4"
      style={{
        background: '#fff',
        borderRadius: '8px',
        border: isP0 && isActive ? '2px solid #ef4444' : '1px solid #e2e8f0',
        boxShadow: isP0 && isActive ? '0 0 8px rgba(239, 68, 68, 0.3)' : 'none',
      }}
    >
      <div className="flex justify-content-between align-items-start">
        <div className="flex flex-column gap-2 flex-1">
          <div className="flex align-items-center gap-2 flex-wrap">
            <SeverityBadge severity={alert.severity} />
            <span className="font-semibold text-slate-700">{alert.alert_name}</span>
            <Tag value={alert.source} />
            {alert.trigger_count > 1 && (
              <Tag value={`x${alert.trigger_count}`} severity="warning" />
            )}
          </div>
          <p className="text-slate-500 m-0">{alert.message}</p>
          <div className="flex gap-4 text-sm text-slate-400">
            <span>触发时间: {dayjs(alert.trigger_time).format('YYYY-MM-DD HH:mm:ss')}</span>
            {alert.labels && Object.keys(alert.labels).length > 0 && (
              <div className="flex gap-1">
                {Object.entries(alert.labels).slice(0, 3).map(([key, value]) => (
                  <Tag key={key} value={`${key}: ${String(value)}`} className="text-xs" />
                ))}
                {Object.keys(alert.labels).length > 3 && (
                  <span className="text-slate-400">+{Object.keys(alert.labels).length - 3}</span>
                )}
              </div>
            )}
          </div>
        </div>

        {showActions && isActive && (
          <div className="flex gap-2">
            <Button
              icon="pi pi-check"
              label="确认"
              size="small"
              onClick={() => onAck?.(alert)}
            />
            <Button
              icon="pi pi-volume-off"
              label="静默"
              size="small"
              severity="warning"
              onClick={() => onQuickSilence?.(alert)}
            />
          </div>
        )}
      </div>

      {alert.acked_by && (
        <div className="mt-2 text-sm text-slate-400">
          已由 {alert.acked_by} 于 {dayjs(alert.acked_at).format('MM-DD HH:mm')} 确认
          {alert.ack_comment && `: ${alert.ack_comment}`}
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/components/AlertCard.tsx frontend/src/components/SeverityBadge.tsx frontend/src/components/PermissionNotice.tsx
git commit -m "feat: migrate common components to PrimeReact"
```

---

### Task 12: 迁移 Alerts 页面

**Files:**
- Modify: `frontend/src/pages/Alerts.tsx`

- [ ] **Step 1: 重写 Alerts 页面使用 PrimeReact**

由于文件较长，完整代码见设计文档中的组件映射。主要变更：
- Table → DataTable
- Modal → Dialog
- Form → 使用受控组件
- Select → Dropdown
- DatePicker → Calendar
- message → useToast

- [ ] **Step 2: 提交**

```bash
git add frontend/src/pages/Alerts.tsx
git commit -m "feat: migrate Alerts page to PrimeReact"
```

---

## 阶段 3-4: 其他页面迁移

由于篇幅限制，以下页面迁移遵循相同模式：

### Task 13-22: 迁移其他页面

每个页面迁移任务包含：
1. 重写页面使用 PrimeReact 组件
2. 替换 message 为 useToast
3. 替换 Modal 为 Dialog
4. 替换 Table 为 DataTable
5. 替换 Form 为受控组件
6. 验证功能正常
7. 提交

**页面迁移顺序：**
- Task 13: DataSources.tsx
- Task 14: Channels.tsx
- Task 15: RouteRules.tsx
- Task 16: Silences.tsx
- Task 17: OnDuty.tsx
- Task 18: Deliveries.tsx
- Task 19: OpsHealth.tsx
- Task 20: Users.tsx
- Task 21: Profile.tsx

---

## 阶段 5: 收尾

### Task 22: 清理和移除旧依赖

**Files:**
- Modify: `frontend/package.json`
- Delete: `frontend/src/FreyaDemo.tsx`
- Delete: `frontend/src/main-demo.tsx`
- Delete: `frontend/index-demo.html`
- Delete: `frontend/vite.demo.config.ts`

- [ ] **Step 1: 移除 Ant Design 依赖**

Run: `cd frontend && pnpm remove antd @ant-design/icons`

- [ ] **Step 2: 删除 Demo 文件**

```bash
rm frontend/src/FreyaDemo.tsx
rm frontend/src/main-demo.tsx
rm frontend/index-demo.html
rm frontend/vite.demo.config.ts
```

- [ ] **Step 3: 清理未使用的导入**

检查所有文件，移除 antd 相关导入

- [ ] **Step 4: 验证构建**

Run: `cd frontend && pnpm build`
Expected: 构建成功，无错误

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "chore: remove Ant Design dependencies and demo files"
```

---

### Task 23: 更新测试文件

**Files:**
- Modify: `frontend/src/App.test.tsx`
- Modify: `frontend/src/pages/Dashboard.test.tsx`
- Modify: `frontend/src/pages/Alerts.test.tsx`
- Modify: `frontend/src/pages/DataSources.test.tsx`
- Modify: `frontend/src/pages/Deliveries.test.tsx`

- [ ] **Step 1: 更新测试文件以使用 PrimeReact**

在测试文件中添加 PrimeReact Provider 包装

- [ ] **Step 2: 运行测试验证**

Run: `cd frontend && pnpm test`
Expected: 所有测试通过

- [ ] **Step 3: 提交**

```bash
git add frontend/src/**/*.test.tsx
git commit -m "test: update tests for PrimeReact migration"
```

---

## 验收清单

- [ ] 所有页面正常渲染
- [ ] 所有表单功能正常（创建、编辑、验证）
- [ ] 所有表格功能正常（排序、筛选、分页）
- [ ] 所有弹窗功能正常（Modal、Drawer）
- [ ] 全局提示正常（Toast）
- [ ] 侧边栏折叠功能正常
- [ ] 路由导航正常
- [ ] 无 Ant Design 相关依赖残留
- [ ] 视觉风格符合 Freya 浅色主题
- [ ] 所有测试通过
