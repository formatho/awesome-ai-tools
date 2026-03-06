package todoqueue

import (
	"encoding/json"
	"time"
)

// Status represents the current state of a TODO item.
type Status string

const (
	// StatusPending indicates the item is waiting to be processed.
	StatusPending Status = "pending"

	// StatusInProgress indicates the item is currently being processed.
	StatusInProgress Status = "in-progress"

	// StatusCompleted indicates the item has been successfully processed.
	StatusCompleted Status = "completed"

	// StatusFailed indicates the item processing has failed.
	StatusFailed Status = "failed"

	// StatusBlocked indicates the item is blocked by unresolved dependencies.
	StatusBlocked Status = "blocked"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the status is a valid status value.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted, StatusFailed, StatusBlocked:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if the status can transition to the target status.
func (s Status) CanTransitionTo(target Status) bool {
	validTransitions := map[Status][]Status{
		StatusPending:    {StatusInProgress, StatusBlocked},
		StatusInProgress: {StatusCompleted, StatusFailed},
		StatusFailed:     {StatusPending}, // For retries
		StatusBlocked:    {StatusPending}, // When dependencies are resolved
		StatusCompleted:  {},              // Terminal state
	}

	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

// Item represents a single TODO item in the queue.
type Item struct {
	// ID is the unique identifier for the item.
	ID string `json:"id"`

	// Priority determines the order of processing. Higher numbers = higher priority.
	Priority int `json:"priority"`

	// Description is a human-readable description of the TODO item.
	Description string `json:"description"`

	// Status is the current state of the item.
	Status Status `json:"status"`

	// Dependencies is a list of item IDs that must be completed before this item.
	Dependencies []string `json:"dependencies,omitempty"`

	// SkillsRequired is a list of skills/capabilities needed to process this item.
	SkillsRequired []string `json:"skills_required,omitempty"`

	// Result contains the result data when the item is completed.
	Result string `json:"result,omitempty"`

	// Error contains the error message when the item fails.
	Error string `json:"error,omitempty"`

	// RetryCount is the number of times this item has been retried.
	RetryCount int `json:"retry_count"`

	// CreatedAt is the timestamp when the item was created.
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is the timestamp when processing started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is the timestamp when processing completed (success or failure).
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// UpdatedAt is the timestamp when the item was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	// Metadata stores additional custom data as JSON.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IsReady checks if the item is ready to be processed.
// An item is ready if it's pending and all dependencies are met.
func (i *Item) IsReady() bool {
	return i.Status == StatusPending
}

// IsTerminal returns true if the item is in a terminal state (completed).
func (i *Item) IsTerminal() bool {
	return i.Status == StatusCompleted
}

// CanRetry checks if the item can be retried given the max retries.
func (i *Item) CanRetry(maxRetries int) bool {
	return i.Status == StatusFailed && i.RetryCount < maxRetries
}

// Duration returns the time elapsed during processing (if completed).
func (i *Item) Duration() time.Duration {
	if i.StartedAt == nil || i.CompletedAt == nil {
		return 0
	}
	return i.CompletedAt.Sub(*i.StartedAt)
}

// clone creates a deep copy of the item.
func (i *Item) clone() *Item {
	if i == nil {
		return nil
	}

	result := &Item{
		ID:          i.ID,
		Priority:    i.Priority,
		Description: i.Description,
		Status:      i.Status,
		Result:      i.Result,
		Error:       i.Error,
		RetryCount:  i.RetryCount,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}

	if i.StartedAt != nil {
		t := *i.StartedAt
		result.StartedAt = &t
	}
	if i.CompletedAt != nil {
		t := *i.CompletedAt
		result.CompletedAt = &t
	}

	if len(i.Dependencies) > 0 {
		result.Dependencies = make([]string, len(i.Dependencies))
		copy(result.Dependencies, i.Dependencies)
	}

	if len(i.SkillsRequired) > 0 {
		result.SkillsRequired = make([]string, len(i.SkillsRequired))
		copy(result.SkillsRequired, i.SkillsRequired)
	}

	if len(i.Metadata) > 0 {
		result.Metadata = make(map[string]interface{}, len(i.Metadata))
		for k, v := range i.Metadata {
			result.Metadata[k] = v
		}
	}

	return result
}

// MarshalMetadata serializes the Metadata field to JSON bytes.
func (i *Item) MarshalMetadata() ([]byte, error) {
	if i.Metadata == nil {
		return nil, nil
	}
	return json.Marshal(i.Metadata)
}

// UnmarshalMetadata deserializes JSON bytes into the Metadata field.
func (i *Item) UnmarshalMetadata(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &i.Metadata)
}

// MarshalDependencies serializes the Dependencies field to JSON bytes.
func (i *Item) MarshalDependencies() ([]byte, error) {
	if len(i.Dependencies) == 0 {
		return nil, nil
	}
	return json.Marshal(i.Dependencies)
}

// UnmarshalDependencies deserializes JSON bytes into the Dependencies field.
func (i *Item) UnmarshalDependencies(data []byte) error {
	if len(data) == 0 {
		i.Dependencies = nil
		return nil
	}
	return json.Unmarshal(data, &i.Dependencies)
}

// MarshalSkillsRequired serializes the SkillsRequired field to JSON bytes.
func (i *Item) MarshalSkillsRequired() ([]byte, error) {
	if len(i.SkillsRequired) == 0 {
		return nil, nil
	}
	return json.Marshal(i.SkillsRequired)
}

// UnmarshalSkillsRequired deserializes JSON bytes into the SkillsRequired field.
func (i *Item) UnmarshalSkillsRequired(data []byte) error {
	if len(data) == 0 {
		i.SkillsRequired = nil
		return nil
	}
	return json.Unmarshal(data, &i.SkillsRequired)
}
