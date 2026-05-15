import axios from 'axios'
import type { User, AuditLog } from '../types'
import { getApiErrorMessage } from './client'

const authClient = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
authClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
authClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  login: async (username: string, password: string): Promise<{ token: string; user: User }> => {
    const res = await authClient.post<{ token: string; user: User }>('/auth/login', {
      username,
      password,
    })
    return res.data
  },

  logout: async (): Promise<void> => {
    try {
      await authClient.post('/auth/logout')
    } finally {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
  },

  refreshToken: async (): Promise<string> => {
    const res = await authClient.post<{ token: string }>('/auth/refresh')
    const token = res.data.token
    localStorage.setItem('token', token)
    return token
  },

  getCurrentUser: async (): Promise<User> => {
    const res = await authClient.get<User>('/users/me')
    return res.data
  },

  listUsers: async (): Promise<User[]> => {
    const res = await authClient.get<User[]>('/users')
    return res.data
  },

  createUser: async (data: {
    username: string
    password: string
    name: string
    email?: string
    role?: User['role']
    force_password_reset?: boolean
  }): Promise<User> => {
    const res = await authClient.post<User>('/users', data)
    return res.data
  },

  updateUser: async (id: number, data: {
    name?: string
    email?: string
    role?: User['role']
    disabled?: boolean
    force_password_reset?: boolean
  }): Promise<User> => {
    const res = await authClient.patch<User>(`/users/${id}`, data)
    return res.data
  },

  deleteUser: async (id: number): Promise<void> => {
    await authClient.delete(`/users/${id}`)
  },

  updateOwnProfile: async (data: {
    name?: string
    email?: string
  }): Promise<User> => {
    const res = await authClient.patch<User>('/users/me/profile', data)
    return res.data
  },

  updateOwnPassword: async (password: string): Promise<void> => {
    await authClient.put('/users/me/password', { password })
  },

  listAuditLogs: async (params: {
    page?: number
    page_size?: number
    action?: string
    result?: string
    start_time?: string
    end_time?: string
  }): Promise<{ items: AuditLog[]; total: number; page: number; page_size: number }> => {
    const res = await authClient.get<{ items: AuditLog[]; total: number; page: number; page_size: number }>('/users/audit-logs', { params })
    return res.data
  },
}

export { getApiErrorMessage }

export default authClient
