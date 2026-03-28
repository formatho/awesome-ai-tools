package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// BetaFeedbackService handles beta feedback business logic
type BetaFeedbackService struct {
	db *sql.DB
}

// NewBetaFeedbackService creates a new beta feedback service
func NewBetaFeedbackService(db *sql.DB) *BetaFeedbackService {
	return &BetaFeedbackService{db: db}
}

// Create creates a new beta feedback entry
func (s *BetaFeedbackService) Create(feedback *models.BetaFeedback) (*models.BetaFeedback, error) {
	if feedback.ID == "" {
		feedback.ID = uuid.New().String()
	}

	query := `
		INSERT INTO beta_feedback (
			id, type, rating, title, description, email, name, priority,
			browser, steps_to_reproduce, expected_behavior, actual_behavior,
			attachments, status, beta_tester_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`

	attachmentsJSON, _ := jsonMarshal(feedback.Attachments)

	err := s.db.QueryRow(
		query,
		feedback.ID,
		feedback.Type,
		feedback.Rating,
		feedback.Title,
		feedback.Description,
		feedback.Email,
		feedback.Name,
		feedback.Priority,
		feedback.Browser,
		feedback.StepsToReproduce,
		feedback.ExpectedBehavior,
		feedback.ActualBehavior,
		attachmentsJSON,
		feedback.Status,
		feedback.BetaTesterID,
		feedback.CreatedAt,
		feedback.UpdatedAt,
	).Scan(&feedback.ID)

	if err != nil {
		return nil, fmt.Errorf("error creating feedback: %w", err)
	}

	return feedback, nil
}

// GetByID retrieves feedback by ID
func (s *BetaFeedbackService) GetByID(id string) (*models.BetaFeedback, error) {
	query := `
		SELECT id, type, rating, title, description, email, name, priority,
			   browser, steps_to_reproduce, expected_behavior, actual_behavior,
			   attachments, status, resolution, assigned_to, beta_tester_id,
			   created_at, updated_at, resolved_at
		FROM beta_feedback
		WHERE id = $1
	`

	feedback := &models.BetaFeedback{}
	var attachmentsJSON []byte
	var resolvedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&feedback.ID,
		&feedback.Type,
		&feedback.Rating,
		&feedback.Title,
		&feedback.Description,
		&feedback.Email,
		&feedback.Name,
		&feedback.Priority,
		&feedback.Browser,
		&feedback.StepsToReproduce,
		&feedback.ExpectedBehavior,
		&feedback.ActualBehavior,
		&attachmentsJSON,
		&feedback.Status,
		&feedback.Resolution,
		&feedback.AssignedTo,
		&feedback.BetaTesterID,
		&feedback.CreatedAt,
		&feedback.UpdatedAt,
		&resolvedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting feedback: %w", err)
	}

	if resolvedAt.Valid {
		feedback.ResolvedAt = &resolvedAt.Time
	}

	if len(attachmentsJSON) > 0 {
		jsonUnmarshal(attachmentsJSON, &feedback.Attachments)
	}

	return feedback, nil
}

// List retrieves feedback with optional filters
func (s *BetaFeedbackService) List(status, feedbackType, priority string) ([]*models.BetaFeedback, error) {
	query := `
		SELECT id, type, rating, title, description, email, name, priority,
			   browser, steps_to_reproduce, expected_behavior, actual_behavior,
			   attachments, status, resolution, assigned_to, beta_tester_id,
			   created_at, updated_at, resolved_at
		FROM beta_feedback
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, status)
		argNum++
	}

	if feedbackType != "" {
		query += fmt.Sprintf(" AND type = $%d", argNum)
		args = append(args, feedbackType)
		argNum++
	}

	if priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", argNum)
		args = append(args, priority)
		argNum++
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []*models.BetaFeedback
	for rows.Next() {
		feedback := &models.BetaFeedback{}
		var attachmentsJSON []byte
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&feedback.ID,
			&feedback.Type,
			&feedback.Rating,
			&feedback.Title,
			&feedback.Description,
			&feedback.Email,
			&feedback.Name,
			&feedback.Priority,
			&feedback.Browser,
			&feedback.StepsToReproduce,
			&feedback.ExpectedBehavior,
			&feedback.ActualBehavior,
			&attachmentsJSON,
			&feedback.Status,
			&feedback.Resolution,
			&feedback.AssignedTo,
			&feedback.BetaTesterID,
			&feedback.CreatedAt,
			&feedback.UpdatedAt,
			&resolvedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning feedback: %w", err)
		}

		if resolvedAt.Valid {
			feedback.ResolvedAt = &resolvedAt.Time
		}

		if len(attachmentsJSON) > 0 {
			jsonUnmarshal(attachmentsJSON, &feedback.Attachments)
		}

		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// UpdateStatus updates feedback status
func (s *BetaFeedbackService) UpdateStatus(id, status, resolution, assignedTo string) (*models.BetaFeedback, error) {
	query := `
		UPDATE beta_feedback
		SET status = $1, resolution = $2, assigned_to = $3, updated_at = $4
	`
	args := []interface{}{status, resolution, assignedTo, time.Now()}
	argNum := 5

	if status == "resolved" || status == "closed" {
		query += fmt.Sprintf(", resolved_at = $%d", argNum)
		args = append(args, time.Now())
		argNum++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id", argNum)
	args = append(args, id)

	err := s.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("error updating feedback: %w", err)
	}

	return s.GetByID(id)
}

// GetStats retrieves feedback statistics
func (s *BetaFeedbackService) GetStats() (*models.BetaFeedbackStats, error) {
	stats := &models.BetaFeedbackStats{
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		ByPriority: make(map[string]int),
	}

	// Total count
	err := s.db.QueryRow("SELECT COUNT(*) FROM beta_feedback").Scan(&stats.Total)
	if err != nil {
		return nil, fmt.Errorf("error getting total count: %w", err)
	}

	// By type
	rows, err := s.db.Query("SELECT type, COUNT(*) FROM beta_feedback GROUP BY type")
	if err != nil {
		return nil, fmt.Errorf("error getting by type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var typ string
		var count int
		rows.Scan(&typ, &count)
		stats.ByType[typ] = count
	}

	// By status
	rows, err = s.db.Query("SELECT status, COUNT(*) FROM beta_feedback GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("error getting by status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats.ByStatus[status] = count
	}

	// By priority
	rows, err = s.db.Query("SELECT priority, COUNT(*) FROM beta_feedback GROUP BY priority")
	if err != nil {
		return nil, fmt.Errorf("error getting by priority: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var priority string
		var count int
		rows.Scan(&priority, &count)
		stats.ByPriority[priority] = count
	}

	// Average rating
	err = s.db.QueryRow("SELECT AVG(rating) FROM beta_feedback WHERE rating > 0").Scan(&stats.AverageRating)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting average rating: %w", err)
	}

	// Recent count (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	err = s.db.QueryRow("SELECT COUNT(*) FROM beta_feedback WHERE created_at > $1", sevenDaysAgo).Scan(&stats.RecentCount)
	if err != nil {
		return nil, fmt.Errorf("error getting recent count: %w", err)
	}

	return stats, nil
}

// Helper functions
func jsonMarshal(v interface{}) ([]byte, error) {
	// Implement or use encoding/json
	return []byte("[]"), nil
}

func jsonUnmarshal(data []byte, v interface{}) error {
	// Implement or use encoding/json
	return nil
}
