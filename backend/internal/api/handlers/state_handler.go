package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
)

// StateHandler handles requests related to agent state persistence.
type StateHandler struct {
	stateSvc *services.StateService
}

// NewStateHandler creates a new state handler instance.
func NewStateHandler(stateSvc *services.StateService) *StateHandler {
	return &StateHandler{
		stateSvc: stateSvc,
	}
}

// SaveState handles POST /api/agent-state - Save agent state to database
func (h *StateHandler) SaveState(c *fiber.Ctx) error {
	agentID := c.Params("agentID")

	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	// Decode request body (state data as JSON object)
	var stateData map[string]interface{}
	if err := c.BodyParser(&stateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Default metadata if not provided
	metadata := make(map[string]interface{})
	if m, ok := stateData["metadata"]; ok {
		if metaMap, ok := m.(map[string]interface{}); ok {
			metadata = metaMap
		}
		delete(stateData, "metadata") // Remove metadata from main data to avoid duplication
	}

	savedState, err := h.stateSvc.SaveState(context.Background(), agentID, stateData, metadata)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save state: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(savedState)
}

// GetAgentState handles GET /api/agent-state/:agentID - Retrieve current state by agent ID
func (h *StateHandler) GetAgentState(c *fiber.Ctx) error {
	agentID := c.Params("agentID")

	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	state, err := h.stateSvc.GetState(agentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Agent state not found: " + err.Error(),
		})
	}

	return c.JSON(state)
}

// GetAgentStateHistory handles GET /api/agent-state/:agentID/history - Retrieve version history
func (h *StateHandler) GetAgentStateHistory(c *fiber.Ctx) error {
	agentID := c.Params("agentID")

	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	// Parse optional limit and offset query parameters
	limit := 10 // Default limit
	offset := 0 // Default offset

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	history, err := h.stateSvc.GetStateHistory(agentID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get history: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"agent_id":       agentID,
		"history":        history,
		"total_versions": len(history),
	})
}

// UpdateAgentState handles PATCH /api/agent-state/:agentID - Update existing state
func (h *StateHandler) UpdateAgentState(c *fiber.Ctx) error {
	agentID := c.Params("agentID")

	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	var stateData map[string]interface{}
	if err := c.BodyParser(&stateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	metadata := make(map[string]interface{})
	if m, ok := stateData["metadata"]; ok {
		if metaMap, ok := m.(map[string]interface{}); ok {
			metadata = metaMap
		}
		delete(stateData, "metadata")
	}

	updatedState, err := h.stateSvc.SaveState(context.Background(), agentID, stateData, metadata)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update state: " + err.Error(),
		})
	}

	return c.JSON(updatedState)
}

// DeleteAgentState handles DELETE /api/agent-state/:agentID - Delete agent state
func (h *StateHandler) DeleteAgentState(c *fiber.Ctx) error {
	agentID := c.Params("agentID")

	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	err := h.stateSvc.DeleteState(context.Background(), agentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete state: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Agent state deleted successfully",
		"agent_id": agentID,
	})
}

// ExportAgentState handles GET /api/agent-state/:agentID/export - Export state data
func (h *StateHandler) ExportAgentState(c *fiber.Ctx) error {
	agentID := c.Params("agentID")

	if agentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Agent ID is required",
		})
	}

	state, err := h.stateSvc.GetState(agentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Agent state not found: " + err.Error(),
		})
	}

	return c.JSON(state)
}

// GetAgentStatesSummary handles GET /api/agent-states - Summary of all agent states
func (h *StateHandler) GetAgentStatesSummary(c *fiber.Ctx) error {
	states, err := h.stateSvc.GetAgentStateSummary()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get state summary: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total_states": len(states),
		"states":       states,
	})
}

// GetStateMetrics handles GET /api/agent-state/metrics - Returns state metrics
func (h *StateHandler) GetStateMetrics(c *fiber.Ctx) error {
	metrics, err := h.stateSvc.GetStateMetrics(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get state metrics: " + err.Error(),
		})
	}

	return c.JSON(metrics)
}

// Helper function to parse integers safely
func parseInt(s string) (int, error) {
	var result int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid character: %c", ch)
		}
		result = result*10 + int64(ch-'0')
	}
	return int(result), nil
}
