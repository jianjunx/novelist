import { create } from 'zustand'
import api from '../api/client'

interface Message { role: 'user' | 'agent'; content: string; agent?: string }
interface AgentState {
  messages: Message[]; isStreaming: boolean; streamContent: string
  sendMessage: (projectId: string, content: string) => Promise<void>
  clearMessages: () => void
}

export const useAgentStore = create<AgentState>((set, get) => ({
  messages: [], isStreaming: false, streamContent: '',
  sendMessage: async (projectId, content) => {
    const userMessage: Message = { role: 'user', content }
    const allMessages = [...get().messages, userMessage]
    set({ messages: allMessages, isStreaming: true, streamContent: '' })

    try {
      const { data } = await api.post('/creator/chat', {
        project_id: projectId,
        messages: allMessages.map(m => ({ role: m.role === 'agent' ? 'assistant' : 'user', content: m.content })),
      })
      const agentMessage: Message = { role: 'agent', content: data.content, agent: 'creator' }
      set({ messages: [...allMessages, agentMessage], isStreaming: false })
    } catch { set({ isStreaming: false }) }
  },
  clearMessages: () => set({ messages: [] }),
}))
