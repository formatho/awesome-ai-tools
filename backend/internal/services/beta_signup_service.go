package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// BetaSignupService handles beta signup operations
type BetaSignupService struct {
	db *sql.DB
}

// NewBetaSignupService creates a new beta signup service
func NewBetaSignupService(db *sql.DB) *BetaSignupService {
	return &BetaSignupService{db: db}
}

// Create adds a new beta signup
func (s *BetaSignupService) Create(req *models.BetaSignupRequest) (*models.BetaSignup, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	query := `
		INSERT INTO beta_signups (id, name, email, role, use_case, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?)
	`

	_, err := s.db.Exec(query, id, req.Name, req.Email, req.Role, req.UseCase, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create beta signup: %w", err)
	}

	return s.GetByID(id)
}

// GetByID retrieves a beta signup by ID
func (s *BetaSignupService) GetByID(id string) (*models.BetaSignup, error) {
	query := `
		SELECT id, name, email, role, use_case, status, notes, created_at, updated_at, reviewed_at
		FROM beta_signups
		WHERE id = ?
	`

	signup := &models.BetaSignup{}
	var reviewedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&signup.ID,
		&signup.Name,
		&signup.Email,
		&signup.Role,
		&signup.UseCase,
		&signup.Status,
		&signup.Notes,
		&signup.CreatedAt,
		&signup.UpdatedAt,
		&reviewedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("beta signup not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get beta signup: %w", err)
	}

	if reviewedAt.Valid {
		signup.ReviewedAt = &reviewedAt.Time
	}

	return signup, nil
}

// GetByEmail retrieves a beta signup by email
func (s *BetaSignupService) GetByEmail(email string) (*models.BetaSignup, error) {
	query := `
		SELECT id, name, email, role, use_case, status, notes, created_at, updated_at, reviewed_at
		FROM beta_signups
		WHERE email = ?
	`

	signup := &models.BetaSignup{}
	var reviewedAt sql.NullTime

	err := s.db.QueryRow(query, email).Scan(
		&signup.ID,
		&signup.Name,
		&signup.Email,
		&signup.Role,
		&signup.UseCase,
		&signup.Status,
		&signup.Notes,
		&signup.CreatedAt,
		&signup.UpdatedAt,
		&reviewedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get beta signup: %w", err)
	}

	if reviewedAt.Valid {
		signup.ReviewedAt = &reviewedAt.Time
	}

	return signup, nil
}

// List retrieves all beta signups
func (s *BetaSignupService) List(status string) ([]*models.BetaSignup, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, name, email, role, use_case, status, notes, created_at, updated_at, reviewed_at
			FROM beta_signups
			WHERE status = ?
			ORDER BY created_at DESC
		`
		args = []interface{}{status}
	} else {
		query = `
			SELECT id, name, email, role, use_case, status, notes, created_at, updated_at, reviewed_at
			FROM beta_signups
			ORDER BY created_at DESC
		`
		args = []interface{}{}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list beta signups: %w", err)
	}
	defer rows.Close()

	var signups []*models.BetaSignup
	for rows.Next() {
		signup := &models.BetaSignup{}
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&signup.ID,
			&signup.Name,
			&signup.Email,
			&signup.Role,
			&signup.UseCase,
			&signup.Status,
			&signup.Notes,
			&signup.CreatedAt,
			&signup.UpdatedAt,
			&reviewedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan beta signup: %w", err)
		}

		if reviewedAt.Valid {
			signup.ReviewedAt = &reviewedAt.Time
		}

		signups = append(signups, signup)
	}

	return signups, nil
}

// UpdateStatus updates the status of a beta signup
func (s *BetaSignupService) UpdateStatus(id string, status string, notes string) (*models.BetaSignup, error) {
	now := time.Now().UTC()

	query := `
		UPDATE beta_signups
		SET status = ?, notes = ?, reviewed_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, status, notes, now, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update beta signup status: %w", err)
	}

	return s.GetByID(id)
}

// GetStats retrieves beta signup statistics
func (s *BetaSignupService) GetStats() (*models.BetaSignupStats, error) {
	stats := &models.BetaSignupStats{
		ByRole: make(map[string]int),
	}

	// Total count
	err := s.db.QueryRow(`SELECT COUNT(*) FROM beta_signups`).Scan(&stats.TotalSignups)
	if err != nil {
		return nil, fmt.Errorf("failed to get total signups: %w", err)
	}

	// Count by status
	rows, err := s.db.Query(`SELECT status, COUNT(*) FROM beta_signups GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}

		switch status {
		case "pending":
			stats.PendingCount = count
		case "accepted":
			stats.AcceptedCount = count
		case "rejected":
			stats.RejectedCount = count
		}
	}

	// Count by role
	rows, err = s.db.Query(`SELECT role, COUNT(*) FROM beta_signups WHERE role != '' GROUP BY role`)
	if err != nil {
		return nil, fmt.Errorf("failed to get role counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, err
		}
		stats.ByRole[role] = count
	}

	return stats, nil
}

// Delete removes a beta signup
func (s *BetaSignupService) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM beta_signups WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete beta signup: %w", err)
	}
	return nil
}
