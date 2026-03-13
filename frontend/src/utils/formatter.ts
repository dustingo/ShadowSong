// Utility functions for formatting data

export const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

export const formatRelativeTime = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 0) return `${days}天前`
  if (hours > 0) return `${hours}小时前`
  if (minutes > 0) return `${minutes}分钟前`
  return `${seconds}秒前`
}

export const getSeverityColor = (severity: string): string => {
  const colors: Record<string, string> = {
    P0: '#ff4d4f',
    P1: '#ff7a45',
    P2: '#ffa940',
    P3: '#1890ff',
  }
  return colors[severity] || '#d9d9d9'
}

export const getSeverityEmoji = (severity: string): string => {
  const emojis: Record<string, string> = {
    P0: '🔴',
    P1: '🟠',
    P2: '🟡',
    P3: '🔵',
  }
  return emojis[severity] || '⚪'
}

export const getSeverityText = (severity: string): string => {
  const texts: Record<string, string> = {
    P0: '核心服务故障，立即处理',
    P1: '重要服务异常，尽快处理',
    P2: '一般告警，关注',
    P3: '低优先级，了解即可',
  }
  return texts[severity] || '未知级别'
}
