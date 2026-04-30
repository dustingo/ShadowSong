---
phase: 18-establish-delivery-ledger
phase_number: 18
verified: 2026-04-30
status: open_threats
threats_total: 9
threats_closed: 8
threats_open: 1
asvs_level: unspecified
block_on: unspecified
---

## OPEN_THREATS

**Phase:** 18 — establish-delivery-ledger  
**Closed:** 8/9 | **Open:** 1/9  
**ASVS Level:** unspecified in provided phase artifacts

### Threat Classification
| Threat ID | Category | Disposition | Verification Method |
|-----------|----------|-------------|---------------------|
| T-18-01 | I | mitigate | Verify snapshot fields and tests exclude secrets in model/service code |
| T-18-02 | R | mitigate | Verify append-only attempt model, uniqueness, and tests |
| T-18-03 | T | mitigate | Verify GORM-only parameterized writes and enum validation in service/model layers |
| T-18-04 | R | mitigate | Verify terminal failure branch calls `MarkFailed` and tests persist failed terminal evidence |
| T-18-05 | D | mitigate | Verify bounded retry/log ordering remains intact and tests cover success, fallback, retry exhausted |
| T-18-06 | T | mitigate | Verify snapshots are frozen from final title/content and route/channel identity before send |
| T-18-07 | I | mitigate | Verify `/api/v1/deliveries*` uses `JWTAuth + RequireCapability(CapabilityViewConfig)` and tests cover 401/403/200 |
| T-18-08 | T | mitigate | Verify fixed filter parsing, bounded pagination, and service/GORM-only queries |
| T-18-09 | D | mitigate | Verify minimal list query surface and bounded pagination |

### Closed
| Threat ID | Category | Disposition | Evidence |
|-----------|----------|-------------|----------|
| T-18-01 | I | mitigate | `internal/models/notification_delivery.go:63`, `internal/delivery/service.go:82`, `internal/delivery/service_test.go:56`, `internal/delivery/service_test.go:234` show channel snapshots only keep identity fields and tests assert no `secret`, `api_key`, or `config`. |
| T-18-02 | R | mitigate | `internal/models/notification_delivery.go:167`, `internal/models/notification_delivery.go:186`, `internal/delivery/service.go:172`, `internal/models/notification_delivery_test.go:72` enforce `delivery_id + attempt_number` uniqueness, append-only attempts, and aggregate-only main record updates. |
| T-18-03 | T | mitigate | `internal/delivery/service.go:143`, `internal/delivery/service.go:172`, `internal/delivery/service.go:205`, `internal/delivery/service.go:253`, `internal/delivery/service.go:293` use GORM `Create`/`Save`/`Where`; `internal/delivery/service.go:73`, `internal/delivery/service.go:157`, `internal/delivery/service.go:222`, `internal/models/notification_delivery.go:136`, `internal/models/notification_delivery.go:190` validate mode, trigger kind, status, and attempt result enums. |
| T-18-04 | R | mitigate | `internal/handlers/webhook.go:1144`, `internal/handlers/webhook.go:1239` call `MarkFailed` in the `terminal_failure` branch; `internal/handlers/webhook_test.go:1011`, `internal/handlers/webhook_test.go:1073` assert failed terminal ledger rows and persisted `final_failure_summary`. |
| T-18-05 | D | mitigate | `internal/handlers/webhook.go:35`, `internal/handlers/webhook.go:1123`, `internal/handlers/webhook.go:1130`, `internal/handlers/webhook.go:1146` preserve 3-attempt retry bounds and canonical log stages while writing the ledger; tests at `internal/handlers/webhook_test.go:693`, `internal/handlers/webhook_test.go:826`, `internal/handlers/webhook_test.go:1011` cover success, fallback, and retry exhausted. |
| T-18-06 | T | mitigate | `internal/handlers/webhook.go:1075`, `internal/handlers/webhook.go:1085`, `internal/handlers/webhook.go:1100`, `internal/handlers/webhook.go:1115` create the delivery only after final render or default fallback is decided; `internal/delivery/service.go:108` freezes rendered title/content and route/channel identity into snapshots; fallback assertions at `internal/handlers/webhook_test.go:880` confirm the stored mode and payload reflect the actual send path. |
| T-18-08 | T | mitigate | `internal/handlers/delivery.go:131` only parses `alert_id`, `trace_id`, `channel_id`, `delivery_status`, `created_from`, `created_to`, `limit`, `offset`; `internal/handlers/delivery.go:167`, `internal/handlers/delivery.go:180` bound pagination; `internal/handlers/delivery.go:87`, `internal/delivery/service.go:293` route all queries through service/GORM with parameterized `Where`. |
| T-18-09 | D | mitigate | `internal/handlers/delivery.go:16`, `internal/handlers/delivery.go:131`, `internal/delivery/service.go:318`, `internal/delivery/service.go:323` keep the list API to bounded pagination plus a minimal fixed filter set, with no unbounded search or aggregate expansion. |

### Open
| Threat ID | Category | Mitigation Expected | Files Searched |
|-----------|----------|---------------------|----------------|
| T-18-07 | I | `/api/v1/deliveries*` must sit behind `JWTAuth + RequireCapability(CapabilityViewConfig)` and tests must cover `401/403/200` for that delivery route. The route protection exists at `internal/router/router.go:156`, and `401/200` are covered at `internal/handlers/delivery_test.go:197` and `internal/handlers/delivery_test.go:216`, but the only `403` assertion uses a separate synthetic route guarded by `CapabilityManageUsers` at `internal/handlers/delivery_test.go:192` rather than the real `/api/v1/deliveries` capability chain. | `internal/router/router.go`, `internal/handlers/delivery_test.go`, `internal/middleware/authorize.go`, `internal/authz/capabilities.go` |

### Unregistered Flags
none

### Threat Flags Scan
No `## Threat Flags` section was present in `18-01-SUMMARY.md`, `18-02-SUMMARY.md`, or `18-03-SUMMARY.md`.

### Accepted Risks Log
none

### Transfer Documentation
none
