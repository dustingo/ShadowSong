# Retrospective

## Milestone: v1.0 — AI Removal Complete

**Shipped:** 2026-04-10  
**Phases:** 4 | **Plans:** 12

### What Was Built

- 移除了后端 AI 配置、路由、运行时文件和持久化残留
- 移除了前端 AI 页面、入口、展示字段和调用链
- 对齐 README、环境基线、codebase map 和验证入口到非 AI 当前态
- 补齐了通知模板原始事件透传、产品内预览和端到端验证脚本

### What Worked

- 先拆除 AI 主链路，再清理文档和验证，回归面更可控
- PowerShell 自清理验证脚本非常适合当前 Windows/Go/React 工作流
- 模板契约和 UI 预览一起交付，避免功能可用但用户不会用

### What Was Inefficient

- `REQUIREMENTS.md` 在执行中没有持续同步，归档时仍保留旧状态
- `STATE.md` 的部分字段格式和 gsd-tools 预期不一致，需要手动收口
- 阶段结束后才做 milestone audit，会让里程碑归档前的状态修正成本变高

### Patterns Established

- 废弃能力下线后必须补脚本化验证，而不是只依赖代码审查
- 模板系统改动应同时交付共享渲染契约、UI 指导和端到端验证
- dirty worktree 下始终按文件精确提交，避免吞掉用户本地改动

### Key Lessons

- 用户能直接感知的字段契约必须显式命名，不能让用户猜内部映射
- 预览接口必须复用 live 渲染路径，否则说明文档会快速漂移
- 里程碑结束前需要一次 requirements/state 对账，否则归档文件会落后于真实执行结果

### Cost Observations

- 主要工作集中在 docs / fix / test 提交，阶段粒度清晰
- 真实验证脚本虽然耗时较长，但能有效提前暴露契约和路由问题

## Cross-Milestone Trends

- v1.0：以“先收主链路、再补证据”的顺序最稳
