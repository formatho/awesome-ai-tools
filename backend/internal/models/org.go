// Package models defines the data structures used by the API.
package models

import (
	"time"
)

// Organization represents an organization in the system.
type Organization struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	OwnerID   string                 `json:"owner_id"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// OrganizationCreate is the request body for creating a new organization.
type OrganizationCreate struct {
	Name     string                 `json:"name"`
	Slug     string                 `json:"slug,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OrganizationUpdate is the request body for updating an organization.
type OrganizationUpdate struct {
	Name     *string                `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OrganizationSwitch is the request body for switching the active organization.
type OrganizationSwitch struct {
	OrganizationID string `json:"organization_id"`
}

// Validate validates the organization creation request.
func (o *OrganizationCreate) Validate() error {
	if o.Name == "" {
		return ErrValidation("name is required")
	}
	return nil
}
