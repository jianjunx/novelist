import { useEffect, useRef, useCallback } from 'react'
import { useAuthStore } from '../stores/authStore'

export function useWebSocket(onMessage?: (data: any) => void) {
  const ws = useRef<WebSocket | null>(null)
  const { token } = useAuthStore()

  useEffect(() => {
    if (!token) return
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws.current = new WebSocket(`${protocol}//${window.location.host}/api/ws?token=${token}`)
    ws.current.onmessage = (e) => {
      try { onMessage?.(JSON.parse(e.data)) } catch {}
    }
    return () => { ws.current?.close() }
  }, [token, onMessage])

  const send = useCallback((data: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) ws.current.send(JSON.stringify(data))
  }, [])

  return { send }
}
