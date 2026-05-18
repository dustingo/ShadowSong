import React, { useEffect, useRef } from 'react'
import { Card } from 'primereact/card'
import { Button } from 'primereact/button'
import { Message } from 'primereact/message'
import { ProgressSpinner } from 'primereact/progressspinner'
import { Chart } from 'primereact/chart'
import { Tag } from 'primereact/tag'
import { useAlertStore } from '../stores/alertStore'
import { useUserStore } from '../stores/userStore'
import { AlertCard } from '../components/AlertCard'
import { StatisticCard, useToast } from '../components'
import type { Alert as AlertItem, GroupedActiveAlert } from '../types'

export const Dashboard: React.FC = () => {
  const token = useUserStore((state) => state.token)
  const toast = useToast()
  const {
    groupedActiveAlerts,
    groupedActiveLoading,
    stats,
    wsConnected,
    fetchGroupedActiveAlerts,
    fetchStats,
    setWsConnected,
    ackAlert,
    quickSilence,
  } = useAlertStore()

  const handleAck = async (alert: AlertItem) => {
    try {
      await ackAlert(alert.alert_id, '')
      toast.showSuccess('已确认')
    } catch {
      toast.showError('确认失败')
    }
  }

  const handleQuickSilence = async (alert: AlertItem) => {
    try {
      await quickSilence(alert.alert_id, 3600)
      toast.showSuccess('已静默')
    } catch {
      toast.showError('静默失败')
    }
  }

  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    fetchGroupedActiveAlerts()
    fetchStats()

    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let isConnecting = false

    const connectWS = () => {
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
          } catch {
            // Ignore parse errors
          }
        }

        ws.onclose = () => {
          isConnecting = false
          setWsConnected(false)
          reconnectTimer = setTimeout(connectWS, 3000)
        }

        ws.onerror = () => {
          isConnecting = false
        }
      } catch {
        isConnecting = false
        reconnectTimer = setTimeout(connectWS, 3000)
      }
    }

    connectWS()

    const interval = setInterval(() => {
      fetchGroupedActiveAlerts()
      fetchStats()
    }, 10000)

    return () => {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
      if (wsRef.current) {
        wsRef.current.onclose = null
        wsRef.current.close()
      }
      clearInterval(interval)
    }
  }, [fetchGroupedActiveAlerts, fetchStats, setWsConnected, token])

  const chartData = {
    labels: (stats?.trend || []).map((t) => t.time),
    datasets: [
      {
        label: '告警数量',
        data: (stats?.trend || []).map((t) => t.count),
        fill: true,
        backgroundColor: 'rgba(16, 185, 129, 0.2)',
        borderColor: '#10B981',
        tension: 0.4,
      },
    ],
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: { color: '#E2E8F0' },
      },
      x: {
        grid: { display: false },
      },
    },
  }

  const sortedGroupedAlerts = [...groupedActiveAlerts].sort((a, b) => {
    if (a.latest_alert.severity === 'P0' && b.latest_alert.severity !== 'P0') return -1
    if (b.latest_alert.severity === 'P0' && a.latest_alert.severity !== 'P0') return 1
    return new Date(b.last_triggered_at).getTime() - new Date(a.last_triggered_at).getTime()
  })

  // Count firing alerts that have been notified but not yet acked
  const pendingAckCount = groupedActiveAlerts.filter((g) =>
    g.latest_alert.status === 'firing'
    && !g.latest_alert.acked_at
    && g.latest_alert.notify_count > 0
  ).length

  const statsCards = [
    { label: '活跃告警', value: stats?.firing || 0, color: 'var(--danger-color)', icon: 'pi pi-bell' },
    { label: 'P0 告警', value: stats?.by_severity?.P0 || 0, color: 'var(--danger-color)', icon: 'pi pi-exclamation-triangle' },
    { label: '待确认告警', value: pendingAckCount, color: 'var(--info-color, #3B82F6)', icon: 'pi pi-clock' },
    { label: '已确认', value: stats?.acked || 0, color: 'var(--success-color)', icon: 'pi pi-check-circle' },
    { label: '已静默', value: stats?.silenced || 0, color: 'var(--warning-color)', icon: 'pi pi-volume-off' },
  ]

  return (
    <div className="flex flex-column gap-4">
      {!wsConnected && (
        <Message
          severity="warn"
          text="实时连接已断开，正在尝试重新连接..."
        />
      )}

      {/* Stats cards */}
      <div className="grid">
        {statsCards.map((stat, index) => (
          <div key={index} className="col-12 md:col-6 lg:col-3">
            <StatisticCard {...stat} />
          </div>
        ))}
      </div>

      {/* Chart */}
      <Card title="24 小时告警趋势" className="shadow-sm">
        <div style={{ height: '250px' }}>
          <Chart type="line" data={chartData} options={chartOptions} />
        </div>
      </Card>

      {/* Active alerts */}
      <Card className="shadow-sm">
        <div className="flex align-items-center justify-content-between mb-3">
          <span className="text-xl font-semibold" style={{ color: 'var(--text-primary)' }}>
            活跃告警 ({groupedActiveAlerts.length})
          </span>
          <Button
            label="查看全部"
            link
            style={{ color: 'var(--primary-color)' }}
            onClick={() => window.location.href = '/alerts'}
          />
        </div>
        {groupedActiveLoading ? (
          <div className="flex justify-content-center p-4">
            <ProgressSpinner />
          </div>
        ) : sortedGroupedAlerts.length === 0 ? (
          <Message severity="success" text="暂无活跃告警，系统运行正常" />
        ) : (
          <div style={{ maxHeight: '600px', overflowY: 'auto' }}>
            {sortedGroupedAlerts.map((grouped) => (
              <div key={grouped.fingerprint} className="mb-3">
                <AlertCard
                  alert={grouped.latest_alert}
                  showActions={true}
                  onAck={handleAck}
                  onQuickSilence={handleQuickSilence}
                />
                {grouped.count > 1 && (
                  <div className="mt-1 ml-4">
                    <Tag
                      value={`共 ${grouped.count} 次`}
                      style={{
                        background: 'var(--warning-light-color)',
                        color: 'var(--warning-color)',
                        border: '1px solid var(--warning-color)',
                      }}
                    />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
