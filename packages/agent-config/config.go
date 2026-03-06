// Package agentconfig provides configuration management for agent orchestration.
// It supports multi-format configuration files (YAML, TOML, JSON) with validation
// and hot reload capabilities.
package agentconfig

// LLMConfig holds LLM-specific configuration parameters.
// These can be set globally and overridden per-agent.
type LLMConfig struct {
	// Provider is the LLM provider (e.g., "openai", "anthropic", "ollama")
	Provider string `yaml:"provider" json:"provider" toml:"provider"`

	// Model is the model identifier (e.g., "gpt-4o", "claude-3-opus")
	Model string `yaml:"model" json:"model" toml:"model"`

	// Temperature controls randomness (0.0 to 2.0)
	Temperature *float64 `yaml:"temperature" json:"temperature" toml:"temperature"`

	// MaxTokens is the maximum tokens in the response
	MaxTokens *int `yaml:"max_tokens" json:"max_tokens" toml:"max_tokens"`

	// TopP controls diversity via nucleus sampling
	TopP *float64 `yaml:"top_p" json:"top_p" toml:"top_p"`

	// FrequencyPenalty reduces repetition (0.0 to 2.0)
	FrequencyPenalty *float64 `yaml:"frequency_penalty" json:"frequency_penalty" toml:"frequency_penalty"`

	// PresencePenalty encourages new topics (0.0 to 2.0)
	PresencePenalty *float64 `yaml:"presence_penalty" json:"presence_penalty" toml:"presence_penalty"`

	// StopSequences are sequences where generation stops
	StopSequences []string `yaml:"stop_sequences" json:"stop_sequences" toml:"stop_sequences"`

	// BaseURL is the API base URL (for custom endpoints)
	BaseURL string `yaml:"base_url" json:"base_url" toml:"base_url"`

	// APIKey is the API key (can also be set via environment)
	APIKey string `yaml:"api_key" json:"api_key" toml:"api_key"`
}

// SkillsConfig defines allowed and denied skills for an agent.
// Skills use a wildcard pattern (e.g., "file.*" matches "file.read", "file.write").
type SkillsConfig struct {
	// Allowed is the whitelist of permitted skills
	Allowed []string `yaml:"allowed" json:"allowed" toml:"allowed"`

	// Denied is the blacklist of forbidden skills
	Denied []string ` json:"denied" yaml:"denied" toml:"denied"`
}

// AgentConfig holds configuration for a specific agent.
// Values here override global defaults.
type AgentConfig struct {
	// Name is the unique identifier for the agent
	Name string `yaml:"-" json:"-" toml:"-"`

	// LLM contains LLM-specific overrides
	LLM *LLMConfig `yaml:"llm" json:"llm" toml:"llm"`

	// Skills defines what capabilities the agent has
	Skills *SkillsConfig `yaml:"skills" json:"skills" toml:"skills"`

	// Timeout is the maximum execution time in seconds
	Timeout *int `yaml:"timeout" json:"timeout" toml:"timeout"`

	// MaxRetries is the number of retry attempts
	MaxRetries *int `yaml:"max_retries" json:"max_retries" toml:"max_retries"`

	// MaxConcurrentTasks limits parallel task execution
	MaxConcurrentTasks *int `yaml:"max_concurrent_tasks" json:"max_concurrent_tasks" toml:"max_concurrent_tasks"`

	// Debug enables verbose logging for this agent
	Debug *bool `yaml:"debug" json:"debug" toml:"debug"`

	// Enabled determines if the agent is active
	Enabled *bool `yaml:"enabled" json:"enabled" toml:"enabled"`

	// Metadata contains arbitrary additional configuration
	Metadata map[string]interface{} `yaml:"metadata" json:"metadata" toml:"metadata"`
}

// GlobalConfig holds global defaults that apply to all agents.
// Agents can override these values in their specific configuration.
type GlobalConfig struct {
	// LLM contains default LLM settings
	LLM *LLMConfig `yaml:"llm" json:"llm" toml:"llm"`

	// Timeout is the default timeout in seconds
	Timeout int `yaml:"timeout" json:"timeout" toml:"timeout"`

	// MaxRetries is the default number of retries
	MaxRetries int `yaml:"max_retries" json:"max_retries" toml:"max_retries"`

	// Debug enables verbose logging globally
	Debug bool `yaml:"debug" json:"debug" toml:"debug"`

	// MaxConcurrentTasks is the default concurrency limit
	MaxConcurrentTasks int `yaml:"max_concurrent_tasks" json:"max_concurrent_tasks" toml:"max_concurrent_tasks"`
}

// Config is the root configuration structure.
// It contains global defaults and per-agent configurations.
type Config struct {
	// Global holds default settings for all agents
	Global GlobalConfig `yaml:"global" json:"global" toml:"global"`

	// Agents contains agent-specific configurations
	Agents map[string]*AgentConfig `yaml:"agents" json:"agents" toml:"agents"`
}

// NewConfig creates a new Config with sensible defaults.
func NewConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			Timeout:            300,
			MaxRetries:         3,
			Debug:              false,
			MaxConcurrentTasks: 5,
			LLM: &LLMConfig{
				Provider:    "openai",
				Model:       "gpt-4o",
				Temperature: ptrFloat64(0.7),
				MaxTokens:   ptrInt(2000),
			},
		},
		Agents: make(map[string]*AgentConfig),
	}
}

// MergeWithGlobal creates a merged AgentConfig with global defaults applied.
// Agent-specific values take precedence over global defaults.
func (a *AgentConfig) MergeWithGlobal(global *GlobalConfig) *AgentConfig {
	merged := &AgentConfig{
		Name:     a.Name,
		Metadata: a.Metadata,
	}

	// Merge LLM config
	if a.LLM != nil {
		merged.LLM = mergeLLMConfig(global.LLM, a.LLM)
	} else {
		merged.LLM = global.LLM
	}

	// Merge skills
	merged.Skills = a.Skills

	// Merge scalar values (agent overrides global)
	if a.Timeout != nil {
		merged.Timeout = a.Timeout
	} else {
		merged.Timeout = &global.Timeout
	}

	if a.MaxRetries != nil {
		merged.MaxRetries = a.MaxRetries
	} else {
		merged.MaxRetries = &global.MaxRetries
	}

	if a.MaxConcurrentTasks != nil {
		merged.MaxConcurrentTasks = a.MaxConcurrentTasks
	} else {
		merged.MaxConcurrentTasks = &global.MaxConcurrentTasks
	}

	if a.Debug != nil {
		merged.Debug = a.Debug
	} else {
		merged.Debug = &global.Debug
	}

	merged.Enabled = a.Enabled

	return merged
}

// mergeLLMConfig merges two LLM configurations, with override taking precedence.
func mergeLLMConfig(base *LLMConfig, override *LLMConfig) *LLMConfig {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := &LLMConfig{
		Provider:         base.Provider,
		Model:            base.Model,
		BaseURL:          base.BaseURL,
		APIKey:           base.APIKey,
		Temperature:      base.Temperature,
		MaxTokens:        base.MaxTokens,
		TopP:             base.TopP,
		FrequencyPenalty: base.FrequencyPenalty,
		PresencePenalty:  base.PresencePenalty,
		StopSequences:    base.StopSequences,
	}

	// Apply overrides
	if override.Provider != "" {
		merged.Provider = override.Provider
	}
	if override.Model != "" {
		merged.Model = override.Model
	}
	if override.BaseURL != "" {
		merged.BaseURL = override.BaseURL
	}
	if override.APIKey != "" {
		merged.APIKey = override.APIKey
	}
	if override.Temperature != nil {
		merged.Temperature = override.Temperature
	}
	if override.MaxTokens != nil {
		merged.MaxTokens = override.MaxTokens
	}
	if override.TopP != nil {
		merged.TopP = override.TopP
	}
	if override.FrequencyPenalty != nil {
		merged.FrequencyPenalty = override.FrequencyPenalty
	}
	if override.PresencePenalty != nil {
		merged.PresencePenalty = override.PresencePenalty
	}
	if len(override.StopSequences) > 0 {
		merged.StopSequences = override.StopSequences
	}

	return merged
}

// Helper functions for creating pointers
func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt(v int) *int {
	return &v
}

func ptrBool(v bool) *bool {
	return &v
}
