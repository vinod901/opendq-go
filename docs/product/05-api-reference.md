# API Reference

Complete API documentation for OpenDQ.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API requests require authentication via Bearer token:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/...
```

For multi-tenant deployments, include tenant header:

```bash
curl -H "X-Tenant: <tenant-slug>" ...
```

## Response Format

All responses are JSON:

```json
{
  "data": {...},
  "error": null
}
```

Error responses:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

---

## Health

### Health Check

```
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

---

## Tenants

### List Tenants

```
GET /api/v1/tenants
```

**Response:**
```json
[
  {
    "id": "tenant-123",
    "name": "Acme Corp",
    "slug": "acme",
    "active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Tenant

```
POST /api/v1/tenants
```

**Request:**
```json
{
  "name": "Acme Corp",
  "slug": "acme",
  "metadata": {
    "industry": "technology"
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "tenant-123",
  "name": "Acme Corp",
  "slug": "acme",
  "active": true
}
```

### Get Tenant

```
GET /api/v1/tenants/{id}
```

### Update Tenant

```
PUT /api/v1/tenants/{id}
```

### Delete Tenant

```
DELETE /api/v1/tenants/{id}
```

---

## Datasources

### List Datasources

```
GET /api/v1/datasources?tenant_id=xxx
```

**Response:**
```json
[
  {
    "id": "ds-123",
    "tenant_id": "tenant-123",
    "name": "Production DB",
    "type": "postgres",
    "active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Datasource

```
POST /api/v1/datasources
```

**Request:**
```json
{
  "name": "Production DB",
  "type": "postgres",
  "description": "Main production database",
  "connection": {
    "host": "db.example.com",
    "port": 5432,
    "database": "production",
    "username": "reader",
    "password": "secret",
    "ssl_mode": "require"
  }
}
```

**Response:** `201 Created`

### Get Datasource

```
GET /api/v1/datasources/{id}
```

### Update Datasource

```
PUT /api/v1/datasources/{id}
```

**Request:**
```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "active": true
}
```

### Delete Datasource

```
DELETE /api/v1/datasources/{id}
```

**Response:** `204 No Content`

### Test Connection

```
POST /api/v1/datasources/test
```

**Request:**
```json
{
  "type": "postgres",
  "connection": {
    "host": "db.example.com",
    "port": 5432,
    "database": "test",
    "username": "test",
    "password": "test"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Connection successful"
}
```

### List Tables

```
GET /api/v1/datasources/{id}/tables
```

**Response:**
```json
[
  {
    "schema": "public",
    "name": "users",
    "type": "table",
    "row_count": 50000
  }
]
```

### List Datasource Checks

```
GET /api/v1/datasources/{id}/checks
```

---

## Checks

### List Checks

```
GET /api/v1/checks?tenant_id=xxx&datasource_id=xxx
```

**Response:**
```json
[
  {
    "id": "check-123",
    "name": "Users Row Count",
    "type": "row_count",
    "table": "users",
    "severity": "high",
    "last_status": "passed",
    "last_run_at": "2024-01-15T10:00:00Z"
  }
]
```

### Create Check

```
POST /api/v1/checks
```

**Request (Row Count):**
```json
{
  "name": "Users Minimum Rows",
  "datasource_id": "ds-123",
  "type": "row_count",
  "table": "users",
  "severity": "high",
  "parameters": {
    "min_rows": 1000
  }
}
```

**Request (Null Check):**
```json
{
  "name": "Email Not Null",
  "datasource_id": "ds-123",
  "type": "null_check",
  "table": "users",
  "column": "email",
  "severity": "critical",
  "parameters": {
    "max_null_percentage": 0
  }
}
```

**Request (Freshness):**
```json
{
  "name": "Orders Freshness",
  "datasource_id": "ds-123",
  "type": "freshness",
  "table": "orders",
  "severity": "high",
  "parameters": {
    "timestamp_column": "created_at",
    "max_age_hours": 24
  }
}
```

**Request (Custom SQL):**
```json
{
  "name": "Active Users",
  "datasource_id": "ds-123",
  "type": "custom_sql",
  "severity": "medium",
  "parameters": {
    "custom_sql": "SELECT COUNT(*) FROM users WHERE status = 'active'",
    "expected_value": "1000"
  },
  "threshold": {
    "type": "absolute",
    "operator": "gte",
    "value": 1000
  }
}
```

### Get Check

```
GET /api/v1/checks/{id}
```

### Update Check

```
PUT /api/v1/checks/{id}
```

### Delete Check

```
DELETE /api/v1/checks/{id}
```

### Run Check

```
POST /api/v1/checks/{id}/run
```

**Response:**
```json
{
  "id": "result-456",
  "check_id": "check-123",
  "status": "passed",
  "actual_value": 1500,
  "expected_value": 1000,
  "message": "Row count 1500 is within expected range",
  "duration": "125ms",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Get Check Results

```
GET /api/v1/checks/{id}/results
```

**Response:**
```json
[
  {
    "id": "result-456",
    "check_id": "check-123",
    "status": "passed",
    "actual_value": 1500,
    "timestamp": "2024-01-15T10:30:00Z"
  }
]
```

---

## Schedules

### List Schedules

```
GET /api/v1/schedules?tenant_id=xxx
```

### Create Schedule

```
POST /api/v1/schedules
```

**Request:**
```json
{
  "name": "Nightly Checks",
  "cron_expression": "0 0 * * *",
  "timezone": "UTC",
  "check_ids": ["check-123", "check-456"],
  "alert_channel_ids": ["channel-789"],
  "enabled": true
}
```

### Get Schedule

```
GET /api/v1/schedules/{id}
```

### Update Schedule

```
PUT /api/v1/schedules/{id}
```

### Delete Schedule

```
DELETE /api/v1/schedules/{id}
```

### Run Schedule Now

```
POST /api/v1/schedules/{id}/run
```

**Response:**
```json
{
  "id": "exec-789",
  "schedule_id": "schedule-123",
  "status": "running",
  "started_at": "2024-01-15T10:30:00Z"
}
```

### Get Schedule Executions

```
GET /api/v1/schedules/{id}/executions
```

---

## Alert Channels

### List Alert Channels

```
GET /api/v1/alerts/channels?tenant_id=xxx
```

### Create Alert Channel

```
POST /api/v1/alerts/channels
```

**Request (Slack):**
```json
{
  "name": "Data Team Slack",
  "type": "slack",
  "config": {
    "webhook_url": "https://hooks.slack.com/services/...",
    "channel": "#data-alerts"
  },
  "enabled": true
}
```

**Request (Email):**
```json
{
  "name": "Data Team Email",
  "type": "email",
  "config": {
    "smtp_host": "smtp.gmail.com",
    "smtp_port": 587,
    "smtp_user": "alerts@company.com",
    "smtp_password": "xxx",
    "from_email": "alerts@company.com",
    "to_emails": ["team@company.com"]
  },
  "enabled": true
}
```

**Request (PagerDuty):**
```json
{
  "name": "On-Call",
  "type": "pagerduty",
  "config": {
    "api_key": "xxx",
    "service_id": "PXXXXXX"
  },
  "enabled": true
}
```

### Get Alert Channel

```
GET /api/v1/alerts/channels/{id}
```

### Update Alert Channel

```
PUT /api/v1/alerts/channels/{id}
```

### Delete Alert Channel

```
DELETE /api/v1/alerts/channels/{id}
```

### Test Alert Channel

```
POST /api/v1/alerts/channels/{id}/test
```

**Response:**
```json
{
  "success": true,
  "message": "Test alert sent successfully"
}
```

### Get Alert History

```
GET /api/v1/alerts/history?channel_id=xxx
```

---

## Views

### List Views

```
GET /api/v1/views?tenant_id=xxx&datasource_id=xxx
```

### Create View

```
POST /api/v1/views
```

**Request:**
```json
{
  "name": "Daily Revenue",
  "datasource_id": "ds-123",
  "sql_query": "SELECT date_trunc('day', created_at) as date, SUM(total) as revenue FROM orders GROUP BY 1",
  "description": "Daily revenue aggregation"
}
```

### Get View

```
GET /api/v1/views/{id}
```

### Update View

```
PUT /api/v1/views/{id}
```

### Delete View

```
DELETE /api/v1/views/{id}
```

### Query View

```
GET /api/v1/views/{id}/query
```

**Response:**
```json
{
  "columns": ["date", "revenue"],
  "rows": [
    {"date": "2024-01-15", "revenue": 50000},
    {"date": "2024-01-14", "revenue": 45000}
  ],
  "row_count": 2
}
```

### Validate View

```
POST /api/v1/views/{id}/validate
```

**Response:**
```json
{
  "valid": true,
  "message": "View is valid"
}
```

### Get View SQL

```
GET /api/v1/views/{id}/sql
```

**Response:**
```json
{
  "sql": "SELECT date_trunc('day', created_at) as date, SUM(total) as revenue FROM orders GROUP BY 1"
}
```

---

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (successful delete) |
| 400 | Bad Request (invalid input) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (no permission) |
| 404 | Not Found |
| 500 | Internal Server Error |

---

## Rate Limiting

API requests are rate limited:
- 1000 requests per minute per user
- 100 concurrent connections

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1705315200
```

---

## Pagination

For list endpoints:

```
GET /api/v1/checks?page=1&per_page=50
```

Response includes pagination info:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 150,
    "total_pages": 3
  }
}
```
