package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			ID:    "test",
			Model: "gpt-4o",
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				Delta        Message `json:"delta,omitempty"`
				FinishReason string  `json:"finish_reason"`
			}{
				{Index: 0, Message: Message{Role: "assistant", Content: "OK"}},
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test",
		BaseURL: server.URL,
	})

	// Empty messages should not panic
	resp, err := provider.Complete(context.Background(), Request{
		Messages: []Message{},
	})

	if err != nil {
		t.Errorf("Unexpected error with empty messages: %v", err)
	}

	if resp.Content != "OK" {
		t.Errorf("Expected 'OK', got '%s'", resp.Content)
	}
}

func TestInvalidAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Incorrect API key provided",
				"type":    "invalid_request_error",
				"code":    "invalid_api_key",
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "invalid",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Fatal("Expected error for invalid API key")
	}

	if !IsAuthenticationError(err) {
		t.Errorf("Expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "The model 'gpt-5' does not exist",
				"type":    "invalid_request_error",
				"code":    "model_not_found",
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Model:    "gpt-5",
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Fatal("Expected error for model not found")
	}

	if !IsModelNotFoundError(err) {
		t.Errorf("Expected ModelNotFoundError, got %T: %v", err, err)
	}
}

func TestContextLengthExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "This model's maximum context length is 4096 tokens",
				"type":    "invalid_request_error",
				"code":    "context_length_exceeded",
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "very long message"}},
	})

	if err == nil {
		t.Fatal("Expected error for context length exceeded")
	}

	if !IsContextLengthError(err) {
		t.Errorf("Expected ContextLengthExceededError, got %T: %v", err, err)
	}
}

func TestMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json {")) // Malformed JSON
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Fatal("Expected error for malformed response")
	}
}

func TestNoChoicesInResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			ID:    "test",
			Model: "gpt-4o",
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				Delta        Message `json:"delta,omitempty"`
				FinishReason string  `json:"finish_reason"`
			}{}, // Empty choices
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Fatal("Expected error for no choices")
	}
}

func TestProviderUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Temporarily Unavailable"))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:     "test",
		BaseURL:    server.URL,
		MaxRetries: 0, // No retries for faster test
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Fatal("Expected error for service unavailable")
	}
}

func TestRetryAfterRateLimit(t *testing.T) {
	attempts := 0
	retryAfter := 2 // seconds

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts < 2 {
			// Set Retry-After header
			w.Header().Set("Retry-After", string(rune(retryAfter)))
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			ID:    "test",
			Model: "gpt-4o",
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				Delta        Message `json:"delta,omitempty"`
				FinishReason string  `json:"finish_reason"`
			}{
				{Index: 0, Message: Message{Role: "assistant", Content: "Success"}},
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:     "test",
		BaseURL:    server.URL,
		MaxRetries: 3,
	})

	resp, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resp.Content != "Success" {
		t.Errorf("Expected 'Success', got '%s'", resp.Content)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}
