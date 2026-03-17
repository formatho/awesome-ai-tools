// Package services provides business logic layer for the API.
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	agentpool "github.com/formatho/agent-orchestrator/packages/agent-pool"
	agentrunner "github.com/formatho/agent-orchestrator/packages/agent-runner"
	llmclient "github.com/formatho/agent-orchestrator/packages/llm-client"
	"github.com/google/uuid"
)

// ErrNoDatabase is returned when database is not available (e.g., in tests)
var ErrNoDatabase = errors.New("database not available")

// AgentService handles agent operations.
type AgentService struct {
	db      *sql.DB
	hub     *websocket.Hub
	pool    *agentpool.Pool
	runner  *agentrunner.AgentRunner
}

// NewAgentService creates a new agent service.
func NewAgentService(db *sql.DB, hub *websocket.Hub) *AgentService {
	service := &AgentService{
		db:     db,
		hub:    hub,
		pool:   agentpool.New(agentpool.Config{MaxAgents: 100}),
		runner: agentrunner.NewAgentRunner(),
	}

	// Reset any agents that are marked as "running" but aren't actually running
	// This can happen when the application restarts
	service.resetStaleRunningAgents()

	return service
}

// List returns all agents. Optionally filtered by organization_id.
func (s *AgentService) List(orgID *string) ([]*models.Agent, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	var query string
	var args []interface{}

	if orgID != nil && *orgID != "" {
		query = `SELECT id, name, status, provider, model, system_prompt, base_url, work_dir, organization_id, config, metadata,
			created_at, updated_at, started_at, stopped_at, error
			FROM agents WHERE organization_id = ? ORDER BY created_at DESC`
		args = append(args, *orgID)
	} else {
		query = `SELECT id, name, status, provider, model, system_prompt, base_url, work_dir, organization_id, config, metadata,
			created_at, updated_at, started_at, stopped_at, error
			FROM agents ORDER BY created_at DESC`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		a := &models.Agent{}
		var config, metadata sql.NullString
		var startedAt, stoppedAt sql.NullTime
		var provider, model, systemPrompt, baseURL, workDir, orgID, agentError sql.NullString

		err := rows.Scan(
			&a.ID, &a.Name, &a.Status, &provider, &model, &systemPrompt, &baseURL, &workDir,
			&orgID, &config, &metadata, &a.CreatedAt, &a.UpdatedAt,
			&startedAt, &stoppedAt, &agentError,
		)
		if err != nil {
			return nil, err
		}

		a.Provider = provider.String
		a.Model = model.String
		a.SystemPrompt = systemPrompt.String
		a.BaseURL = baseURL.String
		a.WorkDir = workDir.String
		a.OrganizationID = orgID.String
		a.Error = agentError.String
		if startedAt.Valid {
			a.StartedAt = &startedAt.Time
		}
		if stoppedAt.Valid {
			a.StoppedAt = &stoppedAt.Time
		}

		if config.Valid && config.String != "" {
			_ = json.Unmarshal([]byte(config.String), &a.Config)
		}
		if metadata.Valid && metadata.String != "" {
			_ = json.Unmarshal([]byte(metadata.String), &a.Metadata)
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

	query := `SELECT id, name, status, provider, model, system_prompt, base_url, work_dir, organization_id, config, metadata,
		created_at, updated_at, started_at, stopped_at, error
		FROM agents WHERE id = ?`

	a := &models.Agent{}
	var config, metadata sql.NullString
	var startedAt, stoppedAt sql.NullTime
	var provider, model, systemPrompt, baseURL, workDir, orgID, agentError sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&a.ID, &a.Name, &a.Status, &provider, &model, &systemPrompt, &baseURL, &workDir,
		&orgID, &config, &metadata, &a.CreatedAt, &a.UpdatedAt,
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
	a.BaseURL = baseURL.String
	a.WorkDir = workDir.String
	a.OrganizationID = orgID.String
	a.Error = agentError.String
	if startedAt.Valid {
		a.StartedAt = &startedAt.Time
	}
	if stoppedAt.Valid {
		a.StoppedAt = &stoppedAt.Time
	}

	if config.Valid && config.String != "" {
		_ = json.Unmarshal([]byte(config.String), &a.Config)
	}
	if metadata.Valid && metadata.String != "" {
		_ = json.Unmarshal([]byte(metadata.String), &a.Metadata)
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

	query := `INSERT INTO agents (id, name, status, provider, model, system_prompt, base_url, work_dir, organization_id, config, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, req.Name, status, req.Provider, req.Model,
		req.SystemPrompt, "", req.WorkDir, req.OrganizationID, string(configJSON), string(metadataJSON), now, now)
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
	if req.OrganizationID != nil {
		query += `, organization_id = ?`
		args = append(args, *req.OrganizationID)
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

	// Get the agent from database
	agent, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Create the agent in the runner if it doesn't exist
	ctx := context.Background()

	// Build agent config
	config := agentrunner.AgentConfig{
		Provider:     agent.Provider,
		Model:        agent.Model,
		APIKey:       s.getAPIKey(agent.Provider),
		BaseURL:      agent.BaseURL,
		MaxTokens:    4096, // Default max tokens
		SystemPrompt: agent.SystemPrompt,
		MemoryLimit:  8,    // Default memory limit
	}

	// Create agent in runner
	_, err = s.runner.CreateAgent(ctx, id, config)
	if err != nil {
		// If agent already exists, that's okay - just continue
		if !errors.Is(err, agentrunner.ErrAgentAlreadyExists) {
			return nil, fmt.Errorf("failed to create agent in runner: %w", err)
		}
	}

	// Start the agent with an initial prompt
	initialPrompt := agent.SystemPrompt
	if initialPrompt == "" {
		initialPrompt = "You are now active and ready to help."
	}

	_, err = s.runner.Start(ctx, id, initialPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to start agent in runner: %w", err)
	}

	// Update database status
	query := `UPDATE agents SET status = ?, updated_at = ?, started_at = ? WHERE id = ?`
	_, err = s.db.Exec(query, models.AgentStatusRunning, now, now, id)
	if err != nil {
		return nil, err
	}

	// Get updated agent
	agent, err = s.Get(id)
	if err != nil {
		return nil, err
	}

	// Start a goroutine to monitor and sync status
	go s.monitorAgentStatus(id)

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

// CreateLLMClient creates an LLM client for the agent using gollm library.
func (s *AgentService) CreateLLMClient(agent *models.Agent, apiKey string) (llmclient.ProviderClient, error) {
	provider := agent.Provider
	if provider == "" {
		provider = "openai"
	}

	// Use gollm for all providers
	config := llmclient.GollmConfig{
		Provider: provider,
		Model:    agent.Model,
		APIKey:   apiKey,
		BaseURL:  agent.BaseURL,
	}

	// For zai provider, use openrouter with custom base URL if needed
	if provider == "zai" && agent.BaseURL != "" {
		config.Provider = "openrouter"
		config.BaseURL = agent.BaseURL
	}

	client, err := llmclient.NewGollmProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create gollm provider: %w", err)
	}

	return client, nil
}

// GetLogs returns logs for an agent, ordered by most recent first.
func (s *AgentService) GetLogs(agentID string, limit int) ([]*models.AgentLog, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Validate limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `SELECT id, agent_id, level, message, metadata, created_at
		FROM agent_logs WHERE agent_id = ?
		ORDER BY created_at DESC LIMIT ?`

	rows, err := s.db.Query(query, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AgentLog
	for rows.Next() {
		log := &models.AgentLog{}
		var metadata sql.NullString

		err := rows.Scan(&log.ID, &log.AgentID, &log.Level, &log.Message, &metadata, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		if metadata.Valid && metadata.String != "" {
			_ = json.Unmarshal([]byte(metadata.String), &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// AddLog adds a log entry for an agent.
func (s *AgentService) AddLog(agentID string, level models.LogLevel, message string, metadata map[string]interface{}) error {
	if s.db == nil {
		return ErrNoDatabase
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	metadataJSON, _ := json.Marshal(metadata)

	query := `INSERT INTO agent_logs (id, agent_id, level, message, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, agentID, level, message, string(metadataJSON), now)
	if err != nil {
		return err
	}

	return nil
}

// getAPIKey retrieves the API key for a given provider from environment variables.
func (s *AgentService) getAPIKey(provider string) string {
	switch provider {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "zai":
		return os.Getenv("ZAI_API_KEY")
	case "groq":
		return os.Getenv("GROQ_API_KEY")
	case "mistral":
		return os.Getenv("MISTRAL_API_KEY")
	case "openrouter":
		return os.Getenv("OPENROUTER_API_KEY")
	default:
		return ""
	}
}

// monitorAgentStatus monitors an agent's status in the runner and syncs it to the database.
func (s *AgentService) monitorAgentStatus(agentID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get status from runner
			status, err := s.runner.GetStatus(agentID)
			if err != nil {
				// Agent not in runner, mark as stopped in database
				s.updateStatus(agentID, models.AgentStatusIdle)
				return
			}

			// Sync status to database based on runner status
			var newStatus models.AgentStatus
			switch status.Status {
			case agentrunner.StatusRunning:
				newStatus = models.AgentStatusRunning
			case agentrunner.StatusPaused:
				newStatus = models.AgentStatusPaused
			case agentrunner.StatusStopped, agentrunner.StatusComplete:
				newStatus = models.AgentStatusIdle
			case agentrunner.StatusError:
				newStatus = models.AgentStatusError
			default:
				newStatus = models.AgentStatusIdle
			}

			// Update database if status changed
			agent, err := s.Get(agentID)
			if err == nil && agent.Status != newStatus {
				s.updateStatus(agentID, newStatus)
			}

			// If agent is stopped or complete, stop monitoring
			if status.Status == agentrunner.StatusStopped || status.Status == agentrunner.StatusComplete {
				return
			}
		}
	}
}
