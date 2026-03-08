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

// ZAIProvider implements the ProviderClient interface for z.ai (GLM)
type ZAIProvider struct {
	apiKey     string
	baseURL    string
	client     *http.Client
	debug      bool
	maxRetries int
}

// ZAIConfig is configuration for z.ai provider
type ZAIConfig struct {
	APIKey     string
	BaseURL    string // Optional, defaults to https://open.bigmodel.cn/api/paas/v4
	Timeout    int    // Timeout in seconds
	MaxRetries int    // Max retry attempts (default: 3)
	Debug      bool
}

// NewZAIProvider creates a new z.ai provider
func NewZAIProvider(config ZAIConfig) *ZAIProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &ZAIProvider{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		debug:      config.Debug,
		maxRetries: maxRetries,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// zaiRequest represents the request body for z.ai API
type zaiRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
}

// zaiResponse represents the response from z.ai API
type zaiResponse struct {
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

// Complete sends a completion request to z.ai with retry logic
func (p *ZAIProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, etc.
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if p.debug {
				fmt.Printf("[ZAI] Retry attempt %d after %v\n", attempt, backoff)
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
func (p *ZAIProvider) doRequest(ctx context.Context, req Request) (*Response, error) {
	// Build request body
	model := req.Model
	if model == "" {
		model = "glm-4" // Default model
	}

	zaiReq := zaiRequest{
		Model:            model,
		Messages:         req.Messages,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stop:             req.Stop,
		Stream:           false,
	}

	body, err := json.Marshal(zaiReq)
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
		fmt.Printf("[ZAI] Request: %s\n", string(body))
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
		fmt.Printf("[ZAI] Response (status %d): %s\n", httpResp.StatusCode, string(respBody))
	}

	// Check for retryable HTTP status codes
	if p.shouldRetryHTTP(httpResp.StatusCode) {
		return nil, &RetryableError{
			Err:    fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(respBody)),
			Status: httpResp.StatusCode,
		}
	}

	// Parse response
	var zaiResp zaiResponse
	if err := json.Unmarshal(respBody, &zaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error and convert to appropriate error type
	if zaiResp.Error != nil {
		return nil, p.convertAPIError(httpResp.StatusCode, zaiResp.Error)
	}

	// Convert to our response format
	if len(zaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := zaiResp.Choices[0]
	return &Response{
		ID:      zaiResp.ID,
		Model:   zaiResp.Model,
		Content: choice.Message.Content,
		Usage:   zaiResp.Usage,
		Choices: []Choice{
			{
				Index:        choice.Index,
				Message:      choice.Message,
				FinishReason: choice.FinishReason,
			},
		},
	}, nil
}

// convertAPIError converts z.ai API errors to our error types
func (p *ZAIProvider) convertAPIError(statusCode int, apiErr *struct {
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
			Provider: "zai",
			Message:  apiErr.Message,
		}
	case "rate_limit_exceeded", "insufficient_quota":
		return &RateLimitError{
			Provider: "zai",
			Message:  apiErr.Message,
		}
	case "model_not_found", "invalid_model":
		return &ModelNotFoundError{
			Provider: "zai",
			Model:    "",
		}
	case "context_length_exceeded":
		return &ContextLengthExceededError{
			Provider: "zai",
			Message:  apiErr.Message,
		}
	}

	// Fallback to generic error
	return fmt.Errorf("ZAI API error: %s (code: %s)", apiErr.Message, errCode)
}

// shouldRetry checks if an error is retryable
func (p *ZAIProvider) shouldRetry(err error) bool {
	var retryErr *RetryableError
	return errors.As(err, &retryErr)
}

// shouldRetryHTTP checks if an HTTP status code is retryable
func (p *ZAIProvider) shouldRetryHTTP(statusCode int) bool {
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

// Stream sends a streaming completion request to z.ai
func (p *ZAIProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	// Build request body
	model := req.Model
	if model == "" {
		model = "glm-4"
	}

	zaiReq := zaiRequest{
		Model:            model,
		Messages:         req.Messages,
		MaxTokens:        req.MaxTokens,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		Stop:             req.Stop,
		Stream:           true,
	}

	body, err := json.Marshal(zaiReq)
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

	// Check HTTP status before streaming
	if resp.StatusCode >= 400 {
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		close(ch)
		if err != nil {
			return nil, fmt.Errorf("HTTP %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Start goroutine to read stream
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				select {
				case ch <- StreamChunk{Error: ctx.Err(), Finished: true}:
				default:
				}
				return
			default:
			}

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
				select {
				case ch <- StreamChunk{Finished: true}:
				case <-ctx.Done():
				}
				return
			}

			// Parse the chunk
			var zaiResp zaiResponse
			if err := json.Unmarshal([]byte(data), &zaiResp); err != nil {
				if p.debug {
					fmt.Printf("[ZAI] Failed to parse chunk: %v\n", err)
				}
				continue
			}

			// Check for error
			if zaiResp.Error != nil {
				if p.debug {
					fmt.Printf("[ZAI] Stream error: %s\n", zaiResp.Error.Message)
				}
				select {
				case ch <- StreamChunk{Error: fmt.Errorf("stream error: %s", zaiResp.Error.Message), Finished: true}:
				case <-ctx.Done():
				}
				return
			}

			// Send chunk
			if len(zaiResp.Choices) > 0 {
				choice := zaiResp.Choices[0]
				chunk := StreamChunk{
					Delta:    choice.Delta,
					Finished: choice.FinishReason != "",
				}

				select {
				case ch <- chunk:
				case <-ctx.Done():
					return
				}

				if choice.FinishReason != "" {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if p.debug {
				fmt.Printf("[ZAI] Scanner error: %v\n", err)
			}
			select {
			case ch <- StreamChunk{Error: err, Finished: true}:
			case <-ctx.Done():
			}
		}
	}()

	return ch, nil
}

// CountTokens counts tokens using z.ai's tokenizer
// This is a rough approximation. For accurate counting, use tiktoken
func (p *ZAIProvider) CountTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	// This is not accurate but works for estimation
	return len(text) / 4
}
