package agent

import (
	"context"
	"errors"
	"testing"
)

// MockSkill is a test skill for testing the runner
type MockSkill struct {
	name        string
	actions     []string
	executeFunc func(ctx context.Context, action string, params map[string]any) (Result, error)
}

func (m *MockSkill) Name() string {
	return m.name
}

func (m *MockSkill) Actions() []string {
	return m.actions
}

func (m *MockSkill) Execute(ctx context.Context, action string, params map[string]any) (Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, action, params)
	}
	return Result{Success: true, Message: "ok"}, nil
}

func TestNewRunner(t *testing.T) {
	config := Config{
		Allowed: []string{"*"},
	}
	runner := NewRunner(config)

	if runner == nil {
		t.Fatal("expected runner to be created")
	}

	if runner.skills == nil {
		t.Error("expected skills map to be initialized")
	}
}

func TestRegister(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	skill := &MockSkill{name: "test", actions: []string{"run"}}

	runner.Register(skill)

	if len(runner.skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(runner.skills))
	}

	if runner.GetSkill("test") == nil {
		t.Error("expected to find registered skill")
	}
}

func TestRegisterAll(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	skills := []Skill{
		&MockSkill{name: "skill1", actions: []string{"run"}},
		&MockSkill{name: "skill2", actions: []string{"run"}},
		&MockSkill{name: "skill3", actions: []string{"run"}},
	}

	runner.RegisterAll(skills...)

	if len(runner.skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(runner.skills))
	}
}

func TestUnregister(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	skill := &MockSkill{name: "test", actions: []string{"run"}}

	runner.Register(skill)
	runner.Unregister("test")

	if runner.GetSkill("test") != nil {
		t.Error("expected skill to be unregistered")
	}
}

func TestExecute_SkillNotFound(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})

	_, err := runner.Execute(context.Background(), Action{
		Skill:  "nonexistent",
		Action: "run",
	})

	if err == nil {
		t.Fatal("expected error for nonexistent skill")
	}

	var skillErr *SkillNotFoundError
	if !errors.As(err, &skillErr) {
		t.Errorf("expected SkillNotFoundError, got %T", err)
	}

	if skillErr.Skill != "nonexistent" {
		t.Errorf("expected skill name 'nonexistent', got %s", skillErr.Skill)
	}
}

func TestExecute_PermissionDenied(t *testing.T) {
	runner := NewRunner(Config{
		Allowed: []string{"allowed:run"},
	})
	runner.Register(&MockSkill{name: "denied", actions: []string{"run"}})

	_, err := runner.Execute(context.Background(), Action{
		Skill:  "denied",
		Action: "run",
	})

	if err == nil {
		t.Fatal("expected permission denied error")
	}

	var permErr *PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Errorf("expected PermissionDeniedError, got %T", err)
	}
}

func TestExecute_UnsupportedAction(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	runner.Register(&MockSkill{name: "test", actions: []string{"run", "stop"}})

	_, err := runner.Execute(context.Background(), Action{
		Skill:  "test",
		Action: "invalid",
	})

	if err == nil {
		t.Fatal("expected error for unsupported action")
	}

	var execErr *ExecutionError
	if !errors.As(err, &execErr) {
		t.Errorf("expected ExecutionError, got %T", err)
	}
}

func TestExecute_Success(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	runner.Register(&MockSkill{
		name:    "test",
		actions: []string{"run"},
		executeFunc: func(ctx context.Context, action string, params map[string]any) (Result, error) {
			return Result{Success: true, Data: "test data", Message: "executed"}, nil
		},
	})

	result, err := runner.Execute(context.Background(), Action{
		Skill:  "test",
		Action: "run",
		Params: map[string]any{"key": "value"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	if result.Data != "test data" {
		t.Errorf("expected 'test data', got %v", result.Data)
	}
}

// Permission Tests

func TestPermissionChecker_AllowAll(t *testing.T) {
	checker := NewPermissionChecker(Config{Allowed: []string{"*"}})

	if err := checker.CheckPermission("file", "read"); err != nil {
		t.Errorf("expected permission granted, got error: %v", err)
	}
}

func TestPermissionChecker_DenyAll(t *testing.T) {
	checker := NewPermissionChecker(Config{Allowed: []string{}, Denied: []string{"*"}})

	err := checker.CheckPermission("file", "read")
	if err == nil {
		t.Fatal("expected permission denied")
	}
}

func TestPermissionChecker_SpecificAllow(t *testing.T) {
	checker := NewPermissionChecker(Config{Allowed: []string{"file:read", "web:fetch"}})

	// Allowed actions
	if err := checker.CheckPermission("file", "read"); err != nil {
		t.Errorf("expected file:read to be allowed, got: %v", err)
	}
	if err := checker.CheckPermission("web", "fetch"); err != nil {
		t.Errorf("expected web:fetch to be allowed, got: %v", err)
	}

	// Denied actions
	if err := checker.CheckPermission("file", "write"); err == nil {
		t.Error("expected file:write to be denied")
	}
	if err := checker.CheckPermission("shell", "run"); err == nil {
		t.Error("expected shell:run to be denied")
	}
}

func TestPermissionChecker_SkillWildcard(t *testing.T) {
	checker := NewPermissionChecker(Config{Allowed: []string{"file:*"}})

	// All file actions should be allowed
	if err := checker.CheckPermission("file", "read"); err != nil {
		t.Errorf("expected file:read to be allowed, got: %v", err)
	}
	if err := checker.CheckPermission("file", "write"); err != nil {
		t.Errorf("expected file:write to be allowed, got: %v", err)
	}

	// Other skills should be denied
	if err := checker.CheckPermission("web", "fetch"); err == nil {
		t.Error("expected web:fetch to be denied")
	}
}

func TestPermissionChecker_DenyOverridesAllow(t *testing.T) {
	checker := NewPermissionChecker(Config{
		Allowed: []string{"*"},
		Denied:  []string{"shell:*"},
	})

	// Shell should be denied even though * is allowed
	if err := checker.CheckPermission("shell", "run"); err == nil {
		t.Error("expected shell:run to be denied (explicit deny)")
	}

	// Other skills should be allowed
	if err := checker.CheckPermission("file", "read"); err != nil {
		t.Errorf("expected file:read to be allowed, got: %v", err)
	}
}

func TestPermissionChecker_SpecificDeny(t *testing.T) {
	checker := NewPermissionChecker(Config{
		Allowed: []string{"file:*"},
		Denied:  []string{"file:delete"},
	})

	// Read and write should work
	if err := checker.CheckPermission("file", "read"); err != nil {
		t.Errorf("expected file:read to be allowed, got: %v", err)
	}
	if err := checker.CheckPermission("file", "write"); err != nil {
		t.Errorf("expected file:write to be allowed, got: %v", err)
	}

	// Delete should be denied
	if err := checker.CheckPermission("file", "delete"); err == nil {
		t.Error("expected file:delete to be denied")
	}
}

func TestPermissionChecker_EmptyAllowedList(t *testing.T) {
	checker := NewPermissionChecker(Config{Allowed: []string{}})

	// Everything should be denied
	if err := checker.CheckPermission("file", "read"); err == nil {
		t.Error("expected permission denied with empty allowed list")
	}
}

// Error Type Tests

func TestSkillNotFoundError(t *testing.T) {
	err := NewSkillNotFoundError("test-skill")

	if err.Error() != "skill not found: test-skill" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestPermissionDeniedError(t *testing.T) {
	err := NewPermissionDeniedError("file", "write", "not in allowed list")

	var permErr *PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Error("expected to be PermissionDeniedError")
	}

	expected := "permission denied: skill=file action=write reason=not in allowed list"
	if err.Error() != expected {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestExecutionError(t *testing.T) {
	err := NewExecutionError("shell", "run", "command failed")

	var execErr *ExecutionError
	if !errors.As(err, &execErr) {
		t.Error("expected to be ExecutionError")
	}

	expected := "execution error: skill=shell action=run error=command failed"
	if err.Error() != expected {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// Context Cancellation Test

func TestExecute_ContextCancellation(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	runner.Register(&MockSkill{
		name:    "test",
		actions: []string{"run"},
		executeFunc: func(ctx context.Context, action string, params map[string]any) (Result, error) {
			select {
			case <-ctx.Done():
				return Result{}, ctx.Err()
			default:
				return Result{Success: true}, nil
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := runner.Execute(ctx, Action{
		Skill:  "test",
		Action: "run",
	})

	if err == nil {
		t.Error("expected error due to context cancellation")
	}
}

// Result Tests

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestResult_JSON(t *testing.T) {
	result := Result{
		Success: true,
		Data:    "test data",
		Message: "operation complete",
		Metadata: map[string]any{
			"key": "value",
		},
	}

	json := result.JSON()

	if json == "" {
		t.Error("expected non-empty JSON output")
	}

	if !contains(json, `"success": true`) && !contains(json, `"success":true`) {
		t.Errorf("JSON should contain success field: %s", json)
	}
}

func TestRunner_ListSkills(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"*"}})
	runner.Register(&MockSkill{name: "file", actions: []string{"read"}})
	runner.Register(&MockSkill{name: "web", actions: []string{"fetch"}})

	skills := runner.ListSkills()

	if len(skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(skills))
	}
}

func TestRunner_CheckPermission(t *testing.T) {
	runner := NewRunner(Config{Allowed: []string{"file:*"}})
	runner.Register(&MockSkill{name: "file", actions: []string{"read"}})
	runner.Register(&MockSkill{name: "web", actions: []string{"fetch"}})

	// File skill should be allowed
	if err := runner.CheckPermission("file", "read"); err != nil {
		t.Errorf("expected file:read to be allowed, got: %v", err)
	}

	// Web skill should be denied
	if err := runner.CheckPermission("web", "fetch"); err == nil {
		t.Error("expected web:fetch to be denied")
	}

	// Nonexistent skill should return SkillNotFoundError
	err := runner.CheckPermission("nonexistent", "run")
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}

	var skillErr *SkillNotFoundError
	if !errors.As(err, &skillErr) {
		t.Errorf("expected SkillNotFoundError, got %T", err)
	}
}
