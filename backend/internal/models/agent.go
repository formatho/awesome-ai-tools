// Package models defines the data structures used by the API.
package models

import (
	"time"
)

// AgentStatus represents the current state of an agent.
type AgentStatus string

const (
	AgentStatusIdle     AgentStatus = "idle"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusPaused   AgentStatus = "paused"
	AgentStatusStopped  AgentStatus = "stopped"
	AgentStatusError    AgentStatus = "error"
)

// Agent represents an AI agent in the system.
type Agent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      AgentStatus            `json:"status"`
	Provider    string                 `json:"provider,omitempty"`
	Model       string                 `json:"model,omitempty"`
	SystemPrompt string                `json:"system_prompt,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	StoppedAt   *time.Time             `json:"stopped_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// AgentCreate is the request body for creating a new agent.
type AgentCreate struct {
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider,omitempty"`
	Model        string                 `json:"model,omitempty"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AgentUpdate is the request body for updating an agent.
type AgentUpdate struct {
	Name         *string                `json:"name,omitempty"`
	Provider     *string                `json:"provider,omitempty"`
	Model        *string                `json:"model,omitempty"`
	SystemPrompt *string                `json:"system_prompt,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the agent creation request.
func (a *AgentCreate) Validate() error {
	if a.Name == "" {
		return ErrValidation("name is required")
	}
	return nil
}
