package handlers

import (
	"log"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// BetaSignupHandler handles beta signup-related HTTP requests
type BetaSignupHandler struct {
	betaSvc *services.BetaSignupService
}

// NewBetaSignupHandler creates a new beta signup handler
func NewBetaSignupHandler(betaSvc *services.BetaSignupService) *BetaSignupHandler {
	return &BetaSignupHandler{
		betaSvc: betaSvc,
	}
}

// Signup handles POST /api/beta-signup
func (h *BetaSignupHandler) Signup(c *fiber.Ctx) error {
	var req models.BetaSignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Basic validation
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Check if email already exists
	existing, err := h.betaSvc.GetByEmail(req.Email)
	if err != nil {
		log.Printf("Error checking existing email %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing signup",
		})
	}

	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "Email already registered",
			"message": "This email is already on our beta list. We'll be in touch soon!",
		})
	}

	// Create signup
	signup, err := h.betaSvc.Create(&req)
	if err != nil {
		log.Printf("Error creating beta signup for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create signup",
		})
	}

	log.Printf("Successfully created beta signup: %s (%s) - Role: %s", signup.Name, signup.Email, signup.Role)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Application received! We'll review it and get back to you within 24 hours.",
		"signup": fiber.Map{
			"id":     signup.ID,
			"name":   signup.Name,
			"email":  signup.Email,
			"status": signup.Status,
		},
	})
}

// List handles GET /api/beta-signups
func (h *BetaSignupHandler) List(c *fiber.Ctx) error {
	status := c.Query("status")

	signups, err := h.betaSvc.List(status)
	if err != nil {
		log.Printf("Error listing beta signups: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list signups",
		})
	}

	return c.JSON(fiber.Map{
		"signups": signups,
		"count":   len(signups),
	})
}

// Get handles GET /api/beta-signups/:id
func (h *BetaSignupHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	signup, err := h.betaSvc.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Signup not found",
		})
	}

	return c.JSON(signup)
}

// UpdateStatus handles PATCH /api/beta-signups/:id/status
func (h *BetaSignupHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":  true,
		"accepted": true,
		"rejected": true,
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Must be: pending, accepted, or rejected",
		})
	}

	signup, err := h.betaSvc.UpdateStatus(id, req.Status, req.Notes)
	if err != nil {
		log.Printf("Error updating beta signup status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update status",
		})
	}

	log.Printf("Updated beta signup %s status to: %s", id, req.Status)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Status updated successfully",
		"signup":  signup,
	})
}

// GetStats handles GET /api/beta-signups/stats
func (h *BetaSignupHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.betaSvc.GetStats()
	if err != nil {
		log.Printf("Error getting beta signup stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get stats",
		})
	}

	return c.JSON(stats)
}

// Delete handles DELETE /api/beta-signups/:id
func (h *BetaSignupHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.betaSvc.Delete(id)
	if err != nil {
		log.Printf("Error deleting beta signup: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete signup",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Signup deleted successfully",
	})
}
