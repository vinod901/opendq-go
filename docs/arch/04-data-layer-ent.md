# Data Layer (Ent ORM)

OpenDQ uses [Ent](https://entgo.io/) as its ORM for type-safe database operations. Ent generates Go code from schema definitions, providing compile-time type safety and powerful query capabilities.

## Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           Ent Architecture                                  │
│                                                                            │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────┐   │
│  │  Schema Files   │───▶│  go generate    │───▶│  Generated Code     │   │
│  │  (ent/schema/)  │    │  ./ent          │    │  (type-safe client) │   │
│  └─────────────────┘    └─────────────────┘    └─────────────────────┘   │
│                                                                            │
│  Schema Definition                           Generated Features:           │
│  - Fields (types, defaults)                  - CRUD operations            │
│  - Edges (relationships)                     - Query builders             │
│  - Indexes                                   - Type-safe filters          │
│  - Hooks                                     - Migration support          │
└────────────────────────────────────────────────────────────────────────────┘
```

## Entity Schemas

Located in `ent/schema/`, each file defines an entity:

### Core Entities

| Entity | Description | File |
|--------|-------------|------|
| Tenant | Organization/workspace | `tenant.go` |
| User | Authenticated users | `user.go` |
| Policy | Data governance policies | `policy.go` |
| Workflow | State machine workflows | `workflow.go` |
| LineageEvent | Data lineage events | `lineage_event.go` |
| Datasource | Database connections | `datasource.go` |
| Check | Data quality checks | `check.go` |
| CheckResult | Check execution results | `check_result.go` |
| Schedule | Scheduled executions | `schedule.go` |
| ScheduleExecution | Execution history | `schedule_execution.go` |
| AlertChannel | Notification channels | `alert_channel.go` |
| AlertHistory | Alert delivery history | `alert_history.go` |
| View | Logical views | `view.go` |

## Schema Structure

### Tenant Schema

```go
// ent/schema/tenant.go
package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type Tenant struct {
    ent.Schema
}

func (Tenant) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").
            Unique().
            Immutable(),
        field.String("name").
            NotEmpty(),
        field.String("slug").
            Unique().
            NotEmpty(),
        field.JSON("metadata", map[string]interface{}{}).
            Optional(),
        field.Bool("active").
            Default(true),
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (Tenant) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("users", User.Type),
        edge.To("policies", Policy.Type),
        edge.To("workflows", Workflow.Type),
        edge.To("lineage_events", LineageEvent.Type),
        edge.To("datasources", Datasource.Type),
        edge.To("checks", Check.Type),
        edge.To("schedules", Schedule.Type),
        edge.To("alert_channels", AlertChannel.Type),
        edge.To("views", View.Type),
    }
}

func (Tenant) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("slug").Unique(),
        index.Fields("active"),
    }
}
```

### Datasource Schema

```go
// ent/schema/datasource.go
package schema

type Datasource struct {
    ent.Schema
}

func (Datasource) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").
            Unique().
            Immutable(),
        field.String("tenant_id"),
        field.String("name").
            NotEmpty(),
        field.String("description").
            Optional(),
        field.String("type").
            NotEmpty(), // postgres, mysql, snowflake, etc.
        field.JSON("connection", map[string]interface{}{}),
        field.JSON("metadata", map[string]interface{}{}).
            Optional(),
        field.Bool("active").
            Default(true),
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (Datasource) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("tenant", Tenant.Type).
            Ref("datasources").
            Unique().
            Required(),
        edge.To("checks", Check.Type),
        edge.To("views", View.Type),
    }
}

func (Datasource) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("tenant_id"),
        index.Fields("tenant_id", "name").Unique(),
        index.Fields("type"),
        index.Fields("active"),
    }
}
```

### Check Schema

```go
// ent/schema/check.go
package schema

type Check struct {
    ent.Schema
}

func (Check) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").
            Unique().
            Immutable(),
        field.String("tenant_id"),
        field.String("datasource_id"),
        field.String("name").
            NotEmpty(),
        field.String("description").
            Optional(),
        field.String("type").
            NotEmpty(), // row_count, null_check, etc.
        field.String("table"),
        field.String("column").
            Optional(),
        field.JSON("parameters", map[string]interface{}{}),
        field.JSON("threshold", map[string]interface{}{}),
        field.String("severity").
            Default("medium"),
        field.JSON("tags", []string{}).
            Optional(),
        field.JSON("metadata", map[string]interface{}{}).
            Optional(),
        field.Bool("active").
            Default(true),
        field.String("schedule_id").
            Optional(),
        field.String("view_id").
            Optional(),
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
        field.Time("last_run_at").
            Optional().
            Nillable(),
        field.String("last_status").
            Optional(),
    }
}

func (Check) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("tenant", Tenant.Type).
            Ref("checks").
            Unique().
            Required(),
        edge.From("datasource", Datasource.Type).
            Ref("checks").
            Unique().
            Required(),
        edge.To("results", CheckResult.Type),
        edge.From("schedule", Schedule.Type).
            Ref("checks").
            Unique(),
        edge.From("view", View.Type).
            Ref("checks").
            Unique(),
    }
}
```

## Code Generation

### Generate Code

```bash
# Generate Ent code
go generate ./ent

# Or use Makefile
make ent-generate
```

### Generated Files

After generation, `ent/` contains:

```
ent/
├── schema/           # Your schema definitions
├── client.go         # Main client
├── tenant.go         # Tenant entity
├── tenant_create.go  # Create builder
├── tenant_update.go  # Update builder
├── tenant_query.go   # Query builder
├── tenant_delete.go  # Delete builder
├── user.go           # User entity
├── ...               # Other entities
├── migrate/          # Migration support
└── runtime.go        # Runtime hooks
```

## Using Ent Client

### Initialize Client

```go
package main

import (
    "context"
    "log"
    
    "entgo.io/ent/dialect"
    _ "github.com/lib/pq"
    "github.com/vinod901/opendq-go/ent"
)

func main() {
    // Create client
    client, err := ent.Open(dialect.Postgres, 
        "host=localhost port=5432 user=postgres dbname=opendq sslmode=disable")
    if err != nil {
        log.Fatalf("failed opening connection: %v", err)
    }
    defer client.Close()
    
    // Run migrations
    if err := client.Schema.Create(context.Background()); err != nil {
        log.Fatalf("failed creating schema: %v", err)
    }
}
```

### CRUD Operations

#### Create

```go
// Create tenant
tenant, err := client.Tenant.
    Create().
    SetID(uuid.New().String()).
    SetName("Acme Corp").
    SetSlug("acme").
    SetMetadata(map[string]interface{}{"industry": "tech"}).
    Save(ctx)

// Create datasource linked to tenant
datasource, err := client.Datasource.
    Create().
    SetID(uuid.New().String()).
    SetTenant(tenant).
    SetName("Production DB").
    SetType("postgres").
    SetConnection(map[string]interface{}{
        "host": "db.example.com",
        "port": 5432,
        "database": "production",
    }).
    Save(ctx)
```

#### Read

```go
// Get by ID
tenant, err := client.Tenant.Get(ctx, "tenant-id")

// Query with filters
tenants, err := client.Tenant.
    Query().
    Where(tenant.ActiveEQ(true)).
    All(ctx)

// Get with edges (relationships)
tenant, err := client.Tenant.
    Query().
    Where(tenant.SlugEQ("acme")).
    WithDatasources().
    WithChecks().
    Only(ctx)
```

#### Update

```go
// Update by ID
tenant, err := client.Tenant.
    UpdateOneID("tenant-id").
    SetName("Acme Corp Updated").
    SetMetadata(map[string]interface{}{"updated": true}).
    Save(ctx)

// Bulk update
count, err := client.Tenant.
    Update().
    Where(tenant.ActiveEQ(false)).
    SetActive(true).
    Save(ctx)
```

#### Delete

```go
// Delete by ID
err := client.Tenant.DeleteOneID("tenant-id").Exec(ctx)

// Bulk delete
count, err := client.Tenant.
    Delete().
    Where(tenant.ActiveEQ(false)).
    Exec(ctx)
```

### Complex Queries

```go
// Get checks for a datasource with results
checks, err := client.Check.
    Query().
    Where(
        check.DatasourceIDEQ("ds-123"),
        check.ActiveEQ(true),
    ).
    WithResults(func(q *ent.CheckResultQuery) {
        q.Order(ent.Desc(checkresult.FieldTimestamp)).
            Limit(10)
    }).
    All(ctx)

// Aggregate query
count, err := client.Check.
    Query().
    Where(
        check.TenantIDEQ("tenant-id"),
        check.LastStatusEQ("failed"),
    ).
    Count(ctx)
```

## Database Migrations

### Auto Migration

```go
// Run auto-migration (development)
if err := client.Schema.Create(ctx); err != nil {
    log.Fatal(err)
}
```

### Versioned Migrations

```go
// Generate versioned migration
import "entgo.io/ent/migrate"

// Create migration directory
if err := client.Schema.Create(ctx,
    migrate.WithDir("./migrations"),
    migrate.WithForeignKeys(true),
); err != nil {
    log.Fatal(err)
}
```

## Entity Relationships

### One-to-Many (Tenant → Datasources)

```go
// Tenant schema
edge.To("datasources", Datasource.Type)

// Datasource schema
edge.From("tenant", Tenant.Type).
    Ref("datasources").
    Unique().
    Required()
```

### Many-to-Many (Check ↔ Schedule)

```go
// Schedule schema
edge.To("checks", Check.Type)

// Check schema
edge.From("schedule", Schedule.Type).
    Ref("checks").
    Unique()
```

### Self-Referential (View → Source Views)

```go
// View can reference other views
edge.To("source_views", View.Type)
edge.From("dependent_views", View.Type).
    Ref("source_views")
```

## Integration with Managers

### TenantManager with Ent

```go
// internal/tenant/tenant.go
type Manager struct {
    client *ent.Client
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
        Save(ctx)
}

func (m *Manager) GetTenant(ctx context.Context, id string) (*ent.Tenant, error) {
    return m.client.Tenant.Get(ctx, id)
}

func (m *Manager) ListTenants(ctx context.Context) ([]*ent.Tenant, error) {
    return m.client.Tenant.
        Query().
        Where(tenant.ActiveEQ(true)).
        All(ctx)
}
```

## Field Types Reference

| Go Type | SQL Type | Notes |
|---------|----------|-------|
| `field.String` | VARCHAR | Use `.NotEmpty()` for required |
| `field.Int` | INTEGER | |
| `field.Int64` | BIGINT | |
| `field.Float64` | DOUBLE | |
| `field.Bool` | BOOLEAN | |
| `field.Time` | TIMESTAMP | |
| `field.JSON` | JSONB | For dynamic data |
| `field.Enum` | VARCHAR | With `.Values()` |
| `field.Bytes` | BYTEA | Binary data |
| `field.UUID` | UUID | PostgreSQL UUID |

## Best Practices

1. **Use Indexes**: Add indexes for frequently queried fields
2. **Soft Delete**: Use `active` field instead of hard delete
3. **Audit Fields**: Always include `created_at`, `updated_at`
4. **UUID IDs**: Use UUIDs for distributed systems
5. **JSON for Flexibility**: Use JSON fields for dynamic metadata
6. **Transactions**: Use transactions for multi-entity operations
7. **Eager Loading**: Use `With*` methods to avoid N+1 queries
