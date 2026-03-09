package llm

import (
	"context"

	goagentmodels "github.com/Protocol-Lattice/go-agent/src/models"
)

// GoAgentAdapter wraps a ProviderClient to implement go-agent's models.Agent interface
// This allows using our existing LLM providers with the go-agent framework
type GoAgentAdapter struct {
	provider   ProviderClient
	model      string
	maxTokens  int
}

// NewGoAgentAdapter creates a new adapter for go-agent
func NewGoAgentAdapter(provider ProviderClient, model string, maxTokens int) *GoAgentAdapter {
	return &GoAgentAdapter{
		provider:  provider,
		model:     model,
		maxTokens: maxTokens,
	}
}

// Generate implements models.Agent interface
func (a *GoAgentAdapter) Generate(ctx context.Context, prompt string) (any, error) {
	resp, err := a.provider.Complete(ctx, Request{
		Model: a.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: a.maxTokens,
	})
	if err != nil {
		return nil, err
	}
	return resp.Content, nil
}

// GenerateWithFiles implements models.Agent interface
func (a *GoAgentAdapter) GenerateWithFiles(ctx context.Context, prompt string, files []goagentmodels.File) (any, error) {
	// For now, just append file names to prompt
	// TODO: Implement proper file handling for multimodal models
	enhancedPrompt := prompt
	for _, file := range files {
		enhancedPrompt += "\n\n[Attachment: " + file.Name + " (" + file.MIME + ")]"
	}

	return a.Generate(ctx, enhancedPrompt)
}

// GenerateStream implements models.Agent interface
func (a *GoAgentAdapter) GenerateStream(ctx context.Context, prompt string) (<-chan goagentmodels.StreamChunk, error) {
	// Create output channel
	outCh := make(chan goagentmodels.StreamChunk, 100)

	// Start streaming from provider
	streamCh, err := a.provider.Stream(ctx, Request{
		Model: a.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: a.maxTokens,
	})
	if err != nil {
		close(outCh)
		return nil, err
	}

	// Convert our StreamChunk to go-agent's StreamChunk
	go func() {
		defer close(outCh)
		var fullText string

		for chunk := range streamCh {
			if chunk.Error != nil {
				outCh <- goagentmodels.StreamChunk{
					Err: chunk.Error,
				}
				return
			}

			fullText += chunk.Delta.Content

			outCh <- goagentmodels.StreamChunk{
				Delta:    chunk.Delta.Content,
				Done:     chunk.Finished,
				FullText: fullText,
			}

			if chunk.Finished {
				return
			}
		}

		// Send final chunk if stream ended without Finished
		outCh <- goagentmodels.StreamChunk{
			Done:     true,
			FullText: fullText,
		}
	}()

	return outCh, nil
}

// Ensure GoAgentAdapter implements models.Agent interface
var _ goagentmodels.Agent = (*GoAgentAdapter)(nil)
