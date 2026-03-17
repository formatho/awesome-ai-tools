// Package services provides business logic for the Agent Orchestrator.
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
)

// StateService manages agent state persistence and retrieval.
type StateService struct {
	db      *sql.DB
	mu      sync.RWMutex // Protect concurrent access to in-memory cache
	cache   map[string]*models.AgentState
}

// NewStateService creates a new state persistence service.
func NewStateService(db *sql.DB) *StateService {
	return &StateService{
		db:      db,
		cache:   make(map[string]*models.AgentState),
	}
}

// SaveState persists the agent's current state to the database.
func (s *StateService) SaveState(ctx context.Context, agentID string, stateData interface{}, metadata map[string]interface{}) (*models.AgentState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	// Get existing state to determine version
	existingState, err := s.GetState(agentID)
	if err != nil && !models.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to get existing state: %w", err)
	}

	var version int
	if existingState != nil {
		version = existingState.Version + 1
	} else {
		version = 1
	}

	// Serialize data
	stateDataJSON, err := json.Marshal(stateData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize state data: %w", err)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	// Insert or update state using raw SQL with RETURNING (PostgreSQL)
	query := `
		INSERT INTO agent_states (agent_id, state_data, metadata, version, updated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT(agent_id) DO UPDATE SET
			state_data = EXCLUDED.state_data,
			metadata = EXCLUDED.metadata,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
		RETURNING id, agent_id, state_data, metadata, version, updated_at, created_at
	`

	var savedState models.AgentState
	err = s.db.QueryRowContext(ctx, query,
		agentID,
		string(stateDataJSON),
		string(metadataJSON),
		version,
		now,
		now,
	).Scan(&savedState.ID, &savedState.AgentID, 
			&savedState.StateData, &savedState.Metadata,
			&savedState.Version, &savedState.UpdatedAt, &savedState.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	// Update cache
	s.cache[agentID] = &savedState

	return &savedState, nil
}

// GetState retrieves the current state for an agent.
func (s *StateService) GetState(agentID string) (*models.AgentState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache first
	if cachedState, exists := s.cache[agentID]; exists {
		return cachedState, nil
	}

	// Query database
	query := `SELECT id, agent_id, state_data, metadata, version, updated_at, created_at 
			  FROM agent_states WHERE agent_id = $1`

	var savedState models.AgentState
	err := s.db.QueryRow(query, agentID).Scan(
		&savedState.ID, &savedState.AgentID,
		&savedState.StateData, &savedState.Metadata,
		&savedState.Version, &savedState.UpdatedAt, &savedState.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFoundWithID("agent state", agentID)
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	state := &savedState

	// Update cache
	s.cache[agentID] = state

	return state, nil
}

// GetStateHistory retrieves the version history for an agent's state.
func (s *StateService) GetStateHistory(agentID string, limit int, offset int) ([]*models.AgentState, error) {
	query := `SELECT id, agent_id, state_data, metadata, version, updated_at, created_at 
			  FROM agent_states WHERE agent_id = $1 ORDER BY version DESC LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(query, agentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get state history: %w", err)
	}
	defer rows.Close()

	var states []*models.AgentState
	for rows.Next() {
		var savedState models.AgentState
		err := rows.Scan(&savedState.ID, &savedState.AgentID,
			&savedState.StateData, &savedState.Metadata,
			&savedState.Version, &savedState.UpdatedAt, &savedState.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		states = append(states, &savedState)
	}

	return states, rows.Err()
}

// UpdateState partially updates an agent's state.
func (s *StateService) UpdateState(ctx context.Context, agentID string, updates map[string]interface{}) (*models.AgentState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current state
	currentState, err := s.GetState(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %w", err)
	}

	// Apply updates based on what's being updated
	if newStateData, exists := updates["state_data"]; exists {
		currentState.StateData = newStateData
	}

	if newMetadata, exists := updates["metadata"]; exists {
		if metadataMap, ok := newMetadata.(map[string]interface{}); ok {
			currentState.Metadata = metadataMap
		} else if metadataJSON, ok := newMetadata.([]byte); ok {
			var temp map[string]interface{}
			json.Unmarshal(metadataJSON, &temp)
			currentState.Metadata = temp
		}
	}

	if currentState.Version > 0 {
		currentState.Version++
	}

	currentState.UpdatedAt = time.Now().UTC()

	// Serialize and save
	stateDataJSON, _ := json.Marshal(currentState.StateData)
	metadataJSON, _ := json.Marshal(currentState.Metadata)

	query := `UPDATE agent_states SET state_data = $1, metadata = $2, version = $3, updated_at = $4 WHERE agent_id = $5`
	result, err := s.db.ExecContext(ctx, query, string(stateDataJSON), string(metadataJSON), currentState.Version, currentState.UpdatedAt, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, models.ErrNotFoundWithID("agent state", agentID)
	}

	// Update cache
	s.cache[agentID] = currentState

	return currentState, nil
}

// DeleteState removes an agent's persistent state.
func (s *StateService) DeleteState(ctx context.Context, agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM agent_states WHERE agent_id = $1`
	result, err := s.db.ExecContext(ctx, query, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete state: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrNotFoundWithID("agent state", agentID)
	}

	// Remove from cache
	delete(s.cache, agentID)

	return nil
}

// GetAgentStateSummary retrieves a summary of all tracked agents' states.
func (s *StateService) GetAgentStateSummary() ([]*models.AgentState, error) {
	query := `SELECT 
				agent_id,
				COUNT(*) as total_versions,
				MAX(version) as latest_version,
				MAX(updated_at) as last_updated
			  FROM agent_states 
			  GROUP BY agent_id ORDER BY last_updated DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get state summary: %w", err)
	}
	defer rows.Close()

	var summaries []*models.AgentState
	for rows.Next() {
		var savedState models.AgentState
		err := rows.Scan(&savedState.ID, &savedState.AgentID,
			&savedState.StateData, &savedState.Metadata,
			&savedState.Version, &savedState.UpdatedAt, &savedState.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan summary row: %w", err)
		}
		summaries = append(summaries, &savedState)
	}

	return summaries, rows.Err()
}

// ExportState exports an agent's complete state history.
func (s *StateService) ExportState(agentID string) (*models.AgentStateExport, error) {
	currentState, err := s.GetState(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %w", err)
	}

	history, err := s.GetStateHistory(agentID, 1000, 0) // All history up to 1000 versions
	if err != nil {
		return nil, fmt.Errorf("failed to get state history: %w", err)
	}

	export := &models.AgentStateExport{
		AgentID:    agentID,
		Current:    currentState,
		History:    history,
		ExportedAt: time.Now().UTC(),
	}

	return export, nil
}

// ImportState imports state from an export file.
func (s *StateService) ImportState(exportData *models.AgentStateExport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if exportData.Current != nil {
		s.cache[exportData.AgentID] = exportData.Current
	}

	return nil
}

// ClearAgentStateHistory clears all historical versions for an agent (keeps current).
func (s *StateService) ClearAgentStateHistory(agentID string) error {
	query := `DELETE FROM agent_states WHERE agent_id = $1 AND version < (SELECT MAX(version) FROM agent_states WHERE agent_id = $1)`
	result, err := s.db.Exec(query, agentID)
	if err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Cleared %d historical versions for agent %s\n", rowsAffected, agentID)

	return nil
}

// GetStateMetrics returns aggregate metrics about state persistence.
func (s *StateService) GetStateMetrics(ctx context.Context) (*models.StateMetrics, error) {
	query := `SELECT 
				COUNT(DISTINCT agent_id) as total_agents_tracked,
				SUM(version) as total_versions_across_all_agents,
				AVG(CAST(version AS FLOAT)) as avg_versions_per_agent,
				MAX(updated_at) as most_recent_state_update
			  FROM agent_states`

	var metrics models.StateMetrics
	err := s.db.QueryRow(query).Scan(
		&metrics.TotalAgentsTracked,
		&metrics.TotalVersionsAcrossAllAgents,
		&metrics.AvgVersionsPerAgent,
		&metrics.MostRecentStateUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to get state metrics: %w", err)
	}

	return &metrics, nil
}
