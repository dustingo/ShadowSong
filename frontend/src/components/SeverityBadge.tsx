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