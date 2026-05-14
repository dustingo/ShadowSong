import React from 'react'
import { Tag } from 'primereact/tag'

interface SeverityBadgeProps {
  severity: string
}

const severityConfig: Record<string, { severity: 'danger' | 'warning' | 'success' | 'info' | 'secondary' | 'contrast', label: string }> = {
  P0: { severity: 'danger', label: '🔴 P0 紧急' },
  P1: { severity: 'warning', label: '🟠 P1 严重' },
  P2: { severity: 'info', label: '🟡 P2 警告' },
  P3: { severity: 'success', label: '🟢 P3 提示' },
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  const config = severityConfig[severity] || { severity: 'secondary', label: severity }
  return <Tag value={config.label} severity={config.severity} />
}