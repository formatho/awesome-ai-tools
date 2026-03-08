package cron

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// RunStatus represents the status of a job execution.
type RunStatus string

const (
	StatusPending   RunStatus = "pending"
	StatusRunning   RunStatus = "running"
	StatusSuccess   RunStatus = "success"
	StatusFailed    RunStatus = "failed"
	StatusRetry     RunStatus = "retry"
	StatusCancelled RunStatus = "cancelled"
)

// RunHistory represents a single execution record of a job.
type RunHistory struct {
	// ID is the unique identifier for this run record.
	ID int64 `json:"id"`

	// JobID is the ID of the job that was executed.
	JobID string `json:"job_id"`

	// ScheduledFor is the time the job was scheduled to run.
	ScheduledFor time.Time `json:"scheduled_for"`

	// StartedAt is when the job actually started execution.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the job finished execution.
	CompletedAt time.Time `json:"completed_at,omitempty"`

	// Status indicates the outcome of the run.
	Status RunStatus `json:"status"`

	// Result contains the output or result data from the job.
	// Stored as JSON for flexibility.
	Result interface{} `json:"result,omitempty"`

	// Error contains the error message if the run failed.
	Error string `json:"error,omitempty"`

	// RetryCount is the number of retry attempts for this run.
	RetryCount int `json:"retry_count"`

	// Duration is the execution duration in milliseconds.
	Duration int64 `json:"duration,omitempty"`

	// Metadata stores additional run-specific information.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HistoryStore manages persistent storage of run history.
type HistoryStore struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewHistoryStore creates a new history store with the given database path.
// If path is empty, an in-memory database is used.
func NewHistoryStore(path string) (*HistoryStore, error) {
	dsn := path
	if dsn == "" {
		dsn = ":memory:"
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and WAL mode for better performance
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	store := &HistoryStore{
		db:   db,
		path: path,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the database schema if it doesn't exist.
func (s *HistoryStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS run_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT NOT NULL,
		scheduled_for DATETIME NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		status TEXT NOT NULL,
		result TEXT,
		error TEXT,
		retry_count INTEGER DEFAULT 0,
		duration INTEGER,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_run_history_job_id ON run_history(job_id);
	CREATE INDEX IF NOT EXISTS idx_run_history_status ON run_history(status);
	CREATE INDEX IF NOT EXISTS idx_run_history_scheduled_for ON run_history(scheduled_for);
	CREATE INDEX IF NOT EXISTS idx_run_history_started_at ON run_history(started_at);

	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		schedule TEXT NOT NULL,
		timezone TEXT,
		agent_id TEXT NOT NULL,
		todo TEXT,
		enabled INTEGER DEFAULT 1,
		last_run DATETIME,
		next_run DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		metadata TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_agent_id ON jobs(agent_id);
	CREATE INDEX IF NOT EXISTS idx_jobs_enabled ON jobs(enabled);
	CREATE INDEX IF NOT EXISTS idx_jobs_next_run ON jobs(next_run);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection.
func (s *HistoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

// AddRun creates a new run history record.
func (s *HistoryStore) AddRun(history *RunHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resultJSON, _ := json.Marshal(history.Result)
	metadataJSON, _ := json.Marshal(history.Metadata)

	var completedAt interface{}
	if !history.CompletedAt.IsZero() {
		completedAt = history.CompletedAt
	}

	query := `
		INSERT INTO run_history (job_id, scheduled_for, started_at, completed_at, status, result, error, retry_count, duration, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		history.JobID,
		history.ScheduledFor,
		history.StartedAt,
		completedAt,
		string(history.Status),
		string(resultJSON),
		history.Error,
		history.RetryCount,
		history.Duration,
		string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to add run history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	history.ID = id

	return nil
}

// UpdateRun updates an existing run history record.
func (s *HistoryStore) UpdateRun(history *RunHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resultJSON, _ := json.Marshal(history.Result)
	metadataJSON, _ := json.Marshal(history.Metadata)

	query := `
		UPDATE run_history
		SET completed_at = ?, status = ?, result = ?, error = ?, retry_count = ?, duration = ?, metadata = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		history.CompletedAt,
		string(history.Status),
		string(resultJSON),
		history.Error,
		history.RetryCount,
		history.Duration,
		string(metadataJSON),
		history.ID,
	)

	return err
}

// GetHistory retrieves run history for a specific job.
func (s *HistoryStore) GetHistory(jobID string, limit int) ([]RunHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, job_id, scheduled_for, started_at, completed_at, status, result, error, retry_count, duration, metadata
		FROM run_history
		WHERE job_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, jobID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var histories []RunHistory
	for rows.Next() {
		var h RunHistory
		var status string
		var resultJSON, metadataJSON sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&h.ID,
			&h.JobID,
			&h.ScheduledFor,
			&h.StartedAt,
			&completedAt,
			&status,
			&resultJSON,
			&h.Error,
			&h.RetryCount,
			&h.Duration,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}

		h.Status = RunStatus(status)
		if completedAt.Valid {
			h.CompletedAt = completedAt.Time
		}
		if resultJSON.Valid && resultJSON.String != "" {
			_ = json.Unmarshal([]byte(resultJSON.String), &h.Result)
		}
		if metadataJSON.Valid && metadataJSON.String != "" {
			_ = json.Unmarshal([]byte(metadataJSON.String), &h.Metadata)
		}

		histories = append(histories, h)
	}

	return histories, rows.Err()
}

// GetAllHistory retrieves all run history with optional filtering.
func (s *HistoryStore) GetAllHistory(limit int) ([]RunHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, job_id, scheduled_for, started_at, completed_at, status, result, error, retry_count, duration, metadata
		FROM run_history
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get all history: %w", err)
	}
	defer rows.Close()

	return s.scanHistories(rows)
}

// GetHistoryByStatus retrieves run history filtered by status.
func (s *HistoryStore) GetHistoryByStatus(status RunStatus, limit int) ([]RunHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, job_id, scheduled_for, started_at, completed_at, status, result, error, retry_count, duration, metadata
		FROM run_history
		WHERE status = ?
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, string(status), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get history by status: %w", err)
	}
	defer rows.Close()

	return s.scanHistories(rows)
}

// GetHistoryInRange retrieves run history within a time range.
func (s *HistoryStore) GetHistoryInRange(start, end time.Time) ([]RunHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, job_id, scheduled_for, started_at, completed_at, status, result, error, retry_count, duration, metadata
		FROM run_history
		WHERE started_at >= ? AND started_at <= ?
		ORDER BY started_at DESC
	`

	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get history in range: %w", err)
	}
	defer rows.Close()

	return s.scanHistories(rows)
}

// DeleteOldHistory removes history records older than the specified duration.
func (s *HistoryStore) DeleteOldHistory(olderThan time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	result, err := s.db.Exec(
		"DELETE FROM run_history WHERE started_at < ?",
		cutoff,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetStats returns statistics about run history.
func (s *HistoryStore) GetStats() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})

	// Total runs
	var totalRuns int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM run_history").Scan(&totalRuns)
	if err != nil {
		return nil, err
	}
	stats["total_runs"] = totalRuns

	// Runs by status
	rows, err := s.db.Query("SELECT status, COUNT(*) FROM run_history GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}
	stats["by_status"] = statusCounts

	// Average duration for successful runs
	var avgDuration sql.NullFloat64
	err = s.db.QueryRow(
		"SELECT AVG(duration) FROM run_history WHERE status = ? AND duration > 0",
		string(StatusSuccess),
	).Scan(&avgDuration)
	if err == nil && avgDuration.Valid {
		stats["avg_duration_ms"] = avgDuration.Float64
	}

	return stats, nil
}

// SaveJob persists a job to the database.
func (s *HistoryStore) SaveJob(job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	todoJSON, _ := json.Marshal(job.TODO)
	metadataJSON, _ := json.Marshal(job.Metadata)

	query := `
		INSERT OR REPLACE INTO jobs (id, name, schedule, timezone, agent_id, todo, enabled, last_run, next_run, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		job.ID,
		job.Name,
		job.Schedule,
		job.Timezone,
		job.AgentID,
		string(todoJSON),
		job.Enabled,
		job.LastRun,
		job.NextRun,
		job.CreatedAt,
		job.UpdatedAt,
		string(metadataJSON),
	)

	return err
}

// LoadJob loads a job from the database by ID.
func (s *HistoryStore) LoadJob(id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, name, schedule, timezone, agent_id, todo, enabled, last_run, next_run, created_at, updated_at, metadata
		FROM jobs
		WHERE id = ?
	`

	var job Job
	var todoJSON, metadataJSON sql.NullString
	var lastRun, nextRun sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.Name,
		&job.Schedule,
		&job.Timezone,
		&job.AgentID,
		&todoJSON,
		&job.Enabled,
		&lastRun,
		&nextRun,
		&job.CreatedAt,
		&job.UpdatedAt,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("job not found")
	}
	if err != nil {
		return nil, err
	}

	if lastRun.Valid {
		job.LastRun = lastRun.Time
	}
	if nextRun.Valid {
		job.NextRun = nextRun.Time
	}
	if todoJSON.Valid && todoJSON.String != "" {
		_ = json.Unmarshal([]byte(todoJSON.String), &job.TODO)
	}
	if metadataJSON.Valid && metadataJSON.String != "" {
		_ = json.Unmarshal([]byte(metadataJSON.String), &job.Metadata)
	}

	return &job, nil
}

// LoadAllJobs loads all jobs from the database.
func (s *HistoryStore) LoadAllJobs() ([]Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, name, schedule, timezone, agent_id, todo, enabled, last_run, next_run, created_at, updated_at, metadata
		FROM jobs
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		var todoJSON, metadataJSON sql.NullString
		var lastRun, nextRun sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.Name,
			&job.Schedule,
			&job.Timezone,
			&job.AgentID,
			&todoJSON,
			&job.Enabled,
			&lastRun,
			&nextRun,
			&job.CreatedAt,
			&job.UpdatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		if lastRun.Valid {
			job.LastRun = lastRun.Time
		}
		if nextRun.Valid {
			job.NextRun = nextRun.Time
		}
		if todoJSON.Valid && todoJSON.String != "" {
			_ = json.Unmarshal([]byte(todoJSON.String), &job.TODO)
		}
		if metadataJSON.Valid && metadataJSON.String != "" {
			_ = json.Unmarshal([]byte(metadataJSON.String), &job.Metadata)
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// DeleteJob removes a job from the database.
func (s *HistoryStore) DeleteJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM jobs WHERE id = ?", id)
	return err
}

// scanHistories is a helper to scan multiple history rows.
func (s *HistoryStore) scanHistories(rows *sql.Rows) ([]RunHistory, error) {
	var histories []RunHistory
	for rows.Next() {
		var h RunHistory
		var status string
		var resultJSON, metadataJSON sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&h.ID,
			&h.JobID,
			&h.ScheduledFor,
			&h.StartedAt,
			&completedAt,
			&status,
			&resultJSON,
			&h.Error,
			&h.RetryCount,
			&h.Duration,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		h.Status = RunStatus(status)
		if completedAt.Valid {
			h.CompletedAt = completedAt.Time
		}
		if resultJSON.Valid && resultJSON.String != "" {
			_ = json.Unmarshal([]byte(resultJSON.String), &h.Result)
		}
		if metadataJSON.Valid && metadataJSON.String != "" {
			_ = json.Unmarshal([]byte(metadataJSON.String), &h.Metadata)
		}

		histories = append(histories, h)
	}

	return histories, rows.Err()
}

// DBPath returns the database file path.
func (s *HistoryStore) DBPath() string {
	return s.path
}

// DBExists checks if the database file exists.
func DBExists(path string) bool {
	if path == "" || path == ":memory:" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
