package llm

import (
	"errors"
	"fmt"
)

// Error types for better error handling

// AuthenticationError indicates an invalid API key or authentication failure
type AuthenticationError struct {
	Provider string
	Message  string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed for %s: %s", e.Provider, e.Message)
}

// RateLimitError indicates rate limiting has been exceeded
type RateLimitError struct {
	Provider   string
	RetryAfter int // seconds to wait before retry
	Message    string
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limit exceeded for %s (retry after %ds): %s", e.Provider, e.RetryAfter, e.Message)
	}
	return fmt.Sprintf("rate limit exceeded for %s: %s", e.Provider, e.Message)
}

// ModelNotFoundError indicates the requested model doesn't exist
type ModelNotFoundError struct {
	Provider string
	Model    string
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("model '%s' not found for provider %s", e.Model, e.Provider)
}

// ContextLengthExceededError indicates the context length was exceeded
type ContextLengthExceededError struct {
	Provider string
	Message  string
}

func (e *ContextLengthExceededError) Error() string {
	return fmt.Sprintf("context length exceeded for %s: %s", e.Provider, e.Message)
}

// InvalidRequestError indicates the request was malformed
type InvalidRequestError struct {
	Provider string
	Field    string
	Message  string
}

func (e *InvalidRequestError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid request for %s (field: %s): %s", e.Provider, e.Field, e.Message)
	}
	return fmt.Sprintf("invalid request for %s: %s", e.Provider, e.Message)
}

// ProviderUnavailableError indicates the provider is temporarily unavailable
type ProviderUnavailableError struct {
	Provider string
	Status   int
	Message  string
}

func (e *ProviderUnavailableError) Error() string {
	return fmt.Sprintf("provider %s unavailable (status %d): %s", e.Provider, e.Status, e.Message)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var retryErr *RetryableError
	var rateLimit *RateLimitError
	var unavailable *ProviderUnavailableError

	if as(err, &retryErr) || as(err, &rateLimit) || as(err, &unavailable) {
		return true
	}
	return false
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	var authErr *AuthenticationError
	return as(err, &authErr)
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	var rateLimit *RateLimitError
	return as(err, &rateLimit)
}

// IsModelNotFoundError checks if an error is a model not found error
func IsModelNotFoundError(err error) bool {
	var modelErr *ModelNotFoundError
	return as(err, &modelErr)
}

// IsContextLengthError checks if an error is a context length error
func IsContextLengthError(err error) bool {
	var ctxErr *ContextLengthExceededError
	return as(err, &ctxErr)
}

// Helper for errors.As (Go 1.13+)
func as(err error, target interface{}) bool {
	return errors.As(err, target)
}
