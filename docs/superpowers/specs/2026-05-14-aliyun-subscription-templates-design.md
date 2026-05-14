---
name: aliyun-subscription-templates
description: 阿里云监控订阅通知的输入/输出模板设计
metadata:
  type: project
---

# 阿里云监控订阅通知模板设计

## 概述

本文档定义了阿里云监控订阅通知（`HandleAliYunSubPush`）的输入/输出模板，用于将阿里云订阅数据转换为本项目统一的 Alert 结构。

## 数据源

阿里云监控订阅推送的原始数据结构（`SubPushPayload`）：

```go
type SubPushPayload struct {
    Severity         string      `json:"severity"`
    UserInfo         SubUserInfo `json:"UserInfo"`
    StrategyName     string      `json:"strategyName"`
    RelatedAlertIds  []string    `json:"relatedAlertIds"`
    GroupingId       string      `json:"groupingId"`
    Project          string      `json:"project"`
    RetriggerTime    int         `json:"retriggerTime"`
    Subscription     interface{} `json:"subscription"`
    BatchId          string      `json:"batchId"`
    UserId           string      `json:"userId"`
    EscalationLevel  int         `json:"escalationLevel"`
    Alert            SubAlert    `json:"alert"`
    AlertCount       int         `json:"alertCount"`
    NextEscalateTime int         `json:"nextEscalateTime"`
    StartTime        int64       `json:"startTime"`      // 毫秒时间戳
    Time             int64       `json:"time"`           // 毫秒时间戳
    AutoResolveTime  int         `json:"autoResolveTime"`
}

type SubAlert struct {
    AlertStatus     string          `json:"alertStatus"`
    TraceId         string          `json:"traceId"`
    Severity        string          `json:"severity"`
    Product         string          `json:"product"`
    GroupId         string          `json:"groupId"`
    EventRawContent interface{}     `json:"eventRawContent"`
    Project         string          `json:"project"`
    Source          string          `json:"source"`
    EventType       string          `json:"eventType"`
    UserId          string          `json:"userId"`
    EventContentMap EventContentMap `json:"eventContentMap"`
    Meta            AlertMeta       `json:"meta"`
    DedupId         string          `json:"dedupId"`
    EventName       string          `json:"eventName"`
    Arn             string          `json:"arn"`
    Timestamp       int             `json:"timestamp"`
}

type AlertMeta struct {
    SysEventMeta SysEventMeta `json:"sysEventMeta"`
}

type SysEventMeta struct {
    RegionNameEn  string `json:"regionNameEn"`
    ResourceId    string `json:"resourceId"`
    Product       string `json:"product"`
    EventNameEn   string `json:"eventNameEn"`
    InstanceName  string `json:"instanceName"`
    Level         string `json:"level"`
    Resource      string `json:"resource"`
    RegionNameZh  string `json:"regionNameZh"`
    GroupId       string `json:"groupId"`
    ServiceTypeEn string `json:"serviceTypeEn"`
    EventType     string `json:"eventType"`
    ServiceTypeZh string `json:"serviceTypeZh"`
    RegionId      string `json:"regionId"`
    EventTime     string `json:"eventTime"`
    Name          string `json:"name"`
    Id            string `json:"id"`
    Status        string `json:"status"`
    EventNameZh   string `json:"eventNameZh"`
}
```

## 字段映射表

**重要说明**: 阿里云推送的 JSON 字段名全部为**小写**，如 `alert`、`userInfo`、`meta.sysEventMeta`。

### 核心字段映射

| 阿里云字段 | 本项目字段 | 转换说明 |
|-----------|-----------|---------|
| `.alert.dedupId` | `alert_id` | 去重ID作为唯一标识 |
| `.alert.meta.sysEventMeta.eventNameZh` | `alert_name` | 事件中文名称 |
| `.alert.meta.sysEventMeta.level` | `severity` | 需通过 `toSeverity` 转换 |
| `.alert.alertStatus` | `status` | 需通过 `toStatus` 转换 |
| `.time` | `trigger_time` | 毫秒时间戳，需通过 `toTime` 转换 |
| 组合字段 | `message` | 用户ID + 事件时间 + 服务类型 + 事件名称 |

### Labels 字段映射

| 阿里云字段 | Labels 字段 | 说明 |
|-----------|------------|------|
| `.alert.meta.sysEventMeta.instanceName` | `instance` | 实例名称 |
| `.alert.meta.sysEventMeta.regionId` | `region` | 区域ID |
| `.alert.meta.sysEventMeta.serviceTypeZh` | `service_type` | 服务类型（中文） |
| `.alert.meta.sysEventMeta.eventType` | `event_type` | 事件类型 |
| `.alert.meta.sysEventMeta.product` | `product` | 产品名称 |
| `.alert.meta.sysEventMeta.resourceId` | `resource_id` | 资源ID |

### Severity 映射规则

| 阿里云 Level | 本项目 Severity |
|-------------|----------------|
| CRITICAL | P0 |
| WARN, WARNING | P1 |
| INFO | P2 |
| DEBUG, LOW | P3 |

### Status 映射规则

| 阿里云 alertStatus | 本项目 Status |
|-------------------|---------------|
| TRIGGERED | firing |
| RESOLVED | resolved |
| 其他 | pending |

## 模板定义

### Input Template

**重要**: `source` 字段必须与 DataSource 的 `name` 字段一致，否则路由规则无法匹配。

```json
{
  "alert_id": "{{.alert.dedupId}}",
  "alert_name": "{{.alert.meta.sysEventMeta.eventNameZh}}",
  "severity": "{{toSeverity .alert.meta.sysEventMeta.level}}",
  "message": "{{.userInfo.aliyunId}} {{.alert.meta.sysEventMeta.eventTime}} {{.alert.meta.sysEventMeta.serviceTypeZh}} {{.alert.meta.sysEventMeta.eventNameZh}}",
  "source": "aliyun_event_push",
  "status": "{{toStatus .alert.alertStatus}}",
  "trigger_time": "{{toTime .time}}",
  "labels": {
    "instance": "{{.alert.meta.sysEventMeta.instanceName}}",
    "region": "{{.alert.meta.sysEventMeta.regionId}}",
    "service_type": "{{.alert.meta.sysEventMeta.serviceTypeZh}}",
    "event_type": "{{.alert.meta.sysEventMeta.eventType}}",
    "product": "{{.alert.meta.sysEventMeta.product}}",
    "resource_id": "{{.alert.meta.sysEventMeta.resourceId}}"
  }
}
```

### Output Template

**注意**: `labels` 字段存储的是 Input Template 渲染结果中的扁平结构，可直接访问。

```json
{"title": "[{{.severity}}] {{.alert_name}}", "content": "实例: {{.labels.instance}}\n区域: {{.labels.region}}\n服务: {{.labels.service_type}}\n时间: {{.trigger_time}}\n\n详情: {{.message}}"}
```

## 测试用例

### 测试数据（使用阿里云实际推送的小写字段名）

```json
{
  "severity": "CRITICAL",
  "userInfo": {
    "aliyunId": "user123@example.com"
  },
  "strategyName": "ECS监控策略",
  "alert": {
    "alertStatus": "TRIGGERED",
    "dedupId": "dedup-abc123",
    "meta": {
      "sysEventMeta": {
        "eventNameZh": "实例状态改变",
        "instanceName": "ecs-production-01",
        "regionId": "cn-hangzhou",
        "serviceTypeZh": "云服务器ECS",
        "eventType": "StatusChange",
        "product": "ECS",
        "resourceId": "i-bp1234567890",
        "level": "CRITICAL",
        "eventTime": "2026-05-14 10:30:00"
      }
    }
  },
  "time": 1715665800000
}
```

### 预期输出 Alert

```json
{
  "alert_id": "dedup-abc123",
  "alert_name": "实例状态改变",
  "severity": "P0",
  "message": "user123@example.com 2026-05-14 10:30:00 云服务器ECS 实例状态改变",
  "source": "aliyun-sub",
  "status": "firing",
  "trigger_time": "2026-05-14T02:30:00Z",
  "labels": {
    "instance": "ecs-production-01",
    "region": "cn-hangzhou",
    "service_type": "云服务器ECS",
    "event_type": "StatusChange",
    "product": "ECS",
    "resource_id": "i-bp1234567890"
  }
}
```

### 预期通知内容

```
标题: [P0] 实例状态改变
内容:
实例: ecs-production-01
区域: cn-hangzhou
服务: 云服务器ECS
时间: 2026-05-14T02:30:00Z

详情: user123@example.com 2026-05-14 10:30:00 云服务器ECS 实例状态改变
```

## 实现要点

1. **字段名大小写**: 阿里云推送的 JSON 字段名全部为**小写**（如 `alert`、`userInfo`、`meta.sysEventMeta`）
2. **时间转换**: `time` 是毫秒时间戳，需要 `toTime` 函数正确处理
3. **空值处理**: 部分字段可能为空，模板应使用 `default` 函数提供默认值
4. **字符转义**: 模板中的特殊字符需要正确转义

## Why

阿里云监控订阅通知是重要的告警数据来源，需要将其转换为本项目统一的 Alert 结构，以便进行后续的路由、去重和通知处理。

## How to apply

1. 在 DataSource 配置中创建名为 `aliyun-sub` 的数据源
2. 将 Input Template 和 Output Template 配置到该数据源
3. 配置对应的 API Key 和路由规则
4. 使用测试用例验证模板渲染结果
