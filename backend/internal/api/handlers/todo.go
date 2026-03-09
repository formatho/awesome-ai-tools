package handlers

import (
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// TODOHandler handles TODO-related requests.
type TODOHandler struct {
	service *services.TODOService
}

// NewTODOHandler creates a new TODO handler.
func NewTODOHandler(service *services.TODOService) *TODOHandler {
	return &TODOHandler{service: service}
}

// List returns all TODOs. Supports organization filtering via X-Organization-ID header.
func (h *TODOHandler) List(c *fiber.Ctx) error {
	var orgID *string
	orgIDHeader := c.Get("X-Organization-ID")
	if orgIDHeader != "" {
		orgID = &orgIDHeader
	}

	todos, err := h.service.List(orgID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(todos)
}

// Get returns a single TODO.
func (h *TODOHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	todo, err := h.service.Get(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(todo)
}

// Create creates a new TODO.
func (h *TODOHandler) Create(c *fiber.Ctx) error {
	var req models.TODOCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	todo, err := h.service.Create(&req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "VALIDATION_ERROR" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(todo)
}

// Update updates a TODO.
func (h *TODOHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.TODOUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	todo, err := h.service.Update(id, &req)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(todo)
}

// Delete deletes a TODO.
func (h *TODOHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(id); err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(204).Send(nil)
}

// Start starts processing a TODO.
func (h *TODOHandler) Start(c *fiber.Ctx) error {
	id := c.Params("id")
	todo, err := h.service.Start(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(todo)
}

// Cancel cancels a TODO.
func (h *TODOHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	todo, err := h.service.Cancel(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(todo)
}
