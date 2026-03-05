// Package store provides database initialization and migrations.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database connection.
func InitDB(path string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // No limit for SQLite

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RunMigrations creates all necessary tables if they don't exist.
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'idle',
			provider TEXT,
			model TEXT,
			system_prompt TEXT,
			config TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			stopped_at DATETIME,
			error TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_created ON agents(created_at)`,

		`CREATE TABLE IF NOT EXISTS todos (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			priority INTEGER NOT NULL DEFAULT 0,
			progress INTEGER NOT NULL DEFAULT 0,
			agent_id TEXT,
			skills TEXT,
			dependencies TEXT,
			config TEXT,
			result TEXT,
			error TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_agent ON todos(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_created ON todos(created_at)`,

		`CREATE TABLE IF NOT EXISTS cron_jobs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			schedule TEXT NOT NULL,
			timezone TEXT DEFAULT 'UTC',
			status TEXT NOT NULL DEFAULT 'active',
			agent_id TEXT NOT NULL,
			task_name TEXT,
			task_config TEXT,
			last_run_at DATETIME,
			next_run_at DATETIME,
			last_result TEXT,
			last_error TEXT,
			run_count INTEGER NOT NULL DEFAULT 0,
			success_count INTEGER NOT NULL DEFAULT 0,
			fail_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_status ON cron_jobs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_next_run ON cron_jobs(next_run_at)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_agent ON cron_jobs(agent_id)`,

		`CREATE TABLE IF NOT EXISTS cron_history (
			id TEXT PRIMARY KEY,
			cron_id TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			status TEXT NOT NULL,
			result TEXT,
			error TEXT,
			metadata TEXT,
			FOREIGN KEY (cron_id) REFERENCES cron_jobs(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_history_cron ON cron_history(cron_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cron_history_started ON cron_history(started_at)`,

		`CREATE TABLE IF NOT EXISTS config (
			id TEXT PRIMARY KEY DEFAULT 'default',
			llm_config TEXT,
			defaults TEXT,
			settings TEXT,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT OR IGNORE INTO config (id, updated_at) VALUES ('default', CURRENT_TIMESTAMP)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Update timestamps trigger
	trigger := `CREATE TRIGGER IF NOT EXISTS update_timestamp 
		AFTER UPDATE ON agents
		BEGIN
			UPDATE agents SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END`
	db.Exec(trigger)

	trigger = `CREATE TRIGGER IF NOT EXISTS update_todo_timestamp 
		AFTER UPDATE ON todos
		BEGIN
			UPDATE todos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END`
	db.Exec(trigger)

	trigger = `CREATE TRIGGER IF NOT EXISTS update_cron_timestamp 
		AFTER UPDATE ON cron_jobs
		BEGIN
			UPDATE cron_jobs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END`
	db.Exec(trigger)

	return nil
}

// Helper functions for common queries

// GetNow returns current time in UTC.
func GetNow() time.Time {
	return time.Now().UTC()
}

// NullTime converts a time to sql.NullTime.
func NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// NullString converts a string to sql.NullString.
func NullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// ScanNullTime scans a sql.NullTime into a *time.Time.
func ScanNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// ScanNullString scans a sql.NullString into a *string.
func ScanNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
