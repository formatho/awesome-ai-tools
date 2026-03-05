# Contributing to go-llm-client

Thank you for your interest in contributing! 🎉

## Quick Start

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/agent-orchestrator.git`
3. Create a branch: `git checkout -b feature/my-feature`
4. Make changes and test: `go test ./packages/llm-client/...`
5. Commit: `git commit -m "feat: add amazing feature"`
6. Push: `git push origin feature/my-feature`
7. Open a Pull Request

## Development Setup

```bash
# Install Go 1.24+
brew install go  # macOS
# or download from https://golang.org/dl/

# Clone and enter the project
cd agent-orchestrator/packages/llm-client

# Run tests
go test -v

# Run tests with coverage
go test -cover

# Format code
go fmt ./...

# Run linter (optional but recommended)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
```

## Project Structure

```
packages/llm-client/
├── client.go           # Main client implementation
├── openai.go           # OpenAI provider
├── errors.go           # Error types
├── register.go         # Provider registration
├── client_test.go      # Tests
├── retry_test.go       # Retry tests
├── examples/           # Example programs
│   ├── basic-usage/
│   ├── streaming/
│   ├── error-handling/
│   ├── retry-configuration/
│   └── concurrent/
├── README.md
└── CHANGELOG.md
```

## Code Style

- Run `go fmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Add comments for exported functions
- Write tests for new functionality
- Keep functions small and focused

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation only
- `test:` Adding tests
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

Examples:
```
feat: add Anthropic provider
fix: handle empty responses correctly
docs: improve README examples
test: add edge case tests for retry logic
```

## Adding a New Provider

1. Create a new file: `providers/anthropic.go` (or `anthropic.go` in root)
2. Implement the `ProviderClient` interface:
   ```go
   type ProviderClient interface {
       Complete(ctx context.Context, req Request) (*Response, error)
       Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)
       CountTokens(text string) int
   }
   ```
3. Add registration function in `register.go`
4. Add tests in `anthropic_test.go`
5. Update README with provider info
6. Add example in `examples/`

## Testing

- Write unit tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Mock external API calls (use `httptest`)
- Test error cases, not just happy paths

Example test:
```go
func TestMyFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"basic", "hello", "Hello!", false},
        {"error", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test code
        })
    }
}
```

## Documentation

- Add godoc comments for all exported types and functions
- Update README.md with examples
- Add examples to `examples/` directory
- Update CHANGELOG.md

## Pull Request Checklist

- [ ] Code compiles: `go build`
- [ ] Tests pass: `go test -v`
- [ ] Code formatted: `go fmt`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages follow convention

## Questions?

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Be respectful and constructive

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🙏
