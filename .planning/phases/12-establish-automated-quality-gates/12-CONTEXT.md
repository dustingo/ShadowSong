# Phase 12: Establish Automated Quality Gates - Context

**Gathered:** 2026-04-21
**Status:** Ready for planning

<domain>
## Phase Boundary

本 phase 只做两类事情：一是把当前已经可本地通过的后端测试、前端 lint、前端测试和前端构建固化成自动执行的 CI 门禁；二是收口会直接影响工程真相的 README、工程命名和 planning 文档表述，使其继续与“非 AI 告警系统”现状一致。范围不包含新增业务能力、替换 CI 平台、重构 Go module 路径或清理所有历史 AI 痕迹。

</domain>

<decisions>
## Implementation Decisions

### CI Scope
- **D-01:** Phase 12 的 CI 必须覆盖四条明确检查链路：`go test ./...`、前端 `pnpm lint`、前端 `pnpm test -- --run`、前端 `pnpm build`。
- **D-02:** CI 以“合并前阻断回归”为目标，工作流输出必须按步骤清晰拆分，不能把多类检查混在一个难以定位失败点的黑箱脚本里。

### CI Platform And Contract
- **D-03:** 在当前仓库没有现成 CI 基线的前提下，优先使用仓库原生的 `.github/workflows` 方案建立 GitHub Actions 门禁，而不是引入外部 CI 服务或自定义执行器。
- **D-04:** CI 命令尽量复用仓库现有真命令和技术栈事实，不额外发明第二套脚本协议；缺少的只是自动化编排，不是新的构建体系。

### Documentation And Naming Cleanup
- **D-05:** Phase 12 只收口会误导当前工程真相的命名和文档，例如 README、前端包名、工程入口表述和 planning 文档里的里程碑描述。
- **D-06:** Go module 路径、历史测试名、旧验证脚本名里保留的 `ai-alert-system` / `no_ai` 痕迹不在本 phase 进行全量迁移，只要不继续对外误导当前产品定位即可。

### Verification Boundary
- **D-07:** Phase 12 完成时必须能在本地或 CI 语义上验证工作流定义正确，并确认 README / planning 真相文件与当前阶段顺序、命名和目标保持一致。
- **D-08:** 如果 CI 工作流依赖仓库托管平台能力而无法在本地完整执行，则至少要验证 YAML、命令引用和关键路径一致性，不伪造“已跑远端 CI”的结论。

### the agent's Discretion
- GitHub Actions 中 job 的拆分粒度、缓存策略和触发条件可由后续计划阶段决定，只要满足“步骤清晰、失败可定位、复用现有命令”。
- 文档对齐可按最小必要修正推进，不要求一次扫清所有历史设计文档中的 AI 时代遗留内容。

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone And Phase Truth
- `.planning/PROJECT.md` — 当前里程碑约束、Active requirement 与“非 AI 告警系统”命名要求
- `.planning/REQUIREMENTS.md` — Phase 12 对应的 `CIV-01` 到 `CIV-03`、`DOCS-01`、`DOCS-02`
- `.planning/ROADMAP.md` — Phase 12 的目标、依赖和 success criteria
- `.planning/STATE.md` — 当前阶段位置与 Phase 13 的依赖顺序

### Existing Build And Test Truth
- `Makefile` — 后端 `go test ./...` 和常用本地命令入口
- `frontend/package.json` — 前端 `lint`、`test`、`build` 脚本真相
- `README.md` — 当前对外项目说明、启动方式和命名表述
- `go.mod` — 当前 Go module 路径和实际工程命名事实

### Live Code And Repo Structure
- `cmd/server/main.go` — 当前后端入口及 module import 现状
- `.planning/phases/11-restore-frontend-quality-baseline/11-VERIFICATION.md` — Phase 11 已验证通过的前端质量基线，可直接作为 CI 输入
- `.planning/phases/10-secure-realtime-alert-access/10-VERIFICATION.md` — 近期已稳定的后端/前端验证真相
- `.github/` — 当前不存在现成 workflow，需要在本 phase 新建

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Makefile:test` 已经直接执行 `go test -v ./...`，适合作为后端质量门禁命令真源。
- `frontend/package.json` 已经具备 `lint`、`test`、`build` 三条前端命令，不需要额外创建新脚本才能接入 CI。
- Phase 11 已经证明这些前端命令在当前仓库可通过，因此 Phase 12 的主要工作是编排和真相对齐。

### Established Patterns
- planning 文档已经把 v1.2 的阶段顺序与 requirement traceability 固化下来，Phase 12 的文档更新需要继续沿用这套真相体系。
- 仓库仍保留部分历史 AI 命名，例如前端包名 `ai-alert-system-frontend`，但对 Go module 和内部 import 的改动面会明显更大，因此本 phase 应只处理对外或低风险命名。
- README 已经明确“当前版本已经完成 AI 能力移除”，说明文档方向正确，但仍有脚本命名与旧包名需要继续收口。

### Integration Points
- `.github/workflows/*.yml` 将成为 Phase 12 的主交付物，直接绑定 `go`, `node`, `pnpm` 和前端工作目录。
- `README.md` 和 `frontend/package.json` 是最直接的工程命名对齐点。
- `.planning/PROJECT.md`、`.planning/ROADMAP.md`、`.planning/REQUIREMENTS.md`、`.planning/STATE.md` 需要在 phase closeout 时同步反映 CI 门禁已建立。

</code_context>

<specifics>
## Specific Ideas

- 当前没有任何 `.github/workflows`，所以 GitHub Actions 是最自然且最小增量的 CI 选项。
- `frontend/package.json` 的包名仍是 `ai-alert-system-frontend`，这属于低风险且直接可见的工程命名项，适合在本 phase 收口。
- README 中“无 AI 验证脚本”的描述可以保留事实，但要避免继续把项目整体叙述绑在旧 AI 对照语境上。

</specifics>

<deferred>
## Deferred Ideas

- 全量迁移 `go.mod` module path 与所有 Go import 路径中的 `ai-alert-system`
- 清理所有历史规格文档、验证脚本和测试名中的旧 AI 残留
- 加入矩阵测试、跨平台 runner、bundle size 检查或更复杂的工程门禁

None of the above belongs in Phase 12 scope.

</deferred>

---

*Phase: 12-establish-automated-quality-gates*
*Context gathered: 2026-04-21*
