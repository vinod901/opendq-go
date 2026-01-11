# Introduction to OpenDQ

## What is OpenDQ?

OpenDQ is an open-source, enterprise-grade **Data Quality and Governance Platform** that helps organizations ensure their data is accurate, consistent, and trustworthy. Built with modern technologies (Go backend, Svelte frontend), OpenDQ provides a comprehensive solution for data quality management at scale.

## 🎯 Key Objectives

1. **Data Quality Assurance** - Continuously monitor and validate data quality
2. **Data Lineage Tracking** - Understand data flow and transformations
3. **Policy Enforcement** - Define and enforce data governance rules
4. **Workflow Automation** - Automate data quality processes
5. **Multi-tenancy Support** - Isolate data and configurations per organization

## 🏗️ Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Svelte)                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ Lineage  │  │ Policies │  │ Workflows│  │ Tenants  │       │
│  │   UI     │  │    UI    │  │    UI    │  │    UI    │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└────────────────────────┬────────────────────────────────────────┘
                         │ REST API
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                     API Gateway (Go)                             │
│  ┌─────────────────────────────────────────────────────┐        │
│  │              Authentication Middleware               │        │
│  │                  (Keycloak OIDC)                    │        │
│  └─────────────────────────────────────────────────────┘        │
└─────────────────────────┬───────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
┌───────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐
│   Lineage    │  │   Policy    │  │  Workflow   │
│   Service    │  │   Service   │  │   Engine    │
└───────┬──────┘  └──────┬──────┘  └──────┬──────┘
        │                │                 │
        │         ┌──────▼──────┐          │
        │         │    Auth     │          │
        └─────────┤  Service    ├──────────┘
                  │  (OpenFGA)  │
                  └──────┬──────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
┌───────▼──────┐  ┌──────▼──────┐  ┌────▼─────┐
│  PostgreSQL  │  │   Marquez   │  │  Redis   │
│   Database   │  │  (Lineage)  │  │  Cache   │
└──────────────┘  └─────────────┘  └──────────┘
```

## 🔑 Core Components

### 1. **Frontend (Svelte)**
- Modern, reactive web interface
- Real-time data quality dashboards
- Interactive lineage visualization
- Policy configuration UI

### 2. **API Gateway (Go)**
- RESTful API endpoints
- Request routing and validation
- Authentication & authorization
- Rate limiting and caching

### 3. **Services**

#### Lineage Service
- Captures data lineage events
- Integrates with OpenLineage
- Provides lineage query APIs
- Visualizes data flow

#### Policy Service
- Defines data quality rules
- Executes policy validations
- Reports policy violations
- Manages policy lifecycle

#### Workflow Engine
- Orchestrates data quality workflows
- Schedules automated checks
- Handles event-driven processes
- Manages task dependencies

#### Authorization Service
- Fine-grained access control (OpenFGA)
- Role-based permissions
- Tenant isolation
- API security

### 4. **Data Stores**

#### PostgreSQL
- Primary application database
- Stores metadata and configurations
- Manages tenant data

#### Marquez
- OpenLineage backend
- Lineage graph storage
- Temporal lineage queries

#### Redis
- Session management
- Response caching
- Real-time data

## 💡 Key Concepts

### Data Lineage
Data lineage tracks the journey of data from source to destination, including all transformations along the way.

```
Source DB → ETL Process → Data Warehouse → Analytics → Reports
    ↓           ↓              ↓              ↓          ↓
  Track     Transform      Aggregate       Analyze    Deliver
```

### Data Quality Policies
Policies define rules that data must satisfy:
- **Completeness**: No missing values in critical fields
- **Accuracy**: Data matches expected patterns/ranges
- **Consistency**: Data is consistent across systems
- **Timeliness**: Data is up-to-date
- **Validity**: Data conforms to business rules

### Workflows
Automated processes that:
1. Collect data quality metrics
2. Execute validation rules
3. Generate alerts on violations
4. Trigger remediation actions

### Multi-tenancy
Isolate data and configurations per organization:
- Each tenant has separate data namespace
- Tenant-specific policies and workflows
- Role-based access within tenants
- Billing and usage tracking per tenant

## 🌟 Why OpenDQ?

### Open Source
- No vendor lock-in
- Community-driven development
- Transparent roadmap
- Free to use and modify

### Cloud Native
- Container-based deployment
- Horizontal scalability
- Kubernetes-ready
- Microservices architecture

### Standards-Based
- OpenLineage integration
- OAuth 2.0 / OIDC authentication
- REST API with OpenAPI spec
- Standard database connectors

### Enterprise-Ready
- Multi-tenant architecture
- Fine-grained authorization
- Audit logging
- High availability support

## 📊 Use Cases

### 1. Data Quality Monitoring
Monitor data quality metrics across your data pipeline and get alerted when quality degrades.

### 2. Compliance & Governance
Ensure data meets regulatory requirements (GDPR, HIPAA, SOX) with automated policy enforcement.

### 3. Data Migration Validation
Validate data during migrations to ensure accuracy and completeness.

### 4. Impact Analysis
Understand downstream impact of schema or data changes before making them.

### 5. Root Cause Analysis
Trace data quality issues back to their source using lineage information.

## 🚦 Next Steps

Ready to get started? Continue to:
- [Quick Start Guide](02-quick-start.md) - Get OpenDQ running in minutes
- [Installation Guide](03-installation.md) - Detailed setup instructions
- [Architecture Overview](04-architecture.md) - Deep dive into system design
