import { create } from 'zustand'
import api from '../api/client'

interface Suggestion {
  type: string
  location: string
  problem: string
  suggestion: string
  priority: number
}

interface DiscussionResult {
  editor_suggestions: Suggestion[]
  reader_feedback: string
  critic_analysis: string
  aggregated: Suggestion[]
}

interface DiscussionState {
  result: DiscussionResult | null
  isDiscussing: boolean
  startDiscussion: (chapterId: string) => Promise<void>
  clearResult: () => void
}

export const useDiscussionStore = create<DiscussionState>((set) => ({
  result: null,
  isDiscussing: false,
  startDiscussion: async (chapterId) => {
    set({ isDiscussing: true, result: null })
    try {
      const { data } = await api.post(`/chapters/${chapterId}/discuss`)
      set({ result: data, isDiscussing: false })
    } catch {
      set({ isDiscussing: false })
    }
  },
  clearResult: () => set({ result: null }),
}))
