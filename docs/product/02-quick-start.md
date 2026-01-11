# Quick Start Guide

Get OpenDQ up and running in 5 minutes!

## 📋 Prerequisites

Before you begin, ensure you have:
- Docker (20.10+) and Docker Compose (2.0+)
- Git
- 4GB RAM minimum (8GB recommended)
- Ports available: 5432, 5000, 3001, 6379, 8081, 8180

## 🚀 Installation Steps

### Step 1: Clone the Repository

```bash
git clone https://github.com/yourusername/opendq-go.git
cd opendq-go
```

### Step 2: Start the Services

```bash
# Start all services in detached mode
docker-compose up -d

# Wait for services to be ready (takes ~30 seconds)
docker-compose ps
```

You should see all services running:
```
NAME                   STATUS              PORTS
opendq-postgres        Up (healthy)        0.0.0.0:5432->5432/tcp
opendq-redis           Up (healthy)        0.0.0.0:6379->6379/tcp
opendq-openfga         Up (healthy)        0.0.0.0:8081->8080/tcp, 0.0.0.0:3002->3000/tcp
opendq-keycloak        Up (healthy)        0.0.0.0:8180->8080/tcp
opendq-marquez         Up (healthy)        0.0.0.0:5000->5000/tcp
opendq-marquez-web     Up                  0.0.0.0:3001->3000/tcp
```

### Step 3: Verify Installation

Check that all services are responding:

```bash
# Check PostgreSQL
docker exec opendq-postgres pg_isready

# Check Redis
docker exec opendq-redis redis-cli ping

# Check Marquez API
curl http://localhost:5000/api/v1/namespaces

# Check Keycloak
curl http://localhost:8180/health/ready

# Check OpenFGA
curl http://localhost:8081/healthz
```

## 🌐 Access the Application

### Web Interfaces

| Service | URL | Credentials |
|---------|-----|-------------|
| **OpenFGA Playground** | http://localhost:3002 | No auth required |
| **Marquez Web** (Lineage UI) | http://localhost:3001 | No auth required (dev mode) |
| **Keycloak Admin** | http://localhost:8180 | admin / admin |
| **Marquez API** | http://localhost:5000/api/v1 | No auth required (dev mode) |

### API Endpoints

The API server (when built and running) will be available at:
```
http://localhost:8080
```

## 🎯 Your First Steps

### 1. View Lineage Data

Open your browser and navigate to:
```
http://localhost:3001
```

This opens the Marquez web UI where you can explore data lineage.

### 2. Create a Namespace

Namespaces organize related datasets. Create one via API:

```bash
curl -X POST http://localhost:5000/api/v1/namespaces/my-namespace \
  -H "Content-Type: application/json" \
  -d '{
    "ownerName": "data-team",
    "description": "My first namespace"
  }'
```

### 3. Register a Dataset

Register a dataset to track:

```bash
curl -X POST http://localhost:5000/api/v1/namespaces/my-namespace/datasets/customers \
  -H "Content-Type: application/json" \
  -d '{
    "type": "DB_TABLE",
    "name": "customers",
    "physicalName": "public.customers",
    "sourceName": "my-database",
    "fields": [
      {
        "name": "id",
        "type": "INTEGER",
        "description": "Customer ID"
      },
      {
        "name": "email",
        "type": "VARCHAR",
        "description": "Customer email"
      }
    ],
    "description": "Customer master table"
  }'
```

### 4. Record Lineage Events

Track data transformations using OpenLineage:

```bash
curl -X POST http://localhost:5000/api/v1/lineage \
  -H "Content-Type: application/json" \
  -d '{
    "eventType": "COMPLETE",
    "eventTime": "2026-01-11T10:00:00.000Z",
    "run": {
      "runId": "550e8400-e29b-41d4-a716-446655440000"
    },
    "job": {
      "namespace": "my-namespace",
      "name": "customer-etl"
    },
    "inputs": [{
      "namespace": "my-namespace",
      "name": "customers"
    }],
    "outputs": [{
      "namespace": "my-namespace",
      "name": "customers_enriched"
    }]
  }'
```

### 5. Explore in the UI

Go back to http://localhost:3001 and you'll see:
- Your namespace listed
- The datasets you registered
- The lineage graph showing data flow

## 🔧 Running the Full Stack

### Option 1: Run Everything at Once

Use the Makefile command to start all services:

```bash
# Start infrastructure + backend + frontend
make dev-all
```

This will:
- Start PostgreSQL, Redis, OpenFGA, Keycloak, Marquez
- Start the OpenDQ backend on http://localhost:8080
- Start the OpenDQ frontend on http://localhost:5173

### Option 2: Run Components Separately

```bash
# Start infrastructure only
make dev

# In a new terminal: Start the backend
make dev-backend

# In another terminal: Start the frontend
make dev-frontend
```

## 🔧 Building the Go Application

To build and run the OpenDQ Go backend:

```bash
# Install dependencies
go mod download

# Build the application
make build

# Run the server
./opendq-server
```

The server will start on `http://localhost:8080`

## 📊 Quick Architecture Diagram

```
┌──────────────┐
│   Browser    │
└──────┬───────┘
       │
       │ HTTP
       ▼
┌──────────────────┐     ┌──────────────┐
│  Marquez Web UI  │────▶│  Marquez API │
│  (Port 3001)     │     │  (Port 5000) │
└──────────────────┘     └──────┬───────┘
                                │
                                ▼
                         ┌──────────────┐
                         │  PostgreSQL  │
                         │  (Port 5432) │
                         └──────────────┘
```

## 🛑 Stopping the Services

When you're done:

```bash
# Stop all services
docker-compose down

# Stop and remove all data (CAUTION: destroys all data!)
docker-compose down -v
```

## ⚠️ Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose logs [service-name]

# Restart a specific service
docker-compose restart [service-name]
```

### Port conflicts
If you see port already in use errors:
```bash
# Check what's using the port
sudo lsof -i :5432

# Change the port mapping in docker-compose.yml
# Example: "5433:5432" instead of "5432:5432"
```

### Database connection issues
```bash
# Verify PostgreSQL is healthy
docker exec opendq-postgres pg_isready -U postgres

# Check connection
docker exec -it opendq-postgres psql -U postgres -d opendq
```

### Services unhealthy
```bash
# Wait longer - some services take time to initialize
docker-compose ps

# Force recreate
docker-compose up -d --force-recreate
```

## 📚 Next Steps

Now that you have OpenDQ running:

1. **[Installation Guide](03-installation.md)** - Learn about production deployment
2. **[Architecture Overview](04-architecture.md)** - Understand the system design
3. **[Database Connectors](09-database-connectors.md)** - Connect to your data sources
4. **[API Reference](13-api-reference.md)** - Explore the API endpoints

## 💡 Tips

- **Development Mode**: The quick start runs in development mode with authentication disabled
- **Data Persistence**: Data is stored in `./volumes/` directory
- **Logs**: View logs with `docker-compose logs -f [service-name]`
- **Clean Start**: Use `docker-compose down -v` for a fresh start

Happy data quality monitoring! 🎉
