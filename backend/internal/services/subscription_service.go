// Package services provides business logic for the API.
package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// SubscriptionService handles subscription-related operations
type SubscriptionService struct {
	db *sql.DB
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(db *sql.DB) *SubscriptionService {
	return &SubscriptionService{db: db}
}

// GetSubscriptionByUserID retrieves a subscription by user ID
func (s *SubscriptionService) GetSubscriptionByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, organization_id, tier, status, stripe_customer_id, 
		       stripe_subscription_id, stripe_price_id, current_period_start, 
		       current_period_end, cancel_at_period_end, canceled_at, trial_start, 
		       trial_end, created_at, updated_at
		FROM subscriptions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	sub := &models.Subscription{}
	var orgID, canceledAt, trialStart, trialEnd, periodStart, periodEnd sql.NullString

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &orgID, &sub.Tier, &sub.Status, &sub.StripeCustomerID,
		&sub.StripeSubscriptionID, &sub.StripePriceID, &periodStart, &periodEnd,
		&sub.CancelAtPeriodEnd, &canceledAt, &trialStart, &trialEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return free tier subscription for users without a subscription
		return s.createFreeSubscription(userID, ""), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	sub.OrganizationID = orgID.String
	if canceledAt.Valid {
		t, _ := time.Parse(time.RFC3339, canceledAt.String)
		sub.CanceledAt = &t
	}
	if trialStart.Valid {
		t, _ := time.Parse(time.RFC3339, trialStart.String)
		sub.TrialStart = &t
	}
	if trialEnd.Valid {
		t, _ := time.Parse(time.RFC3339, trialEnd.String)
		sub.TrialEnd = &t
	}
	if periodStart.Valid {
		t, _ := time.Parse(time.RFC3339, periodStart.String)
		sub.CurrentPeriodStart = &t
	}
	if periodEnd.Valid {
		t, _ := time.Parse(time.RFC3339, periodEnd.String)
		sub.CurrentPeriodEnd = &t
	}

	return sub, nil
}

// GetSubscriptionByStripeCustomerID retrieves a subscription by Stripe customer ID
func (s *SubscriptionService) GetSubscriptionByStripeCustomerID(ctx context.Context, customerID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, organization_id, tier, status, stripe_customer_id, 
		       stripe_subscription_id, stripe_price_id, current_period_start, 
		       current_period_end, cancel_at_period_end, canceled_at, trial_start, 
		       trial_end, created_at, updated_at
		FROM subscriptions
		WHERE stripe_customer_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	sub := &models.Subscription{}
	var orgID, canceledAt, trialStart, trialEnd, periodStart, periodEnd sql.NullString

	err := s.db.QueryRowContext(ctx, query, customerID).Scan(
		&sub.ID, &sub.UserID, &orgID, &sub.Tier, &sub.Status, &sub.StripeCustomerID,
		&sub.StripeSubscriptionID, &sub.StripePriceID, &periodStart, &periodEnd,
		&sub.CancelAtPeriodEnd, &canceledAt, &trialStart, &trialEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found for customer %s", customerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	sub.OrganizationID = orgID.String
	// Parse nullable time fields...
	if canceledAt.Valid {
		t, _ := time.Parse(time.RFC3339, canceledAt.String)
		sub.CanceledAt = &t
	}
	if trialStart.Valid {
		t, _ := time.Parse(time.RFC3339, trialStart.String)
		sub.TrialStart = &t
	}
	if trialEnd.Valid {
		t, _ := time.Parse(time.RFC3339, trialEnd.String)
		sub.TrialEnd = &t
	}
	if periodStart.Valid {
		t, _ := time.Parse(time.RFC3339, periodStart.String)
		sub.CurrentPeriodStart = &t
	}
	if periodEnd.Valid {
		t, _ := time.Parse(time.RFC3339, periodEnd.String)
		sub.CurrentPeriodEnd = &t
	}

	return sub, nil
}

// CreateSubscription creates a new subscription
func (s *SubscriptionService) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	now := time.Now()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	query := `
		INSERT INTO subscriptions (
			id, user_id, organization_id, tier, status, stripe_customer_id,
			stripe_subscription_id, stripe_price_id, current_period_start,
			current_period_end, cancel_at_period_end, canceled_at, trial_start,
			trial_end, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		sub.ID, sub.UserID, sub.OrganizationID, sub.Tier, sub.Status, sub.StripeCustomerID,
		sub.StripeSubscriptionID, sub.StripePriceID, formatTime(sub.CurrentPeriodStart),
		formatTime(sub.CurrentPeriodEnd), sub.CancelAtPeriodEnd, formatTime(sub.CanceledAt),
		formatTime(sub.TrialStart), formatTime(sub.TrialEnd),
		sub.CreatedAt, sub.UpdatedAt,
	)

	return err
}

// UpdateSubscription updates an existing subscription
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	sub.UpdatedAt = time.Now()

	query := `
		UPDATE subscriptions SET
			tier = ?, status = ?, stripe_customer_id = ?, stripe_subscription_id = ?,
			stripe_price_id = ?, current_period_start = ?, current_period_end = ?,
			cancel_at_period_end = ?, canceled_at = ?, trial_start = ?, trial_end = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
		sub.Tier, sub.Status, sub.StripeCustomerID, sub.StripeSubscriptionID,
		sub.StripePriceID, formatTime(sub.CurrentPeriodStart), formatTime(sub.CurrentPeriodEnd),
		sub.CancelAtPeriodEnd, formatTime(sub.CanceledAt), formatTime(sub.TrialStart),
		formatTime(sub.TrialEnd), sub.UpdatedAt, sub.ID,
	)

	return err
}

// UpsertStripeCustomer creates or updates the Stripe customer ID for a user
func (s *SubscriptionService) UpsertStripeCustomer(ctx context.Context, userID, customerID string) error {
	// Check if subscription exists
	existing, err := s.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if existing.StripeCustomerID == "" {
		// No existing subscription record, create one
		return s.CreateSubscription(ctx, &models.Subscription{
			UserID:           userID,
			Tier:             models.TierFree,
			Status:           models.StatusActive,
			StripeCustomerID: customerID,
		})
	}

	// Update existing record with new customer ID
	existing.StripeCustomerID = customerID
	return s.UpdateSubscription(ctx, existing)
}

// UpdateSubscriptionFromStripe updates subscription details from Stripe webhook data
func (s *SubscriptionService) UpdateSubscriptionFromStripe(
	ctx context.Context,
	customerID string,
	subscriptionID string,
	priceID string,
	status models.SubscriptionStatus,
	tier models.SubscriptionTier,
	currentPeriodStart, currentPeriodEnd *time.Time,
	cancelAtPeriodEnd bool,
) error {
	sub, err := s.GetSubscriptionByStripeCustomerID(ctx, customerID)
	if err != nil {
		return err
	}

	sub.StripeSubscriptionID = subscriptionID
	sub.StripePriceID = priceID
	sub.Status = status
	sub.Tier = tier
	sub.CurrentPeriodStart = currentPeriodStart
	sub.CurrentPeriodEnd = currentPeriodEnd
	sub.CancelAtPeriodEnd = cancelAtPeriodEnd

	return s.UpdateSubscription(ctx, sub)
}

// CancelSubscription marks a subscription as canceled
func (s *SubscriptionService) CancelSubscription(ctx context.Context, subscriptionID string) error {
	query := `
		UPDATE subscriptions SET
			status = ?, canceled_at = ?, updated_at = ?
		WHERE stripe_subscription_id = ?
	`

	_, err := s.db.ExecContext(ctx, query,
		models.StatusCanceled, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339),
		subscriptionID,
	)

	return err
}

// RecordPayment creates a payment record
func (s *SubscriptionService) RecordPayment(ctx context.Context, payment *models.PaymentRecord) error {
	if payment.ID == "" {
		payment.ID = uuid.New().String()
	}
	payment.CreatedAt = time.Now()

	query := `
		INSERT INTO payments (
			id, user_id, subscription_id, stripe_invoice_id, stripe_payment_intent_id,
			amount, currency, status, description, paid_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var paidAt interface{}
	if payment.PaidAt != nil {
		paidAt = payment.PaidAt.Format(time.RFC3339)
	}

	_, err := s.db.ExecContext(ctx, query,
		payment.ID, payment.UserID, payment.SubscriptionID, payment.StripeInvoiceID,
		payment.StripePaymentIntentID, payment.Amount, payment.Currency, payment.Status,
		payment.Description, paidAt, payment.CreatedAt,
	)

	return err
}

// GetPaymentHistory retrieves payment history for a user
func (s *SubscriptionService) GetPaymentHistory(ctx context.Context, userID string, limit int) ([]models.PaymentRecord, error) {
	query := `
		SELECT id, user_id, subscription_id, stripe_invoice_id, stripe_payment_intent_id,
		       amount, currency, status, description, paid_at, created_at
		FROM payments
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment history: %w", err)
	}
	defer rows.Close()

	var payments []models.PaymentRecord
	for rows.Next() {
		var p models.PaymentRecord
		var paidAt sql.NullString
		err := rows.Scan(
			&p.ID, &p.UserID, &p.SubscriptionID, &p.StripeInvoiceID, &p.StripePaymentIntentID,
			&p.Amount, &p.Currency, &p.Status, &p.Description, &paidAt, &p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		if paidAt.Valid {
			t, _ := time.Parse(time.RFC3339, paidAt.String)
			p.PaidAt = &t
		}
		payments = append(payments, p)
	}

	return payments, nil
}

// GetTier returns the subscription tier for a user
func (s *SubscriptionService) GetTier(ctx context.Context, userID string) models.SubscriptionTier {
	sub, err := s.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		return models.TierFree
	}
	return sub.Tier
}

// HasFeature checks if a user has access to a feature based on their tier
func (s *SubscriptionService) HasFeature(ctx context.Context, userID string, minTier models.SubscriptionTier) bool {
	tier := s.GetTier(ctx, userID)
	return isTierAtLeast(tier, minTier)
}

// createFreeSubscription creates a default free tier subscription
func (s *SubscriptionService) createFreeSubscription(userID, orgID string) *models.Subscription {
	return &models.Subscription{
		ID:             uuid.New().String(),
		UserID:         userID,
		OrganizationID: orgID,
		Tier:           models.TierFree,
		Status:         models.StatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// isTierAtLeast checks if a tier meets the minimum required tier
func isTierAtLeast(current, minimum models.SubscriptionTier) bool {
	tierOrder := map[models.SubscriptionTier]int{
		models.TierFree:       0,
		models.TierPro:        1,
		models.TierTeam:       2,
		models.TierEnterprise: 3,
	}
	return tierOrder[current] >= tierOrder[minimum]
}

// formatTime safely formats a time pointer for database storage
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
