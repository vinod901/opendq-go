# Deployment Guide

This guide covers various deployment options for the OpenDQ Control Plane platform.

## Prerequisites

- Go 1.24 or higher
- Node.js 20+ (for frontend)
- PostgreSQL 14+
- Docker and Docker Compose (for containerized deployment)

## Local Development

### Using Docker Compose

The easiest way to get started:

```bash
# Start all services (PostgreSQL, OpenFGA, Keycloak, Marquez)
make dev

# Or manually:
docker-compose up -d

# Build and run the backend
go build -o opendq-server ./cmd/server
./opendq-server

# In another terminal, start the frontend
cd frontend
npm install
npm run dev
```

Access:
- Backend API: http://localhost:8080
- Frontend: http://localhost:5173
- Keycloak: http://localhost:8180 (admin/admin)
- OpenFGA: http://localhost:8081
- Marquez: http://localhost:5000
- Marquez Web: http://localhost:3001

### Manual Setup

1. **Set up PostgreSQL**:
```bash
createdb opendq
```

2. **Configure environment variables**:
```bash
cp .env.example .env
# Edit .env with your settings
```

3. **Run the backend**:
```bash
go run ./cmd/server
```

4. **Run the frontend**:
```bash
cd frontend
npm run dev
```

## Docker Deployment

### Build Docker Image

```bash
docker build -t opendq-server:latest .
```

### Run with Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=your_password \
  opendq-server:latest
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.20+)
- kubectl configured
- Helm 3.x (optional)

### Basic Deployment

Create Kubernetes manifests:

**namespace.yaml**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: opendq
```

**configmap.yaml**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: opendq-config
  namespace: opendq
data:
  SERVER_HOST: "0.0.0.0"
  SERVER_PORT: "8080"
  DB_DRIVER: "postgres"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "opendq"
  MULTITENANT_ENABLED: "true"
```

**secret.yaml**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: opendq-secrets
  namespace: opendq
type: Opaque
stringData:
  DB_PASSWORD: "your-password"
  OIDC_CLIENT_SECRET: "your-client-secret"
```

**deployment.yaml**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: opendq-server
  namespace: opendq
spec:
  replicas: 3
  selector:
    matchLabels:
      app: opendq-server
  template:
    metadata:
      labels:
        app: opendq-server
    spec:
      containers:
      - name: opendq-server
        image: opendq-server:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: opendq-config
        - secretRef:
            name: opendq-secrets
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

**service.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: opendq-server
  namespace: opendq
spec:
  selector:
    app: opendq-server
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

Apply the manifests:
```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## Cloud Deployments

### AWS (ECS)

1. Create ECR repository
2. Build and push image
3. Create ECS task definition
4. Create ECS service
5. Configure Application Load Balancer

### Google Cloud (GKE)

1. Create GKE cluster
2. Push image to GCR
3. Apply Kubernetes manifests
4. Configure Cloud SQL for PostgreSQL
5. Set up Ingress with SSL

### Azure (AKS)

1. Create AKS cluster
2. Push image to ACR
3. Apply Kubernetes manifests
4. Configure Azure Database for PostgreSQL
5. Set up Application Gateway

## Production Considerations

### Database

- Use managed database services (AWS RDS, Google Cloud SQL, Azure Database)
- Enable SSL connections
- Set up automated backups
- Configure read replicas for scaling

### Security

- Use TLS/HTTPS for all connections
- Configure OIDC provider (Okta/Keycloak)
- Set up OpenFGA with proper authorization model
- Use secrets management (Vault, AWS Secrets Manager, etc.)
- Enable audit logging

### Monitoring

- Set up Prometheus metrics
- Configure alerting (PagerDuty, Opsgenie)
- Use distributed tracing (Jaeger, Zipkin)
- Implement log aggregation (ELK, Splunk)

### Scaling

- Horizontal pod autoscaling in Kubernetes
- Database connection pooling
- Redis for caching and session management
- CDN for static frontend assets

### High Availability

- Deploy across multiple availability zones
- Use database replicas
- Implement circuit breakers
- Set up health checks and readiness probes

## Environment Variables Reference

### Required

- `DB_HOST`: Database host
- `DB_PASSWORD`: Database password

### Optional

- `SERVER_HOST`: Server bind address (default: 0.0.0.0)
- `SERVER_PORT`: Server port (default: 8080)
- `DB_PORT`: Database port (default: 5432)
- `DB_NAME`: Database name (default: opendq)
- `DB_USER`: Database user (default: postgres)
- `OIDC_ISSUER`: OIDC provider URL
- `OIDC_CLIENT_ID`: OIDC client ID
- `OIDC_CLIENT_SECRET`: OIDC client secret
- `OPENFGA_STORE_ID`: OpenFGA store ID
- `OPENFGA_API_HOST`: OpenFGA API host
- `MULTITENANT_ENABLED`: Enable multi-tenancy (default: true)
- `OPENLINEAGE_ENABLED`: Enable lineage tracking (default: true)
- `OPENLINEAGE_ENDPOINT`: OpenLineage endpoint URL

## Troubleshooting

### Backend won't start

- Check database connectivity
- Verify environment variables
- Check logs for errors

### Authentication issues

- Verify OIDC configuration
- Check redirect URLs
- Ensure client credentials are correct

### Authorization issues

- Verify OpenFGA is running
- Check authorization model is configured
- Ensure tuples are written correctly

### Performance issues

- Enable database query logging
- Check database connection pool settings
- Monitor resource usage (CPU, memory)
- Review slow query logs

## Support

For issues and questions:
- GitHub Issues: https://github.com/vinod901/opendq-go/issues
- Documentation: https://github.com/vinod901/opendq-go
