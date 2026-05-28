# Webhook Channel Enhancement Design

## Problem

The current webhook sender only supports JSON body with no authentication options. This makes it impossible to integrate with external alerting platforms like Odin, which require `application/x-www-form-urlencoded` body format and HTTP Basic Auth.

## Scope

Enhance the existing `webhook` channel type only. No new channel types. Changes span backend sender, handler validation, and frontend configuration UI.

## Design Decisions

- **Config JSON extension** (not new DB columns): New fields added to the existing `config` JSON field on the Channel model. Consistent with how feishu/dingtalk store their config. No database migration needed.
- **Backward compatible**: Missing new fields default to current behavior (POST + JSON + no auth).

## Webhook Config Structure

Current:
```json
{
  "url": "https://example.com/hook",
  "headers": {"X-Custom": "value"}
}
```

Enhanced:
```json
{
  "url": "https://example.com/hook",
  "method": "POST",
  "content_type": "application/json",
  "headers": {"X-Custom": "value"},
  "auth_type": "none",
  "auth_config": {}
}
```

### New Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `method` | string | `"POST"` | HTTP method: POST or PUT |
| `content_type` | string | `"application/json"` | `application/json` or `application/x-www-form-urlencoded` |
| `auth_type` | string | `"none"` | `none`, `basic`, or `custom` |
| `auth_config` | object | `{}` | Varies by auth_type (see below) |

### Auth Config Shapes

**basic:**
```json
{"username": "app_key", "password": "appsecret"}
```

**custom:**
```json
{"header_name": "Authorization", "header_value": "Bearer xxx"}
```

## Backend Changes

### 1. WebhookConfig Struct (`internal/notifier/webhook.go`)

```go
type WebhookConfig struct {
    URL         string            `json:"url"`
    Method      string            `json:"method"`
    ContentType string            `json:"content_type"`
    Headers     map[string]string `json:"headers"`
    AuthType    string            `json:"auth_type"`
    AuthConfig  json.RawMessage   `json:"auth_config"`
}

type BasicAuthConfig struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type CustomAuthConfig struct {
    HeaderName  string `json:"header_name"`
    HeaderValue string `json:"header_value"`
}
```

### 2. Sender Logic (`internal/notifier/webhook.go`)

Parse config, then:

1. **Method**: Use `config.Method` (default POST). Validate to POST/PUT only.
2. **Content Type**:
   - `application/json` (default): Send `content` as JSON body (existing behavior). If `content` is already valid JSON, send as-is; otherwise wrap in `{"title": title, "content": content}`.
   - `application/x-www-form-urlencoded`: Send `content` string directly as form body via `strings.NewReader()`. The output_template should render to `key1=value1&key2=value2` format.
3. **Auth**:
   - `none`: Skip.
   - `basic`: Parse `auth_config` as `BasicAuthConfig`, call `req.SetBasicAuth(username, password)`.
   - `custom`: Parse `auth_config` as `CustomAuthConfig`, add `header_name: header_value` to request headers.
4. **Headers**: Merge custom headers into request (existing behavior). Custom auth header takes precedence if name conflicts.

### 3. Channel Validation (`internal/models/models.go`)

Add validation for webhook config new fields:
- `method` must be empty, POST, or PUT
- `content_type` must be empty, `application/json`, or `application/x-www-form-urlencoded`
- `auth_type` must be empty, none, basic, or custom
- When `auth_type=basic`, `auth_config` must contain `username` and `password`
- When `auth_type=custom`, `auth_config` must contain `header_name` and `header_value`

### 4. Test Channel (`internal/handlers/config.go`)

Update `TestChannel` to construct request using the new config fields so the test accurately reflects production behavior.

## Frontend Changes

### Channels.tsx — Webhook Config Form

Add conditional fields when channel type is `webhook`:

| Field | Component | Condition |
|-------|-----------|-----------|
| URL | InputText | Always |
| Method | Dropdown (POST/PUT) | Always |
| Content Type | Dropdown (JSON/form-urlencoded) | Always |
| Headers | JSON editor | Always |
| Auth Type | Dropdown (none/basic/custom) | Always |
| Username | InputText | auth_type = basic |
| Password | Password | auth_type = basic |
| Header Name | InputText | auth_type = custom |
| Header Value | InputText | auth_type = custom |

### Type Updates (`frontend/src/types/index.ts`)

Extend Channel config type to include new webhook fields.

## Odin Integration Example

To configure an Odin alert channel:
- URL: `http://open.odin.qihoo.net:8360/alarm/message/open/alarm/send`
- Method: POST
- Content Type: `application/x-www-form-urlencoded`
- Auth Type: basic
- Username: app_key
- Password: appsecret
- Output template: `teams=team1,team2&title={{.alert_name}}&app_content={{.message}}`

## Files to Modify

| File | Change |
|------|--------|
| `internal/notifier/webhook.go` | Parse new config fields, support form-urlencoded body, Basic Auth, custom auth |
| `internal/models/models.go` | Add webhook config validation for new fields |
| `internal/handlers/config.go` | Update TestChannel to use new config fields |
| `frontend/src/pages/Channels.tsx` | Add webhook config form fields with conditional rendering |
| `frontend/src/types/index.ts` | Extend Channel config type |
