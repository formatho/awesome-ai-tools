// Package handlers provides HTTP handlers for the REST API.
package handlers

import (
	"strconv"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// AgentHandler handles agent-related requests.
type AgentHandler struct {
	service *services.AgentService
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler(service *services.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

// List returns all agents. Supports organization filtering via X-Organization-ID header.
func (h *AgentHandler) List(c *fiber.Ctx) error {
	var orgID *string
	orgIDHeader := c.Get("X-Organization-ID")
	if orgIDHeader != "" {
		orgID = &orgIDHeader
	}

	agents, err := h.service.List(orgID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(agents)
}

// Get returns a single agent.
func (h *AgentHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	agent, err := h.service.Get(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(agent)
}

// Create creates a new agent.
func (h *AgentHandler) Create(c *fiber.Ctx) error {
	var req models.AgentCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	agent, err := h.service.Create(&req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "VALIDATION_ERROR" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(agent)
}

// Update updates an agent.
func (h *AgentHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.AgentUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	agent, err := h.service.Update(id, &req)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(agent)
}

// Delete deletes an agent.
func (h *AgentHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(id); err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(204).Send(nil)
}

// Pause pauses an agent.
func (h *AgentHandler) Pause(c *fiber.Ctx) error {
	id := c.Params("id")
	agent, err := h.service.Pause(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(agent)
}

// Resume resumes an agent.
func (h *AgentHandler) Resume(c *fiber.Ctx) error {
	id := c.Params("id")
	agent, err := h.service.Resume(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(agent)
}

// Start starts an agent (alias for Resume).
func (h *AgentHandler) Start(c *fiber.Ctx) error {
	id := c.Params("id")
	agent, err := h.service.Start(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(agent)
}

// Stop stops an agent (alias for Pause).
func (h *AgentHandler) Stop(c *fiber.Ctx) error {
	id := c.Params("id")
	agent, err := h.service.Stop(id)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(agent)
}

// Logs returns logs for an agent.
func (h *AgentHandler) Logs(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse query parameters for filtering and pagination
	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	logs, err := h.service.GetLogs(id, limit)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(logs)
}
