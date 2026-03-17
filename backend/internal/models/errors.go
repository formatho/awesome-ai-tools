package models

import (
	"errors"
	"fmt"
)

// AppError represents an application error with a code.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Common errors.
var (
	ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "Resource not found"}
	ErrBadRequest   = &AppError{Code: "BAD_REQUEST", Message: "Invalid request"}
	ErrInternal     = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized"}
	ErrConflict     = &AppError{Code: "CONFLICT", Message: "Resource conflict"}
)

// ErrValidation creates a validation error.
func ErrValidation(msg string) error {
	return &AppError{Code: "VALIDATION_ERROR", Message: msg}
}

// ErrNotFoundWithID creates a not found error with ID.
func ErrNotFoundWithID(resource, id string) error {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s with id '%s' not found", resource, id),
	}
}

// IsNotFoundError checks if error is a not found error.
func IsNotFoundError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == "NOT_FOUND"
	}
	return false
}

// IsConflictError checks if error is a conflict error.
func IsConflictError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == "CONFLICT"
	}
	return false
}

// IsBadRequestError checks if error is a bad request error.
func IsBadRequestError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == "BAD_REQUEST" || appErr.Code == "VALIDATION_ERROR"
	}
	return false
}

// IsUnauthorizedError checks if error is an unauthorized error.
func IsUnauthorizedError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == "UNAUTHORIZED"
	}
	return false
}

// NewAppError creates a new application error with code and message.
func NewAppError(code, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}
