<div align="center">

# ShadowSong

**Multi-Source Alert Management Platform**

统一接收 · 智能路由 · 多通道投递 · 全链路追踪

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-4169E1?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

</div>

---

<p align="center">
  <img src="docs/index.png" alt="ShadowSong Dashboard" width="100%">
</p>

---

## Why ShadowSong?

监控系统各有各的告警通道 — Prometheus 走 Alertmanager、Zabbix 发邮件、Grafana 推 Slack。ShadowSong 把它们统一接入，按规则路由到正确的团队和通道，并记录每条告警从接达到投递的完整生命周期。

## Features

| | |
|---|---|
| **Multi-Source Ingest** | Webhook 接收 + 双阶段模板管道，适配任意 JSON 告警源 |
| **Smart Routing** | 基于级别 / 来源 / Labels 的规则匹配，优先级排序，多通道扇出 |
| **Channel Fan-Out** | 飞书 · 钉钉 · 企微（内置）+ 通用 Webhook（Basic Auth / Custom Header） |
| **Delivery Tracking** | 逐条投递记录、重试 / 重放、升级链、完整审计日志 |
| **Dedup & Group** | 指纹去重 + Labels 分组聚合，减少告警噪音 |
| **Real-Time Push** | WebSocket 告警实时更新到 Dashboard |
| **Silence Management** | 时间窗口静默，从告警一键创建 |
| **RBAC** | admin / operator / viewer 三角色，5 项权限能力，全操作审计 |

## Template Pipeline

```
Webhook JSON ──▶ input_template ──▶ Alert (normalized)
                                         │
                  output_template ──▶ Notification ──▶ Channel
```

Available variables: `.alert_name` · `.severity` · `.message` · `.source` · `.labels` · `.event.*` · `.trigger_time`

Helper functions: `default` · `lookup` · `toJson` · `get`

Severity mapping: `critical` → P0 · `warning` / `error` → P1 · `info` → P2 · `debug` → P3

## Quick Start

> **Prerequisites:** Go 1.25+, Node.js 18+, pnpm 10+, Docker

```bash
# 1. Start PostgreSQL & Redis
make docker-up

# 2. Set JWT_SECRET in .env (≥ 32 chars for production)

# 3. Start backend
make run

# 4. Start frontend
make frontend-dev
```

Open http://localhost:5173

## RBAC

| Role | View Alerts | Process Alerts | View Config | Manage Config | Manage Users |
|------|:---:|:---:|:---:|:---:|:---:|
| **admin** | ✓ | ✓ | ✓ | ✓ | ✓ |
| **operator** | ✓ | ✓ | ✓ | — | — |
| **viewer** | ✓ | — | ✓ | — | — |

## Tech Stack

| Backend | Frontend |
|---------|----------|
| Go · Gin · GORM | React 18 · TypeScript |
| PostgreSQL · Redis | PrimeReact · PrimeFlex |
| JWT · RBAC | Zustand · ECharts · Monaco Editor |
| | Vite · pnpm |

## Development

```bash
make help              # All available commands
make test              # Run all tests
make lint              # Lint backend
make frontend-lint     # Lint frontend
make frontend-test     # Run frontend tests
```

## License

[MIT](LICENSE)