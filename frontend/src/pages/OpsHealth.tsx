import React, { useEffect, useState } from 'react'
import { Card, Table, Statistic, Row, Col, Spin, Alert } from 'antd'
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
        setError('Failed to load ops health data')
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  if (loading) return <Spin />
  if (error) return <Alert type="error" message={error} />

  return (
    <div>
      <Card title="System Metrics (24h)">
        <Row gutter={16}>
          <Col span={4}>
            <Statistic title="Webhook Ingest" value={metrics?.webhook_ingest_total || 0} />
          </Col>
          <Col span={4}>
            <Statistic title="Success" value={metrics?.notification_send_success_total || 0} valueStyle={{ color: '#3f8600' }} />
          </Col>
          <Col span={4}>
            <Statistic title="Failures" value={metrics?.notification_send_failure_total || 0} valueStyle={{ color: '#cf1322' }} />
          </Col>
          <Col span={4}>
            <Statistic title="Retries" value={metrics?.notification_retry_total || 0} />
          </Col>
          <Col span={4}>
            <Statistic title="Terminal Failures" value={metrics?.notification_terminal_failure_total || 0} valueStyle={{ color: '#cf1322' }} />
          </Col>
        </Row>
      </Card>

      <Card title="Channel Health (24h)" style={{ marginTop: 16 }}>
        <Table
          dataSource={channelHealth}
          rowKey="channel_id"
          columns={[
            { title: 'Channel', dataIndex: 'channel_name', key: 'name' },
            { title: 'Total', dataIndex: 'total_deliveries', key: 'total' },
            { title: 'Success Rate', dataIndex: 'success_rate', key: 'rate', render: (v: number) => `${(v * 100).toFixed(1)}%` },
            { title: 'Failed', dataIndex: 'failed', key: 'failed' },
            { title: 'Last Error', dataIndex: 'last_failure', key: 'error', render: (v: ChannelHealthResponse['last_failure']) => v?.error_message || '-' },
          ]}
        />
      </Card>
    </div>
  )
}

export default OpsHealth