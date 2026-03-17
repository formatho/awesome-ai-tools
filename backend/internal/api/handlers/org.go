// Package handlers provides HTTP handlers for the REST API.
package handlers

import (
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// OrgHandler handles organization-related requests.
type OrgHandler struct {
	service *services.OrgService
}

// NewOrgHandler creates a new organization handler.
func NewOrgHandler(service *services.OrgService) *OrgHandler {
	return &OrgHandler{service: service}
}

// List returns all organizations.
// @Summary List organizations
// @Description Get all organizations
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {array} models.Organization
// @Router /api/organizations [get]
func (h *OrgHandler) List(c *fiber.Ctx) error {
	orgs, err := h.service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orgs)
}

// Get returns a single organization.
// @Summary Get organization
// @Description Get a single organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} models.Organization
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id} [get]
func (h *OrgHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	org, err := h.service.Get(id)
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(org)
}

// GetBySlug returns a single organization by slug.
// @Summary Get organization by slug
// @Description Get a single organization by slug
// @Tags organizations
// @Accept json
// @Produce json
// @Param slug path string true "Organization slug"
// @Success 200 {object} models.Organization
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/slug/{slug} [get]
func (h *OrgHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	org, err := h.service.GetBySlug(slug)
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(org)
}

// GetByOwner returns all organizations for a specific owner.
// @Summary Get organizations by owner
// @Description Get all organizations for a specific owner
// @Tags organizations
// @Accept json
// @Produce json
// @Param ownerId path string true "Owner ID"
// @Success 200 {array} models.Organization
// @Router /api/organizations/owner/{ownerId} [get]
func (h *OrgHandler) GetByOwner(c *fiber.Ctx) error {
	ownerID := c.Params("ownerId")
	orgs, err := h.service.GetByOwner(ownerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orgs)
}

// Create creates a new organization.
// @Summary Create organization
// @Description Create a new organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param org body models.OrganizationCreate true "Organization data"
// @Param X-Owner-ID header string true "Owner ID"
// @Success 201 {object} models.Organization
// @Failure 400 {object} fiber.Map
// @Router /api/organizations [post]
func (h *OrgHandler) Create(c *fiber.Ctx) error {
	var req models.OrganizationCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get owner ID from header
	ownerID := c.Get("X-Owner-ID")
	if ownerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "X-Owner-ID header is required"})
	}

	org, err := h.service.Create(&req, ownerID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "VALIDATION_ERROR" {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(org)
}

// Update updates an organization.
// @Summary Update organization
// @Description Update an existing organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param org body models.OrganizationUpdate true "Organization data"
// @Success 200 {object} models.Organization
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id} [put]
func (h *OrgHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.OrganizationUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	org, err := h.service.Update(id, &req)
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(org)
}

// Delete deletes an organization.
// @Summary Delete organization
// @Description Delete an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 204
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id} [delete]
func (h *OrgHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.Delete(id)
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

// SwitchOrganization allows switching the active organization.
// @Summary Switch organization
// @Description Switch the active organization for the current user
// @Tags organizations
// @Accept json
// @Produce json
// @Param switch body models.OrganizationSwitch true "Organization switch data"
// @Success 200 {object} fiber.Map
// @Router /api/organizations/switch [post]
func (h *OrgHandler) SwitchOrganization(c *fiber.Ctx) error {
	var req models.OrganizationSwitch
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Verify organization exists
	org, err := h.service.Get(req.OrganizationID)
	if err != nil {
		if err.Error() == "organization not found" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// In a real implementation, this would update the user's session
	// For now, just return the organization
	return c.JSON(fiber.Map{
		"message":        "Organization switched successfully",
		"organization":   org,
		"organizationId": org.ID,
	})
}

// ListMembers returns all members of an organization.
// @Summary List organization members
// @Description Get all team members for an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {array} models.UserOrgMember
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id}/members [get]
func (h *OrgHandler) ListMembers(c *fiber.Ctx) error {
	id := c.Params("id")
	members, err := h.service.ListMembers(id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "NOT_FOUND" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Include user details if available
	result := make([]fiber.Map, len(members))
	for i, member := range members {
		userData := fiber.Map{
			"id":           member.ID,
			"user_id":      member.UserID,
			"role":         member.Role,
			"status":       member.Status,
			"joined_at":    member.JoinedAt,
			"metadata":     member.Metadata,
		}
		
		// Try to get user details from service if available
		if h.service.HasUserService() {
			user, err := h.service.GetUserByID(member.UserID)
			if err == nil && user != nil {
				userData["email"] = user.Email
				userData["name"] = user.Name
				userData["avatar_url"] = user.AvatarURL
			}
		}
		
		result[i] = userData
	}

	return c.JSON(result)
}

// InviteMember invites a new member to an organization by email.
// @Summary Invite team member
// @Description Send invitation to join the organization via email
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param invite body models.UserOrgMemberCreate true "Invitation data"
// @Success 201 {object} models.UserOrgMember
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id}/members [post]
func (h *OrgHandler) InviteMember(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.UserOrgMemberCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	member, err := h.service.InviteMember(id, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "NOT_FOUND" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "CONFLICT" {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(member)
}

// GetMember returns a specific member of an organization.
// @Summary Get team member
// @Description Get details about a specific team member
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Success 200 {object} models.UserOrgMember
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id}/members/{userId} [get]
func (h *OrgHandler) GetMember(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Params("userId")
	
	member, err := h.service.GetMember(orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(member)
}

// UpdateMemberRole updates a team member's role.
// @Summary Update member role
// @Description Change the role of a team member within an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID"
// @Param update body models.MemberRoleUpdate true "Role update data"
// @Success 200 {object} models.UserOrgMember
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id}/members/{userId}/role [patch]
func (h *OrgHandler) UpdateMemberRole(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Params("userId")
	
	var req models.MemberRoleUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	member, err := h.service.UpdateMemberRole(orgID, userID, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == "NOT_FOUND" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(member)
}

// RemoveMember removes a team member from an organization.
// @Summary Remove team member
// @Description Remove a user from the organization (they lose access)
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID to remove"
// @Success 204
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id}/members/{userId} [delete]
func (h *OrgHandler) RemoveMember(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Params("userId")
	
	err := h.service.RemoveMember(orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

// KickMember is an alias for RemoveMember (for API clarity).
// @Summary Kick team member
// @Description Kick a user from the organization immediately
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param userId path string true "User ID to kick"
// @Success 204
// @Failure 404 {object} fiber.Map
// @Router /api/organizations/{id}/members/{userId}/kick [delete]
func (h *OrgHandler) KickMember(c *fiber.Ctx) error {
	return h.RemoveMember(c)
}

