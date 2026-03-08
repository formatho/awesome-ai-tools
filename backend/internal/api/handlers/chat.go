// Package handlers provides HTTP handlers for the REST API.
package handlers

import (
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// ChatHandler handles chat-related requests.
type ChatHandler struct {
	chatSvc *services.ChatService
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(chatSvc *services.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

// GetHistory returns the chat history for an agent.
func (h *ChatHandler) GetHistory(c *fiber.Ctx) error {
	agentID := c.Params("id")

	messages, err := h.chatSvc.GetHistory(agentID)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(messages)
}

// SendMessage sends a message to an agent and returns the response.
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	agentID := c.Params("id")

	var req models.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	response, err := h.chatSvc.SendMessage(agentID, req.Message)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(response)
}

// ClearHistory clears all chat messages for an agent.
func (h *ChatHandler) ClearHistory(c *fiber.Ctx) error {
	agentID := c.Params("id")

	if err := h.chatSvc.ClearHistory(agentID); err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(204).Send(nil)
}
