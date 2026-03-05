# go-llm-client

**Unified Go client for LLM providers.** One interface for OpenAI, Anthropic, Ollama, and local models.

[![Go Reference](https://pkg.go.dev/badge/github.com/formatho/agent-orchestrator/packages/llm-client.svg)](https://pkg.go.dev/github.com/formatho/agent-orchestrator/packages/llm-client)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/formatho/agent-orchestrator/packages/llm-client)](https://goreportcard.com/report/github.com/formatho/agent-orchestrator/packages/llm-client)

---

## Why go-llm-client?

- **🔄 Multi-provider** — Same API for OpenAI, Anthropic, Ollama, and local models
- **⚡ Streaming support** — Real-time token streaming for responsive UX
- **🔁 Auto-retry** — Exponential backoff on transient failures (429, 500, etc.)
- **🎯 Type-safe errors** — Catch specific error types (rate limit, auth, context length)
- **📊 Token counting** — Built-in token estimation per provider
- **🔧 Per-request override** — Change provider/model on the fly

---

## Installation

```bash
go get github.com/formatho/agent-orchestrator/packages/llm-client
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "os"

    llm "github.com/formatho/agent-orchestrator/packages/llm-client"
)

func main() {
    // Create client
    client := llm.NewClient(llm.Config{
        Provider: llm.ProviderOpenAI,
        Model:    "gpt-4o",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })

    // Register OpenAI provider
    llm.RegisterOpenAI(client, llm.OpenAIConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    // Simple completion
    response, err := client.Simple(context.Background(), "Hello, world!")
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

---

## Examples

The [`examples/`](./examples/) directory contains complete working examples:

| Example | Description |
|---------|-------------|
| [basic-usage](./examples/basic-usage) | Simple completion |
| [streaming](./examples/streaming) | Real-time token streaming |
| [error-handling](./examples/error-handling) | Type-safe error handling |
| [retry-configuration](./examples/retry-configuration) | Custom retry behavior |
| [concurrent](./examples/concurrent) | Multiple concurrent requests |

Run any example:
```bash
cd examples/basic-usage
go run main.go
```

---

## Features

### ✅ Streaming

```go
stream, err := client.Stream(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "user", Content: "Tell me a story"},
    },
})
if err != nil {
    panic(err)
}

for chunk := range stream {
    fmt.Print(chunk.Delta.Content)
    if chunk.Finished {
        break
    }
}
```

### ✅ Error Handling

```go
_, err := client.Simple(ctx, "Hello")
if err != nil {
    if llm.IsAuthenticationError(err) {
        // Handle invalid API key
    } else if llm.IsRateLimitError(err) {
        // Handle rate limiting
    } else if llm.IsModelNotFoundError(err) {
        // Handle invalid model
    } else if llm.IsContextLengthError(err) {
        // Handle context length exceeded
    }
}
```

### ✅ Per-Request Override

```go
// Default is OpenAI, but use Anthropic for this request
response, err := client.Complete(ctx, llm.Request{
    Provider: llm.ProviderAnthropic,
    Model:    "claude-3-opus",
    Messages: []llm.Message{
        {Role: "user", Content: "Hello!"},
    },
})
```

### ✅ Retry Configuration

The client automatically retries on transient failures:

- **Retryable errors:** 429 (rate limit), 500, 502, 503, 504
- **Backoff:** Exponential (1s, 2s, 4s, ...)
- **Default:** 3 retries (4 total attempts)

```go
client := llm.NewClient(llm.Config{
    Provider:   llm.ProviderOpenAI,
    Model:      "gpt-4o",
    APIKey:     os.Getenv("OPENAI_API_KEY"),
    MaxRetries: 5, // 6 total attempts (1 initial + 5 retries)
})
```

---

## Providers

| Provider | Status | Notes |
|----------|--------|-------|
| OpenAI | ✅ | GPT-4, GPT-3.5, etc. |
| Anthropic | 🚧 | Claude models |
| Ollama | 🚧 | Local models |
| Local | 📋 | Custom endpoints |

---

## API Reference

### `NewClient(config Config) *Client`

Creates a new LLM client.

### `Complete(ctx, Request) (*Response, error)`

Sends a completion request.

### `Stream(ctx, Request) (<-chan StreamChunk, error)`

Streams completion tokens.

### `Simple(ctx, prompt) (string, error)`

Convenience method for simple prompts.

### `CountTokens(text) int`

Counts tokens in text.

---

## License

MIT

---

*Part of [Agent Orchestrator](https://github.com/formatho/agent-orchestrator)*
