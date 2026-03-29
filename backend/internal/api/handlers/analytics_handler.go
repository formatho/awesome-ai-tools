package handlers

import (
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	analyticsSvc *services.AnalyticsService
	subSvc       *services.SubscriptionService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsSvc *services.AnalyticsService, subSvc *services.SubscriptionService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsSvc: analyticsSvc,
		subSvc:       subSvc,
	}
}

// TrackEvent handles generic event tracking
func (h *AnalyticsHandler) TrackEvent(c *fiber.Ctx) error {
	var req models.EventTrackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.SessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session_id is required"})
	}
	if req.EventType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "event_type is required"})
	}
	if req.EventName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "event_name is required"})
	}

	// Get user ID if authenticated (optional for anonymous tracking)
	userID, _ := c.Locals("user_id").(string)

	// Extract request metadata
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	referrer := c.Get("Referer")

	// Track the event
	err := h.analyticsSvc.TrackEvent(c.Context(), &models.AnalyticsEvent{
		UserID:     userID,
		SessionID:  req.SessionID,
		EventType:  req.EventType,
		EventName:  req.EventName,
		Properties: req.Properties,
		PageURL:    req.PageURL,
		Referrer:   referrer,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track event"})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"tracked":  true,
		"event_id": req.EventName,
	})
}

// TrackPricingView handles pricing page view tracking
func (h *AnalyticsHandler) TrackPricingView(c *fiber.Ctx) error {
	var req struct {
		SessionID string `json:"session_id"`
		PageURL   string `json:"page_url"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.SessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session_id is required"})
	}

	userID, _ := c.Locals("user_id").(string)
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	referrer := c.Get("Referer")

	err := h.analyticsSvc.TrackPricingPageView(c.Context(), userID, req.SessionID, req.PageURL, referrer, userAgent, ipAddress)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track pricing view"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"tracked": "pricing_view",
	})
}

// TrackCheckoutClick handles checkout button click tracking
func (h *AnalyticsHandler) TrackCheckoutClick(c *fiber.Ctx) error {
	var req struct {
		SessionID    string `json:"session_id"`
		Plan         string `json:"plan"`
		BillingCycle string `json:"billing_cycle"`
		PageURL      string `json:"page_url"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.SessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session_id is required"})
	}

	userID, _ := c.Locals("user_id").(string)
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	err := h.analyticsSvc.TrackCheckoutClick(c.Context(), userID, req.SessionID, req.Plan, req.BillingCycle, req.PageURL, userAgent, ipAddress)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track checkout click"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"tracked": "checkout_click",
		"plan":    req.Plan,
	})
}

// TrackCheckoutStart handles checkout session start tracking
func (h *AnalyticsHandler) TrackCheckoutStart(c *fiber.Ctx) error {
	var req struct {
		SessionID      string `json:"session_id"`
		Plan           string `json:"plan"`
		BillingCycle   string `json:"billing_cycle"`
		PriceID        string `json:"price_id"`
		StripeSession  string `json:"stripe_session"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	err := h.analyticsSvc.TrackCheckoutStart(c.Context(), userID, req.SessionID, req.Plan, req.BillingCycle, req.PriceID, req.StripeSession)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track checkout start"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"tracked": "checkout_start",
	})
}

// TrackCheckoutSuccess handles successful checkout tracking
func (h *AnalyticsHandler) TrackCheckoutSuccess(c *fiber.Ctx) error {
	var req struct {
		SessionID     string `json:"session_id"`
		Plan          string `json:"plan"`
		BillingCycle  string `json:"billing_cycle"`
		StripeSession string `json:"stripe_session"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	err := h.analyticsSvc.TrackCheckoutSuccess(c.Context(), userID, req.SessionID, req.Plan, req.BillingCycle, req.StripeSession)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track checkout success"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"tracked": "checkout_success",
	})
}

// TrackCheckoutCancel handles checkout cancellation tracking
func (h *AnalyticsHandler) TrackCheckoutCancel(c *fiber.Ctx) error {
	var req struct {
		SessionID     string `json:"session_id"`
		StripeSession string `json:"stripe_session"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, _ := c.Locals("user_id").(string)

	err := h.analyticsSvc.TrackCheckoutCancel(c.Context(), userID, req.SessionID, req.StripeSession)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track checkout cancel"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"tracked": "checkout_cancel",
	})
}

// GetFunnel returns conversion funnel metrics
// Admin only endpoint
func (h *AnalyticsHandler) GetFunnel(c *fiber.Ctx) error {
	// Check if user is admin (you'll need to implement admin check)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// For now, allow any authenticated user to view funnel
	// In production, add admin role check here
	_ = userID

	// Parse query parameters
	period := c.Query("period", "week")
	var startDate, endDate time.Time
	endDate = time.Now()

	switch period {
	case "day":
		startDate = endDate.AddDate(0, 0, -1)
	case "week":
		startDate = endDate.AddDate(0, 0, -7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "quarter":
		startDate = endDate.AddDate(0, -3, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	case "all":
		startDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		startDate = endDate.AddDate(0, 0, -7)
	}

	// Custom date range
	if start := c.Query("start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = t
		}
	}
	if end := c.Query("end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = t
		}
	}

	funnel, err := h.analyticsSvc.GetConversionFunnel(c.Context(), models.FunnelQuery{
		StartDate: startDate,
		EndDate:   endDate,
		Period:    models.FunnelPeriod(period),
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get funnel data"})
	}

	return c.JSON(funnel)
}

// GetFunnelSummary returns a quick summary of funnel performance
func (h *AnalyticsHandler) GetFunnelSummary(c *fiber.Ctx) error {
	// Check authentication
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	summary, err := h.analyticsSvc.GetFunnelSummary(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get funnel summary"})
	}

	return c.JSON(summary)
}

// GetDailyStats returns daily breakdown of funnel metrics
func (h *AnalyticsHandler) GetDailyStats(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Parse date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default to last 30 days

	if start := c.Query("start"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = t
		}
	}
	if end := c.Query("end"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			endDate = t
		}
	}

	stats, err := h.analyticsSvc.GetDailyFunnelStats(c.Context(), models.FunnelQuery{
		StartDate:  startDate,
		EndDate:    endDate,
		GroupByDay: true,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get daily stats"})
	}

	return c.JSON(fiber.Map{
		"period": map[string]string{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
		"daily": stats,
	})
}

// GetSessionEvents returns all events for a session
func (h *AnalyticsHandler) GetSessionEvents(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "session_id is required"})
	}

	events, err := h.analyticsSvc.GetSessionEvents(c.Context(), sessionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get session events"})
	}

	return c.JSON(fiber.Map{
		"session_id": sessionID,
		"events":     events,
		"count":      len(events),
	})
}

// GetDashboard returns analytics dashboard data
func (h *AnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Get funnel summary
	summary, err := h.analyticsSvc.GetFunnelSummary(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get analytics summary"})
	}

	// Get last 7 days daily stats
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	dailyStats, err := h.analyticsSvc.GetDailyFunnelStats(c.Context(), models.FunnelQuery{
		StartDate:  startDate,
		EndDate:    endDate,
		GroupByDay: true,
	})
	if err != nil {
		dailyStats = []models.DailyFunnelStats{} // Don't fail on daily stats error
	}

	return c.JSON(fiber.Map{
		"summary":      summary,
		"daily_trend":  dailyStats,
		"last_updated": time.Now(),
	})
}

// GetABTestResults returns results for a specific A/B test
func (h *AnalyticsHandler) GetABTestResults(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	testID := c.Params("testId")
	if testID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "test_id is required"})
	}

	// Get A/B test results from service
	results, err := h.analyticsSvc.GetABTestResults(c.Context(), testID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get A/B test results"})
	}

	return c.JSON(results)
}

// GetAllABTests returns summary of all active A/B tests
func (h *AnalyticsHandler) GetAllABTests(c *fiber.Ctx) error {
	_, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Get all A/B test summaries
	tests, err := h.analyticsSvc.GetAllABTests(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get A/B tests"})
	}

	return c.JSON(fiber.Map{
		"tests":        tests,
		"total":        len(tests),
		"last_updated": time.Now(),
	})
}
