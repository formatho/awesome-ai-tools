# Agent Orchestrator - API Documentation

## Base URL

```
Development: http://localhost:18765/api
Production:  https://api.formatho.com/api
```

## Overview

The Agent Orchestrator API provides endpoints for managing AI agents, TODO queues, cron jobs, configurations, and organizations. The API is RESTful and uses JSON for request/response bodies.

## Authentication

**Status:** Authentication is not yet implemented. All endpoints are currently public.

**Planned:** JWT-based Bearer token authentication

```http
Authorization: Bearer <token>
```

## Headers

### Common Headers

| Header | Description | Required |
|--------|-------------|-----------|
| `Content-Type` | Request content type | Yes (for POST/PUT) |
| `X-Organization-ID` | Organization ID for filtering | No |
| `X-Owner-ID` | Owner user ID (for org creation) | Yes (for POST /organizations) |

### Example

```http
GET /api/agents
X-Organization-ID: 123e4567-e89b-12d3-a456-426614174000
Content-Type: application/json
```

## Response Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 204 | No Content - Successful with no return body |
| 400 | Bad Request - Invalid input |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error - Server error |

## Endpoints

### Organizations

#### List Organizations

```http
GET /api/organizations
```

**Response:** Array of organizations

#### Create Organization

```http
POST /api/organizations
X-Owner-ID: <owner-id>
Content-Type: application/json

{
  "name": "My Company",
  "slug": "my-company",
  "settings": {},
  "metadata": {}
}
```

#### Get Organization

```http
GET /api/organizations/{id}
```

#### Update Organization

```http
PUT /api/organizations/{id}
Content-Type: application/json

{
  "name": "Updated Name",
  "settings": {}
}
```

#### Delete Organization

```http
DELETE /api/organizations/{id}
```

#### Get Organization by Slug

```http
GET /api/organizations/slug/{slug}
```

#### Get Organizations by Owner

```http
GET /api/organizations/owner/{ownerId}
```

#### Switch Organization

```http
POST /api/organizations/switch
Content-Type: application/json

{
  "organization_id": "<org-id>"
}
```

### Agents

#### List Agents

```http
GET /api/agents
X-Organization-ID: <org-id>  # Optional filter
```

**Response:** Array of agents

#### Create Agent

```http
POST /api/agents
X-Organization-ID: <org-id>  # Optional filter
Content-Type: application/json

{
  "name": "My Agent",
  "provider": "openai",
  "model": "gpt-4o",
  "system_prompt": "You are a helpful assistant.",
  "work_dir": "~/sandbox",
  "organization_id": "<org-id>",  # Optional
  "config": {},
  "metadata": {}
}
```

#### Get Agent

```http
GET /api/agents/{id}
```

#### Update Agent

```http
PUT /api/agents/{id}
Content-Type: application/json

{
  "name": "Updated Agent Name",
  "model": "gpt-4o",
  "system_prompt": "Updated prompt"
}
```

#### Delete Agent

```http
DELETE /api/agents/{id}
```

#### Start Agent

```http
POST /api/agents/{id}/start
```

#### Stop Agent

```http
POST /api/agents/{id}/stop
```

#### Pause Agent

```http
POST /api/agents/{id}/pause
```

#### Resume Agent

```http
POST /api/agents/{id}/resume
```

### TODOs

#### List TODOs

```http
GET /api/todos
X-Organization-ID: <org-id>  # Optional filter
```

**Response:** Array of TODOs

#### Create TODO

```http
POST /api/todos
X-Organization-ID: <org-id>  # Optional filter
Content-Type: application/json

{
  "title": "Process data",
  "description": "Process the uploaded CSV file",
  "priority": 5,
  "agent_id": "<agent-id>",
  "organization_id": "<org-id>",  # Optional
  "skills": ["python", "data-analysis"],
  "dependencies": [],
  "config": {}
}
```

#### Get TODO

```http
GET /api/todos/{id}
```

#### Update TODO

```http
PUT /api/todos/{id}
Content-Type: application/json

{
  "title": "Updated title",
  "status": "completed",
  "progress": 100
}
```

#### Delete TODO

```http
DELETE /api/todos/{id}
```

#### Start TODO

```http
POST /api/todos/{id}/start
```

#### Cancel TODO

```http
POST /api/todos/{id}/cancel
```

### Cron Jobs

#### List Cron Jobs

```http
GET /api/cron
X-Organization-ID: <org-id>  # Optional filter
```

**Response:** Array of cron jobs

#### Create Cron Job

```http
POST /api/cron
X-Organization-ID: <org-id>  # Optional filter
Content-Type: application/json

{
  "name": "Daily Backup",
  "schedule": "0 0 * * *",
  "timezone": "UTC",
  "agent_id": "<agent-id>",
  "organization_id": "<org-id>",  # Optional
  "task_name": "backup",
  "task_config": {}
}
```

#### Get Cron Job

```http
GET /api/cron/{id}
```

#### Update Cron Job

```http
PUT /api/cron/{id}
Content-Type: application/json

{
  "name": "Updated Job Name",
  "schedule": "0 1 * * *"
}
```

#### Delete Cron Job

```http
DELETE /api/cron/{id}
```

#### Pause Cron Job

```http
POST /api/cron/{id}/pause
```

#### Resume Cron Job

```http
POST /api/cron/{id}/resume
```

#### Get Cron History

```http
GET /api/cron/{id}/history
```

### Configuration

#### Get Configuration

```http
GET /api/config
```

#### Update Configuration

```http
PUT /api/config
Content-Type: application/json

{
  "llm_config": {},
  "defaults": {},
  "settings": {}
}
```

#### Test LLM Configuration

```http
POST /api/config/test-llm
Content-Type: application/json

{
  "provider": "openai",
  "model": "gpt-4o",
  "api_key": "<api-key>"
}
```

### System

#### Get System Status

```http
GET /api/system/status
```

**Response:**
```json
{
  "uptime_seconds": 3600,
  "started_at": "2026-03-09T12:00:00Z",
  "version": "1.0.0",
  "counts": {
    "agents": 5,
    "todos": 10,
    "crons": 3
  },
  "resources": {
    "goroutines": 15,
    "memory_mb": 128,
    "num_cpu": 8
  }
}
```

#### Health Check

```http
GET /api/system/health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-03-09T12:00:00Z"
}
```

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1234567890
}
```

## WebSocket Connection

### Connect to WebSocket

```javascript
const ws = new WebSocket('ws://localhost:18765/ws');
```

### Events

The WebSocket broadcasts real-time updates for:

- Agent status changes
- TODO status updates
- Cron job executions
- Agent creations/deletions

### Example Messages

```javascript
// Agent status update
{
  "type": "agent_status",
  "agent_id": "agent-1",
  "status": "running"
}

// TODO status update
{
  "type": "todo_status",
  "todo_id": "todo-1",
  "status": "running"
}
```

## Data Models

### Organization

```json
{
  "id": "uuid",
  "name": "Organization Name",
  "slug": "organization-name",
  "owner_id": "owner-uuid",
  "settings": {},
  "metadata": {},
  "created_at": "2026-03-09T00:00:00Z",
  "updated_at": "2026-03-09T00:00:00Z"
}
```

### Agent

```json
{
  "id": "uuid",
  "name": "Agent Name",
  "status": "idle|running|paused|stopped|error",
  "provider": "openai|anthropic|ollama",
  "model": "gpt-4o",
  "system_prompt": "System prompt text",
  "work_dir": "~/sandbox",
  "organization_id": "org-uuid",
  "config": {},
  "metadata": {},
  "created_at": "2026-03-09T00:00:00Z",
  "updated_at": "2026-03-09T00:00:00Z",
  "started_at": "2026-03-09T00:00:00Z",
  "stopped_at": "2026-03-09T00:00:00Z",
  "error": ""
}
```

### TODO

```json
{
  "id": "uuid",
  "title": "Task title",
  "description": "Task description",
  "status": "pending|queued|running|completed|failed|cancelled",
  "priority": 5,
  "progress": 50,
  "agent_id": "agent-uuid",
  "organization_id": "org-uuid",
  "skills": ["skill1", "skill2"],
  "dependencies": ["todo-id-1"],
  "config": {},
  "result": {},
  "error": "",
  "created_at": "2026-03-09T00:00:00Z",
  "updated_at": "2026-03-09T00:00:00Z",
  "started_at": "2026-03-09T00:00:00Z",
  "completed_at": "2026-03-09T00:00:00Z"
}
```

### Cron Job

```json
{
  "id": "uuid",
  "name": "Job Name",
  "schedule": "0 0 * * *",
  "timezone": "UTC",
  "status": "active|paused|disabled",
  "agent_id": "agent-uuid",
  "organization_id": "org-uuid",
  "task_name": "task-name",
  "task_config": {},
  "last_run_at": "2026-03-09T00:00:00Z",
  "next_run_at": "2026-03-10T00:00:00Z",
  "last_result": "",
  "last_error": "",
  "run_count": 10,
  "success_count": 9,
  "fail_count": 1,
  "created_at": "2026-03-09T00:00:00Z",
  "updated_at": "2026-03-09T00:00:00Z"
}
```

## Error Handling

All errors follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common Errors

- `"name is required"` - Missing required field
- `"organization not found"` - Invalid organization ID
- `"agent not found"` - Invalid agent ID
- `"todo not found"` - Invalid TODO ID
- `"cron job not found"` - Invalid cron job ID

## Rate Limiting

**Status:** Not yet implemented

**Planned:** Rate limiting per user/IP

## Pagination

**Status:** Not yet implemented

**Current:** All endpoints return full lists

**Planned:** Support for `page` and `limit` parameters

## Filtering

**Implemented:** Organization-based filtering via `X-Organization-ID` header

**Planned:** Additional filtering options (status, date range, etc.)

## Sorting

**Current:** Default sorting by `created_at` DESC

**Planned:** Customizable sort parameters

## Versioning

**Current Version:** v1.0.0

**Versioning Strategy:** URL path versioning (e.g., `/api/v1/organizations`)

## Support

- **Documentation:** https://docs.formatho.com
- **GitHub:** https://github.com/formatho/agent-orchestrator
- **Email:** support@formatho.com

## Changelog

### v1.0.0 (2026-03-09)
- Initial API release
- Agent management
- TODO queue management
- Cron job scheduling
- Organization support with filtering
- WebSocket real-time updates
