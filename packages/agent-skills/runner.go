package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Runner is the skill execution engine that manages skill registration
// and executes skills with permission checking.
type Runner struct {
	mu      sync.RWMutex
	skills  map[string]Skill
	checker *PermissionChecker
	config  Config
}

// NewRunner creates a new skill runner with the given configuration.
func NewRunner(config Config) *Runner {
	return &Runner{
		skills:  make(map[string]Skill),
		checker: NewPermissionChecker(config),
		config:  config,
	}
}

// Register adds a skill to the runner.
// If a skill with the same name already exists, it will be replaced.
func (r *Runner) Register(skill Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[skill.Name()] = skill
}

// RegisterAll adds multiple skills to the runner at once.
func (r *Runner) RegisterAll(skills ...Skill) {
	for _, skill := range skills {
		r.Register(skill)
	}
}

// Unregister removes a skill from the runner.
func (r *Runner) Unregister(skillName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.skills, skillName)
}

// GetSkill retrieves a registered skill by name.
// Returns nil if the skill is not found.
func (r *Runner) GetSkill(name string) Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.skills[name]
}

// ListSkills returns the names of all registered skills.
func (r *Runner) ListSkills() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	return names
}

// Execute runs a skill action after checking permissions.
// It returns the result of the execution or an error if:
// - The skill is not found (SkillNotFoundError)
// - The action is not permitted (PermissionDeniedError)
// - The action is not supported by the skill (ExecutionError)
// - The execution itself fails (ExecutionError)
func (r *Runner) Execute(ctx context.Context, action Action) (Result, error) {
	// Look up the skill
	skill := r.GetSkill(action.Skill)
	if skill == nil {
		return Result{}, NewSkillNotFoundError(action.Skill)
	}

	// Check permissions
	if err := r.checker.CheckPermission(action.Skill, action.Action); err != nil {
		return Result{}, err
	}

	// Validate that the action is supported
	if !r.isActionSupported(skill, action.Action) {
		return Result{}, NewExecutionError(
			action.Skill,
			action.Action,
			fmt.Sprintf("skill '%s' does not support action '%s'. Supported actions: %s",
				action.Skill, action.Action, strings.Join(skill.Actions(), ", ")),
		)
	}

	// Execute the skill action
	result, err := skill.Execute(ctx, action.Action, action.Params)
	if err != nil {
		// Wrap non-agent errors in ExecutionError
		if _, ok := err.(*ExecutionError); !ok {
			return Result{}, NewExecutionError(action.Skill, action.Action, err.Error())
		}
		return Result{}, err
	}

	return result, nil
}

// isActionSupported checks if the skill supports the given action.
func (r *Runner) isActionSupported(skill Skill, action string) bool {
	for _, a := range skill.Actions() {
		if a == action {
			return true
		}
	}
	return false
}

// CheckPermission is a convenience method to check if an action would be permitted.
func (r *Runner) CheckPermission(skillName, action string) error {
	// First check if skill exists
	if r.GetSkill(skillName) == nil {
		return NewSkillNotFoundError(skillName)
	}
	return r.checker.CheckPermission(skillName, action)
}

// Config returns the current permission configuration.
func (r *Runner) Config() Config {
	return r.config
}
