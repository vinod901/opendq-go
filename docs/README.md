# OpenDQ Documentation

Welcome to the OpenDQ documentation. This documentation is organized into two main sections:

## 📚 Documentation Structure

### [Architecture Documentation](arch/README.md)

Technical documentation for developers working on or extending OpenDQ:

- [Architecture Overview](arch/01-architecture-overview.md) - System design and component interactions
- [Authentication](arch/02-authentication.md) - OIDC/Keycloak authentication flow
- [Authorization](arch/03-authorization.md) - OpenFGA relationship-based access control
- [Data Layer (Ent ORM)](arch/04-data-layer-ent.md) - Entity schemas and database design
- [API Design](arch/05-api-design.md) - HTTP handlers and RESTful API structure
- [Datasources](arch/06-datasources.md) - Datasource connectivity and connectors
- [Data Quality Checks](arch/07-data-quality-checks.md) - Check definitions and execution
- [Scheduling & Alerting](arch/08-scheduling-alerting.md) - Scheduled executions and notifications
- [Multi-Tenancy](arch/09-multi-tenancy.md) - Tenant isolation and management
- [Development Guide](arch/10-development-guide.md) - Local development setup and workflow

### [Product Documentation](product/README.md)

User-focused documentation for using OpenDQ:

- [Introduction](product/01-introduction.md) - What is OpenDQ and why use it
- [Quick Start](product/02-quick-start.md) - Get started in 5 minutes
- [User Guide](product/03-user-guide.md) - Complete user guide
- [Feature Guides](product/04-features.md) - In-depth feature documentation
- [API Reference](product/05-api-reference.md) - Complete API documentation

## 🚀 Quick Links

| Topic | Link |
|-------|------|
| **Get Started** | [Quick Start Guide](product/02-quick-start.md) |
| **Run Locally** | [Development Guide](arch/10-development-guide.md) |
| **API Reference** | [API Documentation](product/05-api-reference.md) |
| **Architecture** | [Architecture Overview](arch/01-architecture-overview.md) |

## 🎯 What is OpenDQ?

OpenDQ is an open-source, enterprise-grade **Data Quality and Governance Platform** that helps organizations ensure their data is accurate, consistent, and trustworthy.

### Key Features

- **Multi-Datasource Connectivity**: Connect to databases, data warehouses, and cloud storage
- **Data Quality Checks**: Row count, null checks, freshness, custom SQL, and more
- **Scheduled Execution**: Run checks automatically on a schedule
- **Alerting**: Get notified on failures via Slack, email, PagerDuty, etc.
- **Multi-Tenant**: Support for multiple organizations
- **Fine-Grained Authorization**: Role-based access control with OpenFGA
- **Data Lineage**: Track data flow with OpenLineage integration

### Technology Stack

- **Backend**: Go with Ent ORM
- **Frontend**: SvelteKit
- **Authentication**: OIDC (Keycloak/Okta compatible)
- **Authorization**: OpenFGA
- **Data Lineage**: OpenLineage (Marquez)
- **Database**: PostgreSQL
