package skills

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	agent "github.com/formatho/agent-orchestrator/packages/agent-skills"
)

// ShellSkill provides shell command execution for AI agents.
// It allows running shell commands with proper timeout and context support.
type ShellSkill struct {
	// DefaultTimeout is the default timeout for command execution.
	// Default is 60 seconds if not set.
	DefaultTimeout time.Duration

	// AllowedCommands is a whitelist of allowed commands.
	// If empty, all commands are allowed (subject to ShellRunner validation).
	AllowedCommands []string

	// ForbiddenCommands is a blacklist of forbidden commands.
	// These commands will always be blocked.
	ForbiddenCommands []string
}

// NewShellSkill creates a new ShellSkill with default settings.
func NewShellSkill() *ShellSkill {
	return &ShellSkill{
		DefaultTimeout: 60 * time.Second,
		ForbiddenCommands: []string{
			"rm -rf /",
			"mkfs",
			"dd if=/dev/zero",
		},
	}
}

// Name returns the skill name.
func (s *ShellSkill) Name() string {
	return "shell"
}

// Actions returns the list of supported actions.
func (s *ShellSkill) Actions() []string {
	return []string{"run"}
}

// Execute performs the specified shell action.
func (s *ShellSkill) Execute(ctx context.Context, action string, params map[string]any) (agent.Result, error) {
	switch action {
	case "run":
		return s.run(ctx, params)
	default:
		return agent.Result{}, agent.NewExecutionError("shell", action,
			fmt.Sprintf("unknown action: %s", action))
	}
}

// run executes a shell command and returns the output.
func (s *ShellSkill) run(ctx context.Context, params map[string]any) (agent.Result, error) {
	// Get command parameter
	cmdRaw, ok := params["command"]
	if !ok {
		return agent.Result{}, agent.NewExecutionError("shell", "run", "missing required parameter: command")
	}

	command, ok := cmdRaw.(string)
	if !ok {
		return agent.Result{}, agent.NewExecutionError("shell", "run", "parameter 'command' must be a string")
	}

	// Security check: validate command
	if err := s.validateCommand(command); err != nil {
		return agent.Result{}, agent.NewExecutionError("shell", "run", err.Error())
	}

	// Get timeout from params or use default
	timeout := s.DefaultTimeout
	if timeoutMs, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Determine shell and args based on OS
	var cmd *exec.Cmd
	shell := "/bin/sh"
	flag := "-c"

	// Check if a specific shell is requested
	if shellPath, ok := params["shell"].(string); ok {
		shell = shellPath
	}

	// Build command
	cmd = exec.CommandContext(ctx, shell, flag, command)

	// Set working directory if specified
	if cwd, ok := params["cwd"].(string); ok {
		cmd.Dir = cwd
	}

	// Set environment variables if specified
	if env, ok := params["env"].(map[string]any); ok {
		// Start with current environment
		cmd.Env = append(cmd.Env, s.envMapToList(env)...)
	}

	// Capture both stdout and stderr
	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	// Check for context cancellation
	if ctx.Err() == context.DeadlineExceeded {
		return agent.Result{
			Success: false,
			Data:    string(output),
			Message: fmt.Sprintf("command timed out after %v", timeout),
			Metadata: map[string]any{
				"command":  command,
				"duration": duration.String(),
				"timeout":  timeout.String(),
				"timedOut": true,
			},
		}, agent.NewExecutionError("shell", "run", "command timed out")
	}

	// Check for context cancellation
	if ctx.Err() == context.Canceled {
		return agent.Result{
			Success: false,
			Data:    string(output),
			Message: "command was cancelled",
			Metadata: map[string]any{
				"command":   command,
				"cancelled": true,
			},
		}, agent.NewExecutionError("shell", "run", "command was cancelled")
	}

	// Build result
	result := agent.Result{
		Success: err == nil,
		Data:    string(output),
		Message: fmt.Sprintf("command executed in %v", duration),
		Metadata: map[string]any{
			"command":  command,
			"duration": duration.String(),
			"timedOut": false,
		},
	}

	// Handle command error
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Metadata["exitCode"] = exitErr.ExitCode()
			result.Message = fmt.Sprintf("command exited with code %d: %s",
				exitErr.ExitCode(), strings.TrimSpace(string(output)))
		} else {
			result.Message = fmt.Sprintf("command failed: %s", err.Error())
		}
	}

	// Include exit code if successful
	if err == nil {
		result.Metadata["exitCode"] = 0
	}

	return result, nil
}

// validateCommand checks if the command is allowed to run.
func (s *ShellSkill) validateCommand(command string) error {
	// Check forbidden commands
	for _, forbidden := range s.ForbiddenCommands {
		if strings.Contains(command, forbidden) {
			return fmt.Errorf("command contains forbidden pattern: %s", forbidden)
		}
	}

	// If allowed commands list is set, check it
	if len(s.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range s.AllowedCommands {
			if strings.HasPrefix(command, allowedCmd) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command not in allowed list: %s", command)
		}
	}

	return nil
}

// envMapToList converts an environment map to a list of KEY=VALUE strings.
func (s *ShellSkill) envMapToList(env map[string]any) []string {
	var result []string
	for key, value := range env {
		result = append(result, fmt.Sprintf("%s=%v", key, value))
	}
	return result
}
