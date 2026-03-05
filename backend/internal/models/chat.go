// Package models defines the data structures used by the API.
package models

import (
	"time"
)

// ChatRole represents who sent a message.
type ChatRole string

const (
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
)

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Role      ChatRole  `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRequest is the request body for sending a chat message.
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse is the response for a chat message.
type ChatResponse struct {
	UserMessage      *ChatMessage `json:"user_message"`
	AssistantMessage *ChatMessage `json:"assistant_message"`
}

// Validate validates the chat request.
func (c *ChatRequest) Validate() error {
	if c.Message == "" {
		return ErrValidation("message is required")
	}
	return nil
}
