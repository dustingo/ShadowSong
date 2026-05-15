/**
 * ShadowSong 主题配置
 * 专业蓝调风格 - 适用于运维告警监控系统
 */

export const theme = {
  colors: {
    // 主色调
    primary: '#3B82F6',
    primaryHover: '#2563EB',
    primaryLight: '#EFF6FF',

    // 次要色
    secondary: '#6366F1',
    secondaryLight: '#EEF2FF',

    // 语义色
    success: '#10B981',
    successLight: '#ECFDF5',
    warning: '#F59E0B',
    warningLight: '#FFFBEB',
    danger: '#EF4444',
    dangerLight: '#FEF2F2',
    info: '#06B6D4',
    infoLight: '#ECFEFF',

    // 表面色
    surface: {
      ground: '#F8FAFC',
      card: '#FFFFFF',
      border: '#E2E8F0',
      hover: '#F1F5F9',
    },

    // 文字色
    text: {
      primary: '#1E293B',
      secondary: '#64748B',
      disabled: '#94A3B8',
      inverse: '#FFFFFF',
    },
  },

  // 告警严重程度配置
  severity: {
    P0: {
      label: 'P0 紧急',
      color: '#EF4444',
      bgColor: '#FEF2F2',
      borderColor: '#FECACA',
    },
    P1: {
      label: 'P1 严重',
      color: '#F97316',
      bgColor: '#FFF7ED',
      borderColor: '#FED7AA',
    },
    P2: {
      label: 'P2 警告',
      color: '#F59E0B',
      bgColor: '#FFFBEB',
      borderColor: '#FDE68A',
    },
    P3: {
      label: 'P3 提示',
      color: '#10B981',
      bgColor: '#ECFDF5',
      borderColor: '#A7F3D0',
    },
  } as Record<string, { label: string; color: string; bgColor: string; borderColor: string }>,

  // 状态配置
  status: {
    firing: {
      label: '活跃',
      color: '#EF4444',
      bgColor: '#FEF2F2',
    },
    resolved: {
      label: '已恢复',
      color: '#10B981',
      bgColor: '#ECFDF5',
    },
    silenced: {
      label: '已静默',
      color: '#F59E0B',
      bgColor: '#FFFBEB',
    },
    acked: {
      label: '已确认',
      color: '#10B981',
      bgColor: '#ECFDF5',
    },
  },

  // 间距
  spacing: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '1rem',
    lg: '1.5rem',
    xl: '2rem',
  },

  // 圆角
  borderRadius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    full: '9999px',
  },

  // 阴影
  shadows: {
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  },
} as const

export type Theme = typeof theme
