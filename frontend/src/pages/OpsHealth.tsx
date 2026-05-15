import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { ProgressSpinner } from 'primereact/progressspinner'
import { Message } from 'primereact/message'
import { StatisticCard } from '../components'
import { metricsApi, MetricsResponse, channelHealthApi, ChannelHealthResponse, channelApi } from '../api/client'
import type { Channel } from '../types'

const OpsHealth: React.FC = () => {
  const [metrics, setMetrics] = useState<MetricsResponse | null>(null)
  const [channelHealth, setChannelHealth] = useState<ChannelHealthResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [metricsRes, channelsRes] = await Promise.all([
          metricsApi.get('24h'),
          channelApi.list(),
        ])

        setMetrics(metricsRes as unknown as MetricsResponse)

        // Fetch health for each channel
        const healthPromises = (channelsRes as unknown as Channel[]).map((ch: Channel) =>
          channelHealthApi.get(ch.id, '24h').then((r) => r as unknown as ChannelHealthResponse).catch(() => null)
        )
        const healthResults = await Promise.all(healthPromises)
        setChannelHealth(healthResults.filter(Boolean) as ChannelHealthResponse[])
      } catch (e) {
        setError('加载运维健康数据失败')
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  if (loading) {
    return (
      <div className="flex justify-content-center align-items-center" style={{ minHeight: '200px' }}>
        <ProgressSpinner />
      </div>
    )
  }
  if (error) return <Message severity="error" text={error} />

  const successRateBodyTemplate = (rowData: ChannelHealthResponse) => {
    return `${(rowData.success_rate * 100).toFixed(1)}%`
  }

  const lastErrorBodyTemplate = (rowData: ChannelHealthResponse) => {
    return rowData.last_failure?.error_message || '-'
  }

  return (
    <div>
      <Card title="系统指标 (24小时)" className="shadow-sm">
        <div className="grid">
          <div className="col-12 md:col-6 lg:col-2">
            <StatisticCard
              label="Webhook 接收"
              value={metrics?.webhook_ingest_total || 0}
              icon="pi pi-sign-in"
              color="var(--primary-color)"
            />
          </div>
          <div className="col-12 md:col-6 lg:col-2">
            <StatisticCard
              label="发送成功"
              value={metrics?.notification_send_success_total || 0}
              icon="pi pi-check-circle"
              color="var(--success-color)"
            />
          </div>
          <div className="col-12 md:col-6 lg:col-2">
            <StatisticCard
              label="发送失败"
              value={metrics?.notification_send_failure_total || 0}
              icon="pi pi-times-circle"
              color="var(--danger-color)"
            />
          </div>
          <div className="col-12 md:col-6 lg:col-2">
            <StatisticCard
              label="重试次数"
              value={metrics?.notification_retry_total || 0}
              icon="pi pi-refresh"
              color="var(--warning-color)"
            />
          </div>
          <div className="col-12 md:col-6 lg:col-2">
            <StatisticCard
              label="终态失败"
              value={metrics?.notification_terminal_failure_total || 0}
              icon="pi pi-exclamation-triangle"
              color="var(--danger-color)"
            />
          </div>
        </div>
      </Card>

      <Card title="渠道健康度 (24小时)" className="mt-4 shadow-sm">
        <DataTable
          value={channelHealth}
          dataKey="channel_id"
        >
          <Column field="channel_name" header="渠道名称" />
          <Column field="total_deliveries" header="投递总数" />
          <Column field="success_rate" header="成功率" body={successRateBodyTemplate} />
          <Column field="failed" header="失败数" />
          <Column field="last_failure" header="最近错误" body={lastErrorBodyTemplate} />
        </DataTable>
      </Card>
    </div>
  )
}

export default OpsHealth
