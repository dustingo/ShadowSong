import React from 'react'
import { Tag } from 'primereact/tag'
import { Button } from 'primereact/button'
import { SeverityBadge } from './SeverityBadge'
import dayjs from 'dayjs'
import type { Alert } from '../types'

interface AlertCardProps {
  alert: Alert
  onAck?: (alert: Alert) => void
  onQuickSilence?: (alert: Alert) => void
  showActions?: boolean
}

export const AlertCard: React.FC<AlertCardProps> = ({
  alert,
  onAck,
  onQuickSilence,
  showActions = true,
}) => {
  const isP0 = alert.severity === 'P0'
  const isActive = alert.status === 'firing'

  return (
    <div
      className="mb-3 p-4"
      style={{
        background: 'var(--surface-card)',
        borderRadius: '8px',
        border: isP0 && isActive ? '2px solid var(--danger-color)' : '1px solid var(--surface-border)',
        boxShadow: isP0 && isActive ? '0 0 8px rgba(239, 68, 68, 0.3)' : 'none',
      }}
    >
      <div className="flex justify-content-between align-items-start">
        <div className="flex flex-column gap-2 flex-1">
          <div className="flex align-items-center gap-2 flex-wrap">
            <SeverityBadge severity={alert.severity} />
            <span className="font-semibold" style={{ color: 'var(--text-primary)' }}>{alert.alert_name}</span>
            <Tag
              value={alert.source}
              style={{
                background: 'var(--surface-hover)',
                color: 'var(--text-secondary)',
                border: '1px solid var(--surface-border)',
              }}
            />
            {alert.trigger_count > 1 && (
              <Tag
                value={`x${alert.trigger_count}`}
                style={{
                  background: 'var(--warning-light-color)',
                  color: 'var(--warning-color)',
                  border: '1px solid var(--warning-color)',
                }}
              />
            )}
          </div>
          <p className="m-0" style={{ color: 'var(--text-secondary)' }}>{alert.message}</p>
          <div className="flex gap-4 text-sm" style={{ color: 'var(--text-disabled)' }}>
            <span>触发时间: {dayjs(alert.trigger_time).format('YYYY-MM-DD HH:mm:ss')}</span>
            {alert.labels && Object.keys(alert.labels).length > 0 && (
              <div className="flex gap-1">
                {Object.entries(alert.labels).slice(0, 3).map(([key, value]) => (
                  <Tag
                    key={key}
                    value={`${key}: ${String(value)}`}
                    style={{
                      background: 'var(--surface-hover)',
                      color: 'var(--text-secondary)',
                      fontSize: '0.75rem',
                    }}
                  />
                ))}
                {Object.keys(alert.labels).length > 3 && (
                  <span style={{ color: 'var(--text-disabled)' }}>+{Object.keys(alert.labels).length - 3}</span>
                )}
              </div>
            )}
          </div>
        </div>

        {showActions && isActive && (
          <div className="flex gap-2">
            <Button
              icon="pi pi-check"
              label="确认"
              size="small"
              style={{ background: 'var(--primary-color)', border: 'none' }}
              onClick={() => onAck?.(alert)}
            />
            <Button
              icon="pi pi-volume-off"
              label="静默"
              size="small"
              style={{
                background: 'var(--warning-light-color)',
                color: 'var(--warning-color)',
                border: '1px solid var(--warning-color)',
              }}
              onClick={() => onQuickSilence?.(alert)}
            />
          </div>
        )}
      </div>

      {alert.acked_by && (
        <div className="mt-2 text-sm" style={{ color: 'var(--text-disabled)' }}>
          已由 {alert.acked_by} 于 {dayjs(alert.acked_at).format('MM-DD HH:mm')} 确认
          {alert.ack_comment && `: ${alert.ack_comment}`}
        </div>
      )}
    </div>
  )
}