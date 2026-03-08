// Package services provides business logic layer for the API.
package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	agentpool "github.com/formatho/agent-orchestrator/packages/agent-pool"
	llmclient "github.com/formatho/agent-orchestrator/packages/llm-client"
	"github.com/google/uuid"
)

// ErrNoDatabase is returned when database is not available (e.g., in tests)
var ErrNoDatabase = errors.New("database not available")

// AgentService handles agent operations.
type AgentService struct {
	db   *sql.DB
	hub  *websocket.Hub
	pool *agentpool.Pool
	mu   sync.RWMutex
}

// NewAgentService creates a new agent service.
func NewAgentService(db *sql.DB, hub *websocket.Hub) *AgentService {
	service := &AgentService{
		db:   db,
		hub:  hub,
		pool: agentpool.New(agentpool.Config{MaxAgents: 100}),
	}

	// Reset any agents that are marked as "running" but aren't actually running
	// This can happen when the application restarts
	service.resetStaleRunningAgents()

	return service
}

// List returns all agents.
func (s *AgentService) List() ([]*models.Agent, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, name, status, provider, model, system_prompt, work_dir, config, metadata,
		created_at, updated_at, started_at, stopped_at, error
		FROM agents ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		a := &models.Agent{}
		var config, metadata sql.NullString
		var startedAt, stoppedAt sql.NullTime
		var provider, model, systemPrompt, workDir, agentError sql.NullString

		err := rows.Scan(
			&a.ID, &a.Name, &a.Status, &provider, &model, &systemPrompt, &workDir,
			&config, &metadata, &a.CreatedAt, &a.UpdatedAt,
			&startedAt, &stoppedAt, &agentError,
		)
		if err != nil {
			return nil, err
		}

		a.Provider = provider.String
		a.Model = model.String
		a.SystemPrompt = systemPrompt.String
		a.WorkDir = workDir.String
		a.Error = agentError.String
		a.StartedAt = &startedAt.Time
		if !startedAt.Valid {
			a.StartedAt = nil
		}
		a.StoppedAt = &stoppedAt.Time
		if !stoppedAt.Valid {
			a.StoppedAt = nil
		}

		if config.Valid && config.String != "" {
			json.Unmarshal([]byte(config.String), &a.Config)
		}
		if metadata.Valid && metadata.String != "" {
			json.Unmarshal([]byte(metadata.String), &a.Metadata)
		}

		agents = append(agents, a)
	}

	return agents, nil
}

// Get returns a single agent by ID.
func (s *AgentService) Get(id string) (*models.Agent, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, name, status, provider, model, system_prompt, work_dir, config, metadata,
		created_at, updated_at, started_at, stopped_at, error
		FROM agents WHERE id = ?`

	a := &models.Agent{}
	var config, metadata sql.NullString
	var startedAt, stoppedAt sql.NullTime
	var provider, model, systemPrompt, workDir, agentError sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&a.ID, &a.Name, &a.Status, &provider, &model, &systemPrompt, &workDir,
		&config, &metadata, &a.CreatedAt, &a.UpdatedAt,
		&startedAt, &stoppedAt, &agentError,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	a.Provider = provider.String
	a.Model = model.String
	a.SystemPrompt = systemPrompt.String
	a.WorkDir = workDir.String
	a.Error = agentError.String
	if startedAt.Valid {
		a.StartedAt = &startedAt.Time
	}
	if stoppedAt.Valid {
		a.StoppedAt = &stoppedAt.Time
	}

	if config.Valid && config.String != "" {
		json.Unmarshal([]byte(config.String), &a.Config)
	}
	if metadata.Valid && metadata.String != "" {
		json.Unmarshal([]byte(metadata.String), &a.Metadata)
	}

	return a, nil
}

// Create creates a new agent.
func (s *AgentService) Create(req *models.AgentCreate) (*models.Agent, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if database is available
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	status := models.AgentStatusIdle

	configJSON, _ := json.Marshal(req.Config)
	metadataJSON, _ := json.Marshal(req.Metadata)

	query := `INSERT INTO agents (id, name, status, provider, model, system_prompt, work_dir, config, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, req.Name, status, req.Provider, req.Model,
		req.SystemPrompt, req.WorkDir, string(configJSON), string(metadataJSON), now, now)
	if err != nil {
		return nil, err
	}

	agent, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Broadcast creation event
	s.hub.BroadcastAgentCreated(agent)

	return agent, nil
}

// Update updates an existing agent.
func (s *AgentService) Update(id string, req *models.AgentUpdate) (*models.Agent, error) {
	// Check if agent exists
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE agents SET updated_at = ?`
	args := []interface{}{now}

	if req.Name != nil {
		query += `, name = ?`
		args = append(args, *req.Name)
	}
	if req.Provider != nil {
		query += `, provider = ?`
		args = append(args, *req.Provider)
	}
	if req.Model != nil {
		query += `, model = ?`
		args = append(args, *req.Model)
	}
	if req.SystemPrompt != nil {
		query += `, system_prompt = ?`
		args = append(args, *req.SystemPrompt)
	}
	if req.WorkDir != nil {
		query += `, work_dir = ?`
		args = append(args, *req.WorkDir)
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		query += `, config = ?`
		args = append(args, string(configJSON))
	}
	if req.Metadata != nil {
		metadataJSON, _ := json.Marshal(req.Metadata)
		query += `, metadata = ?`
		args = append(args, string(metadataJSON))
	}

	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Delete deletes an agent.
func (s *AgentService) Delete(id string) error {
	if _, err := s.Get(id); err != nil {
		return err
	}

	_, err := s.db.Exec(`DELETE FROM agents WHERE id = ?`, id)
	if err != nil {
		return err
	}

	s.hub.Broadcast(websocket.EventAgentDeleted, map[string]string{"id": id})
	return nil
}

// Pause pauses an agent.
func (s *AgentService) Pause(id string) (*models.Agent, error) {
	return s.updateStatus(id, models.AgentStatusPaused)
}

// Resume resumes a paused agent.
func (s *AgentService) Resume(id string) (*models.Agent, error) {
	return s.updateStatus(id, models.AgentStatusRunning)
}

// Start starts an agent (alias for Resume).
func (s *AgentService) Start(id string) (*models.Agent, error) {
	now := time.Now().UTC()
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	query := `UPDATE agents SET status = ?, updated_at = ?, started_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, models.AgentStatusRunning, now, now, id)
	if err != nil {
		return nil, err
	}

	agent, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastAgentStatus(id, models.AgentStatusRunning)
	return agent, nil
}

// Stop stops an agent (alias for Pause).
func (s *AgentService) Stop(id string) (*models.Agent, error) {
	now := time.Now().UTC()
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	query := `UPDATE agents SET status = ?, updated_at = ?, stopped_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, models.AgentStatusIdle, now, now, id)
	if err != nil {
		return nil, err
	}

	agent, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastAgentStatus(id, models.AgentStatusIdle)
	return agent, nil
}

func (s *AgentService) updateStatus(id string, status models.AgentStatus) (*models.Agent, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE agents SET status = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, now, id)
	if err != nil {
		return nil, err
	}

	agent, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastAgentStatus(id, status)
	return agent, nil
}

// resetStaleRunningAgents resets any agents that are marked as "running" or "paused"
// to "idle" status. This is called on service initialization to handle the case
// where the application was restarted while agents were marked as running.
func (s *AgentService) resetStaleRunningAgents() {
	// Skip if no database connection (e.g., in tests)
	if s.db == nil {
		return
	}

	now := time.Now().UTC()

	// Reset all running and paused agents to idle, with an error message
	query := `UPDATE agents
		SET status = ?, updated_at = ?, stopped_at = ?, error = ?
		WHERE status IN (?, ?)`

	_, err := s.db.Exec(query, models.AgentStatusIdle, now, now,
		"Agent was not properly stopped (application restart)",
		models.AgentStatusRunning, models.AgentStatusPaused)

	if err != nil {
		// Log the error but don't fail to start the service
		fmt.Printf("Warning: failed to reset stale running agents: %v\n", err)
	}
}

// CreateLLMClient creates an LLM client for the agent.
func (s *AgentService) CreateLLMClient(agent *models.Agent, apiKey string) (llmclient.ProviderClient, error) {
	provider := agent.Provider
	if provider == "" {
		provider = "openai"
	}

	switch provider {
	case "openai":
		return llmclient.NewOpenAIProvider(llmclient.OpenAIConfig{
			APIKey: apiKey,
		}), nil
	case "anthropic":
		return llmclient.NewAnthropicProvider(llmclient.AnthropicConfig{
			APIKey: apiKey,
			Model:  agent.Model,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}
