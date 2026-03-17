// Package agentrunner provides an execution engine for AI agents using the go-agent framework.
package agentrunner

import (
	"context"
	"fmt"
	"sync"
	"time"

	goagent "github.com/Protocol-Lattice/go-agent"
	goagentsvc "github.com/formatho/agent-orchestrator/packages/goagent"
	agentpkg "github.com/formatho/agent-orchestrator/packages/agent-skills"
)

// AgentRunner handles the execution of AI agents.
type AgentRunner struct {
	mu       sync.RWMutex
	active   map[string]*ActiveAgent
	skills   *agentpkg.Runner
	goAgent  *goagentService
}

// ActiveAgent represents a running agent with its state.
type ActiveAgent struct {
	ID         string
	Agent      *goagent.Agent
	SessionID  string
	Context    context.Context
	Cancel     context.CancelFunc
	Status     AgentStatus
	Error      error
	Result     any
	Tasks      []Task
	TaskIndex  int
	LastUpdate time.Time
}

// AgentStatus represents the execution status of an agent.
type AgentStatus string

const (
	StatusIdle    AgentStatus = "idle"
	StatusRunning AgentStatus = "running"
	StatusPaused  AgentStatus = "paused"
	StatusStopped AgentStatus = "stopped"
	StatusError   AgentStatus = "error"
	StatusComplete AgentStatus = "complete"
)

// Task represents a task for the agent to execute.
type Task struct {
	ID        string
	Prompt    string
	Result    any
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// goagentService wraps the go-agent service for agent creation.
type goagentService struct {
	mu       sync.RWMutex
	services map[string]*goagentsvc.AgentService
}

// getOrCreateService gets or creates a go-agent service for the given agent ID.
func (g *goagentService) getOrCreateService(agentID string) *goagentsvc.AgentService {
	g.mu.RLock()
	svc, exists := g.services[agentID]
	g.mu.RUnlock()

	if exists {
		return svc
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Double-check after acquiring write lock
	if svc, exists := g.services[agentID]; exists {
		return svc
	}

	// Create new service
	svc = goagentsvc.NewAgentService()
	g.services[agentID] = svc

	return svc
}

// NewAgentRunner creates a new agent runner with default configuration.
func NewAgentRunner() *AgentRunner {
	return &AgentRunner{
		active:   make(map[string]*ActiveAgent),
		skills:   agentpkg.NewRunner(agentpkg.Config{}),
		goAgent:  &goagentService{services: make(map[string]*goagentsvc.AgentService)},
	}
}

// NewAgentRunnerWithConfig creates a new agent runner with custom configuration.
func NewAgentRunnerWithConfig(cfg Config) *AgentRunner {
	return &AgentRunner{
		active:   make(map[string]*ActiveAgent),
		skills:   agentpkg.NewRunner(agentpkg.Config{}),
		goAgent:  &goagentService{services: make(map[string]*goagentsvc.AgentService)},
	}
}

// CreateAgent creates a new AI agent with the given configuration.
func (r *AgentRunner) CreateAgent(ctx context.Context, agentID string, config AgentConfig) (*ActiveAgent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if agent already exists
	if _, exists := r.active[agentID]; exists {
		return nil, ErrAgentAlreadyExists
	}

	agent := &ActiveAgent{
		ID:         agentID,
		SessionID:  fmt.Sprintf("session-%s", time.Now().Format("20060102-150405")),
		Status:     StatusIdle,
		LastUpdate: time.Now(),
	}

	// Create go-agent service for this agent ID
	goAgentSvc := r.goAgent.getOrCreateService(agentID)

	// Create the actual agent using go-agent framework
	gAgent, err := goAgentSvc.CreateAgent(ctx, fmt.Sprintf("agent-%s", agentID), goagentsvc.AgentConfig{
		Provider:     config.Provider,
		Model:        config.Model,
		APIKey:       config.APIKey,
		BaseURL:      config.BaseURL,
		MaxTokens:    config.MaxTokens,
		SystemPrompt: config.SystemPrompt,
		ContextLimit: config.MemoryLimit,
		EnableMemory: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	agent.Agent = gAgent

	r.active[agentID] = agent
	return agent, nil
}

// Start begins executing an agent with the given prompt.
func (r *AgentRunner) Start(ctx context.Context, agentID string, initialPrompt string) (*ActiveAgent, error) {
	r.mu.Lock()
	agent, exists := r.active[agentID]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Agent == nil {
		return nil, fmt.Errorf("agent not initialized")
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithCancel(ctx)
	agent.Context = execCtx
	agent.Cancel = cancel

	// Run the agent in a goroutine
	go r.executeAgent(agent, initialPrompt)

	return agent, nil
}

// executeAgent runs the agent's main loop.
func (r *AgentRunner) executeAgent(agent *ActiveAgent, initialPrompt string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			r.mu.Lock()
			agent.Status = StatusError
			agent.Error = fmt.Errorf("panic: %v", recovered)
			agent.Cancel()
			r.mu.Unlock()
		}
	}()

	// Start execution
	r.mu.Lock()
	agent.Status = StatusRunning
	agent.LastUpdate = time.Now()
	r.mu.Unlock()

	// Send initial prompt to agent
	result, err := agent.Agent.Generate(agent.Context, agent.SessionID, initialPrompt)
	if err != nil {
		r.mu.Lock()
		agent.Error = err
		agent.Status = StatusError
		agent.Cancel()
		r.mu.Unlock()
		return
	}

	// Process result
	r.mu.Lock()
	agent.Result = result
	agent.LastUpdate = time.Now()
	agent.Status = StatusComplete
	r.mu.Unlock()
}

// SendPrompt sends a prompt to an active agent.
func (r *AgentRunner) SendPrompt(ctx context.Context, agentID string, prompt string) error {
	r.mu.RLock()
	agent, exists := r.active[agentID]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	result, err := agent.Agent.Generate(ctx, agent.SessionID, prompt)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	r.mu.Lock()
	agent.Result = result
	agent.LastUpdate = time.Now()
	r.mu.Unlock()

	return nil
}

// Stop stops an active agent.
func (r *AgentRunner) Stop(agentID string) error {
	r.mu.Lock()
	agent, exists := r.active[agentID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Cancel != nil {
		agent.Cancel()
	}

	r.mu.Lock()
	agent.Status = StatusStopped
	agent.LastUpdate = time.Now()
	r.mu.Unlock()

	return nil
}

// Pause pauses an active agent.
func (r *AgentRunner) Pause(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.active[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	agent.Status = StatusPaused
	agent.LastUpdate = time.Now()
	return nil
}

// Resume resumes a paused agent.
func (r *AgentRunner) Resume(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.active[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if agent.Status != StatusPaused {
		return nil // Already running or idle
	}

	// Create new execution context
	execCtx, cancel := context.WithCancel(agent.Context)
	agent.Context = execCtx
	agent.Cancel = cancel
	agent.Status = StatusRunning
	agent.LastUpdate = time.Now()

	go r.executeAgent(agent, "") // Continue from current state
	return nil
}

// GetStatus returns the status of an agent.
func (r *AgentRunner) GetStatus(agentID string) (*ActiveAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.active[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	// Return a copy to avoid race conditions
	return &ActiveAgent{
		ID:         agent.ID,
		SessionID:  agent.SessionID,
		Status:     agent.Status,
		Error:      agent.Error,
		Result:     agent.Result,
		LastUpdate: agent.LastUpdate,
	}, nil
}

// List returns all active agents.
func (r *AgentRunner) List() []*ActiveAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ActiveAgent, 0, len(r.active))
	for _, agent := range r.active {
		result = append(result, &ActiveAgent{
			ID:         agent.ID,
			SessionID:  agent.SessionID,
			Status:     agent.Status,
			Error:      agent.Error,
			Result:     agent.Result,
			LastUpdate: agent.LastUpdate,
		})
	}

	return result
}

// Delete removes an agent from the runner.
func (r *AgentRunner) Delete(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.active[agentID]; !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	delete(r.active, agentID)
	return nil
}
