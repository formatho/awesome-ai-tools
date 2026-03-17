package services

import (
	"database/sql"
	"testing"

	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupOrgFilteringDB creates a test database with organizations and resources
func setupOrgFilteringDB(t *testing.T) (*sql.DB, string, string, string, string) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create all tables
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

		CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'idle',
			provider TEXT,
			model TEXT,
			system_prompt TEXT,
			base_url TEXT DEFAULT '',
			work_dir TEXT DEFAULT '~/sandbox',
			organization_id TEXT,
			config TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			stopped_at DATETIME,
			error TEXT
		);

		CREATE TABLE IF NOT EXISTS todos (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			priority INTEGER NOT NULL DEFAULT 0,
			progress INTEGER NOT NULL DEFAULT 0,
			agent_id TEXT,
			organization_id TEXT,
			skills TEXT,
			dependencies TEXT,
			config TEXT,
			result TEXT,
			error TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS cron_jobs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			schedule TEXT NOT NULL,
			timezone TEXT DEFAULT 'UTC',
			status TEXT NOT NULL DEFAULT 'active',
			agent_id TEXT NOT NULL,
			organization_id TEXT,
			task_name TEXT,
			task_config TEXT,
			last_run_at DATETIME,
			next_run_at DATETIME,
			last_result TEXT,
			last_error TEXT,
			run_count INTEGER NOT NULL DEFAULT 0,
			success_count INTEGER NOT NULL DEFAULT 0,
			fail_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	// Create organizations
	org1ID := uuid.New().String()
	org2ID := uuid.New().String()

	_, err = db.Exec(`INSERT INTO organizations (id, name, slug, owner_id) VALUES (?, ?, ?, ?)`,
		org1ID, "Org 1", "org-1", "user-1")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO organizations (id, name, slug, owner_id) VALUES (?, ?, ?, ?)`,
		org2ID, "Org 2", "org-2", "user-2")
	require.NoError(t, err)

	// Create agents
	agent1ID := uuid.New().String()
	agent2ID := uuid.New().String()
	agentNoOrgID := uuid.New().String()

	_, err = db.Exec(`INSERT INTO agents (id, name, organization_id) VALUES (?, ?, ?)`,
		agent1ID, "Agent in Org 1", org1ID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO agents (id, name, organization_id) VALUES (?, ?, ?)`,
		agent2ID, "Agent in Org 2", org2ID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO agents (id, name, organization_id) VALUES (?, ?, ?)`,
		agentNoOrgID, "Agent with no org", nil)
	require.NoError(t, err)

	// Create todos
	todo1ID := uuid.New().String()
	todo2ID := uuid.New().String()
	todoNoOrgID := uuid.New().String()

	_, err = db.Exec(`INSERT INTO todos (id, title, organization_id) VALUES (?, ?, ?)`,
		todo1ID, "TODO in Org 1", org1ID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO todos (id, title, organization_id) VALUES (?, ?, ?)`,
		todo2ID, "TODO in Org 2", org2ID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO todos (id, title, organization_id) VALUES (?, ?, ?)`,
		todoNoOrgID, "TODO with no org", nil)
	require.NoError(t, err)

	// Create cron jobs
	cron1ID := uuid.New().String()
	cron2ID := uuid.New().String()
	cronNoOrgID := uuid.New().String()
	agentRefID := agent1ID // Use a valid agent ID

	_, err = db.Exec(`INSERT INTO cron_jobs (id, name, schedule, agent_id, organization_id) VALUES (?, ?, ?, ?, ?)`,
		cron1ID, "Cron in Org 1", "* * * * *", agentRefID, org1ID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO cron_jobs (id, name, schedule, agent_id, organization_id) VALUES (?, ?, ?, ?, ?)`,
		cron2ID, "Cron in Org 2", "* * * * *", agentRefID, org2ID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO cron_jobs (id, name, schedule, agent_id, organization_id) VALUES (?, ?, ?, ?, ?)`,
		cronNoOrgID, "Cron with no org", "* * * * *", agentRefID, nil)
	require.NoError(t, err)

	return db, org1ID, org2ID, agentRefID, agentNoOrgID
}

// TestAgentService_OrganizationFiltering tests organization filtering in agent service
func TestAgentService_OrganizationFiltering(t *testing.T) {
	db, org1ID, org2ID, _, _ := setupOrgFilteringDB(t)
	defer db.Close()

	hub := websocket.NewHub()
	go hub.Run()
	svc := NewAgentService(db, hub)

	t.Run("List all agents (no filter)", func(t *testing.T) {
		agents, err := svc.List(nil)
		require.NoError(t, err)
		assert.Len(t, agents, 3)

		t.Logf("✅ Listed %d agents (no filter)", len(agents))
	})

	t.Run("List agents for Org 1", func(t *testing.T) {
		agents, err := svc.List(&org1ID)
		require.NoError(t, err)
		assert.Len(t, agents, 1)
		assert.Equal(t, "Agent in Org 1", agents[0].Name)
		assert.Equal(t, org1ID, agents[0].OrganizationID)

		t.Logf("✅ Listed %d agents for Org 1", len(agents))
	})

	t.Run("List agents for Org 2", func(t *testing.T) {
		agents, err := svc.List(&org2ID)
		require.NoError(t, err)
		assert.Len(t, agents, 1)
		assert.Equal(t, "Agent in Org 2", agents[0].Name)

		t.Logf("✅ Listed %d agents for Org 2", len(agents))
	})

	t.Run("Create agent with organization", func(t *testing.T) {
		req := &models.AgentCreate{
			Name:           "New Agent in Org 1",
			OrganizationID: org1ID,
		}

		agent, err := svc.Create(req)
		require.NoError(t, err)
		assert.Equal(t, org1ID, agent.OrganizationID)

		// Verify it shows up in Org 1's list
		agents, err := svc.List(&org1ID)
		require.NoError(t, err)
		assert.Len(t, agents, 2)

		t.Logf("✅ Created agent with org_id: %s", agent.OrganizationID)
	})
}

// TestTODOService_OrganizationFiltering tests organization filtering in TODO service
func TestTODOService_OrganizationFiltering(t *testing.T) {
	db, org1ID, org2ID, _, _ := setupOrgFilteringDB(t)
	defer db.Close()

	hub := websocket.NewHub()
	go hub.Run()
	svc := NewTODOService(db, hub)

	t.Run("List all todos (no filter)", func(t *testing.T) {
		todos, err := svc.List(nil)
		require.NoError(t, err)
		assert.Len(t, todos, 3)

		t.Logf("✅ Listed %d todos (no filter)", len(todos))
	})

	t.Run("List todos for Org 1", func(t *testing.T) {
		todos, err := svc.List(&org1ID)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, "TODO in Org 1", todos[0].Title)
		assert.Equal(t, org1ID, todos[0].OrganizationID)

		t.Logf("✅ Listed %d todos for Org 1", len(todos))
	})

	t.Run("List todos for Org 2", func(t *testing.T) {
		todos, err := svc.List(&org2ID)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, "TODO in Org 2", todos[0].Title)

		t.Logf("✅ Listed %d todos for Org 2", len(todos))
	})

	t.Run("Create TODO with organization", func(t *testing.T) {
		req := &models.TODOCreate{
			Title:          "New TODO in Org 1",
			OrganizationID: org1ID,
		}

		todo, err := svc.Create(req)
		require.NoError(t, err)
		assert.Equal(t, org1ID, todo.OrganizationID)

		// Verify it shows up in Org 1's list
		todos, err := svc.List(&org1ID)
		require.NoError(t, err)
		assert.Len(t, todos, 2)

		t.Logf("✅ Created TODO with org_id: %s", todo.OrganizationID)
	})
}

// TestCronService_OrganizationFiltering tests organization filtering in cron service
func TestCronService_OrganizationFiltering(t *testing.T) {
	db, org1ID, org2ID, agentRefID, _ := setupOrgFilteringDB(t)
	defer db.Close()

	hub := websocket.NewHub()
	go hub.Run()
	svc := NewCronService(db, hub)

	t.Run("List all cron jobs (no filter)", func(t *testing.T) {
		crons, err := svc.List(nil)
		require.NoError(t, err)
		assert.Len(t, crons, 3)

		t.Logf("✅ Listed %d cron jobs (no filter)", len(crons))
	})

	t.Run("List cron jobs for Org 1", func(t *testing.T) {
		crons, err := svc.List(&org1ID)
		require.NoError(t, err)
		assert.Len(t, crons, 1)
		assert.Equal(t, "Cron in Org 1", crons[0].Name)
		assert.Equal(t, org1ID, crons[0].OrganizationID)

		t.Logf("✅ Listed %d cron jobs for Org 1", len(crons))
	})

	t.Run("List cron jobs for Org 2", func(t *testing.T) {
		crons, err := svc.List(&org2ID)
		require.NoError(t, err)
		assert.Len(t, crons, 1)
		assert.Equal(t, "Cron in Org 2", crons[0].Name)

		t.Logf("✅ Listed %d cron jobs for Org 2", len(crons))
	})

	t.Run("Create cron job with organization", func(t *testing.T) {
		req := &models.CronCreate{
			Name:           "New Cron in Org 1",
			Schedule:       "0 * * * *",
			AgentID:        agentRefID,
			OrganizationID: org1ID,
		}

		cron, err := svc.Create(req)
		require.NoError(t, err)
		assert.Equal(t, org1ID, cron.OrganizationID)

		// Verify it shows up in Org 1's list
		crons, err := svc.List(&org1ID)
		require.NoError(t, err)
		assert.Len(t, crons, 2)

		t.Logf("✅ Created cron job with org_id: %s", cron.OrganizationID)
	})
}

// TestOrganizationID_EmptyTests tests edge cases with empty organization IDs
func TestOrganizationID_EmptyTests(t *testing.T) {
	db, _, _, _, _ := setupOrgFilteringDB(t)
	defer db.Close()

	hub := websocket.NewHub()
	go hub.Run()
	agentSvc := NewAgentService(db, hub)
	todoSvc := NewTODOService(db, hub)
	cronSvc := NewCronService(db, hub)

	t.Run("List with empty organization ID string", func(t *testing.T) {
		emptyStr := ""
		agents, err := agentSvc.List(&emptyStr)
		require.NoError(t, err)
		assert.Len(t, agents, 3) // Should return all

		todos, err := todoSvc.List(&emptyStr)
		require.NoError(t, err)
		assert.Len(t, todos, 3) // Should return all

		crons, err := cronSvc.List(&emptyStr)
		require.NoError(t, err)
		assert.Len(t, crons, 3) // Should return all

		t.Logf("✅ Empty string returns all resources")
	})
}
