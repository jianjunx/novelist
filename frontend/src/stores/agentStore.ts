import { create } from 'zustand'
import api from '../api/client'

interface BrainstormData {
  characters?: Array<{ name: string; role: string; personality: string; background: string; appearance: string }>
  worldSettings?: Array<{ category: string; content: string }>
  outlines?: Array<{ act: number; chapter_num: number; summary: string }>
}

interface Message {
  role: 'user' | 'agent'
  content: string
  agent?: string
  options?: string[]
  complete?: boolean
  data?: BrainstormData
}

interface AgentState {
  messages: Message[]
  isStreaming: boolean
  streamContent: string
  brainstormData: BrainstormData | null
  sendMessage: (projectId: string, content: string) => Promise<void>
  clearMessages: () => void
  setBrainstormData: (data: BrainstormData | null) => void
}

export const useAgentStore = create<AgentState>((set, get) => ({
  messages: [],
  isStreaming: false,
  streamContent: '',
  brainstormData: null,
  sendMessage: async (projectId, content) => {
    const userMessage: Message = { role: 'user', content }
    const allMessages = [...get().messages, userMessage]
    set({ messages: allMessages, isStreaming: true, streamContent: '' })

    try {
      const { data } = await api.post('/creator/chat', {
        project_id: projectId,
        messages: allMessages.map(m => ({ role: m.role === 'agent' ? 'assistant' : 'user', content: m.content })),
      })
      const agentMessage: Message = {
        role: 'agent',
        content: data.content,
        agent: 'creator',
        options: data.options,
        complete: data.complete,
        data: data.data,
      }
      set({ messages: [...allMessages, agentMessage], isStreaming: false })
      if (data.data) {
        set({ brainstormData: data.data })
      }
    } catch { set({ isStreaming: false }) }
  },
  clearMessages: () => set({ messages: [], brainstormData: null }),
  setBrainstormData: (data) => set({ brainstormData: data }),
}))
