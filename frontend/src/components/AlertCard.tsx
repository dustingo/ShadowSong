import React from 'react'
import { Tag, Typography, Space, Button, Tooltip } from 'antd'
import { CheckOutlined } from '@ant-design/icons'
import type { Alert } from '../types'
import { SeverityBadge } from './SeverityBadge'
import dayjs from 'dayjs'

const { Text } = Typography

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

  const handleAck = () => {
    onAck?.(alert)
  }

  const handleQuickSilence = () => {
    onQuickSilence?.(alert)
  }

  return (
    <div
      style={{
        background: '#fff',
        borderRadius: 8,
        padding: 16,
        marginBottom: 12,
        border: isP0 && isActive ? '2px solid #ff4d4f' : '1px solid #f0f0f0',
        boxShadow: isP0 && isActive ? '0 0 8px rgba(255, 77, 79, 0.3)' : 'none',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <Space direction="vertical" size={4} style={{ flex: 1 }}>
          <Space align="center">
            <SeverityBadge severity={alert.severity} />
            <Text strong style={{ fontSize: 16 }}>{alert.alert_name}</Text>
            <Tag color="blue">{alert.source}</Tag>
            {alert.trigger_count > 1 && (
              <Tooltip title="触发次数">
                <Tag color="orange">x{alert.trigger_count}</Tag>
              </Tooltip>
            )}
          </Space>
          <Text type="secondary">{alert.message}</Text>
          <Space size="large">
            <Text type="secondary" style={{ fontSize: 12 }}>
              触发时间: {dayjs(alert.trigger_time).format('YYYY-MM-DD HH:mm:ss')}
            </Text>
            {alert.labels && Object.keys(alert.labels).length > 0 && (
              <Space size={4}>
                {Object.entries(alert.labels).slice(0, 3).map(([key, value]) => (
                  <Tag key={key} style={{ margin: 0 }}>{key}: {String(value)}</Tag>
                ))}
                {Object.keys(alert.labels).length > 3 && (
                  <Text type="secondary">+{Object.keys(alert.labels).length - 3}</Text>
                )}
              </Space>
            )}
          </Space>
        </Space>
        {showActions && isActive && (
          <Space>
            <Button
              type="primary"
              size="small"
              icon={<CheckOutlined />}
              onClick={handleAck}
            >
              确认
            </Button>
            <Button
              size="small"
              onClick={handleQuickSilence}
            >
              静默
            </Button>
          </Space>
        )}
      </div>

      {alert.acked_by && (
        <div style={{ marginTop: 8 }}>
          <Text type="secondary">
            已由 {alert.acked_by} 于 {dayjs(alert.acked_at).format('MM-DD HH:mm')} 确认
            {alert.ack_comment && `: ${alert.ack_comment}`}
          </Text>
        </div>
      )}
    </div>
  )
}
