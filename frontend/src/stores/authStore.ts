import { create } from 'zustand'
import api from '../api/client'

interface User { id: string; username: string }
interface AuthState {
  user: User | null
  token: string | null
  isLoading: boolean
  error: string | null
  login: (u: string, p: string) => Promise<void>
  register: (u: string, p: string) => Promise<void>
  logout: () => void
  checkAuth: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: localStorage.getItem('token'),
  isLoading: false,
  error: null,
  login: async (u, p) => {
    set({ isLoading: true, error: null })
    try {
      const { data } = await api.post('/auth/login', { username: u, password: p })
      localStorage.setItem('token', data.token)
      set({ user: data.user, token: data.token, isLoading: false })
    } catch (e: any) {
      set({ error: e.response?.data?.error || 'Login failed', isLoading: false })
      throw e
    }
  },
  register: async (u, p) => {
    set({ isLoading: true, error: null })
    try {
      const { data } = await api.post('/auth/register', { username: u, password: p })
      localStorage.setItem('token', data.token)
      set({ user: data.user, token: data.token, isLoading: false })
    } catch (e: any) {
      set({ error: e.response?.data?.error || 'Registration failed', isLoading: false })
      throw e
    }
  },
  logout: () => {
    localStorage.removeItem('token')
    set({ user: null, token: null })
  },
  checkAuth: async () => {
    const token = localStorage.getItem('token')
    if (!token) { set({ user: null }); return }
    try {
      const { data } = await api.get('/auth/me')
      set({ user: data })
    } catch {
      localStorage.removeItem('token')
      set({ user: null, token: null })
    }
  },
}))
