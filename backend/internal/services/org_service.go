// Package services provides business logic layer for the API.
package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

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
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", cerr)
		}
	}()

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
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", cerr)
		}
	}()

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
	_, err := s.Get(id)
	if err != nil {
		return err
	}

	// Delete organization
	query := `DELETE FROM organizations WHERE id = ?`
	_, err = s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
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
