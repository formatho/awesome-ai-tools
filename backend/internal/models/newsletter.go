package models

import "time"

// NewsletterSubscriber represents an email newsletter subscription
type NewsletterSubscriber struct {
	ID             string     `json:"id"`
	Email          string     `json:"email"`
	Source         string     `json:"source"` // homepage, pricing, blog
	SubscribedAt   time.Time  `json:"subscribed_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
	Metadata       string     `json:"metadata,omitempty"` // JSON string
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewsletterStats represents statistics about newsletter subscriptions
type NewsletterStats struct {
	TotalSubscribers int            `json:"total_subscribers"`
	ActiveSubscribers int           `json:"active_subscribers"`
	BySource         map[string]int `json:"by_source"`
	GrowthLast30Days int            `json:"growth_last_30_days"`
}

// SubscribeRequest represents the request body for subscribing
type SubscribeRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Source   string `json:"source" validate:"required"`
	Metadata string `json:"metadata,omitempty"` // Optional JSON string
}
