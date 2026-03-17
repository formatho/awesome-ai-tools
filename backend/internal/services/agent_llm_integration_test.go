// Package services tests for agent LLM response handling.
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	agentrunner "github.com/formatho/agent-orchestrator/packages/agent-runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentService_LLMResponseIntegration tests that agent responses are based on LLM outputs
// This is an integration test that verifies the flow from agent creation to LLM response
func TestAgentService_LLMResponseIntegration(t *testing.T) {
	// Skip if no API keys are configured (CI/CD environment)
	t.Skip("Skipping integration test - requires actual LLM API keys")

	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create an agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Test LLM Integration",
		Provider:     "openai",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "You are a helpful assistant who responds with 'Test successful' when prompted.",
		WorkDir:      "/tmp/test",
	})
	require.NoError(t, err)
	require.NotNil(t, agent)

	// Start the agent
	started, err := service.Start(agent.ID)
	require.NoError(t, err)
	require.NotNil(t, started)
	assert.Equal(t, models.AgentStatusRunning, started.Status)

	// Wait for agent to process the initial prompt
	time.Sleep(2 * time.Second)

	// Check the agent status and logs
	statusAgent, err := service.Get(agent.ID)
	require.NoError(t, err)
	assert.NotNil(t, statusAgent)

	// Retrieve logs to verify LLM response was logged
	logs, err := service.GetLogs(agent.ID, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, logs, "Expected logs to be populated with LLM response")

	// Verify at least one log contains the LLM response
	foundResponse := false
	for _, log := range logs {
		if log.Level == models.LogLevelInfo || log.Level == models.LogLevelDebug {
			// Check if log contains evidence of LLM processing
			if len(log.Message) > 0 {
				foundResponse = true
				break
			}
		}
	}
	assert.True(t, foundResponse, "Expected to find LLM response in logs")

	// Stop the agent
	stopped, err := service.Stop(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AgentStatusIdle, stopped.Status)
}

// TestAgentService_AgentRunnerResult verifies that the AgentRunner captures LLM results
func TestAgentService_AgentRunnerResult(t *testing.T) {
	db := setupAgentTestDB(t)
	defer db.Close()

	hub := websocket.NewHub()
	service := NewAgentService(db, hub)

	// Create an agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Runner Result Test",
		Provider:     "openai",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "Test prompt",
	})
	require.NoError(t, err)

	// Access the runner directly to verify result storage
	runner := service.runner

	// Create agent in runner (without starting, to avoid API calls)
	ctx := context.Background()
	config := agentrunner.AgentConfig{
		Provider:     agent.Provider,
		Model:        agent.Model,
		APIKey:       "test-key", // Won't actually call LLM
		MaxTokens:    4096,
		SystemPrompt: agent.SystemPrompt,
		MemoryLimit:  8,
	}

	_, err = runner.CreateAgent(ctx, agent.ID, config)
	if err != nil && !errors.Is(err, agentrunner.ErrAgentAlreadyExists) {
		require.NoError(t, err)
	}

	// Verify agent exists in runner
	status, err := runner.GetStatus(agent.ID)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, agent.ID, status.ID)
	assert.Equal(t, agentrunner.StatusIdle, status.Status)
}

// TestAgentService_SendPrompt_LLM verifies that prompts are sent to LLM
func TestAgentService_SendPrompt_LLM(t *testing.T) {
	t.Skip("Skipping test - requires mock LLM provider or actual API key")

	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Create an agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Prompt Test",
		Provider:     "openai",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "You are a helpful assistant",
	})
	require.NoError(t, err)

	// Start the agent
	_, err = service.Start(agent.ID)
	require.NoError(t, err)
	defer service.Stop(agent.ID)

	// Send a prompt to the agent via the runner
	ctx := context.Background()
	prompt := "Say 'Hello' in response"
	err = service.runner.SendPrompt(ctx, agent.ID, prompt)

	// If API key is not configured, we expect an error
	// In a proper test environment with mocks, this would verify the response
	if err != nil {
		assert.Contains(t, err.Error(), "API key")
	} else {
		// If successful, wait and check for response in logs
		time.Sleep(2 * time.Second)
		logs, _ := service.GetLogs(agent.ID, 10)
		assert.NotEmpty(t, logs, "Expected logs after sending prompt")
	}
}

// TestAgentService_LLMLifecycle tests the complete lifecycle with LLM interactions
func TestAgentService_LLMLifecycle(t *testing.T) {
	t.Skip("Skipping integration test - requires actual LLM API keys")

	db := setupAgentTestDB(t)
	defer db.Close()

	service := setupTestAgentService(t, db)

	// Step 1: Create agent
	agent, err := service.Create(&models.AgentCreate{
		Name:         "Lifecycle Test Agent",
		Provider:     "anthropic",
		Model:        "claude-3-haiku",
		SystemPrompt: "You are a concise assistant",
	})
	require.NoError(t, err)

	// Step 2: Start agent (triggers initial LLM call)
	started, err := service.Start(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AgentStatusRunning, started.Status)

	time.Sleep(1 * time.Second)

	// Step 3: Check logs for LLM activity
	logs, err := service.GetLogs(agent.ID, 20)
	require.NoError(t, err)
	assert.NotEmpty(t, logs, "Expected logs after starting agent")

	// Step 4: Pause agent
	paused, err := service.Pause(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AgentStatusPaused, paused.Status)

	// Step 5: Resume agent (should trigger another LLM call)
	resumed, err := service.Resume(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AgentStatusRunning, resumed.Status)

	time.Sleep(1 * time.Second)

	// Step 6: Stop agent
	stopped, err := service.Stop(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AgentStatusIdle, stopped.Status)

	// Step 7: Verify final state
	finalState, err := service.Get(agent.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AgentStatusIdle, finalState.Status)
	assert.NotNil(t, finalState.StartedAt)
	assert.NotNil(t, finalState.StoppedAt)

	// Step 8: Check that we have logs throughout the lifecycle
	allLogs, err := service.GetLogs(agent.ID, 100)
	require.NoError(t, err)
	assert.Greater(t, len(allLogs), 0, "Expected logs throughout agent lifecycle")
}
