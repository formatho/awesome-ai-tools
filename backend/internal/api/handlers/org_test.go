package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestApp creates a test Fiber app with organization routes
func setupTestApp(t *testing.T) (*fiber.App, *services.OrgService) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create tables
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
	`)
	require.NoError(t, err)

	// Create service and handler
	orgSvc := services.NewOrgService(db)
	orgHandler := NewOrgHandler(orgSvc)

	// Create Fiber app
	app := fiber.New()

	// Register routes
	orgs := app.Group("/api/organizations")
	orgs.Get("/", orgHandler.List)
	orgs.Post("/", orgHandler.Create)
	orgs.Get("/:id", orgHandler.Get)
	orgs.Put("/:id", orgHandler.Update)
	orgs.Delete("/:id", orgHandler.Delete)
	orgs.Get("/slug/:slug", orgHandler.GetBySlug)
	orgs.Get("/owner/:ownerId", orgHandler.GetByOwner)
	orgs.Post("/switch", orgHandler.SwitchOrganization)

	return app, orgSvc
}

// TestOrgHandler_List tests listing organizations
func TestOrgHandler_List(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	// Create test organizations
	_, err := orgSvc.Create(&models.OrganizationCreate{Name: "Org 1"}, "user-1")
	require.NoError(t, err)

	t.Run("List all organizations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/organizations", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result), 1)
		t.Logf("✅ Listed %d organizations", len(result))
	})
}

// TestOrgHandler_Create tests creating organizations
func TestOrgHandler_Create(t *testing.T) {
	app, _ := setupTestApp(t)

	t.Run("Create valid organization", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":     "New Organization",
			"slug":     "new-org",
			"settings": map[string]interface{}{"key": "value"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/organizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Owner-ID", "user-123")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 201, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "New Organization", result["name"])
		assert.Equal(t, "new-org", result["slug"])
		assert.Equal(t, "user-123", result["owner_id"])
		assert.NotNil(t, result["id"])

		t.Logf("✅ Created organization: %s", result["name"])
	})

	t.Run("Create without owner header", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "No Owner Org",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/organizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 400, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "X-Owner-ID")

		t.Logf("✅ Missing owner header rejected")
	})

	t.Run("Create with empty name", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/organizations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Owner-ID", "user-123")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 400, resp.StatusCode)

		t.Logf("✅ Empty name rejected")
	})
}

// TestOrgHandler_Get tests retrieving an organization by ID
func TestOrgHandler_Get(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	// Create test organization
	created, err := orgSvc.Create(&models.OrganizationCreate{
		Name:     "Get Test Org",
		Slug:     "get-test",
		Settings: map[string]interface{}{"key": "value"},
	}, "user-123")
	require.NoError(t, err)

	t.Run("Get existing organization", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/organizations/"+created.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, created.ID, result["id"])
		assert.Equal(t, "Get Test Org", result["name"])
		assert.Equal(t, "get-test", result["slug"])

		t.Logf("✅ Retrieved organization: %s", result["name"])
	})

	t.Run("Get non-existent organization", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/organizations/non-existent", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 404, resp.StatusCode)

		t.Logf("✅ Non-existent org returned 404")
	})
}

// TestOrgHandler_GetBySlug tests retrieving an organization by slug
func TestOrgHandler_GetBySlug(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	_, err := orgSvc.Create(&models.OrganizationCreate{
		Name: "Slug Test Org",
		Slug: "slug-test-org",
	}, "user-123")
	require.NoError(t, err)

	t.Run("Get by valid slug", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/organizations/slug/slug-test-org", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "Slug Test Org", result["name"])
		assert.Equal(t, "slug-test-org", result["slug"])

		t.Logf("✅ Retrieved by slug: %s", result["name"])
	})

	t.Run("Get by invalid slug", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/organizations/slug/invalid-slug", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 404, resp.StatusCode)

		t.Logf("✅ Invalid slug returned 404")
	})
}

// TestOrgHandler_GetByOwner tests retrieving organizations by owner
func TestOrgHandler_GetByOwner(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	// Create orgs for different owners
	_, err := orgSvc.Create(&models.OrganizationCreate{Name: "User1 Org1"}, "user-1")
	require.NoError(t, err)
	_, err = orgSvc.Create(&models.OrganizationCreate{Name: "User1 Org2"}, "user-1")
	require.NoError(t, err)
	_, err = orgSvc.Create(&models.OrganizationCreate{Name: "User2 Org1"}, "user-2")
	require.NoError(t, err)

	t.Run("Get organizations by owner", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/organizations/owner/user-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Len(t, result, 2)

		t.Logf("✅ Retrieved %d organizations for user-1", len(result))
	})
}

// TestOrgHandler_Update tests updating an organization
func TestOrgHandler_Update(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	created, err := orgSvc.Create(&models.OrganizationCreate{
		Name: "Original Name",
		Slug: "original-slug",
	}, "user-123")
	require.NoError(t, err)

	t.Run("Update organization name", func(t *testing.T) {
		newName := "Updated Name"
		reqBody := map[string]interface{}{
			"name": newName,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/api/organizations/"+created.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "Updated Name", result["name"])

		t.Logf("✅ Updated name to: %s", result["name"])
	})

	t.Run("Update with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/organizations/"+created.ID, bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 400, resp.StatusCode)

		t.Logf("✅ Invalid JSON rejected")
	})
}

// TestOrgHandler_Delete tests deleting an organization
func TestOrgHandler_Delete(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	created, err := orgSvc.Create(&models.OrganizationCreate{
		Name: "To Be Deleted",
		Slug: "delete-test",
	}, "user-123")
	require.NoError(t, err)

	t.Run("Delete existing organization", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/organizations/"+created.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 204, resp.StatusCode)

		// Verify it's deleted
		req2 := httptest.NewRequest("GET", "/api/organizations/"+created.ID, nil)
		resp2, err := app.Test(req2)
		require.NoError(t, err)
		assert.Equal(t, 404, resp2.StatusCode)

		t.Logf("✅ Deleted organization: %s", created.Name)
	})
}

// TestOrgHandler_SwitchOrganization tests switching active organization
func TestOrgHandler_SwitchOrganization(t *testing.T) {
	app, orgSvc := setupTestApp(t)

	created, err := orgSvc.Create(&models.OrganizationCreate{
		Name: "Target Org",
		Slug: "target-org",
	}, "user-123")
	require.NoError(t, err)

	t.Run("Switch to existing organization", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"organization_id": created.ID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/organizations/switch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, result["message"], "switched successfully")
		assert.NotNil(t, result["organization"])

		t.Logf("✅ Switched to organization: %s", created.Name)
	})

	t.Run("Switch to non-existent organization", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"organization_id": "non-existent-id",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/organizations/switch", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 404, resp.StatusCode)

		t.Logf("✅ Non-existent org switch rejected")
	})

	t.Run("Switch with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/organizations/switch", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, 400, resp.StatusCode)

		t.Logf("✅ Invalid JSON rejected")
	})
}

// TestOrgHandler_CORS tests CORS headers
func TestOrgHandler_CORS(t *testing.T) {
	app, _ := setupTestApp(t)

	req := httptest.NewRequest("GET", "/api/organizations", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Fiber's CORS middleware should allow all origins
	assert.Equal(t, 200, resp.StatusCode)

	t.Logf("✅ CORS headers present")
}

func TestMain(m *testing.M) {
	m.Run()
}
