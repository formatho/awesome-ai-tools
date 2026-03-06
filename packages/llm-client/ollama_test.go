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

func TestOllamaProvider_Complete(t *testing.T) {
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
				assert.Equal(t, "/api/chat", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(ollamaResponse{
					Model: "llama3.2",
					Done:  true,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Hello from Ollama!",
					},
					EvalCount:       20,
					PromptEvalCount: 10,
				})
			},
			request: Request{
				Messages: []Message{
					{Role: "user", Content: "Say hello"},
				},
			},
			expectedError: false,
			expectedResp: &Response{
				Model:   "llama3.2",
				Content: "Hello from Ollama!",
				Usage:   Usage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
			},
		},
		{
			name: "completion with custom model",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req ollamaRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "codellama", req.Model)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(ollamaResponse{
					Model: "codellama",
					Done:  true,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Code response",
					},
				})
			},
			request: Request{
				Model: "codellama",
				Messages: []Message{
					{Role: "user", Content: "Test"},
				},
			},
			expectedError: false,
			expectedResp: &Response{
				Model:   "codellama",
				Content: "Code response",
			},
		},
		{
			name: "completion with all parameters",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				var req ollamaRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, 0.7, req.Options.Temperature)
				assert.Equal(t, 0.9, req.Options.TopP)
				assert.Equal(t, 100, req.Options.NumPredict)
				assert.Equal(t, []string{"stop"}, req.Options.Stop)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(ollamaResponse{
					Model: "llama3.2",
					Done:  true,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "OK",
					},
				})
			},
			request: Request{
				Messages:         []Message{{Role: "user", Content: "Test"}},
				Temperature:      0.7,
				TopP:             0.9,
				MaxTokens:        100,
				Stop:             []string{"stop"},
				FrequencyPenalty: 0.5,
				PresencePenalty:  0.3,
			},
			expectedError: false,
		},
		{
			name: "model not found error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ollamaResponse{
					Error: "model 'nonexistent' not found",
				})
			},
			request: Request{
				Model:    "nonexistent",
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				assert.True(t, IsModelNotFoundError(err))
			},
		},
		{
			name: "context length exceeded",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ollamaResponse{
					Error: "context length exceeded",
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
			name: "server error (500)",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ollamaResponse{
					Error: "Internal server error",
				})
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedError: true,
			validateError: func(t *testing.T, err error) {
				// After retries are exhausted, we get max retries error
				assert.Contains(t, err.Error(), "max retries")
			},
		},
		{
			name: "incomplete response (done=false)",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(ollamaResponse{
					Model: "llama3.2",
					Done:  false,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "Partial",
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
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ollamaResponse{
						Model: "llama3.2",
						Done:  true,
						Message: struct {
							Role    string `json:"role"`
							Content string `json:"content"`
						}{
							Role:    "assistant",
							Content: "Success",
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

			provider := NewOllamaProvider(OllamaConfig{
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
					if tt.expectedResp.Model != "" {
						assert.Equal(t, tt.expectedResp.Model, resp.Model)
					}
					if tt.expectedResp.Usage.TotalTokens > 0 {
						assert.Equal(t, tt.expectedResp.Usage, resp.Usage)
					}
				}
			}
		})
	}
}

func TestOllamaProvider_Stream(t *testing.T) {
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
				assert.True(t, strings.Contains(r.URL.Path, "/api/chat"))

				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				chunks := []string{"Hello", " ", "from", " ", "Ollama"}
				for i, chunk := range chunks {
					resp := ollamaGenerateResponse{
						Model:    "llama3.2",
						Response: chunk,
						Done:     i == len(chunks)-1,
					}
					jsonData, _ := json.Marshal(resp)
					w.Write(jsonData)
					w.Write([]byte("\n"))
					flusher.Flush()
					time.Sleep(10 * time.Millisecond)
				}
			},
			request: Request{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedChunks: []string{"Hello", " ", "from", " ", "Ollama"},
		},
		{
			name: "stream with custom model",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				var req ollamaRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, "mistral", req.Model)

				resp := ollamaGenerateResponse{
					Model:    "mistral",
					Response: "Test",
					Done:     true,
				}
				jsonData, _ := json.Marshal(resp)
				w.Write(jsonData)
				w.Write([]byte("\n"))
				flusher.Flush()
			},
			request: Request{
				Model:    "mistral",
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			expectedChunks: []string{"Test"},
		},
		{
			name: "stream with parameters",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected flusher")
				}

				var req ollamaRequest
				json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, 0.8, req.Options.Temperature)
				assert.Equal(t, 50, req.Options.NumPredict)

				resp := ollamaGenerateResponse{
					Model:    "llama3.2",
					Response: "Param test",
					Done:     true,
				}
				jsonData, _ := json.Marshal(resp)
				w.Write(jsonData)
				w.Write([]byte("\n"))
				flusher.Flush()
			},
			request: Request{
				Messages:    []Message{{Role: "user", Content: "Test"}},
				Temperature: 0.8,
				MaxTokens:   50,
			},
			expectedChunks: []string{"Param test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			provider := NewOllamaProvider(OllamaConfig{
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

func TestOllamaProvider_CountTokens(t *testing.T) {
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

	provider := NewOllamaProvider(OllamaConfig{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := provider.CountTokens(tt.text)
			assert.GreaterOrEqual(t, count, tt.minCount)
			assert.LessOrEqual(t, count, tt.maxCount)
		})
	}
}

func TestOllamaProvider_ErrorHandling(t *testing.T) {
	t.Run("network error", func(t *testing.T) {
		provider := NewOllamaProvider(OllamaConfig{
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

		provider := NewOllamaProvider(OllamaConfig{
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

		provider := NewOllamaProvider(OllamaConfig{
			BaseURL: server.URL,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})
}

func TestOllamaProvider_DefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Should default to llama3.2
		assert.Equal(t, "llama3.2", req.Model)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ollamaResponse{
			Model: "llama3.2",
			Done:  true,
			Message: struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    "assistant",
				Content: "OK",
			},
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
}

func TestOllamaProvider_DefaultBaseURL(t *testing.T) {
	provider := NewOllamaProvider(OllamaConfig{})

	// Should default to localhost:11434
	assert.Equal(t, "http://localhost:11434", provider.baseURL)
}

func TestOllamaProvider_RetryLogic(t *testing.T) {
	t.Run("retry on 500", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "OK",
				},
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
			BaseURL:    server.URL,
			MaxRetries: 3,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("retry on 503", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "OK",
				},
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
			BaseURL:    server.URL,
			MaxRetries: 3,
		})

		_, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("no retry on 404 (model not found)", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ollamaResponse{
				Error: "model not found",
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
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

func TestOllamaProvider_SpecialCases(t *testing.T) {
	t.Run("empty messages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "OK",
				},
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
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
			var req ollamaRequest
			json.NewDecoder(r.Body).Decode(&req)

			assert.Len(t, req.Messages, 4)
			assert.Equal(t, "system", req.Messages[0].Role)
			assert.Equal(t, "user", req.Messages[1].Role)
			assert.Equal(t, "assistant", req.Messages[2].Role)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "OK",
				},
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
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
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "你好世界 🌍",
				},
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
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
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: largeContent,
				},
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Generate large response"}},
		})

		assert.NoError(t, err)
		assert.True(t, len(resp.Content) > 50000)
	})

	t.Run("response with token usage", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ollamaResponse{
				Model: "llama3.2",
				Done:  true,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: "Test",
				},
				EvalCount:       25,
				PromptEvalCount: 15,
			})
		}))
		defer server.Close()

		provider := NewOllamaProvider(OllamaConfig{
			BaseURL: server.URL,
		})

		resp, err := provider.Complete(context.Background(), Request{
			Messages: []Message{{Role: "user", Content: "Test"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, 15, resp.Usage.PromptTokens)
		assert.Equal(t, 25, resp.Usage.CompletionTokens)
		assert.Equal(t, 40, resp.Usage.TotalTokens)
	})
}

func TestOllamaProvider_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL:    server.URL,
		MaxRetries: 2,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries")
	assert.Equal(t, 3, attempts) // Initial + 2 retries
}

func TestOllamaProvider_InvalidRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ollamaResponse{
			Error: "Invalid temperature value",
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid")
}

func TestOllamaProvider_NoAuthRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ollama doesn't require authentication headers
		assert.Empty(t, r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("x-api-key"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ollamaResponse{
			Model: "llama3.2",
			Done:  true,
			Message: struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    "assistant",
				Content: "OK",
			},
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
}

func TestOllamaProvider_ResponseID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ollamaResponse{
			Model: "llama3.2",
			Done:  true,
			Message: struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    "assistant",
				Content: "Test",
			},
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
	})

	resp, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Contains(t, resp.ID, "ollama-")
}

func TestOllamaProvider_FinishReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ollamaResponse{
			Model: "llama3.2",
			Done:  true,
			Message: struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    "assistant",
				Content: "Test",
			},
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
	})

	resp, err := provider.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Choices, 1)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
}
