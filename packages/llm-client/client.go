// Package llm provides a unified interface for multiple LLM providers.
//
// This library offers a simple, consistent API for working with different LLM providers
// (OpenAI, Anthropic, Ollama, etc.) with built-in retry logic, streaming support, and
// comprehensive error handling.
//
// # Quick Start
//
// Create a client and make a simple completion:
//
//	client := llm.NewClient(llm.Config{
//	    Provider: llm.ProviderOpenAI,
//	    Model:    "gpt-4o",
//	    APIKey:   os.Getenv("OPENAI_API_KEY"),
//	})
//
//	llm.RegisterOpenAI(client, llm.OpenAIConfig{
//	    APIKey: os.Getenv("OPENAI_API_KEY"),
//	})
//
//	response, err := client.Simple(context.Background(), "Hello!")
//
// # Streaming
//
// For real-time token streaming:
//
//	stream, err := client.Stream(ctx, llm.Request{
//	    Messages: []llm.Message{
//	        {Role: "user", Content: "Tell me a story"},
//	    },
//	})
//
//	for chunk := range stream {
//	    fmt.Print(chunk.Delta.Content)
//	}
//
// # Error Handling
//
// The library provides typed errors for better error handling:
//
//	if llm.IsRateLimitError(err) {
//	    // Handle rate limiting
//	}
//
// # Retry Logic
//
// Automatic retry with exponential backoff for transient failures:
//
//	client := llm.NewClient(llm.Config{
//	    Provider:   llm.ProviderOpenAI,
//	    Model:      "gpt-4o",
//	    APIKey:     os.Getenv("OPENAI_API_KEY"),
//	    MaxRetries: 3, // Retry 3 times (4 total attempts)
//	})
package llm

import (
	"context"
	"fmt"
	"io"
)

// Provider represents an LLM provider (OpenAI, Anthropic, Ollama, etc.)
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderOllama    Provider = "ollama"
	ProviderLocal     Provider = "local"
	ProviderZAI       Provider = "zai"
)

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`           // "system", "user", "assistant"
	Content string `json:"content"`        // Message content
	Name    string `json:"name,omitempty"` // Optional name for function calling
}

// Request represents a completion request
type Request struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`    // Override default model
	Provider    Provider  `json:"provider,omitempty"` // Override default provider
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`

	// Streaming
	Stream bool `json:"stream,omitempty"`

	// Advanced
	TopP             float64  `json:"top_p,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
}

// Response represents a completion response
type Response struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Content string   `json:"content"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice represents a single completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Delta    Message `json:"delta"`
	Finished bool    `json:"finished"`
	Error    error   `json:"-"` // Stream error, if any
}

// Config represents client configuration
type Config struct {
	Provider Provider `json:"provider"`
	Model    string   `json:"model"`
	APIKey   string   `json:"api_key"`
	BaseURL  string   `json:"base_url,omitempty"` // For custom endpoints

	// Retry configuration
	MaxRetries int `json:"max_retries,omitempty"`
	Timeout    int `json:"timeout,omitempty"` // seconds

	// Logging
	Debug bool `json:"debug,omitempty"`
}

// Client is the main LLM client interface
type Client struct {
	config    Config
	providers map[Provider]ProviderClient
}

// ProviderClient is the interface that all providers must implement
type ProviderClient interface {
	Complete(ctx context.Context, req Request) (*Response, error)
	Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)
	CountTokens(text string) int
}

// NewClient creates a new LLM client
func NewClient(config Config) *Client {
	return &Client{
		config:    config,
		providers: make(map[Provider]ProviderClient),
	}
}

// Complete sends a completion request
func (c *Client) Complete(ctx context.Context, req Request) (*Response, error) {
	provider := req.Provider
	if provider == "" {
		provider = c.config.Provider
	}

	client, ok := c.providers[provider]
	if !ok {
		// Initialize provider on first use
		if err := c.initProvider(provider); err != nil {
			return nil, err
		}
		client = c.providers[provider]
	}

	return client.Complete(ctx, req)
}

// Stream sends a streaming completion request
func (c *Client) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	req.Stream = true

	provider := req.Provider
	if provider == "" {
		provider = c.config.Provider
	}

	client, ok := c.providers[provider]
	if !ok {
		if err := c.initProvider(provider); err != nil {
			return nil, err
		}
		client = c.providers[provider]
	}

	return client.Stream(ctx, req)
}

// CountTokens counts tokens in the given text
func (c *Client) CountTokens(text string) int {
	client, ok := c.providers[c.config.Provider]
	if !ok {
		// Use default tokenizer if provider not initialized
		return len(text) / 4 // Rough estimate
	}
	return client.CountTokens(text)
}

// initProvider initializes a provider client
func (c *Client) initProvider(provider Provider) error {
	// Provider must be registered using RegisterOpenAI, RegisterAnthropic, etc.
	// or set explicitly using SetProvider
	return fmt.Errorf("provider %s not registered - use RegisterOpenAI() or SetProvider() to initialize", provider)
}

// SetProvider sets a custom provider implementation
func (c *Client) SetProvider(provider Provider, client ProviderClient) {
	c.providers[provider] = client
}

// Simple is a convenience method for simple completions
func (c *Client) Simple(ctx context.Context, prompt string) (string, error) {
	resp, err := c.Complete(ctx, Request{
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// SimpleStream is a convenience method for simple streaming completions
func (c *Client) SimpleStream(ctx context.Context, prompt string, writer io.Writer) error {
	ch, err := c.Stream(ctx, Request{
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return err
	}

	for chunk := range ch {
		if chunk.Error != nil {
			return chunk.Error
		}
		if _, err := writer.Write([]byte(chunk.Delta.Content)); err != nil {
			return err
		}
		if chunk.Finished {
			break
		}
	}
	return nil
}
