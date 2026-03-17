// Package services provides business logic layer for the API.
package services

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// PermissionService handles permission checking and management.
type PermissionService struct {
	db           *sql.DB
	cache        sync.Map // map[string]bool for cached permissions
	cacheExpiry  time.Time
	cacheMu      sync.RWMutex
}

// NewPermissionService creates a new permission service.
func NewPermissionService(db *sql.DB) *PermissionService {
	return &PermissionService{
		db:     db,
		cache:  sync.Map{},
	}
}

// CheckPermission checks if a user has a specific permission for an organization.
// Returns nil if permitted, error if denied.
func (s *PermissionService) CheckPermission(userID string, orgID string, resource string, action string) error {
	permissionKey := fmt.Sprintf("%s:%s:%s", userID, orgID, resource+":"+action)

	// Check cache first (for performance)
	s.cacheMu.RLock()
	if s.cacheExpiry.After(time.Now()) {
		if hasPerm, ok := s.cache.Load(permissionKey); ok && hasPerm == true {
			s.cacheMu.RUnlock()
			return nil
		}
		s.cacheMu.RUnlock()
	} else {
		s.cacheMu.RUnlock()
	}

	// Build permission key in standard format
	permissionID := resource + ":" + action

	// Check if user has explicit permission assignment
	hasExplicit, err := s.checkExplicitPermission(userID, orgID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check explicit permissions: %w", err)
	}
	if hasExplicit {
		s.cachePerm(permissionKey, true)
		return nil
	}

	// Check role-based permissions
	hasRolePerm, err := s.checkRolePermission(userID, orgID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check role permissions: %w", err)
	}
	if hasRolePerm {
		s.cachePerm(permissionKey, true)
		return nil
	}

	// Permission denied
	s.cachePerm(permissionKey, false)
	return &models.PermissionError{
		UserID:       userID,
		Resource:     resource,
		Action:       action,
		RequiredPerm: permissionID,
	}
}

// checkExplicitPermission checks for user-specific permission assignments.
func (s *PermissionService) checkExplicitPermission(userID string, orgID string, permissionID string) (bool, error) {
	query := `SELECT id FROM user_permissions 
		WHERE user_id = ? AND organization_id = ? AND permission = ?`

	var count int64
	err := s.db.QueryRow(query, userID, orgID, permissionID).Scan(&count)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// checkRolePermission checks permissions based on user's role in organization.
func (s *PermissionService) checkRolePermission(userID string, orgID string, permissionID string) (bool, error) {
	// Get user's role in the organization
	var role models.UserRole
	err := s.db.QueryRow(`SELECT role FROM user_org_members 
		WHERE user_id = ? AND organization_id = ? AND status = 'active'`, userID, orgID).Scan(&role)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check if role has this permission
	hasPermission := s.roleHasPermission(role, permissionID)
	if hasPermission {
		return true, nil
	}

	// Check inherited permissions from lower roles (owner > admin > member > viewer)
	inheritedRoles := models.RoleHierarchy[string(role)]
	for _, inheritedRole := range inheritedRoles {
		if s.roleHasPermission(models.UserRole(inheritedRole), permissionID) {
			return true, nil
		}
	}

	return false, nil
}

// roleHasPermission checks if a specific role has a permission.
func (s *PermissionService) roleHasPermission(role models.UserRole, permissionID string) bool {
	requiredAction := strings.Split(permissionID, ":")[1]
	
	switch role {
	case models.UserRoleOwner:
		return s.roleOwnerHasPermission(permissionID, requiredAction)
	case models.UserRoleAdmin:
		return s.roleAdminHasPermission(permissionID, requiredAction)
	case models.UserRoleMember:
		return s.roleMemberHasPermission(permissionID, requiredAction)
	case models.UserRoleViewer:
		return s.roleViewerHasPermission(permissionID, requiredAction)
	default:
		return false
	}
}

// roleOwnerHasPermission checks owner permissions.
func (s *PermissionService) roleOwnerHasPermission(permissionID, action string) bool {
	if permissionID == "admin:full_access" && action != "" {
		return true
	}

	switch permissionID {
	case "agent:create", "agent:read", "agent:update", "agent:delete":
		return action == "create" || action == "read" || action == "update" || action == "delete"
	case "config:create", "config:read", "config:update", "config:delete":
		return action == "create" || action == "read" || action == "update" || action == "delete"
	case "cron_jobs:create", "cron_jobs:read", "cron_jobs:update", "cron_jobs:delete":
		return action == "create" || action == "read" || action == "update" || action == "delete"
	case "team:manage", "team:settings":
		return action == "invite" || action == "remove" || action == "update_role" || action == "update"
	default:
		return false
	}
}

// roleAdminHasPermission checks admin permissions.
func (s *PermissionService) roleAdminHasPermission(permissionID, action string) bool {
	switch permissionID {
	case "agent:create", "agent:read", "agent:update", "agent:delete":
		return action == "create" || action == "read" || action == "update" || action == "delete"
	case "config:create", "config:read", "config:update", "config:delete":
		return action == "create" || action == "read" || action == "update" || action == "delete"
	case "cron_jobs:create", "cron_jobs:read", "cron_jobs:update", "cron_jobs:delete":
		return action == "create" || action == "read" || action == "update" || action == "delete"
	case "team:manage", "team:settings":
		return action == "invite" || action == "remove" || action == "update_role" || action == "update"
	default:
		return false
	}
}

// roleMemberHasPermission checks member permissions.
func (s *PermissionService) roleMemberHasPermission(permissionID, action string) bool {
	switch permissionID {
	case "agent:create", "agent:read", "agent:update":
		return action == "create" || action == "read" || action == "update"
	case "config:read":
		return action == "read"
	case "cron_jobs:read":
		return action == "read"
	default:
		return false
	}
}

// roleViewerHasPermission checks viewer permissions.
func (s *PermissionService) roleViewerHasPermission(permissionID, action string) bool {
	switch permissionID {
	case "agent:read", "config:read", "cron_jobs:read":
		return action == "read"
	default:
		return false
	}
}

// cachePerm caches a permission check result.
func (s *PermissionService) cachePerm(key string, hasPerm bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache.Store(key, hasPerm)

	// Update expiry to 1 minute from now for all cached permissions
	now := time.Now().Add(time.Minute)
	if s.cacheExpiry.Before(now) {
		s.cacheExpiry = now
	}
}

// GrantPermission grants an explicit permission to a user.
func (s *PermissionService) GrantPermission(userID string, orgID string, permissionID string, grantedBy string) error {
	// Check if permission already exists
	var existingCount int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM user_permissions 
		WHERE user_id = ? AND organization_id = ? AND permission = ?`, userID, orgID, permissionID).Scan(&existingCount)

	if err != nil {
		return fmt.Errorf("failed to check existing permissions: %w", err)
	}

	if existingCount > 0 {
		return models.NewAppError("CONFLICT", "permission already granted")
	}

	id := uuid.New().String()
	query := `INSERT INTO user_permissions (id, user_id, organization_id, permission, granted_by) 
		VALUES (?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, id, userID, orgID, permissionID, grantedBy)
	if err != nil {
		return fmt.Errorf("failed to grant permission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.NewAppError("CONFLICT", "failed to create permission assignment")
	}

	// Clear cache for this user
	s.clearUserCache(userID, orgID)

	return nil
}

// RevokePermission revokes an explicit permission from a user.
func (s *PermissionService) RevokePermission(userID string, orgID string, permissionID string) error {
	query := `DELETE FROM user_permissions 
		WHERE user_id = ? AND organization_id = ? AND permission = ?`

	result, err := s.db.Exec(query, userID, orgID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.NewAppError("NOT_FOUND", "permission not found")
	}

	s.clearUserCache(userID, orgID)
	return nil
}

// GetUserPermissions returns all permissions for a user in an organization.
func (s *PermissionService) GetUserPermissions(userID string, orgID string) ([]string, error) {
	query := `SELECT permission FROM user_permissions 
		WHERE user_id = ? AND organization_id = ?`

	rows, err := s.db.Query(query, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// clearUserCache clears cache entries for a user in an organization.
func (s *PermissionService) clearUserCache(userID string, orgID string) {
	prefix := fmt.Sprintf("%s:%s:", userID, orgID)
	s.cache.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(key.(string), prefix) {
			s.cache.Delete(key)
		}
		return true
	})
}

// HasPermissionForResource returns all actions a user can perform on a resource.
func (s *PermissionService) HasPermissionForResource(userID string, orgID string, resource string) ([]string, error) {
	actions := []string{"create", "read", "update", "delete"}
	var allowedActions []string

	for _, action := range actions {
		if err := s.CheckPermission(userID, orgID, resource, action); err == nil {
			allowedActions = append(allowedActions, action)
		}
	}

	return allowedActions, nil
}

// CheckOrganizationAccess validates if a user has access to an organization.
func (s *PermissionService) CheckOrganizationAccess(userID string, orgID string) error {
	var memberCount int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM user_org_members 
		WHERE user_id = ? AND organization_id = ? AND status = 'active'`, userID, orgID).Scan(&memberCount)

	if err == sql.ErrNoRows || memberCount == 0 {
		return models.NewAppError("NOT_FOUND", "organization access denied")
	}
	if err != nil {
		return fmt.Errorf("failed to check organization access: %w", err)
	}

	return nil
}

// BulkCheckPermissions checks multiple permissions at once.
func (s *PermissionService) BulkCheckPermissions(userID string, orgID string, requests []models.PermissionCheckRequest) (map[string]bool, error) {
	results := make(map[string]bool)

	for _, req := range requests {
		key := fmt.Sprintf("%s:%s", req.Resource, req.Action)
		if err := s.CheckPermission(userID, orgID, req.Resource, req.Action); err == nil {
			results[key] = true
		} else {
			results[key] = false
		}
	}

	return results, nil
}

// PermissionMiddleware creates a Fiber middleware for permission checking.
func (s *PermissionService) PermissionMiddleware(resource string, action string) func(c any) error {
	return func(c any) error {
		// This will be used in Fiber router setup
		// For now, return placeholder
		return nil
	}
}
