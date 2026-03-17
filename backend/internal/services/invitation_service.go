// Package services provides business logic layer for the API.
package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// InvitationService handles team member invitations.
type InvitationService struct {
	db                *sql.DB
	tokenManager      *models.InvitationTokenManager
	emailSender       EmailSender // Interface for sending emails
	defaultExpiryHours int
}

// EmailSender defines the interface for email sending.
type EmailSender interface {
	SendInvitation(email string, invitationName string, link string, role string, message string) error
	SendInvitationReminder(invitationID string, email string) error
}

// MockEmailSender is a mock implementation for testing.
type MockEmailSender struct{}

func (m *MockEmailSender) SendInvitation(email string, invitationName string, link string, role string, message string) error {
	// In production, this would send an actual email
	fmt.Printf("Would send invitation email to: %s\n", email)
	fmt.Printf("Link: %s\n", link)
	fmt.Printf("Role: %s\n", role)
	if message != "" {
		fmt.Printf("Message: %s\n", message)
	}
	return nil
}

func (m *MockEmailSender) SendInvitationReminder(invitationID string, email string) error {
	fmt.Printf("Would send reminder for invitation: %s to: %s\n", invitationID, email)
	return nil
}

// NewInvitationService creates a new invitation service.
func NewInvitationService(db *sql.DB, emailSender EmailSender) *InvitationService {
	if db == nil {
		panic("database is required for InvitationService")
	}

	if emailSender == nil {
		emailSender = &MockEmailSender{}
	}

	return &InvitationService{
		db:                db,
		tokenManager:      models.NewInvitationTokenManager([]byte("invitation-secret-key")),
		emailSender:       emailSender,
		defaultExpiryHours: 72, // 3 days default
	}
}

// Create creates a new invitation.
func (s *InvitationService) Create(orgID string, req *models.InvitationCreate, createdBy string) (*models.Invitation, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify organization exists
	orgService := NewOrgService(s.db)
	if _, err := orgService.Get(orgID); err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if user already invited to this organization
	var existingCount int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM invitations 
		WHERE organization_id = ? AND email = ? AND status IN ('pending', 'accepted')`, orgID, req.Email).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invitations: %w", err)
	}
	if existingCount > 0 {
		return nil, models.NewAppError("CONFLICT", "user already has pending invitation")
	}

	// Check if user is already an active member (only if UserID is provided)
	var memberCount int64
	if req.UserID != nil && *req.UserID != "" {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM user_org_members
			WHERE organization_id = ? AND user_id = ? AND status = 'active'`, orgID, *req.UserID).Scan(&memberCount)
		if err != nil {
			return nil, fmt.Errorf("failed to check membership: %w", err)
		}
		if memberCount > 0 {
			return nil, models.NewAppError("CONFLICT", "user is already a member of this organization")
		}
	}

	// Generate unique token
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
	if req.ExpiresIn <= 0 {
		expiresAt = time.Now().Add(time.Duration(s.defaultExpiryHours) * time.Hour)
	}

	// Create invitation record
	id := uuid.New().String()
	invitation := &models.Invitation{
		ID:             id,
		UserID:         "", // Empty for email-based invitations (user will create account later)
		OrganizationID: orgID,
		Email:          req.Email,
		Role:           req.Role,
		Message:        req.Message,
		Status:         models.InvitationStatusPending,
		Token:          token,
		ExpiresAt:      expiresAt,
		SentAt:         time.Now(),
		CreatedBy:      createdBy,
		Metadata:       req.Metadata,
	}

	// Insert into database - use NULL for user_id when it's empty (email-based invitation)
	query := `INSERT INTO invitations (id, user_id, organization_id, email, role, message, status, token, expires_at, sent_at, created_by) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var userIDParam interface{} = ""
	if req.UserID != nil && *req.UserID != "" {
		userIDParam = *req.UserID
	} else {
		userIDParam = nil // NULL for email-based invitations
	}

	result, err := s.db.Exec(query, invitation.ID, userIDParam, orgID, invitation.Email, 
		invitation.Role, invitation.Message, invitation.Status, invitation.Token,
		invitation.ExpiresAt, invitation.SentAt, invitation.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, models.NewAppError("CONFLICT", "failed to create invitation")
	}

	// Send invitation email if it's a new user (no UserID provided)
	if req.UserID == nil || *req.UserID == "" {
		go s.sendInvitationEmail(invitation)
	}

	return invitation, nil
}

// sendInvitationEmail sends the invitation email asynchronously.
func (s *InvitationService) sendInvitationEmail(invitation *models.Invitation) error {
	link := fmt.Sprintf("https://app.formatho.com/invite/accept?token=%s", invitation.Token)

	invitationName := "Team Member" // Will be set when user accepts
	return s.emailSender.SendInvitation(
		invitation.Email,
		invitationName,
		link,
		string(invitation.Role),
		invitation.Message,
	)
}

// Accept processes an invitation acceptance.
func (s *InvitationService) Accept(token string, req *models.InvitationAccept) (*models.UserOrgMember, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Find invitation by token
	var id, userID, orgID, email, roleStr, statusStr, createdBy string
	var message, tokenHash string
	var expiresAt time.Time
	var sentAt sql.NullTime

	query := `SELECT id, user_id, organization_id, email, role, message, status, token,
		expires_at, sent_at, created_by FROM invitations WHERE token = ?`

	err := s.db.QueryRow(query, token).Scan(&id, &userID, &orgID, &email, &roleStr, 
		&message, &statusStr, &tokenHash, &expiresAt, &sentAt, &createdBy)
	if err == sql.ErrNoRows {
		return nil, models.NewAppError("NOT_FOUND", "invalid invitation token")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}

	status := models.InvitationStatus(statusStr)
	if status == models.InvitationStatusExpired || status == models.InvitationStatusCancelled {
		return nil, models.NewAppError("BAD_REQUEST", "invitation has expired or been cancelled")
	}

	// Verify token matches
	if !s.tokenManager.VerifyToken(req.Token, id, email) {
		return nil, models.NewAppError("UNAUTHORIZED", "invalid invitation token")
	}

	// Check if already accepted
	if status == models.InvitationStatusAccepted {
		// Return error - invitation already accepted
		return nil, models.NewAppError("ALREADY_ACCEPTED", "invitation has already been accepted")
	}

	// Use user_id from invitation if available
	actualUserID := userID

	// Create or update user account if needed
	if actualUserID == "" {
		// New user - create account
		actualUserID, err = s.createUserAccount(email, req.Name, req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to create user account: %w", err)
		}
	}

	// Add user to organization with the invited role
	memberService := NewOrgService(s.db)
	member, err := memberService.InviteMember(orgID, &models.UserOrgMemberCreate{
		UserID: actualUserID,
		Role:   models.UserRole(roleStr),
		Metadata: map[string]interface{}{"invited_by_token": id},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Update invitation status
	now := time.Now()
	updateQuery := `UPDATE invitations SET status = 'accepted', accepted_at = ? WHERE id = ?`
	_, err = s.db.Exec(updateQuery, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	return member, nil
}

// createUserAccount creates a new user account.
func (s *InvitationService) createUserAccount(email string, name string, password string) (string, error) {
	// This would integrate with your user service
	// For now, return placeholder logic
	
	if email == "" || password == "" {
		return "", models.NewAppError("BAD_REQUEST", "email and password required for new account")
	}

	// TODO: Implement actual user creation with password hashing
	// This should call a UserService to create the user
	
	return "user-" + uuid.New().String(), nil
}

// Cancel cancels an invitation.
func (s *InvitationService) Cancel(invitationID string, cancelledBy string) error {
	if s.db == nil {
		return ErrNoDatabase
	}

	// Check if invitation exists and is cancellable
	var status models.InvitationStatus
	err := s.db.QueryRow(`SELECT status FROM invitations WHERE id = ?`, invitationID).Scan(&status)
	if err == sql.ErrNoRows {
		return models.NewAppError("NOT_FOUND", "invitation not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check invitation: %w", err)
	}

	if status != models.InvitationStatusPending {
		return models.NewAppError("BAD_REQUEST", "only pending invitations can be cancelled")
	}

	// Update status
	updateQuery := `UPDATE invitations SET status = 'cancelled', updated_at = ? WHERE id = ?`
	result, err := s.db.Exec(updateQuery, time.Now(), invitationID)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.NewAppError("NOT_FOUND", "invitation not found")
	}

	return nil
}

// List returns all invitations for an organization.
func (s *InvitationService) List(orgID string, filter *models.InvitationFilter) ([]*models.Invitation, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, user_id, organization_id, email, role, message, status, token, 
		expires_at, sent_at, created_by FROM invitations WHERE organization_id = ?`

	args := []interface{}{orgID}

	if filter != nil {
		if filter.Status != nil {
			query += ` AND status = $` + fmt.Sprintf("%d", len(args)+1)
			args = append(args, *filter.Status)
		}
		if filter.Role != nil {
			query += ` AND role = $` + fmt.Sprintf("%d", len(args)+1)
			args = append(args, *filter.Role)
		}
		if !filter.Since.IsZero() {
			query += ` AND sent_at >= $` + fmt.Sprintf("%d", len(args)+1)
			args = append(args, filter.Since)
		}
		if !filter.Until.IsZero() {
			query += ` AND sent_at <= $` + fmt.Sprintf("%d", len(args)+1)
			args = append(args, filter.Until)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*models.Invitation
	for rows.Next() {
		var inv models.Invitation
		var tokenHash string
		
		err := rows.Scan(&inv.ID, &inv.UserID, &inv.OrganizationID, &inv.Email, 
			&inv.Role, &inv.Message, &inv.Status, &tokenHash, &inv.ExpiresAt, 
			&inv.SentAt, &inv.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		// Never include token in response
		invitations = append(invitations, &inv)
	}

	return invitations, nil
}

// Get retrieves a specific invitation.
func (s *InvitationService) Get(invitationID string) (*models.Invitation, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, user_id, organization_id, email, role, message, status, token, 
		expires_at, sent_at, created_by FROM invitations WHERE id = ?`

	var inv models.Invitation
	var tokenHash string
	
	err := s.db.QueryRow(query, invitationID).Scan(&inv.ID, &inv.UserID, &inv.OrganizationID, 
		&inv.Email, &inv.Role, &inv.Message, &inv.Status, &tokenHash, &inv.ExpiresAt, 
		&inv.SentAt, &inv.CreatedBy)
	if err == sql.ErrNoRows {
		return nil, models.NewAppError("NOT_FOUND", "invitation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &inv, nil
}

// GetByToken retrieves an invitation by its verification token.
func (s *InvitationService) GetByToken(token string) (*models.Invitation, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, user_id, organization_id, email, role, message, status, token, 
		expires_at, sent_at, created_by FROM invitations WHERE token = ?`

	var inv models.Invitation
	var tokenHash string
	
	err := s.db.QueryRow(query, token).Scan(&inv.ID, &inv.UserID, &inv.OrganizationID, 
		&inv.Email, &inv.Role, &inv.Message, &inv.Status, &tokenHash, &inv.ExpiresAt, 
		&inv.SentAt, &inv.CreatedBy)
	if err == sql.ErrNoRows {
		return nil, models.NewAppError("NOT_FOUND", "invalid invitation token")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &inv, nil
}

// Reject marks an invitation as rejected.
func (s *InvitationService) Reject(invitationID string) error {
	if s.db == nil {
		return ErrNoDatabase
	}

	var status models.InvitationStatus
	err := s.db.QueryRow(`SELECT status FROM invitations WHERE id = ?`, invitationID).Scan(&status)
	if err == sql.ErrNoRows {
		return models.NewAppError("NOT_FOUND", "invitation not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check invitation: %w", err)
	}

	if status != models.InvitationStatusPending {
		return models.NewAppError("BAD_REQUEST", "only pending invitations can be rejected")
	}

	updateQuery := `UPDATE invitations SET status = 'rejected' WHERE id = ?`
	result, err := s.db.Exec(updateQuery, invitationID)
	if err != nil {
		return fmt.Errorf("failed to reject invitation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.NewAppError("NOT_FOUND", "invitation not found")
	}

	return nil
}

// ExpireExpiredInvitations expires all invitations that have passed their expiry time.
func (s *InvitationService) ExpireExpiredInvitations() (int64, error) {
	if s.db == nil {
		return 0, ErrNoDatabase
	}

	now := time.Now()
	updateQuery := `UPDATE invitations SET status = 'expired' WHERE status = 'pending' AND expires_at < ?`
	result, err := s.db.Exec(updateQuery, now)
	if err != nil {
		return 0, fmt.Errorf("failed to expire invitations: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// GetStats returns invitation statistics for an organization.
func (s *InvitationService) GetStats(orgID string) (*models.InvitationStats, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT 
		COUNT(*),
		SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END)
	FROM invitations WHERE organization_id = ?`

	stats := &models.InvitationStats{}
	err := s.db.QueryRow(query, orgID).Scan(
		&stats.TotalSent, 
		&stats.Pending, 
		&stats.Accepted, 
		&stats.Rejected, 
		&stats.Expired, 
		&stats.Cancelled)

	if err != nil {
		return nil, fmt.Errorf("failed to get invitation stats: %w", err)
	}

	return stats, nil
}

// generateSecureToken generates a cryptographically secure random token.
func (s *InvitationService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
