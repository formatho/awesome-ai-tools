// Package main demonstrates proper error handling with go-llm-client
package main

import (
	"context"
	"fmt"
	"os"

	llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
	client := llm.NewClient(llm.Config{
		Provider:   llm.ProviderOpenAI,
		Model:      "gpt-4o",
		APIKey:     "invalid-key", // Intentionally invalid
		MaxRetries: 2,
	})

	llm.RegisterOpenAI(client, llm.OpenAIConfig{
		APIKey: "invalid-key",
	})

	// Try a completion
	_, err := client.Simple(context.Background(), "Hello")
	if err != nil {
		// Check error type
		if llm.IsAuthenticationError(err) {
			fmt.Println("❌ Authentication failed - check your API key")
		} else if llm.IsRateLimitError(err) {
			fmt.Println("⏳ Rate limit exceeded - please wait and retry")
		} else if llm.IsModelNotFoundError(err) {
			fmt.Println("🔍 Model not found - check model name")
		} else if llm.IsContextLengthError(err) {
			fmt.Println("📏 Context too long - reduce message size")
		} else if llm.IsRetryable(err) {
			fmt.Println("🔄 Transient error - will retry automatically")
		} else {
			fmt.Printf("❓ Unknown error: %v\n", err)
		}

		os.Exit(1)
	}

	fmt.Println("Success!")
}
