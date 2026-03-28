package models

import "time"

// BetaFeedback represents feedback from beta testers
type BetaFeedback struct {
	ID           string     `json:"id"`
	UserEmail    string     `json:"user_email"`
	UserName     string     `json:"user_name"`
	Category     string     `json:"category"` // bug, feature_request, testimonial, general
	Subject      string     `json:"subject"`
	Message      string     `json:"message"`
	Rating       int        `json:"rating,omitempty"` // 1-5 stars
	Status       string     `json:"status"` // new, acknowledged, in_progress, resolved, closed
	Priority     string     `json:"priority"` // low, medium, high, critical
	Response     string     `json:"response,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty"`
	PageURL      string     `json:"page_url,omitempty"`
	Screenshot   string     `json:"screenshot,omitempty"` // Base64 encoded or URL
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// BetaFeedbackRequest represents the request body for submitting feedback
type BetaFeedbackRequest struct {
	UserEmail  string   `json:"user_email" validate:"required,email"`
	UserName   string   `json:"user_name" validate:"required"`
	Category   string   `json:"category" validate:"required,oneof=bug feature_request testimonial general"`
	Subject    string   `json:"subject" validate:"required"`
	Message    string   `json:"message" validate:"required"`
	Rating     int      `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Tags       []string `json:"tags,omitempty"`
	UserAgent  string   `json:"user_agent,omitempty"`
	PageURL    string   `json:"page_url,omitempty"`
	Screenshot string   `json:"screenshot,omitempty"`
}

// BetaFeedbackStats represents statistics about feedback
type BetaFeedbackStats struct {
	TotalFeedback   int            `json:"total_feedback"`
	ByCategory      map[string]int `json:"by_category"`
	ByStatus        map[string]int `json:"by_status"`
	AverageRating   float64        `json:"average_rating"`
	RecentFeedback  int            `json:"recent_feedback"` // Last 7 days
}
