import React, { useEffect, useRef } from 'react'
import { Row, Col, Card, Statistic, Spin, Typography, Space, Alert, message } from 'antd'
import ReactECharts from 'echarts-for-react'
import { useAlertStore } from '../stores/alertStore'
import { useUserStore } from '../stores/userStore'
import { AlertCard } from '../components/AlertCard'

const { Text } = Typography

export const Dashboard: React.FC = () => {
  const token = useUserStore((state) => state.token)
  const {
    activeAlerts,
    stats,
    loading,
    wsConnected,
    fetchActiveAlerts,
    fetchStats,
    setWsConnected,
    ackAlert,
    quickSilence,
  } = useAlertStore()

  const handleAck = async (alert: any) => {
    try {
      await ackAlert(alert.alert_id, '')
      message.success('已确认')
    } catch (error) {
      message.error('确认失败')
    }
  }

  const handleQuickSilence = async (alert: any) => {
    try {
      await quickSilence(alert.alert_id, 3600)
      message.success('已静默')
    } catch (error) {
      message.error('静默失败')
    }
  }

  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    fetchActiveAlerts()
    fetchStats()

    // WebSocket connection
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let isConnecting = false

    const connectWS = () => {
      // Prevent multiple concurrent connections
      if (isConnecting || wsRef.current?.readyState === WebSocket.OPEN) {
        return
      }
      if (!token) {
        setWsConnected(false)
        return
      }

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws/alerts?token=${encodeURIComponent(token)}`

      try {
        isConnecting = true
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          console.log('WebSocket connected')
          isConnecting = false
          setWsConnected(true)
        }

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            if (data.type === 'new_alert') {
              useAlertStore.getState().addAlert(data.alert)
            } else if (data.type === 'update_alert') {
              useAlertStore.getState().updateAlert(data.alert)
            }
          } catch (e) {
            console.error('Failed to parse WS message:', e)
          }
        }

        ws.onclose = () => {
          console.log('WebSocket disconnected')
          isConnecting = false
          setWsConnected(false)
          // Reconnect after 3 seconds
          reconnectTimer = setTimeout(connectWS, 3000)
        }

        ws.onerror = () => {
          console.error('WebSocket error, will reconnect...')
          isConnecting = false
        }
      } catch (error) {
        console.error('Failed to create WebSocket:', error)
        isConnecting = false
        reconnectTimer = setTimeout(connectWS, 3000)
      }
    }

    connectWS()

    // Poll for updates
    const interval = setInterval(() => {
      fetchActiveAlerts()
      fetchStats()
    }, 10000)

    return () => {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
      if (wsRef.current) {
        wsRef.current.onclose = null // Prevent reconnect on cleanup
        wsRef.current.close()
      }
      clearInterval(interval)
    }
  }, [fetchActiveAlerts, fetchStats, setWsConnected, token])

  const getTrendOption = () => {
    const trendData = stats?.trend || []
    return {
      tooltip: {
        trigger: 'axis',
      },
      xAxis: {
        type: 'category',
        data: trendData.map((t) => t.time),
        boundaryGap: false,
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          data: trendData.map((t) => t.count),
          type: 'line',
          smooth: true,
          areaStyle: {
            color: 'rgba(24, 144, 255, 0.2)',
          },
          lineStyle: {
            color: '#1890ff',
          },
        },
      ],
    }
  }

  const sortedAlerts = [...activeAlerts].sort((a, b) => {
    // P0 first
    if (a.severity === 'P0' && b.severity !== 'P0') return -1
    if (b.severity === 'P0' && a.severity !== 'P0') return 1
    // Then by trigger time
    return new Date(b.trigger_time).getTime() - new Date(a.trigger_time).getTime()
  })

  return (
    <div>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        {/* Connection status */}
        {!wsConnected && (
          <Alert
            message="实时连接已断开"
            description="正在尝试重新连接..."
            type="warning"
            showIcon
          />
        )}

        {/* Stats cards */}
        <Row gutter={16}>
          <Col span={6}>
            <Card>
              <Statistic
                title="活跃告警"
                value={stats?.firing || 0}
                valueStyle={{ color: '#ff4d4f' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="P0 告警"
                value={stats?.by_severity?.P0 || 0}
                valueStyle={{ color: '#ff4d4f', fontWeight: 'bold' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已确认"
                value={stats?.acked || 0}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已静默"
                value={stats?.silenced || 0}
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
        </Row>

        {/* Trend chart */}
        <Card title="24 小时告警趋势">
          <ReactECharts option={getTrendOption()} style={{ height: 250 }} showLoading={loading} />
        </Card>

        {/* Active alerts */}
        <Card
          title={
            <Space>
              <span>活跃告警</span>
              <Text type="secondary">({activeAlerts.length})</Text>
            </Space>
          }
          extra={<a href="/alerts">查看全部</a>}
        >
          {loading ? (
            <div style={{ textAlign: 'center', padding: 40 }}>
              <Spin />
            </div>
          ) : sortedAlerts.length === 0 ? (
            <Alert message="暂无活跃告警" description="系统运行正常" type="success" showIcon />
          ) : (
            <div style={{ maxHeight: 600, overflowY: 'auto' }}>
              {sortedAlerts.map((alert) => (
                <AlertCard
                  key={alert.alert_id}
                  alert={alert}
                  showActions={true}
                  onAck={handleAck}
                  onQuickSilence={handleQuickSilence}
                />
              ))}
            </div>
          )}
        </Card>
      </Space>
    </div>
  )
}
