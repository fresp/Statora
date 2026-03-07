import { useEffect, useRef, useCallback } from 'react'

type WSHandler = (event: { type: string; data: unknown }) => void

export function useWebSocket(onMessage: WSHandler) {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>()

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => console.log('WebSocket connected')

    ws.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data)
        onMessage(event)
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected, reconnecting in 3s...')
      reconnectTimer.current = setTimeout(connect, 3000)
    }

    ws.onerror = () => ws.close()
  }, [onMessage])

  useEffect(() => {
    connect()
    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])
}
