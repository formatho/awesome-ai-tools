package models

import (
	"time"
)

// Config represents the global system configuration.
type Config struct {
	ID        string                 `json:"id"`
	LLMConfig *LLMConfig             `json:"llm,omitempty"`
	Defaults  map[string]interface{} `json:"defaults,omitempty"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// LLMConfig holds LLM provider configuration.
type LLMConfig struct {
	Provider         string   `json:"provider"`
	Model            string   `json:"model,omitempty"`
	APIKey           string   `json:"api_key,omitempty"`
	BaseURL          string   `json:"base_url,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	StopSequences    []string `json:"stop_sequences,omitempty"`
}

// ConfigUpdate is the request body for updating configuration.
type ConfigUpdate struct {
	LLMConfig *LLMConfig             `json:"llm,omitempty"`
	Defaults  map[string]interface{} `json:"defaults,omitempty"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
}

// LLMTestRequest is the request to test LLM connection.
type LLMTestRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"model,omitempty"`
}

// LLMTestResponse is the result of LLM connection test.
type LLMTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency_ms,omitempty"`
}
