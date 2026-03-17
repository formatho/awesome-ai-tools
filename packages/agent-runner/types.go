package agentrunner

import "time"

// AgentConfig holds configuration for creating an AI agent.
type AgentConfig struct {
	// LLM Provider configuration
	Provider   string
	Model      string
	APIKey     string
	BaseURL    string

	// Execution settings
	MaxTokens    int
	Temperature  float64
	SystemPrompt string

	// Memory settings
	MemoryLimit int
}

// AgentResult represents the result of an agent execution.
type AgentResult struct {
	ID        string
	Status    AgentStatus
	Result    any
	Error     error
	Duration  time.Duration
	CreatedAt time.Time
}
