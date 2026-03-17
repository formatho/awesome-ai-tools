package services

import (
	"database/sql"
	"testing"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "Failed to create test database")

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			owner_id TEXT NOT NULL,
			settings TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS user_org_members (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			role TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err, "Failed to create test tables")

	return db
}

// TestOrgService_Create tests creating a new organization
func TestOrgService_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	t.Run("Create valid organization", func(t *testing.T) {
		req := &models.OrganizationCreate{
			Name: "Test Organization",
			Slug: "test-org",
			Settings: map[string]interface{}{
				"max_agents": 10,
			},
			Metadata: map[string]interface{}{
				"trial": true,
			},
		}

		org, err := svc.Create(req, "user-123")
		require.NoError(t, err, "Failed to create organization")
		require.NotNil(t, org)

		assert.Equal(t, "Test Organization", org.Name)
		assert.Equal(t, "test-org", org.Slug)
		assert.Equal(t, "user-123", org.OwnerID)
		assert.NotNil(t, org.ID)
		assert.NotNil(t, org.CreatedAt)
		assert.NotNil(t, org.UpdatedAt)
		assert.Equal(t, float64(10), org.Settings["max_agents"])
		assert.True(t, org.Metadata["trial"].(bool))

		t.Logf("✅ Organization created: ID=%s, Name=%s", org.ID, org.Name)
	})

	t.Run("Create organization with auto-generated slug", func(t *testing.T) {
		req := &models.OrganizationCreate{
			Name: "My Awesome Organization",
		}

		org, err := svc.Create(req, "user-456")
		require.NoError(t, err)

		assert.Equal(t, "My Awesome Organization", org.Name)
		assert.Equal(t, "my-awesome-organization", org.Slug)

		t.Logf("✅ Auto-generated slug: %s", org.Slug)
	})

	t.Run("Create organization with empty name", func(t *testing.T) {
		req := &models.OrganizationCreate{
			Name: "",
		}

		_, err := svc.Create(req, "user-789")
		assert.Error(t, err, "Expected error for empty name")
		assert.Contains(t, err.Error(), "name is required")

		t.Logf("✅ Validation working: empty name rejected")
	})

	t.Run("Create organization with duplicate slug", func(t *testing.T) {
		req1 := &models.OrganizationCreate{
			Name: "First Org",
			Slug: "duplicate-slug",
		}
		_, err := svc.Create(req1, "user-111")
		require.NoError(t, err)

		req2 := &models.OrganizationCreate{
			Name: "Second Org",
			Slug: "duplicate-slug",
		}
		_, err = svc.Create(req2, "user-222")
		assert.Error(t, err, "Expected error for duplicate slug")

		t.Logf("✅ Duplicate slug rejected")
	})
}

// TestOrgService_List tests listing organizations
func TestOrgService_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	// Create test organizations
	orgs := []*models.OrganizationCreate{
		{Name: "Org 1", Slug: "org-1"},
		{Name: "Org 2", Slug: "org-2"},
		{Name: "Org 3", Slug: "org-3"},
	}

	for _, org := range orgs {
		_, err := svc.Create(org, "user-123")
		require.NoError(t, err)
	}

	t.Run("List all organizations", func(t *testing.T) {
		result, err := svc.List()
		require.NoError(t, err)
		require.Len(t, result, 3)

		t.Logf("✅ Listed %d organizations", len(result))
	})
}

// TestOrgService_Get tests retrieving an organization by ID
func TestOrgService_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	t.Run("Get existing organization", func(t *testing.T) {
		req := &models.OrganizationCreate{
			Name:     "Get Test Org",
			Slug:     "get-test",
			Settings: map[string]interface{}{"key": "value"},
		}

		created, err := svc.Create(req, "user-123")
		require.NoError(t, err)

		fetched, err := svc.Get(created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, "Get Test Org", fetched.Name)
		assert.Equal(t, "get-test", fetched.Slug)
		assert.Equal(t, "value", fetched.Settings["key"])

		t.Logf("✅ Retrieved organization: %s", fetched.Name)
	})

	t.Run("Get non-existent organization", func(t *testing.T) {
		_, err := svc.Get("non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		t.Logf("✅ Non-existent org rejected")
	})
}

// TestOrgService_GetBySlug tests retrieving an organization by slug
func TestOrgService_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	req := &models.OrganizationCreate{
		Name: "Slug Test Org",
		Slug: "slug-test-org",
	}
	created, err := svc.Create(req, "user-123")
	require.NoError(t, err)

	t.Run("Get by valid slug", func(t *testing.T) {
		fetched, err := svc.GetBySlug("slug-test-org")
		require.NoError(t, err)

		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, "Slug Test Org", fetched.Name)

		t.Logf("✅ Retrieved by slug: %s", fetched.Name)
	})

	t.Run("Get by invalid slug", func(t *testing.T) {
		_, err := svc.GetBySlug("invalid-slug")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		t.Logf("✅ Invalid slug rejected")
	})
}

// TestOrgService_GetByOwner tests retrieving organizations by owner ID
func TestOrgService_GetByOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	// Create orgs for different owners
	_, err := svc.Create(&models.OrganizationCreate{Name: "User1 Org1"}, "user-1")
	require.NoError(t, err)
	_, err = svc.Create(&models.OrganizationCreate{Name: "User1 Org2"}, "user-1")
	require.NoError(t, err)
	_, err = svc.Create(&models.OrganizationCreate{Name: "User2 Org1"}, "user-2")
	require.NoError(t, err)

	t.Run("Get organizations for owner", func(t *testing.T) {
		orgs, err := svc.GetByOwner("user-1")
		require.NoError(t, err)
		require.Len(t, orgs, 2)

		t.Logf("✅ Retrieved %d organizations for user-1", len(orgs))
	})

	t.Run("Get organizations for non-existent owner", func(t *testing.T) {
		orgs, err := svc.GetByOwner("user-999")
		require.NoError(t, err)
		require.Len(t, orgs, 0)

		t.Logf("✅ Empty list for non-existent owner")
	})
}

// TestOrgService_Update tests updating an organization
func TestOrgService_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	created, err := svc.Create(&models.OrganizationCreate{
		Name:     "Original Name",
		Slug:     "original-slug",
		Settings: map[string]interface{}{"old": "value"},
	}, "user-123")
	require.NoError(t, err)

	t.Run("Update name", func(t *testing.T) {
		newName := "Updated Name"
		update := &models.OrganizationUpdate{
			Name: &newName,
		}

		updated, err := svc.Update(created.ID, update)
		require.NoError(t, err)

		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "original-slug", updated.Slug) // Slug unchanged

		t.Logf("✅ Updated name to: %s", updated.Name)
	})

	t.Run("Update settings", func(t *testing.T) {
		newSettings := map[string]interface{}{
			"new_key":  "new_value",
			"max_users": 100,
		}
		update := &models.OrganizationUpdate{
			Settings: newSettings,
		}

		updated, err := svc.Update(created.ID, update)
		require.NoError(t, err)

		assert.Equal(t, "new_value", updated.Settings["new_key"])
		assert.Equal(t, float64(100), updated.Settings["max_users"])

		t.Logf("✅ Updated settings")
	})

	t.Run("Update non-existent organization", func(t *testing.T) {
		newName := "Non-existent"
		update := &models.OrganizationUpdate{
			Name: &newName,
		}

		_, err := svc.Update("non-existent-id", update)
		assert.Error(t, err)

		t.Logf("✅ Non-existent org update rejected")
	})
}

// TestOrgService_Delete tests deleting an organization
func TestOrgService_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewOrgService(db)

	created, err := svc.Create(&models.OrganizationCreate{
		Name: "To Be Deleted",
		Slug: "delete-test",
	}, "user-123")
	require.NoError(t, err)

	t.Run("Delete existing organization", func(t *testing.T) {
		err := svc.Delete(created.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = svc.Get(created.ID)
		assert.Error(t, err)

		t.Logf("✅ Deleted organization: %s", created.Name)
	})

	t.Run("Delete non-existent organization", func(t *testing.T) {
		err := svc.Delete("non-existent-id")
		assert.Error(t, err)

		t.Logf("✅ Non-existent org deletion rejected")
	})
}

// TestOrgService_Errors tests error handling without database
func TestOrgService_Errors(t *testing.T) {
	svc := NewOrgService(nil)

	t.Run("List without database", func(t *testing.T) {
		_, err := svc.List()
		assert.Equal(t, ErrNoDatabase, err)

		t.Logf("✅ NoDatabase error handled")
	})

	t.Run("Create without database", func(t *testing.T) {
		req := &models.OrganizationCreate{Name: "Test"}
		_, err := svc.Create(req, "user-123")
		assert.Equal(t, ErrNoDatabase, err)

		t.Logf("✅ NoDatabase error handled")
	})
}

func TestMain(m *testing.M) {
	m.Run()
}
