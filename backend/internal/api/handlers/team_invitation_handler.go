// Package handlers provides HTTP request handlers.
package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
)

// TeamInvitationHandler handles team invitation endpoints.
type TeamInvitationHandler struct {
	invitationService *services.InvitationService
	authService       *services.AuthService
}

// NewTeamInvitationHandler creates a new invitation handler.
func NewTeamInvitationHandler(invitationService *services.InvitationService, authService *services.AuthService) *TeamInvitationHandler {
	return &TeamInvitationHandler{
		invitationService: invitationService,
		authService:       authService,
	}
}

// RegisterRoutes registers the invitation routes.
func (h *TeamInvitationHandler) RegisterRoutes(router fiber.Router) {
	routes := router.Group("/team/invitations")
	
	routes.Post("/", h.CreateInvitation)
	routes.Get("/", h.ListInvitations)
	routes.Get("/:id", h.GetInvitation)
	routes.Delete("/:id", h.CancelInvitation)
	routes.Post("/accept", h.AcceptInvitation)
	routes.Post("/reject/:id", h.RejectInvitation)
	
	// Invitation verification (public endpoint, no auth required)
	routes.Get("/verify-token", h.VerifyToken)
	
	// Statistics
	routes.Get("/stats", h.GetStats)
}

// CreateInvitation creates a new team invitation.
// @Summary Create Team Invitation
// @Description Send an invitation to join the organization as a team member
// @Tags team, invitations
// @Accept json
// @Produce json
// @Param request body models.InvitationCreate true "Invitation details"
// @Success 201 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 403 {object} fiber.Map
// @Failure 409 {object} fiber.Map
// @Router /team/invitations [post]
func (h *TeamInvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	var req models.InvitationCreate
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user has permission to manage team members
	if err := h.authService.CheckPermission(claims.UserID, claims.OrgID, "team", "manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You must have team management permission to send invitations",
		})
	}

	orgService := services.NewOrgService(h.authService.DB())

	// Verify user is a member of the organization
	if _, err := orgService.GetMember(claims.OrgID, claims.UserID); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You must be a member of this organization to send invitations",
			"details": err.Error(),
			"org_id_from_token": claims.OrgID,
			"user_id_from_token": claims.UserID,
		})
	}

	invitation, err := h.invitationService.Create(claims.OrgID, &req, "") // Pass empty string for email-based invitation (no user yet)
	if err != nil {
		if models.IsConflictError(err) {
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "User already has pending invitation or is a member",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create invitation",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"invitation": fiber.Map{
			"id":             invitation.ID,
			"email":          invitation.Email,
			"role":           invitation.Role,
			"status":         invitation.Status,
			"expires_at":     invitation.ExpiresAt.Format(time.RFC3339),
			"sented_at":      invitation.SentAt.Format(time.RFC3339),
			"created_by":     invitation.CreatedBy,
		},
	})
}

// ListInvitations returns all invitations for an organization.
func (h *TeamInvitationHandler) ListInvitations(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Check permission to view invitations
	if err := h.authService.CheckPermission(claims.UserID, claims.OrgID, "team", "manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You must have team management permission to view invitations",
		})
	}

	filter := &models.InvitationFilter{}
	
	// Parse optional query parameters
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.InvitationStatus(statusStr)
		filter.Status = &status
	}
	
	if roleStr := c.Query("role"); roleStr != "" {
		role := models.UserRole(roleStr)
		filter.Role = &role
	}

	invitations, err := h.invitationService.List(claims.OrgID, filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list invitations",
		})
	}

	// Convert to response format (exclude tokens for security)
	var invitationsResponse []fiber.Map
	for _, inv := range invitations {
		invitationsResponse = append(invitationsResponse, fiber.Map{
			"id":             inv.ID,
			"email":          inv.Email,
			"user_id":        inv.UserID,
			"role":           inv.Role,
			"message":        inv.Message,
			"status":         inv.Status,
			"expires_at":     inv.ExpiresAt.Format(time.RFC3339),
			"sented_at":      inv.SentAt.Format(time.RFC3339),
			"created_by":     inv.CreatedBy,
			"accepted_at":    nil, // Will be set if accepted
		})
		
		if inv.AcceptedAt != nil {
			invitationsResponse[len(invitationsResponse)-1]["accepted_at"] = inv.AcceptedAt.Format(time.RFC3339)
		}
	}

	return c.JSON(fiber.Map{
		"invitations": invitationsResponse,
		"count":       len(invitations),
	})
}

// GetInvitation retrieves a specific invitation.
func (h *TeamInvitationHandler) GetInvitation(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	invitationID := c.Params("id")
	
	inv, err := h.invitationService.Get(invitationID)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Invitation not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get invitation",
		})
	}

	// Verify user has permission to view this invitation (must be member)
	if err := h.authService.CheckPermission(claims.UserID, inv.OrganizationID, "team", "manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to view this invitation",
		})
	}

	return c.JSON(fiber.Map{
		"id":             inv.ID,
		"user_id":        inv.UserID,
		"organization_id": inv.OrganizationID,
		"email":          inv.Email,
		"role":           inv.Role,
		"message":        inv.Message,
		"status":         inv.Status,
		"expires_at":     inv.ExpiresAt.Format(time.RFC3339),
		"sented_at":      inv.SentAt.Format(time.RFC3339),
		"accepted_at":    nil, // Will be set if accepted
	})
}

// CancelInvitation cancels an invitation.
func (h *TeamInvitationHandler) CancelInvitation(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	invitationID := c.Params("id")
	
	err = h.invitationService.Cancel(invitationID, claims.UserID)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Invitation not found",
			})
		}
		if models.IsBadRequestError(err) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel invitation",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invitation cancelled successfully",
	})
}

// AcceptInvitation processes an invitation acceptance.
func (h *TeamInvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
	var req models.InvitationAccept
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Token == "" || req.Name == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "token, name, and password are required",
		})
	}

	member, err := h.invitationService.Accept(req.Token, &req)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Invalid or expired invitation token",
			})
		}
		if models.IsUnauthorizedError(err) {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid invitation token",
			})
		}
		if models.IsBadRequestError(err) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to accept invitation",
		})
	}

	// Return member info after successful acceptance
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invitation accepted successfully. You are now a team member.",
		"member": fiber.Map{
			"id":                 member.ID,
			"user_id":            member.UserID,
			"organization_id":    member.OrganizationID,
			"role":               member.Role,
			"status":             member.Status,
			"joined_at":          member.JoinedAt.Format(time.RFC3339),
		},
	})
}

// RejectInvitation marks an invitation as rejected.
func (h *TeamInvitationHandler) RejectInvitation(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	invitationID := c.Params("id")
	
	err = h.invitationService.Reject(invitationID)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Invitation not found",
			})
		}
		if models.IsBadRequestError(err) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reject invitation",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invitation rejected successfully",
	})
}

// VerifyToken verifies an invitation token.
func (h *TeamInvitationHandler) VerifyToken(c *fiber.Ctx) error {
	token := c.Query("token")
	
	if token == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "token query parameter is required",
		})
	}

	inv, err := h.invitationService.GetByToken(token)
	if err != nil {
		if models.IsNotFoundError(err) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Invalid invitation token",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify token",
		})
	}

	if inv.Status != models.InvitationStatusPending {
		statusMsg := ""
		switch inv.Status {
		case models.InvitationStatusAccepted:
			statusMsg = "This invitation has already been accepted"
		case models.InvitationStatusRejected:
			statusMsg = "This invitation was rejected"
		case models.InvitationStatusExpired:
			statusMsg = "This invitation has expired"
		case models.InvitationStatusCancelled:
			statusMsg = "This invitation was cancelled"
		}
		
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": statusMsg,
			"status": inv.Status,
		})
	}

	if inv.IsExpired() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invitation has expired",
			"expires_at": inv.ExpiresAt.Format(time.RFC3339),
		})
	}

	return c.JSON(fiber.Map{
		"valid": true,
		"invitation_id": inv.ID,
		"email":          inv.Email,
		"role":           inv.Role,
		"organization_id": inv.OrganizationID,
		"expires_at":     inv.ExpiresAt.Format(time.RFC3339),
	})
}

// GetStats returns invitation statistics for an organization.
func (h *TeamInvitationHandler) GetStats(c *fiber.Ctx) error {
	claims, err := h.authService.ValidateToken(c.Cookies("access_token"))
	if err != nil || claims == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	stats, err := h.invitationService.GetStats(claims.OrgID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get invitation statistics",
		})
	}

	return c.JSON(fiber.Map{
		"organization_id": claims.OrgID,
		"stats": fiber.Map{
			"total_sent":      stats.TotalSent,
			"pending":         stats.Pending,
			"accepted":        stats.Accepted,
			"rejected":        stats.Rejected,
			"expired":         stats.Expired,
			"cancelled":       stats.Cancelled,
			"acceptance_rate": calculateAcceptanceRate(stats),
		},
	})
}

// Helper function to calculate acceptance rate.
func calculateAcceptanceRate(stats *models.InvitationStats) float64 {
	if stats.TotalSent == 0 {
		return 0
	}
	return float64(stats.Accepted) / float64(stats.TotalSent) * 100
}
