package agent

import (
	"fmt"
)

// SkillNotFoundError is returned when attempting to execute
// a skill that has not been registered.
type SkillNotFoundError struct {
	Skill string
}

// NewSkillNotFoundError creates a new SkillNotFoundError.
func NewSkillNotFoundError(skill string) *SkillNotFoundError {
	return &SkillNotFoundError{Skill: skill}
}

// Error implements the error interface.
func (e *SkillNotFoundError) Error() string {
	return fmt.Sprintf("skill not found: %s", e.Skill)
}

// PermissionDeniedError is returned when an action is not permitted
// by the permission configuration.
type PermissionDeniedError struct {
	Skill  string
	Action string
	Reason string
}

// NewPermissionDeniedError creates a new PermissionDeniedError.
func NewPermissionDeniedError(skill, action, reason string) *PermissionDeniedError {
	return &PermissionDeniedError{
		Skill:  skill,
		Action: action,
		Reason: reason,
	}
}

// Error implements the error interface.
func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("permission denied: skill=%s action=%s reason=%s",
		e.Skill, e.Action, e.Reason)
}

// ExecutionError is returned when a skill execution fails.
type ExecutionError struct {
	Skill  string
	Action string
	Err    string
}

// NewExecutionError creates a new ExecutionError.
func NewExecutionError(skill, action, err string) *ExecutionError {
	return &ExecutionError{
		Skill:  skill,
		Action: action,
		Err:    err,
	}
}

// Error implements the error interface.
func (e *ExecutionError) Error() string {
	return fmt.Sprintf("execution error: skill=%s action=%s error=%s",
		e.Skill, e.Action, e.Err)
}

// Unwrap returns the underlying error message for use with errors.Is/As.
func (e *ExecutionError) Unwrap() error {
	return fmt.Errorf(e.Err)
}
