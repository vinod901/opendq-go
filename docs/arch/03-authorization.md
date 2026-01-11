# Authorization

OpenDQ uses OpenFGA for fine-grained, relationship-based access control (ReBAC). This enables flexible permission models that go beyond simple role-based access control.

## Overview

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                        Authorization Flow                                     │
│                                                                              │
│  ┌─────────────────┐                                                         │
│  │ Authenticated   │                                                         │
│  │ Request         │                                                         │
│  └────────┬────────┘                                                         │
│           │                                                                  │
│           ▼                                                                  │
│  ┌─────────────────┐     ┌────────────────────────────────────────────┐     │
│  │ Authz Middleware│────▶│ OpenFGA Check                               │     │
│  │                 │     │ Check(user:alice, member, tenant:acme)     │     │
│  └────────┬────────┘     └────────────────────────────────────────────┘     │
│           │                               │                                  │
│           │                               ▼                                  │
│           │                      ┌────────────────┐                         │
│           │                      │    OpenFGA     │                         │
│           │                      │    Server      │                         │
│           │                      └────────┬───────┘                         │
│           │                               │                                  │
│           │◀───── allowed: true/false ────┘                                  │
│           │                                                                  │
│           ▼                                                                  │
│  ┌─────────────────┐                                                         │
│  │ Continue to     │ (if allowed)                                            │
│  │ Handler         │                                                         │
│  └─────────────────┘                                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

## OpenFGA Concepts

### 1. Authorization Model

The authorization model defines object types, relations, and how permissions are computed.

```json
// openfga-model.json
{
  "schema_version": "1.1",
  "type_definitions": [
    {
      "type": "user"
    },
    {
      "type": "tenant",
      "relations": {
        "owner": { "this": {} },
        "admin": { "this": {} },
        "editor": { "this": {} },
        "viewer": { "this": {} },
        "member": {
          "union": {
            "child": [
              { "this": {} },
              { "computedUserset": { "relation": "owner" } },
              { "computedUserset": { "relation": "admin" } },
              { "computedUserset": { "relation": "editor" } },
              { "computedUserset": { "relation": "viewer" } }
            ]
          }
        }
      }
    },
    {
      "type": "policy",
      "relations": {
        "parent": { "this": {} },
        "owner": { /* inherited from parent tenant */ },
        "editor": { /* inherited from parent tenant */ },
        "viewer": { /* inherited from parent tenant */ }
      }
    }
  ]
}
```

### 2. Relationship Tuples

Tuples define the actual permissions:

```
user:alice  owner   tenant:acme
user:bob    editor  tenant:acme
user:charlie viewer tenant:acme
tenant:acme parent  policy:policy-123
```

### 3. Check Queries

```
Check(user:alice, member, tenant:acme) → true  (alice is owner, which implies member)
Check(user:charlie, editor, tenant:acme) → false (charlie is only viewer)
```

## Authorization Manager

The Authorization Manager (`internal/authorization/authorization.go`) wraps OpenFGA operations:

### Configuration

```bash
# Environment variables
OPENFGA_STORE_ID=01HQ2N3YK4567890ABCDEFGH
OPENFGA_API_HOST=http://localhost:8081
OPENFGA_AUTH_MODEL=01HQ2N3YK4567890MODEL12
```

### Initialization

```go
authzManager, err := authorization.NewManager(authorization.Config{
    APIHost:   cfg.OpenFGA.APIHost,
    StoreID:   cfg.OpenFGA.StoreID,
    AuthModel: cfg.OpenFGA.AuthModel,
})
```

### Key Methods

#### 1. Check - Verify Permission

```go
// Check if user has permission
func (m *Manager) Check(ctx context.Context, user, relation, object string) (bool, error) {
    body := client.ClientCheckRequest{
        User:     user,      // e.g., "user:alice"
        Relation: relation,  // e.g., "member"
        Object:   object,    // e.g., "tenant:acme"
    }
    data, err := m.client.Check(ctx).Body(body).Execute()
    if err != nil {
        return false, err
    }
    return data.GetAllowed(), nil
}

// Example usage
allowed, err := authzManager.Check(ctx, "user:alice", "editor", "policy:policy-123")
if !allowed {
    return ErrForbidden
}
```

#### 2. WriteTuple - Grant Permission

```go
// Grant a permission
func (m *Manager) WriteTuple(ctx context.Context, user, relation, object string) error {
    body := client.ClientWriteRequest{
        Writes: []openfga.TupleKey{
            {
                User:     user,
                Relation: relation,
                Object:   object,
            },
        },
    }
    _, err := m.client.Write(ctx).Body(body).Execute()
    return err
}

// Example: Grant alice admin access to tenant
err := authzManager.WriteTuple(ctx, "user:alice", "admin", "tenant:acme")
```

#### 3. DeleteTuple - Revoke Permission

```go
// Revoke a permission
func (m *Manager) DeleteTuple(ctx context.Context, user, relation, object string) error {
    body := client.ClientWriteRequest{
        Deletes: []openfga.TupleKeyWithoutCondition{
            {
                User:     user,
                Relation: relation,
                Object:   object,
            },
        },
    }
    _, err := m.client.Write(ctx).Body(body).Execute()
    return err
}
```

#### 4. ListObjects - Find Accessible Resources

```go
// List all objects a user can access
func (m *Manager) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
    body := client.ClientListObjectsRequest{
        User:     user,
        Relation: relation,
        Type:     objectType,
    }
    data, err := m.client.ListObjects(ctx).Body(body).Execute()
    if err != nil {
        return nil, err
    }
    return data.GetObjects(), nil
}

// Example: Get all tenants user can view
tenants, err := authzManager.ListObjects(ctx, "user:alice", "member", "tenant")
// Returns: ["tenant:acme", "tenant:beta"]
```

## Helper Functions

### Format User/Object Identifiers

```go
// FormatUser creates a user identifier
func FormatUser(userType, userID string) string {
    return fmt.Sprintf("%s:%s", userType, userID)
}
// Example: FormatUser("user", "alice") → "user:alice"

// FormatObject creates an object identifier
func FormatObject(objectType, objectID string) string {
    return fmt.Sprintf("%s:%s", objectType, objectID)
}
// Example: FormatObject("tenant", "acme") → "tenant:acme"
```

### Tenant Access Helpers

```go
// Grant user access to tenant
func (m *Manager) GrantTenantAccess(ctx context.Context, userID, tenantID, relation string) error {
    user := FormatUser(TypeUser, userID)
    object := FormatObject(TypeTenant, tenantID)
    return m.WriteTuple(ctx, user, relation, object)
}

// Check tenant access
func (m *Manager) CheckTenantAccess(ctx context.Context, userID, tenantID, relation string) (bool, error) {
    user := FormatUser(TypeUser, userID)
    object := FormatObject(TypeTenant, tenantID)
    return m.Check(ctx, user, relation, object)
}

// Revoke tenant access
func (m *Manager) RevokeTenantAccess(ctx context.Context, userID, tenantID, relation string) error {
    user := FormatUser(TypeUser, userID)
    object := FormatObject(TypeTenant, tenantID)
    return m.DeleteTuple(ctx, user, relation, object)
}
```

## Authorization Middleware

The middleware checks permissions before requests reach handlers:

```go
type AuthzMiddleware struct {
    authzManager *authorization.Manager
}

func (m *AuthzMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip for public endpoints
        if isPublicEndpoint(r.URL.Path) {
            next.ServeHTTP(w, r)
            return
        }

        // Get user from context (set by auth middleware)
        userID, ok := r.Context().Value(contextKeyUserID).(string)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Get tenant from context (set by tenant middleware)
        tenantSlug, err := tenant.GetTenantSlug(r.Context())
        if err != nil {
            http.Error(w, "Tenant not found", http.StatusBadRequest)
            return
        }

        // Check if user has access to tenant
        allowed, err := m.authzManager.CheckTenantAccess(
            r.Context(),
            userID,
            tenantSlug,
            authorization.RelationMember,
        )
        if err != nil {
            http.Error(w, "Authorization check failed", http.StatusInternalServerError)
            return
        }

        if !allowed {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Permission Hierarchy

### Tenant Relations

```
owner
  └── admin
        └── editor
              └── viewer
                    └── member
```

- **owner**: Full control over tenant and all resources
- **admin**: Manage users and settings
- **editor**: Create and modify resources
- **viewer**: Read-only access
- **member**: Basic access (computed from all above)

### Resource Inheritance

Resources inherit permissions from their parent tenant:

```
tenant:acme
    │
    ├── policy:policy-123 (inherits from acme)
    ├── workflow:workflow-456 (inherits from acme)
    └── datasource:ds-789 (inherits from acme)
```

When alice is editor of tenant:acme:
- alice is automatically editor of all policies in acme
- alice can create new policies in acme
- alice can modify existing policies in acme

## Common Authorization Patterns

### 1. Creating Resources

```go
// When creating a resource, establish parent relationship
func (h *Handler) createPolicy(w http.ResponseWriter, r *http.Request) {
    // ... create policy in database ...
    
    // Link policy to tenant
    err := h.authzManager.WriteTuple(ctx, 
        authorization.FormatObject("tenant", policy.TenantID),
        "parent",
        authorization.FormatObject("policy", policy.ID),
    )
}
```

### 2. Checking Resource Access

```go
// Before modifying a resource, check permission
func (h *Handler) updatePolicy(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(contextKeyUserID).(string)
    policyID := extractIDFromPath(r.URL.Path, "/api/v1/policies")
    
    allowed, err := h.authzManager.Check(ctx,
        authorization.FormatUser("user", userID),
        "editor",
        authorization.FormatObject("policy", policyID),
    )
    if !allowed {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // ... proceed with update ...
}
```

### 3. Listing Resources

```go
// List only resources user can access
func (h *Handler) listPolicies(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(contextKeyUserID).(string)
    
    // Get all policies user can view
    policyIDs, err := h.authzManager.ListObjects(ctx,
        authorization.FormatUser("user", userID),
        "viewer",
        "policy",
    )
    
    // Fetch policies from database
    policies := h.policyManager.GetPoliciesByIDs(ctx, policyIDs)
    // ...
}
```

## OpenFGA Setup

### 1. Create Store

```bash
# Using FGA CLI
fga store create --name opendq

# Note the store ID for configuration
```

### 2. Create Authorization Model

```bash
# Upload model
fga model write --store-id $STORE_ID < openfga-model.json

# Note the authorization model ID
```

### 3. Initialize Relationships

```bash
# Create initial admin user
fga tuple write --store-id $STORE_ID \
    user:admin owner tenant:default
```

## OpenFGA Playground (UI)

The docker-compose includes OpenFGA Playground for testing:

```
http://localhost:3000
```

Features:
- Visual authorization model editor
- Test permission checks
- Manage relationship tuples
- Debug authorization queries

## Best Practices

1. **Principle of Least Privilege**: Grant minimum required permissions
2. **Use Inheritance**: Leverage tenant inheritance for resource permissions
3. **Cache Checks**: Cache frequently checked permissions
4. **Audit Trail**: Log authorization decisions for compliance
5. **Test Thoroughly**: Use OpenFGA Playground to test authorization model

## Troubleshooting

### Common Issues

1. **Store Not Found**: Verify OPENFGA_STORE_ID is correct
2. **Model Not Found**: Verify OPENFGA_AUTH_MODEL is correct
3. **Connection Refused**: Check OpenFGA server is running
4. **Permission Denied**: Check relationship tuples exist

### Debugging

```bash
# List all tuples in store
fga tuple read --store-id $STORE_ID

# Test a check
fga query check --store-id $STORE_ID \
    --user user:alice \
    --relation member \
    --object tenant:acme
```
