package models

import (
	"time"
)

// AgentState represents the persistent state of an agent.
type AgentState struct {
	ID        int64                     `json:"id"`
	AgentID   string                    `json:"agent_id" db:"agent_id"`
	StateData interface{}               `json:"state_data" db:"state_data"` // JSON serialized data
	Metadata  map[string]interface{}    `json:"metadata" db:"metadata"`     // JSON serialized metadata
	Version   int                       `json:"version" db:"version"`
	UpdatedAt time.Time                 `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time                 `json:"created_at" db:"created_at"`
}

// AgentStateSummary represents a summary of an agent's state.
type AgentStateSummary struct {
	AgentID          string    `json:"agent_id"`
	TotalVersions    int       `json:"total_versions"`
	LatestVersion    int       `json:"latest_version"`
	LastUpdated      time.Time `json:"last_updated"`
}

// AgentStateExport represents a complete export of agent state with history.
type AgentStateExport struct {
	AgentID    string        `json:"agent_id"`
	Current    *AgentState   `json:"current,omitempty"`
	History    []*AgentState `json:"history,omitempty"`
	ExportedAt time.Time     `json:"exported_at"`
}

// StateMetrics represents aggregate metrics about state persistence.
type StateMetrics struct {
	TotalAgentsTracked          int       `json:"total_agents_tracked"`
	TotalVersionsAcrossAllAgents int       `json:"total_versions_across_all_agents"`
	AvgVersionsPerAgent         float64   `json:"avg_versions_per_agent"`
	MostRecentStateUpdate       time.Time `json:"most_recent_state_update"`
}

// IsErrNotFound checks if error is a not found error for agent state.
func IsErrNotFound(err error) bool {
	return IsNotFoundError(err)
}
