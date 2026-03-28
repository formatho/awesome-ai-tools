package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// BetaFeedbackService handles beta feedback operations
type BetaFeedbackService struct {
	db *sql.DB
}

// NewBetaFeedbackService creates a new beta feedback service
func NewBetaFeedbackService(db *sql.DB) *BetaFeedbackService {
	return &BetaFeedbackService{db: db}
}

// Create adds new feedback
func (s *BetaFeedbackService) Create(req *models.BetaFeedbackRequest) (*models.BetaFeedback, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	// Convert tags to JSON string for storage
	tagsJSON := "[]"
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		tagsJSON = string(tagsBytes)
	}

	query := `
		INSERT INTO beta_feedback (id, user_email, user_name, category, subject, message, rating, status, priority, tags, page_url, user_agent, screenshot, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'new', 'medium', ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		id,
		req.UserEmail,
		req.UserName,
		req.Category,
		req.Subject,
		req.Message,
		req.Rating,
		tagsJSON,
		req.PageURL,
		req.UserAgent,
		req.Screenshot,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	return s.GetByID(id)
}

// GetByID retrieves feedback by ID
func (s *BetaFeedbackService) GetByID(id string) (*models.BetaFeedback, error) {
	query := `
		SELECT id, user_email, user_name, category, subject, message, rating, status, priority, tags, page_url, user_agent, screenshot, response, created_at, updated_at, resolved_at
		FROM beta_feedback
		WHERE id = ?
	`

	feedback := &models.BetaFeedback{}
	var tags, response sql.NullString
	var resolvedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&feedback.ID,
		&feedback.UserEmail,
		&feedback.UserName,
		&feedback.Category,
		&feedback.Subject,
		&feedback.Message,
		&feedback.Rating,
		&feedback.Status,
		&feedback.Priority,
		&tags,
		&feedback.PageURL,
		&feedback.UserAgent,
		&feedback.Screenshot,
		&response,
		&feedback.CreatedAt,
		&feedback.UpdatedAt,
		&resolvedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("feedback not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	// Parse tags from JSON
	if tags.Valid {
		var tagList []string
		if err := json.Unmarshal([]byte(tags.String), &tagList); err == nil {
			feedback.Tags = tagList
		}
	}

	if response.Valid {
		feedback.Response = response.String
	}
	if resolvedAt.Valid {
		feedback.ResolvedAt = &resolvedAt.Time
	}

	return feedback, nil
}

// List retrieves feedback with optional filters
func (s *BetaFeedbackService) List(status, category string, limit int) ([]*models.BetaFeedback, error) {
	var query string
	var args []interface{}

	if status != "" && category != "" {
		query = `
			SELECT id, user_email, user_name, category, subject, message, rating, status, priority, tags, page_url, user_agent, screenshot, response, created_at, updated_at, resolved_at
			FROM beta_feedback
			WHERE status = ? AND category = ?
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{status, category, limit}
	} else if status != "" {
		query = `
			SELECT id, user_email, user_name, category, subject, message, rating, status, priority, tags, page_url, user_agent, screenshot, response, created_at, updated_at, resolved_at
			FROM beta_feedback
			WHERE status = ?
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{status, limit}
	} else if category != "" {
		query = `
			SELECT id, user_email, user_name, category, subject, message, rating, status, priority, tags, page_url, user_agent, screenshot, response, created_at, updated_at, resolved_at
			FROM beta_feedback
			WHERE category = ?
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{category, limit}
	} else {
		query = `
			SELECT id, user_email, user_name, category, subject, message, rating, status, priority, tags, page_url, user_agent, screenshot, response, created_at, updated_at, resolved_at
			FROM beta_feedback
			ORDER BY created_at DESC
			LIMIT ?
		`
		args = []interface{}{limit}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback: %w", err)
	}
	defer rows.Close()

	var feedbackList []*models.BetaFeedback
	for rows.Next() {
		feedback := &models.BetaFeedback{}
		var tags, response sql.NullString
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&feedback.ID,
			&feedback.UserEmail,
			&feedback.UserName,
			&feedback.Category,
			&feedback.Subject,
			&feedback.Message,
			&feedback.Rating,
			&feedback.Status,
			&feedback.Priority,
			&tags,
			&feedback.PageURL,
			&feedback.UserAgent,
			&feedback.Screenshot,
			&response,
			&feedback.CreatedAt,
			&feedback.UpdatedAt,
			&resolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback: %w", err)
		}

		// Parse tags from JSON
		if tags.Valid {
			var tagList []string
			if err := json.Unmarshal([]byte(tags.String), &tagList); err == nil {
				feedback.Tags = tagList
			}
		}

		if response.Valid {
			feedback.Response = response.String
		}
		if resolvedAt.Valid {
			feedback.ResolvedAt = &resolvedAt.Time
		}

		feedbackList = append(feedbackList, feedback)
	}

	return feedbackList, nil
}

// UpdateStatus updates the status of feedback
func (s *BetaFeedbackService) UpdateStatus(id string, status string, response string) (*models.BetaFeedback, error) {
	now := time.Now().UTC()

	var resolvedAt interface{}
	if status == "resolved" || status == "closed" {
		resolvedAt = now
	} else {
		resolvedAt = nil
	}

	query := `
		UPDATE beta_feedback
		SET status = ?, response = ?, resolved_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query, status, response, resolvedAt, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update feedback status: %w", err)
	}

	return s.GetByID(id)
}

// GetStats retrieves feedback statistics
func (s *BetaFeedbackService) GetStats() (*models.BetaFeedbackStats, error) {
	stats := &models.BetaFeedbackStats{
		ByCategory: make(map[string]int),
		ByStatus:   make(map[string]int),
	}

	// Total count
	err := s.db.QueryRow(`SELECT COUNT(*) FROM beta_feedback`).Scan(&stats.TotalFeedback)
	if err != nil {
		return nil, fmt.Errorf("failed to get total feedback: %w", err)
	}

	// Average rating
	err = s.db.QueryRow(`SELECT COALESCE(AVG(rating), 0) FROM beta_feedback WHERE rating > 0`).Scan(&stats.AverageRating)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get average rating: %w", err)
	}

	// Count by category
	rows, err := s.db.Query(`SELECT category, COUNT(*) FROM beta_feedback GROUP BY category`)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		stats.ByCategory[category] = count
	}

	// Count by status
	rows, err = s.db.Query(`SELECT status, COUNT(*) FROM beta_feedback GROUP BY status`)
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
		stats.ByStatus[status] = count
	}

	// Recent feedback (last 7 days)
	weekAgo := time.Now().UTC().AddDate(0, 0, -7)
	err = s.db.QueryRow(`SELECT COUNT(*) FROM beta_feedback WHERE created_at >= ?`, weekAgo).Scan(&stats.RecentFeedback)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent feedback: %w", err)
	}

	return stats, nil
}

// Delete removes feedback
func (s *BetaFeedbackService) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM beta_feedback WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete feedback: %w", err)
	}
	return nil
}
