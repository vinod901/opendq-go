# API Design

OpenDQ provides a RESTful API for all platform operations. The API follows REST conventions with JSON payloads.

## API Architecture

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           API Structure                                     │
│                                                                            │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                     HTTP Router (ServeMux)                           │  │
│  │                                                                      │  │
│  │  /health                   → HealthCheck                            │  │
│  │  /api/v1/tenants          → Handler.handleTenants                   │  │
│  │  /api/v1/policies         → Handler.handlePolicies                  │  │
│  │  /api/v1/workflows        → Handler.handleWorkflows                 │  │
│  │  /api/v1/lineage          → Handler.handleLineage                   │  │
│  │  /api/v1/datasources      → DataQualityHandler.handleDatasources    │  │
│  │  /api/v1/checks           → DataQualityHandler.handleChecks         │  │
│  │  /api/v1/schedules        → DataQualityHandler.handleSchedules      │  │
│  │  /api/v1/alerts/channels  → DataQualityHandler.handleAlertChannels  │  │
│  │  /api/v1/views            → DataQualityHandler.handleViews          │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────┘
```

## Handler Organization

### Core Handler (`api/http/handler.go`)

Handles platform-level resources:

```go
type Handler struct {
    tenantManager   *tenant.Manager
    policyManager   *policy.Manager
    workflowEngine  *workflow.Engine
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    // Health check
    mux.HandleFunc("/health", h.HealthCheck)
    
    // Tenant routes
    mux.HandleFunc("/api/v1/tenants", h.handleTenants)
    mux.HandleFunc("/api/v1/tenants/", h.handleTenant)
    
    // Policy routes
    mux.HandleFunc("/api/v1/policies", h.handlePolicies)
    mux.HandleFunc("/api/v1/policies/", h.handlePolicy)
    
    // Workflow routes
    mux.HandleFunc("/api/v1/workflows", h.handleWorkflows)
    mux.HandleFunc("/api/v1/workflows/", h.handleWorkflow)
    
    // Lineage routes
    mux.HandleFunc("/api/v1/lineage", h.handleLineage)
}
```

### Data Quality Handler (`api/http/dq_handler.go`)

Handles data quality-specific resources:

```go
type DataQualityHandler struct {
    datasourceManager *datasource.Manager
    checkManager      *check.Manager
    schedulerManager  *scheduler.Manager
    alertManager      *alerting.Manager
    viewManager       *view.Manager
}

func (h *DataQualityHandler) RegisterRoutes(mux *http.ServeMux) {
    // Datasource routes
    mux.HandleFunc("/api/v1/datasources", h.handleDatasources)
    mux.HandleFunc("/api/v1/datasources/", h.handleDatasource)
    mux.HandleFunc("/api/v1/datasources/test", h.testDatasourceConnection)
    
    // Check routes
    mux.HandleFunc("/api/v1/checks", h.handleChecks)
    mux.HandleFunc("/api/v1/checks/", h.handleCheck)
    
    // Schedule routes
    mux.HandleFunc("/api/v1/schedules", h.handleSchedules)
    mux.HandleFunc("/api/v1/schedules/", h.handleSchedule)
    
    // Alert channel routes
    mux.HandleFunc("/api/v1/alerts/channels", h.handleAlertChannels)
    mux.HandleFunc("/api/v1/alerts/channels/", h.handleAlertChannel)
    mux.HandleFunc("/api/v1/alerts/history", h.getAlertHistory)
    
    // View routes
    mux.HandleFunc("/api/v1/views", h.handleViews)
    mux.HandleFunc("/api/v1/views/", h.handleView)
}
```

## API Endpoints Reference

### Health Check

```
GET /health

Response:
{
    "status": "healthy"
}
```

### Tenants

```
GET    /api/v1/tenants           List all tenants
POST   /api/v1/tenants           Create a tenant
GET    /api/v1/tenants/{id}      Get tenant by ID
PUT    /api/v1/tenants/{id}      Update tenant
DELETE /api/v1/tenants/{id}      Delete tenant
```

### Policies

```
GET    /api/v1/policies               List policies (filter: ?tenant_id=xxx)
POST   /api/v1/policies               Create policy
GET    /api/v1/policies/{id}          Get policy by ID
PUT    /api/v1/policies/{id}          Update policy
DELETE /api/v1/policies/{id}          Delete policy
```

### Workflows

```
GET    /api/v1/workflows              List workflows
POST   /api/v1/workflows              Create workflow
GET    /api/v1/workflows/{id}         Get workflow by ID
POST   /api/v1/workflows/{id}         Trigger state transition
```

### Datasources

```
GET    /api/v1/datasources                List datasources (?tenant_id=xxx)
POST   /api/v1/datasources                Create datasource
POST   /api/v1/datasources/test           Test connection (without saving)
GET    /api/v1/datasources/{id}           Get datasource
PUT    /api/v1/datasources/{id}           Update datasource
DELETE /api/v1/datasources/{id}           Delete datasource
GET    /api/v1/datasources/{id}/tables    List tables in datasource
GET    /api/v1/datasources/{id}/checks    List checks for datasource
```

### Checks

```
GET    /api/v1/checks                  List checks (?tenant_id=xxx&datasource_id=xxx)
POST   /api/v1/checks                  Create check
GET    /api/v1/checks/{id}             Get check
PUT    /api/v1/checks/{id}             Update check
DELETE /api/v1/checks/{id}             Delete check
POST   /api/v1/checks/{id}/run         Execute check immediately
GET    /api/v1/checks/{id}/results     Get check execution history
```

### Schedules

```
GET    /api/v1/schedules                    List schedules (?tenant_id=xxx)
POST   /api/v1/schedules                    Create schedule
GET    /api/v1/schedules/{id}               Get schedule
PUT    /api/v1/schedules/{id}               Update schedule
DELETE /api/v1/schedules/{id}               Delete schedule
POST   /api/v1/schedules/{id}/run           Run schedule immediately
GET    /api/v1/schedules/{id}/executions    Get execution history
```

### Alert Channels

```
GET    /api/v1/alerts/channels              List alert channels
POST   /api/v1/alerts/channels              Create channel
GET    /api/v1/alerts/channels/{id}         Get channel
PUT    /api/v1/alerts/channels/{id}         Update channel
DELETE /api/v1/alerts/channels/{id}         Delete channel
POST   /api/v1/alerts/channels/{id}/test    Send test alert
GET    /api/v1/alerts/history               Get alert history
```

### Views

```
GET    /api/v1/views                   List views (?tenant_id=xxx&datasource_id=xxx)
POST   /api/v1/views                   Create view
GET    /api/v1/views/{id}              Get view
PUT    /api/v1/views/{id}              Update view
DELETE /api/v1/views/{id}              Delete view
GET    /api/v1/views/{id}/query        Execute view query
POST   /api/v1/views/{id}/validate     Validate view definition
GET    /api/v1/views/{id}/sql          Get generated SQL
```

## Request/Response Patterns

### Collection List Pattern

```go
func (h *Handler) listTenants(w http.ResponseWriter, r *http.Request) {
    tenants, err := h.tenantManager.ListTenants(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tenants)
}
```

### Create Pattern

```go
func (h *Handler) createTenant(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name     string                 `json:"name"`
        Slug     string                 `json:"slug"`
        Metadata map[string]interface{} `json:"metadata"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    tenant, err := h.tenantManager.CreateTenant(r.Context(), req.Name, req.Slug, req.Metadata)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(tenant)
}
```

### ID Extraction Pattern

```go
func extractIDFromPath(path, prefix string) string {
    path = strings.TrimPrefix(path, prefix)
    path = strings.TrimPrefix(path, "/")
    parts := strings.Split(path, "/")
    if len(parts) > 0 {
        return parts[0]
    }
    return ""
}

// Usage: /api/v1/datasources/abc123/tables
// extractIDFromPath(path, "/api/v1/datasources") → "abc123"
```

### Sub-Resource Pattern

```go
func (h *DataQualityHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
    id := extractIDFromPath(r.URL.Path, "/api/v1/checks")
    
    // Check for sub-resources
    if strings.Contains(r.URL.Path, "/run") {
        h.runCheck(w, r, id)
        return
    }
    if strings.Contains(r.URL.Path, "/results") {
        h.getCheckResults(w, r, id)
        return
    }
    
    // Handle main resource
    switch r.Method {
    case http.MethodGet:
        h.getCheck(w, r, id)
    case http.MethodPut:
        h.updateCheck(w, r, id)
    case http.MethodDelete:
        h.deleteCheck(w, r, id)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
```

## Request Examples

### Create Datasource

```bash
curl -X POST http://localhost:8080/api/v1/datasources \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant: acme" \
  -d '{
    "name": "Production Database",
    "type": "postgres",
    "connection": {
      "host": "db.example.com",
      "port": 5432,
      "database": "production",
      "username": "reader",
      "password": "secret",
      "ssl_mode": "require"
    },
    "description": "Main production database"
  }'
```

### Create Check

```bash
curl -X POST http://localhost:8080/api/v1/checks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant: acme" \
  -d '{
    "name": "Users Table Row Count",
    "datasource_id": "ds-123",
    "type": "row_count",
    "table": "users",
    "severity": "high",
    "parameters": {
      "min_rows": 1000
    },
    "threshold": {
      "type": "absolute",
      "value": 1000,
      "operator": "gte"
    }
  }'
```

### Run Check

```bash
curl -X POST http://localhost:8080/api/v1/checks/check-123/run \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant: acme"
```

### Create Schedule

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant: acme" \
  -d '{
    "name": "Nightly Checks",
    "cron_expression": "0 0 * * *",
    "check_ids": ["check-123", "check-456"],
    "alert_channel_ids": ["channel-789"],
    "timezone": "UTC",
    "enabled": true
  }'
```

## Error Handling

### Standard Error Responses

```go
// 400 Bad Request - Invalid input
http.Error(w, "Invalid JSON payload", http.StatusBadRequest)

// 401 Unauthorized - Missing/invalid auth
http.Error(w, "Unauthorized", http.StatusUnauthorized)

// 403 Forbidden - No permission
http.Error(w, "Forbidden", http.StatusForbidden)

// 404 Not Found - Resource doesn't exist
http.Error(w, "Resource not found", http.StatusNotFound)

// 500 Internal Server Error - Server error
http.Error(w, err.Error(), http.StatusInternalServerError)
```

### Error Response Format

```json
{
    "error": "check not found: check-123",
    "code": "NOT_FOUND",
    "details": {
        "resource": "check",
        "id": "check-123"
    }
}
```

## Headers

### Request Headers

| Header | Description | Required |
|--------|-------------|----------|
| `Authorization` | Bearer JWT token | Yes (except public endpoints) |
| `Content-Type` | `application/json` | Yes (for POST/PUT) |
| `X-Tenant` | Tenant slug | Yes (multi-tenant mode) |

### Response Headers

| Header | Description |
|--------|-------------|
| `Content-Type` | `application/json` |
| `Access-Control-Allow-Origin` | CORS origin |
| `Access-Control-Allow-Methods` | Allowed HTTP methods |

## Pagination (Future)

```
GET /api/v1/checks?page=1&per_page=50

Response:
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

## OpenAPI Documentation

OpenAPI/Swagger documentation can be served at `/api/docs`:

```go
// Add to handler registration
mux.HandleFunc("/api/docs", h.serveOpenAPISpec)
mux.HandleFunc("/api/docs/", h.serveSwaggerUI)
```

The OpenAPI specification provides:
- Interactive API documentation
- Request/response schemas
- Try-it-out functionality
- Client SDK generation

See [10-development-guide.md](10-development-guide.md) for OpenAPI setup details.
