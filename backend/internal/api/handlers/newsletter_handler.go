package handlers

import (
	"log"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// NewsletterHandler handles newsletter-related HTTP requests
type NewsletterHandler struct {
	newsletterSvc *services.NewsletterService
}

// NewNewsletterHandler creates a new newsletter handler
func NewNewsletterHandler(newsletterSvc *services.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{
		newsletterSvc: newsletterSvc,
	}
}

// Subscribe handles POST /api/newsletter/subscribe
func (h *NewsletterHandler) Subscribe(c *fiber.Ctx) error {
	var req models.SubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Basic validation
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	if req.Source == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Source is required",
		})
	}

	// Subscribe
	subscriber, err := h.newsletterSvc.Subscribe(&req)
	if err != nil {
		log.Printf("Error subscribing email %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to subscribe",
		})
	}

	log.Printf("Successfully subscribed: %s from %s", subscriber.Email, subscriber.Source)

	// TODO: Send welcome email via Resend/ConvertKit integration

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":    true,
		"message":    "Successfully subscribed!",
		"subscriber": subscriber,
	})
}

// Unsubscribe handles POST /api/newsletter/unsubscribe
func (h *NewsletterHandler) Unsubscribe(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	err := h.newsletterSvc.Unsubscribe(req.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Successfully unsubscribed",
	})
}

// GetStats handles GET /api/newsletter/stats
func (h *NewsletterHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.newsletterSvc.GetStats()
	if err != nil {
		log.Printf("Error getting newsletter stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get stats",
		})
	}

	return c.JSON(stats)
}

// GetRecentSubscribers handles GET /api/newsletter/recent
func (h *NewsletterHandler) GetRecentSubscribers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit > 100 {
		limit = 100 // Cap at 100
	}

	subscribers, err := h.newsletterSvc.GetRecentSubscribers(limit)
	if err != nil {
		log.Printf("Error getting recent subscribers: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subscribers",
		})
	}

	return c.JSON(fiber.Map{
		"subscribers": subscribers,
		"count":       len(subscribers),
	})
}

// ExportSubscribers handles GET /api/newsletter/export
func (h *NewsletterHandler) ExportSubscribers(c *fiber.Ctx) error {
	subscribers, err := h.newsletterSvc.ExportSubscribers()
	if err != nil {
		log.Printf("Error exporting subscribers: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export subscribers",
		})
	}

	return c.JSON(fiber.Map{
		"subscribers": subscribers,
		"count":       len(subscribers),
	})
}
