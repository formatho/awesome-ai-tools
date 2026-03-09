package llm

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestZAIProvider_Complete(t *testing.T) {
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY not set")
	}

	provider := NewZAIProvider(ZAIConfig{
		APIKey: apiKey,
		Debug:  true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Complete(ctx, Request{
		Model: "glm-4.7",
		Messages: []Message{
			{Role: "user", Content: "Say 'Connection successful' in exactly those words."},
		},
		MaxTokens: 20,
	})

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Expected non-empty content")
	}

	t.Logf("Response: %s", resp.Content)
	t.Logf("Model: %s", resp.Model)
	t.Logf("Usage: %+v", resp.Usage)
}

func TestZAIProvider_Complete_WithCustomEndpoint(t *testing.T) {
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		t.Skip("ZAI_API_KEY not set")
	}

	provider := NewZAIProvider(ZAIConfig{
		APIKey:  apiKey,
		BaseURL: "https://api.z.ai/api/coding/paas/v4",
		Debug:   true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Complete(ctx, Request{
		Model: "glm-4.7",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 50,
	})

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Expected non-empty content")
	}

	t.Logf("Response: %s", resp.Content)
}

func TestZAIProvider_DefaultBaseURL(t *testing.T) {
	provider := NewZAIProvider(ZAIConfig{
		APIKey: "test-key",
	})

	expected := "https://api.z.ai/api/coding/paas/v4"
	if provider.baseURL != expected {
		t.Errorf("Expected baseURL %s, got %s", expected, provider.baseURL)
	}
}

func TestZAIProvider_CustomBaseURL(t *testing.T) {
	customURL := "https://custom.endpoint.com/v1"
	provider := NewZAIProvider(ZAIConfig{
		APIKey:  "test-key",
		BaseURL: customURL,
	})

	if provider.baseURL != customURL {
		t.Errorf("Expected baseURL %s, got %s", customURL, provider.baseURL)
	}
}

func TestZAIProvider_Defaults(t *testing.T) {
	provider := NewZAIProvider(ZAIConfig{
		APIKey: "test-key",
	})

	if provider.client.Timeout != 60*time.Second {
		t.Errorf("Expected default timeout 60s, got %v", provider.client.Timeout)
	}

	if provider.maxRetries != 3 {
		t.Errorf("Expected default maxRetries 3, got %d", provider.maxRetries)
	}
}
