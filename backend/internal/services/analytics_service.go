package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// AnalyticsService handles analytics event tracking and funnel analysis
type AnalyticsService struct {
	db *sql.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// TrackEvent records an analytics event
func (s *AnalyticsService) TrackEvent(ctx context.Context, event *models.AnalyticsEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	event.CreatedAt = time.Now()

	// Serialize properties to JSON
	var propsJSON []byte
	var err error
	if event.Properties != nil {
		propsJSON, err = json.Marshal(event.Properties)
		if err != nil {
			return fmt.Errorf("failed to marshal properties: %w", err)
		}
	}

	query := `
		INSERT INTO analytics_events (
			id, user_id, session_id, event_type, event_name, properties,
			page_url, referrer, user_agent, ip_address, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		event.ID, event.UserID, event.SessionID, event.EventType, event.EventName,
		propsJSON, event.PageURL, event.Referrer, event.UserAgent, event.IPAddress,
		event.CreatedAt,
	)

	return err
}

// TrackFunnelEvent is a convenience method for tracking funnel-specific events
func (s *AnalyticsService) TrackFunnelEvent(
	ctx context.Context,
	userID, sessionID string,
	eventType models.AnalyticsEventType,
	eventName string,
	properties map[string]interface{},
	pageURL, referrer, userAgent, ipAddress string,
) error {
	return s.TrackEvent(ctx, &models.AnalyticsEvent{
		UserID:     userID,
		SessionID:  sessionID,
		EventType:  eventType,
		EventName:  eventName,
		Properties: properties,
		PageURL:    pageURL,
		Referrer:   referrer,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	})
}

// TrackPricingPageView records a pricing page view
func (s *AnalyticsService) TrackPricingPageView(ctx context.Context, userID, sessionID, pageURL, referrer, userAgent, ipAddress string) error {
	return s.TrackFunnelEvent(ctx, userID, sessionID, models.EventPageView, "pricing_view", map[string]interface{}{
		"page": "pricing",
	}, pageURL, referrer, userAgent, ipAddress)
}

// TrackCheckoutClick records a checkout button click
func (s *AnalyticsService) TrackCheckoutClick(ctx context.Context, userID, sessionID, plan, billingCycle string, pageURL, userAgent, ipAddress string) error {
	return s.TrackFunnelEvent(ctx, userID, sessionID, models.EventButtonClick, "checkout_click", map[string]interface{}{
		"plan":          plan,
		"billing_cycle": billingCycle,
	}, pageURL, "", userAgent, ipAddress)
}

// TrackCheckoutStart records a checkout session creation
func (s *AnalyticsService) TrackCheckoutStart(ctx context.Context, userID, sessionID, plan, billingCycle, priceID, stripeSessionID string) error {
	return s.TrackFunnelEvent(ctx, userID, sessionID, models.EventCheckoutStart, "checkout_start", map[string]interface{}{
		"plan":           plan,
		"billing_cycle":  billingCycle,
		"price_id":       priceID,
		"stripe_session": stripeSessionID,
	}, "", "", "", "")
}

// TrackCheckoutSuccess records a successful checkout
func (s *AnalyticsService) TrackCheckoutSuccess(ctx context.Context, userID, sessionID, plan, billingCycle, stripeSessionID string) error {
	return s.TrackFunnelEvent(ctx, userID, sessionID, models.EventCheckoutSuccess, "checkout_success", map[string]interface{}{
		"plan":           plan,
		"billing_cycle":  billingCycle,
		"stripe_session": stripeSessionID,
	}, "", "", "", "")
}

// TrackCheckoutCancel records a canceled checkout
func (s *AnalyticsService) TrackCheckoutCancel(ctx context.Context, userID, sessionID, stripeSessionID string) error {
	return s.TrackFunnelEvent(ctx, userID, sessionID, models.EventCheckoutCancel, "checkout_cancel", map[string]interface{}{
		"stripe_session": stripeSessionID,
	}, "", "", "", "")
}

// GetConversionFunnel returns funnel metrics for a time period
func (s *AnalyticsService) GetConversionFunnel(ctx context.Context, query models.FunnelQuery) (*models.ConversionFunnel, error) {
	funnel := &models.ConversionFunnel{
		Period:    query.Period,
		StartDate: query.StartDate,
		EndDate:   query.EndDate,
	}

	// Count pricing page views
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM analytics_events 
		WHERE event_type = ? AND event_name = 'pricing_view'
		AND created_at >= ? AND created_at <= ?
	`, models.EventPageView, query.StartDate, query.EndDate).Scan(&funnel.PricingViews)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count pricing views: %w", err)
	}

	// Count checkout clicks
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM analytics_events 
		WHERE event_type = ? AND event_name = 'checkout_click'
		AND created_at >= ? AND created_at <= ?
	`, models.EventButtonClick, query.StartDate, query.EndDate).Scan(&funnel.CheckoutClicks)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count checkout clicks: %w", err)
	}

	// Count checkout starts
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM analytics_events 
		WHERE event_type = ? AND event_name = 'checkout_start'
		AND created_at >= ? AND created_at <= ?
	`, models.EventCheckoutStart, query.StartDate, query.EndDate).Scan(&funnel.CheckoutStarts)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count checkout starts: %w", err)
	}

	// Count checkout successes
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM analytics_events 
		WHERE event_type = ? AND event_name = 'checkout_success'
		AND created_at >= ? AND created_at <= ?
	`, models.EventCheckoutSuccess, query.StartDate, query.EndDate).Scan(&funnel.CheckoutSuccess)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count checkout successes: %w", err)
	}

	// Count checkout cancels
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM analytics_events 
		WHERE event_type = ? AND event_name = 'checkout_cancel'
		AND created_at >= ? AND created_at <= ?
	`, models.EventCheckoutCancel, query.StartDate, query.EndDate).Scan(&funnel.CheckoutCancels)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count checkout cancels: %w", err)
	}

	// Calculate conversion rates
	if funnel.PricingViews > 0 {
		funnel.PricingToCheckoutRate = float64(funnel.CheckoutClicks) / float64(funnel.PricingViews) * 100
		funnel.OverallConversionRate = float64(funnel.CheckoutSuccess) / float64(funnel.PricingViews) * 100
	}
	if funnel.CheckoutStarts > 0 {
		funnel.CheckoutToSuccessRate = float64(funnel.CheckoutSuccess) / float64(funnel.CheckoutStarts) * 100
		funnel.CheckoutAbandonRate = float64(funnel.CheckoutCancels) / float64(funnel.CheckoutStarts) * 100
	}

	// Get unique user counts
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT session_id) FROM analytics_events 
		WHERE created_at >= ? AND created_at <= ?
	`, query.StartDate, query.EndDate).Scan(&funnel.UniqueVisitors)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count unique visitors: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT session_id) FROM analytics_events 
		WHERE event_name = 'pricing_view'
		AND created_at >= ? AND created_at <= ?
	`, query.StartDate, query.EndDate).Scan(&funnel.UniquePricingViewers)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count unique pricing viewers: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT session_id) FROM analytics_events 
		WHERE event_name IN ('checkout_click', 'checkout_start')
		AND created_at >= ? AND created_at <= ?
	`, query.StartDate, query.EndDate).Scan(&funnel.UniqueCheckoutUsers)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count unique checkout users: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT session_id) FROM analytics_events 
		WHERE event_name = 'checkout_success'
		AND created_at >= ? AND created_at <= ?
	`, query.StartDate, query.EndDate).Scan(&funnel.UniqueConvertedUsers)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to count unique converted users: %w", err)
	}

	return funnel, nil
}

// GetDailyFunnelStats returns daily breakdown of funnel metrics
func (s *AnalyticsService) GetDailyFunnelStats(ctx context.Context, query models.FunnelQuery) ([]models.DailyFunnelStats, error) {
	queryStr := `
		SELECT 
			DATE(created_at) as day,
			SUM(CASE WHEN event_name = 'pricing_view' THEN 1 ELSE 0 END) as pricing_views,
			SUM(CASE WHEN event_name = 'checkout_click' THEN 1 ELSE 0 END) as checkout_clicks,
			SUM(CASE WHEN event_name = 'checkout_start' THEN 1 ELSE 0 END) as checkout_starts,
			SUM(CASE WHEN event_name = 'checkout_success' THEN 1 ELSE 0 END) as checkout_success
		FROM analytics_events
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY DATE(created_at)
		ORDER BY day ASC
	`

	rows, err := s.db.QueryContext(ctx, queryStr, query.StartDate, query.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily stats: %w", err)
	}
	defer rows.Close()

	var stats []models.DailyFunnelStats
	for rows.Next() {
		var day models.DailyFunnelStats
		var dateStr string
		err := rows.Scan(&dateStr, &day.PricingViews, &day.CheckoutClicks, &day.CheckoutStarts, &day.CheckoutSuccess)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stats: %w", err)
		}

		day.Date, _ = time.Parse("2006-01-02", dateStr)
		
		// Calculate daily conversion rate
		if day.PricingViews > 0 {
			day.ConversionRate = float64(day.CheckoutSuccess) / float64(day.PricingViews) * 100
		}

		stats = append(stats, day)
	}

	return stats, nil
}

// GetEvents retrieves raw events for analysis
func (s *AnalyticsService) GetEvents(ctx context.Context, eventType models.AnalyticsEventType, limit int, offset int) ([]models.AnalyticsEvent, error) {
	query := `
		SELECT id, user_id, session_id, event_type, event_name, properties,
		       page_url, referrer, user_agent, ip_address, created_at
		FROM analytics_events
		WHERE event_type = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []models.AnalyticsEvent
	for rows.Next() {
		var event models.AnalyticsEvent
		var propsJSON []byte
		var userID, pageURL, referrer, userAgent, ipAddress sql.NullString

		err := rows.Scan(
			&event.ID, &userID, &event.SessionID, &event.EventType, &event.EventName,
			&propsJSON, &pageURL, &referrer, &userAgent, &ipAddress, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		event.UserID = userID.String
		event.PageURL = pageURL.String
		event.Referrer = referrer.String
		event.UserAgent = userAgent.String
		event.IPAddress = ipAddress.String

		if len(propsJSON) > 0 {
			json.Unmarshal(propsJSON, &event.Properties)
		}

		events = append(events, event)
	}

	return events, nil
}

// GetSessionEvents retrieves all events for a specific session
func (s *AnalyticsService) GetSessionEvents(ctx context.Context, sessionID string) ([]models.AnalyticsEvent, error) {
	query := `
		SELECT id, user_id, session_id, event_type, event_name, properties,
		       page_url, referrer, user_agent, ip_address, created_at
		FROM analytics_events
		WHERE session_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query session events: %w", err)
	}
	defer rows.Close()

	var events []models.AnalyticsEvent
	for rows.Next() {
		var event models.AnalyticsEvent
		var propsJSON []byte
		var userID, pageURL, referrer, userAgent, ipAddress sql.NullString

		err := rows.Scan(
			&event.ID, &userID, &event.SessionID, &event.EventType, &event.EventName,
			&propsJSON, &pageURL, &referrer, &userAgent, &ipAddress, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		event.UserID = userID.String
		event.PageURL = pageURL.String
		event.Referrer = referrer.String
		event.UserAgent = userAgent.String
		event.IPAddress = ipAddress.String

		if len(propsJSON) > 0 {
			json.Unmarshal(propsJSON, &event.Properties)
		}

		events = append(events, event)
	}

	return events, nil
}

// DeleteOldEvents removes events older than the retention period
func (s *AnalyticsService) DeleteOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM analytics_events WHERE created_at < ?
	`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old events: %w", err)
	}

	return result.RowsAffected()
}

// GetFunnelSummary returns a quick summary of funnel performance
func (s *AnalyticsService) GetFunnelSummary(ctx context.Context) (map[string]interface{}, error) {
	now := time.Now()
	dayStart := now.AddDate(0, 0, -1)
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	dayFunnel, err := s.GetConversionFunnel(ctx, models.FunnelQuery{
		StartDate: dayStart,
		EndDate:   now,
		Period:    models.PeriodDay,
	})
	if err != nil {
		return nil, err
	}

	weekFunnel, err := s.GetConversionFunnel(ctx, models.FunnelQuery{
		StartDate: weekStart,
		EndDate:   now,
		Period:    models.PeriodWeek,
	})
	if err != nil {
		return nil, err
	}

	monthFunnel, err := s.GetConversionFunnel(ctx, models.FunnelQuery{
		StartDate: monthStart,
		EndDate:   now,
		Period:    models.PeriodMonth,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"last_24h": map[string]interface{}{
			"pricing_views":     dayFunnel.PricingViews,
			"checkout_clicks":   dayFunnel.CheckoutClicks,
			"checkout_success":  dayFunnel.CheckoutSuccess,
			"conversion_rate":   dayFunnel.OverallConversionRate,
		},
		"last_7d": map[string]interface{}{
			"pricing_views":     weekFunnel.PricingViews,
			"checkout_clicks":   weekFunnel.CheckoutClicks,
			"checkout_success":  weekFunnel.CheckoutSuccess,
			"conversion_rate":   weekFunnel.OverallConversionRate,
		},
		"last_30d": map[string]interface{}{
			"pricing_views":     monthFunnel.PricingViews,
			"checkout_clicks":   monthFunnel.CheckoutClicks,
			"checkout_success":  monthFunnel.CheckoutSuccess,
			"conversion_rate":   monthFunnel.OverallConversionRate,
		},
	}, nil
}
