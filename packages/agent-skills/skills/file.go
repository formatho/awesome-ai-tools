package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agent "github.com/formatho/agent-orchestrator/packages/agent-skills"
)

// FileSkill provides file system operations for AI agents.
// It supports reading, writing, deleting, and listing files and directories.
type FileSkill struct {
	// BaseDir restricts all operations to this directory and its subdirectories.
	// If empty, operations are allowed anywhere.
	BaseDir string
}

// NewFileSkill creates a new FileSkill with an optional base directory restriction.
func NewFileSkill(baseDir string) *FileSkill {
	return &FileSkill{BaseDir: baseDir}
}

// Name returns the skill name.
func (s *FileSkill) Name() string {
	return "file"
}

// Actions returns the list of supported actions.
func (s *FileSkill) Actions() []string {
	return []string{"read", "write", "delete", "list"}
}

// Execute performs the specified file action.
func (s *FileSkill) Execute(ctx context.Context, action string, params map[string]any) (agent.Result, error) {
	// Get the path parameter
	pathRaw, ok := params["path"]
	if !ok {
		return agent.Result{}, agent.NewExecutionError("file", action, "missing required parameter: path")
	}

	path, ok := pathRaw.(string)
	if !ok {
		return agent.Result{}, agent.NewExecutionError("file", action, "parameter 'path' must be a string")
	}

	// Resolve and validate path
	fullPath, err := s.resolvePath(path)
	if err != nil {
		return agent.Result{}, agent.NewExecutionError("file", action, err.Error())
	}

	switch action {
	case "read":
		return s.readFile(ctx, fullPath)
	case "write":
		return s.writeFile(ctx, fullPath, params)
	case "delete":
		return s.deleteFile(ctx, fullPath, params)
	case "list":
		return s.listFiles(ctx, fullPath, params)
	default:
		return agent.Result{}, agent.NewExecutionError("file", action,
			fmt.Sprintf("unknown action: %s", action))
	}
}

// resolvePath resolves the path and ensures it's within the base directory if set.
func (s *FileSkill) resolvePath(path string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// If base directory is set, ensure path is within it
	if s.BaseDir != "" {
		baseAbs, err := filepath.Abs(s.BaseDir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve base directory: %w", err)
		}

		// Check if path is within base directory
		if !strings.HasPrefix(absPath, baseAbs) {
			return "", fmt.Errorf("path '%s' is outside allowed directory '%s'", path, s.BaseDir)
		}
	}

	return absPath, nil
}

// readFile reads the contents of a file.
func (s *FileSkill) readFile(ctx context.Context, path string) (agent.Result, error) {
	select {
	case <-ctx.Done():
		return agent.Result{}, ctx.Err()
	default:
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return agent.Result{}, agent.NewExecutionError("file", "read", err.Error())
	}

	return agent.Result{
		Success: true,
		Data:    string(content),
		Message: fmt.Sprintf("successfully read %d bytes from %s", len(content), path),
		Metadata: map[string]any{
			"path": path,
			"size": len(content),
		},
	}, nil
}

// writeFile writes content to a file.
func (s *FileSkill) writeFile(ctx context.Context, path string, params map[string]any) (agent.Result, error) {
	select {
	case <-ctx.Done():
		return agent.Result{}, ctx.Err()
	default:
	}

	contentRaw, ok := params["content"]
	if !ok {
		return agent.Result{}, agent.NewExecutionError("file", "write", "missing required parameter: content")
	}

	content, ok := contentRaw.(string)
	if !ok {
		return agent.Result{}, agent.NewExecutionError("file", "write", "parameter 'content' must be a string")
	}

	// Check if we should create parent directories
	mkdir := false
	if v, ok := params["mkdir"].(bool); ok {
		mkdir = v
	}

	if mkdir {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return agent.Result{}, agent.NewExecutionError("file", "write",
				fmt.Sprintf("failed to create parent directories: %s", err))
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return agent.Result{}, agent.NewExecutionError("file", "write", err.Error())
	}

	return agent.Result{
		Success: true,
		Message: fmt.Sprintf("successfully wrote %d bytes to %s", len(content), path),
		Metadata: map[string]any{
			"path": path,
			"size": len(content),
		},
	}, nil
}

// deleteFile deletes a file or directory.
func (s *FileSkill) deleteFile(ctx context.Context, path string, params map[string]any) (agent.Result, error) {
	select {
	case <-ctx.Done():
		return agent.Result{}, ctx.Err()
	default:
	}

	// Check if recursive delete
	recursive := false
	if v, ok := params["recursive"].(bool); ok {
		recursive = v
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return agent.Result{}, agent.NewExecutionError("file", "delete", err.Error())
	}

	if info.IsDir() {
		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
	} else {
		err = os.Remove(path)
	}

	if err != nil {
		return agent.Result{}, agent.NewExecutionError("file", "delete", err.Error())
	}

	return agent.Result{
		Success: true,
		Message: fmt.Sprintf("successfully deleted %s", path),
		Metadata: map[string]any{
			"path":      path,
			"is_dir":    info.IsDir(),
			"recursive": recursive,
		},
	}, nil
}

// listFiles lists the contents of a directory.
func (s *FileSkill) listFiles(ctx context.Context, path string, params map[string]any) (agent.Result, error) {
	select {
	case <-ctx.Done():
		return agent.Result{}, ctx.Err()
	default:
	}

	// Check if recursive
	recursive := false
	if v, ok := params["recursive"].(bool); ok {
		recursive = v
	}

	var files []map[string]any

	if recursive {
		err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			files = append(files, map[string]any{
				"path":  walkPath,
				"name":  info.Name(),
				"isDir": info.IsDir(),
				"size":  info.Size(),
			})
			return nil
		})
		if err != nil {
			return agent.Result{}, agent.NewExecutionError("file", "list", err.Error())
		}
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return agent.Result{}, agent.NewExecutionError("file", "list", err.Error())
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, map[string]any{
				"path":  filepath.Join(path, entry.Name()),
				"name":  entry.Name(),
				"isDir": entry.IsDir(),
				"size":  info.Size(),
			})
		}
	}

	return agent.Result{
		Success: true,
		Data:    files,
		Message: fmt.Sprintf("listed %d items in %s", len(files), path),
		Metadata: map[string]any{
			"path":      path,
			"count":     len(files),
			"recursive": recursive,
		},
	}, nil
}
