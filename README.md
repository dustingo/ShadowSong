# 游戏运维告警系统

面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多个数据源的告警信息。当前版本已经完成 AI 能力移除，并支持通知模板原始事件透传与数据源模板预览。

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

正常启动后端只需要数据库、Redis、服务端口和 JWT 相关配置。仓库根目录已提供本地开发用的 `.env` 基线，可按实际环境调整以下配置：

项目常规启动不需要 `OPENAI_*` 或 `AI_*` 变量。

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
│   └── server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── models/          # 数据模型
│   ├── handlers/        # HTTP 处理器
│   ├── router/          # 路由配置
│   ├── processor/       # 告警处理器
│   └── notifier/        # 通知推送
├── frontend/            # 前端项目
└── docker-compose.yml   # Docker 配置
```

## 模板链路

数据源模板分两段执行：

1. `input_template`
   作用：把原始 webhook JSON 标准化成内部 `Alert` JSON。
   要求：模板输出必须是合法 JSON，并至少包含 `alert_id`、`alert_name`、`severity`、`message`、`source`、`status`、`trigger_time` 这些核心字段中的必需项。

2. `output_template`
   作用：把标准化后的告警渲染成最终通知内容。
   可用字段：
   - 标准字段：`.alert_name`、`.message`、`.status`、`.source`、`.labels`
   - 严重级别：`.severity` / `.severity_code`
   - 原始严重级别：`.severity_raw`
   - 原始事件：`.event.xxx`
   - 可读别名：`.alert.severity_code`、`.alert.severity_raw`

严重级别标准化映射：

- `critical` -> `P0`
- `warning` / `error` -> `P1`
- `info` -> `P2`
- `debug` -> `P3`

示例 `input_template`：

```gotemplate
{
  "alert_id": "{{ .alertId }}",
  "alert_name": "{{ .alertName }}",
  "severity": "{{ .severity }}",
  "message": "{{ .summary }}",
  "source": "custom-source",
  "status": "{{ .status }}",
  "trigger_time": "{{ .startsAt }}"
}
```

示例 `output_template`：

```gotemplate
{{ if eq .severity_raw "critical" }}
严重告警
{{ else if eq .severity_raw "warning" }}
⚠️ 一般告警
{{ else }}
ℹ️ 提示信息
{{ end }}

名称: {{ .alert_name }}
实例: {{ .event.labels.instance }}
描述: {{ default .event.description .message }}
```

## 模板预览与验证

- 前端数据源页面支持模板预览，会调用 `POST /api/v1/datasources/preview`
- 后端真实通知透传验证脚本：`pwsh -ExecutionPolicy Bypass -File scripts/verify_template_passthrough.ps1`
- 后端无 AI 闭环验证脚本：`pwsh -ExecutionPolicy Bypass -File scripts/verify_backend_no_ai.ps1`
- 前端无 AI 构建/残留扫描脚本：`pwsh -ExecutionPolicy Bypass -File scripts/verify_frontend_no_ai.ps1`

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

启动服务后可先访问 `http://localhost:8080/health` 验证服务是否正常启动。

当前接口以 Gin 路由和 `internal/handlers/` 中的实现为准；常用接口包括认证、告警列表/统计、配置管理、Webhook 接入、数据源模板预览和 WebSocket 推送。

## 许可证

MIT
