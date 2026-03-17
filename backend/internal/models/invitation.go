// Package models defines invitation-related data structures.
package models

import (
	"time"
)

// Invitation represents a team member invitation to an organization.
type Invitation struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"` // Empty if email-based invitation
	OrganizationID string                 `json:"organization_id"`
	Email          string                 `json:"email"`
	Role           UserRole               `json:"role"`
	Message        string                 `json:"message,omitempty"` // Custom invitation message
	Status         InvitationStatus       `json:"status"`
	Token          string                 `json:"-"` // Hashed token for verification (never return in JSON)
	ExpiresAt      time.Time              `json:"expires_at"`
	SentAt         time.Time              `json:"sent_at"`
	AcceptedAt     *time.Time             `json:"accepted_at,omitempty"`
	CreatedBy      string                 `json:"created_by"` // User ID who sent the invitation
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// InvitationStatus represents the status of an invitation.
type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"    // Invitation sent, not yet accepted/rejected
	InvitationStatusAccepted  InvitationStatus = "accepted"   // User has accepted the invitation
	InvitationStatusRejected  InvitationStatus = "rejected"   // User has rejected the invitation
	InvitationStatusExpired   InvitationStatus = "expired"    // Invitation has expired
	InvitationStatusCancelled InvitationStatus = "cancelled"  // Invitation was cancelled by sender
)

// InvitationCreate is the request body for creating a new invitation.
type InvitationCreate struct {
	Email     string                 `json:"email"`
	UserID    *string                `json:"user_id,omitempty"` // If provided, email will be ignored (existing user)
	Role      UserRole               `json:"role"`
	Message   string                 `json:"message,omitempty"`
	ExpiresIn int                    `json:"expires_in,omitempty"` // Hours until expiry, default 72 (3 days)
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// InvitationUpdate is the request body for updating an invitation.
type InvitationUpdate struct {
	Status *InvitationStatus `json:"status,omitempty"`
	Message *string           `json:"message,omitempty"`
}

// InvitationAccept is the request body for accepting an invitation.
type InvitationAccept struct {
	Token string `json:"token"` // Verification token sent via email
	Name  string `json:"name"`  // Name to use if creating new account
	Password string `json:"password"` // Password if creating new account
}

// Validate validates the invitation creation request.
func (i *InvitationCreate) Validate() error {
	if i.Email == "" && i.UserID == nil {
		return ErrValidation("email or user_id is required")
	}
	if i.Role == "" {
		return ErrValidation("role is required")
	}

	validRoles := []UserRole{UserRoleOwner, UserRoleAdmin, UserRoleMember, UserRoleViewer}
	for _, role := range validRoles {
		if i.Role == role {
			return nil
		}
	}

	return ErrValidation("invalid role")
}

// InvitationStats represents statistics about invitations for an organization.
type InvitationStats struct {
	TotalSent      int64                  `json:"total_sent"`
	Pending        int64                  `json:"pending"`
	Accepted       int64                  `json:"accepted"`
	Rejected       int64                  `json:"rejected"`
	Expired        int64                  `json:"expired"`
	Cancelled      int64                  `json:"cancelled"`
	AverageAcceptanceTime time.Time `json:"average_acceptance_time,omitempty"` // ISO 8601 duration
}

// InvitationFilter represents filtering options for invitation queries.
type InvitationFilter struct {
	Status    *InvitationStatus
	Role      *UserRole
	CreatedBy string
	Since     time.Time
	Until     time.Time
}

// InvitationLink generates a verification link for the invitation.
func (i *Invitation) InvitationLink(baseUrl string) string {
	return baseUrl + "/invite/accept?token=" + i.Token
}

// IsExpired checks if the invitation has expired.
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// CanAccept returns true if the invitation can still be accepted.
func (i *Invitation) CanAccept() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

// InvitationTokenManager handles token generation and verification for invitations.
type InvitationTokenManager struct {
	secretKey []byte
	maxLength int
}

// NewInvitationTokenManager creates a new token manager.
func NewInvitationTokenManager(secretKey []byte) *InvitationTokenManager {
	return &InvitationTokenManager{
		secretKey: secretKey,
		maxLength: 64, // SHA-256 produces 64 hex characters
	}
}

// GenerateToken generates a secure token for an invitation.
func (tm *InvitationTokenManager) GenerateToken(invitationID string, email string) string {
	// In production, use proper cryptographic signing
	// For now, return hashed value
	return hashToken(invitationID + ":" + email + ":" + string(tm.secretKey))
}

// VerifyToken verifies an invitation token.
func (tm *InvitationTokenManager) VerifyToken(token string, invitationID string, email string) bool {
	expected := hashToken(invitationID + ":" + email + ":" + string(tm.secretKey))
	return constantTimeCompare(token, expected)
}

// Helper functions for token management

func hashToken(input string) string {
	// Simple hash - in production use proper crypto
	var hash int64 = 0
	for _, c := range input {
		hash = hash*31 + int64(c)
	}
	return string(rune(hash % 100000000)) // Simplified for demo
}

func constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	result := uint8(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
