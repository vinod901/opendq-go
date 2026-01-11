# Development Guide

This guide covers local development setup, running services, and common development workflows.

## Prerequisites

- **Go 1.24+**: [Download](https://go.dev/dl/)
- **Docker & Docker Compose**: [Download](https://www.docker.com/products/docker-desktop/)
- **Node.js 18+** (for frontend): [Download](https://nodejs.org/)
- **Git**: [Download](https://git-scm.com/)

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/vinod901/opendq-go.git
cd opendq-go
```

### 2. Start Infrastructure Services

```bash
# Start all services (PostgreSQL, OpenFGA, Keycloak, Marquez, Redis)
make dev

# Or using docker-compose directly
docker-compose up -d
```

### 3. Wait for Services

```bash
# Check service status
docker-compose ps

# Services should show as "healthy"
```

### 4. Run the Backend

```bash
# Copy environment file
cp .env.example .env

# Run the server
make run

# Or directly
go run ./cmd/server
```

### 5. Run the Frontend (Optional)

```bash
cd frontend
npm install
npm run dev
```

## Development Commands

### Makefile Commands

```bash
make help              # Show all available commands
make deps              # Download Go dependencies
make build             # Build the server binary
make run               # Run the server
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make fmt               # Format Go code
make lint              # Run linter (requires golangci-lint)
make vet               # Run go vet
make ent-generate      # Generate Ent code
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make docker-logs       # View Docker logs
make dev               # Start full development environment
make install-tools     # Install development tools
```

### Running All Services Together

```bash
# Start all services including frontend, backend, and infrastructure
make dev-all
```

This starts:
- PostgreSQL (port 5432)
- OpenFGA (ports 8081, 8082, 3000)
- OpenFGA Playground UI (port 3000)
- Keycloak (port 8180)
- Marquez API (port 5000)
- Marquez Web (port 3001)
- Redis (port 6379)
- OpenDQ Backend (port 8080)
- OpenDQ Frontend (port 5173)

## Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| OpenDQ API | http://localhost:8080 | - |
| OpenDQ Frontend | http://localhost:5173 | - |
| OpenFGA Playground | http://localhost:3000 | - |
| Keycloak Admin | http://localhost:8180 | admin / admin |
| Marquez Web | http://localhost:3001 | - |
| Marquez API | http://localhost:5000/api/v1 | - |

## Configuration

### Environment Variables

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
DB_PASSWORD=postgres
DB_SSLMODE=disable

# OIDC Authentication
OIDC_ISSUER=http://localhost:8180/realms/master
OIDC_CLIENT_ID=opendq-client
OIDC_CLIENT_SECRET=your_secret
OIDC_REDIRECT_URL=http://localhost:8080/auth/callback

# OpenFGA Authorization
OPENFGA_STORE_ID=
OPENFGA_API_HOST=http://localhost:8081
OPENFGA_AUTH_MODEL=

# Multi-Tenancy
MULTITENANT_ENABLED=true
MULTITENANT_ISOLATION=namespace

# OpenLineage
OPENLINEAGE_ENABLED=true
OPENLINEAGE_ENDPOINT=http://localhost:5000
OPENLINEAGE_NAMESPACE=opendq
```

## Keycloak Setup (Authentication)

### Create OIDC Client

1. Open Keycloak: http://localhost:8180
2. Login: admin / admin
3. Go to Clients → Create
4. Settings:
   - Client ID: `opendq-client`
   - Client Protocol: `openid-connect`
   - Root URL: `http://localhost:8080`
5. Save and configure:
   - Access Type: `confidential`
   - Valid Redirect URIs: `http://localhost:8080/auth/callback`
   - Web Origins: `http://localhost:8080`
6. Copy client secret from Credentials tab
7. Update `.env`: `OIDC_CLIENT_SECRET=<secret>`

### Create Test User

1. Go to Users → Add User
2. Set username and email
3. Set password in Credentials tab

## OpenFGA Setup (Authorization)

### Using OpenFGA Playground

1. Open: http://localhost:3000
2. Create a new store
3. Upload authorization model from `openfga-model.json`
4. Note the Store ID and Model ID
5. Update `.env`:
   ```
   OPENFGA_STORE_ID=<store_id>
   OPENFGA_AUTH_MODEL=<model_id>
   ```

### Using FGA CLI

```bash
# Install FGA CLI
go install github.com/openfga/cli/cmd/fga@latest

# Create store
fga store create --api-url http://localhost:8081 --name opendq

# Write model
fga model write --api-url http://localhost:8081 --store-id <store_id> < openfga-model.json
```

## Database Migrations

### Using Ent

```bash
# Generate Ent code
make ent-generate

# Auto-migrate (development only)
# Migrations run automatically on server start
```

### Manual Migration

```bash
# Connect to PostgreSQL
docker exec -it opendq-postgres psql -U postgres -d opendq

# Run SQL commands
\dt  # List tables
```

## Testing

### Run All Tests

```bash
make test
```

### Run Specific Tests

```bash
# Run tests for a package
go test ./internal/check/...

# Run with verbose output
go test -v ./internal/check/...

# Run a specific test
go test -v -run TestRowCountCheck ./internal/check/...
```

### Test Coverage

```bash
make test-coverage
# Opens coverage.html in browser
```

## Linting

### Install Linter

```bash
make install-tools
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Run Linter

```bash
make lint
```

## API Development

### Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# List datasources (requires auth in production)
curl http://localhost:8080/api/v1/datasources \
  -H "X-Tenant: default"

# Create datasource
curl -X POST http://localhost:8080/api/v1/datasources \
  -H "Content-Type: application/json" \
  -H "X-Tenant: default" \
  -d '{
    "name": "Test DB",
    "type": "postgres",
    "connection": {
      "host": "localhost",
      "port": 5432,
      "database": "test",
      "username": "test",
      "password": "test"
    }
  }'
```

### OpenAPI Documentation

OpenAPI spec is available at `/api/docs` when the server is running.

## Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
opendq-go/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go
├── internal/            # Private application code
│   ├── auth/            # OIDC authentication
│   ├── authorization/   # OpenFGA authorization
│   ├── workflow/        # Workflow engine
│   ├── lineage/         # OpenLineage client
│   ├── tenant/          # Multi-tenant management
│   ├── policy/          # Policy engine
│   ├── datasource/      # Datasource connectors
│   ├── check/           # Data quality checks
│   ├── scheduler/       # Check scheduling
│   ├── alerting/        # Alert notifications
│   ├── view/            # Logical views
│   └── middleware/      # HTTP middleware
├── pkg/                 # Public packages
│   └── config/          # Configuration
├── api/
│   └── http/            # HTTP handlers
├── ent/
│   └── schema/          # Ent entity schemas
├── frontend/            # SvelteKit frontend
├── docker/              # Docker configs
├── docs/
│   ├── arch/            # Architecture docs
│   └── product/         # Product docs
├── docker-compose.yml   # Service orchestration
├── Makefile             # Build commands
├── go.mod               # Go dependencies
└── .env.example         # Environment template
```

## Adding New Features

### 1. Add Entity Schema

```go
// ent/schema/myentity.go
type MyEntity struct {
    ent.Schema
}

func (MyEntity) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique().Immutable(),
        // ... fields
    }
}
```

### 2. Generate Code

```bash
make ent-generate
```

### 3. Add Business Logic

```go
// internal/myentity/myentity.go
type Manager struct {
    client *ent.Client
}

func (m *Manager) Create(ctx context.Context, ...) (*ent.MyEntity, error) {
    // Implementation
}
```

### 4. Add HTTP Handler

```go
// api/http/handler.go
func (h *Handler) handleMyEntities(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### 5. Register Routes

```go
mux.HandleFunc("/api/v1/myentities", h.handleMyEntities)
```

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker-compose logs postgres
docker-compose logs openfga

# Restart services
docker-compose restart
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker exec opendq-postgres pg_isready

# Check connection
docker exec -it opendq-postgres psql -U postgres -d opendq
```

### OpenFGA Errors

```bash
# Check OpenFGA logs
docker-compose logs openfga

# Verify store exists
curl http://localhost:8081/stores
```

### Port Conflicts

```bash
# Find process using port
lsof -i :8080

# Change port in .env or docker-compose.yml
```

## IDE Setup

### VS Code

Recommended extensions:
- Go (golang.go)
- Svelte for VS Code
- Docker
- YAML

Settings (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true
}
```

### GoLand

- Enable Go Modules integration
- Configure golangci-lint as external tool
