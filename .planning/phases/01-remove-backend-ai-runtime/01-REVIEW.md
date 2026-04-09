---
phase: 01-remove-backend-ai-runtime
reviewed: 2026-04-09T09:12:27.5874242Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - D:/goproject/shadowsongAI/internal/config/config.go
  - D:/goproject/shadowsongAI/internal/router/router.go
  - D:/goproject/shadowsongAI/internal/models/alert.go
  - D:/goproject/shadowsongAI/internal/models/models.go
  - D:/goproject/shadowsongAI/internal/database/postgres.go
  - D:/goproject/shadowsongAI/internal/handlers/config.go
  - D:/goproject/shadowsongAI/internal/config/config_test.go
  - D:/goproject/shadowsongAI/internal/router/router_test.go
  - D:/goproject/shadowsongAI/scripts/verify_backend_no_ai.ps1
findings:
  critical: 0
  warning: 2
  info: 1
  total: 3
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-04-09T09:12:27.5874242Z
**Depth:** standard
**Files Reviewed:** 9
**Status:** issues_found

## Summary

本次审查覆盖了 Phase 01 指定的后端源码、回归测试和验证脚本，并结合 3 份 PLAN/SUMMARY 核对“移除后端 AI 运行时”的预期结果。`go test ./internal/config ./internal/router` 通过，`scripts/verify_backend_no_ai.ps1` 在当前工作树内可跑通主链路，但仍有 2 个会影响“移除完整性/失败清理稳定性”的警告项，以及 1 个测试覆盖缺口。

## Warnings

### WR-01: 现有数据库不会真正清理历史 AI 表和字段

**File:** `D:/goproject/shadowsongAI/internal/database/postgres.go:44-66`
**Issue:** 当前改动只是把 AI 模型从 `tables` 列表里移除，但 `InitDB` 只会 `CreateTable`/`AutoMigrate` 仍保留的模型，不会删除旧表或旧列。也就是说，已升级过的环境里，历史 AI 表和 `alerts` 上的 AI 字段仍会保留。对“完整移除 AI 相关配置/结构”的 phase 目标来说，这会留下持久化残留。
**Fix:**
```go
// one-time cleanup for removed AI schema
if migrator.HasTable("ai_logs") {
	_ = migrator.DropTable("ai_logs")
}
if migrator.HasTable("silence_recommendations") {
	_ = migrator.DropTable("silence_recommendations")
}

for _, col := range []string{
	"ai_summary",
	"ai_root_cause",
	"ai_severity",
	"ai_suggestions",
	"ai_tags",
} {
	if migrator.HasColumn(&models.Alert{}, col) {
		_ = migrator.DropColumn(&models.Alert{}, col)
	}
}
```

### WR-02: 验证脚本在服务异常退出时可能遗留测试数据

**File:** `D:/goproject/shadowsongAI/scripts/verify_backend_no_ai.ps1:523-526`
**Issue:** `finally` 里只有在 `$script:ServerProcess` 存在且 `-not HasExited` 时才调用 `Clear-TestData`。但测试数据是通过 `docker exec ... psql` 写入的，和服务进程是否还活着无关。如果服务在写入测试数据后崩掉，清理会被跳过，数据库会残留 `users/data_sources/channels/route_rules/alerts/silence_rules` 测试记录。
**Fix:**
```powershell
$script:Seeded = $false

# after successful seed
$script:Seeded = $true

finally {
  try {
    if ($script:Seeded) {
      Clear-TestData
    }
  } catch {
    Add-Content -LiteralPath $script:RunLog -Value ("CLEANUP_ERROR: " + $_.Exception.Message)
  }

  Stop-Resources
}
```

## Info

### IN-01: 缺少数据库迁移清理的回归测试

**File:** `D:/goproject/shadowsongAI/internal/database/postgres.go:44-66`
**Issue:** Phase 01 新增了配置和路由回归测试，但没有任何测试锁定“AI 表/字段不会再出现在迁移逻辑里”这个事实。后续如果有人把 AI 模型重新加回迁移列表，当前测试集不会报错。
**Fix:** 增加 `internal/database/postgres_test.go`，至少断言迁移目标不包含 `AILog`/`SilenceRecommendation`，并对历史 schema 清理逻辑加一个回归用例。

---

_Reviewed: 2026-04-09T09:12:27.5874242Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
