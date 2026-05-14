import React from 'react'
import { Card } from 'primereact/card'

interface StatisticCardProps {
  title: string
  value: number | string
  icon: string
  color: string
  suffix?: string
}

export const StatisticCard: React.FC<StatisticCardProps> = ({
  title,
  value,
  icon,
  color,
  suffix,
}) => {
  return (
    <Card className="shadow-sm border-0">
      <div className="flex align-items-center justify-content-between">
        <div>
          <div className="text-sm text-slate-500 mb-1">{title}</div>
          <div className="text-3xl font-bold" style={{ color }}>
            {value}
            {suffix && <span className="text-lg ml-1">{suffix}</span>}
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