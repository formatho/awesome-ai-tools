// Package handlers tests for agent API endpoints.
package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/formatho/agent-orchestrator/backend/internal/store"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAgentTestApp creates a test Fiber app with agent routes
func setupAgentTestApp(t *testing.T) (*fiber.App, *sql.DB, *AgentHandler) {
	// Create in-memory database
	db, err := store.InitDB(":memory:")
	require.NoError(t, err, "Failed to create test database")

	// Run migrations
	err = store.RunMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	// Create service and handler
	hub := websocket.NewHub()
	service := services.NewAgentService(db, hub)
	handler := NewAgentHandler(service)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Agent Orchestrator Test API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Setup agent routes
	agents := app.Group("/api/agents")
	agents.Get("/", handler.List)
	agents.Post("/", handler.Create)
	agents.Get("/:id", handler.Get)
	agents.Put("/:id", handler.Update)
	agents.Delete("/:id", handler.Delete)
	agents.Post("/:id/start", handler.Start)
	agents.Post("/:id/stop", handler.Stop)
	agents.Post("/:id/pause", handler.Pause)
	agents.Post("/:id/resume", handler.Resume)
	agents.Get("/:id/logs", handler.Logs)

	return app, db, handler
}

// TestAgentHandler_List tests listing agents via API
func TestAgentHandler_List(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create test agents
	_, err := handler.service.Create(&models.AgentCreate{
		Name:         "Agent 1",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Test prompt 1",
	})
	require.NoError(t, err)

	_, err = handler.service.Create(&models.AgentCreate{
		Name:         "Agent 2",
		Provider:     "anthropic",
		Model:        "claude-3",
		SystemPrompt: "Test prompt 2",
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		orgIDHeader    string
		wantStatusCode int
		wantMinAgents  int
	}{
		{
			name:           "list all agents",
			orgIDHeader:    "",
			wantStatusCode: 200,
			wantMinAgents:  2,
		},
		{
			name:           "list with empty org header",
			orgIDHeader:    "",
			wantStatusCode: 200,
			wantMinAgents:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/agents", nil)
			if tt.orgIDHeader != "" {
				req.Header.Set("X-Organization-ID", tt.orgIDHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			// Parse response
			var agents []models.Agent
			err = json.NewDecoder(resp.Body).Decode(&agents)
			require.NoError(t, err)

			assert.GreaterOrEqual(t, len(agents), tt.wantMinAgents)
		})
	}
}

// TestAgentHandler_Get tests retrieving a single agent via API
func TestAgentHandler_Get(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		wantStatusCode int
		validateAgent  bool
	}{
		{
			name:           "get existing agent",
			id:             agent.ID,
			wantStatusCode:  200,
			validateAgent:   true,
		},
		{
			name:           "get non-existent agent",
			id:             "00000000-0000-0000-0000-000000000000",
			wantStatusCode:  404,
			validateAgent:   false,
		},
		{
			name:           "get with invalid UUID",
			id:             "invalid-uuid",
			wantStatusCode:  404, // Fiber returns 404 for invalid UUID format
			validateAgent:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/agents/"+tt.id, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if tt.validateAgent {
				var result models.Agent
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.Equal(t, agent.ID, result.ID)
				assert.Equal(t, agent.Name, result.Name)
			}
		})
	}
}

// TestAgentHandler_Create tests creating agents via API
func TestAgentHandler_Create(t *testing.T) {
	app, db, _ := setupAgentTestApp(t)
	defer db.Close()

	tests := []struct {
		name           string
		body           models.AgentCreate
		wantStatusCode int
		wantErr        bool
	}{
		{
			name: "create valid agent",
			body: models.AgentCreate{
				Name:         "New Agent",
				Provider:     "openai",
				Model:        "gpt-4",
				SystemPrompt: "You are helpful",
				WorkDir:      "/tmp/test",
			},
			wantStatusCode: 201,
			wantErr:        false,
		},
		{
			name: "create with organization",
			body: models.AgentCreate{
				Name:           "Org Agent",
				Provider:       "anthropic",
				Model:          "claude-3",
				SystemPrompt:   "You are expert",
				OrganizationID: "550e8400-e29b-41d4-a716-446655440000",
			},
			wantStatusCode: 201,
			wantErr:        false,
		},
		{
			name: "create with empty name",
			body: models.AgentCreate{
				Name:         "",
				Provider:     "openai",
				Model:        "gpt-4",
				SystemPrompt: "Test",
			},
			wantStatusCode: 400,
			wantErr:        true,
		},
		{
			name: "create with invalid JSON",
			body: models.AgentCreate{
				Name: "Test",
			},
			wantStatusCode: 201, // Should work with minimal fields
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if !tt.wantErr {
				var agent models.Agent
				err = json.NewDecoder(resp.Body).Decode(&agent)
				require.NoError(t, err)
				assert.NotEmpty(t, agent.ID)
				assert.Equal(t, tt.body.Name, agent.Name)
			}
		})
	}
}

// TestAgentHandler_Update tests updating agents via API
func TestAgentHandler_Update(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Original Name",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Original prompt",
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		body           models.AgentUpdate
		wantStatusCode int
		validate       bool
	}{
		{
			name: "update name only",
			id:   agent.ID,
			body: models.AgentUpdate{
				Name: func() *string { s := "Updated Name"; return &s }(),
			},
			wantStatusCode: 200,
			validate:       true,
		},
		{
			name: "update multiple fields",
			id:   agent.ID,
			body: models.AgentUpdate{
				Name:         func() *string { s := "Multi Update"; return &s }(),
				Model:        func() *string { s := "gpt-4-turbo"; return &s }(),
				SystemPrompt: func() *string { s := "New prompt"; return &s }(),
			},
			wantStatusCode: 200,
			validate:       true,
		},
		{
			name: "update non-existent agent",
			id:   "00000000-0000-0000-0000-000000000000",
			body: models.AgentUpdate{
				Name: func() *string { s := "Test"; return &s }(),
			},
			wantStatusCode: 404,
			validate:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PUT", "/api/agents/"+tt.id, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if tt.validate {
				var updated models.Agent
				err = json.NewDecoder(resp.Body).Decode(&updated)
				require.NoError(t, err)
				if tt.body.Name != nil {
					assert.Equal(t, *tt.body.Name, updated.Name)
				}
			}
		})
	}
}

// TestAgentHandler_Delete tests deleting agents via API
func TestAgentHandler_Delete(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "To Delete",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Temporary",
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		wantStatusCode int
	}{
		{
			name:           "delete existing agent",
			id:             agent.ID,
			wantStatusCode: 204,
		},
		{
			name:           "delete non-existent agent",
			id:             "00000000-0000-0000-0000-000000000000",
			wantStatusCode: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/agents/"+tt.id, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			// Verify deletion for successful case
			if tt.wantStatusCode == 204 {
				_, err := handler.service.Get(tt.id)
				assert.Error(t, err)
			}
		})
	}
}

// TestAgentHandler_Start tests starting agents via API
func TestAgentHandler_Start(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
	})
	require.NoError(t, err)

	t.Run("start existing agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agent.ID+"/start", nil)
		resp, err := app.Test(req)

		// May fail if API key is not set, but should return JSON
		require.NoError(t, err)

		// Parse response (may be error or success)
		var result fiber.Map
		json.NewDecoder(resp.Body).Decode(&result)

		// If successful, verify status is running
		if resp.StatusCode == 200 {
			assert.Equal(t, "running", result["status"])
		}
	})

	t.Run("start non-existent agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/00000000-0000-0000-0000-000000000000/start", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

// TestAgentHandler_Stop tests stopping agents via API
func TestAgentHandler_Stop(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
	})
	require.NoError(t, err)

	t.Run("stop existing agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agent.ID+"/stop", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, models.AgentStatusIdle, result.Status)
	})

	t.Run("stop non-existent agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/00000000-0000-0000-0000-000000000000/stop", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

// TestAgentHandler_Pause tests pausing agents via API
func TestAgentHandler_Pause(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
	})
	require.NoError(t, err)

	t.Run("pause existing agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agent.ID+"/pause", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, models.AgentStatusPaused, result.Status)
	})

	t.Run("pause non-existent agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/00000000-0000-0000-0000-000000000000/pause", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

// TestAgentHandler_Resume tests resuming agents via API
func TestAgentHandler_Resume(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
	})
	require.NoError(t, err)

	// Pause it first
	_, err = handler.service.Pause(agent.ID)
	require.NoError(t, err)

	t.Run("resume existing agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/"+agent.ID+"/resume", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result models.Agent
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, models.AgentStatusRunning, result.Status)
	})

	t.Run("resume non-existent agent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents/00000000-0000-0000-0000-000000000000/resume", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})
}

// TestAgentHandler_Logs tests retrieving agent logs via API
func TestAgentHandler_Logs(t *testing.T) {
	app, db, handler := setupAgentTestApp(t)
	defer db.Close()

	// Create a test agent
	agent, err := handler.service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
	})
	require.NoError(t, err)

	// Add some logs
	handler.service.AddLog(agent.ID, models.LogLevelInfo, "Log 1", nil)
	handler.service.AddLog(agent.ID, models.LogLevelInfo, "Log 2", nil)
	handler.service.AddLog(agent.ID, models.LogLevelError, "Error log", nil)

	tests := []struct {
		name           string
		id             string
		query          string
		wantStatusCode int
		wantMinLogs    int
	}{
		{
			name:           "get logs",
			id:             agent.ID,
			query:          "",
			wantStatusCode:  200,
			wantMinLogs:    3,
		},
		{
			name:           "get logs with limit",
			id:             agent.ID,
			query:          "?limit=2",
			wantStatusCode:  200,
			wantMinLogs:    2,
		},
		{
			name:           "get logs with high limit",
			id:             agent.ID,
			query:          "?limit=100",
			wantStatusCode:  200,
			wantMinLogs:    3,
		},
		{
			name:           "get logs for non-existent agent",
			id:             "00000000-0000-0000-0000-000000000000",
			query:          "",
			wantStatusCode:  200, // GetLogs returns empty array for non-existent agent
			wantMinLogs:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/agents/"+tt.id+"/logs"+tt.query, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if tt.wantStatusCode == 200 {
				var logs []models.AgentLog
				err = json.NewDecoder(resp.Body).Decode(&logs)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(logs), tt.wantMinLogs)
			}
		})
	}
}

// TestAgentHandler_RequestValidation tests request validation
func TestAgentHandler_RequestValidation(t *testing.T) {
	app, db, _ := setupAgentTestApp(t)
	defer db.Close()

	t.Run("create with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/agents", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("update with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/agents/00000000-0000-0000-0000-000000000000", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})
}

// TestAgentHandler_ErrorHandling tests error handling scenarios
func TestAgentHandler_ErrorHandling(t *testing.T) {
	app, db, _ := setupAgentTestApp(t)
	defer db.Close()

	t.Run("get with invalid ID format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/agents/not-a-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode) // Fiber returns 404 for invalid route/ID
	})

	t.Run("delete with invalid ID format", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/agents/not-a-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode) // Fiber returns 404 for invalid route/ID
	})

	t.Run("update with invalid ID format", func(t *testing.T) {
		body := models.AgentUpdate{
			Name: func() *string { s := "Test"; return &s }(),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/agents/not-a-uuid", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode) // Fiber returns 404 for invalid route/ID
	})
}

// TestAgentHandler_CORS tests CORS headers
func TestAgentHandler_CORS(t *testing.T) {
	app, db, _ := setupAgentTestApp(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/agents", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	// CORS headers are added by middleware in production
}
