# 设计文档

## 概述

游戏运维 AI 告警系统是一个基于 Go 语言开发的智能告警管理平台。系统采用微服务架构，使用 Redis 作为消息队列和缓存，PostgreSQL 作为持久化存储，前端采用现代化的 React + TypeScript 技术栈。

系统核心功能包括：
- 多数据源告警接入与标准化
- 智能去重与聚合
- AI 驱动的根因分析
- 灵活的路由与多渠道推送
- 实时监控大盘

## 技术栈

### 后端
- **语言**: Go 1.21+
- **Web 框架**: Gin
- **数据库**: PostgreSQL 14+
- **缓存/队列**: Redis 7+
- **ORM**: GORM
- **模板引擎**: Go Template
- **AI 集成**: OpenAI API / 自定义 LLM

### 前端
- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **包管理器**: pnpm
- **UI 库**: Ant Design 5.x（浅色主题）
- **状态管理**: Zustand
- **HTTP 客户端**: Axios
- **WebSocket**: native WebSocket API
- **图表**: ECharts
- **代码编辑器**: Monaco Editor（用于模板编辑）

## 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        外部数据源                              │
│  (Prometheus, Zabbix, CloudWatch, Custom...)                │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTP POST
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                     API Gateway (Gin)                        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Webhook Handler (/webhook/{source_name})           │   │
│  └──────────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────────┘
                     │ Write to Stream
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Redis Stream (alerts:normalized)           │
└────────────────────┬────────────────────────────────────────┘
                     │ Consumer Group
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  Alert Processor (Go Worker)                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Deduplicator │→ │ Silence      │→ │ Aggregator   │      │
│  │              │  │ Matcher      │  │ (Window)     │      │
│  └──────────────┘  └──────────────┘  └──────┬───────┘      │
└────────────────────────────────────────────┬────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      AI Analyzer                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  - Root Cause Analysis                               │   │
│  │  - Severity Assessment                               │   │
│  │  - Suggestion Generation                             │   │
│  └──────────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Notification Router                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Feishu       │  │ DingTalk     │  │ WeCom        │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      PostgreSQL                              │
│  - alerts (告警表)                                            │
│  - data_sources (数据源配置)                                  │
│  - silence_rules (静默规则)                                   │
│  - route_rules (路由规则)                                     │
│  - channels (推送渠道)                                        │
│  - ai_logs (AI 处理日志)                                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Frontend (React)                          │
│  - Dashboard (实时大盘)                                       │
│  - Alert Management (告警管理)                               │
│  - Configuration (配置管理)                                   │
│  - AI Assistant (AI 助手)                                    │
└─────────────────────────────────────────────────────────────┘
```



## 组件与接口

### 后端服务组件

#### 1. API Server (Gin)

**职责**: 
- 接收外部 webhook 请求
- 提供 RESTful API 给前端
- WebSocket 连接管理

**主要模块**:
```go
// handlers/webhook.go
type WebhookHandler struct {
    dataSourceRepo DataSourceRepository
    redisClient    *redis.Client
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
    sourceName := c.Param("source_name")
    // 1. 验证数据源是否启用
    // 2. 读取原始数据
    // 3. 使用 input_template 渲染
    // 4. 生成统一告警结构
    // 5. 写入 Redis Stream
}

// handlers/api.go
type APIHandler struct {
    // CRUD operations for configuration
}

// handlers/websocket.go
type WebSocketHandler struct {
    // Real-time alert push
}
```

#### 2. Alert Processor (Worker)

**职责**:
- 从 Redis Stream 消费告警
- 执行去重、静默、聚合逻辑

**主要模块**:
```go
// processor/consumer.go
type AlertConsumer struct {
    redisClient *redis.Client
    db          *gorm.DB
}

func (c *AlertConsumer) Start() {
    // Consumer Group: alert-processor
    // Stream: alerts:normalized
}

// processor/deduplicator.go
type Deduplicator struct {
    // 计算指纹
    // 检查去重 TTL
}

// processor/silence_matcher.go
type SilenceMatcher struct {
    // 匹配静默规则
}

// processor/aggregator.go
type Aggregator struct {
    // Session Window 聚合
    // 触发条件检查
}
```

#### 3. AI Analyzer

**职责**:
- 调用 AI 模型分析告警
- 生成摘要、根因、建议

**主要模块**:
```go
// ai/analyzer.go
type AIAnalyzer struct {
    client AIClient
    db     *gorm.DB
}

type AnalysisResult struct {
    Summary      string
    RootCause    string
    Severity     string
    Suggestions  []string
    IsAutoHeal   bool
    Tags         []string
}

func (a *AIAnalyzer) Analyze(ctx context.Context, group *AggregatedAlert) (*AnalysisResult, error) {
    // 1. 构建 AI 上下文（当前告警 + 历史记录）
    // 2. 调用 AI API
    // 3. 解析结果
    // 4. 验证 severity
    // 5. 记录日志
}
```

#### 4. Notification Router

**职责**:
- 路由规则匹配
- 多渠道推送

**主要模块**:
```go
// notifier/router.go
type NotificationRouter struct {
    ruleRepo    RouteRuleRepository
    channelRepo ChannelRepository
    onDutyRepo  OnDutyRepository
}

func (r *NotificationRouter) Route(alert *ProcessedAlert) []Channel {
    // 1. 匹配路由规则
    // 2. 获取值班人员渠道
    // 3. 合并渠道列表
}

// notifier/channels/feishu.go
type FeishuChannel struct {
    webhookURL string
    secret     string
}

func (f *FeishuChannel) Send(alert *ProcessedAlert, template string) error {
    // 使用 output_template 渲染消息
    // 发送 Interactive Card
}

// notifier/channels/dingtalk.go
// notifier/channels/wecom.go
// notifier/channels/webhook.go
```



### 数据模型

#### 核心数据结构

```go
// models/alert.go
type Alert struct {
    AlertID      string            `gorm:"primaryKey" json:"alert_id"`
    Source       string            `gorm:"index" json:"source"`
    AlertName    string            `gorm:"index" json:"alert_name"`
    Severity     string            `gorm:"index" json:"severity"` // P0/P1/P2/P3
    Message      string            `json:"message"`
    Labels       datatypes.JSON    `json:"labels"` // map[string]string
    Fingerprint  string            `gorm:"uniqueIndex:idx_fingerprint_time" json:"fingerprint"`
    TriggerTime  time.Time         `gorm:"index" json:"trigger_time"`
    ReceivedAt   time.Time         `gorm:"index" json:"received_at"`
    Status       string            `gorm:"index" json:"status"` // pending/firing/acked/silenced/resolved/deduplicated
    Raw          datatypes.JSON    `json:"raw"`
    
    // AI 分析结果
    AISummary    string            `json:"ai_summary"`
    AIRootCause  string            `json:"ai_root_cause"`
    AISeverity   string            `json:"ai_severity"`
    AISuggestions datatypes.JSON   `json:"ai_suggestions"` // []string
    AITags       datatypes.JSON    `json:"ai_tags"` // []string
    
    // 确认信息
    AckedBy      string            `json:"acked_by"`
    AckedAt      *time.Time        `json:"acked_at"`
    AckComment   string            `json:"ack_comment"`
    
    // 统计
    TriggerCount int               `json:"trigger_count"` // 去重计数
    
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
}

// models/data_source.go
type DataSource struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    Name            string    `gorm:"uniqueIndex" json:"name"` // 不可修改
    DisplayName     string    `json:"display_name"`
    InputTemplate   string    `gorm:"type:text" json:"input_template"`
    OutputTemplate  string    `gorm:"type:text" json:"output_template"`
    GroupByLabels   datatypes.JSON `json:"group_by_labels"` // []string
    Enabled         bool      `json:"enabled"`
    LastTriggerAt   *time.Time `json:"last_trigger_at"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// models/silence_rule.go
type SilenceRule struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    Name             string         `json:"name"`
    Comment          string         `json:"comment"`
    Source           string         `json:"source"` // 空表示匹配所有
    AlertNamePattern string         `json:"alert_name_pattern"` // 正则
    Severities       datatypes.JSON `json:"severities"` // []string
    LabelMatchers    datatypes.JSON `json:"label_matchers"` // []LabelMatcher
    StartsAt         time.Time      `json:"starts_at"`
    EndsAt           time.Time      `json:"ends_at"`
    CreatedBy        string         `json:"created_by"`
    CreatedAt        time.Time      `json:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at"`
}

type LabelMatcher struct {
    Key     string `json:"key"`
    Pattern string `json:"pattern"` // 正则
}

// models/route_rule.go
type RouteRule struct {
    ID              uint           `gorm:"primaryKey" json:"id"`
    Name            string         `json:"name"`
    Priority        int            `gorm:"index" json:"priority"` // 越小优先级越高
    Severities      datatypes.JSON `json:"severities"` // []string
    Sources         datatypes.JSON `json:"sources"` // []string
    LabelMatchers   datatypes.JSON `json:"label_matchers"` // []LabelMatcher
    ChannelIDs      datatypes.JSON `json:"channel_ids"` // []uint
    TimeRanges      datatypes.JSON `json:"time_ranges"` // []TimeRange
    Enabled         bool           `json:"enabled"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
}

type TimeRange struct {
    StartTime string `json:"start_time"` // HH:MM
    EndTime   string `json:"end_time"`   // HH:MM
}

// models/channel.go
type Channel struct {
    ID         uint           `gorm:"primaryKey" json:"id"`
    Name       string         `json:"name"`
    Type       string         `json:"type"` // feishu/dingtalk/wecom/webhook
    Config     datatypes.JSON `json:"config"` // 根据类型不同存储不同配置
    Enabled    bool           `json:"enabled"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
}

// models/on_duty.go
type OnDuty struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    UserID     string    `json:"user_id"`
    UserName   string    `json:"user_name"`
    ChannelID  uint      `json:"channel_id"`
    StartTime  time.Time `json:"start_time"`
    EndTime    time.Time `json:"end_time"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

// models/ai_log.go
type AILog struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    AlertID     string         `gorm:"index" json:"alert_id"`
    Input       datatypes.JSON `json:"input"`
    Output      datatypes.JSON `json:"output"`
    IsAccurate  *bool          `json:"is_accurate"` // null/true/false
    Feedback    string         `json:"feedback"`
    Duration    int            `json:"duration"` // ms
    CreatedAt   time.Time      `json:"created_at"`
}

// models/dead_letter.go
type DeadLetter struct {
    ID         uint           `gorm:"primaryKey" json:"id"`
    Source     string         `json:"source"`
    RawData    datatypes.JSON `json:"raw_data"`
    Error      string         `json:"error"`
    CreatedAt  time.Time      `json:"created_at"`
}
```



### API 接口设计

#### Webhook 接口

```
POST /webhook/:source_name
Content-Type: application/json

Request Body: 任意 JSON（由数据源决定）

Response:
{
  "success": true,
  "alert_id": "uuid",
  "message": "Alert received"
}

Error Response:
{
  "success": false,
  "error": "Data source not found or disabled"
}
```

#### 数据源管理 API

```
# 列表
GET /api/v1/data-sources
Response: {
  "data": [
    {
      "id": 1,
      "name": "prometheus",
      "display_name": "Prometheus",
      "webhook_url": "https://domain.com/webhook/prometheus",
      "group_by_labels": ["host", "zone"],
      "enabled": true,
      "last_trigger_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 10
}

# 创建
POST /api/v1/data-sources
Request: {
  "name": "prometheus",
  "display_name": "Prometheus",
  "input_template": "...",
  "output_template": "...",
  "group_by_labels": ["host", "zone"]
}

# 更新
PUT /api/v1/data-sources/:id
Request: {
  "display_name": "Prometheus Updated",
  "input_template": "...",
  "output_template": "...",
  "group_by_labels": ["host"]
}

# 删除
DELETE /api/v1/data-sources/:id

# 测试输入模板
POST /api/v1/data-sources/test-input-template
Request: {
  "template": "...",
  "sample_data": {...}
}
Response: {
  "success": true,
  "result": {
    "alert_name": "HighMemory",
    "severity": "P1",
    "message": "Memory usage is high",
    "labels": {...}
  }
}

# 测试输出模板
POST /api/v1/data-sources/test-output-template
Request: {
  "template": "...",
  "alert_data": {...}
}
Response: {
  "success": true,
  "result": "渲染后的消息内容"
}

# 启用/禁用
PATCH /api/v1/data-sources/:id/toggle
Request: {
  "enabled": true
}
```

#### 告警管理 API

```
# 列表（支持筛选）
GET /api/v1/alerts?page=1&size=20&severity=P0,P1&source=prometheus&status=firing&start_time=...&end_time=...&labels=host:server-01

Response: {
  "data": [...],
  "total": 100,
  "page": 1,
  "size": 20
}

# 详情
GET /api/v1/alerts/:alert_id

# 确认告警
POST /api/v1/alerts/:alert_id/ack
Request: {
  "comment": "已处理，重启服务解决"
}

# 通过 token 确认（用于消息按钮）
POST /api/v1/alerts/ack-by-token
Request: {
  "token": "signed-token",
  "comment": "已处理"
}

# 大盘统计
GET /api/v1/alerts/dashboard/stats
Response: {
  "p0_count": 2,
  "p1_count": 5,
  "p2_count": 10,
  "p3_count": 3,
  "firing_count": 20,
  "trend": [
    {"time": "2024-01-01T00:00:00Z", "count": 10},
    ...
  ]
}

# 活跃告警列表（仅 firing 状态）
GET /api/v1/alerts/dashboard/active
```

#### 静默规则 API

```
# 列表
GET /api/v1/silence-rules?active=true

# 创建
POST /api/v1/silence-rules
Request: {
  "name": "维护窗口",
  "comment": "数据库升级",
  "source": "prometheus",
  "alert_name_pattern": ".*Database.*",
  "severities": ["P2", "P3"],
  "label_matchers": [
    {"key": "env", "pattern": "prod"}
  ],
  "starts_at": "2024-01-01T00:00:00Z",
  "ends_at": "2024-01-01T02:00:00Z"
}

# 更新
PUT /api/v1/silence-rules/:id

# 删除（提前取消）
DELETE /api/v1/silence-rules/:id

# 快捷创建（从告警）
POST /api/v1/silence-rules/from-alert
Request: {
  "alert_id": "uuid",
  "duration": "1h" // 1h/4h/today/custom
}
```

#### 路由规则 API

```
# 列表
GET /api/v1/route-rules

# 创建
POST /api/v1/route-rules
Request: {
  "name": "P0告警路由",
  "priority": 1,
  "severities": ["P0"],
  "sources": ["prometheus"],
  "label_matchers": [...],
  "channel_ids": [1, 2],
  "time_ranges": [
    {"start_time": "09:00", "end_time": "18:00"}
  ]
}

# 更新
PUT /api/v1/route-rules/:id

# 删除
DELETE /api/v1/route-rules/:id

# 批量更新优先级（拖拽排序）
POST /api/v1/route-rules/reorder
Request: {
  "orders": [
    {"id": 1, "priority": 1},
    {"id": 2, "priority": 2}
  ]
}
```

#### 推送渠道 API

```
# 列表
GET /api/v1/channels

# 创建
POST /api/v1/channels
Request: {
  "name": "运维飞书群",
  "type": "feishu",
  "config": {
    "webhook_url": "https://...",
    "secret": "..."
  }
}

# 更新
PUT /api/v1/channels/:id

# 删除
DELETE /api/v1/channels/:id

# 测试发送
POST /api/v1/channels/:id/test
```

#### 值班管理 API

```
# 列表
GET /api/v1/on-duty?start_time=...&end_time=...

# 创建
POST /api/v1/on-duty
Request: {
  "user_id": "user123",
  "user_name": "张三",
  "channel_id": 1,
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-02T00:00:00Z"
}

# 更新
PUT /api/v1/on-duty/:id

# 删除
DELETE /api/v1/on-duty/:id

# 当前值班人员
GET /api/v1/on-duty/current
```

#### AI 功能 API

```
# 智能问答
POST /api/v1/ai/chat
Request: {
  "question": "最近一周 P0 告警有多少？",
  "session_id": "uuid" // 用于保持上下文
}
Response: {
  "answer": "最近一周共有 15 条 P0 告警",
  "chart_data": {...}, // 可选
  "data_range": "基于最近 7 天告警数据"
}

# 处置建议查询
POST /api/v1/ai/suggestions
Request: {
  "alert_id": "uuid"
}
Response: {
  "suggestions": [
    "1. 检查内存使用情况",
    "2. 重启相关服务",
    "3. 查看日志文件"
  ],
  "history_count": 5,
  "note": "基于 5 条历史处置记录"
}

# AI 处理日志
GET /api/v1/ai/logs?page=1&size=20

# 标记 AI 准确性
POST /api/v1/ai/logs/:id/feedback
Request: {
  "is_accurate": true,
  "feedback": "分析准确"
}

# AI 准确率统计
GET /api/v1/ai/accuracy?days=7
Response: {
  "total": 100,
  "accurate": 85,
  "inaccurate": 10,
  "unmarked": 5,
  "accuracy_rate": 0.85
}

# 静默规则推荐
GET /api/v1/ai/silence-recommendations
Response: {
  "recommendations": [
    {
      "id": "rec-1",
      "alert_fingerprint": "...",
      "frequency": 8,
      "auto_heal_rate": 0.9,
      "suggested_rule": {
        "name": "自动恢复告警静默",
        "source": "prometheus",
        "alert_name_pattern": "HighMemory",
        "duration": "4h"
      }
    }
  ]
}

# 采纳推荐
POST /api/v1/ai/silence-recommendations/:id/accept
Request: {
  "rule": {...} // 可修改后的规则
}

# 忽略推荐
POST /api/v1/ai/silence-recommendations/:id/ignore
```

#### WebSocket 接口

```
WS /ws/alerts

# 客户端连接后，服务端实时推送新告警
Message Format:
{
  "type": "new_alert",
  "data": {
    "alert_id": "uuid",
    "severity": "P0",
    "alert_name": "HighMemory",
    ...
  }
}

# 心跳
{
  "type": "ping"
}

# 断线重连机制：客户端每 3 秒尝试重连
```



## 前端设计

### 技术架构

```
src/
├── api/                    # API 调用封装
│   ├── alerts.ts
│   ├── dataSources.ts
│   ├── channels.ts
│   ├── routes.ts
│   ├── silence.ts
│   ├── onDuty.ts
│   └── ai.ts
├── components/             # 通用组件
│   ├── AlertCard/         # 告警卡片
│   ├── CodeEditor/        # 模板编辑器（Monaco）
│   ├── SeverityBadge/     # 级别徽章
│   ├── TimeRangePicker/   # 时间范围选择
│   └── ...
├── pages/                  # 页面组件
│   ├── Dashboard/         # 告警大盘
│   ├── AlertManagement/   # 告警管理
│   ├── DataSources/       # 数据源管理
│   ├── Channels/          # 推送渠道
│   ├── Routes/            # 路由规则
│   ├── Silence/           # 静默管理
│   ├── OnDuty/            # 值班管理
│   └── AIAssistant/       # AI 助手
├── stores/                 # 状态管理（Zustand）
│   ├── alertStore.ts
│   ├── configStore.ts
│   └── userStore.ts
├── hooks/                  # 自定义 Hooks
│   ├── useWebSocket.ts
│   ├── useAlerts.ts
│   └── ...
├── utils/                  # 工具函数
│   ├── formatter.ts
│   ├── validator.ts
│   └── ...
├── types/                  # TypeScript 类型定义
│   ├── alert.ts
│   ├── dataSource.ts
│   └── ...
└── App.tsx
```

### 页面详细设计

#### 1. 告警大盘 (Dashboard)

**路由**: `/dashboard`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  Header: 游戏运维 AI 告警系统                                  │
├─────────────────────────────────────────────────────────────┤
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐                    │
│  │ 🔴 P0│  │ 🟠 P1│  │ 🟡 P2│  │ 🔵 P3│                    │
│  │  2   │  │  5   │  │  10  │  │  3   │                    │
│  └──────┘  └──────┘  └──────┘  └──────┘                    │
├─────────────────────────────────────────────────────────────┤
│  最近 24 小时告警趋势                                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  [ECharts 折线图]                                      │  │
│  └───────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│  活跃告警列表 (status=firing)                                │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 🔴 P0 | HighMemory | prometheus | 2024-01-01 10:00   │  │
│  │     主机: server-01 | 触发次数: 3                      │  │
│  │     [确认] [详情]                                      │  │
│  ├───────────────────────────────────────────────────────┤  │
│  │ 🟠 P1 | DiskFull | zabbix | 2024-01-01 09:50         │  │
│  │     主机: server-02 | 触发次数: 1                      │  │
│  │     [确认] [详情]                                      │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**组件结构**:
```tsx
// pages/Dashboard/index.tsx
export const Dashboard: React.FC = () => {
  const { alerts, stats } = useAlerts();
  const { connect, disconnect } = useWebSocket('/ws/alerts');
  
  useEffect(() => {
    connect((message) => {
      // 处理新告警推送
      if (message.type === 'new_alert') {
        // 更新列表
        // 播放提示音（P0/P1）
        // 显示通知
      }
    });
    
    return () => disconnect();
  }, []);
  
  return (
    <div className="dashboard">
      <StatsCards stats={stats} />
      <TrendChart data={stats.trend} />
      <ActiveAlertList 
        alerts={alerts.filter(a => a.status === 'firing')}
        onAck={handleAck}
      />
    </div>
  );
};
```

**关键功能**:
- WebSocket 实时推送，断线自动重连（3秒间隔）
- P0 告警红色高亮并置顶
- 点击确认弹出对话框填写备注
- 支持快速筛选（按级别）

#### 2. 告警管理 (AlertManagement)

**路由**: `/alerts`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  筛选条件                                                      │
│  级别: [P0][P1][P2][P3]  来源: [下拉]  状态: [下拉]          │
│  时间: [时间范围选择器]  Labels: [key:value 搜索]            │
│  [搜索] [重置]                                                │
├─────────────────────────────────────────────────────────────┤
│  告警列表                                                      │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 🔴 P0 | HighMemory | prometheus | firing | 10:00     │  │
│  │ ▼ 展开详情                                             │  │
│  │   ┌─────────────────────────────────────────────────┐ │  │
│  │   │ 原始告警详情:                                    │ │  │
│  │   │ - 主机: server-01                                │ │  │
│  │   │ - 内存使用: 95%                                  │ │  │
│  │   │ - 触发时间: 2024-01-01 10:00:00                 │ │  │
│  │   │ - 触发次数: 3                                    │ │  │
│  │   ├─────────────────────────────────────────────────┤ │  │
│  │   │ AI 分析结果:                                     │ │  │
│  │   │ 摘要: 内存使用率持续超过阈值                     │ │  │
│  │   │ 根因: 可能存在内存泄漏                           │ │  │
│  │   │ 建议:                                            │ │  │
│  │   │   1. 检查应用日志                                │ │  │
│  │   │   2. 重启相关服务                                │ │  │
│  │   │ 标签: [内存泄漏] [性能问题]                      │ │  │
│  │   │ [问 AI] [确认处理] [快速静默]                    │ │  │
│  │   └─────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────┘  │
│  [分页: 1 2 3 ... 10]                                        │
└─────────────────────────────────────────────────────────────┘
```

**组件结构**:
```tsx
// pages/AlertManagement/index.tsx
export const AlertManagement: React.FC = () => {
  const [filters, setFilters] = useState<AlertFilters>({});
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const { data, loading } = useAlerts(filters);
  
  return (
    <div className="alert-management">
      <FilterBar filters={filters} onChange={setFilters} />
      <AlertList 
        alerts={data.alerts}
        expandedIds={expandedIds}
        onToggleExpand={handleToggleExpand}
        onAck={handleAck}
        onSilence={handleQuickSilence}
        onAskAI={handleAskAI}
      />
      <Pagination {...data.pagination} />
    </div>
  );
};

// components/AlertCard/index.tsx
export const AlertCard: React.FC<{alert: Alert}> = ({ alert }) => {
  return (
    <Card className="alert-card">
      <AlertHeader alert={alert} />
      {expanded && (
        <AlertDetails 
          alert={alert}
          aiAnalysis={alert.aiAnalysis}
        />
      )}
    </Card>
  );
};
```

**关键功能**:
- 高级筛选（多条件组合）
- 展开/折叠详情
- 快速静默（预填充匹配条件）
- 问 AI 按钮（调用处置建议 API）

#### 3. 数据源管理 (DataSources)

**路由**: `/data-sources`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  数据源列表                                    [+ 新增数据源]  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ prometheus | Prometheus监控                           │  │
│  │ Webhook: https://domain.com/webhook/prometheus        │  │
│  │ 分组维度: host, zone | 状态: ✅ 启用                   │  │
│  │ 最近触发: 2024-01-01 10:00                            │  │
│  │ [编辑] [禁用] [删除]                                   │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**编辑/新增对话框**:
```
┌─────────────────────────────────────────────────────────────┐
│  新增数据源                                          [X]      │
├─────────────────────────────────────────────────────────────┤
│  名称 (name): [prometheus_____] (创建后不可修改)             │
│  显示名称: [Prometheus监控_____]                             │
│                                                              │
│  输入模板 (Input Template):                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ [Monaco Editor - Go Template 语法高亮]                │  │
│  │ {{- $alert := . -}}                                   │  │
│  │ alert_name: {{ $alert.labels.alertname }}            │  │
│  │ severity: {{ mapValue $alert.labels.severity         │  │
│  │              "critical" "P0" "warning" "P1" }}        │  │
│  │ ...                                                   │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                              │
│  输出模板 (Output Template):                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ [Monaco Editor - Go Template 语法高亮]                │  │
│  │ 🔴 {{ .Severity }} 告警                               │  │
│  │ 告警名称: {{ .AlertName }}                            │  │
│  │ 来源: {{ .Source }}                                   │  │
│  │ ...                                                   │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                              │
│  分组维度 (group_by_labels):                                 │
│  [host] [zone] [+ 添加]                                     │
│                                                              │
│  ┌─ 测试区 ─────────────────────────────────────────────┐  │
│  │ 输入样本数据 (JSON):                                   │  │
│  │ ┌─────────────────────────────────────────────────┐   │  │
│  │ │ {                                                │   │  │
│  │ │   "labels": {                                    │   │  │
│  │ │     "alertname": "HighMemory",                   │   │  │
│  │ │     "severity": "critical"                       │   │  │
│  │ │   }                                              │   │  │
│  │ │ }                                                │   │  │
│  │ └─────────────────────────────────────────────────┘   │  │
│  │ [测试输入模板] [测试输出模板]                          │  │
│  │                                                        │  │
│  │ 解析结果:                                              │  │
│  │ ✅ 输入模板测试通过                                    │  │
│  │ {                                                      │  │
│  │   "alert_name": "HighMemory",                         │  │
│  │   "severity": "P0",                                   │  │
│  │   ...                                                  │  │
│  │ }                                                      │  │
│  │                                                        │  │
│  │ ✅ 输出模板测试通过                                    │  │
│  │ 🔴 P0 告警                                             │  │
│  │ 告警名称: HighMemory                                   │  │
│  │ ...                                                    │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  [取消] [保存]                                               │
└─────────────────────────────────────────────────────────────┘
```

**组件结构**:
```tsx
// pages/DataSources/index.tsx
export const DataSources: React.FC = () => {
  const { dataSources, loading } = useDataSources();
  const [editingSource, setEditingSource] = useState<DataSource | null>(null);
  
  return (
    <div className="data-sources">
      <Button onClick={() => setEditingSource({} as DataSource)}>
        新增数据源
      </Button>
      <DataSourceList 
        sources={dataSources}
        onEdit={setEditingSource}
        onToggle={handleToggle}
        onDelete={handleDelete}
      />
      {editingSource && (
        <DataSourceModal 
          source={editingSource}
          onClose={() => setEditingSource(null)}
          onSave={handleSave}
        />
      )}
    </div>
  );
};

// components/DataSourceModal/index.tsx
export const DataSourceModal: React.FC<Props> = ({ source, onSave }) => {
  const [inputTemplate, setInputTemplate] = useState(source.input_template);
  const [outputTemplate, setOutputTemplate] = useState(source.output_template);
  const [sampleData, setSampleData] = useState('{}');
  const [testResult, setTestResult] = useState<TestResult | null>(null);
  
  const handleTestInput = async () => {
    const result = await api.testInputTemplate({
      template: inputTemplate,
      sample_data: JSON.parse(sampleData)
    });
    setTestResult(result);
  };
  
  const handleTestOutput = async () => {
    const result = await api.testOutputTemplate({
      template: outputTemplate,
      alert_data: testResult?.result
    });
    setTestResult(prev => ({ ...prev, outputResult: result }));
  };
  
  return (
    <Modal>
      <Form>
        <Input label="名称" disabled={!!source.id} />
        <Input label="显示名称" />
        <MonacoEditor 
          label="输入模板"
          language="go-template"
          value={inputTemplate}
          onChange={setInputTemplate}
        />
        <MonacoEditor 
          label="输出模板"
          language="go-template"
          value={outputTemplate}
          onChange={setOutputTemplate}
        />
        <TagInput label="分组维度" />
        
        <TestPanel>
          <JsonEditor 
            label="样本数据"
            value={sampleData}
            onChange={setSampleData}
          />
          <Button onClick={handleTestInput}>测试输入模板</Button>
          <Button onClick={handleTestOutput}>测试输出模板</Button>
          {testResult && <TestResult result={testResult} />}
        </TestPanel>
        
        <Button onClick={onSave}>保存</Button>
      </Form>
    </Modal>
  );
};
```

**关键功能**:
- Monaco Editor 代码高亮和自动补全
- 实时模板测试（输入和输出分别测试）
- 保存前必须通过测试验证
- Webhook URL 自动生成并可复制



#### 4. 推送渠道管理 (Channels)

**路由**: `/channels`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  推送渠道列表                                  [+ 新增渠道]    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 运维飞书群 | 飞书 | ✅ 启用                            │  │
│  │ Webhook: https://open.feishu.cn/... (已脱敏)          │  │
│  │ [编辑] [测试] [禁用] [删除]                            │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**新增/编辑对话框**:
```
┌─────────────────────────────────────────────────────────────┐
│  新增推送渠道                                        [X]      │
├─────────────────────────────────────────────────────────────┤
│  名称: [运维飞书群_____]                                     │
│  类型: [飞书 ▼]                                              │
│                                                              │
│  配置:                                                       │
│  Webhook URL: [https://open.feishu.cn/..._____]             │
│  Secret: [••••••••_____]                                     │
│                                                              │
│  [测试发送] [取消] [保存]                                    │
└─────────────────────────────────────────────────────────────┘
```

**组件结构**:
```tsx
// pages/Channels/index.tsx
export const Channels: React.FC = () => {
  const { channels } = useChannels();
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null);
  
  const handleTest = async (id: number) => {
    await api.testChannel(id);
    message.success('测试消息已发送');
  };
  
  return (
    <div className="channels">
      <Button onClick={() => setEditingChannel({} as Channel)}>
        新增渠道
      </Button>
      <ChannelList 
        channels={channels}
        onEdit={setEditingChannel}
        onTest={handleTest}
        onToggle={handleToggle}
        onDelete={handleDelete}
      />
      {editingChannel && (
        <ChannelModal 
          channel={editingChannel}
          onClose={() => setEditingChannel(null)}
          onSave={handleSave}
        />
      )}
    </div>
  );
};

// components/ChannelModal/index.tsx
export const ChannelModal: React.FC<Props> = ({ channel, onSave }) => {
  const [type, setType] = useState(channel.type || 'feishu');
  const [config, setConfig] = useState(channel.config || {});
  
  const renderConfigForm = () => {
    switch (type) {
      case 'feishu':
        return (
          <>
            <Input 
              label="Webhook URL"
              value={config.webhook_url}
              onChange={v => setConfig({...config, webhook_url: v})}
            />
            <Input 
              label="Secret"
              type="password"
              value={config.secret}
              onChange={v => setConfig({...config, secret: v})}
            />
          </>
        );
      case 'dingtalk':
        // 类似飞书
      case 'wecom':
        return (
          <Input 
            label="Webhook URL"
            value={config.webhook_url}
            onChange={v => setConfig({...config, webhook_url: v})}
          />
        );
      case 'webhook':
        return (
          <>
            <Input label="URL" />
            <Select label="Method" options={['POST', 'GET']} />
            <KeyValueEditor label="Headers" />
            <MonacoEditor label="Body Template (可选)" />
          </>
        );
    }
  };
  
  return (
    <Modal>
      <Form>
        <Input label="名称" />
        <Select 
          label="类型"
          options={[
            { label: '飞书', value: 'feishu' },
            { label: '钉钉', value: 'dingtalk' },
            { label: '企业微信', value: 'wecom' },
            { label: '自定义 Webhook', value: 'webhook' }
          ]}
          value={type}
          onChange={setType}
        />
        {renderConfigForm()}
        <Button onClick={handleTest}>测试发送</Button>
        <Button onClick={onSave}>保存</Button>
      </Form>
    </Modal>
  );
};
```

#### 5. 路由规则管理 (Routes)

**路由**: `/routes`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  路由规则列表 (可拖拽排序)                     [+ 新增规则]    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ ☰ 1. P0告警路由                                        │  │
│  │    匹配: P0 | 所有来源                                  │  │
│  │    推送: 运维飞书群, 值班电话                          │  │
│  │    时间: 全天                                           │  │
│  │    [编辑] [删除]                                        │  │
│  ├───────────────────────────────────────────────────────┤  │
│  │ ☰ 2. 数据库告警路由                                    │  │
│  │    匹配: P0,P1 | prometheus | labels: service=db      │  │
│  │    推送: DBA飞书群                                      │  │
│  │    时间: 09:00-18:00                                   │  │
│  │    [编辑] [删除]                                        │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**新增/编辑对话框**:
```
┌─────────────────────────────────────────────────────────────┐
│  新增路由规则                                        [X]      │
├─────────────────────────────────────────────────────────────┤
│  规则名称: [P0告警路由_____]                                 │
│                                                              │
│  匹配条件:                                                   │
│  级别: [☑ P0] [☐ P1] [☐ P2] [☐ P3]                         │
│  来源: [☑ prometheus] [☐ zabbix] [☐ cloudwatch]            │
│  Labels 条件:                                                │
│    [host___] 匹配 [server-.*___] [+ 添加]                   │
│                                                              │
│  目标渠道:                                                   │
│  [☑ 运维飞书群] [☐ DBA飞书群] [☐ 值班电话]                  │
│                                                              │
│  通知时间段 (可选):                                          │
│  [09:00] - [18:00] [+ 添加时间段]                           │
│                                                              │
│  [取消] [保存]                                               │
└─────────────────────────────────────────────────────────────┘
```

**组件结构**:
```tsx
// pages/Routes/index.tsx
export const Routes: React.FC = () => {
  const { rules } = useRouteRules();
  const [editingRule, setEditingRule] = useState<RouteRule | null>(null);
  
  const handleReorder = async (newOrder: RouteRule[]) => {
    await api.reorderRouteRules(
      newOrder.map((r, i) => ({ id: r.id, priority: i + 1 }))
    );
  };
  
  return (
    <div className="routes">
      <Button onClick={() => setEditingRule({} as RouteRule)}>
        新增规则
      </Button>
      <DraggableList 
        items={rules}
        onReorder={handleReorder}
        renderItem={(rule) => (
          <RouteRuleCard 
            rule={rule}
            onEdit={() => setEditingRule(rule)}
            onDelete={() => handleDelete(rule.id)}
          />
        )}
      />
      {editingRule && (
        <RouteRuleModal 
          rule={editingRule}
          onClose={() => setEditingRule(null)}
          onSave={handleSave}
        />
      )}
    </div>
  );
};
```

#### 6. 静默管理 (Silence)

**路由**: `/silence`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  [活跃规则] [历史规则]                         [+ 新增静默]    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 维护窗口                                               │  │
│  │ 备注: 数据库升级维护                                   │  │
│  │ 匹配: prometheus | .*Database.* | P2,P3              │  │
│  │ 生效时间: 2024-01-01 00:00 - 02:00                   │  │
│  │ 剩余: 1小时30分                                        │  │
│  │ [编辑] [取消静默]                                      │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**新增对话框**:
```
┌─────────────────────────────────────────────────────────────┐
│  新增静默规则                                        [X]      │
├─────────────────────────────────────────────────────────────┤
│  规则名称: [维护窗口_____]                                   │
│  备注: [数据库升级维护_____]                                 │
│                                                              │
│  匹配条件:                                                   │
│  来源: [prometheus ▼]                                        │
│  告警名称 (正则): [.*Database.*_____]                        │
│  级别: [☐ P0] [☐ P1] [☑ P2] [☑ P3]                         │
│  Labels 条件:                                                │
│    [env___] 匹配 [prod___] [+ 添加]                         │
│                                                              │
│  生效时间:                                                   │
│  [1小时] [4小时] [今天结束] [自定义]                         │
│  开始: [2024-01-01 00:00]                                   │
│  结束: [2024-01-01 02:00]                                   │
│                                                              │
│  [取消] [保存]                                               │
└─────────────────────────────────────────────────────────────┘
```

#### 7. 值班管理 (OnDuty)

**路由**: `/on-duty`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  值班排班                                      [+ 新增排班]    │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  [日历视图 - FullCalendar]                            │  │
│  │                                                        │  │
│  │  2024年1月                                             │  │
│  │  ┌────┬────┬────┬────┬────┬────┬────┐                │  │
│  │  │ 日 │ 一 │ 二 │ 三 │ 四 │ 五 │ 六 │                │  │
│  │  ├────┼────┼────┼────┼────┼────┼────┤                │  │
│  │  │    │ 1  │ 2  │ 3  │ 4  │ 5  │ 6  │                │  │
│  │  │    │张三│张三│李四│李四│王五│王五│                │  │
│  │  ├────┼────┼────┼────┼────┼────┼────┤                │  │
│  │  │ 7  │ 8  │ 9  │...                                 │  │
│  │  └────┴────┴────┴────────────────────┘                │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**新增对话框**:
```
┌─────────────────────────────────────────────────────────────┐
│  新增值班排班                                        [X]      │
├─────────────────────────────────────────────────────────────┤
│  值班人员: [张三 ▼]                                          │
│  通知渠道: [值班电话 ▼]                                      │
│  开始时间: [2024-01-01 00:00]                               │
│  结束时间: [2024-01-02 00:00]                               │
│                                                              │
│  [取消] [保存]                                               │
└─────────────────────────────────────────────────────────────┘
```

#### 8. AI 助手 (AIAssistant)

**路由**: `/ai-assistant`

**布局**:
```
┌─────────────────────────────────────────────────────────────┐
│  [智能问答] [静默推荐] [AI日志]                              │
├─────────────────────────────────────────────────────────────┤
│  智能问答                                                     │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 💬 对话历史                                            │  │
│  │                                                        │  │
│  │ 👤 最近一周 P0 告警有多少？                            │  │
│  │                                                        │  │
│  │ 🤖 最近一周共有 15 条 P0 告警，主要来自以下服务：     │  │
│  │    ┌─────────────────────────────────────────────┐   │  │
│  │    │ [ECharts 柱状图]                            │   │  │
│  │    │ prometheus: 8                               │   │  │
│  │    │ zabbix: 5                                   │   │  │
│  │    │ cloudwatch: 2                               │   │  │
│  │    └─────────────────────────────────────────────┘   │  │
│  │    基于最近 7 天告警数据                           │  │
│  │                                                        │  │
│  │ 👤 这些告警的主要原因是什么？                          │  │
│  │                                                        │  │
│  │ 🤖 根据 AI 分析，主要原因包括：                        │  │
│  │    1. 内存泄漏 (40%)                                  │  │
│  │    2. 磁盘空间不足 (30%)                              │  │
│  │    3. 网络抖动 (20%)                                  │  │
│  │    4. 其他 (10%)                                      │  │
│  │                                                        │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ [输入问题...___________________________] [发送]        │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**静默推荐 Tab**:
```
┌─────────────────────────────────────────────────────────────┐
│  AI 静默规则推荐                                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 💡 推荐静默规则                                        │  │
│  │                                                        │  │
│  │ 告警: HighMemory (prometheus)                         │  │
│  │ 最近 7 天出现: 8 次                                    │  │
│  │ 自愈率: 90%                                            │  │
│  │ 触发规律: 每天 02:00-03:00                            │  │
│  │                                                        │  │
│  │ 建议静默规则:                                          │  │
│  │ - 来源: prometheus                                    │  │
│  │ - 告警名称: HighMemory                                │  │
│  │ - 时间: 每天 02:00-03:00                              │  │
│  │ - 持续时间: 4小时                                      │  │
│  │                                                        │  │
│  │ [修改] [采纳] [忽略]                                   │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**AI 日志 Tab**:
```
┌─────────────────────────────────────────────────────────────┐
│  AI 处理日志                                                  │
│  准确率统计: 近7天 85% | 近30天 82%                          │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 2024-01-01 10:00 | HighMemory | ✅ 准确                │  │
│  │ 输入: {...}                                            │  │
│  │ 输出: 摘要: 内存使用率持续超过阈值...                  │  │
│  │ 反馈: 分析准确，建议有效                               │  │
│  │ [查看详情]                                             │  │
│  ├───────────────────────────────────────────────────────┤  │
│  │ 2024-01-01 09:50 | DiskFull | ❌ 不准确                │  │
│  │ 输入: {...}                                            │  │
│  │ 输出: 摘要: 磁盘空间不足...                            │  │
│  │ 反馈: 根因分析有误，实际是日志文件过大                 │  │
│  │ [查看详情]                                             │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**组件结构**:
```tsx
// pages/AIAssistant/index.tsx
export const AIAssistant: React.FC = () => {
  const [activeTab, setActiveTab] = useState('chat');
  
  return (
    <div className="ai-assistant">
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="智能问答" key="chat">
          <AIChat />
        </TabPane>
        <TabPane tab="静默推荐" key="recommendations">
          <SilenceRecommendations />
        </TabPane>
        <TabPane tab="AI日志" key="logs">
          <AILogs />
        </TabPane>
      </Tabs>
    </div>
  );
};

// components/AIChat/index.tsx
export const AIChat: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [sessionId] = useState(() => uuidv4());
  
  const handleSend = async () => {
    const userMessage = { role: 'user', content: input };
    setMessages(prev => [...prev, userMessage]);
    
    const response = await api.aiChat({
      question: input,
      session_id: sessionId
    });
    
    const aiMessage = {
      role: 'assistant',
      content: response.answer,
      chartData: response.chart_data,
      dataRange: response.data_range
    };
    setMessages(prev => [...prev, aiMessage]);
    setInput('');
  };
  
  return (
    <div className="ai-chat">
      <MessageList messages={messages} />
      <Input 
        value={input}
        onChange={setInput}
        onPressEnter={handleSend}
        suffix={<Button onClick={handleSend}>发送</Button>}
      />
    </div>
  );
};
```

### UI 设计规范

#### 配色方案（浅色系）

```css
/* 主色调 */
--primary-color: #1890ff;      /* 蓝色 - 主要操作 */
--success-color: #52c41a;      /* 绿色 - 成功状态 */
--warning-color: #faad14;      /* 橙色 - 警告 */
--error-color: #f5222d;        /* 红色 - 错误/P0 */

/* 告警级别颜色 */
--severity-p0: #f5222d;        /* 红色 */
--severity-p1: #fa8c16;        /* 橙色 */
--severity-p2: #fadb14;        /* 黄色 */
--severity-p3: #1890ff;        /* 蓝色 */

/* 背景色 */
--bg-primary: #ffffff;         /* 主背景 */
--bg-secondary: #f0f2f5;       /* 次背景 */
--bg-tertiary: #fafafa;        /* 三级背景 */

/* 文字颜色 */
--text-primary: #000000d9;     /* 主文字 */
--text-secondary: #00000073;   /* 次文字 */
--text-disabled: #00000040;    /* 禁用文字 */

/* 边框颜色 */
--border-color: #d9d9d9;
```

#### 组件样式

```tsx
// 告警级别徽章
<Badge 
  color={severityColors[severity]}
  text={`${severityEmojis[severity]} ${severity}`}
/>

// 状态标签
<Tag color={statusColors[status]}>
  {statusLabels[status]}
</Tag>

// 卡片阴影
box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);

// 悬停效果
&:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
  transition: all 0.3s ease;
}
```



## 错误处理

### 1. Webhook 接收错误

**场景**: 数据源禁用、模板渲染失败

**处理策略**:
```go
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
    // 1. 数据源不存在或禁用
    if !dataSource.Enabled {
        c.JSON(403, gin.H{"error": "Data source disabled"})
        return
    }
    
    // 2. 模板渲染失败
    alert, err := h.renderTemplate(dataSource.InputTemplate, rawData)
    if err != nil {
        // 生成降级告警
        degradedAlert := h.createDegradedAlert(rawData, err)
        h.saveToStream(degradedAlert)
        
        // 写入死信日志
        h.saveDeadLetter(dataSource.Source, rawData, err)
        
        c.JSON(200, gin.H{
            "alert_id": degradedAlert.AlertID,
            "warning": "Template rendering failed, degraded alert created"
        })
        return
    }
    
    c.JSON(200, gin.H{"alert_id": alert.AlertID})
}
```

### 2. AI 调用错误

**场景**: 超时、API 失败

**处理策略**:
```go
func (a *AIAnalyzer) Analyze(ctx context.Context, group *AggregatedAlert) (*AnalysisResult, error) {
    ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
    defer cancel()
    
    result, err := a.client.Analyze(ctx, group)
    if err != nil {
        // 降级处理
        return &AnalysisResult{
            Summary:   group.Alerts[0].Message,
            RootCause: "AI analysis unavailable",
            Severity:  group.MaxSeverity,
            Suggestions: []string{"Please investigate manually"},
        }, nil
    }
    
    // 验证 severity 不低于原始两个档位
    if !a.validateSeverity(result.Severity, group.MaxSeverity) {
        result.Severity = group.MaxSeverity
    }
    
    return result, nil
}
```

### 3. 推送失败

**场景**: 网络错误、渠道配置错误

**处理策略**:
```go
func (n *Notifier) Send(alert *ProcessedAlert, channel *Channel) error {
    var lastErr error
    
    for i := 0; i < 3; i++ {
        err := n.doSend(alert, channel)
        if err == nil {
            return nil
        }
        
        lastErr = err
        time.Sleep(5 * time.Second)
    }
    
    // 记录失败日志
    log.Error("Failed to send notification after 3 retries",
        "alert_id", alert.AlertID,
        "channel", channel.Name,
        "error", lastErr)
    
    return lastErr
}
```

### 4. 数据库错误

**场景**: 连接失败、查询超时

**处理策略**:
```go
// 使用连接池和重试机制
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    ConnPool: &gorm.ConnPool{
        MaxIdleConns: 10,
        MaxOpenConns: 100,
        ConnMaxLifetime: time.Hour,
    },
})

// 查询超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

var alerts []Alert
if err := db.WithContext(ctx).Find(&alerts).Error; err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return nil, errors.New("database query timeout")
    }
    return nil, err
}
```

### 5. Redis 错误

**场景**: 连接断开、Stream 消费失败

**处理策略**:
```go
// 自动重连
client := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})

// Consumer Group 自动恢复
for {
    streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
        Group:    "alert-processor",
        Consumer: "worker-1",
        Streams:  []string{"alerts:normalized", ">"},
        Count:    10,
        Block:    time.Second,
    }).Result()
    
    if err != nil {
        if err == redis.Nil {
            continue
        }
        log.Error("Failed to read from stream", "error", err)
        time.Sleep(time.Second)
        continue
    }
    
    // 处理消息
    for _, stream := range streams {
        for _, message := range stream.Messages {
            if err := processAlert(message); err != nil {
                log.Error("Failed to process alert", "error", err)
                continue
            }
            // ACK 消息
            client.XAck(ctx, stream.Stream, "alert-processor", message.ID)
        }
    }
}
```

## 测试策略

### 单元测试

**覆盖范围**:
- 模板渲染函数
- 指纹计算函数
- 静默规则匹配
- 路由规则匹配
- AI 结果验证

**示例**:
```go
// template/renderer_test.go
func TestRenderInputTemplate(t *testing.T) {
    tests := []struct {
        name     string
        template string
        data     map[string]interface{}
        want     *Alert
        wantErr  bool
    }{
        {
            name: "valid prometheus alert",
            template: `
alert_name: {{ .labels.alertname }}
severity: {{ mapValue .labels.severity "critical" "P0" "warning" "P1" }}
message: {{ .annotations.summary }}
`,
            data: map[string]interface{}{
                "labels": map[string]string{
                    "alertname": "HighMemory",
                    "severity":  "critical",
                },
                "annotations": map[string]string{
                    "summary": "Memory usage is high",
                },
            },
            want: &Alert{
                AlertName: "HighMemory",
                Severity:  "P0",
                Message:   "Memory usage is high",
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := RenderInputTemplate(tt.template, tt.data)
            if (err != nil) != tt.wantErr {
                t.Errorf("RenderInputTemplate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("RenderInputTemplate() = %v, want %v", got, tt.want)
            }
        })
    }
}

// processor/fingerprint_test.go
func TestCalculateFingerprint(t *testing.T) {
    alert := &Alert{
        Source:    "prometheus",
        AlertName: "HighMemory",
        Labels: map[string]string{
            "host": "server-01",
            "zone": "east-1",
            "env":  "prod",
        },
    }
    
    groupByLabels := []string{"host", "zone"}
    
    fingerprint := CalculateFingerprint(alert, groupByLabels)
    
    // 验证指纹格式
    assert.Len(t, fingerprint, 64) // SHA256 hex length
    
    // 验证相同输入产生相同指纹
    fingerprint2 := CalculateFingerprint(alert, groupByLabels)
    assert.Equal(t, fingerprint, fingerprint2)
    
    // 验证不同 label 顺序产生相同指纹
    groupByLabels2 := []string{"zone", "host"}
    fingerprint3 := CalculateFingerprint(alert, groupByLabels2)
    assert.Equal(t, fingerprint, fingerprint3)
}
```

### 集成测试

**覆盖范围**:
- Webhook 端到端流程
- 告警处理流水线
- 数据库操作
- Redis Stream 消费

**示例**:
```go
// integration/webhook_test.go
func TestWebhookEndToEnd(t *testing.T) {
    // 设置测试环境
    db := setupTestDB(t)
    redis := setupTestRedis(t)
    server := setupTestServer(t, db, redis)
    
    // 创建测试数据源
    dataSource := &DataSource{
        Name:          "test-prometheus",
        InputTemplate: testTemplate,
        Enabled:       true,
    }
    db.Create(dataSource)
    
    // 发送 webhook 请求
    payload := map[string]interface{}{
        "labels": map[string]string{
            "alertname": "HighMemory",
            "severity":  "critical",
        },
    }
    
    resp := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/webhook/test-prometheus", 
        bytes.NewBuffer(toJSON(payload)))
    
    server.ServeHTTP(resp, req)
    
    // 验证响应
    assert.Equal(t, 200, resp.Code)
    
    var result map[string]interface{}
    json.Unmarshal(resp.Body.Bytes(), &result)
    assert.NotEmpty(t, result["alert_id"])
    
    // 验证 Redis Stream
    messages, err := redis.XRead(&redis.XReadArgs{
        Streams: []string{"alerts:normalized", "0"},
        Count:   1,
    }).Result()
    assert.NoError(t, err)
    assert.Len(t, messages[0].Messages, 1)
    
    // 验证数据库
    var alert Alert
    db.Where("alert_id = ?", result["alert_id"]).First(&alert)
    assert.Equal(t, "HighMemory", alert.AlertName)
    assert.Equal(t, "P0", alert.Severity)
}
```

### 性能测试

**测试场景**:
- 高并发 webhook 接收（1000 req/s）
- 大量告警聚合（10000 alerts/window）
- 数据库查询性能

**示例**:
```go
// benchmark/webhook_test.go
func BenchmarkWebhookHandler(b *testing.B) {
    server := setupBenchmarkServer(b)
    payload := createTestPayload()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp := httptest.NewRecorder()
            req := httptest.NewRequest("POST", "/webhook/test", 
                bytes.NewBuffer(payload))
            server.ServeHTTP(resp, req)
            
            if resp.Code != 200 {
                b.Fatalf("unexpected status code: %d", resp.Code)
            }
        }
    })
}
```

### 前端测试

**单元测试** (Jest + React Testing Library):
```tsx
// components/AlertCard/AlertCard.test.tsx
describe('AlertCard', () => {
  it('renders P0 alert with red highlight', () => {
    const alert = {
      alert_id: '123',
      severity: 'P0',
      alert_name: 'HighMemory',
      status: 'firing',
    };
    
    render(<AlertCard alert={alert} />);
    
    expect(screen.getByText('🔴 P0')).toBeInTheDocument();
    expect(screen.getByText('HighMemory')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '确认' })).toBeInTheDocument();
  });
  
  it('calls onAck when confirm button clicked', () => {
    const handleAck = jest.fn();
    const alert = createTestAlert();
    
    render(<AlertCard alert={alert} onAck={handleAck} />);
    
    fireEvent.click(screen.getByRole('button', { name: '确认' }));
    
    expect(handleAck).toHaveBeenCalledWith(alert.alert_id);
  });
});
```

**E2E 测试** (Playwright):
```typescript
// e2e/dashboard.spec.ts
test('dashboard displays active alerts and updates in real-time', async ({ page }) => {
  await page.goto('/dashboard');
  
  // 验证统计卡片
  await expect(page.locator('[data-testid="p0-count"]')).toContainText('2');
  
  // 验证活跃告警列表
  const alertList = page.locator('[data-testid="active-alerts"]');
  await expect(alertList.locator('.alert-card')).toHaveCount(5);
  
  // 模拟 WebSocket 推送新告警
  await page.evaluate(() => {
    window.mockWebSocket.send({
      type: 'new_alert',
      data: {
        alert_id: 'new-123',
        severity: 'P0',
        alert_name: 'NewAlert',
      },
    });
  });
  
  // 验证新告警出现
  await expect(alertList.locator('.alert-card')).toHaveCount(6);
  await expect(page.locator('[data-testid="p0-count"]')).toContainText('3');
});
```



## 正确性属性

正确性属性（Correctness Properties）是关于系统行为的形式化规范，用于验证系统在所有有效输入下都能正确运行。每个属性都是一个可执行的测试，通过属性测试（Property-Based Testing）来验证。

### 属性 1：Webhook 地址唯一性

*对于任意*两个不同的数据源名称，系统生成的 webhook 地址应该不同，且格式为 `/webhook/{source_name}`

**验证需求**: 需求 1.1

### 属性 2：数据源必填字段验证

*对于任意*缺少必填字段（name、display_name、input_template、output_template、group_by_labels）的数据源创建请求，系统应该拒绝并返回错误

**验证需求**: 需求 1.2

### 属性 3：数据源名称不可变

*对于任意*已创建的数据源，尝试修改其 name 字段的请求应该被拒绝

**验证需求**: 需求 1.5

### 属性 4：禁用数据源返回 403

*对于任意*被禁用的数据源，向其 webhook 地址发送的请求应该返回 403 状态码

**验证需求**: 需求 1.6

### 属性 5：模板必填字段验证

*对于任意*input_template，如果渲染后缺少 alert_name、severity 或 message 任一字段，保存数据源配置应该失败

**验证需求**: 需求 1.7

### 属性 6：Severity 值验证

*对于任意*input_template 渲染的 severity 值，如果不是 P0、P1、P2、P3 之一，保存数据源配置应该失败

**验证需求**: 需求 1.8

### 属性 7：模板测试往返一致性

*对于任意*有效的 input_template 和样本数据，测试接口应该返回成功或明确的错误信息，不应该崩溃或超时

**验证需求**: 需求 2.3, 2.4

### 属性 8：模板渲染失败降级

*对于任意*导致 input_template 渲染失败的输入，系统应该生成降级告警，其中 severity=P1，message 包含原始数据前 500 字节，labels 为空

**验证需求**: 需求 2.5

### 属性 9：渲染失败记录死信

*对于任意*导致 input_template 渲染失败的输入，系统应该在死信日志中记录原始数据和错误信息

**验证需求**: 需求 2.6

### 属性 10：指纹计算确定性

*对于任意*告警和 group_by_labels 配置，多次计算应该产生相同的指纹值

**验证需求**: 需求 4.1

### 属性 11：指纹计算顺序无关性

*对于任意*告警，无论 labels 的输入顺序如何，只要 group_by_labels 中指定的 label 值相同，计算出的指纹应该相同

**验证需求**: 需求 4.3

### 属性 12：指纹计算范围限制

*对于任意*告警，添加或修改不在 group_by_labels 中的 label，不应该改变指纹值

**验证需求**: 需求 4.2

### 属性 13：去重 TTL 内识别重复

*对于任意*告警，如果在去重 TTL 时间内发送相同指纹的告警，第二个告警应该被标记为 deduplicated 状态

**验证需求**: 需求 4.4, 4.5

### 属性 14：去重计数递增

*对于任意*活跃告警，每次收到相同指纹的重复告警时，其 trigger_count 应该增加 1

**验证需求**: 需求 4.6

### 属性 15：静默规则时间范围验证

*对于任意*静默规则，如果当前时间不在其 starts_at 到 ends_at 范围内，该规则不应该匹配任何告警

**验证需求**: 需求 5.2

### 属性 16：静默规则全条件匹配

*对于任意*静默规则和告警，只有当告警满足规则的所有配置条件（source、alert_name_pattern、severities、label_matchers）时，才应该被静默

**验证需求**: 需求 5.3, 5.4, 5.5, 5.6, 5.7

### 属性 17：分组键与指纹一致性

*对于任意*告警，其分组键（group_key）和指纹（fingerprint）应该使用相同的计算逻辑，即相同的告警产生相同的分组键和指纹

**验证需求**: 需求 6.1, 6.2

### 属性 18：聚合窗口大小限制

*对于任意*聚合窗口，其包含的告警数量不应该超过 100 条

**验证需求**: 需求 6.4

### 属性 19：AI Severity 调整限制

*对于任意*聚合告警组，AI 输出的 severity 不应该低于原始最高 severity 两个档位（例如原始 P0，AI 最低 P2；原始 P1，AI 最低 P3）

**验证需求**: 需求 7.7

### 属性 20：告警确认幂等性

*对于任意*已确认的告警（status=acked），再次尝试确认应该被拒绝并返回错误提示

**验证需求**: 需求 10.5

### 属性 21：确认 Token 验证

*对于任意*无效或伪造的确认 token，系统应该拒绝确认请求

**验证需求**: 需求 10.6

### 属性 22：确认 Token 过期

*对于任意*超过 24 小时的确认 token，系统应该拒绝确认请求并返回过期错误

**验证需求**: 需求 10.7

### 属性 23：推送重试机制

*对于任意*推送失败的告警，系统应该重试最多 3 次，每次间隔 5 秒

**验证需求**: 需求 9.11

### 属性 24：路由规则优先级顺序

*对于任意*告警，系统应该按照路由规则的 priority 从小到大顺序匹配，并在匹配到第一条规则后停止

**验证需求**: 需求 8.1, 8.6

### 属性 25：值班渠道叠加

*对于任意*P0 或 P1 告警，如果存在当前值班人员，其渠道应该被叠加到路由匹配结果中，而不是替换

**验证需求**: 需求 8.10, 需求 12.4

### 属性 26：WebSocket 断线重连

*对于任意*WebSocket 连接断开事件，客户端应该在 3 秒后自动尝试重连

**验证需求**: 需求 15.3

### 属性 27：数据保留期限

*对于任意*超过 90 天的已处置告警，系统应该将其归档到历史表

**验证需求**: 需求 26.4

### 属性 28：死信日志保留期限

*对于任意*超过 30 天的死信日志，系统应该自动删除

**验证需求**: 需求 26.3

