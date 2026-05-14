import React, { createContext, useContext, useRef, ReactNode } from 'react'
import { Toast } from 'primereact/toast'
import type { ToastMessage } from 'primereact/toast'

interface ToastContextType {
  show: (message: ToastMessage) => void
  showSuccess: (summary: string, detail?: string) => void
  showError: (summary: string, detail?: string) => void
  showInfo: (summary: string, detail?: string) => void
  showWarn: (summary: string, detail?: string) => void
  clear: () => void
}

const ToastContext = createContext<ToastContextType | null>(null)

export const useToast = (): ToastContextType => {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context
}

interface ToastProviderProps {
  children: ReactNode
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const toastRef = useRef<Toast>(null)

  const show = (message: ToastMessage) => {
    toastRef.current?.show(message)
  }

  const showSuccess = (summary: string, detail?: string) => {
    show({ severity: 'success', summary, detail, life: 3000 })
  }

  const showError = (summary: string, detail?: string) => {
    show({ severity: 'error', summary, detail, life: 5000 })
  }

  const showInfo = (summary: string, detail?: string) => {
    show({ severity: 'info', summary, detail, life: 3000 })
  }

  const showWarn = (summary: string, detail?: string) => {
    show({ severity: 'warn', summary, detail, life: 4000 })
  }

  const clear = () => {
    toastRef.current?.clear()
  }

  return (
    <ToastContext.Provider value={{ show, showSuccess, showError, showInfo, showWarn, clear }}>
      <Toast ref={toastRef} position="top-right" />
      {children}
    </ToastContext.Provider>
  )
}