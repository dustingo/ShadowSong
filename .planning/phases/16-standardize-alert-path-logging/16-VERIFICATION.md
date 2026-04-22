---
phase: 16-standardize-alert-path-logging
verified: 2026-04-22T01:46:00Z
status: passed
scope: webhook-to-notification alert-path logging
---

# Phase 16: Standardize Alert Path Logging Verification

**Phase Goal:** 把 webhook 到通知发送主链路的日志字段契约固化为可回归、可检索、可复用的仓库真源，同时保持 Phase 14 的 `trace_id` 生命周期和 Phase 15 的 retry/`terminal_failure` 诊断路径不变。  
**Verified:** 2026-04-22T01:46:00Z  
**Status:** passed

## Scope Statement

本 phase 只覆盖 `internal/handlers/webhook.go` 的 webhook-to-notification logging contract：

- 保留现有 `log.Logger` seam
- 保留既有 `stage` taxonomy
- 保留 `trace_id`、`attempt`、`max_attempts`、`terminal_failure` 等 Phase 14/15 契约
- 不引入 `log/slog`
- 不引入 JSON logs、外部观测平台、WebSocket logging cleanup 或 repo-wide logging rewrite

## Automated Commands Executed

```bash
go test ./internal/handlers -run "TestWebhookHandler(.*Lifecycle.*|.*Route.*|.*SendNotification.*|.*Terminal.*|.*Panic.*|.*Trace.*|.*Redis.*|.*Logging.*)" -count=1
go test ./internal/handlers -count=1
```

## Canonical Field Vocabulary

Phase 16 固化的 webhook alert-path canonical fields：

| Field | Meaning | Notes |
| --- | --- | --- |
| `stage` | 生命周期阶段名 | 必须继续使用既有 taxonomy，不重新命名 |
| `trace_id` | 单次 webhook 请求的服务端关联标识 | 继续作为排障回查主键 |
| `alert_id` | 告警主键 | 与 `trace_id` 互补，不替代 |
| `fingerprint` | 去重/聚合指纹 | 用于 dedup 和 route 关联 |
| `source` | 数据源名称 | webhook source truth |
| `channel_id` | 通知渠道 ID | 有 channel 上下文时出现 |
| `channel_name` | 通知渠道名 | 有 channel 上下文时出现 |
| `channel_type` | 通知渠道类型 | 有 channel 上下文时出现 |
| `attempt` | 当前发送尝试次数 | 发送路径固定字段 |
| `max_attempts` | 本次发送最大尝试次数 | 与 `attempt` 配对 |
| `error` | 当前失败上下文 | 作为字段存在，消息文本不再承担机器契约职责 |
| `matched_channels` | 路由命中的渠道数量 | Phase 16 从 message 提升为字段 |
| `mode` | 通知内容模式 | 典型值：`rendered`、`default` |

## Searchable Stage Taxonomy

以下 stage 名称在 Phase 16 后仍然是稳定检索词：

```text
ingest
persist
dedup
redis_publish
route_match
notification_entry
send_attempt
send_notification
terminal_failure
async_panic
```

## Representative Field Examples

以下示例只展示 bounded field contract，不包含原始 payload 或敏感内容：

```text
stage=route_match alert_id=alert-route-structured fingerprint=fp-route-structured matched_channels=1 source=prometheus trace_id=trace-route-structured matched route rules
stage=send_attempt alert_id=alert-terminal attempt=3 channel_id=44 channel_name=ops-webhook channel_type=webhook error=channel 44 (...) send failed... max_attempts=3 mode=rendered source=prometheus trace_id=trace-terminal notification attempt recorded
stage=terminal_failure alert_id=alert-terminal attempt=3 channel_id=44 channel_name=ops-webhook channel_type=webhook error=channel 44 (...) send failed... max_attempts=3 mode=rendered source=prometheus trace_id=trace-terminal retry budget exhausted for rendered notification
```

## Operator Troubleshooting Path

当值班人员从 `terminal_failure` 或 `send_attempt` 开始排查时，搜索路径应保持如下：

1. 先以失败日志中的 `trace_id` 作为主键，例如 `trace_id=trace-terminal`
2. 在同一 `trace_id` 下确认发送链路：
   `stage=terminal_failure` 或 `stage=send_attempt`
3. 回查通知入口：
   `stage=notification_entry trace_id=<same>`
4. 回查路由命中：
   `stage=route_match trace_id=<same>`
   这里应看到 `matched_channels=<n>`
5. 回查 Redis handoff：
   `stage=redis_publish trace_id=<same>`
6. 回查落库：
   `stage=persist trace_id=<same>`
7. 回查接入：
   `stage=ingest trace_id=<same>`

简化检索链路示例：

```text
terminal_failure -> send_attempt -> notification_entry -> route_match -> redis_publish -> persist -> ingest
```

这条回查路径直接延续 Phase 14 的 lifecycle evidence 和 Phase 15 的 retry evidence，而不是重新定义新的观测模型。

## Carry-Forward Contract

- Phase 14 verification truth: `.planning/phases/14-establish-alert-trace-context/14-VERIFICATION.md`
  继续提供 `ingest`、`persist`、`dedup`、`redis_publish`、`route_match`、`notification_entry` 的 `trace_id` 生命周期证据。
- Phase 15 verification truth: `.planning/phases/15-harden-notification-retry-boundaries/15-VERIFICATION.md`
  继续提供 `attempt`、`max_attempts`、`send_attempt`、`send_notification`、`terminal_failure` 的 retry/terminal contract。

Phase 16 的新增真相是：

- `matched_channels` 成为 route-match named field
- `mode` 成为 send-path named field
- `channel_type` 在 channel-bearing logs 中保持一致
- 机器可读值优先进入字段，而不是依赖消息文本解析

## Regression Truth

`internal/handlers/webhook_test.go` 当前锁定了以下 field-level contract：

- `stage=route_match` 与 `matched_channels` 同行出现
- `stage=send_attempt` 保留 `trace_id`、`attempt`、`max_attempts`、`mode`、`channel_type`
- `stage=terminal_failure` 保留 `trace_id`、`attempt`、`max_attempts`、`mode`、`channel_type`
- datasource/render fallback 仍使用同一 field vocabulary，而不是重新发明字段名
- 断言不依赖完整错误消息文案，只要求稳定字段存在并可解析

## Conclusion

Phase 16 已把 webhook-to-notification logging contract 固化为仓库内真源，但范围仍然局限于 `internal/handlers/webhook.go` 现有 `log.Logger` seam。它不声称已经完成 repo-wide 日志统一，也不代表已经接入新的日志基础设施。
