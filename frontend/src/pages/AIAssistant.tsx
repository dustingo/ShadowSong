import React, { useState, useRef, useEffect } from 'react'
import {
  Card,
  Input,
  Button,
  Space,
  List,
  Tag,
  Typography,
  Spin,
  message,
  Avatar,
} from 'antd'
import { SendOutlined, RobotOutlined, UserOutlined, CheckOutlined, CloseOutlined, DeleteOutlined } from '@ant-design/icons'
import ReactMarkdown from 'react-markdown'
import { aiApi } from '../api/client'

const { Paragraph } = Typography

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

interface AILog {
  id: number
  alert_id: string
  alert_name: string
  input: string
  output: string
  accurate: boolean | null
  created_at: string
}

interface SilenceRecommendation {
  id: number
  alert_name: string
  source: string
  frequency: number
  suggested_duration: number
  reason: string
}

export const AIAssistant: React.FC = () => {
  const [activeTab, setActiveTab] = useState('chat')

  return (
    <div style={{ height: 'calc(100vh - 120px)', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, minHeight: 0 }}>
        {activeTab === 'chat' && <AIChatPage />}
        {activeTab === 'silence' && <SilenceRecommendations />}
        {activeTab === 'logs' && <AILogs />}
      </div>
      <div style={{ position: 'sticky', bottom: 0, left: 0, right: 0, zIndex: 100, background: '#fff' }}>
        <TabButtons activeTab={activeTab} onChange={setActiveTab} />
      </div>
    </div>
  )
}

const TabButtons: React.FC<{ activeTab: string; onChange: (tab: string) => void }> = ({ activeTab, onChange }) => {
  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      gap: 8,
      padding: '12px 24px',
      background: '#fff',
      borderTop: '1px solid #f0f0f0',
    }}>
      <Button
        type={activeTab === 'chat' ? 'primary' : 'default'}
        onClick={() => onChange('chat')}
      >
        💬 智能问答
      </Button>
      <Button
        type={activeTab === 'silence' ? 'primary' : 'default'}
        onClick={() => onChange('silence')}
      >
        🔔 静默推荐
      </Button>
      <Button
        type={activeTab === 'logs' ? 'primary' : 'default'}
        onClick={() => onChange('logs')}
      >
        📝 AI 日志
      </Button>
    </div>
  )
}

// ============ AI Chat Page ============

const AIChatPage: React.FC = () => {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, loading])

  const handleSend = async () => {
    if (!input.trim() || loading) return

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date().toISOString(),
    }

    setMessages((prev) => [...prev, userMessage])
    const question = input
    setInput('')
    setLoading(true)

    try {
      const res = await aiApi.chat({ message: question }) as any
      const assistantMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: res.reply,
        timestamp: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, assistantMessage])
    } catch (error: any) {
      console.error('AI chat error:', error)
      const errorMsg = error?.response?.data?.error || error?.message || 'AI 响应失败'
      message.error(errorMsg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 220px)', background: '#fff', borderRadius: 8, overflow: 'hidden' }}>
      {/* 消息列表 - 可滚动 */}
      <div style={{ flex: 1, overflowY: 'auto', padding: '24px', background: '#fafafa' }}>
        {messages.length === 0 ? (
          <div style={{ textAlign: 'center', paddingTop: 60, color: '#999' }}>
            <RobotOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <Paragraph style={{ color: '#999' }}>我是 AI 助手，有什么可以帮您？</Paragraph>
          </div>
        ) : (
          <div style={{ maxWidth: 900, margin: '0 auto' }}>
            {messages.map((msg) => (
              <div
                key={msg.id}
                style={{
                  display: 'flex',
                  marginBottom: 20,
                  flexDirection: msg.role === 'user' ? 'row-reverse' : 'row',
                  alignItems: 'flex-start',
                }}
              >
                <Avatar
                  size={40}
                  icon={msg.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
                  style={{
                    backgroundColor: msg.role === 'user' ? '#52c41a' : '#1890ff',
                    margin: msg.role === 'user' ? '0 0 0 12' : '0 12 0 0',
                    flexShrink: 0,
                  }}
                />
                <div
                  style={{
                    flex: 1,
                    maxWidth: 'calc(100% - 100px)',
                    padding: '14px 18px',
                    borderRadius: 12,
                    background: msg.role === 'user' ? '#1890ff' : '#fff',
                    boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
                  }}
                >
                  {msg.role === 'assistant' ? (
                    <ReactMarkdown
                      components={{
                        h1: ({ children }) => <h1 style={{ fontSize: '1.5em', fontWeight: 600, margin: '16px 0 12px 0', borderBottom: '1px solid #eee', paddingBottom: 8 }}>{children}</h1>,
                        h2: ({ children }) => <h2 style={{ fontSize: '1.3em', fontWeight: 600, margin: '14px 0 10px 0', borderBottom: '1px solid #eee', paddingBottom: 6 }}>{children}</h2>,
                        h3: ({ children }) => <h3 style={{ fontSize: '1.1em', fontWeight: 600, margin: '12px 0 8px 0' }}>{children}</h3>,
                        h4: ({ children }) => <h4 style={{ fontSize: '1em', fontWeight: 600, margin: '10px 0 6px 0' }}>{children}</h4>,
                        p: ({ children }) => <p style={{ margin: '0 0 12px 0', lineHeight: 1.7 }}>{children}</p>,
                        ul: ({ children }) => <ul style={{ margin: '0 0 12px 24px', lineHeight: 1.8 }}>{children}</ul>,
                        ol: ({ children }) => <ol style={{ margin: '0 0 12px 24px', lineHeight: 1.8 }}>{children}</ol>,
                        li: ({ children, node }) => {
                          const hasSublist = node?.children?.some((child: any) => child.type === 'ul' || child.type === 'ol')
                          return <li style={{ marginBottom: hasSublist ? 8 : 4 }}>{children}</li>
                        },
                        code: ({ className, children, ...props }) => {
                          const match = /language-(\w+)/.exec(className || '')
                          const isInline = !match && !className
                          if (isInline) {
                            return <code style={{ background: '#f5f5f5', padding: '2px 6px', borderRadius: 4, fontSize: '0.9em', fontFamily: 'monospace' }} {...props}>{children}</code>
                          }
                          return <code className={className} style={{ background: '#2d2d2d', color: '#f8f8f2', padding: '2px 6px', borderRadius: 4, fontSize: '0.9em', fontFamily: 'monospace' }} {...props}>{children}</code>
                        },
                        pre: ({ children }) => <pre style={{ background: '#2d2d2d', color: '#f8f8f2', padding: 14, borderRadius: 8, overflow: 'auto', margin: '0 0 14px 0', fontSize: '0.9em', lineHeight: 1.5 }}>{children}</pre>,
                        blockquote: ({ children }) => <blockquote style={{ borderLeft: '4px solid #1890ff', margin: '0 0 12px 0', padding: '8px 16px', background: '#f9f9f9', borderRadius: '0 6px 6px 0' }}>{children}</blockquote>,
                        a: ({ href, children }) => <a href={href} style={{ color: '#1890ff', textDecoration: 'underline' }} target="_blank" rel="noopener noreferrer">{children}</a>,
                        table: ({ children }) => <table style={{ borderCollapse: 'collapse', width: '100%', margin: '0 0 14px 0', fontSize: '0.9em' }}>{children}</table>,
                        thead: ({ children }) => <thead style={{ background: '#f5f5f5' }}>{children}</thead>,
                        th: ({ children }) => <th style={{ border: '1px solid #ddd', padding: '8px 12px', textAlign: 'left', fontWeight: 600 }}>{children}</th>,
                        td: ({ children }) => <td style={{ border: '1px solid #ddd', padding: '8px 12px' }}>{children}</td>,
                        hr: () => <hr style={{ border: 'none', borderTop: '1px solid #eee', margin: '16px 0' }} />,
                        strong: ({ children }) => <strong style={{ fontWeight: 600 }}>{children}</strong>,
                        em: ({ children }) => <em style={{ fontStyle: 'italic' }}>{children}</em>,
                        del: ({ children }) => <del style={{ textDecoration: 'line-through', color: '#999' }}>{children}</del>,
                      }}
                    >
                      {msg.content}
                    </ReactMarkdown>
                  ) : (
                    <Paragraph style={{ margin: 0, color: '#fff', whiteSpace: 'pre-wrap', lineHeight: 1.7 }}>{msg.content}</Paragraph>
                  )}
                </div>
              </div>
            ))}
            {loading && (
              <div style={{ display: 'flex', alignItems: 'center', color: '#999', paddingLeft: 52 }}>
                <Spin size="small" style={{ marginRight: 10 }} />
                <span>AI 正在思考...</span>
              </div>
            )}
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 输入框 - 固定底部 */}
      <div style={{ padding: '14px 24px', borderTop: '1px solid #f0f0f0', background: '#fff' }}>
        <div style={{ maxWidth: 900, margin: '0 auto' }}>
          <Space.Compact style={{ width: '100%' }}>
            <Input.TextArea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="输入您的问题..."
              autoSize={{ minRows: 1, maxRows: 3 }}
              onPressEnter={(e) => {
                if (!e.shiftKey) {
                  e.preventDefault()
                  handleSend()
                }
              }}
              disabled={loading}
              style={{ borderRadius: '8px 0 0 8px' }}
            />
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={handleSend}
              loading={loading}
              style={{ height: 'auto', borderRadius: '0 8px 8px 0', padding: '8px 18px' }}
            />
          </Space.Compact>
          <div style={{ marginTop: 8, color: '#bbb', fontSize: 12, textAlign: 'center' }}>
            Enter 发送 · Shift + Enter 换行
          </div>
        </div>
      </div>
    </div>
  )
}

// ============ Silence Recommendations ============

const SilenceRecommendations: React.FC = () => {
  const [recommendations, setRecommendations] = useState<SilenceRecommendation[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchRecommendations()
  }, [])

  const fetchRecommendations = async () => {
    setLoading(true)
    try {
      const data = await aiApi.silenceRecommendations() as unknown as SilenceRecommendation[]
      setRecommendations(data)
    } catch (error) {
      message.error('获取推荐失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAdopt = async (id: number) => {
    try {
      await aiApi.adoptSilenceRecommendation(id)
      message.success('已采纳')
      fetchRecommendations()
    } catch (error) {
      message.error('采纳失败')
    }
  }

  const handleIgnore = async (id: number) => {
    try {
      await aiApi.ignoreSilenceRecommendation(id)
      setRecommendations((prev) => prev.filter((r) => r.id !== id))
    } catch (error) {
      message.error('忽略失败')
    }
  }

  return (
    <Card title="🔔 静默推荐" style={{ height: 'calc(100vh - 220px)', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      <div style={{ flex: 1, overflowY: 'auto' }}>
        <List
          loading={loading}
          dataSource={recommendations}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button type="primary" size="small" icon={<CheckOutlined />} onClick={() => handleAdopt(item.id)}>采纳</Button>,
                <Button size="small" icon={<CloseOutlined />} onClick={() => handleIgnore(item.id)}>忽略</Button>,
              ]}
            >
              <List.Item.Meta
                title={<Space><Tag color="blue">{item.alert_name}</Tag><span style={{ color: '#999' }}>{item.source}</span></Space>}
                description={
                  <div>
                    <div>24小时内触发 <b>{item.frequency}</b> 次</div>
                    <div style={{ color: '#999' }}>推荐静默: {Math.round(item.suggested_duration / 3600)} 小时</div>
                    <div style={{ color: '#999' }}>原因: {item.reason}</div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
        {recommendations.length === 0 && !loading && (
          <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无推荐</div>
        )}
      </div>
    </Card>
  )
}

// ============ AI Logs ============

const AILogs: React.FC = () => {
  const [logs, setLogs] = useState<AILog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchLogs()
  }, [page])

  const fetchLogs = async () => {
    setLoading(true)
    try {
      const res = await aiApi.logs({ page, page_size: 10 }) as unknown as { list: AILog[]; total: number }
      setLogs(res.list)
      setTotal(res.total)
    } catch (error) {
      message.error('获取日志失败')
    } finally {
      setLoading(false)
    }
  }

  const handleMarkAccuracy = async (id: number, accurate: boolean) => {
    try {
      await aiApi.markAccuracy(id, accurate)
      setLogs((prev) => prev.map((log) => (log.id === id ? { ...log, accurate } : log)))
    } catch (error) {
      message.error('标记失败')
    }
  }

  const handleDeleteLog = async (id: number) => {
    try {
      await aiApi.deleteLog(id)
      setLogs((prev) => prev.filter((log) => log.id !== id))
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const accurateCount = logs.filter((l) => l.accurate === true).length
  const inaccurateCount = logs.filter((l) => l.accurate === false).length

  return (
    <Card
      title="📝 AI 处理日志"
      extra={<Space><Tag color="green">准确: {accurateCount}</Tag><Tag color="red">不准确: {inaccurateCount}</Tag></Space>}
      style={{ height: 'calc(100vh - 220px)', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}
    >
      <div style={{ flex: 1, overflowY: 'auto' }}>
        <List
          loading={loading}
          dataSource={logs}
          pagination={{ current: page, total, pageSize: 10, onChange: setPage }}
          renderItem={(item) => (
            <List.Item
              actions={[
                <Button size="small" onClick={() => handleMarkAccuracy(item.id, true)}>准确</Button>,
                <Button size="small" danger onClick={() => handleMarkAccuracy(item.id, false)}>不准确</Button>,
                <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDeleteLog(item.id)} />,
              ]}
            >
              <List.Item.Meta
                title={<Space><Tag>{item.alert_name}</Tag>{item.accurate === true && <Tag color="green">准确</Tag>}{item.accurate === false && <Tag color="red">不准确</Tag>}</Space>}
                description={<div><Paragraph ellipsis={{ rows: 2 }} style={{ margin: 0 }}>{item.output}</Paragraph><span style={{ color: '#999', fontSize: 12 }}>{new Date(item.created_at).toLocaleString()}</span></div>}
              />
            </List.Item>
          )}
        />
      </div>
    </Card>
  )
}
