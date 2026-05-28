// Common types for the application

export type JsonValue = string | number | boolean | null | JsonObject | JsonValue[]

export interface JsonObject {
  [key: string]: JsonValue
}

export interface Alert {
  alert_id: string
  source: string
  alert_name: string
  severity: 'P0' | 'P1' | 'P2' | 'P3'
  message: string
  labels: Record<string, string>
  fingerprint: string
  trigger_time: string
  received_at: string
  status: 'pending' | 'firing' | 'acked' | 'silenced' | 'resolved' | 'deduplicated'
  raw: JsonValue
  acked_by?: string
  acked_at?: string
  ack_comment?: string
  trigger_count: number
  last_notified_at: string | null
  notify_count: number
  created_at: string
  updated_at: string
}

export interface GroupedActiveAlert {
  fingerprint: string
  latest_alert: Alert
  count: number
  first_triggered_at: string
  last_triggered_at: string
}

export interface DataSource {
  id: number
  name: string
  display_name: string
  api_key?: string

  // 去重/聚合配置
  deduplicate_enabled?: boolean
  deduplicate_window?: number // 秒
  group_enabled?: boolean
  group_window?: number // 秒

  input_template: string
  output_template: string
  group_by_labels: string[]
  enabled: boolean
  last_trigger_at?: string
  created_at: string
  updated_at: string
}

export interface DataSourcePreviewRequest {
  datasource_id?: number
  source_name?: string
  input_template?: string
  output_template?: string
  sample_payload: JsonObject | JsonObject[]
}

export interface DataSourcePreviewResponse {
  normalized_alert: Alert
  rendered: {
    title: string
    content: string
  }
  context_preview: {
    top_level_keys: string[]
    event_keys: string[]
    label_keys: string[]
    alert_keys: string[]
  }
}

export interface WebhookAuthConfig {
  username?: string
  password?: string
  header_name?: string
  header_value?: string
}

export interface Channel {
  id: number
  name: string
  type: 'feishu' | 'dingtalk' | 'wecom' | 'webhook' | 'email'
  config: JsonObject & {
    webhook_url?: string
    secret?: string
    url?: string
    method?: string
    content_type?: string
    headers?: Record<string, string> | string
    template?: string
    auth_type?: string
    auth_config?: WebhookAuthConfig
    from_name?: string
    rate_limit?: number
  }
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface SmtpConfig {
  id?: number
  host: string
  port: number
  username: string
  password: string
  from_addr: string
  from_name: string
  tls: boolean
  enabled: boolean
  updated_at: string
}

export interface RouteRule {
  id: number
  name: string
  priority: number
  severities: string[]
  sources: string[]
  label_matchers: LabelMatcher[]
  channel_ids: number[]
  recipients: string[]
  time_ranges: TimeRange[]
  enabled: boolean
  escalation_enabled: boolean
  escalation_timeout: number
  escalation_max_repeats: number
  created_at: string
  updated_at: string
}

export interface LabelMatcher {
  key: string
  pattern: string
}

export interface TimeRange {
  start_time: string
  end_time: string
}

export interface SilenceRule {
  id: number
  name: string
  comment: string
  source?: string
  alert_name_pattern?: string
  severities: string[]
  label_matchers: LabelMatcher[]
  starts_at: string
  ends_at: string
  enabled?: boolean
  created_by: string
  created_at: string
  updated_at: string
}

export interface User {
  id: number
  username: string
  name: string
  email?: string
  role: 'admin' | 'operator' | 'viewer'
  disabled_at?: string
  force_password_reset?: boolean
  created_at: string
  updated_at: string
}

export interface FinalFailureSummary {
  retryable: boolean
  result: string
  error_message: string
  http_status?: number
  attempt_count: number
  trigger_kind: string
}

export interface DeliveryAttempt {
  id: number
  attempt_number: number
  result: string
  retryable: boolean
  error_message: string
  http_status?: number
  duration_ms: number
  trigger_kind: string
  created_at: string
}

export interface AlertSnapshot {
  alert_id: string
  trace_id?: string
  source: string
  alert_name: string
  severity: string
  message: string
  trigger_time?: string
  fingerprint: string
  status: string
  labels?: JsonObject
}

export interface ChannelSnapshot {
  id: number
  name: string
  type: string
  enabled: boolean
}

export interface RouteSnapshot {
  id: number
  name: string
  priority: number
  enabled: boolean
  channel_ids?: number[]
}

export interface RenderedPayloadSnapshot {
  title: string
  content: string
}

export interface Delivery {
  id: number
  alert_id: string
  trace_id: string
  channel_id: number
  route_rule_id?: number
  delivery_status: string
  delivery_mode: string
  attempt_count: number
  final_failure_summary?: FinalFailureSummary
  alert_snapshot: AlertSnapshot
  channel_snapshot: ChannelSnapshot
  route_snapshot?: RouteSnapshot
  rendered_payload_snapshot: RenderedPayloadSnapshot
  last_attempt_at?: string
  last_success_at?: string
  created_at: string
  updated_at: string
  attempts: DeliveryAttempt[]
}

export interface DeliveryListResponse {
  list: Delivery[]
  total: number
}

export interface DeliveryFilters {
  alert_id?: string
  trace_id?: string
  channel_id?: number
  delivery_status?: string
  created_from?: string
  created_to?: string
  limit?: number
  offset?: number
}

export interface DeliveryRecoveryRequest {
  reason: string
}

export interface DeliveryRecoveryResult {
  recovery_id: number
  action: 'retry' | 'replay'
  status: 'succeeded' | 'failed' | 'rejected'
  original_delivery_id: number
  result_delivery_id?: number
  error_message?: string
}

export interface AuditLog {
  id: number
  actor_user_id: number
  actor_username: string
  actor_role: string
  action: string
  target_type: string
  target_id: string
  result: string
  detail: string
  created_at: string
}

export interface AuditLogListResponse {
  items: AuditLog[]
  total: number
  page: number
  page_size: number
}

export interface BatchAckRequest {
  alert_ids: string[]
  comment: string
}

export interface BatchSilenceRequest {
  alert_ids: string[]
  duration: number
}

export interface BatchResult {
  updated: number
  skipped: number
  errors: string[]
}