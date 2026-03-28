package models

import "time"

// BetaFeedback represents a beta tester's feedback
type BetaFeedback struct {
	ID               string    `json:"id" db:"id"`
	Type             string    `json:"type" db:"type"` // bug, feature, testimonial, general
	Rating           int       `json:"rating" db:"rating"` // 1-5
	Title            string    `json:"title" db:"title"`
	Description      string    `json:"description" db:"description"`
	Email            string    `json:"email" db:"email"`
	Name             string    `json:"name" db:"name"`
	Priority         string    `json:"priority" db:"priority"` // low, medium, high, critical
	Browser          string    `json:"browser" db:"browser"`
	StepsToReproduce string    `json:"steps_to_reproduce,omitempty" db:"steps_to_reproduce"`
	ExpectedBehavior string    `json:"expected_behavior,omitempty" db:"expected_behavior"`
	ActualBehavior   string    `json:"actual_behavior,omitempty" db:"actual_behavior"`
	Attachments      []string  `json:"attachments,omitempty" db:"attachments"`
	Status           string    `json:"status" db:"status"` // new, in_progress, resolved, closed
	Resolution       string    `json:"resolution,omitempty" db:"resolution"`
	AssignedTo       string    `json:"assigned_to,omitempty" db:"assigned_to"`
	BetaTesterID     string    `json:"beta_tester_id,omitempty" db:"beta_tester_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// BetaFeedbackRequest is the request body for submitting feedback
type BetaFeedbackRequest struct {
	Type              string   `json:"type"`
	Rating            int      `json:"rating"`
	Title             string   `json:"title" validate:"required"`
	Description       string   `json:"description" validate:"required"`
	Email             string   `json:"email" validate:"required,email"`
	Name              string   `json:"name"`
	Priority          string   `json:"priority"`
	Browser           string   `json:"browser"`
	StepsToReproduce  string   `json:"steps_to_reproduce"`
	ExpectedBehavior  string   `json:"expected_behavior"`
	ActualBehavior    string   `json:"actual_behavior"`
	Attachments       []string `json:"attachments"`
	BetaTesterID      string   `json:"beta_tester_id"`
}

// BetaFeedbackStats represents statistics about beta feedback
type BetaFeedbackStats struct {
	Total         int            `json:"total"`
	ByType        map[string]int `json:"by_type"`
	ByStatus      map[string]int `json:"by_status"`
	ByPriority    map[string]int `json:"by_priority"`
	AverageRating float64        `json:"average_rating"`
	RecentCount   int            `json:"recent_count"` // Last 7 days
}
