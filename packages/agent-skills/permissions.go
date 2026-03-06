package agent

import (
	"fmt"
	"strings"
)

// Config holds the permission configuration for the Runner.
type Config struct {
	// Allowed is a list of skill:action patterns that are permitted.
	// Use "*" to allow all actions for a skill, or "*" as the skill to allow all.
	Allowed []string `json:"allowed"`

	// Denied is a list of skill:action patterns that are explicitly blocked.
	// Denied patterns always override allowed patterns.
	Denied []string `json:"denied"`
}

// PermissionChecker handles permission validation for skill execution.
type PermissionChecker struct {
	config Config
}

// NewPermissionChecker creates a new permission checker with the given configuration.
func NewPermissionChecker(config Config) *PermissionChecker {
	return &PermissionChecker{config: config}
}

// CheckPermission verifies if a skill action is allowed to execute.
// Returns nil if permitted, or PermissionDeniedError if not.
//
// Permission logic:
// 1. If explicitly denied, always block
// 2. If not in allowed list, block by default
// 3. Otherwise, allow
func (p *PermissionChecker) CheckPermission(skillName, action string) error {
	// Build the full identifier
	skillAction := fmt.Sprintf("%s:%s", skillName, action)
	skillAny := fmt.Sprintf("%s:*", skillName)

	// First check if explicitly denied
	for _, pattern := range p.config.Denied {
		if pattern == "*" ||
			pattern == skillAction ||
			pattern == skillAny ||
			strings.HasPrefix(pattern, "*:") {
			// Check if this global deny pattern matches
			if pattern == "*" || pattern == skillAny || pattern == skillAction {
				return NewPermissionDeniedError(skillName, action,
					fmt.Sprintf("action '%s' is explicitly denied by configuration", skillAction))
			}
			// Check pattern like "*:delete" which denies delete on all skills
			if strings.HasPrefix(pattern, "*:") {
				actionPattern := strings.TrimPrefix(pattern, "*:")
				if actionPattern == action || actionPattern == "*" {
					return NewPermissionDeniedError(skillName, action,
						fmt.Sprintf("action '%s' is globally denied", action))
				}
			}
		}
	}

	// If no allowed list specified, deny by default
	if len(p.config.Allowed) == 0 {
		return NewPermissionDeniedError(skillName, action,
			"no actions are allowed (empty allowed list)")
	}

	// Check if allowed
	for _, pattern := range p.config.Allowed {
		switch pattern {
		case "*":
			// Allow everything
			return nil
		case skillAny:
			// Allow all actions for this skill
			return nil
		case skillAction:
			// Allow this specific action
			return nil
		default:
			// Check pattern like "file:*" or "*:read"
			if strings.HasPrefix(pattern, "*:") {
				// Pattern like "*:read" - allow read on all skills
				actionPattern := strings.TrimPrefix(pattern, "*:")
				if actionPattern == action || actionPattern == "*" {
					return nil
				}
			}
			if strings.HasSuffix(pattern, ":*") {
				// Pattern like "file:*" - allow all actions on file skill
				skillPattern := strings.TrimSuffix(pattern, ":*")
				if skillPattern == skillName {
					return nil
				}
			}
		}
	}

	// Not found in allowed list
	return NewPermissionDeniedError(skillName, action,
		fmt.Sprintf("action '%s' is not in the allowed list", skillAction))
}

// IsAllowed is a convenience method that returns true if the action is permitted.
func (p *PermissionChecker) IsAllowed(skillName, action string) bool {
	return p.CheckPermission(skillName, action) == nil
}
