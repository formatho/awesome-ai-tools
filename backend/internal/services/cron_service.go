package services

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	cronagents "github.com/formatho/agent-orchestrator/packages/cron-agents"
	"github.com/google/uuid"
)

// CronService handles cron job operations.
type CronService struct {
	db        *sql.DB
	hub       *websocket.Hub
	scheduler *cronagents.Scheduler
}

// NewCronService creates a new cron service.
func NewCronService(db *sql.DB, hub *websocket.Hub) *CronService {
	scheduler, _ := cronagents.NewScheduler(cronagents.DefaultConfig())
	return &CronService{
		db:        db,
		hub:       hub,
		scheduler: scheduler,
	}
}

// List returns all cron jobs.
func (s *CronService) List() ([]*models.Cron, error) {
	query := `SELECT id, name, schedule, timezone, status, agent_id, task_name, task_config,
		last_run_at, next_run_at, last_result, last_error, run_count, success_count, fail_count,
		created_at, updated_at
		FROM cron_jobs ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Cron
	for rows.Next() {
		c := &models.Cron{}
		var taskName, lastResult, lastError sql.NullString
		var lastRunAt, nextRunAt sql.NullTime
		var taskConfig sql.NullString

		err := rows.Scan(
			&c.ID, &c.Name, &c.Schedule, &c.Timezone, &c.Status, &c.AgentID, &taskName, &taskConfig,
			&lastRunAt, &nextRunAt, &lastResult, &lastError, &c.RunCount, &c.SuccessCount, &c.FailCount,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		c.TaskName = taskName.String
		c.LastResult = lastResult.String
		c.LastError = lastError.String

		c.LastRunAt = &lastRunAt.Time
		if !lastRunAt.Valid {
			c.LastRunAt = nil
		}
		c.NextRunAt = &nextRunAt.Time
		if !nextRunAt.Valid {
			c.NextRunAt = nil
		}

		if taskConfig.Valid && taskConfig.String != "" {
			json.Unmarshal([]byte(taskConfig.String), &c.TaskConfig)
		}

		jobs = append(jobs, c)
	}

	return jobs, nil
}

// Get returns a single cron job by ID.
func (s *CronService) Get(id string) (*models.Cron, error) {
	query := `SELECT id, name, schedule, timezone, status, agent_id, task_name, task_config,
		last_run_at, next_run_at, last_result, last_error, run_count, success_count, fail_count,
		created_at, updated_at
		FROM cron_jobs WHERE id = ?`

	c := &models.Cron{}
	var taskName, lastResult, lastError sql.NullString
	var lastRunAt, nextRunAt sql.NullTime
	var taskConfig sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&c.ID, &c.Name, &c.Schedule, &c.Timezone, &c.Status, &c.AgentID, &taskName, &taskConfig,
		&lastRunAt, &nextRunAt, &lastResult, &lastError, &c.RunCount, &c.SuccessCount, &c.FailCount,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFoundWithID("CronJob", id)
	}
	if err != nil {
		return nil, err
	}

	c.TaskName = taskName.String
	c.LastResult = lastResult.String
	c.LastError = lastError.String

	c.LastRunAt = &lastRunAt.Time
	if !lastRunAt.Valid {
		c.LastRunAt = nil
	}
	c.NextRunAt = &nextRunAt.Time
	if !nextRunAt.Valid {
		c.NextRunAt = nil
	}

	if taskConfig.Valid && taskConfig.String != "" {
		json.Unmarshal([]byte(taskConfig.String), &c.TaskConfig)
	}

	return c, nil
}

// Create creates a new cron job.
func (s *CronService) Create(req *models.CronCreate) (*models.Cron, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	status := models.CronStatusActive
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	taskConfigJSON, _ := json.Marshal(req.TaskConfig)

	query := `INSERT INTO cron_jobs (id, name, schedule, timezone, status, agent_id, task_name, task_config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, req.Name, req.Schedule, timezone, status, req.AgentID,
		req.TaskName, string(taskConfigJSON), now, now)
	if err != nil {
		return nil, err
	}

	job, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastCronTriggered(id, map[string]string{"action": "created"})
	return job, nil
}

// Update updates an existing cron job.
func (s *CronService) Update(id string, req *models.CronUpdate) (*models.Cron, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE cron_jobs SET updated_at = ?`
	args := []interface{}{now}

	if req.Name != nil {
		query += `, name = ?`
		args = append(args, *req.Name)
	}
	if req.Schedule != nil {
		query += `, schedule = ?`
		args = append(args, *req.Schedule)
	}
	if req.Timezone != nil {
		query += `, timezone = ?`
		args = append(args, *req.Timezone)
	}
	if req.AgentID != nil {
		query += `, agent_id = ?`
		args = append(args, *req.AgentID)
	}
	if req.TaskName != nil {
		query += `, task_name = ?`
		args = append(args, *req.TaskName)
	}
	if req.TaskConfig != nil {
		configJSON, _ := json.Marshal(req.TaskConfig)
		query += `, task_config = ?`
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

// Delete deletes a cron job.
func (s *CronService) Delete(id string) error {
	if _, err := s.Get(id); err != nil {
		return err
	}

	_, err := s.db.Exec(`DELETE FROM cron_jobs WHERE id = ?`, id)
	return err
}

// Pause pauses a cron job.
func (s *CronService) Pause(id string) (*models.Cron, error) {
	return s.updateStatus(id, models.CronStatusPaused)
}

// Resume resumes a paused cron job.
func (s *CronService) Resume(id string) (*models.Cron, error) {
	return s.updateStatus(id, models.CronStatusActive)
}

func (s *CronService) updateStatus(id string, status models.CronStatus) (*models.Cron, error) {
	if _, err := s.Get(id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	query := `UPDATE cron_jobs SET status = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, now, id)
	if err != nil {
		return nil, err
	}

	job, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.hub.BroadcastCronStatus(id, string(status))
	return job, nil
}

// GetHistory returns the execution history for a cron job.
func (s *CronService) GetHistory(id string, limit int) ([]*models.CronHistory, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, cron_id, started_at, ended_at, status, result, error, metadata
		FROM cron_history WHERE cron_id = ? ORDER BY started_at DESC LIMIT ?`

	rows, err := s.db.Query(query, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.CronHistory
	for rows.Next() {
		h := &models.CronHistory{}
		var endedAt sql.NullTime
		var result, errMsg, metadata sql.NullString

		err := rows.Scan(
			&h.ID, &h.CronID, &h.StartedAt, &endedAt, &h.Status, &result, &errMsg, &metadata,
		)
		if err != nil {
			return nil, err
		}

		h.EndedAt = &endedAt.Time
		if !endedAt.Valid {
			h.EndedAt = nil
		}
		h.Result = result.String
		h.Error = errMsg.String

		if metadata.Valid && metadata.String != "" {
			json.Unmarshal([]byte(metadata.String), &h.Metadata)
		}

		history = append(history, h)
	}

	return history, nil
}

// RecordRun records a cron job execution in history.
func (s *CronService) RecordRun(cronID string, status string, result string, errMsg string) error {
	id := uuid.New().String()
	now := time.Now().UTC()

	query := `INSERT INTO cron_history (id, cron_id, started_at, ended_at, status, result, error)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, cronID, now, now, status, result, errMsg)
	if err != nil {
		return err
	}

	// Update cron job stats
	updateQuery := `UPDATE cron_jobs SET 
		last_run_at = ?, last_result = ?, last_error = ?, 
		run_count = run_count + 1, 
		success_count = CASE WHEN ? = 'success' THEN success_count + 1 ELSE success_count END,
		fail_count = CASE WHEN ? = 'failed' THEN fail_count + 1 ELSE fail_count END,
		updated_at = ?
		WHERE id = ?`

	_, err = s.db.Exec(updateQuery, now, result, errMsg, status, status, now, cronID)
	if err != nil {
		return err
	}

	s.hub.BroadcastCronTriggered(cronID, map[string]string{"status": status})
	return nil
}

// GetScheduler returns the underlying scheduler.
func (s *CronService) GetScheduler() *cronagents.Scheduler {
	return s.scheduler
}
