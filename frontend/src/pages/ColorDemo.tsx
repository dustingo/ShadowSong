import React from 'react'
import { Button } from 'primereact/button'
import { Card } from 'primereact/card'
import { Tag } from 'primereact/tag'
import { InputText } from 'primereact/inputtext'
import { Dropdown } from 'primereact/dropdown'
import { DataTable } from 'primereact/datatable'
import { Column } from 'primereact/column'
import { Chart } from 'primereact/chart'

/**
 * 配色方案 Demo 页面
 * 展示 ShadowSong 项目的完整配色系统
 */

// 配色变量定义
const colors = {
  primary: '#3B82F6',
  primaryHover: '#2563EB',
  primaryLight: '#EFF6FF',
  secondary: '#6366F1',
  secondaryLight: '#EEF2FF',
  success: '#10B981',
  successLight: '#ECFDF5',
  warning: '#F59E0B',
  warningLight: '#FFFBEB',
  danger: '#EF4444',
  dangerLight: '#FEF2F2',
  info: '#06B6D4',
  infoLight: '#ECFEFF',
  surface: {
    ground: '#F8FAFC',
    card: '#FFFFFF',
    border: '#E2E8F0',
    hover: '#F1F5F9',
  },
  text: {
    primary: '#1E293B',
    secondary: '#64748B',
    disabled: '#94A3B8',
    inverse: '#FFFFFF',
  },
}

// 告警严重程度配置
const severityConfig = [
  { level: 'P0', label: '紧急', color: colors.danger, bgColor: colors.dangerLight },
  { level: 'P1', label: '严重', color: '#F97316', bgColor: '#FFF7ED' },
  { level: 'P2', label: '警告', color: colors.warning, bgColor: colors.warningLight },
  { level: 'P3', label: '提示', color: colors.success, bgColor: colors.successLight },
]

// 示例数据
const sampleAlerts = [
  { id: 1, name: 'CPU使用率过高', severity: 'P0', source: 'Prometheus', status: 'firing', time: '2024-01-15 10:30:00' },
  { id: 2, name: '内存不足', severity: 'P1', source: 'Node Exporter', status: 'firing', time: '2024-01-15 10:25:00' },
  { id: 3, name: '磁盘空间预警', severity: 'P2', source: 'Zabbix', status: 'resolved', time: '2024-01-15 10:20:00' },
  { id: 4, name: '服务响应延迟', severity: 'P3', source: 'Grafana', status: 'firing', time: '2024-01-15 10:15:00' },
]

export const ColorDemo: React.FC = () => {
  const chartData = {
    labels: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
    datasets: [
      {
        label: '告警数量',
        data: [12, 8, 15, 25, 18, 10],
        fill: true,
        backgroundColor: 'rgba(59, 130, 246, 0.2)',
        borderColor: colors.primary,
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
        grid: { color: colors.surface.border },
      },
      x: {
        grid: { display: false },
      },
    },
  }

  return (
    <div style={{ background: colors.surface.ground, minHeight: '100vh', padding: '2rem' }}>
      {/* 页面标题 */}
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ color: colors.text.primary, margin: 0, fontSize: '1.75rem', fontWeight: 600 }}>
          ShadowSong 配色方案
        </h1>
        <p style={{ color: colors.text.secondary, margin: '0.5rem 0 0 0' }}>
          专业蓝调风格 - 适用于运维告警监控系统
        </p>
      </div>

      {/* 色板展示 */}
      <Card style={{ marginBottom: '1.5rem', border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>色彩系统</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: '1rem' }}>
          {[
            { name: 'Primary', color: colors.primary, desc: '主色调' },
            { name: 'Secondary', color: colors.secondary, desc: '次要色' },
            { name: 'Success', color: colors.success, desc: '成功/正常' },
            { name: 'Warning', color: colors.warning, desc: '警告' },
            { name: 'Danger', color: colors.danger, desc: '危险/紧急' },
            { name: 'Info', color: colors.info, desc: '信息' },
          ].map((item) => (
            <div key={item.name} style={{ textAlign: 'center' }}>
              <div
                style={{
                  width: '100%',
                  height: '60px',
                  borderRadius: '8px',
                  background: item.color,
                  marginBottom: '0.5rem',
                  boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
                }}
              />
              <div style={{ color: colors.text.primary, fontWeight: 500, fontSize: '0.875rem' }}>{item.name}</div>
              <div style={{ color: colors.text.secondary, fontSize: '0.75rem' }}>{item.desc}</div>
              <div style={{ color: colors.text.disabled, fontSize: '0.75rem', fontFamily: 'monospace' }}>{item.color}</div>
            </div>
          ))}
        </div>
      </Card>

      {/* 统计卡片 */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
        {[
          { label: '活跃告警', value: 23, color: colors.danger, icon: 'pi-bell' },
          { label: 'P0 告警', value: 3, color: colors.danger, icon: 'pi-exclamation-triangle' },
          { label: '已确认', value: 45, color: colors.success, icon: 'pi-check-circle' },
          { label: '已静默', value: 12, color: colors.warning, icon: 'pi-volume-off' },
        ].map((stat, index) => (
          <Card key={index} style={{ border: `1px solid ${colors.surface.border}` }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <div style={{ color: colors.text.secondary, fontSize: '0.875rem', marginBottom: '0.25rem' }}>{stat.label}</div>
                <div style={{ color: stat.color, fontSize: '2rem', fontWeight: 700 }}>{stat.value}</div>
              </div>
              <div
                style={{
                  width: '48px',
                  height: '48px',
                  borderRadius: '12px',
                  background: `${stat.color}15`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <i className={`pi ${stat.icon}`} style={{ color: stat.color, fontSize: '1.25rem' }} />
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* 告警严重程度标签 */}
      <Card style={{ marginBottom: '1.5rem', border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>告警严重程度</h2>
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
          {severityConfig.map((s) => (
            <Tag
              key={s.level}
              value={`${s.level} ${s.label}`}
              style={{
                background: s.bgColor,
                color: s.color,
                border: `1px solid ${s.color}30`,
                fontWeight: 500,
              }}
            />
          ))}
        </div>
      </Card>

      {/* 按钮样式 */}
      <Card style={{ marginBottom: '1.5rem', border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>按钮样式</h2>
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <Button label="主要按钮" style={{ background: colors.primary, border: 'none' }} />
          <Button label="次要按钮" severity="secondary" />
          <Button label="成功" severity="success" />
          <Button label="警告" severity="warning" />
          <Button label="危险" severity="danger" />
          <Button label="信息" severity="info" />
          <Button label=" outlined" outlined />
          <Button label="文字按钮" link style={{ color: colors.primary }} />
        </div>
      </Card>

      {/* 表单元素 */}
      <Card style={{ marginBottom: '1.5rem', border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>表单元素</h2>
        <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <span className="p-float-label">
            <InputText id="demo-input" style={{ width: '200px' }} />
            <label htmlFor="demo-input">输入框</label>
          </span>
          <Dropdown
            placeholder="下拉选择"
            options={[
              { label: '选项一', value: 1 },
              { label: '选项二', value: 2 },
            ]}
            style={{ width: '200px' }}
          />
          <div className="p-input-icon-left">
            <i className="pi pi-search" style={{ color: colors.text.disabled }} />
            <InputText placeholder="搜索..." style={{ width: '200px' }} />
          </div>
        </div>
      </Card>

      {/* 数据表格 */}
      <Card style={{ marginBottom: '1.5rem', border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>告警列表</h2>
        <DataTable value={sampleAlerts} stripedRows style={{ fontSize: '0.875rem' }}>
          <Column field="name" header="告警名称" />
          <Column
            field="severity"
            header="严重程度"
            body={(row) => {
              const config = severityConfig.find(s => s.level === row.severity)
              return (
                <Tag
                  value={row.severity}
                  style={{
                    background: config?.bgColor,
                    color: config?.color,
                    border: `1px solid ${config?.color}30`,
                  }}
                />
              )
            }}
          />
          <Column field="source" header="来源" />
          <Column
            field="status"
            header="状态"
            body={(row) => (
              <Tag
                value={row.status === 'firing' ? '活跃' : '已恢复'}
                severity={row.status === 'firing' ? 'danger' : 'success'}
              />
            )}
          />
          <Column field="time" header="触发时间" />
          <Column
            header="操作"
            body={() => (
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <Button icon="pi pi-check" size="small" style={{ background: colors.primary, border: 'none' }} />
                <Button icon="pi pi-volume-off" size="small" severity="warning" />
              </div>
            )}
          />
        </DataTable>
      </Card>

      {/* 图表 */}
      <Card style={{ border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>24小时告警趋势</h2>
        <div style={{ height: '250px' }}>
          <Chart type="line" data={chartData} options={chartOptions} />
        </div>
      </Card>

      {/* CSS 变量定义 */}
      <Card style={{ marginTop: '1.5rem', border: `1px solid ${colors.surface.border}` }}>
        <h2 style={{ color: colors.text.primary, fontSize: '1.25rem', marginBottom: '1rem' }}>CSS 变量定义</h2>
        <pre style={{
          background: colors.surface.ground,
          padding: '1rem',
          borderRadius: '8px',
          fontSize: '0.75rem',
          overflow: 'auto',
          color: colors.text.primary,
        }}>
{`:root {
  /* 主色调 */
  --primary-color: ${colors.primary};
  --primary-hover-color: ${colors.primaryHover};
  --primary-light-color: ${colors.primaryLight};

  /* 语义色 */
  --success-color: ${colors.success};
  --warning-color: ${colors.warning};
  --danger-color: ${colors.danger};
  --info-color: ${colors.info};

  /* 表面色 */
  --surface-ground: ${colors.surface.ground};
  --surface-card: ${colors.surface.card};
  --surface-border: ${colors.surface.border};
  --surface-hover: ${colors.surface.hover};

  /* 文字色 */
  --text-primary: ${colors.text.primary};
  --text-secondary: ${colors.text.secondary};
  --text-disabled: ${colors.text.disabled};
}`}
        </pre>
      </Card>
    </div>
  )
}
