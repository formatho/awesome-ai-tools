// Package services provides business logic layer for the API.
package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
)

// OrgService handles organization operations.
type OrgService struct {
	db *sql.DB
}

// NewOrgService creates a new organization service.
func NewOrgService(db *sql.DB) *OrgService {
	return &OrgService{db: db}
}

// List returns all organizations.
func (s *OrgService) List() ([]*models.Organization, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, name, slug, owner_id, settings, metadata, created_at, updated_at
		FROM organizations ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		o := &models.Organization{}
		var settings, metadata sql.NullString

		err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerID, &settings, &metadata,
			&o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if settings.Valid && settings.String != "" {
			_ = json.Unmarshal([]byte(settings.String), &o.Settings)
		}
		if metadata.Valid && metadata.String != "" {
			_ = json.Unmarshal([]byte(metadata.String), &o.Metadata)
		}

		orgs = append(orgs, o)
	}

	return orgs, nil
}

// Get returns a single organization by ID.
func (s *OrgService) Get(id string) (*models.Organization, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, name, slug, owner_id, settings, metadata, created_at, updated_at
		FROM organizations WHERE id = ?`

	o := &models.Organization{}
	var settings, metadata sql.NullString

	err := s.db.QueryRow(query, id).Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerID, &settings, &metadata,
		&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}

	if settings.Valid && settings.String != "" {
		_ = json.Unmarshal([]byte(settings.String), &o.Settings)
	}
	if metadata.Valid && metadata.String != "" {
		_ = json.Unmarshal([]byte(metadata.String), &o.Metadata)
	}

	return o, nil
}

// GetBySlug returns a single organization by slug.
func (s *OrgService) GetBySlug(slug string) (*models.Organization, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, name, slug, owner_id, settings, metadata, created_at, updated_at
		FROM organizations WHERE slug = ?`

	o := &models.Organization{}
	var settings, metadata sql.NullString

	err := s.db.QueryRow(query, slug).Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerID, &settings, &metadata,
		&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}

	if settings.Valid && settings.String != "" {
		_ = json.Unmarshal([]byte(settings.String), &o.Settings)
	}
	if metadata.Valid && metadata.String != "" {
		_ = json.Unmarshal([]byte(metadata.String), &o.Metadata)
	}

	return o, nil
}

// GetByOwner returns all organizations for a specific owner.
func (s *OrgService) GetByOwner(ownerID string) ([]*models.Organization, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, name, slug, owner_id, settings, metadata, created_at, updated_at
		FROM organizations WHERE owner_id = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		o := &models.Organization{}
		var settings, metadata sql.NullString

		err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerID, &settings, &metadata,
			&o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if settings.Valid && settings.String != "" {
			_ = json.Unmarshal([]byte(settings.String), &o.Settings)
		}
		if metadata.Valid && metadata.String != "" {
			_ = json.Unmarshal([]byte(metadata.String), &o.Metadata)
		}

		orgs = append(orgs, o)
	}

	return orgs, nil
}

// Create creates a new organization.
func (s *OrgService) Create(req *models.OrganizationCreate, ownerID string) (*models.Organization, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate ID
	id := uuid.New().String()

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// Serialize settings and metadata
	var settingsJSON, metadataJSON []byte
	var err error

	if req.Settings != nil {
		settingsJSON, err = json.Marshal(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize settings: %w", err)
		}
	}

	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize metadata: %w", err)
		}
	}

	// Insert organization
	query := `INSERT INTO organizations (id, name, slug, owner_id, settings, metadata)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query, id, req.Name, slug, ownerID,
		settingsJSON, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Return the created organization
	return s.Get(id)
}

// Update updates an existing organization.
func (s *OrgService) Update(id string, req *models.OrganizationUpdate) (*models.Organization, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Check if organization exists
	_, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Build update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.Settings != nil {
		settingsJSON, err := json.Marshal(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize settings: %w", err)
		}
		updates = append(updates, fmt.Sprintf("settings = $%d", argIndex))
		args = append(args, settingsJSON)
		argIndex++
	}

	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize metadata: %w", err)
		}
		updates = append(updates, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	if len(updates) == 0 {
		// Nothing to update
		return s.Get(id)
	}

	// Execute update
	query := fmt.Sprintf("UPDATE organizations SET %s WHERE id = $%d",
		joinUpdates(updates), argIndex)
	args = append(args, id)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return s.Get(id)
}

// Delete deletes an organization.
func (s *OrgService) Delete(id string) error {
	if s.db == nil {
		return ErrNoDatabase
	}

	// Check if organization exists
	if _, err := s.Get(id); err != nil {
		return err
	}

	// Check if organization has members before deleting
	var memberCount int64
	var query string
	query = `SELECT COUNT(*) FROM user_org_members WHERE organization_id = ?`
	err := s.db.QueryRow(query, id).Scan(&memberCount)
	if err != nil {
		return fmt.Errorf("failed to count members: %w", err)
	}
	if memberCount > 0 {
		return models.NewAppError("CONFLICT", "cannot delete organization with members")
	}

	// Delete organization
	query = `DELETE FROM organizations WHERE id = ?`
	_, err = s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// ListMembers returns all members of an organization.
func (s *OrgService) ListMembers(orgID string) ([]models.UserOrgMember, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Verify organization exists
	_, err := s.Get(orgID)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, user_id, organization_id, role, status, joined_at, metadata 
	          FROM user_org_members WHERE organization_id = ?`
	rows, err := s.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()

	var members []models.UserOrgMember
	for rows.Next() {
		var member models.UserOrgMember
		var metadataJSON []byte
		
		err := rows.Scan(&member.ID, &member.UserID, &member.OrganizationID, 
			&member.Role, &member.Status, &member.JoinedAt, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &member.Metadata)
		}

		members = append(members, member)
	}

	return members, nil
}

// InviteMember invites a new member to an organization.
func (s *OrgService) InviteMember(orgID string, req *models.UserOrgMemberCreate) (*models.UserOrgMember, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	// Verify organization exists
	if _, err := s.Get(orgID); err != nil {
		return nil, err
	}

	member := &models.UserOrgMember{
		ID:             uuid.New().String(),
		UserID:         req.UserID,
		OrganizationID: orgID,
		Role:           req.Role,
		Status:         "pending", // default status for invitations
		JoinedAt:       time.Now(),
		Metadata:       req.Metadata,
	}

	if member.Status == "" {
		member.Status = "pending"
	}

	// Check if user already exists in organization (by email or user_id)
	var existingCount int64
	var err error
	err = s.db.QueryRow(`SELECT COUNT(*) FROM user_org_members 
		WHERE organization_id = ? AND (user_id = ? OR metadata->>'email' = ?)`,		orgID, req.UserID, req.Email).Scan(&existingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing members: %w", err)
	}

	if existingCount > 0 {
		return nil, models.NewAppError("CONFLICT", "member already exists in organization")
	}

	metadataJSON, _ := json.Marshal(req.Metadata)
	query := `INSERT INTO user_org_members (id, user_id, organization_id, role, status, joined_at, metadata) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.Exec(query, member.ID, member.UserID, orgID, member.Role, member.Status, member.JoinedAt, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to invite member: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, models.NewAppError("CONFLICT", "failed to create member")
	}

	return member, nil
}

// GetMember returns a specific member of an organization.
func (s *OrgService) GetMember(orgID string, userID string) (*models.UserOrgMember, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, user_id, organization_id, role, status, joined_at, metadata 
	          FROM user_org_members WHERE organization_id = ? AND user_id = ?`
	var member models.UserOrgMember
	var metadataJSON []byte
	
	err := s.db.QueryRow(query, orgID, userID).Scan(&member.ID, &member.UserID, 
		&member.OrganizationID, &member.Role, &member.Status, &member.JoinedAt, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, models.NewAppError("NOT_FOUND", "member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &member.Metadata)
	}

	return &member, nil
}

// UpdateMemberRole updates a member's role.
func (s *OrgService) UpdateMemberRole(orgID string, userID string, req *models.MemberRoleUpdate) (*models.UserOrgMember, error) {
	if s.db == nil {
		return nil, ErrNoDatabase
	}

	query := `SELECT id, user_id, organization_id, role, status, joined_at, metadata 
	          FROM user_org_members WHERE organization_id = ? AND user_id = ?`
	var member models.UserOrgMember
	var metadataJSON []byte
	
	err := s.db.QueryRow(query, orgID, userID).Scan(&member.ID, &member.UserID, 
		&member.OrganizationID, &member.Role, &member.Status, &member.JoinedAt, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, models.NewAppError("NOT_FOUND", "member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &member.Metadata)
	}

	member.Role = req.Role
	
	updateQuery := `UPDATE user_org_members SET role = ?, updated_at = ? WHERE organization_id = ? AND user_id = ?`
	result, err := s.db.Exec(updateQuery, member.Role, time.Now(), orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, models.NewAppError("NOT_FOUND", "member not found")
	}

	return &member, nil
}

// RemoveMember removes a member from an organization.
func (s *OrgService) RemoveMember(orgID string, userID string) error {
	if s.db == nil {
		return ErrNoDatabase
	}

	query := `DELETE FROM user_org_members WHERE organization_id = ? AND user_id = ?`
	result, err := s.db.Exec(query, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.NewAppError("NOT_FOUND", "member not found")
	}

	return nil
}

// HasUserService checks if a user service is available for fetching user details.
// This is a stub method that always returns false for now.
func (s *OrgService) HasUserService() bool {
	return false
}

// GetUserByID fetches user details by user ID.
// This is a stub method that always returns nil for now.
// In the future, this would query a users table or external user service.
func (s *OrgService) GetUserByID(userID string) (*models.User, error) {
	return nil, models.NewAppError("NOT_FOUND", "user service not available")
}

// Helper functions

func generateSlug(name string) string {
	// Simple slug generation - can be enhanced
	slug := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c + 32)
		} else if c == ' ' {
			slug += "-"
		}
	}
	return slug
}

func joinUpdates(updates []string) string {
	result := ""
	for i, u := range updates {
		if i > 0 {
			result += ", "
		}
		result += u
	}
	return result
}
