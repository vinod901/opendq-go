# OpenDQ Go - Enterprise Control-Plane Platform

An enterprise-grade control-plane platform built with Go and SvelteKit for data quality, governance, and lineage management.

## Features

### Core Capabilities
- **Multi-Tenant Architecture**: Isolated namespaces for multiple organizations
- **Policy-Driven**: Flexible policy engine for data governance and compliance
- **Workflow-Aware**: State machine-based workflow orchestration
- **Authorization**: Fine-grained access control using OpenFGA
- **Authentication**: OIDC integration (Okta/Keycloak compatible)
- **Data Lineage**: OpenLineage compatible for end-to-end data tracking
- **Extensible**: Plugin architecture for custom integrations

### Technology Stack
- **Backend**: Go 1.24+
- **ORM**: Ent for type-safe entity modeling
- **Authorization**: OpenFGA for relationship-based access control
- **Authentication**: go-oidc for OIDC/OAuth2 integration
- **Workflow Engine**: looplab/fsm for finite state machines
- **Lineage**: OpenLineage compatible event publishing
- **Frontend**: SvelteKit (to be implemented)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (SvelteKit)                  │
│                   Dashboard & Management UI                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTPS/REST
                         │
┌────────────────────────▼────────────────────────────────────┐
│                     API Gateway / Router                     │
│           ┌─────────────────────────────────────┐           │
│           │   Middleware Chain                  │           │
│           │  - CORS                             │           │
│           │  - Authentication (OIDC)            │           │
│           │  - Tenant Resolution                │           │
│           │  - Authorization (OpenFGA)          │           │
│           └─────────────────────────────────────┘           │
└─────────────────────────┬───────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          │               │               │
┌─────────▼─────┐ ┌──────▼──────┐ ┌─────▼──────┐
│   Tenant Mgmt  │ │Policy Engine│ │  Workflow  │
│               │ │             │ │   Engine   │
└───────────────┘ └─────────────┘ └────────────┘
          │               │               │
          └───────────────┼───────────────┘
                          │
          ┌───────────────┼───────────────┐
          │               │               │
┌─────────▼─────┐ ┌──────▼──────┐ ┌─────▼──────┐
│   OpenFGA     │ │ OIDC Provider│ │ OpenLineage│
│ Authorization │ │(Okta/Keycloak│ │   Backend  │
└───────────────┘ └─────────────┘ └────────────┘
          │
┌─────────▼─────────────────────┐
│      Database (PostgreSQL)     │
│     - Tenants                  │
│     - Users                    │
│     - Policies                 │
│     - Workflows                │
│     - Lineage Events           │
└───────────────────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.24 or higher
- PostgreSQL 14+
- OpenFGA server (optional, for authorization)
- OIDC provider (Okta/Keycloak) (optional, for authentication)

### Configuration

Create a `.env` file or set environment variables:

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=opendq
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSLMODE=disable

# OIDC Authentication (optional)
OIDC_ISSUER=https://your-issuer.okta.com
OIDC_CLIENT_ID=your_client_id
OIDC_CLIENT_SECRET=your_client_secret
OIDC_REDIRECT_URL=http://localhost:8080/auth/callback

# OpenFGA Authorization (optional)
OPENFGA_STORE_ID=your_store_id
OPENFGA_API_HOST=http://localhost:8081
OPENFGA_AUTH_MODEL=your_model_id

# Multi-Tenancy
MULTITENANT_ENABLED=true
MULTITENANT_ISOLATION=namespace

# OpenLineage (optional)
OPENLINEAGE_ENABLED=true
OPENLINEAGE_ENDPOINT=http://localhost:5000
OPENLINEAGE_NAMESPACE=opendq
```

### Build and Run

```bash
# Build the application
go build -o opendq-server ./cmd/server

# Run the server
./opendq-server
```

Or run directly:

```bash
go run ./cmd/server
```

### Development with Docker Compose

A Docker Compose setup will be provided to run all required services.

## API Endpoints

### Health Check
```
GET /health
```

### Tenant Management
```
GET    /api/v1/tenants       - List all tenants
POST   /api/v1/tenants       - Create a tenant
GET    /api/v1/tenants/{id}  - Get tenant details
PUT    /api/v1/tenants/{id}  - Update tenant
DELETE /api/v1/tenants/{id}  - Delete tenant
```

### Policy Management
```
GET    /api/v1/policies       - List policies
POST   /api/v1/policies       - Create a policy
GET    /api/v1/policies/{id}  - Get policy details
PUT    /api/v1/policies/{id}  - Update policy
DELETE /api/v1/policies/{id}  - Delete policy
```

### Workflow Management
```
GET    /api/v1/workflows       - List workflows
POST   /api/v1/workflows       - Create a workflow
GET    /api/v1/workflows/{id}  - Get workflow details
POST   /api/v1/workflows/{id}  - Trigger workflow transition
```

### Data Lineage
```
GET    /api/v1/lineage  - Query lineage events
POST   /api/v1/lineage  - Create lineage event
```

## Data Models

### Tenant
Multi-tenant isolation with namespace support.

### User
OIDC-authenticated users with role-based access.

### Policy
Policy definitions for data governance and compliance.

### Workflow
State machine-based workflows with FSM transitions.

### LineageEvent
OpenLineage-compatible data lineage events.

## Workflow Examples

### Data Quality Workflow
States: pending → running → validating → passed/failed → completed

### Approval Workflow
States: draft → submitted → under_review → approved/rejected

### Data Pipeline Workflow
States: pending → extracting → transforming → loading → completed

## Authorization Model (OpenFGA)

The platform uses OpenFGA for fine-grained authorization:

- **Tenants**: Organization-level isolation
- **Relations**: owner, admin, editor, viewer, member
- **Resources**: tenants, policies, workflows, lineage events

Example: Check if user can edit a policy
```go
allowed, err := authzManager.Check(ctx, "user:alice", "editor", "policy:policy-123")
```

## OpenLineage Integration

The platform emits OpenLineage-compatible events for data lineage tracking:

```json
{
  "eventType": "START",
  "eventTime": "2024-01-01T00:00:00Z",
  "run": {
    "runId": "run-123",
    "facets": {}
  },
  "job": {
    "namespace": "opendq",
    "name": "data-pipeline-1",
    "facets": {}
  },
  "inputs": [...],
  "outputs": [...]
}
```

## Development

### Project Structure
```
opendq-go/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── auth/            # OIDC authentication
│   ├── authorization/   # OpenFGA integration
│   ├── workflow/        # FSM workflow engine
│   ├── lineage/         # OpenLineage client
│   ├── tenant/          # Multi-tenant management
│   ├── policy/          # Policy engine
│   └── middleware/      # HTTP middleware
├── pkg/
│   ├── config/          # Configuration management
│   └── models/          # Shared models
├── api/
│   └── http/            # HTTP handlers
├── ent/
│   └── schema/          # Ent entity schemas
└── frontend/            # SvelteKit frontend (TODO)
```

### Adding a New Entity

1. Create schema in `ent/schema/`
2. Generate code: `go generate ./ent`
3. Add business logic in `internal/`
4. Add HTTP handlers in `api/http/`

### Adding a New Workflow

1. Define workflow in `internal/workflow/engine.go`
2. Register in `RegisterStandardWorkflows()`
3. Add API endpoints for transitions

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/workflow
```

## Deployment

### Docker
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o opendq-server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/opendq-server /usr/local/bin/
ENTRYPOINT ["opendq-server"]
```

### Kubernetes
Helm charts and Kubernetes manifests will be provided.

## Security

- OIDC/OAuth2 authentication
- OpenFGA for fine-grained authorization
- Multi-tenant isolation
- TLS/HTTPS support
- Audit logging (planned)

## Contributing

Contributions are welcome! Please follow these guidelines:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions, please open a GitHub issue.

## Roadmap

- [x] Core backend architecture
- [x] OIDC authentication
- [x] OpenFGA authorization
- [x] Workflow engine
- [x] OpenLineage integration
- [x] Multi-tenant support
- [ ] Ent code generation and database setup
- [ ] Complete API implementation
- [ ] SvelteKit frontend
- [ ] Docker Compose setup
- [ ] Kubernetes deployment
- [ ] Comprehensive tests
- [ ] API documentation (OpenAPI)
- [ ] Admin CLI tool
- [ ] Metrics and monitoring
- [ ] Audit logging