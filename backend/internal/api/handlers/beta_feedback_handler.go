package handlers

import (
	"log"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// BetaFeedbackHandler handles beta feedback-related HTTP requests
type BetaFeedbackHandler struct {
	feedbackSvc *services.BetaFeedbackService
}

// NewBetaFeedbackHandler creates a new beta feedback handler
func NewBetaFeedbackHandler(feedbackSvc *services.BetaFeedbackService) *BetaFeedbackHandler {
	return &BetaFeedbackHandler{
		feedbackSvc: feedbackSvc,
	}
}

// Submit handles POST /api/beta-feedback
func (h *BetaFeedbackHandler) Submit(c *fiber.Ctx) error {
	var req models.BetaFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Basic validation
	if req.UserEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	if req.UserName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	if req.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Category is required",
		})
	}

	if req.Subject == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subject and message are required",
		})
	}

	// Create feedback
	feedback, err := h.feedbackSvc.Create(&req)
	if err != nil {
		log.Printf("Error creating beta feedback from %s: %v", req.UserEmail, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to submit feedback",
		})
	}

	log.Printf("Received beta feedback: %s from %s - Category: %s", feedback.Subject, feedback.UserEmail, feedback.Category)

	// TODO: Send email notification to founders
	// TODO: Create TODO item in agent-todo system

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":  true,
		"message":  "Thank you for your feedback! We'll review it shortly.",
		"feedback": feedback,
	})
}

// List handles GET /api/beta-feedback
func (h *BetaFeedbackHandler) List(c *fiber.Ctx) error {
	status := c.Query("status")
	category := c.Query("category")
	limit := c.QueryInt("limit", 50)

	if limit > 200 {
		limit = 200 // Cap at 200
	}

	feedbackList, err := h.feedbackSvc.List(status, category, limit)
	if err != nil {
		log.Printf("Error listing beta feedback: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list feedback",
		})
	}

	return c.JSON(fiber.Map{
		"feedback": feedbackList,
		"count":    len(feedbackList),
	})
}

// Get handles GET /api/beta-feedback/:id
func (h *BetaFeedbackHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	feedback, err := h.feedbackSvc.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Feedback not found",
		})
	}

	return c.JSON(feedback)
}

// UpdateStatus handles PATCH /api/beta-feedback/:id/status
func (h *BetaFeedbackHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Status   string `json:"status"`
		Response string `json:"response"`
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
		"new":          true,
		"acknowledged": true,
		"in_progress":  true,
		"resolved":     true,
		"closed":       true,
	}

	if !validStatuses[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Must be: new, acknowledged, in_progress, resolved, or closed",
		})
	}

	feedback, err := h.feedbackSvc.UpdateStatus(id, req.Status, req.Response)
	if err != nil {
		log.Printf("Error updating beta feedback status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update status",
		})
	}

	log.Printf("Updated beta feedback %s status to: %s", id, req.Status)

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Status updated successfully",
		"feedback": feedback,
	})
}

// GetStats handles GET /api/beta-feedback/stats
func (h *BetaFeedbackHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.feedbackSvc.GetStats()
	if err != nil {
		log.Printf("Error getting beta feedback stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get stats",
		})
	}

	return c.JSON(stats)
}

// Delete handles DELETE /api/beta-feedback/:id
func (h *BetaFeedbackHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.feedbackSvc.Delete(id)
	if err != nil {
		log.Printf("Error deleting beta feedback: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete feedback",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Feedback deleted successfully",
	})
}
