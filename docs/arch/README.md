# OpenDQ Architecture Documentation

This documentation provides comprehensive technical details about how OpenDQ is architected and developed.

## Table of Contents

1. [Architecture Overview](01-architecture-overview.md) - System design and component interactions
2. [Authentication](02-authentication.md) - OIDC/Keycloak authentication flow
3. [Authorization](03-authorization.md) - OpenFGA relationship-based access control
4. [Data Layer (Ent ORM)](04-data-layer-ent.md) - Entity schemas and database design
5. [API Design](05-api-design.md) - HTTP handlers and RESTful API structure
6. [Datasources](06-datasources.md) - Datasource connectivity and connectors
7. [Data Quality Checks](07-data-quality-checks.md) - Check definitions and execution
8. [Scheduling & Alerting](08-scheduling-alerting.md) - Scheduled executions and notifications
9. [Multi-Tenancy](09-multi-tenancy.md) - Tenant isolation and management
10. [Development Guide](10-development-guide.md) - Local development setup and workflow

## Quick Reference

### Project Structure

```
opendq-go/
├── cmd/
│   └── server/              # Main application entry point
├── internal/
│   ├── auth/                # OIDC authentication
│   ├── authorization/       # OpenFGA integration
│   ├── workflow/            # FSM workflow engine
│   ├── lineage/             # OpenLineage client
│   ├── tenant/              # Multi-tenant management
│   ├── policy/              # Policy engine
│   ├── datasource/          # Datasource connectivity
│   ├── check/               # Data quality checks
│   ├── scheduler/           # Scheduled check execution
│   ├── alerting/            # Alert notifications
│   ├── view/                # Logical views
│   └── middleware/          # HTTP middleware
├── pkg/
│   └── config/              # Configuration management
├── api/
│   └── http/                # HTTP handlers
├── ent/
│   └── schema/              # Ent entity schemas
├── frontend/                # SvelteKit frontend
├── docker/                  # Docker configurations
└── docs/
    ├── arch/                # Architecture documentation (you are here)
    └── product/             # Product documentation
```

### Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Backend | Go 1.24+ | API server, business logic |
| ORM | Ent | Type-safe entity modeling |
| Authentication | go-oidc | OIDC/OAuth2 integration |
| Authorization | OpenFGA | Relationship-based access control |
| Workflow | looplab/fsm | Finite state machines |
| Lineage | OpenLineage | Data lineage tracking |
| Frontend | SvelteKit | Web application |
| Database | PostgreSQL | Primary data store |
| Cache | Redis | Session & response caching |

### Key Design Principles

1. **Multi-Tenancy First**: Every resource is tenant-scoped
2. **Fine-Grained Authorization**: OpenFGA for relationship-based access control
3. **Standards-Based**: OpenLineage, OpenAPI, OIDC compliance
4. **Extensible**: Plugin architecture for connectors and integrations
5. **Cloud Native**: Container-ready, horizontally scalable
