---
status: resolved
trigger: "Investigate issue: vite-ws-proxy-socket-error"
created: 2026-04-10T00:00:00+08:00
updated: 2026-04-10T11:40:00+08:00
---

## Current Focus

hypothesis: Verified in code: websocket upgrades were rejected because `internal/handlers/websocket.go` only allowed origins without the Vite dev-server port, so browser `Origin` values like `http://localhost:5173` failed `CheckOrigin`.
test: Human verification in the real Vite dev workflow by opening the dashboard and observing whether `/ws/alerts` connects without proxy socket errors.
expecting: Vite should stop logging `ws proxy socket error` for `/ws/alerts`, and the dashboard should show the realtime connection as connected.
next_action: Have the user run the frontend/backend locally and confirm the original reproduction no longer triggers the proxy socket errors.

## Symptoms

expected: Dashboard WebSocket should connect cleanly through the Vite `/ws` proxy in local development, with no proxy socket errors.
actual: Vite logs repeated `ws proxy socket error` messages: `write ECONNABORTED` and `read ECONNRESET`.
errors: `11:30:16 [vite] ws proxy socket error: Error: write ECONNABORTED ...` and `11:30:16 [vite] ws proxy socket error: Error: read ECONNRESET ...`
reproduction: Run the frontend in dev mode, open the dashboard page, let it create `new WebSocket(`${protocol}//${window.location.host}/ws/alerts`)`, and observe Vite terminal output.
started: Reported after the recent non-AI cleanup milestone; exact first bad commit unknown.

## Eliminated

## Evidence

- timestamp: 2026-04-10T00:00:00+08:00
  checked: frontend/src/pages/Dashboard.tsx
  found: The dashboard opens `new WebSocket(`${protocol}//${window.location.host}/ws/alerts`)`, so in Vite dev it targets the dev server host such as `localhost:5173`.
  implication: The browser-originated websocket request will carry the Vite dev-server origin, not the backend origin on port 8080.

- timestamp: 2026-04-10T00:00:00+08:00
  checked: frontend/vite.config.ts
  found: Vite proxies `/ws` to `ws://localhost:8080` with `ws: true`.
  implication: The backend receives proxied websocket upgrades originating from the browser's dev-server page, so origin checks must tolerate the Vite origin.

- timestamp: 2026-04-10T00:00:00+08:00
  checked: internal/handlers/websocket.go
  found: `websocket.Upgrader.CheckOrigin` only allows exact matches for `http://127.0.0.1` and `http://localhost`, excluding port-qualified origins like `http://localhost:5173` and `http://127.0.0.1:5173`.
  implication: Standard Vite dev-server origins will be rejected during websocket upgrade.

- timestamp: 2026-04-10T00:00:00+08:00
  checked: internal/router/router.go
  found: `/ws/alerts` is exposed directly without auth or alternate origin handling.
  implication: The upgrader's origin gate is the primary backend control point for accepting or rejecting the proxied websocket connection.

- timestamp: 2026-04-10T11:37:26+08:00
  checked: internal/handlers/websocket.go
  found: The fix now parses the `Origin` header and allows loopback hosts `localhost`, `127.0.0.1`, and `::1` regardless of dev-server port.
  implication: Browser websocket upgrades from Vite dev origins such as `http://localhost:5173` should now pass backend origin validation.

- timestamp: 2026-04-10T11:37:26+08:00
  checked: go test ./internal/handlers ./internal/router
  found: Both packages compile and test successfully after the origin-handling change.
  implication: The websocket fix is syntactically and structurally valid within the backend route/handler path.

## Resolution

root_cause: The websocket upgrader rejects Vite dev-server origins because its `CheckOrigin` allowlist only matches `http://localhost` and `http://127.0.0.1` without ports.
fix: Updated `internal/handlers/websocket.go` so `CheckOrigin` parses the request origin and allows loopback hosts (`localhost`, `127.0.0.1`, `::1`) regardless of port, which covers Vite dev-server websocket upgrades.
verification: `go test ./internal/handlers ./internal/router` passed after the fix, and the user confirmed the Vite dev workflow no longer logs `ws proxy socket error` for `/ws/alerts`.
files_changed: [internal/handlers/websocket.go]
