# Deferred Items

- `go test ./... -count=1` currently fails in `internal/router` because `internal/router/router_test.go` calls `handlers.NewWebhookHandler(nil, nil)`, and the constructor still dereferences a nil Redis client. This is pre-existing test debt outside plan `14-02` scope.
- `go test ./... -count=1` also hit a Windows-specific file lock while building `internal/database.test.exe` (`The process cannot access the file because it is being used by another process`). This is environmental and unrelated to the webhook lifecycle changes in plan `14-02`.
