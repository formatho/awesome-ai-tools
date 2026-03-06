package agentconfig

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string      // The field that failed validation
	Message string      // Human-readable error message
	Value   interface{} // The invalid value (optional)
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("%s: %s (got: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

// Error implements the error interface.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}

	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are any validation errors.
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Validator provides configuration validation.
type Validator struct {
	rules []ValidationRule
}

// ValidationRule is a function that validates a config and returns errors.
type ValidationRule func(config *Config) ValidationErrors

// NewValidator creates a new Validator with default rules.
func NewValidator() *Validator {
	v := &Validator{
		rules: make([]ValidationRule, 0),
	}

	// Add default validation rules
	v.AddRule(validateGlobalConfig)
	v.AddRule(validateAgentConfigs)
	v.AddRule(validateSkillPatterns)

	return v
}

// AddRule adds a custom validation rule.
func (v *Validator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate runs all validation rules and returns any errors.
func (v *Validator) Validate(config *Config) error {
	if config == nil {
		return &ValidationError{
			Field:   "config",
			Message: "configuration is nil",
		}
	}

	var allErrors ValidationErrors

	for _, rule := range v.rules {
		if errs := rule(config); len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}

	if allErrors.HasErrors() {
		return allErrors
	}

	return nil
}

// validateGlobalConfig validates the global configuration.
func validateGlobalConfig(config *Config) ValidationErrors {
	var errs ValidationErrors

	// Validate timeout
	if config.Global.Timeout < 0 {
		errs = append(errs, &ValidationError{
			Field:   "global.timeout",
			Message: "timeout must be non-negative",
			Value:   config.Global.Timeout,
		})
	}

	// Validate max retries
	if config.Global.MaxRetries < 0 {
		errs = append(errs, &ValidationError{
			Field:   "global.max_retries",
			Message: "max_retries must be non-negative",
			Value:   config.Global.MaxRetries,
		})
	}

	// Validate max concurrent tasks
	if config.Global.MaxConcurrentTasks < 0 {
		errs = append(errs, &ValidationError{
			Field:   "global.max_concurrent_tasks",
			Message: "max_concurrent_tasks must be non-negative",
			Value:   config.Global.MaxConcurrentTasks,
		})
	}

	// Validate global LLM config
	if config.Global.LLM != nil {
		errs = append(errs, validateLLMConfig(config.Global.LLM, "global.llm")...)
	}

	return errs
}

// validateAgentConfigs validates all agent configurations.
func validateAgentConfigs(config *Config) ValidationErrors {
	var errs ValidationErrors

	for name, agent := range config.Agents {
		if agent == nil {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("agents.%s", name),
				Message: "agent configuration is nil",
			})
			continue
		}

		prefix := fmt.Sprintf("agents.%s", name)

		// Validate agent name
		if name == "" {
			errs = append(errs, &ValidationError{
				Field:   prefix,
				Message: "agent name cannot be empty",
			})
		}

		// Validate agent name format (alphanumeric, dash, underscore)
		if !isValidName(name) {
			errs = append(errs, &ValidationError{
				Field:   prefix,
				Message: "agent name must contain only alphanumeric characters, dashes, and underscores",
				Value:   name,
			})
		}

		// Validate timeout
		if agent.Timeout != nil && *agent.Timeout < 0 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.timeout", prefix),
				Message: "timeout must be non-negative",
				Value:   *agent.Timeout,
			})
		}

		// Validate max retries
		if agent.MaxRetries != nil && *agent.MaxRetries < 0 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.max_retries", prefix),
				Message: "max_retries must be non-negative",
				Value:   *agent.MaxRetries,
			})
		}

		// Validate max concurrent tasks
		if agent.MaxConcurrentTasks != nil && *agent.MaxConcurrentTasks < 0 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.max_concurrent_tasks", prefix),
				Message: "max_concurrent_tasks must be non-negative",
				Value:   *agent.MaxConcurrentTasks,
			})
		}

		// Validate LLM config
		if agent.LLM != nil {
			errs = append(errs, validateLLMConfig(agent.LLM, fmt.Sprintf("%s.llm", prefix))...)
		}
	}

	return errs
}

// validateLLMConfig validates LLM configuration.
func validateLLMConfig(llm *LLMConfig, prefix string) ValidationErrors {
	var errs ValidationErrors

	if llm == nil {
		return errs
	}

	// Validate temperature range (0-2)
	if llm.Temperature != nil {
		if *llm.Temperature < 0 || *llm.Temperature > 2 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.temperature", prefix),
				Message: "temperature must be between 0 and 2",
				Value:   *llm.Temperature,
			})
		}
	}

	// Validate max_tokens
	if llm.MaxTokens != nil {
		if *llm.MaxTokens < 0 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.max_tokens", prefix),
				Message: "max_tokens must be non-negative",
				Value:   *llm.MaxTokens,
			})
		}
	}

	// Validate top_p range (0-1)
	if llm.TopP != nil {
		if *llm.TopP < 0 || *llm.TopP > 1 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.top_p", prefix),
				Message: "top_p must be between 0 and 1",
				Value:   *llm.TopP,
			})
		}
	}

	// Validate frequency_penalty range (0-2)
	if llm.FrequencyPenalty != nil {
		if *llm.FrequencyPenalty < 0 || *llm.FrequencyPenalty > 2 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.frequency_penalty", prefix),
				Message: "frequency_penalty must be between 0 and 2",
				Value:   *llm.FrequencyPenalty,
			})
		}
	}

	// Validate presence_penalty range (0-2)
	if llm.PresencePenalty != nil {
		if *llm.PresencePenalty < 0 || *llm.PresencePenalty > 2 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("%s.presence_penalty", prefix),
				Message: "presence_penalty must be between 0 and 2",
				Value:   *llm.PresencePenalty,
			})
		}
	}

	return errs
}

// validateSkillPatterns validates skill patterns in allowed/denied lists.
func validateSkillPatterns(config *Config) ValidationErrors {
	var errs ValidationErrors

	// Skill pattern: namespace.action or namespace.*
	skillPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_*-]+)*$`)

	for name, agent := range config.Agents {
		if agent == nil || agent.Skills == nil {
			continue
		}

		prefix := fmt.Sprintf("agents.%s.skills", name)

		// Validate allowed skills
		for i, skill := range agent.Skills.Allowed {
			if skill == "" {
				errs = append(errs, &ValidationError{
					Field:   fmt.Sprintf("%s.allowed[%d]", prefix, i),
					Message: "skill pattern cannot be empty",
				})
			} else if !skillPattern.MatchString(skill) {
				errs = append(errs, &ValidationError{
					Field:   fmt.Sprintf("%s.allowed[%d]", prefix, i),
					Message: "invalid skill pattern format (expected: namespace.action or namespace.*)",
					Value:   skill,
				})
			}
		}

		// Validate denied skills
		for i, skill := range agent.Skills.Denied {
			if skill == "" {
				errs = append(errs, &ValidationError{
					Field:   fmt.Sprintf("%s.denied[%d]", prefix, i),
					Message: "skill pattern cannot be empty",
				})
			} else if !skillPattern.MatchString(skill) {
				errs = append(errs, &ValidationError{
					Field:   fmt.Sprintf("%s.denied[%d]", prefix, i),
					Message: "invalid skill pattern format (expected: namespace.action or namespace.*)",
					Value:   skill,
				})
			}
		}

		// Check for conflicts (same skill in both allowed and denied)
		if len(agent.Skills.Allowed) > 0 && len(agent.Skills.Denied) > 0 {
			allowedSet := make(map[string]bool)
			for _, s := range agent.Skills.Allowed {
				allowedSet[s] = true
			}

			for _, denied := range agent.Skills.Denied {
				if allowedSet[denied] {
					errs = append(errs, &ValidationError{
						Field:   prefix,
						Message: fmt.Sprintf("skill '%s' appears in both allowed and denied lists", denied),
						Value:   denied,
					})
				}
			}
		}
	}

	return errs
}

// isValidName checks if a name is valid (alphanumeric, dash, underscore).
func isValidName(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if !isAlphaNumeric(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
