---
phase: 02
slug: remove-frontend-ai-surfaces
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-09
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | TypeScript compiler + Vite build + repo grep checks |
| **Config file** | `frontend/package.json` |
| **Quick run command** | `pnpm build` |
| **Full suite command** | `pnpm build` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run `pnpm build` for tasks that changed TypeScript contracts or imports; otherwise run the task's local grep command first, then `pnpm build` at the end of the plan.
- **After every plan wave:** Run `pnpm build`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | FEAI-01 | T-02-01 / — | AI page file/export is removed so no dead route target remains | static grep | `powershell -NoProfile -Command "if (Test-Path 'frontend/src/pages/AIAssistant.tsx') { exit 1 }; rg -n 'AIAssistant' frontend/src/pages/index.ts; if ($LASTEXITCODE -eq 0) { exit 1 }"` | ✅ | ⬜ pending |
| 02-01-02 | 01 | 1 | FEAI-01 | T-02-01 / T-02-02 | Shell no longer registers `/ai`, AI menu label, or AI route import | static grep | `powershell -NoProfile -Command "rg -n '/ai|AIAssistant|RobotOutlined|AI 助手|游戏运维 AI 告警系统' frontend/src/App.tsx; if ($LASTEXITCODE -eq 0) { exit 1 }"` | ✅ | ⬜ pending |
| 02-02-01 | 02 | 1 | FEAI-02 | T-02-03 / T-02-04 | Dashboard and shared alert card expose only real alert operations, with no AI calls or modal state | static grep | `powershell -NoProfile -Command "rg -n 'onAskAI|问 AI|AI 分析|AI 响应|aiApi|ReactMarkdown|handleAskAI|handleSendToAI|aiModalVisible|aiResponse' frontend/src/components/AlertCard.tsx frontend/src/pages/Dashboard.tsx; if ($LASTEXITCODE -eq 0) { exit 1 }"` | ✅ | ⬜ pending |
| 02-02-02 | 02 | 1 | FEAI-02 | T-02-05 / — | Alerts expanded content shows operational fields only | static grep | `powershell -NoProfile -Command "rg -n 'AI 分析|ai_summary|ai_root_cause|ai_suggestions' frontend/src/pages/Alerts.tsx; if ($LASTEXITCODE -eq 0) { exit 1 }"` | ✅ | ⬜ pending |
| 02-03-01 | 03 | 2 | FEAI-03 | T-02-06 / T-02-07 | Shared API/type contracts no longer define AI-only endpoints or fields | static grep | `powershell -NoProfile -Command "rg -n --hidden --glob '!frontend/dist/**' 'aiApi|/ai/|ai_summary|ai_root_cause|ai_suggestions|ai_tags|ai_severity' frontend/src frontend/index.html; if ($LASTEXITCODE -eq 0) { exit 1 }"` | ✅ | ⬜ pending |
| 02-03-02 | 03 | 2 | FEAI-03 | T-02-08 / — | Touched titles are non-AI and the frontend still builds cleanly | build | `pnpm build` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Authenticated shell no longer exposes an AI entry in the sidebar | FEAI-01 | Visual confirmation of menu collapse is faster in-browser than encoding as another CLI snapshot step | Start the frontend, log in, confirm the sidebar jumps directly from `值班管理` to no further item and navigation still works for remaining modules |
| Dashboard layout collapses cleanly after AI card/modal removal | FEAI-02 | Visual spacing/hierarchy is covered by UI-SPEC and best checked in-browser after automation passes | Open dashboard with active alerts, verify connection warning/P0 cards remain focal and no empty AI container or modal trigger remains |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-04-09
