package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// Result represents the outcome of a skill execution.
// It contains output data and optional metadata about the execution.
type Result struct {
	// Success indicates whether the execution completed successfully
	Success bool `json:"success"`

	// Data contains the result data from the skill execution
	Data any `json:"data,omitempty"`

	// Message provides a human-readable description of the result
	Message string `json:"message,omitempty"`

	// Metadata contains additional information about the execution
	Metadata map[string]any `json:"metadata,omitempty"`
}

// JSON returns the result as a JSON string for logging or serialization.
func (r Result) JSON() string {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"success": %v, "error": "failed to marshal result"}`, r.Success)
	}
	return string(b)
}

// Action represents a request to execute a specific skill action.
type Action struct {
	// Skill is the name of the skill to execute
	Skill string `json:"skill"`

	// Action is the specific action to perform within the skill
	Action string `json:"action"`

	// Params are the parameters for the action
	Params map[string]any `json:"params,omitempty"`
}

// Skill is the interface that all skills must implement.
// Each skill has a name, a list of supported actions, and an Execute method.
type Skill interface {
	// Name returns the unique identifier for this skill
	Name() string

	// Actions returns the list of actions this skill supports
	Actions() []string

	// Execute performs the specified action with the given parameters
	Execute(ctx context.Context, action string, params map[string]any) (Result, error)
}
