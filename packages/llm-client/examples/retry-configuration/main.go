// Package main demonstrates custom retry configuration
package main

import (
	"context"
	"fmt"
	"os"

	llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
	// Create client with custom retry configuration
	client := llm.NewClient(llm.Config{
		Provider:   llm.ProviderOpenAI,
		Model:      "gpt-4o",
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		MaxRetries: 5, // Try 6 times total (1 initial + 5 retries)
	})

	// Register with matching config
	llm.RegisterOpenAI(client, llm.OpenAIConfig{
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		MaxRetries: 5,
		Debug:      true, // Enable debug to see retry attempts
	})

	// Make a request with built-in retry
	response, err := client.Simple(context.Background(), "Hello!")
	if err != nil {
		fmt.Printf("Error after all retries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", response)

	// You can also override retry behavior per-request
	response, err = client.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "user", Content: "Hi"},
		},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", response.Content)
}
