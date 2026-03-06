package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAnthropicProvider_Complete(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		request       Request
		expectedError bool
		expectedResp  *Response
		validateError func(t *testing.T, err error)
	}{
		{
			name: "successful completion",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.NotEmpty(t, r.Header.Get("x-api-key"))
				assert.NotEmpty(t, r.Header.Get("anthropic-version"))

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(anthropicResponse{
					ID:    "msg-123",
					Type:  "message",
					Role:  "assistant",
					Model: "claude-3-5-sonnet-20241022",
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{
						{Type: "text", Text: "Hello from Claude!"},
					},
					StopReason: "end_turn",
					Usage: struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					}{
						InputTokens:  10,
						OutputTokens: 5,
					},
				})
			},
			request: Request{
				Messages: []Message{
					{Role: "user", Content: "Say hello"},
				},
			},
			expectedError: false,
			expectedResp: &Response{
				ID:      "msg-123",
				Model:   "claude-3-5-sonnet-20241022",
				Content: "Hello from Claude!",
				Usage:   Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			},
		},
		{
			name: "completion with custom model",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req anthropicRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "claude-3-opus-20240229", req.Model)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(anthropicResponse{
					ID:    "msg-456",
					Model: "claude-3-opus-20240229",
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{
						{Type: "text", Text: "Response"},
					},
				})
			},
			request: Request{
				Model: "claude-3-opus-20240229",
				Messages: []Message{
					{Role: "user", Content: "Test"},
				},
			},
			expectedError: false,
			expectedResp: &Response{
				ID:      "msg-456",
				Model:   "claude-3-opus-20240229",
				Content: "Response",
			},
		},
		{
			name: "completion with system prompt",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req anthropicRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "You are a helpful assistant", req.System)
				assert.Len(t, req.Messages, 1) // System message should be extracted

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(anthropicResponse{
					ID:    "msg-789",
					Model: "claude-3-5-sonnet-20241022",
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{
						{Type: "text", Text: "Understood"},
					},
				})
			},
			request: Request{
				Messages: []Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello"},
				},
			},
			expectedError: false,
		},
		{
			name: "completion with all parameters",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req anthropicRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, 0.7, req.Temperature)
				assert.Equal(t, 0.9, req.TopP)
				assert.Equal(t, 100, req.MaxTokens)
				assert.Equal(t, []string{"stop"}, req.Stop)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(anthropicResponse{
					ID:    "msg-params",
					Model: "claude-3-5-sonnet-20241022",
					Content: []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{
						{Type: "text", Text: "OK"},
					},
				})
			},
			request: Request{
				Messages:    []Message{{Role: "user", Content: "Test"}},
				Temperature: 0.7,
				TopP:        0.9,
				MaxTokens:   100,
				Stop:        []string{"stop"},
			},
			expectedError: false,
		},
		{
			name: "rate limit error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(anthropicResponse{
					Error: &struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					}{
						Type:    "rate_limit_error",
						Message: "Rate limit exceeded",
					},
				})
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				// After retries are exhausted, we should get a max retries error
				assert.Contains(t, err.Error(), "max retries")
			},
		},
		{
			name: "authentication error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(anthropicResponse{
					Error: &struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					}{
						Type:    "authentication_error",
						Message: "Invalid API key",
					},
				})
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				assert.True(t, IsAuthenticationError(err))
			},
		},
		{
			name: "context length exceeded",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(anthropicResponse{
					Error: &struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					}{
						Type:    "invalid_request_error",
						Message: "context length exceeded: max 200000 tokens",
					},
				})
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Very long message"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				assert.True(t, IsContextLengthError(err))
			},
		},
		{
			name: "model not found",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(anthropicResponse{
					Error: &struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					}{
						Type:    "not_found_error",
						Message: "Model not found: claude-4",
					},
				})
			},
			request: Request{
				Model:    "claude-4",
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				assert.True(t, IsModelNotFoundError(err))
			},
		},
		{
			name: "overloaded error (529)",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(529)
				json.NewEncoder(w).Encode(anthropicResponse{
					Error: &struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					}{
						Type:    "overloaded_error",
						Message: "Overloaded",
					},
				})
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
		},
		{
			name: "server error with retry",
			serverHandler: func() func(w http.ResponseWriter, r *http.Request) {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts < 2 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(anthropicResponse{
						ID:    "msg-retry",
						Model: "claude-3-5-sonnet-20241022",
						Content: []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						}{
							{Type: "text", Text: "Success"},
						},
					})
				}
			}(),
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			provider := NewAnthropicProvider(AnthropicConfig{
				APIKey:     "test-key",
				BaseURL:    server.URL,
				MaxRetries: 1, // Reduce retries for faster tests
			})

			resp, err := provider.Complete(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.validateError != nil {
					tt.validateError(t, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectedResp != nil {
					assert.Equal(t, tt.expectedResp.Content, resp.Content)
					if tt.expectedResp.ID != "" {
						assert.Equal(t, tt.expectedResp.ID, resp.ID)
					}
					if tt.expectedResp.Model != "" {
						assert.Equal(t, tt.expectedResp.Model, resp.Model)
					}
				}
			}
		})
	}
}

func TestAnthropicProvider_Stream(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		request        Request
		expectedChunks []string
	}{
		{
			name: "successful streaming",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				chunks := []string{"Hello", " ", "from", " ", "Claude"}
				for _, chunk := range chunks {
					event := struct {
						Type  string `json:"type"`
						Delta struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"delta"`
					}{
						Type: "content_block_delta",
						Delta: struct {
							Type string `json:"type"`
							Text string `json:"text"`
						}{
							Type: "text_delta",
							Text: chunk,
						},
					}
					jsonData, _ := json.Marshal(event)
					w.Write([]byte("data: " + string(jsonData) + "\n\n"))
					flusher.Flush()
					time.Sleep(10 * time.Millisecond)
				}

				stopEvent := struct {
					Type string `json:"type"`
				}{
					Type: "message_stop",
				}
				jsonData, _ := json.Marshal(stopEvent)
				w.Write([]byte("data: " + string(jsonData) + "\n\n"))
				flusher.Flush()
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedChunks: []string{"Hello", " ", "from", " ", "Claude"},
		},
		{
			name: "stream with system prompt",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				// Verify system prompt is in the request
				var req anthropicRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "Be helpful", req.System)

				event := struct {
					Type  string `json:"type"`
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}{
					Type: "content_block_delta",
					Delta: struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{
						Type: "text_delta",
						Text: "Response",
					},
				}
				jsonData, _ := json.Marshal(event)
				w.Write([]byte("data: " + string(jsonData) + "\n\n"))
				flusher.Flush()

				stopEvent := struct {
					Type string `json:"type"`
				}{
					Type: "message_stop",
				}
				jsonData, _ = json.Marshal(stopEvent)
				w.Write([]byte("data: " + string(jsonData) + "\n\n"))
				flusher.Flush()
			},
			request: Request{
				Messages: []Message{
					{Role: "system", Content: "Be helpful"},
					{Role: "user", Content: "Test"},
				},
			},
			expectedChunks: []string{"Response"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			provider := NewAnthropicProvider(AnthropicConfig{
				APIKey:  "test-key",
				BaseURL: server.URL,
				Debug:   false,
			})

			stream, err := provider.Stream(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.NotNil(t, stream)

			var chunks []string
			for chunk := range stream {
				if chunk.Delta.Content != "" {
					chunks = append(chunks, chunk.Delta.Content)
				}
			}

			assert.Equal(t, tt.expectedChunks, chunks)
		})
	}
}

func TestAnthropicProvider_CountTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int
		maxCount int
	}{
		{
			name:     "empty string",
			text:     "",
			minCount: 0,
			maxCount: 1,
		},
		{
			name:     "short text",
			text:     "Hello world",
			minCount: 2,
			maxCount: 3,
		},
		{
			name:     "longer text",
			text:     "This is a longer piece of text that should have more tokens",
			minCount: 12,
			maxCount: 20,
		},
	}

	provider := NewAnthropicProvider(AnthropicConfig{APIKey: "test"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := provider.CountTokens(tt.text)
			assert.GreaterOrEqual(t, count, tt.minCount)
			assert.LessOrEqual(t, count, tt.maxCount)
		})
	}
}

func TestAnthropicProvider_ErrorHandling(t *testing.T) {
	t.Run("network error", func(t *testing.T) {
		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: "http://localhost:99999",
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := provider.Complete(ctx, Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	})

	t.Run("malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json {"))
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})

	t.Run("no content in response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no content")
	})
}

func TestAnthropicProvider_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Should default to claude-3-5-sonnet-20241022
		assert.Equal(t, "claude-3-5-sonnet-20241022", req.Model)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(anthropicResponse{
			ID:    "test",
			Model: "claude-3-5-sonnet-20241022",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: "OK"},
			},
		})
	}))
	defer server.Close()

	provider := NewAnthropicProvider(AnthropicConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
}

func TestAnthropicProvider_DefaultMaxTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req anthropicRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Should default to 4096 when not specified
		assert.Equal(t, 4096, req.MaxTokens)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(anthropicResponse{
			ID:    "test",
			Model: "claude-3-5-sonnet-20241022",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: "OK"},
			},
		})
	}))
	defer server.Close()

	provider := NewAnthropicProvider(AnthropicConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
}

func TestAnthropicProvider_RetryLogic(t *testing.T) {
	t.Run("retry on 429", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "OK"},
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:     "test-key",
			BaseURL:    server.URL,
			MaxRetries: 3,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("retry on 529 (overloaded)", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(529)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "OK"},
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:     "test-key",
			BaseURL:    server.URL,
			MaxRetries: 3,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("no retry on 400", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(anthropicResponse{
				Error: &struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "invalid_request_error",
					Message: "Bad request",
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:     "test-key",
			BaseURL:    server.URL,
			MaxRetries: 3,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Equal(t, 1, attempts)
	})
}

func TestAnthropicProvider_SpecialCases(t *testing.T) {
	t.Run("empty messages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "OK"},
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("multiple messages with system", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req anthropicRequest
			json.NewDecoder(r.Body).Decode(&req)

			// System should be extracted and not in messages
			assert.Equal(t, "You are helpful", req.System)
			assert.Len(t, req.Messages, 3) // 2 user messages + 1 assistant message

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "OK"},
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hi"},
				{Role: "assistant", Content: "Hello!"},
				{Role: "user", Content: "How are you?"},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("unicode content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: "你好世界 🌍"},
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test unicode"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, "你好世界 🌍", resp.Content)
	})

	t.Run("large response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			largeContent := strings.Repeat("large ", 10000)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(anthropicResponse{
				ID:    "test",
				Model: "claude-3-5-sonnet-20241022",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "text", Text: largeContent},
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Generate large response"}},
		})

		assert.NoError(t, err)
		assert.True(t, len(resp.Content) > 50000)
	})

	t.Run("invalid request error with field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(anthropicResponse{
				Error: &struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "invalid_request_error",
					Message: "Invalid temperature value",
				},
			})
		}))
		defer server.Close()

		provider := NewAnthropicProvider(AnthropicConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid")
	})
}

func TestAnthropicProvider_AuthenticationHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify required headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(anthropicResponse{
			ID:    "test",
			Model: "claude-3-5-sonnet-20241022",
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: "OK"},
			},
		})
	}))
	defer server.Close()

	provider := NewAnthropicProvider(AnthropicConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
}
