package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v84"
)

// StripeHandler handles Stripe-related requests
type StripeHandler struct {
	authSvc      *services.AuthService
	subSvc       *services.SubscriptionService
	analyticsSvc *services.AnalyticsService
	client       *stripe.Client
	webhookSec   string
	frontendURL  string
}

// NewStripeHandler creates a new Stripe handler
func NewStripeHandler(authSvc *services.AuthService, subSvc *services.SubscriptionService, analyticsSvc ...*services.AnalyticsService) *StripeHandler {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	
	var analytics *services.AnalyticsService
	if len(analyticsSvc) > 0 {
		analytics = analyticsSvc[0]
	}

	return &StripeHandler{
		authSvc:      authSvc,
		subSvc:       subSvc,
		analyticsSvc: analytics,
		client:       stripe.NewClient(secretKey),
		webhookSec:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
		frontendURL:  getEnvOrDefault("FRONTEND_URL", "http://localhost:5173"),
	}
}

// getEnvOrDefault returns env var or default
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// CheckoutRequest represents the checkout session request
type CheckoutRequest struct {
	PriceID      string `json:"priceId"`
	Plan         string `json:"plan"`
	BillingCycle string `json:"billingCycle"`
	SuccessURL   string `json:"successUrl"`
	CancelURL    string `json:"cancelUrl"`
}

// PortalRequest represents the customer portal request
type PortalRequest struct {
	ReturnURL string `json:"returnUrl"`
}

// CreateCheckoutSession creates a Stripe checkout session
func (h *StripeHandler) CreateCheckoutSession(c *fiber.Ctx) error {
	var req CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Get or create Stripe customer
	customerID, err := h.getOrCreateCustomer(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create customer"})
	}

	// Default URLs if not provided
	successURL := req.SuccessURL
	if successURL == "" {
		successURL = h.frontendURL + "/settings/billing?success=true"
	}
	cancelURL := req.CancelURL
	if cancelURL == "" {
		cancelURL = h.frontendURL + "/pricing?canceled=true"
	}

	// Create checkout session params
	params := &stripe.CheckoutSessionCreateParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id":       userID,
			"plan":          req.Plan,
			"billing_cycle": req.BillingCycle,
		},
	}

	// Create the session using v84 client
	sess, err := h.client.V1CheckoutSessions.Create(context.TODO(), params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to create session: %v", err)})
	}

	// Track checkout start event for funnel analytics
	if h.analyticsSvc != nil {
		sessionID := c.Cookies("analytics_session_id")
		if sessionID == "" {
			sessionID = userID // Fallback to user ID if no session
		}
		h.analyticsSvc.TrackCheckoutStart(c.Context(), userID, sessionID, req.Plan, req.BillingCycle, req.PriceID, sess.ID)
	}

	return c.JSON(fiber.Map{"sessionId": sess.ID, "url": sess.URL})
}

// CreatePortalSession creates a Stripe customer portal session
func (h *StripeHandler) CreatePortalSession(c *fiber.Ctx) error {
	var req PortalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Get customer ID from subscription
	sub, err := h.subSvc.GetSubscriptionByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get subscription"})
	}

	if sub.StripeCustomerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "No Stripe customer ID found. Please start a subscription first."})
	}

	// Default return URL
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = h.frontendURL + "/settings/billing"
	}

	// Create portal session
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(sub.StripeCustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	portalSession, err := h.client.V1BillingPortalSessions.Create(context.TODO(), params)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create portal session"})
	}

	return c.JSON(fiber.Map{"url": portalSession.URL})
}

// HandleWebhook handles Stripe webhook events
func (h *StripeHandler) HandleWebhook(c *fiber.Ctx) error {
	payload := c.Body()
	signature := c.Get("Stripe-Signature")

	// Verify webhook signature and construct event
	event, err := h.client.ConstructEvent(payload, signature, h.webhookSec)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Webhook signature verification failed: %v", err)})
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(c.Context(), event)
	case "customer.subscription.created":
		h.handleSubscriptionCreated(c.Context(), event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(c.Context(), event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(c.Context(), event)
	case "invoice.payment_succeeded":
		h.handlePaymentSucceeded(c.Context(), event)
	case "invoice.payment_failed":
		h.handlePaymentFailed(c.Context(), event)
	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	return c.JSON(fiber.Map{"received": true})
}

// handleCheckoutCompleted handles checkout.session.completed event
func (h *StripeHandler) handleCheckoutCompleted(ctx context.Context, event stripe.Event) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		fmt.Printf("Error parsing checkout session: %v\n", err)
		return
	}

	userID := sess.Metadata["user_id"]
	plan := sess.Metadata["plan"]
	billingCycle := sess.Metadata["billing_cycle"]

	fmt.Printf("Checkout completed for user %s: plan=%s\n", userID, plan)

	// Track successful checkout for funnel analytics
	if h.analyticsSvc != nil {
		h.analyticsSvc.TrackCheckoutSuccess(ctx, userID, userID, plan, billingCycle, sess.ID)
	}

	// Update subscription with customer and subscription IDs
	if sess.Customer != nil && sess.Subscription != nil {
		sub, err := h.subSvc.GetSubscriptionByUserID(ctx, userID)
		if err != nil {
			fmt.Printf("Error getting subscription: %v\n", err)
			return
		}

		sub.StripeCustomerID = sess.Customer.ID
		sub.StripeSubscriptionID = sess.Subscription.ID
		sub.Tier = planToTier(plan)
		sub.Status = models.StatusActive

		if err := h.subSvc.UpdateSubscription(ctx, sub); err != nil {
			fmt.Printf("Error updating subscription: %v\n", err)
		}
	}
}

// handleSubscriptionCreated handles customer.subscription.created event
func (h *StripeHandler) handleSubscriptionCreated(ctx context.Context, event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Error parsing subscription: %v\n", err)
		return
	}

	fmt.Printf("Subscription created: %s\n", subscription.ID)

	// Get customer to find user ID
	customerID := subscription.Customer.ID
	sub, err := h.subSvc.GetSubscriptionByStripeCustomerID(ctx, customerID)
	if err != nil {
		fmt.Printf("Error finding subscription for customer %s: %v\n", customerID, err)
		return
	}

	// Update subscription details
	sub.StripeSubscriptionID = subscription.ID
	sub.Status = stripeStatusToModel(subscription.Status)
	
	// Get period info from the first subscription item
	if len(subscription.Items.Data) > 0 {
		item := subscription.Items.Data[0]
		sub.Tier = priceIDToTier(item.Price.ID)
		sub.StripePriceID = item.Price.ID
		
		if item.CurrentPeriodStart != 0 {
			t := time.Unix(item.CurrentPeriodStart, 0)
			sub.CurrentPeriodStart = &t
		}
		if item.CurrentPeriodEnd != 0 {
			t := time.Unix(item.CurrentPeriodEnd, 0)
			sub.CurrentPeriodEnd = &t
		}
	}

	if err := h.subSvc.UpdateSubscription(ctx, sub); err != nil {
		fmt.Printf("Error updating subscription: %v\n", err)
	}
}

// handleSubscriptionUpdated handles customer.subscription.updated event
func (h *StripeHandler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Error parsing subscription: %v\n", err)
		return
	}

	fmt.Printf("Subscription updated: %s, status: %s\n", subscription.ID, subscription.Status)

	customerID := subscription.Customer.ID
	sub, err := h.subSvc.GetSubscriptionByStripeCustomerID(ctx, customerID)
	if err != nil {
		fmt.Printf("Error finding subscription: %v\n", err)
		return
	}

	sub.Status = stripeStatusToModel(subscription.Status)
	sub.CancelAtPeriodEnd = subscription.CancelAtPeriodEnd

	// Get period info from the first subscription item
	if len(subscription.Items.Data) > 0 {
		item := subscription.Items.Data[0]
		sub.StripePriceID = item.Price.ID
		sub.Tier = priceIDToTier(sub.StripePriceID)
		
		if item.CurrentPeriodStart != 0 {
			t := time.Unix(item.CurrentPeriodStart, 0)
			sub.CurrentPeriodStart = &t
		}
		if item.CurrentPeriodEnd != 0 {
			t := time.Unix(item.CurrentPeriodEnd, 0)
			sub.CurrentPeriodEnd = &t
		}
	}
	
	if subscription.CanceledAt != 0 {
		t := time.Unix(subscription.CanceledAt, 0)
		sub.CanceledAt = &t
	}

	if err := h.subSvc.UpdateSubscription(ctx, sub); err != nil {
		fmt.Printf("Error updating subscription: %v\n", err)
	}
}

// handleSubscriptionDeleted handles customer.subscription.deleted event
func (h *StripeHandler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		fmt.Printf("Error parsing subscription: %v\n", err)
		return
	}

	fmt.Printf("Subscription deleted: %s\n", subscription.ID)

	// Cancel subscription in database
	if err := h.subSvc.CancelSubscription(ctx, subscription.ID); err != nil {
		fmt.Printf("Error canceling subscription: %v\n", err)
	}
}

// handlePaymentSucceeded handles invoice.payment_succeeded event
func (h *StripeHandler) handlePaymentSucceeded(ctx context.Context, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		fmt.Printf("Error parsing invoice: %v\n", err)
		return
	}

	fmt.Printf("Payment succeeded for invoice: %s, amount: %f\n", invoice.ID, float64(invoice.AmountPaid)/100)

	// Get subscription by customer ID
	sub, err := h.subSvc.GetSubscriptionByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		fmt.Printf("Error finding subscription: %v\n", err)
		return
	}

	// Record payment
	payment := &models.PaymentRecord{
		UserID:          sub.UserID,
		SubscriptionID:  sub.ID,
		StripeInvoiceID: invoice.ID,
		Amount:          float64(invoice.AmountPaid) / 100,
		Currency:        string(invoice.Currency),
		Status:          "succeeded",
		Description:     fmt.Sprintf("Subscription payment - %s", sub.Tier),
		PaidAt:          timePtr(time.Now()),
	}

	if err := h.subSvc.RecordPayment(ctx, payment); err != nil {
		fmt.Printf("Error recording payment: %v\n", err)
	}
}

// handlePaymentFailed handles invoice.payment_failed event
func (h *StripeHandler) handlePaymentFailed(ctx context.Context, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		fmt.Printf("Error parsing invoice: %v\n", err)
		return
	}

	fmt.Printf("Payment failed for invoice: %s, attempt: %d\n", invoice.ID, invoice.AttemptCount)

	// Get subscription
	sub, err := h.subSvc.GetSubscriptionByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		fmt.Printf("Error finding subscription: %v\n", err)
		return
	}

	// Record failed payment
	payment := &models.PaymentRecord{
		UserID:          sub.UserID,
		SubscriptionID:  sub.ID,
		StripeInvoiceID: invoice.ID,
		Amount:          float64(invoice.AmountDue) / 100,
		Currency:        string(invoice.Currency),
		Status:          "failed",
		Description:     fmt.Sprintf("Failed payment attempt #%d", invoice.AttemptCount),
	}

	if err := h.subSvc.RecordPayment(ctx, payment); err != nil {
		fmt.Printf("Error recording failed payment: %v\n", err)
	}

	// Update subscription status if this was the final attempt
	if invoice.AttemptCount >= 3 {
		sub.Status = models.StatusPastDue
		if err := h.subSvc.UpdateSubscription(ctx, sub); err != nil {
			fmt.Printf("Error updating subscription status: %v\n", err)
		}
	}
}

// GetSubscriptionStatus returns the current subscription status for a user
func (h *StripeHandler) GetSubscriptionStatus(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	sub, err := h.subSvc.GetSubscriptionByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get subscription"})
	}

	limits := models.GetTierLimits(sub.Tier)

	return c.JSON(fiber.Map{
		"tier":         sub.Tier,
		"status":       sub.Status,
		"limits":       limits,
		"trialEnd":     sub.TrialEnd,
		"periodEnd":    sub.CurrentPeriodEnd,
		"cancelAtEnd":  sub.CancelAtPeriodEnd,
	})
}

// GetPaymentHistory returns payment history for the user
func (h *StripeHandler) GetPaymentHistory(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	payments, err := h.subSvc.GetPaymentHistory(c.Context(), userID, 50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get payment history"})
	}

	return c.JSON(fiber.Map{"payments": payments})
}

// GetPricing returns available pricing tiers
func (h *StripeHandler) GetPricing(c *fiber.Ctx) error {
	tiers := models.DefaultPricingTiers()
	return c.JSON(fiber.Map{"tiers": tiers})
}

// getOrCreateCustomer gets an existing customer or creates a new one
func (h *StripeHandler) getOrCreateCustomer(ctx context.Context, userID string) (string, error) {
	// Check database for existing customer ID
	sub, err := h.subSvc.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if sub.StripeCustomerID != "" {
		return sub.StripeCustomerID, nil
	}

	// Create new customer in Stripe
	params := &stripe.CustomerCreateParams{
		Metadata: map[string]string{
			"user_id": userID,
		},
	}

	cust, err := h.client.V1Customers.Create(context.TODO(), params)
	if err != nil {
		return "", err
	}

	// Store customer ID in database
	if err := h.subSvc.UpsertStripeCustomer(ctx, userID, cust.ID); err != nil {
		fmt.Printf("Warning: failed to store customer ID: %v\n", err)
	}

	return cust.ID, nil
}

// Helper functions

func planToTier(plan string) models.SubscriptionTier {
	switch plan {
	case "pro":
		return models.TierPro
	case "team":
		return models.TierTeam
	case "enterprise":
		return models.TierEnterprise
	default:
		return models.TierFree
	}
}

func priceIDToTier(priceID string) models.SubscriptionTier {
	// Map Stripe price IDs to tiers
	// In production, this should come from configuration
	if priceID == "" {
		return models.TierFree
	}

	// Check price ID patterns (configure these in env vars)
	switch {
	case contains(priceID, "pro"):
		return models.TierPro
	case contains(priceID, "team"):
		return models.TierTeam
	case contains(priceID, "enterprise"):
		return models.TierEnterprise
	default:
		return models.TierPro // Default to Pro for paid plans
	}
}

func stripeStatusToModel(status stripe.SubscriptionStatus) models.SubscriptionStatus {
	switch status {
	case stripe.SubscriptionStatusActive:
		return models.StatusActive
	case stripe.SubscriptionStatusPastDue:
		return models.StatusPastDue
	case stripe.SubscriptionStatusCanceled:
		return models.StatusCanceled
	case stripe.SubscriptionStatusTrialing:
		return models.StatusTrialing
	case stripe.SubscriptionStatusIncomplete:
		return models.StatusIncomplete
	default:
		return models.StatusActive
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func timePtr(t time.Time) *time.Time {
	return &t
}
