package agentrunner

import "time"

// Config holds configuration for the agent runner.
type Config struct {
	// DefaultTimeout is the default timeout for agent execution (default: 5 minutes)
	DefaultTimeout time.Duration `json:"default_timeout"`

	// MaxTokens is the maximum tokens per response (default: 4096)
	MaxTokens int `json:"max_tokens"`

	// Temperature is the LLM temperature (default: 0.7)
	Temperature float64 `json:"temperature"`

	// SystemPrompt is the default system prompt
	SystemPrompt string `json:"system_prompt"`

	// EnableMemory enables persistent memory for agents (default: true)
	EnableMemory bool `json:"enable_memory"`

	// MemoryLimit is the context window limit (default: 8 messages)
	MemoryLimit int `json:"memory_limit"`

	// MaxConcurrentAgents limits concurrent agent executions (default: unlimited)
	MaxConcurrentAgents int `json:"max_concurrent_agents"`

	// RetryPolicy defines retry behavior for failed operations
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`
}

// RetryPolicy holds retry configuration.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retries (default: 3)
	MaxRetries int `json:"max_retries"`

	// InitialDelay is the initial delay between retries (default: 1 second)
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay is the maximum delay between retries (default: 30 seconds)
	MaxDelay time.Duration `json:"max_delay"`

	// Multiplier is the exponential backoff multiplier (default: 2.0)
	Multiplier float64 `json:"multiplier"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DefaultTimeout:      5 * time.Minute,
		MaxTokens:           4096,
		Temperature:         0.7,
		SystemPrompt:        "You are a helpful AI assistant.",
		EnableMemory:        true,
		MemoryLimit:         8,
		MaxConcurrentAgents: 10,
		RetryPolicy: &RetryPolicy{
			MaxRetries:   3,
			InitialDelay: time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c Config) Validate() error {
	if c.DefaultTimeout < 0 {
		return ErrInvalidConfig{"default_timeout": "must be non-negative"}
	}
	if c.MaxTokens <= 0 {
		return ErrInvalidConfig{"max_tokens": "must be positive"}
	}
	if c.Temperature < 0 || c.Temperature > 1 {
		return ErrInvalidConfig{"temperature": "must be between 0 and 1"}
	}
	if c.MemoryLimit <= 0 {
		return ErrInvalidConfig{"memory_limit": "must be positive"}
	}
	if c.MaxConcurrentAgents < 0 {
		return ErrInvalidConfig{"max_concurrent_agents": "cannot be negative"}
	}

	if c.RetryPolicy != nil {
		if err := c.RetryPolicy.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the retry policy configuration.
func (p *RetryPolicy) Validate() error {
	if p.MaxRetries < 0 {
		return ErrInvalidConfig{"max_retries": "cannot be negative"}
	}
	if p.InitialDelay <= 0 {
		return ErrInvalidConfig{"initial_delay": "must be positive"}
	}
	if p.MaxDelay <= 0 {
		return ErrInvalidConfig{"max_delay": "must be positive"}
	}
	if p.Multiplier <= 1.0 {
		return ErrInvalidConfig{"multiplier": "must be greater than 1.0"}
	}

	return nil
}
