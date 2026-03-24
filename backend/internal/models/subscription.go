// Package models defines the data structures used by the API.
package models

import (
	"time"
)

// SubscriptionTier represents the subscription tier
type SubscriptionTier string

const (
	TierFree       SubscriptionTier = "free"
	TierPro        SubscriptionTier = "pro"
	TierTeam       SubscriptionTier = "team"
	TierEnterprise SubscriptionTier = "enterprise"
)

// SubscriptionStatus represents the subscription status
type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusPastDue   SubscriptionStatus = "past_due"
	StatusCanceled  SubscriptionStatus = "canceled"
	StatusTrialing  SubscriptionStatus = "trialing"
	StatusIncomplete SubscriptionStatus = "incomplete"
)

// UsageLimit defines tier-specific limits
type UsageLimit struct {
	MaxAgents         int `json:"max_agents"`
	MaxTasksPerDay    int `json:"max_tasks_per_day"`
	MaxAgentPoolSize  int `json:"max_agent_pool_size"`
	MaxStorageMB      int `json:"max_storage_mb"`
	MaxAPICallsPerDay int `json:"max_api_calls_per_day"`
}

// GetTierLimits returns limits for a subscription tier
func GetTierLimits(tier SubscriptionTier) UsageLimit {
	switch tier {
	case TierFree:
		return UsageLimit{
			MaxAgents:         3,
			MaxTasksPerDay:    100,
			MaxAgentPoolSize:  1,
			MaxStorageMB:      100,
			MaxAPICallsPerDay: 1000,
		}
	case TierPro:
		return UsageLimit{
			MaxAgents:         -1, // Unlimited
			MaxTasksPerDay:    -1,
			MaxAgentPoolSize:  50,
			MaxStorageMB:      5000,
			MaxAPICallsPerDay: -1,
		}
	case TierTeam:
		return UsageLimit{
			MaxAgents:         -1,
			MaxTasksPerDay:    -1,
			MaxAgentPoolSize:  -1,
			MaxStorageMB:      20000,
			MaxAPICallsPerDay: -1,
		}
	case TierEnterprise:
		return UsageLimit{
			MaxAgents:         -1,
			MaxTasksPerDay:    -1,
			MaxAgentPoolSize:  -1,
			MaxStorageMB:      -1,
			MaxAPICallsPerDay: -1,
		}
	default:
		return GetTierLimits(TierFree)
	}
}

// PricingTier represents a pricing tier for checkout
type PricingTier struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	PriceMonthly  float64 `json:"price_monthly"`
	PriceYearly   float64 `json:"price_yearly"`
	StripePriceID string  `json:"stripe_price_id"`
	Description   string  `json:"description"`
	Features      []string `json:"features"`
}

// DefaultPricingTiers returns the default pricing configuration
func DefaultPricingTiers() map[string]PricingTier {
	return map[string]PricingTier{
		"pro": {
			ID:            "pro",
			Name:          "Pro",
			PriceMonthly:  49.00,
			PriceYearly:   468.00, // ~$39/month
			StripePriceID: "price_pro_monthly", // Replace with actual Stripe price IDs
			Description:   "For individual power users",
			Features: []string{
				"Unlimited agents",
				"Unlimited tasks",
				"Agent pools (up to 50)",
				"5GB storage",
				"Advanced skills library",
				"Priority support",
			},
		},
		"team": {
			ID:            "team",
			Name:          "Team",
			PriceMonthly:  149.00,
			PriceYearly:   1428.00, // ~$119/month
			StripePriceID: "price_team_monthly",
			Description:   "For teams and organizations",
			Features: []string{
				"Everything in Pro",
				"Unlimited team members",
				"Team collaboration",
				"20GB storage",
				"Admin controls",
				"SSO (coming soon)",
			},
		},
		"enterprise": {
			ID:            "enterprise",
			Name:          "Enterprise",
			PriceMonthly:  499.00,
			PriceYearly:   4788.00, // ~$399/month
			StripePriceID: "price_enterprise_monthly",
			Description:   "For large organizations",
			Features: []string{
				"Everything in Team",
				"Unlimited everything",
				"SSO/SAML",
				"Custom integrations",
				"Dedicated support",
				"SLA guarantee",
			},
		},
	}
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                   string             `json:"id"`
	UserID               string             `json:"user_id"`
	OrganizationID       string             `json:"organization_id,omitempty"`
	Tier                 SubscriptionTier   `json:"tier"`
	Status               SubscriptionStatus `json:"status"`
	StripeCustomerID     string             `json:"stripe_customer_id"`
	StripeSubscriptionID string             `json:"stripe_subscription_id"`
	StripePriceID        string             `json:"stripe_price_id"`
	CurrentPeriodStart   *time.Time         `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty"`
	TrialStart           *time.Time         `json:"trial_start,omitempty"`
	TrialEnd             *time.Time         `json:"trial_end,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// PaymentRecord represents a payment transaction
type PaymentRecord struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	SubscriptionID       string     `json:"subscription_id"`
	StripeInvoiceID      string     `json:"stripe_invoice_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id"`
	Amount               float64    `json:"amount"`
	Currency             string     `json:"currency"`
	Status               string     `json:"status"` // succeeded, failed, pending
	Description          string     `json:"description"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}
