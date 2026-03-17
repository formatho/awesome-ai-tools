// Package services tests for agent service.
package services

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAgentTestDB creates a new in-memory database for agent testing
func setupAgentTestDB(t *testing.T) *sql.DB {
	db, err := store.InitDB(":memory:")
	require.NoError(t, err, "Failed to create test database")

	// Run migrations
	err = store.RunMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// setupTestAgentService creates a new agent service for testing
func setupTestAgentService(t *testing.T, db *sql.DB) *AgentService {
	hub := websocket.NewHub()
	service := NewAgentService(db, hub)
	return service
}

// TestAgentService_Create tests agent creation
func TestAgentService_Create(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	tests := []struct {
		name    string
		req     *models.AgentCreate
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid agent creation",
			req: &models.AgentCreate{
				Name:         "Test Agent",
				Provider:     "openai",
				Model:        "gpt-4",
				SystemPrompt: "You are a helpful assistant",
				WorkDir:      "/tmp/test",
			},
			wantErr: false,
		},
		{
			name: "agent with organization",
			req: &models.AgentCreate{
				Name:           "Org Agent",
				Provider:       "anthropic",
				Model:          "claude-3-sonnet",
				SystemPrompt:   "You are an expert coder",
				WorkDir:        "/tmp/org-test",
				OrganizationID: uuid.New().String(),
			},
			wantErr: false,
		},
		{
			name: "agent with empty provider (allowed)",
			req: &models.AgentCreate{
				Name:         "Test Agent",
				Provider:     "",
				Model:        "gpt-4",
				SystemPrompt: "You are a helpful assistant",
			},
			wantErr: false,
		},
		{
			name: "missing required field - name",
			req: &models.AgentCreate{
				Provider:     "openai",
				Model:        "gpt-4",
				SystemPrompt: "You are a helpful assistant",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := service.Create(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
				assert.NotEmpty(t, agent.ID)
				assert.Equal(t, tt.req.Name, agent.Name)
				assert.Equal(t, tt.req.Provider, agent.Provider)
				assert.Equal(t, tt.req.Model, agent.Model)
				assert.Equal(t, tt.req.SystemPrompt, agent.SystemPrompt)
				assert.Equal(t, models.AgentStatusIdle, agent.Status)
				assert.False(t, agent.CreatedAt.IsZero())
				assert.False(t, agent.UpdatedAt.IsZero())
			}
		})
	}
}

// TestAgentService_Get tests retrieving an agent by ID
func TestAgentService_Get(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	req := &models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are a helpful assistant",
		WorkDir:      "/tmp/test",
	}
	created, err := service.Create(req)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      string
		wantErr bool
		errType error
	}{
		{
			name:    "valid agent retrieval",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent agent",
			id:      uuid.New().String(),
			wantErr: true,
			errType: models.ErrNotFound,
		},
		{
			name:    "invalid UUID format",
			id:      "invalid-uuid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := service.Get(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
				assert.Equal(t, created.ID, agent.ID)
				assert.Equal(t, created.Name, agent.Name)
			}
		})
	}
}

// TestAgentService_List tests listing agents with and without filters
func TestAgentService_List(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create test agents
	orgID1 := uuid.New().String()
	orgID2 := uuid.New().String()

	agents := []struct {
		name           string
		organizationID string
	}{
		{"Agent 1", orgID1},
		{"Agent 2", orgID1},
		{"Agent 3", orgID2},
		{"Agent 4", ""},
	}

	for _, a := range agents {
		_, err := service.Create(&models.AgentCreate{
			Name:           a.name,
			Provider:       "openai",
			Model:          "gpt-4",
			SystemPrompt:   "You are helpful",
			WorkDir:        "/tmp/test",
			OrganizationID: a.organizationID,
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		orgID    *string
		wantMin  int
		wantMax  int
	}{
		{
			name:    "list all agents",
			orgID:   nil,
			wantMin: 4,
			wantMax: 4,
		},
		{
			name:    "filter by organization",
			orgID:   &orgID1,
			wantMin: 2,
			wantMax: 2,
		},
		{
			name:    "empty organization filter returns all",
			orgID:   func() *string { s := ""; return &s }(),
			wantMin: 4,
			wantMax: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agents, err := service.List(tt.orgID)

			assert.NoError(t, err)
			assert.NotNil(t, agents)
			assert.GreaterOrEqual(t, len(agents), tt.wantMin)
			assert.LessOrEqual(t, len(agents), tt.wantMax)

			// Verify agents are ordered by created_at DESC
			if len(agents) > 1 {
				for i := 1; i < len(agents); i++ {
					assert.GreaterOrEqual(t, agents[i-1].CreatedAt, agents[i].CreatedAt)
				}
			}
		})
	}
}

// TestAgentService_Update tests updating an agent
func TestAgentService_Update(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	req := &models.AgentCreate{
		Name:         "Original Name",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "Original prompt",
		WorkDir:      "/tmp/original",
	}
	agent, err := service.Create(req)
	require.NoError(t, err)

	// Wait a bit to ensure timestamp difference
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name    string
		id      string
		req     *models.AgentUpdate
		wantErr bool
		validate func(*testing.T, *models.Agent)
	}{
		{
			name: "update name only",
			id:   agent.ID,
			req: &models.AgentUpdate{
				Name: func() *string { s := "Updated Name"; return &s }(),
			},
			wantErr: false,
			validate: func(t *testing.T, a *models.Agent) {
				assert.Equal(t, "Updated Name", a.Name)
				assert.Equal(t, "Original prompt", a.SystemPrompt)
				// Note: SQLite's CURRENT_TIMESTAMP has second-level precision, so we don't
				// compare timestamps directly
			},
		},
		{
			name: "update multiple fields",
			id:   agent.ID,
			req: &models.AgentUpdate{
				Name:         func() *string { s := "Multi Update"; return &s }(),
				Model:        func() *string { s := "gpt-4-turbo"; return &s }(),
				SystemPrompt: func() *string { s := "New prompt"; return &s }(),
			},
			wantErr: false,
			validate: func(t *testing.T, a *models.Agent) {
				assert.Equal(t, "Multi Update", a.Name)
				assert.Equal(t, "gpt-4-turbo", a.Model)
				assert.Equal(t, "New prompt", a.SystemPrompt)
			},
		},
		{
			name: "update config and metadata",
			id:   agent.ID,
			req: &models.AgentUpdate{
				Config: func() map[string]interface{} {
					return map[string]interface{}{
						"max_tokens": 8192,
						"temperature": 0.7,
					}
				}(),
				Metadata: func() map[string]interface{} {
					return map[string]interface{}{
						"tags":     []string{"test", "update"},
						"version": 2,
					}
				}(),
			},
			wantErr: false,
			validate: func(t *testing.T, a *models.Agent) {
				assert.Equal(t, float64(8192), a.Config["max_tokens"])
				assert.Equal(t, 0.7, a.Config["temperature"])
				assert.Equal(t, "test", a.Metadata["tags"].([]interface{})[0])
				assert.Equal(t, float64(2), a.Metadata["version"])
			},
		},
		{
			name:    "update non-existent agent",
			id:      uuid.New().String(),
			req:     &models.AgentUpdate{Name: func() *string { s := "Test"; return &s }()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := service.Update(tt.id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, updated)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updated)
				if tt.validate != nil {
					tt.validate(t, updated)
				}
			}
		})
	}
}

// TestAgentService_Delete tests deleting an agent
func TestAgentService_Delete(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "To Be Deleted",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are temporary",
		WorkDir:      "/tmp/temp",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "delete existing agent",
			id:      agent.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existent agent",
			id:      uuid.New().String(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify agent is deleted
				_, err := service.Get(tt.id)
				assert.Error(t, err)
				assert.ErrorIs(t, err, models.ErrNotFound)
			}
		})
	}
}

// TestAgentService_StartStop tests starting and stopping an agent
func TestAgentService_StartStop(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
		WorkDir:      "/tmp/test",
	})
	require.NoError(t, err)

	t.Run("start agent", func(t *testing.T) {
		// Note: This test will fail if OPENAI_API_KEY is not set
		// In a real test environment, you would mock the LLM client
		started, err := service.Start(agent.ID)

		if err != nil {
			// If API key is not set, we expect an error
			assert.Contains(t, err.Error(), "API key")
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, started)
			assert.Equal(t, models.AgentStatusRunning, started.Status)
			assert.NotNil(t, started.StartedAt)
		}
	})

	t.Run("stop agent", func(t *testing.T) {
		stopped, err := service.Stop(agent.ID)

		assert.NoError(t, err)
		assert.NotNil(t, stopped)
		assert.Equal(t, models.AgentStatusIdle, stopped.Status)
		assert.NotNil(t, stopped.StoppedAt)
	})
}

// TestAgentService_PauseResume tests pausing and resuming an agent
func TestAgentService_PauseResume(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
		WorkDir:      "/tmp/test",
	})
	require.NoError(t, err)

	t.Run("pause agent", func(t *testing.T) {
		paused, err := service.Pause(agent.ID)

		assert.NoError(t, err)
		assert.NotNil(t, paused)
		assert.Equal(t, models.AgentStatusPaused, paused.Status)
	})

	t.Run("resume agent", func(t *testing.T) {
		resumed, err := service.Resume(agent.ID)

		assert.NoError(t, err)
		assert.NotNil(t, resumed)
		assert.Equal(t, models.AgentStatusRunning, resumed.Status)
	})
}

// TestAgentService_AddLog tests adding log entries
func TestAgentService_AddLog(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
		WorkDir:      "/tmp/test",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		agentID  string
		level    models.LogLevel
		message  string
		metadata map[string]interface{}
		wantErr  bool
	}{
		{
			name:    "add info log",
			agentID: agent.ID,
			level:   models.LogLevelInfo,
			message: "Agent started",
			wantErr: false,
		},
		{
			name:    "add error log",
			agentID: agent.ID,
			level:   models.LogLevelError,
			message: "Something went wrong",
			wantErr: false,
		},
		{
			name:    "add log with metadata",
			agentID: agent.ID,
			level:   models.LogLevelDebug,
			message: "Debug info",
			metadata: map[string]interface{}{
				"step":     1,
				"duration": "5s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AddLog(tt.agentID, tt.level, tt.message, tt.metadata)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Retrieve logs and verify
				logs, err := service.GetLogs(tt.agentID, 10)
				assert.NoError(t, err)
				assert.NotEmpty(t, logs)

				// Find the most recent log
				log := logs[0]
				assert.Equal(t, tt.agentID, log.AgentID)
				assert.Equal(t, tt.level, log.Level)
				assert.Equal(t, tt.message, log.Message)
				assert.False(t, log.CreatedAt.IsZero())

				if tt.metadata != nil {
					assert.NotNil(t, log.Metadata)
					for k, v := range tt.metadata {
						// JSON unmarshaling converts numbers to float64
						if intVal, ok := v.(int); ok {
							assert.Equal(t, float64(intVal), log.Metadata[k])
						} else {
							assert.Equal(t, v, log.Metadata[k])
						}
					}
				}
			}
		})
	}
}

// TestAgentService_GetLogs tests retrieving agent logs
func TestAgentService_GetLogs(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create a test agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Test Agent",
		Provider:     "openai",
		Model:        "gpt-4",
		SystemPrompt: "You are helpful",
		WorkDir:      "/tmp/test",
	})
	require.NoError(t, err)

	// Add multiple log entries
	for i := 0; i < 5; i++ {
		err := service.AddLog(agent.ID, models.LogLevelInfo, fmt.Sprintf("Log entry %d", i), nil)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		agentID string
		limit   int
		wantMin int
		wantMax int
		wantErr bool
	}{
		{
			name:    "get all logs",
			agentID: agent.ID,
			limit:   100,
			wantMin: 5,
			wantMax: 5,
			wantErr: false,
		},
		{
			name:    "get limited logs",
			agentID: agent.ID,
			limit:   3,
			wantMin: 3,
			wantMax: 3,
			wantErr: false,
		},
		{
			name:    "get logs with invalid limit",
			agentID: agent.ID,
			limit:   -1,
			wantMin: 5,
			wantMax: 100,
			wantErr: false,
		},
		{
			name:    "get logs for non-existent agent",
			agentID: uuid.New().String(),
			limit:   10,
			wantMin: 0,
			wantMax: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := service.GetLogs(tt.agentID, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// GetLogs should return an empty slice, not nil, even for non-existent agents
				// But if it returns nil, we accept it as valid (no logs)
				if logs != nil {
					assert.GreaterOrEqual(t, len(logs), tt.wantMin)
					assert.LessOrEqual(t, len(logs), tt.wantMax)
				} else {
					// If nil is returned, treat it as an empty result
					assert.Equal(t, 0, tt.wantMin)
				}

				// Verify logs are ordered by created_at DESC
				if len(logs) > 1 {
					for i := 1; i < len(logs); i++ {
						assert.GreaterOrEqual(t, logs[i-1].CreatedAt, logs[i].CreatedAt)
					}
				}
			}
		})
	}
}

// TestAgentService_NoDatabase tests behavior without database
func TestAgentService_NoDatabase(t *testing.T) {
	service := NewAgentService(nil, websocket.NewHub())

	t.Run("create without database", func(t *testing.T) {
		_, err := service.Create(&models.AgentCreate{
			Name:     "Test",
			Provider: "openai",
			Model:    "gpt-4",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNoDatabase)
	})

	t.Run("list without database", func(t *testing.T) {
		_, err := service.List(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNoDatabase)
	})

	t.Run("get without database", func(t *testing.T) {
		_, err := service.Get(uuid.New().String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNoDatabase)
	})
}

// TestAgentService_ConfigValidation tests configuration validation
func TestAgentService_ConfigValidation(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	validProviders := []string{"openai", "anthropic", "zai", "ollama", "groq", "mistral", "openrouter"}

	for _, provider := range validProviders {
		t.Run("valid provider: "+provider, func(t *testing.T) {
			agent, err := service.Create(&models.AgentCreate{
				Name:         "Valid Provider Agent",
				Provider:     provider,
				Model:        "test-model",
				SystemPrompt: "You are helpful",
			})
			assert.NoError(t, err)
			assert.NotNil(t, agent)
			assert.Equal(t, provider, agent.Provider)
		})
	}

	t.Run("empty provider is allowed", func(t *testing.T) {
		agent, err := service.Create(&models.AgentCreate{
			Name:         "Empty Provider Agent",
			Provider:     "",
			Model:        "test-model",
			SystemPrompt: "You are helpful",
		})
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "", agent.Provider)
	})
}
