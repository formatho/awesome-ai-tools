package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetryOnRateLimit(t *testing.T) {
	attempts := 0

	// Create mock server that returns 429 twice, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts < 3 {
			// Return 429 (rate limit) for first 2 attempts
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}

		// Third attempt succeeds
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			ID:    "test-123",
			Model: "gpt-4o",
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				Delta        Message `json:"delta,omitempty"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Index:   0,
					Message: Message{Role: "assistant", Content: "Hello!"},
				},
			},
			Usage: Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		})
	}))
	defer server.Close()

	// Create provider with low timeout for faster test
	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		MaxRetries: 3,
	})

	// Make request
	resp, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if resp.Content != "Hello!" {
		t.Errorf("Expected 'Hello!', got '%s'", resp.Content)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryOnServerError(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			ID:    "test-456",
			Model: "gpt-4o",
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				Delta        Message `json:"delta,omitempty"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Index:   0,
					Message: Message{Role: "assistant", Content: "Success"},
				},
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		MaxRetries: 3,
	})

	resp, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	if err != nil {
		t.Fatalf("Expected success, got: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	if resp.Content != "Success" {
		t.Errorf("Expected 'Success', got '%s'", resp.Content)
	}
}

func TestMaxRetriesExceeded(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		MaxRetries: 2, // Will try 3 times total (initial + 2 retries)
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	if err == nil {
		t.Fatal("Expected error when max retries exceeded")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if err.Error() != "max retries (2) exceeded, last error: HTTP 503: " {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestNoRetryOnClientError(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest) // 400 - should not retry
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid request",
				"type":    "invalid_request_error",
				"code":    "invalid_api_key",
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		MaxRetries: 3,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	if err == nil {
		t.Fatal("Expected error for bad request")
	}

	// Should only try once (no retry)
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry on 400), got %d", attempts)
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Slow response
		w.WriteHeader(http.StatusOK)
	}))
	// Close server in background to avoid blocking
	go func() {
		time.Sleep(500 * time.Millisecond)
		server.Close()
	}()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Timeout: 5, // Long timeout
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := provider.Complete(ctx, Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	if err == nil {
		t.Fatal("Expected context cancellation error")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
	}
}

func TestRetryableError(t *testing.T) {
	err := &RetryableError{
		Err:    fmt.Errorf("test error"),
		Status: 429,
	}

	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}

	if err.Status != 429 {
		t.Errorf("Expected status 429, got %d", err.Status)
	}
}

func TestShouldRetryHTTP(t *testing.T) {
	provider := &OpenAIProvider{}

	retryable := []int{429, 500, 502, 503, 504}
	for _, code := range retryable {
		if !provider.shouldRetryHTTP(code) {
			t.Errorf("Expected status %d to be retryable", code)
		}
	}

	nonRetryable := []int{200, 400, 401, 403, 404}
	for _, code := range nonRetryable {
		if provider.shouldRetryHTTP(code) {
			t.Errorf("Expected status %d to NOT be retryable", code)
		}
	}
}
