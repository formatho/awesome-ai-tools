// Package goagent provides integration with the Protocol-Lattice/go-agent framework
package goagent

import (
	"context"
	"fmt"

	goagent "github.com/Protocol-Lattice/go-agent"
	"github.com/Protocol-Lattice/go-agent/src/memory"
	"github.com/Protocol-Lattice/go-agent/src/memory/engine"
	llmclient "github.com/formatho/agent-orchestrator/packages/llm-client"
)

// AgentConfig holds configuration for creating a go-agent
type AgentConfig struct {
	// LLM Provider configuration
	Provider   string
	Model      string
	APIKey     string
	BaseURL    string
	MaxTokens  int

	// Agent configuration
	SystemPrompt string
	ContextLimit int

	// Memory configuration
	EnableMemory bool
	MemoryLimit  int
}

// AgentService provides go-agent functionality
type AgentService struct {
	agents map[string]*goagent.Agent
}

// NewAgentService creates a new agent service
func NewAgentService() *AgentService {
	return &AgentService{
		agents: make(map[string]*goagent.Agent),
	}
}

// CreateAgent creates a new go-agent with the given configuration
func (s *AgentService) CreateAgent(ctx context.Context, id string, config AgentConfig) (*goagent.Agent, error) {
	// Create LLM provider
	var provider llmclient.ProviderClient
	var err error

	switch config.Provider {
	case "openai":
		provider = llmclient.NewOpenAIProvider(llmclient.OpenAIConfig{
			APIKey: config.APIKey,
		})
	case "anthropic":
		provider = llmclient.NewAnthropicProvider(llmclient.AnthropicConfig{
			APIKey: config.APIKey,
			Model:  config.Model,
		})
	case "ollama":
		provider = llmclient.NewOllamaProvider(llmclient.OllamaConfig{
			BaseURL: config.BaseURL,
		})
	case "zai":
		provider = llmclient.NewZAIProvider(llmclient.ZAIConfig{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	// Create go-agent adapter
	modelAdapter := llmclient.NewGoAgentAdapter(provider, config.Model, config.MaxTokens)

	// Create session memory
	memOpts := engine.DefaultOptions()
	sessionMemory := memory.NewSessionMemory(
		memory.NewMemoryBankWithStore(memory.NewInMemoryStore()),
		config.ContextLimit,
	).WithEngine(memory.NewEngine(memory.NewInMemoryStore(), memOpts))

	// Create agent options
	opts := goagent.Options{
		Model:        modelAdapter,
		Memory:       sessionMemory,
		SystemPrompt: config.SystemPrompt,
		ContextLimit: config.ContextLimit,
	}

	if opts.ContextLimit <= 0 {
		opts.ContextLimit = 8
	}

	if opts.SystemPrompt == "" {
		opts.SystemPrompt = "You are a helpful AI assistant."
	}

	// Create agent
	agent, err := goagent.New(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Store agent
	s.agents[id] = agent

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (s *AgentService) GetAgent(id string) (*goagent.Agent, bool) {
	agent, ok := s.agents[id]
	return agent, ok
}

// DeleteAgent removes an agent
func (s *AgentService) DeleteAgent(id string) {
	delete(s.agents, id)
}

// Generate generates a response from an agent
func (s *AgentService) Generate(ctx context.Context, agentID, sessionID, prompt string) (string, error) {
	agent, ok := s.GetAgent(agentID)
	if !ok {
		return "", fmt.Errorf("agent not found: %s", agentID)
	}

	result, err := agent.Generate(ctx, sessionID, prompt)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	// Result is any, need type assertion
	switch v := result.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ListAgents returns all agent IDs
func (s *AgentService) ListAgents() []string {
	ids := make([]string, 0, len(s.agents))
	for id := range s.agents {
		ids = append(ids, id)
	}
	return ids
}
