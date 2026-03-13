# 游戏运维 AI 告警系统

智能化的告警管理平台，用于统一接收、处理、聚合和分发来自多个数据源的告警信息。

## 技术栈

### 后端
- Go 1.21+
- Gin Web Framework
- PostgreSQL 14+
- Redis 7+
- GORM

### 前端
- React 18 + TypeScript
- Vite
- pnpm
- Ant Design 5.x
- Zustand
- ECharts

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- pnpm
- Docker & Docker Compose

### 安装依赖

```bash
# 后端依赖
make deps

# 前端依赖
make frontend-install
```

### 启动开发环境

1. 启动 PostgreSQL 和 Redis：
```bash
make docker-up
```

2. 配置环境变量：
```bash
cp .env.example .env
# 编辑 .env 文件，配置必要的参数
```

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
│   └── server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── models/          # 数据模型
│   ├── handlers/        # HTTP 处理器
│   ├── router/          # 路由配置
│   ├── processor/       # 告警处理器
│   ├── ai/              # AI 分析器
│   └── notifier/        # 通知推送
├── frontend/            # 前端项目
└── docker-compose.yml   # Docker 配置
```

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

## API 文档

启动服务后访问 http://localhost:8080/api/v1/ping 测试 API 连接。

详细 API 文档请参考 `.kiro/specs/ai-alert-system/design.md`。

## 许可证

MIT
