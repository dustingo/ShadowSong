import React from 'react'
import { Card } from 'primereact/card'

interface StatisticCardProps {
  label: string
  value: number | string
  icon: string
  color: string
  suffix?: string
}

export const StatisticCard: React.FC<StatisticCardProps> = ({
  label,
  value,
  icon,
  color,
  suffix,
}) => {
  return (
    <Card
      className="shadow-sm"
      style={{
        border: '1px solid var(--surface-border)',
        borderRadius: '8px',
      }}
    >
      <div className="flex align-items-center justify-content-between">
        <div>
          <div
            className="text-sm mb-1"
            style={{ color: 'var(--text-secondary)' }}
          >
            {label}
          </div>
          <div
            className="text-3xl font-bold"
            style={{ color }}
          >
            {value}
            {suffix && (
              <span className="text-lg ml-1" style={{ color: 'var(--text-secondary)' }}>
                {suffix}
              </span>
            )}
          </div>
        </div>
        <div
          className="flex align-items-center justify-content-center"
          style={{
            width: '48px',
            height: '48px',
            borderRadius: '12px',
            background: `${color}15`,
          }}
        >
          <i className={`${icon} text-xl`} style={{ color }} />
        </div>
      </div>
    </Card>
  )
}