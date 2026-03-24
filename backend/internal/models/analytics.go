package models

import "time"

// AnalyticsEvent represents a tracked analytics event
type AnalyticsEvent struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id"`
	EventType   AnalyticsEventType     `json:"event_type"`
	EventName   string                 `json:"event_name"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	PageURL     string                 `json:"page_url,omitempty"`
	Referrer    string                 `json:"referrer,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AnalyticsEventType defines the type of analytics event
type AnalyticsEventType string

const (
	// Funnel events
	EventPageView        AnalyticsEventType = "page_view"
	EventButtonClick     AnalyticsEventType = "button_click"
	EventFormSubmit      AnalyticsEventType = "form_submit"
	EventCheckoutStart   AnalyticsEventType = "checkout_start"
	EventCheckoutSuccess AnalyticsEventType = "checkout_success"
	EventCheckoutCancel  AnalyticsEventType = "checkout_cancel"

	// User journey events
	EventSignup          AnalyticsEventType = "signup"
	EventLogin           AnalyticsEventType = "login"
	EventTrialStart      AnalyticsEventType = "trial_start"
	EventTrialEnd        AnalyticsEventType = "trial_end"
	EventUpgrade         AnalyticsEventType = "upgrade"
	EventDowngrade       AnalyticsEventType = "downgrade"
	EventChurn           AnalyticsEventType = "churn"
)

// FunnelStage represents a stage in the conversion funnel
type FunnelStage string

const (
	FunnelStageAwareness    FunnelStage = "awareness"     // Landing/home page
	FunnelStageInterest     FunnelStage = "interest"      // Features page
	FunnelStageConsideration FunnelStage = "consideration" // Pricing page
	FunnelStageIntent       FunnelStage = "intent"        // Checkout started
	FunnelStageConversion   FunnelStage = "conversion"    // Payment completed
)

// ConversionFunnel represents the full conversion funnel metrics
type ConversionFunnel struct {
	Period           FunnelPeriod `json:"period"`
	StartDate        time.Time    `json:"start_date"`
	EndDate          time.Time    `json:"end_date"`
	
	// Stage counts
	PageViews        int64 `json:"page_views"`         // Total page views
	PricingViews     int64 `json:"pricing_views"`      // Pricing page views
	CheckoutClicks   int64 `json:"checkout_clicks"`    // Checkout button clicks
	CheckoutStarts   int64 `json:"checkout_starts"`    // Checkout sessions created
	CheckoutSuccess  int64 `json:"checkout_success"`   // Completed checkouts
	CheckoutCancels  int64 `json:"checkout_cancels"`   // Canceled checkouts
	
	// Conversion rates
	PricingToCheckoutRate   float64 `json:"pricing_to_checkout_rate"`   // pricing -> checkout click
	CheckoutToSuccessRate   float64 `json:"checkout_to_success_rate"`   // checkout start -> success
	OverallConversionRate   float64 `json:"overall_conversion_rate"`    // pricing -> success
	CheckoutAbandonRate     float64 `json:"checkout_abandon_rate"`      // checkout -> cancel
	
	// Unique users
	UniqueVisitors       int64 `json:"unique_visitors"`
	UniquePricingViewers int64 `json:"unique_pricing_viewers"`
	UniqueCheckoutUsers  int64 `json:"unique_checkout_users"`
	UniqueConvertedUsers int64 `json:"unique_converted_users"`
}

// FunnelPeriod represents the time period for funnel analysis
type FunnelPeriod string

const (
	PeriodDay     FunnelPeriod = "day"
	PeriodWeek    FunnelPeriod = "week"
	PeriodMonth   FunnelPeriod = "month"
	PeriodQuarter FunnelPeriod = "quarter"
	PeriodYear    FunnelPeriod = "year"
	PeriodAll     FunnelPeriod = "all"
)

// FunnelQuery represents parameters for querying funnel data
type FunnelQuery struct {
	StartDate  time.Time    `json:"start_date"`
	EndDate    time.Time    `json:"end_date"`
	Period     FunnelPeriod `json:"period"`
	UserID     string       `json:"user_id,omitempty"`
	SessionID  string       `json:"session_id,omitempty"`
	GroupByDay bool         `json:"group_by_day"`
}

// DailyFunnelStats represents funnel stats for a single day
type DailyFunnelStats struct {
	Date             time.Time `json:"date"`
	PricingViews     int64     `json:"pricing_views"`
	CheckoutClicks   int64     `json:"checkout_clicks"`
	CheckoutStarts   int64     `json:"checkout_starts"`
	CheckoutSuccess  int64     `json:"checkout_success"`
	ConversionRate   float64   `json:"conversion_rate"`
}

// EventTrackRequest represents an event tracking request
type EventTrackRequest struct {
	SessionID  string                 `json:"session_id"`
	EventType  AnalyticsEventType     `json:"event_type"`
	EventName  string                 `json:"event_name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	PageURL    string                 `json:"page_url,omitempty"`
	Referrer   string                 `json:"referrer,omitempty"`
}

// CheckoutFunnelEvent is a specialized event for checkout tracking
type CheckoutFunnelEvent struct {
	SessionID     string `json:"session_id"`
	UserID        string `json:"user_id,omitempty"`
	EventType     string `json:"event_type"`
	Plan          string `json:"plan,omitempty"`
	BillingCycle  string `json:"billing_cycle,omitempty"`
	PriceID       string `json:"price_id,omitempty"`
	SessionURL    string `json:"session_url,omitempty"`
	StripeSession string `json:"stripe_session,omitempty"`
}

// GetFunnelStage determines the funnel stage from an event type
func GetFunnelStage(eventType AnalyticsEventType) FunnelStage {
	switch eventType {
	case EventPageView:
		return FunnelStageAwareness
	case EventButtonClick:
		return FunnelStageConsideration
	case EventCheckoutStart:
		return FunnelStageIntent
	case EventCheckoutSuccess:
		return FunnelStageConversion
	case EventCheckoutCancel:
		return FunnelStageIntent
	default:
		return FunnelStageAwareness
	}
}
