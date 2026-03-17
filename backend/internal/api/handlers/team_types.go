package handlers

import "time"

// InviteMemberRequest represents a request to invite a member to a team
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

// MemberResponse represents a team member's information
type MemberResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name,omitempty"`
	Role         string    `json:"role"` // admin, member
	JoinedAt     time.Time `json:"joined_at"`
	Status       string    `json:"status"` // active, pending, inactive
	LastActiveAt time.Time `json:"last_active_at,omitempty"`
}

// MemberInvitationResponse represents the response after inviting a member
type MemberInvitationResponse struct {
	Message      string    `json:"message"`
	OrganizationID string   `json:"organization_id"`
	MemberEmail  string    `json:"member_email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"` // pending, accepted, rejected
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// InvitationVerificationResponse represents the result of verifying an invitation token
type InvitationVerificationResponse struct {
	Valid     bool      `json:"valid"`
	Message   string    `json:"message,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// MemberStatistics contains statistics about team members
type MemberStatistics struct {
	Total        int `json:"total"`
	Active       int `json:"active"`
	Pending      int `json:"pending"`
	Admins       int `json:"admins"`
	MembersCount int `json:"members"` // regular members (not admins)
}

// AgentStatistics contains statistics about agents in the organization
type AgentStatistics struct {
	Total            int     `json:"total"`
	ActiveToday      int     `json:"active_today"`
	AvgExecutionTime float64 `json:"avg_execution_time_minutes,omitempty"` // average execution time per task
	SuccessRate      float64 `json:"success_rate_percent,omitempty"`      // percentage of successful tasks
}

// OrganizationStatsResponse contains organization statistics
type OrganizationStatsResponse struct {
	OrganizationID string          `json:"organization_id"`
	Members        MemberStatistics `json:"members"`
	Agents         AgentStatistics  `json:"agents"`
}
