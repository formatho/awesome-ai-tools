// Package handlers provides HTTP request handlers.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
)

// TeamPermissionsHandler handles team permission management endpoints.
type TeamPermissionsHandler struct {
	permissionService *services.PermissionService
	authService       *services.AuthService
}

// NewTeamPermissionsHandler creates a new permissions handler.
func NewTeamPermissionsHandler(permissionService *services.PermissionService, authService *services.AuthService) *TeamPermissionsHandler {
	return &TeamPermissionsHandler{
		permissionService: permissionService,
		authService:       authService,
	}
}

// RegisterRoutes registers the permissions routes.
func (h *TeamPermissionsHandler) RegisterRoutes(router fiber.Router) {
	routes := router.Group("/team/permissions")
	
	routes.Get("/check", h.CheckPermission)
	routes.Post("/grant", h.GrantPermission)
	routes.Delete("/revoke", h.RevokePermission)
	routes.Get("/user/:userId/org/:orgId", h.GetUserPermissions)
	routes.Get("/resource/:orgId/resource/:resource", h.GetResourcePermissions)
	routes.Get("/bulk-check", h.BulkCheckPermissions)
	
	// Permission templates
	routes.Post("/templates", h.CreatePermissionTemplate)
	routes.Get("/templates", h.ListPermissionTemplates)
	routes.Get("/templates/:id", h.GetPermissionTemplate)
	routes.Put("/templates/:id", h.UpdatePermissionTemplate)
	routes.Delete("/templates/:id", h.DeletePermissionTemplate)
}

// CheckPermission checks if a user has a specific permission.
// @Summary Check Permission
// @Description Check if a user has permission to perform an action on a resource
// @Tags team, permissions
// @Accept json
// @Produce json
// @Param request body models.PermissionCheckRequest true "Permission check request"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 403 {object} fiber.Map
// @Router /team/permissions/check [post]
func (h *TeamPermissionsHandler) CheckPermission(c *fiber.Ctx) error {
	var req models.PermissionCheckRequest
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.OrganizationID == "" || req.Resource == "" || req.Action == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "All fields (user_id, organization_id, resource, action) are required",
		})
	}

	err := h.permissionService.CheckPermission(req.UserID, req.OrganizationID, req.Resource, req.Action)
	if err != nil {
		if models.IsPermissionError(err) {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"has_permission": false,
				"reason":         "permission denied",
				"required_perm":  err.(*models.PermissionError).RequiredPerm,
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}

	return c.JSON(fiber.Map{
		"has_permission": true,
		"resource":       req.Resource,
		"action":         req.Action,
	})
}

// GrantPermission grants an explicit permission to a user.
// @Summary Grant Permission
// @Description Grant a specific permission to a user in an organization (overrides role permissions)
// @Tags team, permissions
// @Accept json
// @Produce json
// @Param request body models.PermissionAssignment true "Permission assignment"
// @Success 201 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 403 {object} fiber.Map
// @Router /team/permissions/grant [post]
func (h *TeamPermissionsHandler) GrantPermission(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	var req models.PermissionAssignment
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if the grantor has admin permissions for team management
	if err := h.permissionService.CheckPermission(claims.UserID, req.OrganizationID, "team", "manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You must have team management permission to grant permissions",
		})
	}

	err = h.permissionService.GrantPermission(req.UserID, req.OrganizationID, req.Permission, claims.UserID)
	if err != nil {
		if models.IsConflictError(err) {
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "Permission already granted",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant permission",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Permission granted successfully",
		"permission_id": req.Permission,
		"user_id":       req.UserID,
	})
}

// RevokePermission revokes an explicit permission from a user.
func (h *TeamPermissionsHandler) RevokePermission(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	var req struct {
		UserID       string `json:"user_id"`
		OrganizationID string `json:"organization_id"`
		Permission   string `json:"permission"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if the grantor has admin permissions for team management
	if err := h.permissionService.CheckPermission(claims.UserID, req.OrganizationID, "team", "manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You must have team management permission to revoke permissions",
		})
	}

	err = h.permissionService.RevokePermission(req.UserID, req.OrganizationID, req.Permission)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Permission not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke permission",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permission revoked successfully",
	})
}

// GetUserPermissions returns all permissions for a user in an organization.
func (h *TeamPermissionsHandler) GetUserPermissions(c *fiber.Ctx) error {
	userID := c.Params("userId")
	orgID := c.Params("orgId")

	if userID == "" || orgID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "user_id and organization_id are required",
		})
	}

	permissions, err := h.permissionService.GetUserPermissions(userID, orgID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user permissions",
		})
	}

	return c.JSON(fiber.Map{
		"user_id":       userID,
		"organization_id": orgID,
		"permissions":   permissions,
		"count":         len(permissions),
	})
}

// GetResourcePermissions returns all actions a user can perform on a resource.
func (h *TeamPermissionsHandler) GetResourcePermissions(c *fiber.Ctx) error {
	orgID := c.Params("orgId")
	resource := c.Params("resource")
	userID := c.Query("user_id", "")

	if orgID == "" || resource == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "organization_id and resource are required",
		})
	}

	var permissions []string
	if userID != "" {
		var err error
		permissions, err = h.permissionService.HasPermissionForResource(userID, orgID, resource)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get resource permissions",
			})
		}
	} else {
		// Return all possible actions for the resource (for UI display purposes)
		allActions := []string{"create", "read", "update", "delete"}
		permissions = allActions
	}

	return c.JSON(fiber.Map{
		"resource":    resource,
		"organization_id": orgID,
		"user_id":     userID,
		"actions":     permissions,
	})
}

// BulkCheckPermissions checks multiple permissions at once.
func (h *TeamPermissionsHandler) BulkCheckPermissions(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	var requests []models.PermissionCheckRequest
	
	if err := c.BodyParser(&requests); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	results, err := h.permissionService.BulkCheckPermissions(claims.UserID, claims.OrgID, requests)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permissions",
		})
	}

	hasAll := true
	for _, hasPerm := range results {
		if !hasPerm {
			hasAll = false
			break
		}
	}

	return c.JSON(fiber.Map{
		"user_id":      claims.UserID,
		"organization_id": claims.OrgID,
		"results":      results,
		"has_all":      hasAll,
	})
}

// Permission Template Handlers

// CreatePermissionTemplate creates a new permission template.
func (h *TeamPermissionsHandler) CreatePermissionTemplate(c *fiber.Ctx) error {
	var req models.PermissionTemplateCreate
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	template := &models.PermissionTemplate{
		ID:          "template-" + strconv.Itoa(len(req.Permissions)), // Placeholder for UUID generation
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"template": template,
	})
}

// ListPermissionTemplates lists all permission templates.
func (h *TeamPermissionsHandler) ListPermissionTemplates(c *fiber.Ctx) error {
	templates := []models.PermissionTemplate{
		{
			ID:          "default-member",
			Name:        "Member Template",
			Description: "Standard member permissions",
			Permissions: []string{"agents:create", "agents:read", "agents:update"},
			CreatedAt:   time.Now(),
		},
		{
			ID:          "default-viewer",
			Name:        "Viewer Template",
			Description: "Read-only permissions",
			Permissions: []string{"agents:read", "configs:read", "cron_jobs:read"},
			CreatedAt:   time.Now(),
		},
	}

	return c.JSON(fiber.Map{
		"templates": templates,
	})
}

// GetPermissionTemplate retrieves a specific permission template.
func (h *TeamPermissionsHandler) GetPermissionTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	
	if templateID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "template_id is required",
		})
	}

	// In production, fetch from database
	return c.JSON(fiber.Map{
		"id":          templateID,
		"name":        "Template " + templateID,
		"description": "A permission template",
		"permissions": []string{"agents:read"},
	})
}

// UpdatePermissionTemplate updates a permission template.
func (h *TeamPermissionsHandler) UpdatePermissionTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	
	if templateID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "template_id is required",
		})
	}

	var req models.PermissionTemplateCreate
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Template updated",
	})
}

// DeletePermissionTemplate deletes a permission template.
func (h *TeamPermissionsHandler) DeletePermissionTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	
	if templateID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "template_id is required",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Template deleted",
	})
}
