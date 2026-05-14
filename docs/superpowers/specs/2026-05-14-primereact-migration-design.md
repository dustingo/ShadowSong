---
name: primereact-migration
description: Migrate frontend from Ant Design to PrimeReact with Freya-style layout
---

# PrimeReact 迁移设计文档

## 概述

将游戏运维告警系统的前端从 Ant Design 迁移到 PrimeReact，采用 Freya 模板的浅色主题风格，包含可折叠侧边栏布局。

## 目标

- 替换所有 Ant Design 组件为 PrimeReact 组件
- 实现 Freya 风格的浅色主题布局
- 保持现有功能不变
- 提升视觉美观度

## 技术栈变更

| 项目 | 当前 | 迁移后 |
|------|------|--------|
| UI 组件库 | Ant Design 5.x | PrimeReact 10.x |
| 图标库 | @ant-design/icons | PrimeIcons |
| CSS 工具 | 内联样式 | PrimeFlex + 内联样式 |
| 图表库 | ECharts | ECharts（保持不变） |
| 代码编辑器 | Monaco Editor | Monaco Editor（保持不变） |
| 状态管理 | Zustand | Zustand（保持不变） |
| 路由 | React Router | React Router（保持不变） |

## 依赖变更

### 新增依赖
- `primereact` - PrimeReact 核心组件库
- `primeicons` - PrimeReact 图标库
- `primeflex` - PrimeReact CSS 工具类
- `chart.js` - PrimeReact Chart 组件依赖

### 移除依赖
- `antd`
- `@ant-design/icons`

## 视觉风格

### 主题配置
- **主题**: Lara Light Indigo
- **侧边栏**: 白色背景 `#ffffff`，右边框 `#e2e8f0`
- **主色调**: 翠绿色 `#10b981`
- **内容区背景**: 浅灰 `#f8fafc`
- **卡片背景**: 白色 `#ffffff`

### 布局结构

```
┌─────────────────────────────────────────────────────────────┐
│  Header (60px)                                               │
│  页面标题 │ 搜索框 │ 通知图标 │ 用户头像                       │
└─────────────────────────────────────────────────────────────┘
┌──────────┬──────────────────────────────────────────────────┐
│ Sidebar  │  Content                                          │
│ (220px)  │  padding: 16px                                    │
│ 可折叠   │  统计卡片 / 图表 / 表格等内容                        │
│ 白色背景 │                                                    │
└──────────┴──────────────────────────────────────────────────┘
```

### 侧边栏特性
- 展开宽度: 220px
- 折叠宽度: 60px
- 选中项: 浅绿背景 `bg-emerald-50`，右边框 `#10b981`
- 未选中项: 悬停时浅灰背景
- 底部折叠按钮

## 页面迁移清单

### 核心布局 (1个文件)
| 文件 | 改动 |
|------|------|
| App.tsx | 重写 Layout、Sider、Header、Menu 组件 |

### 公共组件 (4个文件)
| 文件 | 改动 |
|------|------|
| AlertCard.tsx | 替换 Tag、Button、Typography |
| SeverityBadge.tsx | 替换 Tag |
| PermissionNotice.tsx | 替换 Alert |
| CodeEditor.tsx | 保持不变（Monaco Editor） |

### 页面组件 (12个文件)
| 文件 | 主要组件替换 |
|------|-------------|
| Login.tsx | Form → Form, Input → InputText, Button → Button |
| Dashboard.tsx | Card, Statistic, Table, Alert → Card, 自定义统计, DataTable, Message |
| Alerts.tsx | Table, Modal, Form, Select, DatePicker → DataTable, Dialog, Form, Dropdown, Calendar |
| Deliveries.tsx | Table, Drawer, Modal, Descriptions → DataTable, Sidebar, Dialog, 自定义描述列表 |
| OpsHealth.tsx | Card, Statistic, Table → Card, 自定义统计, DataTable |
| DataSources.tsx | Table, Modal, Form, Drawer → DataTable, Dialog, Form, Sidebar |
| Channels.tsx | Table, Modal, Form, Select → DataTable, Dialog, Form, Dropdown |
| RouteRules.tsx | Table, Modal, Form → DataTable, Dialog, Form |
| Silences.tsx | Table, Modal, Form, Tabs → DataTable, Dialog, Form, TabView |
| OnDuty.tsx | Table, Modal, Form → DataTable, Dialog, Form |
| Users.tsx | Table, Modal, Form, Popconfirm → DataTable, Dialog, Form, ConfirmDialog |
| Profile.tsx | Card, Form → Card, Form |

## 组件映射表

| Ant Design | PrimeReact | 备注 |
|------------|------------|------|
| Layout/Sider/Header/Content | 自定义 div 布局 | 使用 PrimeFlex 工具类 |
| Menu | 自定义菜单 | 简单的 div + onClick |
| Table | DataTable | 功能更强大 |
| Card | Card | API 略有不同 |
| Form | Forms | 需要重写验证逻辑 |
| Modal | Dialog | 属性名不同 |
| Drawer | Sidebar | 作为右侧抽屉使用 |
| message | Toast | 全局提示改用 useToast |
| Tag | Tag | API 相似 |
| Button | Button | 图标使用 icon 属性 |
| Input | InputText | 属性名不同 |
| Input.Password | Password | 独立组件 |
| Input.TextArea | InputTextarea | 属性名不同 |
| Select | Dropdown / SelectButton | 根据场景选择 |
| DatePicker | Calendar | 日期选择器 |
| RangePicker | Calendar (selectionMode="range") | 范围选择 |
| Switch | InputSwitch | API 相似 |
| Tabs | TabView | 结构不同 |
| Alert | Message | 用于提示信息 |
| Spin | ProgressSpinner | 加载指示器 |
| Badge | Badge | API 相似 |
| Avatar | Avatar | API 相似 |
| Popconfirm | ConfirmDialog | 确认弹窗 |
| Typography.Text | span + className | 使用普通 HTML |
| Typography.Title | h1-h6 + className | 使用普通 HTML |
| Space | div + flex | 使用 PrimeFlex |
| Row/Col | div + grid | 使用 PrimeFlex grid |
| Statistic | 自定义组件 | 简单数值展示 |
| Descriptions | 自定义组件 | 键值对展示 |
| Divider | Divider | API 相似 |

## 全局配置

### main.tsx
```tsx
import { PrimeReactProvider } from 'primereact/api'
import 'primereact/resources/themes/lara-light-indigo/theme.css'
import 'primereact/resources/primereact.min.css'
import 'primeicons/primeicons.css'
import 'primeflex/primeflex.css'

// 中文 locale 配置
import { locale, addLocale } from 'primereact/api'
addLocale('zh-CN', {
  // ... 中文翻译
})
locale('zh-CN')
```

### Toast 全局配置
```tsx
// 在 App.tsx 中使用 ToastService
import { Toast } from 'primereact/toast'
// 使用 useRef 获取 toast 实例，通过 context 传递
```

## 迁移步骤

### 阶段 1: 基础设施
1. 安装 PrimeReact 依赖
2. 配置全局样式和主题
3. 创建布局组件（Sidebar、Header）
4. 创建全局 Toast 配置

### 阶段 2: 核心页面
1. Login.tsx - 登录页
2. App.tsx - 主布局
3. Dashboard.tsx - 告警大盘
4. Alerts.tsx - 告警管理

### 阶段 3: 配置页面
1. DataSources.tsx
2. Channels.tsx
3. RouteRules.tsx
4. Silences.tsx
5. OnDuty.tsx

### 阶段 4: 其他页面
1. Deliveries.tsx
2. OpsHealth.tsx
3. Users.tsx
4. Profile.tsx

### 阶段 5: 收尾
1. 移除 Ant Design 依赖
2. 清理未使用的导入
3. 测试所有页面功能
4. 更新文档

## 风险与注意事项

1. **Form 表单验证**: PrimeReact Forms 的验证逻辑与 Ant Design 不同，需要重写
2. **DatePicker**: PrimeReact 的 Calendar 组件 API 差异较大
3. **message 全局提示**: 需要使用 Toast 组件，通过 ref 调用
4. **样式覆盖**: 部分组件可能需要自定义 CSS 覆盖默认样式
5. **图标**: @ant-design/icons 需要替换为 PrimeIcons，部分图标名称不同

## 验收标准

- [ ] 所有页面正常渲染
- [ ] 所有表单功能正常（创建、编辑、验证）
- [ ] 所有表格功能正常（排序、筛选、分页）
- [ ] 所有弹窗功能正常（Modal、Drawer）
- [ ] 全局提示正常（Toast）
- [ ] 侧边栏折叠功能正常
- [ ] 路由导航正常
- [ ] 无 Ant Design 相关依赖残留
- [ ] 视觉风格符合 Freya 浅色主题
