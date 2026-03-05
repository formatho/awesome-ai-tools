package handlers

import (
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// ConfigHandler handles configuration requests.
type ConfigHandler struct {
	service *services.ConfigService
}

// NewConfigHandler creates a new config handler.
func NewConfigHandler(service *services.ConfigService) *ConfigHandler {
	return &ConfigHandler{service: service}
}

// Get returns the current configuration.
func (h *ConfigHandler) Get(c *fiber.Ctx) error {
	config, err := h.service.Get()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(config)
}

// Update updates the configuration.
func (h *ConfigHandler) Update(c *fiber.Ctx) error {
	var req models.ConfigUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	config, err := h.service.Update(&req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(config)
}

// TestLLM tests an LLM connection.
func (h *ConfigHandler) TestLLM(c *fiber.Ctx) error {
	var req models.LLMTestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// If no provider specified, get from config
	if req.Provider == "" {
		config, err := h.service.Get()
		if err == nil && config.LLMConfig != nil {
			req.Provider = config.LLMConfig.Provider
			if req.APIKey == "" {
				req.APIKey = config.LLMConfig.APIKey
			}
			if req.Model == "" {
				req.Model = config.LLMConfig.Model
			}
		}
	}

	result, err := h.service.TestLLM(&req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
