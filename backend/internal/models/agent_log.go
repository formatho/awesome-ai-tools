// Package models defines the AgentLog model for agent logging.
package models

import (
	"time"
)

// LogLevel represents log severity level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// AgentLog represents a log entry for an agent.
type AgentLog struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// AgentLogCreate is the request body for creating a log entry.
type AgentLogCreate struct {
	AgentID  string                 `json:"agent_id"`
	Level    LogLevel               `json:"level"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
