import { create } from 'zustand'
import { fetchConversations, streamCreatorChat } from '../api/client'

interface BrainstormData {
  characters?: Array<{ name: string; role: string; personality: string; background: string; appearance: string }>
  worldSettings?: Array<{ category: string; content: string }>
  outlines?: Array<{ act: number; chapter_num: number; summary: string }>
}

interface SavedIDs {
  character_ids?: string[]
  world_setting_ids?: string[]
  outline_ids?: string[]
  chapter_ids?: string[]
  volume_id?: string
}

interface Message {
  role: 'user' | 'agent'
  content: string
  agent?: string
  options?: string[]
  complete?: boolean
  data?: BrainstormData
  saved_ids?: SavedIDs
}

function parseAssistantContent(content: string): Message {
  try {
    const data = JSON.parse(content)
    if (data.content) {
      return {
        role: 'agent',
        content: data.content,
        agent: 'creator',
        options: data.options,
        complete: data.complete,
        data: data.data,
        saved_ids: data.saved_ids,
      }
    }
  } catch { /* raw text fallback */ }
  return { role: 'agent', content, agent: 'creator' }
}

function isFinalChunk(chunk: { final?: boolean }): boolean {
  return chunk.final === true
}

interface AgentState {
  messages: Message[]
  isStreaming: boolean
  streamContent: string
  brainstormData: BrainstormData | null
  savedIDs: SavedIDs | null
  sendMessage: (projectId: string, content: string) => Promise<void>
  loadConversations: (projectId: string) => Promise<void>
  clearMessages: () => void
  setBrainstormData: (data: BrainstormData | null) => void
}

export const useAgentStore = create<AgentState>((set, get) => ({
  messages: [],
  isStreaming: false,
  streamContent: '',
  brainstormData: null,
  savedIDs: null,
  sendMessage: async (projectId, content) => {
    const userMessage: Message = { role: 'user', content }
    const allMessages = [...get().messages, userMessage]
    const placeholder: Message = { role: 'agent', content: '', agent: 'creator' }
    set({ messages: [...allMessages, placeholder], isStreaming: true, streamContent: '' })

    const apiMessages = allMessages.map(m => ({
      role: m.role === 'agent' ? 'assistant' : 'user',
      content: m.content,
    }))

    let accumulated = ''

    try {
      await streamCreatorChat(projectId, apiMessages, (chunk) => {
        if (chunk.error) throw new Error(chunk.error)

        if (isFinalChunk(chunk)) {
          const agentMessage: Message = {
            role: 'agent',
            content: chunk.content ?? accumulated,
            agent: 'creator',
            options: chunk.options,
            complete: chunk.complete,
            data: chunk.data,
            saved_ids: chunk.saved_ids,
          }
          set({
            messages: [...allMessages, agentMessage],
            streamContent: agentMessage.content,
            ...(chunk.data ? { brainstormData: chunk.data } : {}),
            ...(chunk.saved_ids ? { savedIDs: chunk.saved_ids } : {}),
          })
        } else if (chunk.content) {
          accumulated += chunk.content
          const msgs = get().messages
          const last = msgs[msgs.length - 1]
          if (last?.role === 'agent') {
            set({
              messages: [...msgs.slice(0, -1), { ...last, content: accumulated }],
              streamContent: accumulated,
            })
          }
        }
      })
      set({ isStreaming: false })
    } catch {
      const msgs = get().messages
      if (msgs.length > 0 && msgs[msgs.length - 1].role === 'agent' && !msgs[msgs.length - 1].content) {
        set({ messages: msgs.slice(0, -1), isStreaming: false })
      } else {
        set({ isStreaming: false })
      }
    }
  },
  loadConversations: async (projectId) => {
    try {
      const records = await fetchConversations(projectId)
      const messages: Message[] = records.map((r) =>
        r.role === 'user' ? { role: 'user', content: r.content } : parseAssistantContent(r.content)
      )
      const lastAgent = [...messages].reverse().find((m) => m.role === 'agent')
      set({
        messages,
        brainstormData: lastAgent?.data ?? null,
        savedIDs: lastAgent?.saved_ids ?? null,
      })
    } catch { /* keep empty state on failure */ }
  },
  clearMessages: () => set({ messages: [], brainstormData: null, savedIDs: null }),
  setBrainstormData: (data) => set({ brainstormData: data }),
}))
