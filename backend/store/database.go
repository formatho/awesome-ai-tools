// Package store provides database initialization and migrations.
package store

import (
	"database/sql"

	"github.com/formatho/agent-orchestrator/backend/internal/store"
)

// InitDB initializes the SQLite database connection.
func InitDB(path string) (*sql.DB, error) {
	return store.InitDB(path)
}

// RunMigrations creates all necessary tables if they don't exist.
func RunMigrations(db *sql.DB) error {
	return store.RunMigrations(db)
}

// DB is an alias for sql.DB for convenience.
type DB = sql.DB
