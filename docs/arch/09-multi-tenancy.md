# Multi-Tenancy

OpenDQ is designed as a multi-tenant platform, allowing multiple organizations (tenants) to use the same deployment while maintaining strict data isolation.

## Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                        Multi-Tenant Architecture                            │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                        Tenant A                                       │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐     │ │
│  │  │Datasources │  │  Checks    │  │ Schedules  │  │  Alerts    │     │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘     │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                        Tenant B                                       │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐     │ │
│  │  │Datasources │  │  Checks    │  │ Schedules  │  │  Alerts    │     │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘     │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                     Shared Infrastructure                             │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐     │ │
│  │  │PostgreSQL  │  │  OpenFGA   │  │  Keycloak  │  │   Redis    │     │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘     │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
```

## Tenant Model

```go
// internal/tenant/tenant.go

type Tenant struct {
    ID       string                 `json:"id"`
    Name     string                 `json:"name"`
    Slug     string                 `json:"slug"`     // URL-friendly identifier
    Metadata map[string]interface{} `json:"metadata"`
    Active   bool                   `json:"active"`
}

// Context keys for tenant information
type contextKey string

const (
    TenantIDKey   contextKey = "tenant_id"
    TenantSlugKey contextKey = "tenant_slug"
)
```

## Tenant Resolution

Tenants are identified from incoming requests via:

### 1. Header-Based Resolution (API)

```go
// X-Tenant header
func extractTenantSlug(r *http.Request) string {
    if tenantHeader := r.Header.Get("X-Tenant"); tenantHeader != "" {
        return tenantHeader
    }
    return ""
}
```

**Usage:**
```bash
curl -H "X-Tenant: acme" http://localhost:8080/api/v1/datasources
```

### 2. Subdomain-Based Resolution (Web)

```go
// acme.opendq.example.com
func extractTenantFromSubdomain(r *http.Request) string {
    host := r.Host
    parts := strings.Split(host, ".")
    if len(parts) > 2 {
        return parts[0] // First part is tenant slug
    }
    return ""
}
```

### 3. Path-Based Resolution (Alternative)

```
/api/v1/tenants/{tenant_slug}/datasources
```

## Tenant Middleware

The tenant middleware resolves and validates tenants:

```go
type TenantMiddleware struct {
    tenantManager *tenant.Manager
}

func (m *TenantMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantSlug := extractTenantSlug(r)
        if tenantSlug == "" {
            http.Error(w, "Tenant not found", http.StatusBadRequest)
            return
        }
        
        // Validate tenant exists and is active
        t, err := m.tenantManager.GetTenantBySlug(r.Context(), tenantSlug)
        if err != nil {
            http.Error(w, "Invalid tenant", http.StatusBadRequest)
            return
        }
        
        if !t.Active {
            http.Error(w, "Tenant is inactive", http.StatusForbidden)
            return
        }
        
        // Add tenant to context
        ctx := tenant.WithTenantSlug(r.Context(), tenantSlug)
        ctx = tenant.WithTenantID(ctx, t.ID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Context Propagation

Tenant information flows through context:

```go
// Set tenant in context (middleware)
ctx := tenant.WithTenantSlug(r.Context(), "acme")
ctx = tenant.WithTenantID(ctx, "tenant-123")

// Get tenant in handlers/services
tenantSlug, err := tenant.GetTenantSlug(ctx)
tenantID, err := tenant.GetTenantID(ctx)
```

## Data Isolation

### Database-Level Isolation

All resources include `tenant_id`:

```go
type Datasource struct {
    ID       string `json:"id"`
    TenantID string `json:"tenant_id"`  // Required
    // ... other fields
}
```

### Query Filtering

All queries filter by tenant:

```go
func (m *Manager) ListDatasources(ctx context.Context, tenantID string) ([]*Datasource, error) {
    var result []*Datasource
    for _, ds := range m.datasources {
        if tenantID == "" || ds.TenantID == tenantID {
            result = append(result, ds)
        }
    }
    return result, nil
}
```

With Ent ORM:

```go
func (m *Manager) ListDatasources(ctx context.Context, tenantID string) ([]*ent.Datasource, error) {
    return m.client.Datasource.
        Query().
        Where(datasource.TenantIDEQ(tenantID)).
        All(ctx)
}
```

## Authorization Integration

OpenFGA provides tenant-level access control:

### Grant Tenant Access

```go
// When user joins a tenant
err := authzManager.WriteTuple(ctx,
    "user:alice",
    "member",
    "tenant:acme",
)
```

### Check Tenant Access

```go
// In middleware
allowed, err := authzManager.Check(ctx,
    fmt.Sprintf("user:%s", userID),
    "member",
    fmt.Sprintf("tenant:%s", tenantSlug),
)
if !allowed {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### List User's Tenants

```go
// Get all tenants user has access to
tenants, err := authzManager.ListObjects(ctx,
    "user:alice",
    "member",
    "tenant",
)
// Returns: ["tenant:acme", "tenant:beta"]
```

## Tenant Onboarding Flow

### 1. Create Tenant

```go
// Admin creates new tenant
tenant, err := tenantManager.CreateTenant(ctx, "Acme Corp", "acme", map[string]interface{}{
    "industry": "technology",
    "plan":     "enterprise",
})
```

### 2. Assign Owner

```go
// Grant owner access to creating user
err := authzManager.WriteTuple(ctx,
    "user:admin123",
    "owner",
    "tenant:"+tenant.ID,
)
```

### 3. Invite Members

```go
// Owner invites team members
err := authzManager.WriteTuple(ctx,
    "user:newuser456",
    "editor",
    "tenant:"+tenant.ID,
)
```

## API Examples

### Create Tenant

```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "Acme Corporation",
    "slug": "acme",
    "metadata": {
      "industry": "technology",
      "plan": "enterprise"
    }
  }'
```

### List Tenants (Admin)

```bash
curl http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Access Tenant Resources

```bash
# All subsequent requests include tenant header
curl http://localhost:8080/api/v1/datasources \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant: acme"
```

## Configuration

### Enable Multi-Tenancy

```bash
# Environment variables
MULTITENANT_ENABLED=true
MULTITENANT_ISOLATION=namespace  # Options: namespace, schema, database
```

### Isolation Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `namespace` | Same database, filtered by tenant_id | Small to medium deployments |
| `schema` | Separate PostgreSQL schemas per tenant | Better isolation |
| `database` | Separate databases per tenant | Maximum isolation, regulatory requirements |

## Tenant Management

### TenantManager

```go
type Manager struct {
    client *ent.Client  // Database client
}

func NewManager(client *ent.Client) *Manager {
    return &Manager{client: client}
}

func (m *Manager) CreateTenant(ctx context.Context, name, slug string, metadata map[string]interface{}) (*ent.Tenant, error) {
    return m.client.Tenant.
        Create().
        SetID(uuid.New().String()).
        SetName(name).
        SetSlug(slug).
        SetMetadata(metadata).
        SetActive(true).
        Save(ctx)
}

func (m *Manager) GetTenantBySlug(ctx context.Context, slug string) (*ent.Tenant, error) {
    return m.client.Tenant.
        Query().
        Where(tenant.SlugEQ(slug)).
        Only(ctx)
}

func (m *Manager) ListTenants(ctx context.Context) ([]*ent.Tenant, error) {
    return m.client.Tenant.
        Query().
        Where(tenant.ActiveEQ(true)).
        All(ctx)
}

func (m *Manager) DeactivateTenant(ctx context.Context, id string) error {
    return m.client.Tenant.
        UpdateOneID(id).
        SetActive(false).
        Exec(ctx)
}
```

## Tenant Roles

| Role | Permissions |
|------|-------------|
| `owner` | Full control, can delete tenant |
| `admin` | Manage users, settings, resources |
| `editor` | Create/modify resources |
| `viewer` | Read-only access |
| `member` | Basic access (implied by all roles) |

## Best Practices

1. **Slug Validation**: Enforce slug uniqueness and format
2. **Default Tenant**: Consider a default tenant for simpler deployments
3. **Tenant Limits**: Implement resource quotas per tenant
4. **Audit Logging**: Log tenant-scoped operations
5. **Data Export**: Provide tenant data export capabilities
6. **Tenant Deletion**: Implement soft delete with grace period

## Single-Tenant Mode

For simpler deployments:

```bash
MULTITENANT_ENABLED=false
```

In single-tenant mode:
- No tenant resolution middleware
- No X-Tenant header required
- All resources belong to implicit default tenant
- Simpler authorization model
