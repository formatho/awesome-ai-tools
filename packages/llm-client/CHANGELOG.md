# Changelog

All notable changes to go-llm-client will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release preparation

---

## [0.1.0] - 2026-03-05

### Added
- **Core Features**
  - Unified LLM client interface
  - OpenAI provider implementation
  - Chat completions (synchronous)
  - Streaming completions
  - Token counting (approximation)
  - Provider registration system

- **Retry Logic**
  - Automatic retry with exponential backoff
  - Retry on: 429, 500, 502, 503, 504 HTTP errors
  - Configurable retry count (default: 3)
  - Context cancellation support

- **Error Handling**
  - Typed errors for better error handling
  - `AuthenticationError` - Invalid API key
  - `RateLimitError` - Rate limit exceeded
  - `ModelNotFoundError` - Model not found
  - `ContextLengthExceededError` - Context too long
  - `InvalidRequestError` - Malformed request
  - `ProviderUnavailableError` - Service unavailable
  - Helper functions: `IsRetryable()`, `IsAuthenticationError()`, etc.

- **Documentation**
  - Comprehensive README with examples
  - Package-level godoc documentation
  - Example programs in `examples/` directory
  - Inline code documentation

- **Testing**
  - 10 unit tests covering core functionality
  - Retry behavior tests
  - Error handling tests
  - Context cancellation tests
  - Mock HTTP server for testing

### Examples
- `basic-usage` - Simple completion
- `streaming` - Real-time token streaming
- `error-handling` - Type-safe error handling
- `retry-configuration` - Custom retry behavior
- `concurrent` - Multiple concurrent requests

### Providers
- ✅ OpenAI (implemented)
- 📋 Anthropic (planned for v0.2.0)
- 📋 Ollama (planned for v0.2.0)
- 📋 Local/custom endpoints (planned)

---

## Future Plans

### [0.2.0] - TBD
- Anthropic provider
- Ollama provider (local models)
- Accurate token counting (tiktoken)
- More edge case tests
- Performance benchmarks

### [0.3.0] - TBD
- Function calling support
- Image inputs (vision models)
- Embedding support
- Async batch processing

---

[Unreleased]: https://github.com/formatho/agent-orchestrator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/formatho/agent-orchestrator/releases/tag/v0.1.0
