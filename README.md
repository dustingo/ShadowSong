# 游戏运维告警系统

面向游戏运维场景的告警管理平台，统一接收、处理、聚合、展示和分发来自多个数据源的告警信息。支持多通道通知投递、路由规则匹配、静默管理、值班调度，以及完整的 RBAC 权限控制。

## 技术栈

### 后端
- Go 1.25.0 + Gin
- PostgreSQL 14+ (GORM)
- Redis 7+
- JWT 认证 + RBAC 授权

### 前端
- React 18 + TypeScript
- PrimeReact 10.x + PrimeFlex
- Zustand 状态管理
- ECharts / Chart.js 图表
- Monaco Editor 模板编辑
- Vite + pnpm

## 功能概览

| 模块 | 说明 |
|------|------|
| 告警管理 | 告警列表、按级别/来源/状态/时间/Labels 筛选、确认、静默 |
| 实时推送 | WebSocket 实时告警推送 |
| 数据源 | 多数据源接入、input/output 模板、模板预览 |
| 通知通道 | 飞书/钉钉/企微/Webhook 通道配置与健康度监控 |
| 路由规则 | 基于级别/来源/Labels 的路由匹配，优先级排序 |
| 静默规则 | 时间窗口静默，支持从告警快速创建 |
| 值班管理 | 值班排班与通道绑定 |
| 投递追踪 | 通知投递记录、重试/重放、恢复审计 |
| 用户管理 | 用户 CRUD、角色(admin/operator/viewer)、审计日志 |
| 运维监控 | 告警统计趋势、通道健康度、系统指标 |

## 快速开始

### 前置要求

- Go 1.25.0
- Node.js 18+ & pnpm
- Docker & Docker Compose

### 安装依赖

```bash
make deps          # 后端依赖
make frontend-install  # 前端依赖
```

### 启动开发环境

1. 启动 PostgreSQL 和 Redis：
```bash
make docker-up
```

2. 配置环境变量（仓库已提供 `.env` 基线，按需调整）：

- `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` / `DB_SSLMODE`
- `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB`
- `SERVER_PORT` / `SERVER_MODE`
- `JWT_SECRET` / `TOKEN_EXPIRY`

3. 启动后端服务：
```bash
make run
```

4. 启动前端开发服务器（新终端）：
```bash
make frontend-dev
```

### 访问应用

- 后端 API: http://localhost:8080
- 前端界面: http://localhost:5173
- 健康检查: http://localhost:8080/health

## 项目结构

```
.
├── cmd/
│   ├── server/           # 应用入口
│   └── roleaudit/        # RBAC 审计工具
├── internal/
│   ├── auth/             # JWT 认证
│   ├── authz/            # RBAC 授权 (5 capabilities, 3 roles)
│   ├── config/           # 配置管理
│   ├── database/         # PostgreSQL + Redis 连接
│   ├── delivery/         # 通知投递服务
│   ├── handlers/         # HTTP 处理器
│   ├── middleware/        # 认证/授权/限流中间件
│   ├── models/           # GORM 数据模型
│   ├── notifier/         # 通知推送
│   ├── router/           # 路由注册
│   ├── routing/          # 路由规则匹配引擎
│   ├── stats/            # 统计查询
│   ├── template/         # Go 模板渲染
│   ├── utils/            # 工具函数
│   └── websocket/        # WebSocket 实时推送
├── frontend/             # 前端项目
│   └── src/
│       ├── api/          # Axios HTTP 客户端
│       ├── authz/        # 前端权限检查
│       ├── components/   # 通用组件
│       ├── pages/        # 页面组件
│       ├── stores/       # Zustand 状态管理
│       └── theme/        # 主题配置
├── scripts/              # 验证脚本
├── docs/                 # 文档
└── docker-compose.yml    # Docker 配置
```

## API 路由

所有接口位于 `/api/v1` 下，需 JWT 认证（公开接口除外）：

| 分组 | 路由 | 说明 |
|------|------|------|
| Auth | `POST /auth/login` `POST /auth/logout` `POST /auth/refresh` | 认证（公开） |
| Alerts | `GET /alerts` `GET /alerts/stats` `GET /alerts/active` `POST /alerts/:id/ack` `POST /alerts/:id/quick-silence` | 告警管理 |
| DataSources | `GET/POST /datasources` `POST /datasources/preview` `PATCH /datasources/:id/toggle` | 数据源 |
| Channels | `GET/POST /channels` `POST /channels/:id/test` `GET /channels/:id/health` | 通知通道 |
| Routes | `GET/POST /routes` `POST /routes/reorder` | 路由规则 |
| Silences | `GET/POST /silences` `POST /silences/from-alert/:alertId` | 静默规则 |
| OnDuty | `GET/POST /onduty` `GET /onduty/current` | 值班管理 |
| Deliveries | `GET /deliveries` `POST /deliveries/:id/retry` `POST /deliveries/:id/replay` | 投递追踪 |
| Users | `GET/POST /users` `PATCH /users/me/profile` `GET /users/audit-logs` | 用户管理 |
| Webhook | `POST /webhook/:source_name` | 告警接入（公开，限流） |
| WebSocket | `GET /ws/alerts` | 实时推送 |
| Metrics | `GET /metrics` | 系统指标 |

## 模板链路

数据源模板分两段执行：

1. **input_template** — 接收 webhook JSON，映射为内部告警字段
2. **output_template** — 将标准化告警渲染为通知内容

可用字段：`.alert_name` `.message` `.status` `.source` `.labels` `.severity` `.severity_raw` `.event.xxx`

严重级别映射：`critical` → P0, `warning`/`error` → P1, `info` → P2, `debug` → P3

## RBAC 权限

| 角色 | view_alerts | process_alerts | view_config | manage_config | manage_users |
|------|:-----------:|:--------------:|:-----------:|:-------------:|:------------:|
| admin | ✓ | ✓ | ✓ | ✓ | ✓ |
| operator | ✓ | ✓ | ✓ | — | — |
| viewer | ✓ | — | ✓ | — | — |

## 开发命令

```bash
make help              # 显示所有可用命令
make build             # 编译应用
make run               # 运行应用
make test              # 运行测试
make docker-up         # 启动 Docker 服务
make docker-down       # 停止 Docker 服务
make frontend-dev      # 启动前端开发服务器
make frontend-build    # 构建前端生产版本
```

## 工程质量门禁

- GitHub Actions 工作流：`.github/workflows/quality-gates.yml`
- 后端：`go test ./...`
- 前端：`pnpm lint`、`pnpm test -- --run`、`pnpm build`

## 许可证

MIT
