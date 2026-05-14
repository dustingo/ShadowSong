import React from 'react'
import { Message } from 'primereact/message'

type PermissionNoticeProps = {
  title?: string
  description?: string
  type?: 'info' | 'warn' | 'error' | 'success'
}

export const PermissionNotice: React.FC<PermissionNoticeProps> = ({
  title = '当前角色无权执行该操作',
  description,
  type = 'warn',
}) => {
  return (
    <Message
      severity={type}
      text={title}
      style={{ marginBottom: '16px' }}
    />
  )
}