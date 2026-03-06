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

func TestOpenAIProvider_Complete(t *testing.T) {
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
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(openAIResponse{
					ID:    "test-123",
					Model: "gpt-4o",
					Choices: []struct {
						Index        int     `json:"index"`
						Message      Message `json:"message"`
						Delta        Message `json:"delta,omitempty"`
						FinishReason string  `json:"finish_reason"`
					}{
						{
							Index:        0,
							Message:      Message{Role: "assistant", Content: "Hello, world!"},
							FinishReason: "stop",
						},
					},
					Usage: Usage{
						PromptTokens:     10,
						CompletionTokens: 5,
						TotalTokens:      15,
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
				ID:      "test-123",
				Model:   "gpt-4o",
				Content: "Hello, world!",
				Usage:   Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			},
		},
		{
			name: "completion with custom model",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req openAIRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "gpt-4", req.Model)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(openAIResponse{
					ID:    "test-456",
					Model: "gpt-4",
					Choices: []struct {
						Index        int     `json:"index"`
						Message      Message `json:"message"`
						Delta        Message `json:"delta,omitempty"`
						FinishReason string  `json:"finish_reason"`
					}{
						{
							Index:        0,
							Message:      Message{Role: "assistant", Content: "Response"},
							FinishReason: "stop",
						},
					},
				})
			},
			request: Request{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Test"},
				},
			},
			expectedError: false,
			expectedResp: &Response{
				ID:      "test-456",
				Model:   "gpt-4",
				Content: "Response",
			},
		},
		{
			name: "completion with all parameters",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req openAIRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, 0.7, req.Temperature)
				assert.Equal(t, 0.9, req.TopP)
				assert.Equal(t, 100, req.MaxTokens)
				assert.Equal(t, []string{"stop"}, req.Stop)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(openAIResponse{
					ID:    "test-789",
					Model: "gpt-4o",
					Choices: []struct {
						Index        int     `json:"index"`
						Message      Message `json:"message"`
						Delta        Message `json:"delta,omitempty"`
						FinishReason string  `json:"finish_reason"`
					}{
						{Index: 0, Message: Message{Role: "assistant", Content: "OK"}},
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
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"message": "Rate limit exceeded",
						"type":    "rate_limit_error",
						"code":    "rate_limit_exceeded",
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
			name: "rate limit error no retry",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"message": "Rate limit exceeded",
						"type":    "rate_limit_error",
						"code":    "rate_limit_exceeded",
					},
				})
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "max retries")
			},
		},
		{
			name: "authentication error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"message": "Invalid API key",
						"type":    "invalid_request_error",
						"code":    "invalid_api_key",
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
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"message": "Context length exceeded",
						"type":    "invalid_request_error",
						"code":    "context_length_exceeded",
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
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"message": "Model not found",
						"type":    "invalid_request_error",
						"code":    "model_not_found",
					},
				})
			},
			request: Request{
				Model:    "gpt-5",
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				assert.True(t, IsModelNotFoundError(err))
			},
		},
		{
			name: "server error with retry",
			serverHandler: func() func(w http.ResponseWriter, r *http.Request) {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts < 2 { // Only fail once (with MaxRetries=1, we'll retry once)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(openAIResponse{
						ID:    "test-retry",
						Model: "gpt-4o",
						Choices: []struct {
							Index        int     `json:"index"`
							Message      Message `json:"message"`
							Delta        Message `json:"delta,omitempty"`
							FinishReason string  `json:"finish_reason"`
						}{
							{Index: 0, Message: Message{Role: "assistant", Content: "Success"}},
						},
					})
				}
			}(),
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: false,
			expectedResp: &Response{
				Content: "Success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			provider := NewOpenAIProvider(OpenAIConfig{
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

func TestOpenAIProvider_Stream(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  func(w http.ResponseWriter, r *http.Request)
		request        Request
		expectedChunks []string
		expectedError  bool
	}{
		{
			name: "successful streaming",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)

				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				chunks := []string{"Hello", " ", "world", "!"}
				for _, chunk := range chunks {
					data := map[string]interface{}{
						"id":    "test-stream",
						"model": "gpt-4o",
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]string{
									"role":    "assistant",
									"content": chunk,
								},
							},
						},
					}
					jsonData, _ := json.Marshal(data)
					w.Write([]byte("data: " + string(jsonData) + "\n\n"))
					flusher.Flush()
					time.Sleep(10 * time.Millisecond)
				}

				w.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedChunks: []string{"Hello", " ", "world", "!"},
		},
		{
			name: "stream with finish reason",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				data := map[string]interface{}{
					"id":    "test-stream",
					"model": "gpt-4o",
					"choices": []map[string]interface{}{
						{
							"index": 0,
							"delta": map[string]string{
								"content": "Final",
							},
							"finish_reason": "stop",
						},
					},
				}
				jsonData, _ := json.Marshal(data)
				w.Write([]byte("data: " + string(jsonData) + "\n\n"))
				flusher.Flush()
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedChunks: []string{"Final"},
		},
		{
			name: "stream with error in response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				data := map[string]interface{}{
					"id":    "test-stream",
					"model": "gpt-4o",
					"error": map[string]string{
						"message": "Stream error",
						"type":    "error",
					},
					"choices": []map[string]interface{}{},
				}
				jsonData, _ := json.Marshal(data)
				w.Write([]byte("data: " + string(jsonData) + "\n\n"))
				flusher.Flush()
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedChunks: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			provider := NewOpenAIProvider(OpenAIConfig{
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

			if tt.expectedChunks == nil {
				assert.Nil(t, chunks)
			} else {
				assert.Equal(t, tt.expectedChunks, chunks)
			}
		})
	}
}

func TestOpenAIProvider_CountTokens(t *testing.T) {
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
		{
			name:     "code sample",
			text:     "func main() { fmt.Println(\"Hello\") }",
			minCount: 8,
			maxCount: 12,
		},
	}

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := provider.CountTokens(tt.text)
			assert.GreaterOrEqual(t, count, tt.minCount)
			assert.LessOrEqual(t, count, tt.maxCount)
		})
	}
}

func TestOpenAIProvider_ErrorHandling(t *testing.T) {
	t.Run("network error", func(t *testing.T) {
		provider := NewOpenAIProvider(OpenAIConfig{
			APIKey:  "test-key",
			BaseURL: "http://localhost:99999", // Invalid port
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

		provider := NewOpenAIProvider(OpenAIConfig{
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

		provider := NewOpenAIProvider(OpenAIConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})

	t.Run("no choices in response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(openAIResponse{
				ID:    "test",
				Model: "gpt-4o",
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					Delta        Message `json:"delta,omitempty"`
					FinishReason string  `json:"finish_reason"`
				}{},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no choices")
	})
}

func TestOpenAIProvider_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req openAIRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Should default to gpt-4o when no model specified
		assert.Equal(t, "gpt-4o", req.Model)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(openAIResponse{
			ID:    "test",
			Model: "gpt-4o",
			Choices: []struct {
				Index        int     `json:"index"`
				Message      Message `json:"message"`
				Delta        Message `json:"delta,omitempty"`
				FinishReason string  `json:"finish_reason"`
			}{
				{Index: 0, Message: Message{Role: "assistant", Content: "OK"}},
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
}

func TestOpenAIProvider_RetryLogic(t *testing.T) {
	t.Run("retry on 429", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(openAIResponse{
				ID:    "test",
				Model: "gpt-4o",
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					Delta        Message `json:"delta,omitempty"`
					FinishReason string  `json:"finish_reason"`
				}{
					{Index: 0, Message: Message{Role: "assistant", Content: "OK"}},
				},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
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
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"message": "Bad request",
					"code":    "invalid_request",
				},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
			APIKey:     "test-key",
			BaseURL:    server.URL,
			MaxRetries: 3,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Equal(t, 1, attempts) // Should not retry
	})
}

func TestOpenAIProvider_SpecialCases(t *testing.T) {
	t.Run("empty messages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(openAIResponse{
				ID:    "test",
				Model: "gpt-4o",
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					Delta        Message `json:"delta,omitempty"`
					FinishReason string  `json:"finish_reason"`
				}{
					{Index: 0, Message: Message{Role: "assistant", Content: "OK"}},
				},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("multiple messages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req openAIRequest
			json.NewDecoder(r.Body).Decode(&req)

			assert.Len(t, req.Messages, 4)
			assert.Equal(t, "system", req.Messages[0].Role)
			assert.Equal(t, "user", req.Messages[1].Role)
			assert.Equal(t, "assistant", req.Messages[2].Role)
			assert.Equal(t, "user", req.Messages[3].Role)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(openAIResponse{
				ID:    "test",
				Model: "gpt-4o",
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					Delta        Message `json:"delta,omitempty"`
					FinishReason string  `json:"finish_reason"`
				}{
					{Index: 0, Message: Message{Role: "assistant", Content: "OK"}},
				},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
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
			json.NewEncoder(w).Encode(openAIResponse{
				ID:    "test",
				Model: "gpt-4o",
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					Delta        Message `json:"delta,omitempty"`
					FinishReason string  `json:"finish_reason"`
				}{
					{Index: 0, Message: Message{Role: "assistant", Content: "你好世界 🌍"}},
				},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
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
			json.NewEncoder(w).Encode(openAIResponse{
				ID:    "test",
				Model: "gpt-4o",
				Choices: []struct {
					Index        int     `json:"index"`
					Message      Message `json:"message"`
					Delta        Message `json:"delta,omitempty"`
					FinishReason string  `json:"finish_reason"`
				}{
					{Index: 0, Message: Message{Role: "assistant", Content: largeContent}},
				},
			})
		}))
		defer server.Close()

		provider := NewOpenAIProvider(OpenAIConfig{
			APIKey:  "test-key",
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Generate large response"}},
		})

		assert.NoError(t, err)
		assert.True(t, len(resp.Content) > 50000)
	})
}
