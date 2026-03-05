// Package main demonstrates basic usage of the go-llm-client library
package main

import (
	"context"
	"fmt"
	"os"

	llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
	// Create client with default configuration
	client := llm.NewClient(llm.Config{
		Provider: llm.ProviderOpenAI,
		Model:    "gpt-4o",
		APIKey:   os.Getenv("OPENAI_API_KEY"),
	})

	// Register the OpenAI provider
	llm.RegisterOpenAI(client, llm.OpenAIConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})

	// Simple completion
	response, err := client.Simple(context.Background(), "What is the capital of France?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", response)
}
