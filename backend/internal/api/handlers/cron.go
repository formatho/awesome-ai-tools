package handlers

import (
	"strconv"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// CronHandler handles cron-related requests.
type CronHandler struct {
	service *services.CronService
}

// NewCronHandler creates a new cron handler.
func NewCronHandler(service *services.CronService) *CronHandler {
	return &CronHandler{service: service}
}

// List returns all cron jobs.
func (h *CronHandler) List(c *fiber.Ctx) error {
	jobs, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(jobs)
}

// Get returns a single cron job.
func (h *CronHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.service.Get(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(job)
}

// Create creates a new cron job.
func (h *CronHandler) Create(c *fiber.Ctx) error {
	var req models.CronCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	job, err := h.service.Create(&req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "VALIDATION_ERROR" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(job)
}

// Update updates a cron job.
func (h *CronHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.CronUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	job, err := h.service.Update(id, &req)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(job)
}

// Delete deletes a cron job.
func (h *CronHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(id); err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(204).Send(nil)
}

// Pause pauses a cron job.
func (h *CronHandler) Pause(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.service.Pause(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(job)
}

// Resume resumes a paused cron job.
func (h *CronHandler) Resume(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.service.Resume(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(job)
}

// GetHistory returns the execution history for a cron job.
func (h *CronHandler) GetHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.service.GetHistory(id, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(history)
}
