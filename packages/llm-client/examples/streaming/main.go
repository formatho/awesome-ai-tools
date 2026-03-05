// Package main demonstrates streaming completion with go-llm-client
package main

import (
	"context"
	"fmt"
	"os"

	llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
	client := llm.NewClient(llm.Config{
		Provider: llm.ProviderOpenAI,
		Model:    "gpt-4o",
		APIKey:   os.Getenv("OPENAI_API_KEY"),
	})

	llm.RegisterOpenAI(client, llm.OpenAIConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})

	// Stream a long response
	stream, err := client.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Tell me a short story about a robot learning to paint"},
		},
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Streaming response:")
	fmt.Println("---")

	for chunk := range stream {
		fmt.Print(chunk.Delta.Content)
		if chunk.Finished {
			break
		}
	}

	fmt.Println("\n---")
	fmt.Println("Stream complete!")
}
