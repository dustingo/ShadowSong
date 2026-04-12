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

const initialState = getInitialState()

export const useUserStore = create<UserState>((set) => ({
  user: initialState.user,
  token: initialState.token,

  setUser: (user) => {
    if (user) {
      localStorage.setItem('user', JSON.stringify(user))
    } else {
      localStorage.removeItem('user')
    }
    set({ user })
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
