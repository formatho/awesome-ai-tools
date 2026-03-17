// Package models defines permission-related data structures.
package models

import (
	"errors"
	"time"
)

// Permission represents a granular permission in the system.
type Permission struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"` // e.g., "agents", "configs", "cron_jobs"
	Actions     []string               `json:"actions"`  // e.g., ["create", "read", "update", "delete"]
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RolePermission represents the mapping between roles and permissions.
type RolePermission struct {
	ID         string                 `json:"id"`
	Role       string                 `json:"role"` // owner, admin, member, viewer
	Permission string                 `json:"permission"` // e.g., "agents:create"
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// PermissionAssignment represents a user's explicit permission assignment (overrides role).
type PermissionAssignment struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	OrganizationID string             `json:"organization_id"`
	Permission   string                 `json:"permission"` // e.g., "agents:create"
	GrantedBy    string                 `json:"granted_by"` // user ID who granted it
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// PermissionCheckRequest is used to check if a user has a specific permission.
type PermissionCheckRequest struct {
	UserID       string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Resource     string `json:"resource"`
	Action       string `json:"action"`
}

// Common permissions for the system
var (
	// Agent permissions
	PermAgentCreate = &Permission{
		ID:          "agent:create",
		Name:        "Create Agents",
		Description: "Ability to create new automation agents",
		Resource:    "agents",
		Actions:     []string{"create"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermAgentRead = &Permission{
		ID:          "agent:read",
		Name:        "View Agents",
		Description: "Ability to view automation agents and their details",
		Resource:    "agents",
		Actions:     []string{"read"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermAgentUpdate = &Permission{
		ID:          "agent:update",
		Name:        "Edit Agents",
		Description: "Ability to modify existing automation agents",
		Resource:    "agents",
		Actions:     []string{"update"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermAgentDelete = &Permission{
		ID:          "agent:delete",
		Name:        "Delete Agents",
		Description: "Ability to delete automation agents",
		Resource:    "agents",
		Actions:     []string{"delete"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Config permissions
	PermConfigCreate = &Permission{
		ID:          "config:create",
		Name:        "Create Configs",
		Description: "Ability to create new agent configurations",
		Resource:    "configs",
		Actions:     []string{"create"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermConfigRead = &Permission{
		ID:          "config:read",
		Name:        "View Configs",
		Description: "Ability to view agent configurations",
		Resource:    "configs",
		Actions:     []string{"read"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermConfigUpdate = &Permission{
		ID:          "config:update",
		Name:        "Edit Configs",
		Description: "Ability to modify agent configurations",
		Resource:    "configs",
		Actions:     []string{"update"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermConfigDelete = &Permission{
		ID:          "config:delete",
		Name:        "Delete Configs",
		Description: "Ability to delete agent configurations",
		Resource:    "configs",
		Actions:     []string{"delete"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Cron permissions
	PermCronCreate = &Permission{
		ID:          "cron:create",
		Name:        "Create Cron Jobs",
		Description: "Ability to create scheduled tasks",
		Resource:    "cron_jobs",
		Actions:     []string{"create"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermCronRead = &Permission{
		ID:          "cron:read",
		Name:        "View Cron Jobs",
		Description: "Ability to view scheduled tasks",
		Resource:    "cron_jobs",
		Actions:     []string{"read"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermCronUpdate = &Permission{
		ID:          "cron:update",
		Name:        "Edit Cron Jobs",
		Description: "Ability to modify scheduled tasks",
		Resource:    "cron_jobs",
		Actions:     []string{"update"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermCronDelete = &Permission{
		ID:          "cron:delete",
		Name:        "Delete Cron Jobs",
		Description: "Ability to delete scheduled tasks",
		Resource:    "cron_jobs",
		Actions:     []string{"delete"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Team management permissions
	PermTeamManage = &Permission{
		ID:          "team:manage",
		Name:        "Manage Team Members",
		Description: "Ability to invite, remove, and manage team members",
		Resource:    "team",
		Actions:     []string{"invite", "remove", "update_role"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	PermTeamSettings = &Permission{
		ID:          "team:settings",
		Name:        "Manage Team Settings",
		Description: "Ability to modify team and organization settings",
		Resource:    "team_settings",
		Actions:     []string{"update"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Admin permissions (only for owners)
	PermAdminFull = &Permission{
		ID:          "admin:full_access",
		Name:        "Full Administrative Access",
		Description: "Complete administrative control over the organization",
		Resource:    "admin",
		Actions:     []string{"read", "update", "delete"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Default role permissions mapping
	DefaultRolePermissions = map[string][]string{
		"owner": {
			PermAgentCreate.ID, PermAgentRead.ID, PermAgentUpdate.ID, PermAgentDelete.ID,
			PermConfigCreate.ID, PermConfigRead.ID, PermConfigUpdate.ID, PermConfigDelete.ID,
			PermCronCreate.ID, PermCronRead.ID, PermCronUpdate.ID, PermCronDelete.ID,
			PermTeamManage.ID, PermTeamSettings.ID, PermAdminFull.ID,
		},
		"admin": {
			PermAgentCreate.ID, PermAgentRead.ID, PermAgentUpdate.ID, PermAgentDelete.ID,
			PermConfigCreate.ID, PermConfigRead.ID, PermConfigUpdate.ID, PermConfigDelete.ID,
			PermCronCreate.ID, PermCronRead.ID, PermCronUpdate.ID, PermCronDelete.ID,
			PermTeamManage.ID, PermTeamSettings.ID,
		},
		"member": {
			PermAgentCreate.ID, PermAgentRead.ID, PermAgentUpdate.ID,
			PermConfigRead.ID, PermCronRead.ID,
		},
		"viewer": {
			PermAgentRead.ID, PermConfigRead.ID, PermCronRead.ID,
		},
	}

	// Role hierarchy for permission inheritance
	RoleHierarchy = map[string][]string{
		"owner":  {"admin", "member", "viewer"},
		"admin":  {"member", "viewer"},
		"member": {"viewer"},
		"viewer": {},
	}
)

// PermissionError represents a permission denial error.
type PermissionError struct {
	UserID       string
	Resource     string
	Action       string
	RequiredPerm string
}

func (e *PermissionError) Error() string {
	return "permission denied: user does not have permission to " + e.Action + " on " + e.Resource
}

// IsPermissionError checks if error is a permission error.
func IsPermissionError(err error) bool {
	var permErr *PermissionError
	return errors.As(err, &permErr)
}

// PermissionTemplate represents a reusable permission template for teams.
type PermissionTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions []string               `json:"permissions"` // list of permission IDs
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// PermissionTemplateCreate is the request body for creating a permission template.
type PermissionTemplateCreate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the permission template creation request.
func (p *PermissionTemplateCreate) Validate() error {
	if p.Name == "" {
		return ErrValidation("name is required")
	}
	if len(p.Permissions) == 0 {
		return ErrValidation("at least one permission is required")
	}
	return nil
}
