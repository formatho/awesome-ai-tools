// Package services integration tests for agent API endpoints.
package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/formatho/agent-orchestrator/backend/internal/api/handlers"
	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTestApp creates a full test application with all routes
func setupIntegrationTestApp(t *testing.T) (*fiber.App, *sql.DB, *AgentService) {
	// Create in-memory database
	db, err := store.InitDB(":memory:")
	require.NoError(t, err, "Failed to create test database")

	// Run migrations
	err = store.RunMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	// Create WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create services
	agentSvc := NewAgentService(db, hub)

	// Create handlers
	agentH := handlers.NewAgentHandler(agentSvc)

	// Create Fiber app with middleware
	app := fiber.New(fiber.Config{
		AppName: "Agent Orchestrator Integration Test API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Setup routes (matching production)
	api := app.Group("/api")
	agents := api.Group("/agents")
	agents.Get("/", agentH.List)
	agents.Post("/", agentH.Create)
	agents.Get("/:id", agentH.Get)
	agents.Put("/:id", agentH.Update)
	agents.Delete("/:id", agentH.Delete)
	agents.Post("/:id/start", agentH.Start)
	agents.Post("/:id/stop", agentH.Stop)
	agents.Post("/:id/pause", agentH.Pause)
	agents.Post("/:id/resume", agentH.Resume)
	agents.Get("/:id/logs", agentH.Logs)

	return app, db, agentSvc
}

// TestAgentIntegration_FullWorkflow tests complete agent workflow
func TestAgentIntegration_FullWorkflow(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	t.Run("create agent", func(t *testing.T) {
		body := models.AgentCreate{
			Name:         "Workflow Agent",
			Provider:     "openai",
			Model:        "gpt-4",
			SystemPrompt: "You are helpful",
			WorkDir:      "/tmp/workflow",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var agent models.Agent
		err = json.NewDecoder(resp.Body).Decode(&agent)
		require.NoError(t, err)

		assert.Equal(t, "Workflow Agent", agent.Name)
		assert.Equal(t, models.AgentStatusIdle, agent.Status)
		assert.NotEmpty(t, agent.ID)
	})
}

// TestAgentIntegration_ListWorkflow tests listing and filtering workflow
func TestAgentIntegration_ListWorkflow(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	// Create agents
	agents := []models.AgentCreate{
		{Name: "Agent A", Provider: "openai", Model: "gpt-4"},
		{Name: "Agent B", Provider: "anthropic", Model: "claude-3"},
		{Name: "Agent C", Provider: "ollama", Model: "llama2"},
	}

	var createdIDs []string
	for _, agentDef := range agents {
		bodyBytes, _ := json.Marshal(agentDef)
		req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)

		var agent models.Agent
		err = json.NewDecoder(resp.Body).Decode(&agent)
		require.NoError(t, err)
		createdIDs = append(createdIDs, agent.ID)
	}

	t.Run("list all agents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result []models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("get specific agent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/"+createdIDs[0], nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var agent models.Agent
		err = json.NewDecoder(resp.Body).Decode(&agent)
		require.NoError(t, err)
		assert.Equal(t, "Agent A", agent.Name)
	})
}

// TestAgentIntegration_UpdateWorkflow tests update workflow
func TestAgentIntegration_UpdateWorkflow(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	// Create agent
	body := models.AgentCreate{
		Name:         "Original Name",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Original",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	var agent models.Agent
	err = json.NewDecoder(resp.Body).Decode(&agent)
	require.NoError(t, err)

	agentID := agent.ID

	// Update agent
	updateBody := models.AgentUpdate{
		Name:         func() *string { s := "Updated Name"; return &s }(),
		SystemPrompt: func() *string { s := "Updated prompt"; return &s }(),
	}
	updateBytes, _ := json.Marshal(updateBody)

	req = httptest.NewRequest("PUT", "/api/agents/"+agentID, bytes.NewReader(updateBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var updated models.Agent
	err = json.NewDecoder(resp.Body).Decode(&updated)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated prompt", updated.SystemPrompt)
}

// TestAgentIntegration_ControlWorkflow tests start/stop/pause/resume workflow
func TestAgentIntegration_ControlWorkflow(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	// Create agent
	body := models.AgentCreate{
		Name:         "Control Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Test",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	var agent models.Agent
	err = json.NewDecoder(resp.Body).Decode(&agent)
	require.NoError(t, err)

	agentID := agent.ID

	t.Run("stop agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agentID+"/stop", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, models.AgentStatusIdle, result.Status)
	})

	t.Run("pause agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agentID+"/pause", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, models.AgentStatusPaused, result.Status)
	})

	t.Run("resume agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agentID+"/resume", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, models.AgentStatusRunning, result.Status)
	})
}

// TestAgentIntegration_LogsWorkflow tests logging workflow
func TestAgentIntegration_LogsWorkflow(t *testing.T) {
	app, db, agentSvc := setupIntegrationTestApp(t)
	defer db.Close()

	// Create agent
	agent, err := agentSvc.Create(&models.AgentCreate{
		Name:         "Logging Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Test",
	})
	require.NoError(t, err)

	// Add logs
	agentSvc.AddLog(agent.ID, models.LogLevelInfo, "Info message", nil)
	agentSvc.AddLog(agent.ID, models.LogLevelWarn, "Warning message", nil)
	agentSvc.AddLog(agent.ID, models.LogLevelError, "Error message", map[string]interface{}{"code": 500})

	t.Run("get logs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/"+agent.ID+"/logs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var logs []models.AgentLog
		err = json.NewDecoder(resp.Body).Decode(&logs)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 3)
	})

	t.Run("get logs with limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/"+agent.ID+"/logs?limit=2", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var logs []models.AgentLog
		err = json.NewDecoder(resp.Body).Decode(&logs)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(logs), 2)
	})
}

// TestAgentIntegration_DeleteWorkflow tests delete workflow
func TestAgentIntegration_DeleteWorkflow(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	// Create agent
	body := models.AgentCreate{
		Name:         "To Delete",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Test",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	var agent models.Agent
	err = json.NewDecoder(resp.Body).Decode(&agent)
	require.NoError(t, err)

	agentID := agent.ID

	// Delete agent
	req = httptest.NewRequest("DELETE", "/api/agents/"+agentID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	// Verify deletion
	req = httptest.NewRequest("GET", "/api/agents/"+agentID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestAgentIntegration_Validation tests validation errors
func TestAgentIntegration_Validation(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	t.Run("create with missing name", func(t *testing.T) {
		body := models.AgentCreate{
			Provider: "openai",
			Model:    "gpt-4",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var result fiber.Map
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Contains(t, result["error"], "name is required")
	})

	t.Run("create with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

// TestAgentIntegration_ErrorScenarios tests various error scenarios
func TestAgentIntegration_ErrorScenarios(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	invalidID := "00000000-0000-0000-0000-000000000000"

	t.Run("get non-existent agent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/"+invalidID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("update non-existent agent", func(t *testing.T) {
		body := models.AgentUpdate{
			Name: func() *string { s := "Test"; return &s }(),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/agents/"+invalidID, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("delete non-existent agent", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/agents/"+invalidID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("control non-existent agent - start", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+invalidID+"/start", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("control non-existent agent - stop", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+invalidID+"/stop", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

// TestAgentIntegration_OrganizationFiltering tests organization-based filtering
func TestAgentIntegration_OrganizationFiltering(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	orgID1 := "550e8400-e29b-41d4-a716-446655440000"
	orgID2 := "660e8400-e29b-41d4-a716-446655440001"

	// Create agents in different organizations
	agents := []struct {
		name   string
		orgID  string
	}{
		{"Org 1 Agent A", orgID1},
		{"Org 1 Agent B", orgID1},
		{"Org 2 Agent A", orgID2},
		{"No Org Agent", ""},
	}

	for _, a := range agents {
		body := models.AgentCreate{
			Name:           a.name,
			Provider:       "openai",
			Model:          "gpt-4",
			OrganizationID: a.orgID,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	}

	t.Run("filter by org 1", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents", nil)
		req.Header.Set("X-Organization-ID", orgID1)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result []models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Count agents from org 1
		org1Count := 0
		for _, agent := range result {
			if agent.OrganizationID == orgID1 {
				org1Count++
			}
		}
		assert.Equal(t, 2, org1Count)
	})

	t.Run("filter by org 2", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents", nil)
		req.Header.Set("X-Organization-ID", orgID2)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result []models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Count agents from org 2
		org2Count := 0
		for _, agent := range result {
			if agent.OrganizationID == orgID2 {
				org2Count++
			}
		}
		assert.Equal(t, 1, org2Count)
	})
}

// TestAgentIntegration_ConcurrentOperations tests concurrent agent operations
func TestAgentIntegration_ConcurrentOperations(t *testing.T) {
	app, db, _ := setupIntegrationTestApp(t)
	defer db.Close()

	t.Run("create multiple agents concurrently", func(t *testing.T) {
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func(i int) {
				body := models.AgentCreate{
					Name:         "Concurrent Agent " + string(rune('A'+i)),
					Provider:     "openai",
					Model:        "gpt-4",
					SystemPrompt: "Test",
				}
				bodyBytes, _ := json.Marshal(body)

				req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				assert.NoError(t, err)
				assert.Equal(t, 201, resp.StatusCode)

				done <- true
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
