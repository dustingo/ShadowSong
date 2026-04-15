---
status: awaiting_human_verify
trigger: "Investigate issue: route-rules-source-options-wrong"
created: 2026-04-10T00:00:00+08:00
updated: 2026-04-10T00:08:00+08:00
---

## Current Focus

hypothesis: Confirmed: Route Rules page bound the `sources` selector to `channels` because it never loaded or consumed `dataSources`.
test: User verifies the Route Rules modal now shows datasource options in “匹配来源” and channel options in “目标渠道”.
expecting: The original functional mismatch is gone in the real UI flow.
next_action: wait for human verification in the browser

## Symptoms

expected: In the Route Rules modal, the `sources` / “匹配来源” selector should list data sources from `/datasources`.
actual: The selector is populated from the channel list, so it shows push channels instead of data sources.
errors: No runtime stack trace reported; this is a functional data-binding bug.
reproduction: Open the frontend Route Rules page, create or edit a rule, inspect the “匹配来源” selector. It currently renders `channels.map(...)`.
started: Reported now during manual frontend verification; exact regression point unknown.

## Eliminated

## Evidence

- timestamp: 2026-04-10T00:00:00+08:00
  checked: frontend/src/pages/RouteRules.tsx
  found: The page destructures `channels` and `fetchChannels` from the config store, calls `fetchChannels()` in `useEffect`, and uses `channels.map(...)` for both the `sources` selector and the `channel_ids` selector.
  implication: The `sources` selector cannot show datasources because the page never loads or references `dataSources`.

- timestamp: 2026-04-10T00:00:00+08:00
  checked: frontend/src/stores/configStore.ts and frontend/src/api/client.ts
  found: The config store already exposes `dataSources` state and `fetchDataSources()`, backed by `dataSourceApi.list()` against `/datasources`.
  implication: The bug is isolated to Route Rules page wiring, not missing API/store support.

- timestamp: 2026-04-10T00:05:00+08:00
  checked: frontend/src/pages/RouteRules.tsx
  found: The page now destructures `dataSources` and `fetchDataSources`, calls `fetchDataSources()` on mount, and renders the `sources` selector from `dataSources` while keeping `channel_ids` rendered from `channels`.
  implication: The selector wiring now matches the intended datasource/channel split.

- timestamp: 2026-04-10T00:08:00+08:00
  checked: frontend build
  found: `pnpm build` completed successfully in `frontend/`, including `tsc` and Vite production build.
  implication: The page change is type-safe and does not break the frontend build.

## Resolution

root_cause: `frontend/src/pages/RouteRules.tsx` only loaded `channels` and reused `channels.map(...)` for both selectors, so the `sources` field was bound to push-channel data instead of datasource data.
fix: Updated `RouteRules.tsx` to load `dataSources` via `fetchDataSources()` and render the `sources` selector from `dataSources`, while leaving `channel_ids` bound to `channels`.
verification: Confirmed by code inspection that the selector data sources are split correctly, and `pnpm build` passed in `frontend/`.
files_changed: [frontend/src/pages/RouteRules.tsx]
