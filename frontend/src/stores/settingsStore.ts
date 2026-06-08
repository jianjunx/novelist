import { create } from 'zustand'
import api from '../api/client'

export interface ModelOption {
  value: string
  label: string
  provider: string
}

export interface SettingsData {
  deepseekKey: string
  claudeKey: string
  openaiKey: string
  localModelUrl: string
  defaultModel: string
  discussionRounds: number
}

interface SettingsState extends SettingsData {
  loading: boolean
  error: string | null
  models: ModelOption[]
  fetchSettings: () => Promise<void>
  fetchModels: () => Promise<void>
  updateSettings: (data: Partial<SettingsData>) => Promise<void>
}

function mapFromApi(data: Record<string, unknown>): SettingsData {
  return {
    deepseekKey: (data.deepseek_key as string) || '',
    claudeKey: (data.claude_key as string) || '',
    openaiKey: (data.openai_key as string) || '',
    localModelUrl: (data.local_model_url as string) || '',
    defaultModel: (data.default_model as string) || 'deepseek-chat',
    discussionRounds: (data.discussion_rounds as number) || 1,
  }
}

function mapToApi(data: Partial<SettingsData>): Record<string, unknown> {
  const result: Record<string, unknown> = {}
  if (data.deepseekKey !== undefined) result.deepseek_key = data.deepseekKey
  if (data.claudeKey !== undefined) result.claude_key = data.claudeKey
  if (data.openaiKey !== undefined) result.openai_key = data.openaiKey
  if (data.localModelUrl !== undefined) result.local_model_url = data.localModelUrl
  if (data.defaultModel !== undefined) result.default_model = data.defaultModel
  if (data.discussionRounds !== undefined) result.discussion_rounds = data.discussionRounds
  return result
}

export const useSettingsStore = create<SettingsState>((set) => ({
  deepseekKey: '',
  claudeKey: '',
  openaiKey: '',
  localModelUrl: '',
  defaultModel: 'deepseek-chat',
  discussionRounds: 1,
  loading: false,
  error: null,
  models: [],

  fetchSettings: async () => {
    set({ loading: true, error: null })
    try {
      const { data } = await api.get('/settings')
      set({ ...mapFromApi(data), loading: false })
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      set({ error: err.response?.data?.error || '加载设置失败', loading: false })
      throw e
    }
  },

  fetchModels: async () => {
    try {
      const { data } = await api.get('/settings/models')
      set({ models: data.models || [] })
    } catch {
      // 静默失败，使用空列表
    }
  },

  updateSettings: async (data) => {
    set({ loading: true, error: null })
    try {
      await api.put('/settings', mapToApi(data))
      set((state) => ({ ...state, ...data, loading: false }))
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string } } }
      set({ error: err.response?.data?.error || '保存设置失败', loading: false })
      throw e
    }
  },
}))
