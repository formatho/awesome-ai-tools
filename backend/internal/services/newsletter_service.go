package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// NewsletterService handles newsletter subscription operations
type NewsletterService struct {
	db *sql.DB
}

// NewNewsletterService creates a new newsletter service
func NewNewsletterService(db *sql.DB) *NewsletterService {
	return &NewsletterService{db: db}
}

// Subscribe adds a new email to the newsletter list
func (s *NewsletterService) Subscribe(req *models.SubscribeRequest) (*models.NewsletterSubscriber, error) {
	// Check if email already exists
	var existing models.NewsletterSubscriber
	err := s.db.QueryRow(
		"SELECT id, email, source, subscribed_at, unsubscribed_at, metadata, created_at, updated_at FROM newsletter_subscribers WHERE email = ?",
		req.Email,
	).Scan(&existing.ID, &existing.Email, &existing.Source, &existing.SubscribedAt, &existing.UnsubscribedAt, &existing.Metadata, &existing.CreatedAt, &existing.UpdatedAt)

	if err == nil {
		// Email exists - check if unsubscribed
		if existing.UnsubscribedAt != nil {
			// Re-subscribe
			_, err = s.db.Exec(
				"UPDATE newsletter_subscribers SET unsubscribed_at = NULL, source = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
				req.Source,
				existing.ID,
			)
			if err != nil {
				return nil, err
			}
			existing.UnsubscribedAt = nil
			existing.Source = req.Source
			return &existing, nil
		}
		// Already subscribed
		return &existing, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Create new subscriber
	id := uuid.New().String()
	now := time.Now()

	metadataJSON := req.Metadata
	if metadataJSON == "" {
		metadataJSON = "{}"
	}

	_, err = s.db.Exec(
		`INSERT INTO newsletter_subscribers (id, email, source, subscribed_at, metadata, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, req.Email, req.Source, now, metadataJSON, now, now,
	)
	if err != nil {
		return nil, err
	}

	return &models.NewsletterSubscriber{
		ID:           id,
		Email:        req.Email,
		Source:       req.Source,
		SubscribedAt: now,
		Metadata:     metadataJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Unsubscribe removes an email from the newsletter list
func (s *NewsletterService) Unsubscribe(email string) error {
	result, err := s.db.Exec(
		"UPDATE newsletter_subscribers SET unsubscribed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE email = ? AND unsubscribed_at IS NULL",
		email,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("email not found or already unsubscribed")
	}

	return nil
}

// GetStats returns newsletter subscription statistics
func (s *NewsletterService) GetStats() (*models.NewsletterStats, error) {
	stats := &models.NewsletterStats{
		BySource: make(map[string]int),
	}

	// Total subscribers
	err := s.db.QueryRow("SELECT COUNT(*) FROM newsletter_subscribers").Scan(&stats.TotalSubscribers)
	if err != nil {
		return nil, err
	}

	// Active subscribers (not unsubscribed)
	err = s.db.QueryRow("SELECT COUNT(*) FROM newsletter_subscribers WHERE unsubscribed_at IS NULL").Scan(&stats.ActiveSubscribers)
	if err != nil {
		return nil, err
	}

	// By source
	rows, err := s.db.Query("SELECT source, COUNT(*) as count FROM newsletter_subscribers WHERE unsubscribed_at IS NULL GROUP BY source")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, err
		}
		stats.BySource[source] = count
	}

	// Growth in last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = s.db.QueryRow(
		"SELECT COUNT(*) FROM newsletter_subscribers WHERE subscribed_at >= ? AND unsubscribed_at IS NULL",
		thirtyDaysAgo,
	).Scan(&stats.GrowthLast30Days)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentSubscribers returns the most recent subscribers
func (s *NewsletterService) GetRecentSubscribers(limit int) ([]models.NewsletterSubscriber, error) {
	rows, err := s.db.Query(
		`SELECT id, email, source, subscribed_at, unsubscribed_at, metadata, created_at, updated_at
		 FROM newsletter_subscribers
		 WHERE unsubscribed_at IS NULL
		 ORDER BY subscribed_at DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscribers []models.NewsletterSubscriber
	for rows.Next() {
		var sub models.NewsletterSubscriber
		err := rows.Scan(
			&sub.ID, &sub.Email, &sub.Source, &sub.SubscribedAt,
			&sub.UnsubscribedAt, &sub.Metadata, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, sub)
	}

	return subscribers, nil
}

// ExportSubscribers exports all active subscribers for email service import
func (s *NewsletterService) ExportSubscribers() ([]map[string]interface{}, error) {
	rows, err := s.db.Query(
		`SELECT id, email, source, subscribed_at, metadata
		 FROM newsletter_subscribers
		 WHERE unsubscribed_at IS NULL
		 ORDER BY subscribed_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscribers []map[string]interface{}
	for rows.Next() {
		var id, email, source, metadata string
		var subscribedAt time.Time
		err := rows.Scan(&id, &email, &source, &subscribedAt, &metadata)
		if err != nil {
			return nil, err
		}

		// Parse metadata JSON
		var metaMap map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &metaMap); err != nil {
			metaMap = make(map[string]interface{})
		}

		subscriber := map[string]interface{}{
			"id":            id,
			"email":         email,
			"source":        source,
			"subscribed_at": subscribedAt,
			"metadata":      metaMap,
		}
		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}
