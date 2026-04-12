import React from 'react'
import { Alert } from 'antd'

type PermissionNoticeProps = {
  title?: string
  description?: string
  type?: 'info' | 'warning' | 'error'
}

export const PermissionNotice: React.FC<PermissionNoticeProps> = ({
  title = '当前角色无权执行该操作',
  description = '你可以继续查看当前页面内容，但无法执行受限操作。',
  type = 'warning',
}) => (
  <Alert
    showIcon
    type={type}
    message={title}
    description={description}
    style={{ marginBottom: 16 }}
  />
)
