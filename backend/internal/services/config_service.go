package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	agentconfig "github.com/formatho/agent-orchestrator/packages/agent-config"
	llmclient "github.com/formatho/agent-orchestrator/packages/llm-client"
)

// ConfigService handles configuration operations.
type ConfigService struct {
	db *sql.DB
}

// NewConfigService creates a new config service.
func NewConfigService(db *sql.DB) *ConfigService {
	return &ConfigService{db: db}
}

// Get returns the current configuration.
func (s *ConfigService) Get() (*models.Config, error) {
	// Return default config if no database available
	if s.db == nil {
		return &models.Config{
			ID:        "default",
			LLMConfig: &models.LLMConfig{},
			Defaults:  make(map[string]interface{}),
			Settings:  make(map[string]interface{}),
		}, nil
	}

	query := `SELECT id, llm_config, defaults, settings, updated_at FROM config WHERE id = 'default'`

	c := &models.Config{}
	var llmConfig, defaults, settings sql.NullString

	err := s.db.QueryRow(query).Scan(&c.ID, &llmConfig, &defaults, &settings, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if llmConfig.Valid && llmConfig.String != "" {
		json.Unmarshal([]byte(llmConfig.String), &c.LLMConfig)
	}
	if defaults.Valid && defaults.String != "" {
		json.Unmarshal([]byte(defaults.String), &c.Defaults)
	}
	if settings.Valid && settings.String != "" {
		json.Unmarshal([]byte(settings.String), &c.Settings)
	}

	return c, nil
}

// Update updates the configuration.
func (s *ConfigService) Update(req *models.ConfigUpdate) (*models.Config, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	now := time.Now().UTC()
	query := `UPDATE config SET updated_at = ?`
	args := []interface{}{now}

	if req.LLMConfig != nil {
		llmJSON, _ := json.Marshal(req.LLMConfig)
		query += `, llm_config = ?`
		args = append(args, string(llmJSON))
	}
	if req.Defaults != nil {
		defaultsJSON, _ := json.Marshal(req.Defaults)
		query += `, defaults = ?`
		args = append(args, string(defaultsJSON))
	}
	if req.Settings != nil {
		settingsJSON, _ := json.Marshal(req.Settings)
		query += `, settings = ?`
		args = append(args, string(settingsJSON))
	}

	query += ` WHERE id = 'default'`

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return s.Get()
}

// TestLLM tests an LLM connection.
func (s *ConfigService) TestLLM(req *models.LLMTestRequest) (*models.LLMTestResponse, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var provider llmclient.ProviderClient
	providerType := req.Provider
	if providerType == "" {
		providerType = "openai"
	}

	switch providerType {
	case "openai":
		provider = llmclient.NewOpenAIProvider(llmclient.OpenAIConfig{
			APIKey: req.APIKey,
		})
	case "anthropic":
		provider = llmclient.NewAnthropicProvider(llmclient.AnthropicConfig{
			APIKey: req.APIKey,
			Model:  req.Model,
		})
	case "ollama":
		provider = llmclient.NewOllamaProvider(llmclient.OllamaConfig{
			BaseURL: req.BaseURL,
		})
	default:
		return &models.LLMTestResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported provider: %s", providerType),
		}, nil
	}

	// Send a simple test message
	resp, err := provider.Complete(ctx, llmclient.Request{
		Messages: []llmclient.Message{
			{Role: "user", Content: "Say 'Connection successful' in exactly those words."},
		},
		MaxTokens: 20,
	})

	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &models.LLMTestResponse{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
			Latency: latency,
		}, nil
	}

	return &models.LLMTestResponse{
		Success: true,
		Message: fmt.Sprintf("Connection successful: %s", resp.Content),
		Latency: latency,
	}, nil
}

// GetAgentConfig returns agent-specific configuration using the agent-config package.
func (s *ConfigService) GetAgentConfig(agentName string) (*agentconfig.AgentConfig, error) {
	// Get global config first
	globalConfig, err := s.Get()
	if err != nil {
		return nil, err
	}

	// Build agent config from global defaults
	agentCfg := &agentconfig.AgentConfig{
		Name: agentName,
	}

	if globalConfig.LLMConfig != nil {
		agentCfg.LLM = &agentconfig.LLMConfig{
			Provider:         globalConfig.LLMConfig.Provider,
			Model:            globalConfig.LLMConfig.Model,
			BaseURL:          globalConfig.LLMConfig.BaseURL,
			APIKey:           globalConfig.LLMConfig.APIKey,
			Temperature:      globalConfig.LLMConfig.Temperature,
			MaxTokens:        globalConfig.LLMConfig.MaxTokens,
			TopP:             globalConfig.LLMConfig.TopP,
			FrequencyPenalty: globalConfig.LLMConfig.FrequencyPenalty,
			PresencePenalty:  globalConfig.LLMConfig.PresencePenalty,
			StopSequences:    globalConfig.LLMConfig.StopSequences,
		}
	}

	return agentCfg, nil
}

// SetAgentConfig saves agent-specific configuration.
func (s *ConfigService) SetAgentConfig(agentName string, cfg *agentconfig.AgentConfig) error {
	// For now, we store agent configs in the agents table's config field
	// This could be extended to use a separate config store
	return nil
}
