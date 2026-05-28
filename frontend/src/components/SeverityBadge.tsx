import React from 'react'
import { Tag } from 'primereact/tag'

interface SeverityBadgeProps {
  severity: string
}

// 告警严重程度配置 - 使用自定义配色
const severityConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  P0: { label: 'P0 紧急', color: 'var(--danger-color)', bgColor: 'var(--danger-light-color)' },
  P1: { label: 'P1 严重', color: '#F97316', bgColor: '#FFF7ED' },
  P2: { label: 'P2 警告', color: 'var(--warning-color)', bgColor: 'var(--warning-light-color)' },
  P3: { label: 'P3 提示', color: 'var(--success-color)', bgColor: 'var(--success-light-color)' },
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  const config = severityConfig[severity]

  if (!config) {
    return <Tag value={severity} severity="secondary" />
  }

  return (
    <Tag
      value={config.label}
      style={{
        background: config.bgColor,
        color: config.color,
        border: `1px solid ${config.color}`,
        fontWeight: 500,
      }}
    />
  )
}

interface StatusBadgeProps {
  status: string
}

const statusConfig: Record<string, { label: string; severity: 'danger' | 'warning' | 'secondary' | 'success' | 'info' }> = {
  firing: { label: '告警中', severity: 'danger' },
  acked: { label: '已确认', severity: 'warning' },
  silenced: { label: '已静默', severity: 'secondary' },
  resolved: { label: '已恢复', severity: 'success' },
  pending: { label: '待处理', severity: 'info' },
  deduplicated: { label: '已去重', severity: 'secondary' },
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const config = statusConfig[status]
  if (!config) {
    return <Tag value={status} severity="secondary" />
  }
  return <Tag value={config.label} severity={config.severity} />
}
