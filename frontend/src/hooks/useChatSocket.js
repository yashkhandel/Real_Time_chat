import { useEffect, useRef, useState, useCallback } from 'react'
import { wsUrl } from '../api/client'

// Manages one WebSocket connection with exponential-backoff reconnect.
// Exposes incoming envelopes via onEnvelope callback and a send() function.
export function useChatSocket(onEnvelope) {
  const wsRef = useRef(null)
  const retryRef = useRef(0)
  const [connected, setConnected] = useState(false)
  const onEnvelopeRef = useRef(onEnvelope)
  onEnvelopeRef.current = onEnvelope

  const connect = useCallback(() => {
    const socket = new WebSocket(wsUrl())
    wsRef.current = socket

    socket.onopen = () => {
      setConnected(true)
      retryRef.current = 0
    }

    socket.onmessage = (event) => {
      try {
        const envelope = JSON.parse(event.data)
        onEnvelopeRef.current?.(envelope)
      } catch {
        // ignore malformed frames
      }
    }

    socket.onclose = () => {
      setConnected(false)
      const delay = Math.min(1000 * 2 ** retryRef.current, 15000)
      retryRef.current += 1
      setTimeout(connect, delay)
    }

    socket.onerror = () => socket.close()
  }, [])

  useEffect(() => {
    connect()
    return () => wsRef.current?.close()
  }, [connect])

  const send = useCallback((payload) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(payload))
    }
  }, [])

  return { connected, send }
}
