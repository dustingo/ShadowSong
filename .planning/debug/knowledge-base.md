# GSD Debug Knowledge Base

Resolved debug sessions. Used by `gsd-debugger` to surface known-pattern hypotheses at the start of new investigations.

---

## vite-ws-proxy-socket-error — Vite websocket proxy reset during local dashboard connection
- **Date:** 2026-04-10
- **Error patterns:** ws proxy socket error, write ECONNABORTED, read ECONNRESET, websocket, vite, localhost:5173, /ws/alerts, origin
- **Root cause:** The websocket upgrader rejected Vite dev-server origins because its `CheckOrigin` allowlist only matched `http://localhost` and `http://127.0.0.1` without ports.
- **Fix:** Updated `internal/handlers/websocket.go` so `CheckOrigin` parses the request origin and allows loopback hosts (`localhost`, `127.0.0.1`, `::1`) regardless of port, which covers Vite dev-server websocket upgrades.
- **Files changed:** internal/handlers/websocket.go
---

## config-channel-update-nil-slice-check — Redundant nil guard in channel config update
- **Date:** 2026-04-13
- **Error patterns:** config update, channel, nil slice check, len(input.Config), json.RawMessage, empty config semantics
- **Root cause:** `UpdateChannel` checked `input.Config != nil` before `len(input.Config) > 0` even though `Config` is a `json.RawMessage` slice type and `len(nil)` is already defined as zero in Go.
- **Fix:** Simplified the update condition to `if len(input.Config) > 0` before assigning `ch.Config`.
- **Files changed:** internal/handlers/config.go
---
