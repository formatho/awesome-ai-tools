type MessageHandler = (data: unknown) => void
type ConnectionHandler = () => void
type ErrorHandler = (error: Event) => void

interface WebSocketOptions {
  url?: string
  reconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  onMessage?: MessageHandler
  onOpen?: ConnectionHandler
  onClose?: ConnectionHandler
  onError?: ErrorHandler
}

class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnect: boolean
  private reconnectInterval: number
  private maxReconnectAttempts: number
  private reconnectAttempts: number = 0
  private messageHandlers: Set<MessageHandler> = new Set()
  private openHandlers: Set<ConnectionHandler> = new Set()
  private closeHandlers: Set<ConnectionHandler> = new Set()
  private errorHandlers: Set<ErrorHandler> = new Set()
  private isIntentionallyClosed: boolean = false

  constructor(options: WebSocketOptions = {}) {
    this.url = options.url || import.meta.env.VITE_WS_URL || 'ws://localhost:18765/ws'
    this.reconnect = options.reconnect ?? true
    this.reconnectInterval = options.reconnectInterval ?? 3000
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 5

    if (options.onMessage) this.messageHandlers.add(options.onMessage)
    if (options.onOpen) this.openHandlers.add(options.onOpen)
    if (options.onClose) this.closeHandlers.add(options.onClose)
    if (options.onError) this.errorHandlers.add(options.onError)
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.isIntentionallyClosed = false
    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
      this.openHandlers.forEach(handler => handler())
    }

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        this.messageHandlers.forEach(handler => handler(data))
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket disconnected')
      this.closeHandlers.forEach(handler => handler())

      // Attempt reconnect if not intentionally closed
      if (this.reconnect && !this.isIntentionallyClosed && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++
        console.log(`Reconnecting... Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`)
        setTimeout(() => this.connect(), this.reconnectInterval)
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      this.errorHandlers.forEach(handler => handler(error))
    }
  }

  disconnect(): void {
    this.isIntentionallyClosed = true
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.error('WebSocket is not connected')
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler)
    return () => this.messageHandlers.delete(handler)
  }

  onOpen(handler: ConnectionHandler): () => void {
    this.openHandlers.add(handler)
    return () => this.openHandlers.delete(handler)
  }

  onClose(handler: ConnectionHandler): () => void {
    this.closeHandlers.add(handler)
    return () => this.closeHandlers.delete(handler)
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler)
    return () => this.errorHandlers.delete(handler)
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// Create singleton instance
export const wsClient = new WebSocketClient()

// Topic-based subscriptions
type TopicHandler = (data: unknown) => void

class TopicManager {
  private subscriptions: Map<string, Set<TopicHandler>> = new Map()

  constructor(private client: WebSocketClient) {
    // Handle incoming messages and route to topic subscribers
    client.onMessage((data: unknown) => {
      if (typeof data === 'object' && data !== null && 'topic' in data) {
        const { topic, payload } = data as { topic: string; payload: unknown }
        const handlers = this.subscriptions.get(topic)
        if (handlers) {
          handlers.forEach(handler => handler(payload))
        }
      }
    })
  }

  subscribe(topic: string, handler: TopicHandler): () => void {
    if (!this.subscriptions.has(topic)) {
      this.subscriptions.set(topic, new Set())
    }
    
    this.subscriptions.get(topic)!.add(handler)

    // Send subscription message to server
    this.client.send({ action: 'subscribe', topic })

    // Return unsubscribe function
    return () => {
      const handlers = this.subscriptions.get(topic)
      if (handlers) {
        handlers.delete(handler)
        if (handlers.size === 0) {
          this.subscriptions.delete(topic)
          this.client.send({ action: 'unsubscribe', topic })
        }
      }
    }
  }
}

export const topicManager = new TopicManager(wsClient)

// Export types
export type { WebSocketOptions, MessageHandler, ConnectionHandler, ErrorHandler }
