# AI Alert Verification Design

## Overview

Design an AI-powered alert verification system that validates whether P0/P1 alerts represent real issues or are false positives caused by monitoring infrastructure problems (probe failures, collection pipeline issues, etc.). Verification runs asynchronously after notification, and results are attached as labels to alerts — no separate notification is sent.

## Architecture

### Hybrid Two-Phase Model

```
Alert → [Verification Dispatcher] → (P0/P1 + not silenced) → Redis Stream: alerts:verify
                                        │
                                        ▼
                               [Phase 1: Quick Screening]
                                ├─ Knowledge Base Match (RAG)
                                ├─ Monitoring Chain Health Check
                                └─ Maintenance Window Match
                                        │
                           Hit → Label + Done    Miss → Phase 2
                                        │
                                        ▼
                               [Phase 2: Agent Deep Verification]
                                ├─ LLM Agent (external API)
                                ├─ Tool: search_knowledge_base
                                ├─ Tool: probe_icmp
                                ├─ Tool: probe_http
                                ├─ Tool: query_prometheus
                                ├─ Tool: query_logs
                                └─ Tool: check_monitoring_chain
                                        │
                                        ▼
                               Label + WebSocket Push
```

### Key Design Decisions

- **Decoupled from notification**: Notifications proceed without waiting for verification. Verification adds supplementary information.
- **Two-phase separation**: Phase 1 is fast (<2s), Phase 2 is slower (5-30s). Each is independent.
- **Read-only tool whitelist**: All probe tools enforce read-only semantics. No write/delete/modify operations allowed.
- **Target restriction**: Probe tools can only access targets declared in Alert.Labels (e.g., `instance`, `service` labels), not arbitrary targets.

## Data Model

### Alert Model Extensions

```go
// New fields on Alert model
VerificationStatus  string     // pending / phase1 / phase2 / verified / suspected_false / unverifiable
VerificationResult  JSON       // Detailed verification result
VerifiedAt          *time.Time // Verification completion time
VerificationTraceID string     // Correlation ID for verification trace
```

### VerificationResult Structure

```go
type VerificationResult struct {
    Phase1Result *Phase1Result `json:"phase1_result,omitempty"`
    Phase2Result *Phase2Result `json:"phase2_result,omitempty"`
    FinalLabel   string        `json:"final_label"`      // verified / suspected_false / unverifiable
    Confidence   float64       `json:"confidence"`       // 0.0-1.0
    Reasoning    string        `json:"reasoning"`        // AI reasoning for auditability
    Evidence     []Evidence    `json:"evidence"`         // Supporting evidence
}

type Phase1Result struct {
    KnowledgeMatch    *KnowledgeMatch `json:"knowledge_match,omitempty"`
    MonitoringChainOK bool            `json:"monitoring_chain_ok"`
    MaintenanceHit    *MaintenanceHit `json:"maintenance_hit,omitempty"`
    SkipPhase2        bool            `json:"skip_phase2"`
    Reason            string          `json:"reason"`
}

type Phase2Result struct {
    LLMModel       string     `json:"llm_model"`
    ToolCalls      []ToolCall `json:"tool_calls"`
    ReasoningSteps []string   `json:"reasoning_steps"`
    TokenUsage     TokenUsage `json:"token_usage"`
}

type Evidence struct {
    Type   string `json:"type"`   // knowledge / probe / metric / log
    Source string `json:"source"`
    Detail string `json:"detail"`
}

type KnowledgeMatch struct {
    DocID      string  `json:"doc_id"`
    Title      string  `json:"title"`
    Similarity float64 `json:"similarity"`
    Category   string  `json:"category"` // incident / known_issue / sop / maintenance
}

type MaintenanceHit struct {
    DocID        string    `json:"doc_id"`
    Title        string    `json:"title"`
    WindowStart  time.Time `json:"window_start"`
    WindowEnd    time.Time `json:"window_end"`
    AffectedScope string   `json:"affected_scope"` // e.g., service names, IPs
}

type TokenUsage struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
}

type ToolCall struct {
    Tool    string `json:"tool"`
    Input   string `json:"input"`
    Output  string `json:"output"`
    Success bool   `json:"success"`
}
```

### Verification State Machine

```
pending → phase1 → (hit) → verified / suspected_false
                  → (miss) → phase2 → verified / suspected_false / unverifiable
```

- `pending`: Awaiting verification
- `phase1`: Phase 1 in progress
- `phase2`: Phase 2 in progress
- `verified`: Confirmed as real alert
- `suspected_false`: Suspected false positive
- `unverifiable`: Cannot verify (timeout, tool unavailable, etc.)

### Knowledge Document Model

```go
type KnowledgeDocument struct {
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Category  string     `json:"category"`  // incident / known_issue / sop / maintenance
    Content   string     `json:"content"`
    Tags      []string   `json:"tags"`
    Embedding []float64  `json:"embedding"` // Stored in vector DB
    Source    string     `json:"source"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    ExpiresAt *time.Time `json:"expires_at"` // For maintenance windows
}
```

## Phase 1: Quick Screening

### Knowledge Base Matching Flow

```
Alert → [Embedding Generation] → [Vector Similarity Search Top-K] → [Threshold Filter] → Match Result
```

1. **Alert vectorization**: Concatenate `AlertName + Severity + Message + Labels`, call Embedding API to generate vector
2. **Vector search**: Search Top-5 similar documents in vector DB, similarity threshold 0.75
3. **Classification matching**:
   - Match `incident` document → label `verified` (known incident pattern, alert is real)
   - Match `known_issue` document → label `suspected_false` (known issue like monitoring chain fault)
   - Match `maintenance` document + alert time within maintenance window → label `suspected_false`
   - Match `sop` document → extract probe suggestions, pass to Phase 2 as reference

### Monitoring Chain Health Check

Executed in parallel with knowledge matching:

1. **Data source probe check**: Query probe/collector status for the alert's DataSource
2. **Collection pipeline latency check**: Compare alert time vs data collection latency
3. **Same-source alert density check**: Abnormally high alert density from same source in short time → possible probe failure

All three checks run in parallel. Any hit marks as `suspected_false` with reason.

### Phase 1 Timeout

- Total timeout: 2 seconds
- Any sub-step timeout does not block; mark that sub-step as `timeout` and continue
- All Phase 1 sub-steps timeout → proceed to Phase 2

### Phase 1 Output

- Hit → label directly, skip Phase 2
- Miss → pass Phase 1 results (including Top-K document summaries, monitoring chain status) as context to Phase 2

## Phase 2: Agent Deep Verification

### LLM Agent Tool Set

All tools enforce **read-only semantic control** — write/delete/modify operations are prohibited:

| Tool | Function | Safety Control |
|------|----------|---------------|
| `search_knowledge_base` | Semantic search knowledge base | Read-only query |
| `probe_icmp` | ICMP ping target | Only registered target network segments |
| `probe_http` | HTTP GET/HEAD request | Only GET/HEAD methods, no POST/PUT/DELETE |
| `query_prometheus` | Execute PromQL query | Only `query`/`query_range` API, no admin API |
| `query_logs` | Search logs | Only search/get, no delete/put |
| `check_monitoring_chain` | Check monitoring collection chain health | Read-only status query |

### Agent Workflow

```
Input: Alert details + Phase 1 results + Available tools
  │
  ▼
LLM plans verification strategy (System Prompt guided)
  │
  ▼
Loop (max 5 rounds):
  │  LLM selects tool → Execute tool → Return result → LLM analyzes
  │  (LLM can exit early if sufficient evidence)
  ▼
LLM synthesizes judgment → Output verification result
```

### System Prompt Core Logic

```
You are an alert authenticity verification assistant. Verify whether the following alert
represents a real issue or is a false positive from the monitoring system.

Verification principles:
1. Prioritize checking monitoring chain health (probe failures, collection latency)
2. Cross-validate: use multiple read-only methods to confirm the described problem exists
3. Multiple independent sources confirming the problem = high confidence
4. Single-source alert with normal probe results = possible false positive

Available tools: [tool list]

Output format:
- final_label: verified / suspected_false / unverifiable
- confidence: 0.0-1.0
- reasoning: reasoning process
- evidence: list of supporting evidence
```

### Safety Controls

1. **Tool-level**: Each tool has built-in read-only restrictions; LLM cannot bypass
2. **Round limit**: Max 5 tool call rounds to prevent infinite loops
3. **Target restriction**: Probe tools can only access targets from Alert.Labels, not arbitrary targets
4. **Timeout control**: Single tool call timeout 10s, total Phase 2 timeout 60s
5. **Cost control**: Max tokens per verification (input 8K + output 2K)

### Phase 2 Failure Handling

- LLM API call failure → label `unverifiable`, record error
- All tool calls timeout → label `unverifiable`
- Round limit reached without conclusion → label `unverifiable` with accumulated evidence

## Knowledge Base Management

### Data Sources

| Source | Category | Import Method | Update Frequency |
|--------|----------|---------------|------------------|
| Historical incident reviews | incident | API import / manual entry | After incident review |
| Known issues list | known_issue | API import / monitoring integration | Real-time |
| Maintenance window schedules | maintenance | API import / calendar integration | On schedule change |
| Operations SOP manuals | sop | Document import / manual entry | On SOP update |
| Monitoring chain health data | monitoring_chain | Auto-collection | Real-time |

### Knowledge Base Management API

```
POST   /api/v1/knowledge/documents     — Create document (auto-generate embedding)
GET    /api/v1/knowledge/documents     — List (filter by category/tag)
GET    /api/v1/knowledge/documents/:id — Get detail
PUT    /api/v1/knowledge/documents/:id — Update document (re-generate embedding)
DELETE /api/v1/knowledge/documents/:id — Delete document
POST   /api/v1/knowledge/search        — Semantic search (for debugging)
POST   /api/v1/knowledge/batch-import  — Batch import
```

### Embedding Generation

- External Embedding API (e.g., text-embedding-3-small)
- Document create/update: async embedding generation
- Alert verification: sync embedding generation (latency requirement)
- Embedding dimension: 1536

### Vector Database: Milvus

- Open-source, cloud-native, large-scale vector search
- Mature Go SDK
- Supports metadata filtering (category, tags) for precision
- Supports hybrid search (vector similarity + scalar filtering)

### Search Strategy

Verification combines scalar filtering with vector search:
1. Filter by `category` (prioritize `incident` and `known_issue`)
2. Filter by `tags` (match alert's source, severity-related tags)
3. Rank by vector similarity

### Document Expiration

- `maintenance` documents: set `ExpiresAt`, auto-exclude after expiry
- `known_issue` documents: manually mark expired after issue resolution
- `incident` documents: no expiry, long-term retention for pattern matching

## System Integration

### Integration Points

1. **webhook.go**: Call `verification.TriggerVerification(alert)` at end of `processAlertNotificationsAsync`
2. **Alert model**: 4 new fields via database migration
3. **WebSocket**: Push label updates on verification completion, reuse existing Hub
4. **RBAC**: Knowledge management requires `manage_config` capability (admin role)
5. **Audit**: Verification operations written to AuditLog

### New Code Structure

```
internal/
  verification/
    ├─ dispatcher.go      — Severity filter + dispatch to Redis stream
    ├─ phase1.go          — Quick screening logic
    ├─ phase2.go          — Agent verification logic
    ├─ tools/
    │   ├─ probe_icmp.go
    │   ├─ probe_http.go
    │   ├─ query_prometheus.go
    │   ├─ query_logs.go
    │   └─ check_chain.go
    ├─ llm_client.go     — LLM API wrapper
    └─ knowledge/
        ├─ service.go    — Knowledge CRUD
        └─ embedding.go  — Embedding generation
  models/
    ├─ alert.go          — Extended with verification fields
    ├─ verification_result.go — New
    └─ knowledge_document.go  — New
  handlers/
    └─ knowledge.go      — Knowledge management API

frontend/src/
  pages/
    Alerts.tsx           — Extended: verification status badges + detail panel
    Knowledge.tsx        — New: knowledge management page
  stores/
    alertStore.ts        — Extended: verification state
```

### Configuration

```
VERIFICATION_ENABLED=true                    — Master switch
VERIFICATION_SEVERITIES=P0,P1               — Severities to verify
VERIFICATION_PHASE1_TIMEOUT=2s              — Phase 1 timeout
VERIFICATION_PHASE2_TIMEOUT=60s             — Phase 2 timeout
VERIFICATION_PHASE2_MAX_ROUNDS=5            — Agent max rounds
VERIFICATION_LLM_MODEL=claude-sonnet-4-6    — LLM model
VERIFICATION_LLM_API_KEY=***                — API Key
VERIFICATION_EMBEDDING_API_URL=***          — Embedding API URL
VERIFICATION_MILVUS_ADDRESS=localhost:19530  — Milvus address
VERIFICATION_SIMILARITY_THRESHOLD=0.75      — Similarity threshold
```

### Frontend Display

- **Verification status badge**: `verified` (green), `suspected_false` (orange), `unverifiable` (gray), `pending` (blue with animation)
- **Verification detail panel**: Click to expand, shows verification process, reasoning, probe results
- **Knowledge management page**: Document CRUD, search testing, import/export

## Error Handling

| Error Scenario | Handling |
|---------------|----------|
| Embedding API timeout/failure | Phase 1 skips knowledge matching, proceed to Phase 2 |
| Vector DB unavailable | Same as above, degrade to Phase 2 |
| Monitoring chain check timeout | Skip that check, don't block |
| LLM API call failure | Label `unverifiable`, record error in VerificationResult |
| Probe tool target unreachable | Record failure evidence, LLM judges with other evidence |
| Single tool timeout (>10s) | Interrupt tool call, return timeout result to LLM |
| Phase 2 total timeout (>60s) | Label `unverifiable` with accumulated evidence |
| All verification fails | Label `unverifiable`, does not affect alert's original flow |

**Core principle**: Verification failure never blocks the alert pipeline. Verification is always supplementary information.

## Observability

### Metrics (Prometheus)

- `verification_dispatched_total` — Alerts dispatched for verification
- `verification_phase1_completed_total{result=verified|suspected_false|skipped}` — Phase 1 result distribution
- `verification_phase2_completed_total{result=verified|suspected_false|unverifiable}` — Phase 2 result distribution
- `verification_duration_seconds{phase=1|2}` — Duration per phase
- `verification_llm_tokens_total{type=input|output}` — LLM token consumption
- `verification_tool_calls_total{tool=...,result=success|failure|timeout}` — Tool call statistics
- `verification_errors_total{component=embedding|milvus|llm|tool}` — Error distribution

### Logging

Each verification records complete VerificationResult JSON with all steps and tool call details.

### Audit

Verification completion writes to AuditLog with operator and result.
