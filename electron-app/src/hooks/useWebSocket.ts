import { useEffect, useState, useCallback, useRef } from 'react'
import { wsClient, topicManager } from '../lib/websocket'

interface UseWebSocketOptions {
  autoConnect?: boolean
  topics?: string[]
}

interface WebSocketState {
  isConnected: boolean
  error: string | null
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { autoConnect = true, topics = [] } = options
  const [state, setState] = useState<WebSocketState>({
    isConnected: false,
    error: null,
  })

  useEffect(() => {
    if (!autoConnect) return

    const unsubOpen = wsClient.onOpen(() => {
      setState(prev => ({ ...prev, isConnected: true, error: null }))
    })

    const unsubClose = wsClient.onClose(() => {
      setState(prev => ({ ...prev, isConnected: false }))
    })

    const unsubError = wsClient.onError(() => {
      setState(prev => ({ ...prev, error: 'Connection error' }))
    })

    wsClient.connect()

    return () => {
      unsubOpen()
      unsubClose()
      unsubError()
    }
  }, [autoConnect])

  const send = useCallback((data: unknown) => {
    wsClient.send(data)
  }, [])

  const connect = useCallback(() => {
    wsClient.connect()
  }, [])

  const disconnect = useCallback(() => {
    wsClient.disconnect()
  }, [])

  return {
    ...state,
    send,
    connect,
    disconnect,
  }
}

export function useTopic<T = unknown>(topic: string, handler: (data: T) => void) {
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    const wrappedHandler = (data: unknown) => {
      handlerRef.current(data as T)
    }

    const unsubscribe = topicManager.subscribe(topic, wrappedHandler)
    return unsubscribe
  }, [topic])
}

export function useAgentEvents(agentId: string) {
  const [events, setEvents] = useState<Array<{ type: string; data: unknown; timestamp: Date }>>([])

  useTopic(`agent:${agentId}`, (data: unknown) => {
    setEvents(prev => [
      ...prev.slice(-99), // Keep last 100 events
      {
        type: 'agent-event',
        data,
        timestamp: new Date(),
      },
    ])
  })

  return events
}

export function useTODOUpdates() {
  const [updates, setUpdates] = useState<Array<{ action: string; todo: unknown; timestamp: Date }>>([])

  useTopic('todos', (data: unknown) => {
    const update = data as { action: string; todo: unknown }
    setUpdates(prev => [
      ...prev.slice(-49), // Keep last 50 updates
      {
        ...update,
        timestamp: new Date(),
      },
    ])
  })

  return updates
}

export function useCronUpdates() {
  const [updates, setUpdates] = useState<Array<{ jobId: string; status: string; timestamp: Date }>>([])

  useTopic('cron', (data: unknown) => {
    const update = data as { jobId: string; status: string }
    setUpdates(prev => [
      ...prev.slice(-49), // Keep last 50 updates
      {
        ...update,
        timestamp: new Date(),
      },
    ])
  })

  return updates
}
