package llm

import (
	"context"
	"fmt"
	"io"

	gollm "github.com/teilomillet/gollm"
)

// GollmProvider wraps the gollm library to implement ProviderClient interface
type GollmProvider struct {
	llm    gollm.LLM
	config GollmConfig
}

// GollmConfig is configuration for gollm provider
type GollmConfig struct {
	Provider string // "openai", "anthropic", "ollama", "groq", "mistral", "openrouter"
	Model    string
	APIKey   string
	BaseURL  string // For custom endpoints (e.g., Ollama)

	// Options
	MaxTokens   int
	Temperature float64
	MaxRetries  int
	Debug       bool
}

// NewGollmProvider creates a new provider using gollm library
func NewGollmProvider(config GollmConfig) (*GollmProvider, error) {
	opts := []gollm.ConfigOption{
		gollm.SetProvider(config.Provider),
	}

	if config.Model != "" {
		opts = append(opts, gollm.SetModel(config.Model))
	}

	if config.APIKey != "" {
		opts = append(opts, gollm.SetAPIKey(config.APIKey))
	}

	if config.MaxTokens > 0 {
		opts = append(opts, gollm.SetMaxTokens(config.MaxTokens))
	}

	if config.Temperature > 0 {
		opts = append(opts, gollm.SetTemperature(config.Temperature))
	}

	if config.MaxRetries > 0 {
		opts = append(opts, gollm.SetMaxRetries(config.MaxRetries))
	}

	if config.Debug {
		opts = append(opts, gollm.SetLogLevel(gollm.LogLevelDebug))
	} else {
		opts = append(opts, gollm.SetLogLevel(gollm.LogLevelOff))
	}

	// Set Ollama endpoint if needed
	if config.BaseURL != "" && config.Provider == "ollama" {
		opts = append(opts, gollm.SetOllamaEndpoint(config.BaseURL))
	}

	llm, err := gollm.NewLLM(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gollm: %w", err)
	}

	return &GollmProvider{
		llm:    llm,
		config: config,
	}, nil
}

// Complete sends a completion request using gollm
func (p *GollmProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	// Build gollm prompt from messages
	var promptText string
	var systemPrompt string

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			if promptText != "" {
				promptText += "\n"
			}
			promptText += msg.Content
		case "assistant":
			// For multi-turn, we'd need memory - for now just skip
		}
	}

	// Create prompt with options
	promptOpts := []gollm.PromptOption{}
	if systemPrompt != "" {
		promptOpts = append(promptOpts, gollm.WithContext(systemPrompt))
	}

	prompt := gollm.NewPrompt(promptText, promptOpts...)

	// Generate response
	response, err := p.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("gollm generate failed: %w", err)
	}

	return &Response{
		Model:   p.config.Model,
		Content: response,
		Choices: []Choice{
			{
				Message: Message{
					Role:    "assistant",
					Content: response,
				},
			},
		},
	}, nil
}

// Stream sends a streaming completion request using gollm
func (p *GollmProvider) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	// Build prompt from messages
	var promptText string
	var systemPrompt string

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			if promptText != "" {
				promptText += "\n"
			}
			promptText += msg.Content
		}
	}

	promptOpts := []gollm.PromptOption{}
	if systemPrompt != "" {
		promptOpts = append(promptOpts, gollm.WithContext(systemPrompt))
	}

	prompt := gollm.NewPrompt(promptText, promptOpts...)

	// Start streaming
	tokenStream, err := p.llm.Stream(ctx, prompt)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("gollm stream failed: %w", err)
	}

	// Convert token stream to our StreamChunk format
	go func() {
		defer close(ch)
		for {
			token, err := tokenStream.Next(ctx)
			if err == io.EOF {
				ch <- StreamChunk{Finished: true}
				return
			}
			if err != nil {
				ch <- StreamChunk{
					Error:    err,
					Finished: true,
				}
				return
			}

			if token == nil {
				ch <- StreamChunk{Finished: true}
				return
			}

			ch <- StreamChunk{
				Delta: Message{
					Role:    "assistant",
					Content: token.Text,
				},
			}
		}
	}()

	return ch, nil
}

// CountTokens estimates token count
func (p *GollmProvider) CountTokens(text string) int {
	// gollm uses tiktoken internally, but for simplicity use approximation
	return len(text) / 4
}

// GetLLM returns the underlying gollm.LLM instance for advanced usage
func (p *GollmProvider) GetLLM() gollm.LLM {
	return p.llm
}
