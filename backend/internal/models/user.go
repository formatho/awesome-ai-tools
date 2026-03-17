// Package models defines user-related data structures.
package models

import (
	"time"
)

// UserRole represents the role of a user in an organization.
type UserRole string

const (
	UserRoleOwner  UserRole = "owner"
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
	UserRoleViewer UserRole = "viewer"
)

// User represents a user in the system.
type User struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Password  string                 `json:"-"` // Never return password in JSON
	Name      string                 `json:"name"`
	AvatarURL string                 `json:"avatar_url,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// UserCreate is the request body for creating a new user.
type UserCreate struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserUpdate is the request body for updating a user.
type UserUpdate struct {
	Name      *string                `json:"name,omitempty"`
	AvatarURL *string                `json:"avatar_url,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// UserLogin is the request body for user login.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates the user creation request.
func (u *UserCreate) Validate() error {
	if u.Email == "" {
		return ErrValidation("email is required")
	}
	if u.Password == "" {
		return ErrValidation("password is required")
	}
	if len(u.Password) < 8 {
		return ErrValidation("password must be at least 8 characters")
	}
	return nil
}

// UserOrgMember represents a user's membership in an organization.
type UserOrgMember struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	OrganizationID string               `json:"organization_id"`
	Role         UserRole               `json:"role"`
	Status       string                 `json:"status"` // active, pending, inactive
	JoinedAt     time.Time              `json:"joined_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UserOrgMemberCreate is the request body for adding a member to an organization.
type UserOrgMemberCreate struct {
	UserID       string      `json:"user_id"`
	Email        string      `json:"email"` // For invitation by email
	Role         UserRole    `json:"role"`
	Status       string      `json:"status,omitempty"` // default: "pending"
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the organization member creation request.
func (m *UserOrgMemberCreate) Validate() error {
	if m.Role == "" {
		return ErrValidation("role is required")
	}
	if m.UserID == "" && m.Email == "" {
		return ErrValidation("user_id or email is required")
	}
	return nil
}

// MemberRoleUpdate is the request body for updating a member's role.
type MemberRoleUpdate struct {
	Role UserRole `json:"role"`
}

// Validate validates the role update request.
func (r *MemberRoleUpdate) Validate() error {
	if r.Role == "" {
		return ErrValidation("role is required")
	}
	if r.Role != UserRoleOwner && r.Role != UserRoleAdmin && 
	   r.Role != UserRoleMember && r.Role != UserRoleViewer {
		return ErrValidation("invalid role")
	}
	return nil
}
