package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	todoqueue "github.com/formatho/agent-orchestrator/packages/todo-queue"
	"github.com/google/uuid"
)

// TODOService handles TODO operations.
type TODOService struct {
	db    *sql.DB
	hub   *websocket.Hub
	queue *todoqueue.Queue
}

// NewTODOService creates a new TODO service.
func NewTODOService(db *sql.DB, hub *websocket.Hub) *TODOService {
	// Initialize todo-queue with in-memory database for now
	// In production, this would use a persistent database
	queue, _ := todoqueue.New(todoqueue.Config{
		DBPath:     ":memory:",
		MaxRetries: 3,
	})
	return &TODOService{
		db:    db,
		hub:   hub,
		queue: queue,
	}
}

// List returns all TODOs. Optionally filtered by organization_id.
func (s *TODOService) List(orgID *string) ([]*models.TODO, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	var query string
	var args []interface{}

	if orgID != nil && *orgID != "" {
		query = `SELECT id, title, description, status, priority, progress, agent_id, organization_id,
			skills, dependencies, config, result, error,
			created_at, updated_at, started_at, completed_at
			FROM todos WHERE organization_id = ? ORDER BY priority DESC, created_at DESC`
		args = append(args, *orgID)
	} else {
		query = `SELECT id, title, description, status, priority, progress, agent_id, organization_id,
			skills, dependencies, config, result, error,
			created_at, updated_at, started_at, completed_at
			FROM todos ORDER BY priority DESC, created_at DESC`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", cerr)
		}
	}()

	var todos []*models.TODO
	for rows.Next() {
		t := &models.TODO{}
		var description, agentID, orgID, skills, deps, config, result, todoError sql.NullString
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&t.ID, &t.Title, &description, &t.Status, &t.Priority, &t.Progress, &agentID,
			&orgID, &skills, &deps, &config, &result, &todoError,
			&t.CreatedAt, &t.UpdatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, err
		}

		t.Description = description.String
		if agentID.Valid {
			t.AgentID = &agentID.String
		}
		t.OrganizationID = orgID.String
		t.Error = todoError.String

		t.StartedAt = &startedAt.Time
		if !startedAt.Valid {
			t.StartedAt = nil
		}
		t.CompletedAt = &completedAt.Time
		if !completedAt.Valid {
			t.CompletedAt = nil
		}

		if skills.Valid && skills.String != "" {
			_ = json.Unmarshal([]byte(skills.String), &t.Skills)
		}
		if deps.Valid && deps.String != "" {
			_ = json.Unmarshal([]byte(deps.String), &t.Dependencies)
		}
		if config.Valid && config.String != "" {
			_ = json.Unmarshal([]byte(config.String), &t.Config)
		}
		if result.Valid && result.String != "" {
			_ = json.Unmarshal([]byte(result.String), &t.Result)
		}

		todos = append(todos, t)
	}

	return todos, nil
}

// Get returns a single TODO by ID.
func (s *TODOService) Get(id string) (*models.TODO, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, title, description, status, priority, progress, agent_id, organization_id,
		skills, dependencies, config, result, error,
		created_at, updated_at, started_at, completed_at
		FROM todos WHERE id = ?`

	t := &models.TODO{}
	var description, agentID, orgID, skills, deps, config, result, todoError sql.NullString
	var startedAt, completedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&t.ID, &t.Title, &description, &t.Status, &t.Priority, &t.Progress, &agentID,
		&orgID, &skills, &deps, &config, &result, &todoError,
		&t.CreatedAt, &t.UpdatedAt, &startedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFoundWithID("TODO", id)
	}
	if err != nil {
		return nil, err
	}

	t.Description = description.String
	if agentID.Valid {
		t.AgentID = &agentID.String
	}
	t.OrganizationID = orgID.String
	t.Error = todoError.String

	t.StartedAt = &startedAt.Time
	if !startedAt.Valid {
		t.StartedAt = nil
	}
	t.CompletedAt = &completedAt.Time
	if !completedAt.Valid {
		t.CompletedAt = nil
	}

	if skills.Valid && skills.String != "" {
		_ = json.Unmarshal([]byte(skills.String), &t.Skills)
	}
	if deps.Valid && deps.String != "" {
		_ = json.Unmarshal([]byte(deps.String), &t.Dependencies)
	}
	if config.Valid && config.String != "" {
		_ = json.Unmarshal([]byte(config.String), &t.Config)
	}
	if result.Valid && result.String != "" {
		_ = json.Unmarshal([]byte(result.String), &t.Result)
	}

	return t, nil
}

// Create creates a new TODO.
func (s *TODOService) Create(req *models.TODOCreate) (*models.TODO, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if database is available
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	status := models.TODOStatusPending

	skillsJSON, _ := json.Marshal(req.Skills)
	depsJSON, _ := json.Marshal(req.Dependencies)
	configJSON, _ := json.Marshal(req.Config)

	query := `INSERT INTO todos (id, title, description, status, priority, agent_id, organization_id, skills, dependencies, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var agentID interface{}
	if req.AgentID != nil {
		agentID = *req.AgentID
	}

	_, err := s.db.Exec(query, id, req.Title, req.Description, status, req.Priority,
		agentID, req.OrganizationID, string(skillsJSON), string(depsJSON), string(configJSON), now, now)
	if err != nil {
		return nil, err
	}

	todo, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Broadcast creation
	s.hub.BroadcastTODOStatus(id, string(status))

	return todo, nil
}

// Update updates an existing TODO.
func (s *TODOService) Update(id string, req *models.TODOUpdate) (*models.TODO, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE todos SET updated_at = ?`
	args := []interface{}{now}

	if req.Title != nil {
		query += `, title = ?`
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		query += `, description = ?`
		args = append(args, *req.Description)
	}
	if req.Priority != nil {
		query += `, priority = ?`
		args = append(args, *req.Priority)
	}
	if req.AgentID != nil {
		query += `, agent_id = ?`
		args = append(args, *req.AgentID)
	}
	if req.OrganizationID != nil {
		query += `, organization_id = ?`
		args = append(args, *req.OrganizationID)
	}
	if req.Skills != nil {
		skillsJSON, _ := json.Marshal(req.Skills)
		query += `, skills = ?`
		args = append(args, string(skillsJSON))
	}
	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		query += `, config = ?`
		args = append(args, string(configJSON))
	}

	query += ` WHERE id = ?`
	args = append(args, id)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return s.Get(id)
}

// Delete deletes a TODO.
func (s *TODOService) Delete(id string) error {
	if _, err := s.Get(id); err != nil {
		return err
	}

	_, err := s.db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	return err
}

// Start starts processing a TODO.
func (s *TODOService) Start(id string) (*models.TODO, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE todos SET status = ?, started_at = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, models.TODOStatusRunning, now, now, id)
	if err != nil {
		return nil, err
	}

	todo, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastTODOStatus(id, string(models.TODOStatusRunning))
	return todo, nil
}

// Cancel cancels a TODO.
func (s *TODOService) Cancel(id string) (*models.TODO, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE todos SET status = ?, completed_at = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, models.TODOStatusCancelled, now, now, id)
	if err != nil {
		return nil, err
	}

	todo, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastTODOStatus(id, string(models.TODOStatusCancelled))
	return todo, nil
}

// UpdateProgress updates the progress of a TODO.
func (s *TODOService) UpdateProgress(id string, progress int, message string) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	_, err := s.db.Exec(`UPDATE todos SET progress = ?, updated_at = ? WHERE id = ?`,
		progress, time.Now().UTC(), id)
	if err != nil {
		return err
	}

	s.hub.BroadcastTODOProgress(id, progress, message)
	return nil
}

// Complete marks a TODO as completed.
func (s *TODOService) Complete(id string, result map[string]interface{}) (*models.TODO, error) {
	now := time.Now().UTC()
	resultJSON, _ := json.Marshal(result)

	query := `UPDATE todos SET status = ?, progress = 100, result = ?, completed_at = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, models.TODOStatusCompleted, string(resultJSON), now, now, id)
	if err != nil {
		return nil, err
	}

	todo, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastTODOStatus(id, string(models.TODOStatusCompleted))
	return todo, nil
}

// Fail marks a TODO as failed.
func (s *TODOService) Fail(id string, errMsg string) (*models.TODO, error) {
	now := time.Now().UTC()

	query := `UPDATE todos SET status = ?, error = ?, completed_at = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, models.TODOStatusFailed, errMsg, now, now, id)
	if err != nil {
		return nil, err
	}

	todo, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastTODOStatus(id, string(models.TODOStatusFailed))
	return todo, nil
}

// GetQueue returns the underlying todo queue.
func (s *TODOService) GetQueue() *todoqueue.Queue {
	return s.queue
}
