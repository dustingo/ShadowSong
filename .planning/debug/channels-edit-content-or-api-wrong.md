---
status: awaiting_human_verify
trigger: "Investigate issue: channels-edit-content-or-api-wrong"
created: 2026-04-10T00:00:00+08:00
updated: 2026-04-10T00:24:00+08:00
---

## Current Focus

hypothesis: The fix is implemented and self-verified; remaining work is confirming in the real UI that editing existing channels shows correct values and saves the expected payloads.
test: Human-check the Push Channels edit flow for Feishu and custom webhook channels in the browser.
expecting: Existing secrets/templates/headers display correctly in the modal and saving preserves working channel config.
next_action: wait for user verification in the running app

## Symptoms

expected: Editing a push channel should load the correct raw channel config into the modal and submit an update payload that preserves or updates the correct backend fields for that channel type.
actual: User reports the edit button content or API is wrong in the Push Channels page.
errors: No stack trace provided; this is a functional edit-flow bug.
reproduction: Open the frontend Push Channels page, click edit on an existing channel, inspect whether the modal fields are populated correctly and whether saving sends the correct data.
started: Reported after other frontend verification issues were found; exact regression point unknown.

## Eliminated

## Evidence

- timestamp: 2026-04-10T00:00:00+08:00
  checked: required context files
  found: `Channels.tsx` loads full channel data via `channelApi.get(record.id)` and binds the returned object directly into AntD form fields under `config.*`.
  implication: The edit modal depends entirely on backend `GetChannel` returning a form-compatible raw config shape.

- timestamp: 2026-04-10T00:00:00+08:00
  checked: backend channel handlers
  found: `ListChannels` masks sensitive config, `GetChannel` returns raw config, and `UpdateChannel` replaces `config` when JSON is present without any type-specific normalization or validation.
  implication: Any frontend/backend config key mismatch will directly surface as broken edit content or corrupted saved config.

- timestamp: 2026-04-10T00:10:00+08:00
  checked: `internal/notifier/notifier.go`
  found: Feishu expects `config.secret`, DingTalk expects `config.secret`, WeCom expects `config.webhook_url`, and custom webhook expects `config.url`, `config.method`, `config.headers` as an object, and `config.template`.
  implication: The current frontend fields `config.sign_key`, `config.body_template`, and textarea-bound `config.headers` are not compatible with the runtime config contract.

- timestamp: 2026-04-10T00:10:00+08:00
  checked: `frontend/src/pages/Channels.tsx`
  found: The form binds Feishu secret to `config.sign_key`, webhook body to `config.body_template`, and webhook headers directly to a textarea despite backend returning an object map.
  implication: Editing existing channels can show blank or malformed values, and saving can replace working backend config with wrong keys or stringified headers.

- timestamp: 2026-04-10T00:24:00+08:00
  checked: `frontend/src/pages/Channels.tsx`
  found: Added bidirectional normalization so edit uses backend config keys, webhook headers are serialized/deserialized for the textarea, and submitted payloads use `secret`, `template`, and object `headers`.
  implication: The modal content and update payload now align with the backend notifier contract.

- timestamp: 2026-04-10T00:24:00+08:00
  checked: `pnpm build`
  found: Frontend production build passed after narrowing the local `channelApi.get()` result to `Channel` in `Channels.tsx`.
  implication: The fix does not introduce TypeScript or bundling regressions in the frontend.

## Resolution

root_cause: The Push Channels form uses config keys and field types that diverge from the backend notifier contract, so raw config from `GetChannel` does not hydrate correctly and `UpdateChannel` persists incompatible JSON for Feishu and custom webhook channels.
fix: Updated `frontend/src/pages/Channels.tsx` to normalize raw backend channel config into form values on edit, convert webhook headers between object and textarea JSON, submit backend-compatible config keys on create/update, and replace incorrect `sign_key`/`body_template` bindings with `secret`/`template`.
verification: `pnpm build` passed in `frontend/`, confirming the updated `Channels.tsx` compiles and bundles successfully.
files_changed: ["frontend/src/pages/Channels.tsx"]
