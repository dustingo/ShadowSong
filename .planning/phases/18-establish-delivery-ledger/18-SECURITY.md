---
phase: 18-establish-delivery-ledger
phase_number: 18
verified: 2026-04-30
status: secured
threats_total: 9
threats_closed: 9
threats_open: 0
asvs_level: unspecified
block_on: unspecified
---

## SECURED

**Phase:** 18 — establish-delivery-ledger  
**Threats Closed:** 9/9  
**ASVS Level:** unspecified in provided phase artifacts

### Threat Verification
| Threat ID | Category | Disposition | Evidence |
|-----------|----------|-------------|----------|
| T-18-01 | I | mitigate | `internal/models/notification_delivery.go:63`, `internal/models/notification_delivery.go:80`, `internal/delivery/service.go:93`, `internal/delivery/service.go:121`, `internal/delivery/service_test.go:61`, `internal/delivery/service_test.go:251` show only non-secret channel identity plus rendered payload are snapshotted, and tests assert no `secret`, `api_key`, or `config` leak into persisted snapshots. |
| T-18-02 | R | mitigate | `internal/models/notification_delivery.go:167`, `internal/models/notification_delivery.go:168`, `internal/models/notification_delivery.go:186`, `internal/models/notification_delivery_test.go:68`, `internal/delivery/service.go:172`, `internal/delivery/service.go:205`, `internal/delivery/service.go:253` enforce `delivery_id + attempt_number` uniqueness, append-only attempts, and aggregate-only terminal updates on the main record. |
| T-18-03 | T | mitigate | `internal/delivery/service.go:143`, `internal/delivery/service.go:172`, `internal/delivery/service.go:205`, `internal/delivery/service.go:253`, `internal/delivery/service.go:280` use GORM `Create`/`Save`/`Where` paths rather than dynamic SQL, while `internal/models/notification_delivery.go:136` and `internal/models/notification_delivery.go:190` validate delivery and attempt enums. |
| T-18-04 | R | mitigate | `internal/handlers/webhook.go:1146`, `internal/handlers/webhook.go:1239` persist failed terminal state through `MarkFailed` on the retry-exhausted branch, and `internal/handlers/webhook_test.go:1050`, `internal/handlers/webhook_test.go:1078` verify terminal failure logs and ledger evidence coexist. |
| T-18-05 | D | mitigate | `internal/handlers/webhook.go:35`, `internal/handlers/webhook.go:1130`, `internal/handlers/webhook.go:1146` preserve the 3-attempt retry budget and canonical `send_attempt` / `terminal_failure` logging, while `internal/handlers/webhook_test.go:713`, `internal/handlers/webhook_test.go:886`, `internal/handlers/webhook_test.go:1050` cover success, fallback, and retry exhaustion. |
| T-18-06 | T | mitigate | `internal/handlers/webhook.go:1096`, `internal/handlers/webhook.go:1103`, `internal/handlers/webhook.go:1106`, `internal/delivery/service.go:121` show deliveries are started only after final rendered or fallback payload selection, and `internal/handlers/webhook_test.go:846`, `internal/handlers/webhook_test.go:886` confirm ledger snapshots match the actual send mode and payload. |
| T-18-07 | I | mitigate | The real route is protected by `internal/router/router.go:155` and `internal/router/router.go:156` with `JWTAuth + RequireCapability(authz.CapabilityViewConfig)`. `internal/handlers/delivery_test.go:172` builds the real `/api/v1/deliveries` chain, `internal/handlers/delivery_test.go:205` proves `401` without a token, `internal/handlers/delivery_test.go:215` proves `403` on the real route for a supported role registered without `view_config`, and `internal/handlers/delivery_test.go:225` plus `internal/handlers/delivery_test.go:233` prove `200` for authorized list and detail requests. `go test ./internal/handlers ./internal/router -run "Test(DeliveryHandler|Router.*Deliveries.*|TestRouterDeliveriesAuthorization)" -count=1` passed on 2026-04-30. |
| T-18-08 | T | mitigate | `internal/handlers/delivery.go:131`, `internal/handlers/delivery.go:168`, `internal/handlers/delivery.go:180`, `internal/handlers/delivery.go:87`, `internal/delivery/service.go:280` restrict filters to a fixed query set, bound `limit`/`offset`, and route execution through service/GORM queries. |
| T-18-09 | D | mitigate | `internal/handlers/delivery.go:81`, `internal/handlers/delivery.go:131`, `internal/delivery/service.go:280`, `internal/handlers/delivery_test.go:80` keep the list surface to bounded pagination with minimal fixed filters and regression-test invalid wide-query inputs. |

### Unregistered Flags
none

### Threat Flags Scan
No `## Threat Flags` section was present in `18-01-SUMMARY.md`, `18-02-SUMMARY.md`, or `18-03-SUMMARY.md`.

### Accepted Risks Log
none

### Transfer Documentation
none
