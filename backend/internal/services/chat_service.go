// Package services provides business logic layer for the API.
package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	llmclient "github.com/formatho/agent-orchestrator/packages/llm-client"
	"github.com/google/uuid"
)

// ChatService handles chat operations.
type ChatService struct {
	db        *sql.DB
	agentSvc  *AgentService
	configSvc *ConfigService
	llmClient *llmclient.Client
}

// NewChatService creates a new chat service.
func NewChatService(db *sql.DB, agentSvc *AgentService, configSvc *ConfigService) *ChatService {
	return &ChatService{
		db:        db,
		agentSvc:  agentSvc,
		configSvc: configSvc,
	}
}

// SetLLMClient sets the LLM client for the service.
func (s *ChatService) SetLLMClient(client *llmclient.Client) {
	s.llmClient = client
}

// GetHistory returns the chat history for an agent (last 50 messages).
func (s *ChatService) GetHistory(agentID string) ([]*models.ChatMessage, error) {
	// Verify agent exists
	if _, err := s.agentSvc.Get(agentID); err != nil {
		return nil, err
	}

	query := `SELECT id, agent_id, role, content, created_at
		FROM chat_messages
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT 50`

	rows, err := s.db.Query(query, agentID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", cerr)
		}
	}()

	var messages []*models.ChatMessage
	for rows.Next() {
		msg := &models.ChatMessage{}
		err := rows.Scan(
			&msg.ID,
			&msg.AgentID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// SendMessage sends a message to an agent and returns the response.
func (s *ChatService) SendMessage(agentID string, message string) (*models.ChatResponse, error) {
	// Verify agent exists
	agent, err := s.agentSvc.Get(agentID)
	if err != nil {
		return nil, err
	}

	// Create user message
	userMsg, err := s.saveMessage(agentID, models.ChatRoleUser, message)
	if err != nil {
		return nil, err
	}

	// Get chat history for context
	history, err := s.GetHistory(agentID)
	if err != nil {
		return nil, err
	}

	// Build LLM messages
	llmMessages := s.buildLLMMessages(agent, history)

	// Call LLM
	assistantContent, err := s.callLLM(llmMessages)
	if err != nil {
		return nil, err
	}

	// Create assistant message
	assistantMsg, err := s.saveMessage(agentID, models.ChatRoleAssistant, assistantContent)
	if err != nil {
		return nil, err
	}

	return &models.ChatResponse{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
	}, nil
}

// saveMessage saves a chat message to the database.
func (s *ChatService) saveMessage(agentID string, role models.ChatRole, content string) (*models.ChatMessage, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	query := `INSERT INTO chat_messages (id, agent_id, role, content, created_at)
		VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, id, agentID, role, content, now)
	if err != nil {
		return nil, err
	}

	return &models.ChatMessage{
		ID:        id,
		AgentID:   agentID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}, nil
}

// buildLLMMessages builds the messages array for the LLM request.
func (s *ChatService) buildLLMMessages(agent *models.Agent, history []*models.ChatMessage) []llmclient.Message {
	messages := make([]llmclient.Message, 0, len(history)+1)

	// Add system prompt if available
	if agent.SystemPrompt != "" {
		messages = append(messages, llmclient.Message{
			Role:    "system",
			Content: agent.SystemPrompt,
		})
	}

	// Add conversation history
	for _, msg := range history {
		messages = append(messages, llmclient.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return messages
}

// callLLM calls the LLM with the given messages.
func (s *ChatService) callLLM(messages []llmclient.Message) (string, error) {
	if s.llmClient == nil {
		// Initialize LLM client from config
		config, err := s.configSvc.Get()
		if err != nil {
			return "", fmt.Errorf("failed to get config: %w", err)
		}

		if config.LLMConfig == nil {
			return "", fmt.Errorf("LLM not configured")
		}

		s.llmClient = llmclient.NewClient(llmclient.Config{
			Provider: llmclient.Provider(config.LLMConfig.Provider),
			Model:    config.LLMConfig.Model,
			APIKey:   config.LLMConfig.APIKey,
			BaseURL:  config.LLMConfig.BaseURL,
		})

		// Register providers based on configuration
		switch config.LLMConfig.Provider {
		case "openai":
			llmclient.RegisterOpenAI(s.llmClient, llmclient.OpenAIConfig{
				APIKey: config.LLMConfig.APIKey,
			})
		case "anthropic":
			llmclient.RegisterAnthropic(s.llmClient, llmclient.AnthropicConfig{
				APIKey: config.LLMConfig.APIKey,
				Model:  config.LLMConfig.Model,
			})
		case "ollama":
			llmclient.RegisterOllama(s.llmClient, llmclient.OllamaConfig{
				BaseURL: config.LLMConfig.BaseURL,
			})
		default:
			return "", fmt.Errorf("unsupported LLM provider: %s", config.LLMConfig.Provider)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := s.llmClient.Complete(ctx, llmclient.Request{
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	return resp.Content, nil
}

// ClearHistory clears all chat messages for an agent.
func (s *ChatService) ClearHistory(agentID string) error {
	// Verify agent exists
	if _, err := s.agentSvc.Get(agentID); err != nil {
		return err
	}

	_, err := s.db.Exec(`DELETE FROM chat_messages WHERE agent_id = ?`, agentID)
	return err
}
