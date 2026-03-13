import React from 'react'
import { Tag } from 'antd'

interface SeverityBadgeProps {
  severity: string
}

const severityConfig: Record<string, { color: string; label: string }> = {
  P0: { color: 'red', label: '🔴 P0 紧急' },
  P1: { color: 'orange', label: '🟠 P1 严重' },
  P2: { color: 'gold', label: '🟡 P2 警告' },
  P3: { color: 'green', label: '🟢 P3 提示' },
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  const config = severityConfig[severity] || { color: 'default', label: severity }

  return (
    <Tag color={config.color}>
      {config.label}
    </Tag>
  )
}
