import axios from 'axios'

const api = axios.create({ baseURL: '/api' })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (r) => r,
  (e) => {
    if (e.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(e)
  }
)

export interface ConversationRecord {
  id: string
  role: string
  content: string
  created_at: string
}

export async function fetchConversations(projectId: string): Promise<ConversationRecord[]> {
  const { data } = await api.get(`/projects/${projectId}/conversations`)
  return data
}

export interface CreatorChatChunk {
  final?: boolean
  content?: string
  options?: string[]
  complete?: boolean
  data?: {
    characters?: Array<{ name: string; role: string; personality: string; background: string; appearance: string }>
    worldSettings?: Array<{ category: string; content: string }>
    outlines?: Array<{ act: number; chapter_num: number; summary: string }>
  }
  saved_ids?: {
    character_ids?: string[]
    world_setting_ids?: string[]
    outline_ids?: string[]
    chapter_ids?: string[]
    volume_id?: string
  }
  error?: string
}

export async function streamCreatorChat(
  projectId: string,
  messages: Array<{ role: string; content: string }>,
  onChunk: (chunk: CreatorChatChunk) => void,
): Promise<void> {
  const token = localStorage.getItem('token')
  const res = await fetch('/api/creator/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'text/event-stream',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ project_id: projectId, messages }),
  })

  if (res.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }

  const reader = res.body?.getReader()
  if (!reader) throw new Error('No response body')

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const parts = buffer.split('\n\n')
    buffer = parts.pop() || ''

    for (const part of parts) {
      for (const line of part.split('\n')) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6).trim()
        if (data === '[DONE]') return
        try {
          onChunk(JSON.parse(data))
        } catch { /* skip malformed chunk */ }
      }
    }
  }
}

export default api
