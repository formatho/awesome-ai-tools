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
		// Organizations table
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			owner_id TEXT NOT NULL,
			settings TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_organizations_owner ON organizations(owner_id)`,

		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'idle',
			provider TEXT,
			model TEXT,
			base_url TEXT,
			system_prompt TEXT,
			work_dir TEXT DEFAULT '~/sandbox',
			organization_id TEXT,
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
			organization_id TEXT,
			skills TEXT,
			dependencies TEXT,
			config TEXT,
			result TEXT,
			error TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME
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
			organization_id TEXT,
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
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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

		`CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_agent ON chat_messages(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_created ON chat_messages(created_at)`,

		`CREATE TABLE IF NOT EXISTS agent_logs (
			id TEXT PRIMARY KEY,
			agent_id TEXT NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_agent ON agent_logs(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created ON agent_logs(created_at)`,

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT,
			avatar_url TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		// Organization members table
		`CREATE TABLE IF NOT EXISTS user_org_members (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			status TEXT NOT NULL DEFAULT 'active',
			joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			metadata TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
			UNIQUE(user_id, organization_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_user ON user_org_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_org ON user_org_members(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_role ON user_org_members(role)`,

		// Team invitations table
		`CREATE TABLE IF NOT EXISTS invitations (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			user_id TEXT,
			organization_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			status TEXT NOT NULL DEFAULT 'pending',
			token TEXT NOT NULL UNIQUE,
			message TEXT,
			expires_at DATETIME NOT NULL,
			sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			accepted_at DATETIME,
			rejected_at DATETIME,
			cancelled_at DATETIME,
			created_by TEXT NOT NULL,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			FOREIGN KEY (created_by) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_org ON invitations(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(status)`,

		// Permissions table
		`CREATE TABLE IF NOT EXISTS permissions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			permission TEXT NOT NULL,
			granted_by TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
			FOREIGN KEY (granted_by) REFERENCES users(id),
			UNIQUE(user_id, organization_id, permission)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_user ON permissions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_org ON permissions(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_perm ON permissions(permission)`,

		// Permission templates table
		`CREATE TABLE IF NOT EXISTS permission_templates (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			permissions TEXT NOT NULL,
			created_by TEXT NOT NULL,
			organization_id TEXT,
			is_default INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES users(id),
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_perm_templates_org ON permission_templates(organization_id)`,

		// Beta signups table
		`CREATE TABLE IF NOT EXISTS beta_signups (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			role TEXT,
			use_case TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_signups_email ON beta_signups(email)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_signups_status ON beta_signups(status)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_signups_created ON beta_signups(created_at)`,

		// Beta feedback table
		`CREATE TABLE IF NOT EXISTS beta_feedback (
			id TEXT PRIMARY KEY,
			user_email TEXT NOT NULL,
			user_name TEXT NOT NULL,
			category TEXT NOT NULL,
			subject TEXT NOT NULL,
			message TEXT NOT NULL,
			rating INTEGER,
			status TEXT NOT NULL DEFAULT 'new',
			priority TEXT NOT NULL DEFAULT 'medium',
			tags TEXT DEFAULT '[]',
			page_url TEXT,
			user_agent TEXT,
			screenshot TEXT,
			response TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_feedback_email ON beta_feedback(user_email)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_feedback_category ON beta_feedback(category)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_feedback_status ON beta_feedback(status)`,
		`CREATE INDEX IF NOT EXISTS idx_beta_feedback_created ON beta_feedback(created_at)`,
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
	_, _ = db.Exec(trigger)

	// Create indexes for organization_id (these are safe to run multiple times with IF NOT EXISTS)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_agents_organization ON agents(organization_id)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_organization ON todos(organization_id)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_cron_organization ON cron_jobs(organization_id)`)

	// Add update timestamp trigger for organizations
	orgTrigger := `CREATE TRIGGER IF NOT EXISTS update_org_timestamp
		AFTER UPDATE ON organizations
		BEGIN
			UPDATE organizations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END`
	_, _ = db.Exec(orgTrigger)

	trigger = `CREATE TRIGGER IF NOT EXISTS update_todo_timestamp 
		AFTER UPDATE ON todos
		BEGIN
			UPDATE todos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END`
	_, _ = db.Exec(trigger)

	trigger = `CREATE TRIGGER IF NOT EXISTS update_cron_timestamp 
		AFTER UPDATE ON cron_jobs
		BEGIN
			UPDATE cron_jobs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END`
	_, _ = db.Exec(trigger)

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
