// Package websocket provides real-time communication via WebSockets.
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	fastws "github.com/fasthttp/websocket"
)

// EventType defines the type of WebSocket events.
type EventType string

const (
	EventAgentCreated   EventType = "agent:created"
	EventAgentStatus    EventType = "agent:status"
	EventAgentDeleted   EventType = "agent:deleted"
	EventTODOProgress   EventType = "todo:progress"
	EventTODOStatus     EventType = "todo:status"
	EventCronTriggered  EventType = "cron:triggered"
	EventCronStatus     EventType = "cron:status"
	EventSystemStatus   EventType = "system:status"
)

// Event represents a WebSocket event message.
type Event struct {
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Event
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))

		case event := <-h.broadcast:
			h.mu.RLock()
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Error marshaling event: %v", err)
				h.mu.RUnlock()
				continue
			}

			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					// Client buffer full, close connection
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(eventType EventType, data interface{}) {
	event := &Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
	h.broadcast <- event
}

// BroadcastAgentCreated broadcasts an agent creation event.
func (h *Hub) BroadcastAgentCreated(agent interface{}) {
	h.Broadcast(EventAgentCreated, agent)
}

// BroadcastAgentStatus broadcasts an agent status change.
func (h *Hub) BroadcastAgentStatus(agentID string, status interface{}) {
	h.Broadcast(EventAgentStatus, map[string]interface{}{
		"agent_id": agentID,
		"status":   status,
	})
}

// BroadcastTODOProgress broadcasts a TODO progress update.
func (h *Hub) BroadcastTODOProgress(todoID string, progress int, message string) {
	h.Broadcast(EventTODOProgress, map[string]interface{}{
		"todo_id":  todoID,
		"progress": progress,
		"message":  message,
	})
}

// BroadcastTODOStatus broadcasts a TODO status change.
func (h *Hub) BroadcastTODOStatus(todoID string, status string) {
	h.Broadcast(EventTODOStatus, map[string]string{
		"todo_id": todoID,
		"status":  status,
	})
}

// BroadcastCronTriggered broadcasts a cron job execution event.
func (h *Hub) BroadcastCronTriggered(cronID string, result interface{}) {
	h.Broadcast(EventCronTriggered, map[string]interface{}{
		"cron_id": cronID,
		"result":  result,
	})
}

// BroadcastCronStatus broadcasts a cron job status change.
func (h *Hub) BroadcastCronStatus(cronID string, status string) {
	h.Broadcast(EventCronStatus, map[string]string{
		"cron_id": cronID,
		"status":  status,
	})
}

// BroadcastSystemStatus broadcasts system status updates.
func (h *Hub) BroadcastSystemStatus(status interface{}) {
	h.Broadcast(EventSystemStatus, status)
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Register registers a client with the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Client represents a WebSocket client connection.
type Client struct {
	hub  *Hub
	conn *fastws.Conn
	send chan []byte
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *fastws.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// ReadPump pumps messages from the WebSocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if fastws.IsUnexpectedCloseError(err, fastws.CloseGoingAway, fastws.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
		// For now, we don't process incoming messages from clients
		// This could be extended for client-to-client messaging
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(fastws.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(fastws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(fastws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
