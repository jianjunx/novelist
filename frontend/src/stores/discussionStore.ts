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

interface MultiRoundDiscussionResult {
  total_rounds: number
  rounds: Record<string, DiscussionResult>
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
      const { data } = await api.post<MultiRoundDiscussionResult>(`/chapters/${chapterId}/discuss`)
      // Extract the last round's result
      const rounds = data.rounds || {}
      const lastRound = rounds[String(data.total_rounds)] || rounds['1'] || null
      set({ result: lastRound, isDiscussing: false })
    } catch {
      set({ isDiscussing: false })
    }
  },
  clearResult: () => set({ result: null }),
}))
