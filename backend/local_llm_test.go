package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestLocalLLM_AgentResponse tests local LLM integration (requires LM Studio running on port 1234 and server on port 18765)
func TestLocalLLM_AgentResponse(t *testing.T) {
	t.Log("=== Testing Agent Response with Local LLM ===")

	// Skip in CI environments where these services are not available
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping local LLM test in CI environment")
	}

	// 1. Check if LM Studio is running
	t.Log("Step 1: Check LM Studio availability")
	resp, err := http.Get("http://localhost:1234/v1/models")
	if err != nil {
		t.Skip("LM Studio not running on port 1234")
	}
	defer resp.Body.Close()

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &models)

	if len(models.Data) == 0 {
		t.Fatal("No models available in LM Studio")
	}

	modelID := models.Data[0].ID
	t.Logf("✅ LM Studio running with model: %s", modelID)

	// 2. Create Agent with local LLM
	t.Log("Step 2: Create Agent with local LLM config")
	agentPayload := map[string]interface{}{
		"name":  "local-llm-agent",
		"model": modelID,
		"type":  "testing",
	}

	agentJSON, _ := json.Marshal(agentPayload)
	agentResp, err := http.Post("http://localhost:18765/api/agents", "application/json", bytes.NewReader(agentJSON))
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	defer agentResp.Body.Close()

	var agent struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.NewDecoder(agentResp.Body).Decode(&agent)

	t.Logf("✅ Agent created: %s (ID: %s)", agent.Name, agent.ID)

	// 3. Test LLM Completion via LM Studio
	t.Log("Step 3: Test LLM completion")
	completionPayload := map[string]interface{}{
		"model": modelID,
		"messages": []map[string]string{
			{"role": "user", "content": "Say 'Hello from Agent Orchestrator!'"},
		},
		"max_tokens": 50,
	}

	completionJSON, _ := json.Marshal(completionPayload)
	start := time.Now()
	llmResp, err := http.Post("http://localhost:1234/v1/chat/completions", "application/json", bytes.NewReader(completionJSON))
	if err != nil {
		t.Fatalf("Failed to call LLM: %v", err)
	}
	defer llmResp.Body.Close()

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(llmResp.Body).Decode(&completion)

	elapsed := time.Since(start)

	if len(completion.Choices) == 0 {
		t.Fatal("No response from LLM")
	}

	response := completion.Choices[0].Message.Content
	t.Logf("✅ LLM Response: %s", response)
	t.Logf("✅ Response time: %v", elapsed)

	// 4. Create TODO for agent
	t.Log("Step 4: Create TODO for agent")
	todoPayload := map[string]interface{}{
		"title":       "Test LLM Integration",
		"description": "Verify agent can respond using local LLM",
		"priority":    9,
		"agent_id":    agent.ID,
	}

	todoJSON, _ := json.Marshal(todoPayload)
	todoResp, err := http.Post("http://localhost:18765/api/todos", "application/json", bytes.NewReader(todoJSON))
	if err != nil {
		t.Fatalf("Failed to create TODO: %v", err)
	}
	defer todoResp.Body.Close()

	var todo struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	json.NewDecoder(todoResp.Body).Decode(&todo)

	t.Logf("✅ TODO created: %s (ID: %s)", todo.Title, todo.ID)

	// 5. Verify Integration
	t.Log("Step 5: Verify integration")
	statusResp, err := http.Get("http://localhost:18765/api/system/status")
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	defer statusResp.Body.Close()

	var status struct {
		Counts struct {
			Agents int `json:"agents"`
			TODOs  int `json:"todos"`
		} `json:"counts"`
	}
	json.NewDecoder(statusResp.Body).Decode(&status)

	t.Logf("✅ System status: %d agents, %d TODOs", status.Counts.Agents, status.Counts.TODOs)

	if status.Counts.Agents == 0 {
		t.Error("Expected at least 1 agent")
	}

	t.Log("✅ Local LLM integration test passed!")
}

// BenchmarkLocalLLM_ResponseTime benchmarks local LLM response time (requires LM Studio running on port 1234)
func BenchmarkLocalLLM_ResponseTime(b *testing.B) {
	// Skip in CI environments where LM Studio is not available
	if os.Getenv("CI") == "true" {
		b.Skip("Skipping benchmark in CI environment")
	}

	modelID := "qwen/qwen3.5-35b-a3b"

	for i := 0; i < b.N; i++ {
		completionPayload := map[string]interface{}{
			"model": modelID,
			"messages": []map[string]string{
				{"role": "user", "content": "Say 'test'"},
			},
			"max_tokens": 10,
		}

		completionJSON, _ := json.Marshal(completionPayload)
		_, err := http.Post("http://localhost:1234/v1/chat/completions", "application/json", bytes.NewReader(completionJSON))
		if err != nil {
			b.Skip("LM Studio not running")
		}
	}
}

// TestMultiModel_Agents tests multiple agents with different models (requires server running on port 18765)
func TestMultiModel_Agents(t *testing.T) {
	t.Log("=== Testing Multiple Agents with Different Models ===")

	// Skip in CI environments where the server is not available
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping multi-model test in CI environment")
	}

	models := []string{
		"qwen/qwen3.5-35b-a3b",
		"local-model-1",
	}

	for i, model := range models {
		agentPayload := map[string]interface{}{
			"name":  fmt.Sprintf("agent-%d", i+1),
			"model": model,
		}

		agentJSON, _ := json.Marshal(agentPayload)
		resp, err := http.Post("http://localhost:18765/api/agents", "application/json", bytes.NewReader(agentJSON))
		if err != nil {
			t.Logf("⚠️  Failed to create agent with model %s: %v", model, err)
			continue
		}
		defer resp.Body.Close()

		t.Logf("✅ Agent created with model: %s", model)
	}
}
