// Common types for the application

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
  raw: any
  acked_by?: string
  acked_at?: string
  ack_comment?: string
  trigger_count: number
  created_at: string
  updated_at: string
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
  sample_payload: Record<string, any> | Array<Record<string, any>>
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

export interface Channel {
  id: number
  name: string
  type: 'feishu' | 'dingtalk' | 'wecom' | 'webhook'
  config: any
  enabled: boolean
  created_at: string
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
  time_ranges: TimeRange[]
  enabled: boolean
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
  created_by: string
  created_at: string
  updated_at: string
}

export interface OnDuty {
  id: number
  user_id: string
  user_name: string
  channel_id: number
  start_time: string
  end_time: string
  created_at: string
  updated_at: string
}

export interface User {
  id: number
  username: string
  name: string
  email?: string
  role: 'admin' | 'operator' | 'viewer'
  created_at: string
  updated_at: string
}
