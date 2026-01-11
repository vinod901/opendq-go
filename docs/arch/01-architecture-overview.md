# Architecture Overview

## System Design

OpenDQ is built as a modular, multi-tenant data quality platform. The architecture follows a layered approach with clear separation of concerns.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Client Layer                                        │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐          │
│  │  Frontend (Web)  │  │   CLI Client     │  │   API Clients    │          │
│  │   (SvelteKit)    │  │                  │  │   (REST/SDK)     │          │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘          │
└───────────┼──────────────────────┼──────────────────────┼────────────────────┘
            │                      │                      │
            │         HTTPS/REST API (JSON)               │
            └──────────────────────┼──────────────────────┘
                                   │
┌──────────────────────────────────▼──────────────────────────────────────────┐
│                          API Gateway Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      Middleware Chain                                │   │
│  │  ┌─────────┐  ┌─────────┐  ┌──────────┐  ┌────────────┐            │   │
│  │  │  CORS   │──│  Auth   │──│  Tenant  │──│   Authz    │            │   │
│  │  └─────────┘  └─────────┘  └──────────┘  └────────────┘            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        HTTP Handlers                                 │   │
│  │  ┌──────────────────┐  ┌─────────────────────────┐                  │   │
│  │  │  Handler (Core)  │  │  DataQualityHandler     │                  │   │
│  │  │  - Tenants       │  │  - Datasources          │                  │   │
│  │  │  - Policies      │  │  - Checks               │                  │   │
│  │  │  - Workflows     │  │  - Schedules            │                  │   │
│  │  │  - Lineage       │  │  - Alerts               │                  │   │
│  │  └──────────────────┘  │  - Views                │                  │   │
│  │                        └─────────────────────────┘                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
┌──────────────────────────────────▼──────────────────────────────────────────┐
│                          Service Layer                                       │
│                                                                              │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐   │
│  │ Tenant Manager│ │ Policy Manager│ │Workflow Engine│ │Lineage Client │   │
│  └───────┬───────┘ └───────┬───────┘ └───────┬───────┘ └───────┬───────┘   │
│          │                 │                 │                 │            │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐   │
│  │  DS Manager   │ │ Check Manager │ │Schedule Manager│ │ Alert Manager │   │
│  │ (Datasources) │ │ (DQ Checks)   │ │  (Scheduler)  │ │  (Alerting)   │   │
│  └───────┬───────┘ └───────┬───────┘ └───────┬───────┘ └───────┬───────┘   │
│          │                 │                 │                 │            │
│  ┌───────────────┐ ┌───────────────┐                                       │
│  │  View Manager │ │  Auth Manager │                                       │
│  │ (Logical Views)│ │ (OIDC Auth)  │                                       │
│  └───────┬───────┘ └───────┬───────┘                                       │
│          │                 │                                                │
│  ┌───────────────┐                                                          │
│  │ Authz Manager │                                                          │
│  │  (OpenFGA)    │                                                          │
│  └───────┬───────┘                                                          │
└──────────┼──────────────────────────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────────────────────────┐
│                          Data Layer                                          │
│                                                                              │
│  ┌───────────────────┐  ┌──────────────────┐  ┌─────────────────────┐      │
│  │    PostgreSQL     │  │     Redis        │  │    External DBs     │      │
│  │  (Primary Store)  │  │    (Cache)       │  │  (Datasources)      │      │
│  └───────────────────┘  └──────────────────┘  └─────────────────────┘      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                   │
┌──────────────────────────────────▼──────────────────────────────────────────┐
│                          External Services                                   │
│                                                                              │
│  ┌───────────────────┐  ┌──────────────────┐  ┌─────────────────────┐      │
│  │     OpenFGA       │  │    Keycloak      │  │     Marquez         │      │
│  │  (Authorization)  │  │ (OIDC Provider)  │  │ (OpenLineage)       │      │
│  └───────────────────┘  └──────────────────┘  └─────────────────────┘      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Component Interaction Flow

### 1. Request Lifecycle

```
Client Request
      │
      ▼
┌─────────────────┐
│ CORS Middleware │ ─── Adds CORS headers
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Auth Middleware │ ─── Validates JWT token (OIDC)
└────────┬────────┘     Extracts user claims
         │              Adds user to context
         ▼
┌─────────────────┐
│Tenant Middleware│ ─── Resolves tenant from header/subdomain
└────────┬────────┘     Adds tenant to context
         │
         ▼
┌─────────────────┐
│ Authz Middleware│ ─── Checks OpenFGA permissions
└────────┬────────┘     Verifies user can access tenant
         │
         ▼
┌─────────────────┐
│  HTTP Handler   │ ─── Processes request
└────────┬────────┘     Calls service layer
         │
         ▼
┌─────────────────┐
│ Service Layer   │ ─── Business logic
└────────┬────────┘     Database operations
         │
         ▼
     Response
```

### 2. Data Quality Check Execution Flow

```
┌────────────────┐
│ User Request   │──▶ POST /api/v1/checks/{id}/run
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Check Manager  │──▶ GetCheck(id)
└───────┬────────┘    Validate check exists & active
        │
        ▼
┌────────────────┐
│ DS Manager     │──▶ GetConnector(datasourceID)
└───────┬────────┘    Get database connector
        │
        ▼
┌────────────────┐
│ Executor       │──▶ Execute check based on type
│ (Check Type)   │    (row_count, null_check, etc.)
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Check Result   │──▶ Store result
└───────┬────────┘    Update check status
        │
        ▼
┌────────────────┐
│ Alert Manager  │──▶ Send alert if threshold breached
└───────┬────────┘
        │
        ▼
     Response
```

### 3. Scheduled Check Execution Flow

```
┌────────────────┐
│ Cron Trigger   │──▶ Schedule due for execution
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Scheduler      │──▶ GetSchedule(id)
│ Manager        │    Get list of checks
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Check Manager  │──▶ RunCheck(checkID) for each check
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Execution      │──▶ Store execution results
│ History        │    Update schedule.LastRunAt
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ Alert Manager  │──▶ Send alerts for failures
└────────────────┘
```

## Key Design Decisions

### 1. Middleware Chain Pattern

The middleware chain provides clean separation of cross-cutting concerns:

```go
// Build middleware chain
var httpHandler http.Handler = mux

// Add CORS middleware
corsMiddleware := middleware.NewCORSMiddleware([]string{"*"})
httpHandler = corsMiddleware.Handle(httpHandler)

// Add authentication middleware
if components.authManager != nil {
    authMiddleware := middleware.NewAuthMiddleware(components.authManager)
    httpHandler = authMiddleware.Handle(httpHandler)
}

// Add tenant middleware
if cfg.MultiTenant.Enabled {
    tenantMiddleware := middleware.NewTenantMiddleware(components.tenantManager)
    httpHandler = tenantMiddleware.Handle(httpHandler)
}

// Add authorization middleware
if components.authzManager != nil {
    authzMiddleware := middleware.NewAuthzMiddleware(components.authzManager)
    httpHandler = authzMiddleware.Handle(httpHandler)
}
```

### 2. Manager Pattern

Each domain area has a dedicated manager that encapsulates business logic:

- **TenantManager**: Tenant CRUD operations
- **PolicyManager**: Policy definition and enforcement
- **WorkflowEngine**: State machine transitions
- **DatasourceManager**: Connection management
- **CheckManager**: Check definition and execution
- **SchedulerManager**: Scheduled execution
- **AlertManager**: Notification channels

### 3. Connector Interface

All datasource connectors implement a common interface:

```go
type Connector interface {
    Connect(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error
    Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)
    GetTables(ctx context.Context) ([]TableInfo, error)
    GetColumns(ctx context.Context, table string) ([]ColumnInfo, error)
    GetRowCount(ctx context.Context, table string) (int64, error)
    Type() Type
}
```

### 4. Context-Based Tenant Resolution

Tenant information flows through the request context:

```go
// Set tenant in context (middleware)
ctx := tenant.WithTenantSlug(r.Context(), tenantSlug)

// Get tenant in handlers/services
tenantSlug, err := tenant.GetTenantSlug(r.Context())
```

## Scalability Considerations

1. **Horizontal Scaling**: Stateless API servers can be scaled horizontally
2. **Database Connection Pooling**: Efficient database connections
3. **Redis Caching**: Session and response caching
4. **Async Processing**: Long-running checks can be processed asynchronously
5. **Multi-Region**: Support for geographic distribution

## Security Layers

1. **Network**: TLS/HTTPS encryption
2. **Authentication**: OIDC JWT token validation
3. **Authorization**: OpenFGA relationship-based access control
4. **Tenant Isolation**: Data segregation per tenant
5. **Input Validation**: Request validation at API layer
