package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Helper for errors.As
var _ = errors.As

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err    error
	Status int
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// OpenAIProvider implements the ProviderClient interface for OpenAI
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	client     *http.Client
	debug      bool
	maxRetries int
}

// OpenAIConfig is configuration for OpenAI provider
type OpenAIConfig struct {
	APIKey     string
	BaseURL    string // Optional, defaults to https://api.openai.com/v1
	Timeout    int    // Timeout in seconds
	MaxRetries int    // Max retry attempts (default: 3)
	Debug      bool
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config OpenAIConfig) *OpenAIProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &OpenAIProvider{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		debug:      config.Debug,
		maxRetries: maxRetries,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// openAIRequest represents the request body for OpenAI API
type openAIRequest struct {
	Model            string        `json:"model"`
	Messages         []Message     `json:"messages"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	Temperature      float64       `json:"temperature,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
}

// openAIResponse represents the response from OpenAI API
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		Delta        Message `json:"delta,omitempty"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Complete sends a completion request to OpenAI with retry logic
func (p *OpenAIProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, etc.
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if p.debug {
				fmt.Printf("[OpenAI] Retry attempt %d after %v\n", attempt, backoff)
			}

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := p.doRequest(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.shouldRetry(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries (%d) exceeded, last error: %w", p.maxRetries, lastErr)
}

// doRequest performs a single HTTP request without retry
func (p *OpenAIProvider) doRequest(ctx context.Context, req Request) (*Response, error) {
	// Build request body
	model := req.Model
	if model == "" {
		model = "gpt-4o" // Default model
	}

	openAIReq := openAIRequest{
		Model:            model,
		Messages:         req.Messages,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stop:             req.Stop,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
		Stream:           false,
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	if p.debug {
		fmt.Printf("[OpenAI] Request: %s\n", string(body))
	}

	// Send request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, &RetryableError{Err: err}
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if p.debug {
		fmt.Printf("[OpenAI] Response (status %d): %s\n", httpResp.StatusCode, string(respBody))
	}

	// Check for retryable HTTP status codes
	if p.shouldRetryHTTP(httpResp.StatusCode) {
		return nil, &RetryableError{
			Err:    fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody)),
			Status: httpResp.StatusCode,
		}
	}

	// Parse response
	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error and convert to appropriate error type
	if openAIResp.Error != nil {
		return nil, p.convertAPIError(httpResp.StatusCode, openAIResp.Error)
	}

	// Convert to our response format
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openAIResp.Choices[0]
	return &Response{
		ID:      openAIResp.ID,
		Model:   openAIResp.Model,
		Content: choice.Message.Content,
		Usage:   openAIResp.Usage,
		Choices: []Choice{
			{
				Index:        choice.Index,
				Message:      choice.Message,
				FinishReason: choice.FinishReason,
			},
		},
	}, nil
}

// convertAPIError converts OpenAI API errors to our error types
func (p *OpenAIProvider) convertAPIError(statusCode int, apiErr *struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}) error {
	errCode := apiErr.Code
	if errCode == "" {
		errCode = apiErr.Type
	}

	switch errCode {
	case "invalid_api_key", "authentication_error":
		return &AuthenticationError{
			Provider: "openai",
			Message:  apiErr.Message,
		}
	case "rate_limit_exceeded", "insufficient_quota":
		return &RateLimitError{
			Provider: "openai",
			Message:  apiErr.Message,
		}
	case "model_not_found", "invalid_model":
		return &ModelNotFoundError{
			Provider: "openai",
			Model:    "", // Could extract from request
		}
	case "context_length_exceeded":
		return &ContextLengthExceededError{
			Provider: "openai",
			Message:  apiErr.Message,
		}
	}

	// Fallback to generic error
	return fmt.Errorf("OpenAI API error: %s (code: %s)", apiErr.Message, errCode)
}

// shouldRetry checks if an error is retryable
func (p *OpenAIProvider) shouldRetry(err error) bool {
	var retryErr *RetryableError
	if errors.As(err, &retryErr) {
		return true
	}
	return false
}

// shouldRetryHTTP checks if an HTTP status code is retryable
func (p *OpenAIProvider) shouldRetryHTTP(statusCode int) bool {
	// Retry on: 429 (rate limit), 500, 502, 503, 504
	retryableCodes := map[int]bool{
		429: true, // Too Many Requests
		500: true, // Internal Server Error
		502: true, // Bad Gateway
		503: true, // Service Unavailable
		504: true, // Gateway Timeout
	}
	return retryableCodes[statusCode]
}

// Stream sends a streaming completion request to OpenAI
func (p *OpenAIProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	// Build request body
	model := req.Model
	if model == "" {
		model = "gpt-4o"
	}

	openAIReq := openAIRequest{
		Model:            model,
		Messages:         req.Messages,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stop:             req.Stop,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
		Stream:           true,
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Start goroutine to read stream
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Check for data prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				ch <- StreamChunk{Finished: true}
				return
			}

			// Parse the chunk
			var openAIResp openAIResponse
			if err := json.Unmarshal([]byte(data), &openAIResp); err != nil {
				if p.debug {
					fmt.Printf("[OpenAI] Failed to parse chunk: %v\n", err)
				}
				continue
			}

			// Check for error
			if openAIResp.Error != nil {
				if p.debug {
					fmt.Printf("[OpenAI] Stream error: %s\n", openAIResp.Error.Message)
				}
				return
			}

			// Send chunk
			if len(openAIResp.Choices) > 0 {
				choice := openAIResp.Choices[0]
				ch <- StreamChunk{
					Delta:    choice.Delta,
					Finished: choice.FinishReason != "",
				}

				if choice.FinishReason != "" {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if p.debug {
				fmt.Printf("[OpenAI] Scanner error: %v\n", err)
			}
		}
	}()

	return ch, nil
}

// CountTokens counts tokens using OpenAI's tokenizer
// This is a rough approximation. For accurate counting, use tiktoken
func (p *OpenAIProvider) CountTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	// This is not accurate but works for estimation
	return len(text) / 4
}
