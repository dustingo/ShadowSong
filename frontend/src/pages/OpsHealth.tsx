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
        setError('加载运维健康数据失败')
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
      <Card title="系统指标 (24小时)">
        <Row gutter={16}>
          <Col span={4}>
            <Statistic title="Webhook 接收" value={metrics?.webhook_ingest_total || 0} />
          </Col>
          <Col span={4}>
            <Statistic title="发送成功" value={metrics?.notification_send_success_total || 0} valueStyle={{ color: '#3f8600' }} />
          </Col>
          <Col span={4}>
            <Statistic title="发送失败" value={metrics?.notification_send_failure_total || 0} valueStyle={{ color: '#cf1322' }} />
          </Col>
          <Col span={4}>
            <Statistic title="重试次数" value={metrics?.notification_retry_total || 0} />
          </Col>
          <Col span={4}>
            <Statistic title="终态失败" value={metrics?.notification_terminal_failure_total || 0} valueStyle={{ color: '#cf1322' }} />
          </Col>
        </Row>
      </Card>

      <Card title="渠道健康度 (24小时)" style={{ marginTop: 16 }}>
        <Table
          dataSource={channelHealth}
          rowKey="channel_id"
          columns={[
            { title: '渠道名称', dataIndex: 'channel_name', key: 'name' },
            { title: '投递总数', dataIndex: 'total_deliveries', key: 'total' },
            { title: '成功率', dataIndex: 'success_rate', key: 'rate', render: (v: number) => `${(v * 100).toFixed(1)}%` },
            { title: '失败数', dataIndex: 'failed', key: 'failed' },
            { title: '最近错误', dataIndex: 'last_failure', key: 'error', render: (v: ChannelHealthResponse['last_failure']) => v?.error_message || '-' },
          ]}
        />
      </Card>
    </div>
  )
}

export default OpsHealth