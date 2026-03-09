# Organization-Based Filtering

## Overview

The Agent Orchestrator API supports multi-tenancy through organization-based isolation. All resources (agents, TODOs, cron jobs) can be scoped to a specific organization, ensuring data separation and security.

## Headers

### X-Organization-ID

**Header Name:** `X-Organization-ID`

**Description:** Optional header to filter resources by organization

**Type:** String (UUID)

**Required:** No (defaults to returning all resources)

**Example:** `X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000`

### X-Owner-ID

**Header Name:** `X-Owner-ID`

**Description:** Required header when creating organizations

**Type:** String (UUID)

**Required:** Yes (for organization creation)

**Example:** `X-Owner-ID: 987e6543-e21b-43d3-b876-532414198111`

## Filtering Behavior

### With Organization ID

When `X-Organization-ID` is provided:

```http
GET /api/agents
X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000
```

Returns only resources belonging to the specified organization.

### Without Organization ID

When `X-Organization-ID` is NOT provided:

```http
GET /api/agents
```

Returns ALL resources (useful for superadmins or debugging).

### Empty Organization ID

An empty string for `X-Organization-ID` behaves the same as no header:

```http
GET /api/agents
X-Organization-ID:
```

Returns ALL resources.

## Examples

### 1. List Agents in an Organization

**Request:**
```bash
curl -X GET http://localhost:18765/api/agents \
  -H "X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000"
```

**Response:**
```json
[
  {
    "id": "agent-1",
    "name": "Production Agent",
    "status": "idle",
    "organization_id": "123e4567-e89b-12d3-a456-426614174000",
    "created_at": "2026-03-09T10:00:00Z",
    "updated_at": "2026-03-09T10:00:00Z"
  }
]
```

### 2. Create Agent in an Organization

**Request:**
```bash
curl -X POST http://localhost:18765/api/agents \
  -H "Content-Type: application/json" \
  -H "X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000" \
  -d '{
    "name": "New Agent",
    "model": "gpt-4o"
  }'
```

**Response:**
```json
{
  "id": "agent-2",
  "name": "New Agent",
  "status": "idle",
  "model": "gpt-4o",
  "organization_id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2026-03-09T14:00:00Z",
  "updated_at": "2026-03-09T14:00:00Z"
}
```

### 3. List TODOs in an Organization

**Request:**
```bash
curl -X GET http://localhost:18765/api/todos \
  -H "X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000"
```

**Response:**
```json
[
  {
    "id": "todo-1",
    "title": "Process data",
    "status": "pending",
    "priority": 5,
    "organization_id": "123e4567-e89b-12d3-a456-426614174000",
    "created_at": "2026-03-09T11:00:00Z"
  }
]
```

### 4. List Cron Jobs in an Organization

**Request:**
```bash
curl -X GET http://localhost:18765/api/cron \
  -H "X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000"
```

**Response:**
```json
[
  {
    "id": "cron-1",
    "name": "Daily Backup",
    "schedule": "0 0 * * *",
    "status": "active",
    "agent_id": "agent-1",
    "organization_id": "123e4567-e89b-12d3-a456-426614174000",
    "created_at": "2026-03-09T09:00:00Z"
  }
]
```

### 5. Create Organization

**Request:**
```bash
curl -X POST http://localhost:18765/api/organizations \
  -H "Content-Type: application/json" \
  -H "X-Owner-ID: user-123" \
  -d '{
    "name": "My Company",
    "slug": "my-company",
    "settings": {
      "max_agents": 10
    }
  }'
```

**Response:**
```json
{
  "id": "org-1",
  "name": "My Company",
  "slug": "my-company",
  "owner_id": "user-123",
  "settings": {
    "max_agents": 10
  },
  "created_at": "2026-03-09T08:00:00Z",
  "updated_at": "2026-03-09T08:00:00Z"
}
```

### 6. Switch Active Organization

**Request:**
```bash
curl -X POST http://localhost:18765/api/organizations/switch \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

**Response:**
```json
{
  "message": "Organization switched successfully",
  "organization": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My Company",
    "slug": "my-company"
  },
  "organizationId": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Supported Endpoints

The following endpoints support organization filtering via `X-Organization-ID` header:

- `GET /api/agents` - List agents
- `POST /api/agents` - Create agent (inherits from header if not specified in body)
- `PUT /api/agents/:id` - Update agent
- `GET /api/todos` - List TODOs
- `POST /api/todos` - Create TODO (inherits from header if not specified in body)
- `PUT /api/todos/:id` - Update TODO
- `GET /api/cron` - List cron jobs
- `POST /api/cron` - Create cron job (inherits from header if not specified in body)
- `PUT /api/cron/:id` - Update cron job

## Database Schema

### Organizations Table

```sql
CREATE TABLE organizations (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  owner_id TEXT NOT NULL,
  settings TEXT,
  metadata TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Agents Table

```sql
CREATE TABLE agents (
  -- ... other columns ...
  organization_id TEXT,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);
```

### TODOs Table

```sql
CREATE TABLE todos (
  -- ... other columns ...
  organization_id TEXT,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);
```

### Cron Jobs Table

```sql
CREATE TABLE cron_jobs (
  -- ... other columns ...
  organization_id TEXT,
  FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);
```

## Best Practices

1. **Always send organization ID** in production to ensure data isolation
2. **Validate ownership** before allowing users to access/modify resources
3. **Use slug-based URLs** for human-readable organization references
4. **Handle 404s** gracefully when organization doesn't exist
5. **Cache organization context** on the client side for performance
6. **Implement RBAC** at the application level for fine-grained permissions

## Security Considerations

- Without `X-Organization-ID`, all resources are returned (useful for superadmins)
- Implement proper authentication to prevent unauthorized organization access
- Consider implementing role-based access control (RBAC) for organizations
- Validate that users have permission to access the specified organization
- Log organization context changes for audit purposes

## Testing

See test files for examples:
- `internal/services/org_service_test.go` - Organization service tests
- `internal/api/handlers/org_test.go` - HTTP handler tests
- `internal/services/org_filtering_test.go` - Filtering behavior tests

## Troubleshooting

### Q: Why am I seeing resources from multiple organizations?

**A:** You're not sending the `X-Organization-ID` header. Include it to filter results.

### Q: Can I move a resource between organizations?

**A:** Yes, update the `organization_id` field via the appropriate API endpoint.

### Q: What happens when an organization is deleted?

**A:** The `organization_id` on associated resources is set to NULL. The resources remain but lose their organization context.

### Q: Can a user belong to multiple organizations?

**A:** Yes, users can be members of multiple organizations. Use the `GET /api/organizations/owner/{ownerId}` endpoint to list all organizations for a user.
