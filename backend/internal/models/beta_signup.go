package models

import "time"

// BetaSignup represents a beta tester application
type BetaSignup struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role,omitempty"`
	UseCase     string     `json:"use_case,omitempty"`
	Status      string     `json:"status"` // pending, accepted, rejected
	Notes       string     `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}

// BetaSignupRequest represents the request body for beta signup
type BetaSignupRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Role    string `json:"role,omitempty"`
	UseCase string `json:"use_case,omitempty"`
}

// BetaSignupStats represents statistics about beta signups
type BetaSignupStats struct {
	TotalSignups   int            `json:"total_signups"`
	PendingCount   int            `json:"pending_count"`
	AcceptedCount  int            `json:"accepted_count"`
	RejectedCount  int            `json:"rejected_count"`
	ByRole         map[string]int `json:"by_role"`
}
