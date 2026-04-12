import { create } from 'zustand'
import type { User } from '../types'

interface UserState {
  user: User | null
  token: string | null
  setUser: (user: User | null) => void
  setToken: (token: string | null) => void
  logout: () => void
}

// 初始化时从 localStorage 读取
const getInitialState = () => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  let user: User | null = null

  if (userStr) {
    try {
      user = JSON.parse(userStr) as User
    } catch {
      user = null
    }
  }

  return { user, token }
}

const normalizeUser = (user: User | null): User | null => {
  if (!user) {
    return null
  }

  return {
    ...user,
    force_password_reset: user.force_password_reset ?? false,
  }
}

const initialState = getInitialState()

export const useUserStore = create<UserState>((set) => ({
  user: normalizeUser(initialState.user),
  token: initialState.token,

  setUser: (user) => {
    const normalizedUser = normalizeUser(user)
    if (normalizedUser) {
      localStorage.setItem('user', JSON.stringify(normalizedUser))
    } else {
      localStorage.removeItem('user')
    }
    set({ user: normalizedUser })
  },

  setToken: (token) => {
    if (token) {
      localStorage.setItem('token', token)
    } else {
      localStorage.removeItem('token')
    }
    set({ token })
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ user: null, token: null })
  },
}))
