package agentpool

import (
	"errors"
	"sync"
	"time"
)

// Common errors
var (
	ErrPoolFull      = errors.New("pool has reached maximum agent capacity")
	ErrAgentNotFound = errors.New("agent not found")
	ErrAgentExists   = errors.New("agent with this ID already exists")
	ErrMemoryLimit   = errors.New("memory limit exceeded")
	ErrCPULimit      = errors.New("CPU limit exceeded")
	ErrPoolClosed    = errors.New("pool is closed")
)

// Pool manages a collection of concurrent agents
type Pool struct {
	config Config
	agents map[string]*Agent
	mu     sync.RWMutex
	hooks  LifecycleHooks
	closed bool
	stopCh chan struct{}

	// Resource tracking
	usedMemory int64
	usedCPU    int

	// Health monitoring
	healthTicker *time.Ticker
}

// New creates a new agent pool with the given configuration
func New(config Config) *Pool {
	if config.MaxAgents <= 0 {
		config.MaxAgents = DefaultConfig().MaxAgents
	}

	p := &Pool{
		config: config,
		agents: make(map[string]*Agent),
		stopCh: make(chan struct{}),
	}

	// Start health monitoring if configured
	if config.HealthCheckInterval > 0 {
		p.startHealthMonitoring()
	}

	return p
}

// Spawn creates and adds a new agent to the pool
func (p *Pool) Spawn(id string, config AgentConfig) (*Agent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}

	// Check if agent already exists
	if _, exists := p.agents[id]; exists {
		return nil, ErrAgentExists
	}

	// Check capacity limits
	if len(p.agents) >= p.config.MaxAgents {
		return nil, ErrPoolFull
	}

	// Check memory limit
	if p.config.MemoryLimit > 0 {
		if p.usedMemory+config.Memory > p.config.MemoryLimit {
			return nil, ErrMemoryLimit
		}
	}

	// Check CPU limit
	if p.config.CPULimit > 0 {
		if p.usedCPU+config.CPU > p.config.CPULimit {
			return nil, ErrCPULimit
		}
	}

	// Create the agent
	agent := NewAgent(id, config)

	// Update resource tracking
	p.usedMemory += config.Memory
	p.usedCPU += config.CPU

	// Add to pool
	p.agents[id] = agent

	// Start the agent
	if err := agent.Start(); err != nil {
		delete(p.agents, id)
		p.usedMemory -= config.Memory
		p.usedCPU -= config.CPU
		return nil, err
	}

	// Fire OnStart hook
	if p.hooks.OnStart != nil {
		go p.hooks.OnStart(agent)
	}

	return agent, nil
}

// Kill terminates and removes an agent from the pool
func (p *Pool) Kill(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPoolClosed
	}

	agent, exists := p.agents[id]
	if !exists {
		return ErrAgentNotFound
	}

	// Fire OnKill hook before stopping
	if p.hooks.OnKill != nil {
		p.hooks.OnKill(agent)
	}

	// Stop the agent
	if err := agent.Stop(); err != nil && err.Error() != "agent is already stopped" {
		// Log but continue cleanup
	}

	// Update resource tracking
	p.usedMemory -= agent.Config.Memory
	p.usedCPU -= agent.Config.CPU

	// Remove from pool
	delete(p.agents, id)

	return nil
}

// Pause pauses a running agent
func (p *Pool) Pause(id string) error {
	p.mu.RLock()
	agent, exists := p.agents[id]
	p.mu.RUnlock()

	if !exists {
		return ErrAgentNotFound
	}

	err := agent.Pause()
	if err != nil {
		return err
	}

	// Fire OnPause hook
	if p.hooks.OnPause != nil {
		p.hooks.OnPause(agent)
	}

	return nil
}

// Resume resumes a paused agent
func (p *Pool) Resume(id string) error {
	p.mu.RLock()
	agent, exists := p.agents[id]
	p.mu.RUnlock()

	if !exists {
		return ErrAgentNotFound
	}

	err := agent.Resume()
	if err != nil {
		return err
	}

	// Fire OnResume hook
	if p.hooks.OnResume != nil {
		p.hooks.OnResume(agent)
	}

	return nil
}

// List returns all agents in the pool
func (p *Pool) List() []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := make([]*Agent, 0, len(p.agents))
	for _, agent := range p.agents {
		agents = append(agents, agent)
	}

	return agents
}

// Get returns a specific agent by ID
func (p *Pool) Get(id string) (*Agent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, exists := p.agents[id]
	if !exists {
		return nil, ErrAgentNotFound
	}

	return agent, nil
}

// SetHooks sets the lifecycle hooks for the pool
func (p *Pool) SetHooks(hooks LifecycleHooks) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hooks = hooks
}

// Size returns the number of agents in the pool
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.agents)
}

// ResourceUsage returns current resource usage
func (p *Pool) ResourceUsage() (memory int64, cpu int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.usedMemory, p.usedCPU
}

// Close stops all agents and closes the pool
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// Stop health monitoring
	if p.healthTicker != nil {
		p.healthTicker.Stop()
	}

	// Stop all agents
	for id, agent := range p.agents {
		if p.hooks.OnKill != nil {
			p.hooks.OnKill(agent)
		}
		agent.Stop()
		delete(p.agents, id)
	}

	// Signal stop
	close(p.stopCh)

	return nil
}

// startHealthMonitoring starts the periodic health check routine
func (p *Pool) startHealthMonitoring() {
	p.healthTicker = time.NewTicker(p.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-p.healthTicker.C:
				p.checkAgentHealth()
			case <-p.stopCh:
				return
			}
		}
	}()
}

// checkAgentHealth checks the health of all agents
func (p *Pool) checkAgentHealth() {
	p.mu.RLock()
	agents := make([]*Agent, 0, len(p.agents))
	for _, agent := range p.agents {
		agents = append(agents, agent)
	}
	p.mu.RUnlock()

	for _, agent := range agents {
		// Check if agent context is cancelled
		select {
		case <-agent.Context().Done():
			agent.SetHealthy(false)

			// If agent has an error, fire OnError hook
			if agent.Error != nil && p.hooks.OnError != nil {
				p.hooks.OnError(agent, agent.Error)
			}
		default:
			agent.SetHealthy(true)
		}

		// Check timeout
		if agent.Config.Timeout > 0 && !agent.StartedAt.IsZero() {
			if time.Since(agent.StartedAt) > agent.Config.Timeout {
				agent.SetError(errors.New("agent timeout exceeded"))
				if p.hooks.OnError != nil {
					p.hooks.OnError(agent, agent.Error)
				}
			}
		}
	}
}

// Stats returns pool statistics
type Stats struct {
	TotalAgents int
	Running     int
	Paused      int
	Idle        int
	Stopped     int
	Error       int
	Complete    int
	UsedMemory  int64
	UsedCPU     int
	MaxAgents   int
	MemoryLimit int64
	CPULimit    int
}

// GetStats returns current pool statistics
func (p *Pool) GetStats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := Stats{
		TotalAgents: len(p.agents),
		UsedMemory:  p.usedMemory,
		UsedCPU:     p.usedCPU,
		MaxAgents:   p.config.MaxAgents,
		MemoryLimit: p.config.MemoryLimit,
		CPULimit:    p.config.CPULimit,
	}

	for _, agent := range p.agents {
		switch agent.Status {
		case StatusRunning:
			stats.Running++
		case StatusPaused:
			stats.Paused++
		case StatusIdle:
			stats.Idle++
		case StatusStopped:
			stats.Stopped++
		case StatusError:
			stats.Error++
		case StatusComplete:
			stats.Complete++
		}
	}

	return stats
}
