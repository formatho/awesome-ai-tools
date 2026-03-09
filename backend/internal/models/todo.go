package models

import (
	"time"
)

// TODOStatus represents the current state of a TODO item.
type TODOStatus string

const (
	TODOStatusPending   TODOStatus = "pending"
	TODOStatusQueued    TODOStatus = "queued"
	TODOStatusRunning   TODOStatus = "running"
	TODOStatusCompleted TODOStatus = "completed"
	TODOStatusFailed    TODOStatus = "failed"
	TODOStatusCancelled TODOStatus = "cancelled"
)

// TODO represents a task item in the queue.
type TODO struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description,omitempty"`
	Status         TODOStatus             `json:"status"`
	Priority       int                    `json:"priority"`
	Progress       int                    `json:"progress"` // 0-100
	AgentID        *string                `json:"agent_id,omitempty"`
	OrganizationID string                 `json:"organization_id,omitempty"`
	Skills         []string               `json:"skills,omitempty"`
	Dependencies   []string               `json:"dependencies,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
}

// TODOCreate is the request body for creating a new TODO.
type TODOCreate struct {
	Title          string                 `json:"title"`
	Description    string                 `json:"description,omitempty"`
	Priority       int                    `json:"priority,omitempty"`
	AgentID        *string                `json:"agent_id,omitempty"`
	OrganizationID string                 `json:"organization_id,omitempty"`
	Skills         []string               `json:"skills,omitempty"`
	Dependencies   []string               `json:"dependencies,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty"`
}

// TODOUpdate is the request body for updating a TODO.
type TODOUpdate struct {
	Title          *string                `json:"title,omitempty"`
	Description    *string                `json:"description,omitempty"`
	Priority       *int                   `json:"priority,omitempty"`
	AgentID        *string                `json:"agent_id,omitempty"`
	OrganizationID *string                `json:"organization_id,omitempty"`
	Skills         []string               `json:"skills,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty"`
}

// Validate validates the TODO creation request.
func (t *TODOCreate) Validate() error {
	if t.Title == "" {
		return ErrValidation("title is required")
	}
	return nil
}
