# 游戏运维告警系统

## What This Is

这是一个面向游戏运维场景的告警管理平台，用于统一接收、处理、聚合、展示和分发来自多种数据源的告警信息。v1.0 已完成 AI 能力移除，并补齐了通知模板原始事件透传与产品内模板预览能力。

## Core Value

运维团队能够稳定地接入、查看、处理并分发告警，而不依赖任何 AI 能力。

## Current State

- 已发版版本：`v1.0 AI Removal Complete`（2026-04-10）
- 当前能力：后端 API、前端控制台、Webhook 接入、通知路由、静默规则、值班管理、模板预览、原始事件字段透传
- 已验证路径：后端无 AI 闭环脚本、前端无 AI 构建/残留扫描、模板 passthrough 端到端验证脚本
- 当前 roadmap/requirements 已归档到 `.planning/milestones/`

## Next Milestone Goals

- 为配置管理和告警处理主链路补充更多自动化测试
- 评估并清理数据库中遗留的历史 AI 字段或表
- 继续改进模板编辑体验，例如更强的字段说明、示例和错误提示

## Constraints

- **Tech stack**: 维持现有 Go + Gin + GORM + PostgreSQL + Redis + React + Vite 技术栈
- **Brownfield**: 必须尊重仓库中的既有结构与未提交改动
- **Continuity**: 不能破坏告警接入、展示、路由、静默和值班主流程

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 先移除 AI 运行时和前端入口，再做文档与验证收口 | 先锁定核心告警主链路，降低回归风险 | ✓ Good |
| 模板能力以“标准字段 + 原始事件上下文”正式契约提供 | 满足自由透传 JSON 字段的同时保持旧模板兼容 | ✓ Good |
| 模板能力进入产品时同步提供后端驱动预览与脚本化验证 | 降低模板作者试错成本，避免隐式契约 | ✓ Good |

<details>
<summary>v1.0 Historical Context</summary>

- v1.0 包含 4 个 phase、12 个 plans，完整归档见 `.planning/milestones/v1.0-ROADMAP.md`
- 里程碑摘要见 `.planning/MILESTONES.md`
- 历史 requirements 见 `.planning/milestones/v1.0-REQUIREMENTS.md`

</details>

---
*Last updated: 2026-04-10 after v1.0 milestone completion*
