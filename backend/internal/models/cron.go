package models

import (
	"time"
)

// CronStatus represents the current state of a cron job.
type CronStatus string

const (
	CronStatusActive   CronStatus = "active"
	CronStatusPaused   CronStatus = "paused"
	CronStatusDisabled CronStatus = "disabled"
)

// Cron represents a scheduled job.
type Cron struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Schedule       string                 `json:"schedule"` // Cron expression
	Timezone       string                 `json:"timezone,omitempty"`
	Status         CronStatus             `json:"status"`
	AgentID        string                 `json:"agent_id"`
	OrganizationID string                 `json:"organization_id,omitempty"`
	TaskName       string                 `json:"task_name,omitempty"`
	TaskConfig     map[string]interface{} `json:"task_config,omitempty"`
	LastRunAt      *time.Time             `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time             `json:"next_run_at,omitempty"`
	LastResult     string                 `json:"last_result,omitempty"`
	LastError      string                 `json:"last_error,omitempty"`
	RunCount       int                    `json:"run_count"`
	SuccessCount   int                    `json:"success_count"`
	FailCount      int                    `json:"fail_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CronCreate is the request body for creating a new cron job.
type CronCreate struct {
	Name           string                 `json:"name"`
	Schedule       string                 `json:"schedule"`
	Timezone       string                 `json:"timezone,omitempty"`
	AgentID        string                 `json:"agent_id"`
	OrganizationID string                 `json:"organization_id,omitempty"`
	TaskName       string                 `json:"task_name,omitempty"`
	TaskConfig     map[string]interface{} `json:"task_config,omitempty"`
}

// CronUpdate is the request body for updating a cron job.
type CronUpdate struct {
	Name           *string                `json:"name,omitempty"`
	Schedule       *string                `json:"schedule,omitempty"`
	Timezone       *string                `json:"timezone,omitempty"`
	AgentID        *string                `json:"agent_id,omitempty"`
	OrganizationID *string                `json:"organization_id,omitempty"`
	TaskName       *string                `json:"task_name,omitempty"`
	TaskConfig     map[string]interface{} `json:"task_config,omitempty"`
}

// CronHistory represents a single execution record.
type CronHistory struct {
	ID        string                 `json:"id"`
	CronID    string                 `json:"cron_id"`
	StartedAt time.Time              `json:"started_at"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
	Status    string                 `json:"status"` // success, failed, timeout
	Result    string                 `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the cron creation request.
func (c *CronCreate) Validate() error {
	if c.Name == "" {
		return ErrValidation("name is required")
	}
	if c.Schedule == "" {
		return ErrValidation("schedule is required")
	}
	if c.AgentID == "" {
		return ErrValidation("agent_id is required")
	}
	return nil
}
