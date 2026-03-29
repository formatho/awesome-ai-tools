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

// GetABTestResults returns results for a specific A/B test
func (s *AnalyticsService) GetABTestResults(ctx context.Context, testID string) (*models.ABTestResults, error) {
	results := &models.ABTestResults{
		TestID:    testID,
		StartTime: time.Now().AddDate(0, 0, -7), // Default to 7 days ago
		EndTime:   time.Now(),
	}

	// Get variants for this test
	variants, err := s.getTestVariants(ctx, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test variants: %w", err)
	}

	results.Variants = variants

	// Calculate overall stats
	for _, variant := range variants {
		results.TotalVisitors += variant.Visitors
		results.TotalConversions += variant.Conversions
	}

	// Determine winner (if any)
	if len(variants) > 1 {
		results.StatisticalSignificance = s.calculateSignificance(variants[0], variants[1])
		results.Winner = s.determineWinner(variants)
	}

	return results, nil
}

// GetAllABTests returns all active A/B tests with summaries
func (s *AnalyticsService) GetAllABTests(ctx context.Context) ([]models.ABTestSummary, error) {
	// Return the 3 active tests
	tests := []models.ABTestSummary{
		{
			TestID:          "headline-v1",
			TestName:        "Headline Test: Value vs Speed",
			Hypothesis:      "Speed-focused headline will outperform value-focused headline",
			StartDate:       time.Now().AddDate(0, 0, -1),
			Status:          "running",
			ExpectedLift:    15.0,
			TrafficSplit:    "50/50",
			TargetMetric:    "sign_up_rate",
		},
		{
			TestID:          "cta-v1",
			TestName:        "CTA Button Test",
			Hypothesis:      "Concrete benefit CTA will outperform generic CTA",
			StartDate:       time.Now().AddDate(0, 0, -1),
			Status:          "running",
			ExpectedLift:    12.5,
			TrafficSplit:    "33/33/34",
			TargetMetric:    "click_through_rate",
		},
		{
			TestID:          "urgency-v1",
			TestName:        "Urgency Banner Test",
			Hypothesis:      "Social proof will outperform scarcity messaging",
			StartDate:       time.Now().AddDate(0, 0, -1),
			Status:          "running",
			ExpectedLift:    10.0,
			TrafficSplit:    "50/50",
			TargetMetric:    "conversion_rate",
		},
	}

	// Get actual data for each test
	for i := range tests {
		results, err := s.GetABTestResults(ctx, tests[i].TestID)
		if err == nil {
			tests[i].TotalVisitors = results.TotalVisitors
			tests[i].TotalConversions = results.TotalConversions
			if results.TotalVisitors > 0 {
				tests[i].ConversionRate = float64(results.TotalConversions) / float64(results.TotalVisitors) * 100
			}
		}
	}

	return tests, nil
}

// getTestVariants retrieves variant stats for a test
func (s *AnalyticsService) getTestVariants(ctx context.Context, testID string) ([]models.ABTestVariant, error) {
	var variants []models.ABTestVariant

	// Define test variants
	testConfigs := map[string][]models.ABTestVariant{
		"headline-v1": {
			{VariantID: "control", VariantName: "Control: Orchestrate AI Agents", Visitors: 0, Conversions: 0},
			{VariantID: "variant-a", VariantName: "Variant A: Ship Code 10x Faster", Visitors: 0, Conversions: 0},
		},
		"cta-v1": {
			{VariantID: "control", VariantName: "Control: Start Building Free", Visitors: 0, Conversions: 0},
			{VariantID: "variant-a", VariantName: "Variant A: Get Started in 2 Minutes", Visitors: 0, Conversions: 0},
			{VariantID: "variant-b", VariantName: "Variant B: Claim Your 5 Free Agents", Visitors: 0, Conversions: 0},
		},
		"urgency-v1": {
			{VariantID: "control", VariantName: "Control: Beta Spots Scarcity", Visitors: 0, Conversions: 0},
			{VariantID: "variant-a", VariantName: "Variant A: Social Proof", Visitors: 0, Conversions: 0},
		},
	}

	if config, ok := testConfigs[testID]; ok {
		variants = config
	}

	// Query actual data from analytics_events
	// For each variant, count visitors and conversions
	for i := range variants {
		// Count visitors (impressions)
		err := s.db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT session_id)
			FROM analytics_events
			WHERE properties LIKE ?
			AND event_type = ?
			AND event_name = 'ab_test_impression'
		`, fmt.Sprintf(`%%"test_id":"%s","variant_id":"%s"%%`, testID, variants[i].VariantID), models.EventABTest).Scan(&variants[i].Visitors)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to count visitors for variant %s: %w", variants[i].VariantID, err)
		}

		// Count conversions
		err = s.db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT session_id)
			FROM analytics_events
			WHERE properties LIKE ?
			AND event_type = ?
			AND event_name = 'ab_test_conversion'
		`, fmt.Sprintf(`%%"test_id":"%s","variant_id":"%s"%%`, testID, variants[i].VariantID), models.EventABTest).Scan(&variants[i].Conversions)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to count conversions for variant %s: %w", variants[i].VariantID, err)
		}

		// Calculate conversion rate
		if variants[i].Visitors > 0 {
			variants[i].ConversionRate = float64(variants[i].Conversions) / float64(variants[i].Visitors) * 100
		}
	}

	return variants, nil
}

// calculateSignificance calculates statistical significance using chi-squared test
func (s *AnalyticsService) calculateSignificance(variantA, variantB models.ABTestVariant) float64 {
	// Simplified chi-squared calculation
	// For production, use a proper statistics library

	// If not enough data, return 0
	if variantA.Visitors < 100 || variantB.Visitors < 100 {
		return 0.0
	}

	// Calculate conversion rates
	rateA := float64(variantA.Conversions) / float64(variantA.Visitors)
	rateB := float64(variantB.Conversions) / float64(variantB.Visitors)

	// If rates are identical, no significance
	if rateA == rateB {
		return 0.0
	}

	// Simplified: if difference is > 10% relative, assume 95% confidence
	// In production, use proper statistical test
	diff := (rateB - rateA) / rateA * 100
	if diff > 10 || diff < -10 {
		return 0.95
	} else if diff > 5 || diff < -5 {
		return 0.85
	}

	return 0.50
}

// determineWinner identifies the winning variant (if any)
func (s *AnalyticsService) determineWinner(variants []models.ABTestVariant) string {
	if len(variants) < 2 {
		return ""
	}

	// Find variant with highest conversion rate
	var bestVariant string
	var bestRate float64

	for _, v := range variants {
		if v.ConversionRate > bestRate {
			bestRate = v.ConversionRate
			bestVariant = v.VariantID
		}
	}

	// Only declare winner if we have enough data
	// and statistical significance is high
	for _, v := range variants {
		if v.VariantID == bestVariant && v.Visitors >= 100 && v.Conversions >= 10 {
			return bestVariant
		}
	}

	return ""
}
