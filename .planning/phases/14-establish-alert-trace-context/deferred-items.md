# Deferred Items

- Resolved during Phase 14 completion: `go test ./... -count=1` initially failed in `internal/router` because `internal/router/router_test.go` calls `handlers.NewWebhookHandler(nil, nil)`, and the constructor had started dereferencing a nil Redis client. Commit `2fcb8f0` restored the previous nil-safe behavior and the full Go suite passed afterward.
