// Package main demonstrates concurrent requests with go-llm-client
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
	client := llm.NewClient(llm.Config{
		Provider:   llm.ProviderOpenAI,
		Model:      "gpt-4o",
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		MaxRetries: 3,
	})

	llm.RegisterOpenAI(client, llm.OpenAIConfig{
		APIKey:     os.Getenv("OPENAI_API_KEY"),
		MaxRetries: 3,
	})

	// Make multiple concurrent requests
	prompts := []string{
		"What is 2+2?",
		"What is the capital of Japan?",
		"Name a primary color",
		"What planet do we live on?",
		"Name a popular programming language",
	}

	var wg sync.WaitGroup
	results := make([]string, len(prompts))
	errors := make([]error, len(prompts))

	start := time.Now()

	for i, prompt := range prompts {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := client.Simple(ctx, p)
			if err != nil {
				errors[idx] = err
				return
			}
			results[idx] = resp
		}(i, prompt)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Print results
	fmt.Printf("Completed %d requests in %v\n\n", len(prompts), elapsed)

	for i, prompt := range prompts {
		fmt.Printf("Q%d: %s\n", i+1, prompt)
		if errors[i] != nil {
			fmt.Printf("   Error: %v\n", errors[i])
		} else {
			fmt.Printf("   A: %s\n", results[i])
		}
		fmt.Println()
	}
}
