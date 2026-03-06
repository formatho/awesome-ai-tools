package skills

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	agent "github.com/formatho/agent-orchestrator/packages/agent-skills"
)

// WebSkill provides HTTP operations for AI agents.
// Currently supports fetching URLs via HTTP GET requests.
type WebSkill struct {
	// Timeout is the maximum duration for HTTP requests.
	// Default is 30 seconds if not set.
	Timeout time.Duration

	// UserAgent is the User-Agent header sent with requests.
	UserAgent string
}

// NewWebSkill creates a new WebSkill with default settings.
func NewWebSkill() *WebSkill {
	return &WebSkill{
		Timeout:   30 * time.Second,
		UserAgent: "agent-skills/1.0",
	}
}

// Name returns the skill name.
func (s *WebSkill) Name() string {
	return "web"
}

// Actions returns the list of supported actions.
func (s *WebSkill) Actions() []string {
	return []string{"fetch"}
}

// Execute performs the specified web action.
func (s *WebSkill) Execute(ctx context.Context, action string, params map[string]any) (agent.Result, error) {
	switch action {
	case "fetch":
		return s.fetch(ctx, params)
	default:
		return agent.Result{}, agent.NewExecutionError("web", action,
			fmt.Sprintf("unknown action: %s", action))
	}
}

// fetch performs an HTTP GET request and returns the response body.
func (s *WebSkill) fetch(ctx context.Context, params map[string]any) (agent.Result, error) {
	// Get URL parameter
	urlRaw, ok := params["url"]
	if !ok {
		return agent.Result{}, agent.NewExecutionError("web", "fetch", "missing required parameter: url")
	}

	url, ok := urlRaw.(string)
	if !ok {
		return agent.Result{}, agent.NewExecutionError("web", "fetch", "parameter 'url' must be a string")
	}

	// Create HTTP client with timeout
	timeout := s.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return agent.Result{}, agent.NewExecutionError("web", "fetch",
			fmt.Sprintf("failed to create request: %s", err))
	}

	// Set User-Agent
	userAgent := s.UserAgent
	if userAgent == "" {
		userAgent = "agent-skills/1.0"
	}
	req.Header.Set("User-Agent", userAgent)

	// Add custom headers if provided
	if headers, ok := params["headers"].(map[string]any); ok {
		for key, value := range headers {
			if strVal, ok := value.(string); ok {
				req.Header.Set(key, strVal)
			}
		}
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return agent.Result{}, agent.NewExecutionError("web", "fetch",
			fmt.Sprintf("request failed: %s", err))
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return agent.Result{}, agent.NewExecutionError("web", "fetch",
			fmt.Sprintf("failed to read response: %s", err))
	}

	duration := time.Since(startTime)

	// Build result
	result := agent.Result{
		Success: resp.StatusCode >= 200 && resp.StatusCode < 300,
		Data:    string(body),
		Message: fmt.Sprintf("HTTP %d %s - %d bytes in %v",
			resp.StatusCode, resp.Status, len(body), duration),
		Metadata: map[string]any{
			"url":        url,
			"statusCode": resp.StatusCode,
			"status":     resp.Status,
			"size":       len(body),
			"duration":   duration.String(),
			"headers":    s.headersToMap(resp.Header),
		},
	}

	// Include error message for non-success status codes
	if !result.Success {
		result.Message = fmt.Sprintf("HTTP request failed with status %d: %s",
			resp.StatusCode, resp.Status)
	}

	return result, nil
}

// headersToMap converts http.Header to a map for JSON serialization.
func (s *WebSkill) headersToMap(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}
